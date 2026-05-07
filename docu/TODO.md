# AI Interview TODO 总索引

> 项目级开发任务清单。本文件是入口，包含跨版本的**设计决策、错误码体系、禁止事项**。具体任务清单按版本拆分到三个子文件。

---

## 版本路线图

| 版本 | 目标 | 文件 |
|---|---|---|
| **v1 — MVP** | 跑通一次完整面试：游客上传简历 → AI 出题 → 语音面试 → 拿报告 → 提交问卷 | [todo-v1.md](./todo-v1.md) |
| **v2 — 好用** | 体验完善 + SFT 数据闭环 + 文字面试 + 多 Provider 运行时切换 | [todo-v2.md](./todo-v2.md) |
| **v3 — 全产品** | 招聘场景 + 用户面经 ChatBot + 监控 + 运营后台 | [todo-v3.md](./todo-v3.md) |

---

## 设计决策与避坑记录

> 对标项目 `interview-guide`（Java 实现）已踩过的坑 + 我方架构决策。**所有新代码必须遵守**。

### 1. 音频传输：WebSocket 全双工（不再用 HTTP per-turn）

- **原方案**：前端 VAD 截断 → `POST /v1/interview/audio` 整段上传 → HTTP ASR
- **新方案**：前端流式 PCM → `WS /v1/interview/ws/{interview_id}` → 服务端 server-VAD + 流式 ASR/LLM/TTS
- **收益**：首包延迟 200-400ms（对标 Qwen3 实时模型），告别 3-5s 等待
- **影响**：handler 层重写 / `useAudio.ts` 重写 / `qwen_asr.go` `qwen_tts.go` 改 WebSocket 实现

### 2. 出题数据源：PostgreSQL + pgvector，不引入 Milvus

- Skill markdown 定义考察范围 + Tag-RAG 检索题库
- 题库表加 `tags`、`related_concepts`、`followup_question_ids` 字段，模拟知识图谱关联（成本 1/10，收益 90%）
- 与简历、未来知识库共用同一存储，精简架构

### 3. 多 Provider：v1 编译期静态、v2 运行时动态

- **v1**：从 `.env` 读 Provider 配置，重启生效
- **v2**：设置页 + 落盘 `~/.ai-interview/llm-providers.yml` + reload 接口

### 4. 评估引擎：统一管道（文字 + 语音共用）

- 对标项目踩坑：先各做一套，后期重构成 `UnifiedEvaluationService`
- **新方案**：从 day 1 抽象 `EvaluationPipeline` 接口

```go
type EvaluationPipeline interface {
    BatchScore(ctx context.Context, turns []Turn) ([]DimensionScore, error)  // 分批评估
    Aggregate(ctx context.Context, scores []DimensionScore) (Report, error)  // 二次汇总
    Fallback(err error) Report                                               // 降级兜底
}
```

- `Turn` 结构同时支持 `text` 和 `audio_transcript` 字段，输入归一化

### 5. Worker 任务状态机

- 状态：`pending → processing → completed / failed`，前端可见
- 处理前预校验实体存在（被删则 ACK 丢弃）
- v2 加死信队列：失败 ≥ 3 次入 DLX，前端显示「重试」按钮

### 6. LLM 限流（v1 必须）

- Redis Lua 滑动窗口
- v1：`IP`（10 次/分钟）+ `USER`（30 次/分钟）
- v2：补 `GLOBAL` 维度
- 关键端点：`/v1/interview/ws/*`、`/v1/interview/code/submit`、`/v1/resume/parse`、`/v1/rag/query`

### 7. 结构化输出强制重试

- 所有 JSON 输出必须经 `StructuredOutputInvoker`
- 解析失败自动重试 3 次 + 降级
- 影响：`evaluator`、`response_analyzer`、`question_selector`、`code_judge`、简历解析

### 8. 简历去重（SHA-256 内容哈希）

- `resumes` 表加 `content_hash` 列，上传时先查重，命中直接返回已分析结果
- TTL：游客简历 24h，登录用户永久

### 9. 回声防护与音色（已知遗留）

