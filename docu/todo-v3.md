# todo-v3 — 全产品 + 运营能力

> **目标**：从面试工具变成完整产品。招聘场景 + 用户面经 ChatBot + 监控 + 运营后台。
>
> **前置依赖**：v1 + v2 全部完成。
>
> **跨版本规范**（错误码、设计决策、禁止事项）见 [TODO.md](./TODO.md)。

---

## 模块清单

- [面试日程日历（AI 解析邀请）](#面试日程日历ai-解析邀请)
- [用户面经 ChatBot](#用户面经-chatbot)
- [知识库管理后台](#知识库管理后台)
- [题库管理 CMS](#题库管理-cms)
- [OpenTelemetry 埋点](#opentelemetry-埋点)
- [testcontainers 集成测试](#testcontainers-集成测试)
- [运营数据看板](#运营数据看板)

---

## 面试日程日历（AI 解析邀请）

> 对标 interview-guide 的 `interviewschedule` 模块。招聘场景刚需。

- [ ] **邀请文本解析**：规则 + LLM 双引擎
  - 飞书 / 腾讯会议 / Zoom 邀请格式识别
  - 提取字段：公司、岗位、时间、会议链接、面试官
  - 必须经 `StructuredOutputInvoker`
- [ ] PostgreSQL `interview_schedules` 表
  - id / user_id / company / position / scheduled_at / meeting_url / status / source_text
- [ ] CRUD API：`/v1/schedule/parse` / `/v1/schedule` (POST/GET/PATCH/DELETE)
- [ ] 错误码：`INTERVIEW_SCHEDULE_NOT_FOUND(9001)` / `SCHEDULE_PARSE_FAILED(9002)`
- [ ] 日历组件（前端）：日 / 周 / 月 / 列表视图
- [ ] 拖拽调整时间（`react-big-calendar` 或类似）
- [ ] 状态流转：`upcoming` / `completed` / `cancelled`
- [ ] **定时任务**：每小时自动把过期面试标 `completed`
- [ ] 面试提醒：浏览器通知（可配置提前时长）
- [ ] 限流：邀请解析端点防刷

---

## 用户面经 ChatBot

> 复用 v1 的 pgvector 基建。**独立模块**，跟核心面试解耦。

### 知识库管理（用户侧）

- [ ] PostgreSQL `knowledge_bases` 表（用户的私有 KB）
  - id / user_id / name / description / status / created_at
- [ ] 文档表 `kb_documents`：`kb_id / file_name / mime_type / s3_key / chunk_count / status`
- [ ] 文档分块表 `kb_chunks`：`document_id / content / embedding (vector 1024) / chunk_index`
- [ ] 文件上传 API（PDF / DOCX / Markdown / TXT）
- [ ] 异步分块 + 向量化（RabbitMQ `kb_vectorize` 队列 + Worker，复用 v1 基建）
- [ ] 错误码：`KNOWLEDGE_BASE_NOT_FOUND(6001)` / `VECTOR_INDEX_FAILED(6002)`

### RAG 问答

- [ ] **查询改写**（先让 LLM 把用户提问优化成检索 query）
- [ ] 相似度阈值 + TopK 策略（默认 K=5，阈值 0.6）
- [ ] **SSE 流式问答** + 打字机效果
- [ ] 多 KB 关联（同时检索多个知识库）
- [ ] 引用来源回显（点击可跳到原文片段）

### 会话管理

- [ ] 表 `chat_sessions` / `chat_messages`
- [ ] 会话置顶 / 重命名 / 删除
- [ ] **虚拟列表渲染优化**（`react-virtuoso`，对标他们的优化）
- [ ] Markdown 渲染（代码高亮 + 数学公式）
- [ ] 限流：用户级别 60/小时

---

## 知识库管理后台

- [ ] 分类管理（标签层级树）
- [ ] 知识库下载（导出原始文档 + 分块）
- [ ] 重新向量化（embedding 模型升级时批量重处理）
- [ ] 全文搜索（对内容 / 标签 / 文件名）
- [ ] 统计信息：总文档数 / 总分块数 / 各分类分布
- [ ] 分享功能：把私有 KB 复制为公共模板

---

## 题库管理 CMS

> v1 题库种子是脚本导入，v3 上后台 CRUD。

- [ ] 题目 CRUD 界面（管理员用）
- [ ] 标签管理：层级标签 + 自动建议
- [ ] **关联推荐编辑**：可视化编辑 `related_concepts` + `followup_question_ids`
- [ ] 重新向量化批量任务（题目内容修改后自动触发）
- [ ] 题目质量打分（基于历史使用率 + 用户反馈）
- [ ] 导入 / 导出 JSON 批量操作
- [ ] 权限：仅管理员（`is_admin=true`）可访问

---

## OpenTelemetry 埋点

- [ ] Hertz 中间件接入 OTel（`go.opentelemetry.io/otel`）
- [ ] 埋点关键路径：
  - LLM 调用（model / latency / token_count / error）
  - DB 查询（query / duration）
  - Redis 操作（command / latency / hit_miss）
  - WebSocket 连接（connect / disconnect / message_count）
  - ASR / TTS 延迟
- [ ] **Prometheus exporter** + `/metrics` 端点
- [ ] **业务指标**：
  - 会话创建成功率
  - 平均会话时长
  - ASR 延迟 P50/P95/P99
  - TTS 首包延迟 P50/P95/P99
  - LLM 响应时间
  - 报告生成成功率
- [ ] Grafana dashboard 模板（`docu/grafana/`）
- [ ] 告警规则（Alertmanager）：错误率 > 1% / P99 延迟 > 2s / 死信队列堆积

---

## testcontainers 集成测试

> v1/v2 主要靠单元测试 + miniredis。v3 补完整集成测试。

- [ ] 引入 `testcontainers-go`
- [ ] 自动起依赖：PostgreSQL（含 pgvector）/ Redis / RabbitMQ / MinIO
- [ ] 关键 Service 集成测试覆盖：
  - 简历解析（PDF → LLM → DB）
  - 面试创建到结束完整链路
  - Tag-RAG 检索
  - 评估管道
  - SFT 导出
- [ ] CI 跑全链路测试（GitHub Actions / 公司 CI）
- [ ] 测试覆盖率目标：service 层 ≥ 80%
- [ ] mock LLM Provider（避免 CI 真调外部 API）

---

## 运营数据看板

- [ ] 管理员后台首页：总用户数 / 总面试数 / 总报告数 / 实时在线
- [ ] 用户行为分析：注册转化、面试完成率、平均评分
- [ ] LLM 成本统计：按 Provider / 模型 / 用户分组
- [ ] **SFT 数据看板**：good/bad 标注比例 / 月度增长曲线 / 导出文件统计
- [ ] 异常监控：错误码分布 / 限流触发次数 / 死信任务数

---

## v3 完成标准

- [ ] 用户能粘贴飞书面试邀请，自动出现在日历
- [ ] 用户面经 ChatBot 可上传文档 + 多轮对话 + 流式回答
- [ ] 题库管理员可在 CMS 直接编辑题目和关联关系
- [ ] Grafana 看板能监控 ASR/TTS/LLM 关键指标
- [ ] 集成测试覆盖率 ≥ 80%，CI 全绿
- [ ] 运营看板能展示 SFT 数据增长
