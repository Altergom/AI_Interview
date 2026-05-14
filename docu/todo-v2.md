# todo-v2 — 好用：体验完善 + 数据闭环

> **目标**：体验对齐对标项目水平 + SFT 数据闭环跑起来 + 文字面试模式。
>
> **前置依赖**：v1 全部完成。
>
> **跨版本规范**（错误码、设计决策、禁止事项）见 [TODO.md](./TODO.md)。

---

## 模块清单

- [多 Provider 运行时切换](#多-provider-运行时切换)
- [Code Judge + Monaco Editor](#code-judge--monaco-editor)
- [文字面试模式](#文字面试模式)
- [死信队列 + 重试](#死信队列--重试)
- [SFT JSONL 导出 Worker](#sft-jsonl-导出-worker)
- [面试中心页 + 历史列表](#面试中心页--历史列表)
- [PDF 报告导出](#pdf-报告导出)
- [评估管道增强](#评估管道增强)
- [WebSocket 断线重连](#websocket-断线重连)
- [限流 GLOBAL 维度补全](#限流-global-维度补全)
- [视频采集与存储](#视频采集与存储)

---

## 多 Provider 运行时切换

> v1 是编译期静态配置（重启生效）。v2 上设置页 + 配置落盘。

- [ ] `LlmProviderRegistry` 改造：`atomic.Pointer[ProviderMap]` 支持热更新
- [ ] 配置落盘：`~/.ai-interview/llm-providers.yml`
  - API Key 用对称加密（AES-GCM），密钥从 `APP_AI_CONFIG_ENCRYPTION_KEY` ENV 读
- [ ] **Reload 接口**：`POST /v1/admin/llm-provider/reload`，从 yml 重新加载
- [ ] **测试连接**：`POST /v1/admin/llm-provider/test`
  - 验证 chat / embedding / ASR / TTS 四类
  - 返回延迟 + 错误信息
- [ ] **默认 Provider 切换**：`POST /v1/admin/settings`
  - 写 `chat_default` / `embedding_default` / `asr_default` / `tts_default`
- [ ] 设置页 UI：
  - Provider 列表（DashScope / Kimi / DeepSeek / GLM / LM Studio）
  - 测试按钮 + 默认切换 + API Key 输入
  - 切换默认模型不需要改代码 / 重启
- [ ] Admin 端点鉴权：仅 `is_admin=true` 的用户可访问
- [ ] 错误码补充：`AI_PROVIDER_TEST_FAILED(7004)` / `AI_PROVIDER_NOT_FOUND(7005)`

---

## Code Judge + Monaco Editor

> v1 阶段状态机不含 `algorithm`，v2 加上。

- [x] `CodeJudgeResult` domain 类型
- [ ] **Code Judge Agent** 实现
  - 输入：代码文本 + 题目信息
  - 输出：`{correctness, time_complexity, space_complexity, issues}`
  - 必须经 `StructuredOutputInvoker`
- [ ] 阶段状态机扩展：`questioning → algorithm → closing`
- [ ] WebSocket 上行新增 `code_submit` 消息处理
- [ ] Code Judge 结果回注 Interview Agent 驱动追问方向：
  - `correctness=true` → 追问复杂度优化
  - `correctness=false` → 引导用户找 bug
- [ ] 前端 Monaco Editor 集成
- [ ] 支持语言切换（Java / Python / Go / C++）
- [ ] 算法阶段显示编辑器，其他阶段隐藏
- [ ] 提交按钮：构造 `code_submit` WS 消息
- [ ] 限流：每分钟最多提交 5 次

---

## 文字面试模式

> 复用语音面试的 Agent 链路，仅替换输入输出层。

- [ ] WebSocket 频道扩展：`mode=text` 不走 ASR/TTS
- [ ] EvaluationPipeline **文字输入分支实现**（v1 已留接口）
- [ ] 文字输入面板（前端）：替代麦克风按钮
- [ ] 流式 LLM token 直接渲染（不经 TTS）
- [ ] 文字 / 语音可对比评估（同一 prompt 不同输入）
- [ ] 面试创建时选择模式：`voice` / `text`

---

## 死信队列 + 重试

> v1 失败直接标 failed。v2 加 DLX + 重试按钮。

- [ ] RabbitMQ DLX 配置（dead-letter-exchange + queue）
- [ ] 失败计数 ≥ 3 次入死信队列
- [ ] DB 状态机：`failed` 状态独立于 `processing`，记录 `error_message` + `retry_count`
- [ ] **重试 API**：`POST /v1/admin/task/{id}/retry`
- [ ] 前端「重试」按钮：报告生成失败 / 简历分析失败时显示
- [ ] 死信告警：日志 + Prometheus（v3 接监控时填充）
- [ ] 重试限流：每个任务最多手动重试 3 次

---

## SFT JSONL 导出 Worker

> 我们的差异化卖点：闭环数据飞轮。

- [x] `SFTMessage` domain 类型
- [ ] `cmd/sft_exporter` 定期 Job（cron 每天凌晨 2 点）
- [ ] 筛选 `quality=good` / `quality=bad` 标注的对话轮次
- [ ] 多轮 conversation 格式化：

```json
{
  "messages": [
    {"role": "system", "content": "你是一个技术面试官..."},
    {"role": "assistant", "content": "面试官的问题"},
    {"role": "user", "content": "用户的回答"},
    {"role": "assistant", "content": "面试官的追问"}
  ],
  "quality": "good",
  "user_feedback": "追问很有启发"
}
```

- [ ] 上传 S3 路径：`/sft/{date}/good.jsonl`、`/sft/{date}/bad.jsonl`
- [ ] 标记已导出（避免重复）：`questionnaire_results` 加 `exported_at` 列
- [ ] JSONL 文件大小阈值切分（单文件最大 100MB）
- [ ] 导出统计：每次 Job 输出 INFO 日志（导出条数 / 文件大小 / S3 路径）

---

## 面试中心页 + 历史列表

- [ ] **面试中心页**：整合文字 / 语音入口
- [ ] **历史面试列表**：状态、时长、得分、删除
- [ ] **详情页**：完整对话记录 + 报告 + 问卷标注
- [ ] **继续面试** / **重新面试** 按钮
- [ ] 列表分页 + 搜索（按公司 / 岗位 / 时间过滤）
- [ ] 删除面试：软删除（保留数据用于 SFT 导出）

---

## PDF 报告导出

> v1 是网页版报告，v2 加 PDF。

- [ ] 选型：`gopdf` 或 `unidoc`，含中文字体支持
- [ ] 字体：内置开源中文字体（避免 GPL 风险，用 `ZhuqueFangsong-Regular.ttf` 类似的）
- [ ] 异步生成：MQ 消息 → Worker → 生成 PDF → 上传 S3
- [ ] 错误码：`EXPORT_PDF_FAILED(5001)`
- [ ] 前端「导出 PDF」按钮 → 轮询任务 → 下载链接
- [ ] PDF 内容：基本信息 + 雷达图 + 维度详情 + 对话摘要 + 优劣势总结

---

## 评估管道增强

> v1 是简单平均，v2 完整版。

- [ ] **分批评估**：每批 8 turn 完整实现（已有骨架）
- [ ] **二次汇总**：跨批分数加权 + 维度归一
- [ ] **降级兜底**：LLM 失败返回基于规则的中等分数（含错误说明）
- [ ] 多维度评分（已有 Report 字段全部填充）
  - knowledge_depth / expression / problem_solving / code_quality / stress_response
- [ ] `weak_points` / `strong_points` JSONB 数组生成
- [ ] 评估并发控制：单次面试最多 3 个评估请求并行

---

## WebSocket 断线重连

- [ ] 前端：`last_message_id` 记录到 sessionStorage
- [ ] 重连后从断点恢复：服务端按 `last_message_id` 重发未确认消息
- [ ] 心跳检测：30s ping/pong，超时关闭
- [ ] 前端断线提示 + 自动重连（指数退避，最多 5 次）
- [ ] 服务端清理：连接断开 5 分钟未恢复 → 自动暂停面试

---

## 限流 GLOBAL 维度补全

> v1 只有 IP + USER。v2 补全。

- [ ] `GLOBAL` 维度限流：保护后端兜底（如 100/min 全局调用）
- [ ] 限流注解支持多维度叠加（任一触发即拒绝）
- [ ] 限流监控埋点（v3 监控接入时填充）

---

## 视频采集与存储

> v1 只采集音频，v2 加视频。

- [ ] 前端 `getUserMedia` 采集摄像头视频流
- [ ] 视频实时传输（或录制后上传 S3）
- [ ] S3 路径：`/video/{interview_id}/full.mp4`
- [ ] 视频存储为可选（用户可关闭）

---

## v2 完成标准

- [ ] 设置页可热切换默认 Provider，重启不丢
- [ ] 文字面试可用，评估对比语音版无差异
- [ ] Code Judge Agent 在算法阶段触发，结果驱动追问
- [ ] SFT JSONL 每天定时导出到 S3，good/bad 分文件
- [ ] 失败任务可在前端点「重试」恢复
- [ ] PDF 报告中文显示正常
- [ ] WebSocket 断线 30s 内重连可恢复面试
- [ ] 首包延迟 < 400ms（对标 interview-guide）