- 对标项目未解决的痛点：无耳机回声泄漏、TTS 音色单一、弱网音频断续
- **我方策略**：
  - v1 接受同样限制，前端 UI 提示「建议佩戴耳机」+ 客户端 AEC（`echoCancellation: true`）
  - v2 探索：客户端 VAD 降噪、多音色支持

### 10. 事务规范

- **禁止**在事务内调用 LLM / S3 / WebSocket（占用 DB 连接，长事务风险）

### 11. 用户面经 ChatBot：Tag-RAG，不上知识图谱

- 工程上 KG（Neo4j 等）维护成本 10x，大模型时代 ROI 在下降
- **方案**：pgvector + tags + related_concepts + followup_question_ids 模拟关联
- v1 Tag-RAG 仅供 AI 出题/追问使用，**不做用户面经 ChatBot**
- v3 再做独立面经 ChatBot，复用 v1 的 pgvector 基建

---

## 错误码体系（完整 10 域）

| 域 | 范围 | 示例 |
|----|------|------|
| 通用 | 1xxx | `BAD_REQUEST(1400)` / `NOT_FOUND(1404)` / `UNAUTHORIZED(1401)` |
| 简历 | 2xxx | `RESUME_NOT_FOUND(2001)` / `RESUME_PARSE_FAILED(2002)` / `RESUME_DUPLICATE(2003)` |
| 面试 | 3xxx | `INTERVIEW_SESSION_NOT_FOUND(3001)` / `INTERVIEW_STAGE_INVALID(3002)` |
| 存储 | 4xxx | `STORAGE_UPLOAD_FAILED(4001)` / `STORAGE_NOT_FOUND(4002)` |
| 导出 | 5xxx | `EXPORT_PDF_FAILED(5001)` |
| 知识库 | 6xxx | `KNOWLEDGE_BASE_NOT_FOUND(6001)` / `VECTOR_INDEX_FAILED(6002)` |
| AI 服务 | 7xxx | `AI_SERVICE_TIMEOUT(7002)` / `AI_STRUCTURED_OUTPUT_FAILED(7003)` |
| 限流 | 8xxx | `RATE_LIMIT_EXCEEDED(8001)` |
| 面试日程 | 9xxx | `INTERVIEW_SCHEDULE_NOT_FOUND(9001)` |
| 语音面试 | 10xxx | `VOICE_SESSION_NOT_FOUND(10001)` / `WS_CONNECTION_FAILED(10002)` |

实现要求（v1）：

- [ ] 定义 `internal/errors/code.go`：ErrorCode 常量表
- [ ] 定义 `internal/errors/biz_error.go`：`type BizError struct { Code ErrorCode; Message string; Cause error }`
- [ ] 全局错误处理中间件：`BizError` → `Result.error(code, message)`
- [ ] 禁止使用 `errors.New()` / `fmt.Errorf()` 在 service / handler 层裸抛业务错误

---

## 禁止事项（写进 CLAUDE.md，编码必须遵守）

- 禁止在事务内调用 LLM / S3 / WebSocket（长事务占用 DB 连接）
- 禁止裸抛 `errors.New(...)` / `fmt.Errorf(...)` 作为业务错误，必须用 `BizError` + `ErrorCode`
- 禁止直接返回 Entity / domain 给前端，必须经 DTO 转换
- 禁止 LLM JSON 输出不走 `StructuredOutputInvoker`（无重试 = 必崩）
- 禁止硬编码 Provider Key / API Endpoint，统一走 `LlmProviderRegistry`
- 禁止 `log.Errorf("xx: %v", err.Error())` —— 必须把 err 作为最后一个 arg 保留堆栈
- 禁止全局 unbounded goroutine pool，必须用受限 worker pool 防 OOM
- 禁止跳过测试合并主分支
- 禁止提交 `.env` 文件
- 禁止使用 `hlog` 或 `slog`，统一走 `ai_interview/internal/log` 包

---

## 文档约定

- 各 v 文件**只列任务**，规范统一在本文件
- 已完成项 `[x]` 保留，归入 v1 对应模块
- 跨版本依赖在子文件开头「前置依赖」小节注明
- 任务粒度：单条 `[ ]` 应能在 0.5-2 天内完成
