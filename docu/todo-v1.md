# todo-v1 — MVP：跑通一次完整面试

> **目标**：游客可以完成一次完整闭环 —— 上传简历 → AI 出题 → 语音面试 → 拿到报告 → 提交问卷标注。
>
> **跨版本规范**（错误码、设计决策、禁止事项）见 [TODO.md](./TODO.md)。

---

## 模块清单

- [基础设施](#基础设施)
- [通用能力](#通用能力)
- [用户与认证](#用户与认证)
- [简历模块](#简历模块)
- [Skill 出题模块](#skill-出题模块)
- [Tag-RAG 题库（仅供 AI）](#tag-rag-题库仅供-ai)
- [WebSocket 通信骨架](#websocket-通信骨架)
- [语音面试（替换现有 Mock）](#语音面试替换现有-mock)
- [Interview Agent](#interview-agent)
- [评估引擎接口](#评估引擎接口)
- [Report Worker（基础版）](#report-worker基础版)
- [问卷标注（SFT 数据采集起点）](#问卷标注sft-数据采集起点)
- [前端基础页面](#前端基础页面)

---

## 基础设施

- [x] 编写 Docker Compose，本地一键启动所有服务（PostgreSQL / Redis / S3 / RabbitMQ）
- [x] 配置各服务环境变量（`.env` + `godotenv` + `internal/config` 集中加载）
- [x] 编写各服务 Dockerfile（根目录 `Dockerfile` 构建 `cmd/api`）
- [x] 配置日志采集（`LOG_FORMAT=json|text` + `slog` 输出 stdout）
- [x] 配置健康检查端点（`GET /health`、`/ready` 及别名）
- [x] 选型并初始化 RabbitMQ（`docker-compose`）
- [x] 定义 Topic: `interview_finished`（`internal/mq` 常量与事件体）
- [x] 初始化数据库，创建基础表（users / interviews / interview_turns / reports / questionnaire_results）
- [x] 编写 migration 脚本
- [x] 定义 Redis Key 与命名方法（`internal/storage/redis/keys`）
- [x] 配置合理 TTL（`RESUME_REDIS_TTL` / `INTERVIEW_STATE_TTL`）
- [x] 定义业务对象路径（简历 / 音频 / 视频 / SFT 前缀）
- [x] **Milvus 初始化**：Docker Compose 加 Milvus standalone + etcd + minio（或复用已有 MinIO）
- [x] 创建 Milvus 集合 `bank_questions_vec`（1024 维，COSINE，IVF_FLAT，nlist=128）
- [x] **Elasticsearch 初始化**：Docker Compose 加 ES 单节点（`elasticsearch:8.x`）
- [x] 创建 ES 索引 `bank_questions`（mapping：`question` text + `tags` keyword + `difficulty` keyword）
- [x] migrations 整理：新增 `bank_questions` 元数据表（不含向量列，仅结构化字段）
- [x] PostgreSQL 连接池配置
- [x] Redis 连接初始化
- [x] S3 客户端配置 + Bucket 权限策略
- [x] LlmProviderRegistry 骨架（v1 仅编译期静态配置，从 `.env` 读）

---

## 通用能力

- [x] 错误码 10 域定义（`internal/errors/code.go`）
- [x] `BizError` 结构 + 全局错误处理中间件（`internal/errors/biz_error.go`）
- [x] 统一响应格式 `Result<T>`（`{ success, data, error }`）
- [x] **StructuredOutputInvoker**：JSON 输出最多重试 3 次 + 降级（`internal/einocore/structured_output.go`）
- [x] **限流中间件** + Redis Lua 滑窗脚本（`internal/middleware/ratelimit/`）
  - 维度：`IP`（10/min）+ `USER`（30/min）
  - Key 设计：`ratelimit:{handler}:{dimension}:{value}`
- [x] **异步任务状态机**：报告完成由 WebSocket 推送；`reports` 表加 `error_message` 标记失败；状态由 RabbitMQ 投递保证
- [x] 任务实体预校验工具（处理前查实体存在，被删则 ACK 丢弃）
- [x] 全局日志中间件（注入 request_id、user_id 到 context）

---

## 用户与认证

- [x] PostgreSQL `users` 表实现（已在 `001_init.sql`）
- [x] JWT 工具（`internal/auth/jwt.go`）：`GenerateToken` / `ValidateToken`，secret 从 `JWT_SECRET` 读
- [x] bcrypt 密码加密（`internal/auth/password.go`）
- [x] 注册 API：检查邮箱 → 加密 → 插 PG → 签发 JWT
- [x] 登录 API：查用户 → 验密码 → 签发 JWT
- [x] 游客模式：`guest_` 前缀 ID + `is_guest=true` + 24h JWT
- [x] HAuth 中间件：JWT 鉴权（支持游客 token）
- [x] 路由初始化（Hertz）

---

## 简历模块

- [x] 定义结构化简历领域模型（`StructuredResume` 等）
- [x] S3 预签名 URL 生成（`internal/service/resume_impl.go`），5 分钟有效
- [x] PDF 文本提取（`pdfcpu` 库）+ 流式处理避免 OOM
- [x] LLM 结构化解析（**必须经 `StructuredOutputInvoker`**）
- [x] **降级策略**：LLM 失败返回空结构体 → 前端显示空表单让用户手填
- [x] **去重**：`resumes` 表加 `content_hash` 列（SHA-256），命中直接返回
- [x] PostgreSQL 主存储 + Redis 1h 缓存（按需回填）
- [x] PDF 备份原始文件到 S3
- [x] 文件上传并发限制（信号量 5 并发）+ 大小限制 3MB
- [x] API：`POST /v1/resume/parse` / `POST /v1/resume/submit` / `GET /v1/resume`
- [x] 限流：`/v1/resume/parse` 接 IP+USER 维度

---

## Skill 出题模块

> 对标 interview-guide：`SKILL.md` 文件驱动出题，比 RAG 题库维护成本低。

- [ ] 创建目录 `internal/skills/{direction}/SKILL.md`，v1 至少覆盖 5 个方向：
  - `go-backend` / `java-backend` / `frontend` / `algorithm` / `ai-agent`
- [ ] 每个 SKILL.md 包含：考察范围、难度分布、追问策略、引用资料路径
- [ ] `SkillLoader`：启动加载 + 文件 hot-reload（不重启服务可迭代提示词）
- [ ] 历史题目跨 turn 去重：Redis Set `interview:{id}:asked_questions`

---

## Tag-RAG 题库（仅供 AI）

> v1 用户感知不到 RAG 存在，Milvus+ES 多路召回只是后台基建。用户面经 ChatBot 推到 v3。

- [x] 定义 `BankQuestion` 领域类型
- [ ] 题库元数据表 `bank_questions`（PgSQL）：`question` / `standard_answer` / `tags` (jsonb) / `related_concepts` (jsonb) / `followup_question_ids` (jsonb) / `vec_status` (pending/done/failed)
- [ ] embedding 服务接入（默认 DashScope `text-embedding-v3`，走 `LlmProviderRegistry`）
- [ ] 异步写入 Worker：题目入库后消费 RabbitMQ `vectorize_task` 队列
  - 写 Milvus：`bank_questions_vec` 集合，field `question_id`（varchar）+ `embedding`（FloatVector 1024）
  - 写 ES：索引 `bank_questions`，字段 `question_text` / `tags` / `difficulty`
  - 成功后更新 PgSQL `vec_status=done`
- [ ] **多路召回检索接口**：技能标签 + Skill 配置 → Top-K 题目
  - **向量召回**：Milvus ANN 搜索，nprobe=16，返回 Top-20
  - **关键词/标签召回**：ES bool query（tags filter + question match），返回 Top-20
  - **RRF 融合**：`score = Σ 1/(k + rank_i)`，k=60，取融合 Top-K
- [ ] 题库种子脚本：导入初始 50-100 道题（覆盖 5 个方向）

---

## WebSocket 通信骨架

> 替换原 SSE 设计。前端流式 PCM，服务端流式 ASR/LLM/TTS。

- [ ] 后端 Hertz WS 端点 `GET /v1/interview/ws/{interview_id}`
- [ ] **上行消息**：`audio_chunk` (PCM bytes) / `control` (start/pause/resume/stop) / `code_submit`
- [ ] **下行消息**：`asr_partial` / `asr_final` / `llm_token` / `tts_audio` / `stage_change` / `error`
- [ ] 前端 AudioWorklet 实时采集 PCM (16kHz / 16bit / mono) 流式发送
- [ ] 客户端 AEC：`echoCancellation: true` + `noiseSuppression: true` + `autoGainControl: true`
- [ ] 服务端推送 PCM 用 Web Audio API 流式播放（边收边播）
- [ ] UI 提示「建议佩戴耳机」（在准备页 + 面试中）
- [ ] WebSocket 鉴权：握手时校验 JWT
- [ ] 限流：建立连接维度限流（IP+USER）

---

## 语音面试（替换现有 Mock）

> **删除** `qwen_asr.go` / `qwen_tts.go` 现有 HTTP 实现，重写为 WebSocket。

- [ ] Qwen3 实时 ASR：`wss://dashscope.aliyuncs.com/api-ws/v1/realtime`
  - 模型 `qwen3-asr-flash-realtime`
  - server VAD（400ms 静音阈值）
  - 实时中间结果 + 最终结果
  - 引入 `github.com/gorilla/websocket` 或 `nhooyr.io/websocket`
- [ ] Qwen3 实时 TTS：模型 `qwen3-tts-flash-realtime`
  - 句子级并发 TTS（边生成边合成边播放）
  - PCM 16kHz 输出
  - 默认音色 Cherry
- [ ] **删除** 现有 `qwen_asr.go` / `qwen_tts.go` HTTP 实现
- [ ] 保留 Mock 服务用于单元测试（`tools/mock.go`）
- [ ] 错误码：`WS_CONNECTION_FAILED(10002)` / `AI_SERVICE_TIMEOUT(7002)`

---

## Interview Agent

- [x] Eino 依赖、`compose` 恒等链占位、SFT→`schema.Message` 桥接
- [x] Supervisor / stage_manager / question_selector / response_analyzer 骨架
- [ ] LLM 流式输出（首 token 即触发 TTS）
- [ ] 多轮对话 history：Redis List 存 + 超长裁剪（保留 system + 最近 N 轮）
- [ ] 阶段切换状态机：`intro → questioning → closing → end`（algorithm 阶段在 v2）
- [ ] 题目状态机：出题 → 等答 → 追问（< N）→ 关闭 → RAG 下题
- [ ] 简历上下文注入到 system prompt
- [ ] 设计提问阶段 system prompt（面试官角色、追问参考 `followup_question_ids`）
- [ ] 设计反问阶段 system prompt（候选人反问公司）
- [ ] **Router**：从 Redis 读状态 → 阶段判断 → 路由到对应 Agent
- [ ] **信息提取 Agent**：自我介绍补充提取 → merge 进 Redis 上下文
- [ ] Redis 状态读写封装（`internal/storage/redis`）
  - 面试状态、对话 history、面试配置
- [ ] 面试状态结构补 `report_status` 字段：

```json
{
  "interview_id": "xxx",
  "stage": "questioning",
  "questions_asked": 3,
  "current_question_followups": 1,
  "report_status": "pending",
  "started_at": "2024-01-01T10:00:00Z"
}
```

---

## 评估引擎接口

> 从 day 1 抽象统一管道，避免后期重构。v1 只实现语音输入分支。

- [x] 定义 `Report` / `ReportDimensions` 领域类型
- [ ] 定义 `EvaluationPipeline` 接口（`internal/einocore/evaluation/pipeline.go`）
- [ ] 定义 `Turn` 输入结构（同时支持 `text` + `audio_transcript`）
- [ ] 实现 `BatchScore`：分批评估（每批 8 turn），降低单次 LLM 上下文压力
- [ ] 实现 `Aggregate` 骨架：跨批分数加权（v1 简单平均，v2 增强）
- [ ] 实现 `Fallback`：LLM 失败返回保底报告（基于规则给中等分数 + 错误说明）
- [ ] 所有 JSON 输出必须经 `StructuredOutputInvoker`
- [ ] v1 实现语音输入分支（文字分支留接口给 v2）

---

## Report Worker（基础版）

- [x] `mq.InterviewFinished` 事件定义
- [ ] RabbitMQ Producer：面试结束发布消息
- [ ] Worker Consumer：消费 → 拉对话记录 → 调评估管道 → 写 PG
- [ ] 任务状态机：`pending → processing → completed / failed`
- [ ] 失败重试 3 次（v1 失败直接 DB 标 failed，**死信队列推 v2**）
- [ ] 报告生成完成后通过 WebSocket 推送通知前端
- [ ] Record Service：`SaveTurn` / `GetInterviewRecord`
- [ ] 音频文件上传 S3：`/audio/{interview_id}/{turn_id}.wav`

---

## 问卷标注（SFT 数据采集起点）

> 我们的差异化卖点。v1 只采集，定期 JSONL 导出推到 v2。

- [x] 定义问卷标注与 SFT/JSONL 行结构（`QuestionnaireResult`、`SFTMessage`）
- [x] `questionnaire_results` 表 schema
- [ ] 提交 API `POST /v1/questionnaire/submit`
- [ ] 接收逐轮标注（quality: good/bad + feedback 文字）
- [ ] **采集策略：good/bad 都采集**（DPO 负样本备用）
- [ ] **数据形态：多轮 conversation**（保存到 PG，定期导出在 v2）
- [ ] 前端问卷页：逐轮展示 ASR 文本 + 打标 good/bad + 反馈输入

---

## 前端基础页面

- [ ] Index 首页：「登录/注册」+ 「游客体验」入口
- [ ] 登录 / 注册页
- [ ] 简历信息页：表单填写 + PDF 上传自动填充 + 解析失败显示空表单
- [ ] 岗位选择页（Go / Java / 前端 / 算法 / AI Agent）
- [ ] 方向选择页（软件开发 / 云平台运维 / Agent 开发等）
- [ ] 准备页：麦克风测试 + AEC 启用提示 + 「建议佩戴耳机」横幅
- [ ] 设备权限异常处理（拒绝授权、设备不存在）
- [ ] 面试间页：左侧对话记录 + 顶部阶段进度（**WebSocket 客户端替代 SSE**）
- [ ] 报告生成等待页（轮询 `report_status`）
- [ ] 报告页：网页版多维度雷达图 + 优劣势总结（**PDF 导出推 v2**）
- [ ] 问卷页：逐轮 good/bad + 反馈
- [ ] 结束页：「再来一次」+ 报告链接
- [ ] 雷达图组件（recharts 或 echarts）
- [ ] 全局 Toast：限流 / 错误码统一展示

---

## v1 完成标准

- [ ] 一个游客能从首页走到结束页，不报错
- [ ] AI 能根据简历 + Skill 出题，至少进行 5 轮有效追问
- [ ] 语音面试首包延迟 < 1s（v2 优化到 < 400ms）
- [ ] 面试结束后 30s 内能拿到报告
- [ ] 问卷数据正确入库
- [ ] 关键端点限流生效，超限返回 8001 错误码
- [ ] 简历重复上传命中去重直接返回
