# 李明
**AI Agent 工程师 / 3 年经验**

138-XXXX-XXXX  |  liming.ai.dev@example.com  |  上海

---

## 教育背景

- **华东师范大学**  计算机科学与技术  硕士  *2020.09 – 2023.06*
- **浙江工业大学**  软件工程  本科  *2016.09 – 2020.06*

## 工作经验

**某科技有限公司 — AI Agent 工程师**  *2023.07 – 至今*

- 主导对话式 Agent 平台核心链路设计，基于 LangChain + Eino 构建多 Agent 编排框架，支持 8 个业务场景上线，日活 12 万。
- 实现基于 Milvus + Elasticsearch 双路召回的 RAG 系统，命中率较纯向量召回提升 18%，P95 延迟 < 400ms。
- 推动 LLM 输出结构化稳定性专项：自研 StructuredOutputInvoker，JSON 解析失败率从 4.2% 降至 0.3%（3 次重试 + 降级）。
- 负责 SFT 数据采集 → 标注 → JSONL 导出管线，累计 SFT 样本 8 万条，DPO 负样本占比 35%。

**某 AI 初创 — LLM 应用开发实习生**  *2022.07 – 2023.06*

- 基于 OpenAI Function Calling 实现智能客服 Agent，工单解决率从 41% 提升到 68%。
- 编写 RAG 评测脚本，使用 ragas 框架对 5 套召回方案做基准测试。

## 项目经验

**多 Agent 模拟面试系统（开源）**  *2024.06 – 2024.12*

- 完整复刻招聘面试链路：Supervisor → stage_manager → question_selector → response_analyzer 四 Agent 协同。
- 技术栈：Go 1.22 + Hertz + Eino 框架；前端 React 19 + Vite。
- 实现实时语音通道：Qwen3-ASR 流式识别 + Qwen3-TTS 句子级并发合成，端到端首包延迟 < 1s。
- 落地 SFT 问卷标注模块：good/bad 双向采集 + UPSERT 幂等 + JSONL 导出，用于后续模型微调。

**智能简历筛选 Agent**  *2023.10 – 2024.03*

- 基于 LLM 结构化输出 + 规则引擎，自动从 PDF 简历提取 30 个字段。
- 解析准确率 92%，相比 GPT-3.5 基线提升 11 个百分点。

## 技能清单

- **LLM**：Prompt 设计、Function Calling、结构化输出、RAG、Embedding
- **框架**：LangChain / Eino / LlamaIndex / DSPy
- **向量库**：Milvus / Qdrant / pgvector
- **后端**：Python (FastAPI) / Go (Hertz, Gin)
- **其他**：Docker、Kubernetes、PostgreSQL、Redis、RabbitMQ
- **英语**：CET-6 / 可阅读英文技术文档与论文
