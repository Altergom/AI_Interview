# AI Interview

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react&logoColor=white)](https://react.dev)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

基于字节三件套（Hertz + Eino）构建的 AI 模拟面试系统。支持语音/文本双模态交互，多阶段面试流程由 Agent 或 Workflow 驱动。

# 依赖

## 基础设施

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.25+ | 后端运行时 |
| Node.js | 18+ | 前端构建 |
| PostgreSQL | 16+ | 业务数据持久化 |
| Redis | 7+ | 会话状态、缓存 |
| MinIO / S3 | — | 简历、音频文件存储 |
| Milvus | 2.4+ | 题库向量检索 |
| Elasticsearch | 8.13+ | 题库关键词/标签检索 |
| RabbitMQ | 3.13+（可选） | 面试完成事件、报告生成 |
| Ollama | 可选 | 本地 LLM / Embedding |

## LLM 提供商（至少配一个）

OpenAI · Qwen（千问）· Doubao（豆包）· Claude · DeepSeek · Gemini

> Qwen 同时提供 ASR（语音识别）和 TTS（语音合成）能力，推荐优先配置。

## 后端核心库

| 库 | 用途 |
|----|------|
| `cloudwego/hertz` | HTTP / WebSocket 框架 |
| `cloudwego/eino` | Agent / Workflow 编排（ADK） |
| `gorm.io/gorm` | ORM |
| `redis/go-redis/v9` | Redis 客户端 |
| `milvus-io/milvus-sdk-go/v2` | 向量数据库客户端 |
| `rabbitmq/amqp091-go` | 消息队列 |
| `golang-jwt/jwt/v5` | JWT 认证 |

## 前端技术栈

React 19 · Vite · TypeScript · Zustand · Monaco Editor · Recharts

# 快速启动

## 1. 拉起基础设施

```bash
docker-compose up -d
```

一键启动 PostgreSQL、Redis、MinIO、Milvus、Elasticsearch、RabbitMQ。

## 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env，至少填写一个 LLM 提供商的 API Key
```

## 3. 启动后端

```bash
go run ./cmd
```

默认监听 `:8080`，启动时自动执行数据库迁移。

## 4. 启动前端

```bash
cd frontend && npm install && npm run dev
```

开发服务器运行在 `http://localhost:5173`。

# 执行流程

```
用户 (语音/文本)
     │
     ▼
┌──────────────────────────────────────────────┐
│  Hertz Gateway (:8080)                       │
│  JWT 认证 → 限流 → WebSocket / HTTP 路由     │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────┐
│  Interview Service                           │
│  SessionManager 管理面试状态 (Redis)          │
└──────────────────┬───────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────┐
│  Eino Interview Graph                        │
│                                              │
│  模式 ①  Workflow 驱动（WORKFLOW_ENABLED=1）  │
│    StageRouter → 阶段Agent → StageManager    │
│    → TransitionNode（状态机裁决）             │
│                                              │
│  模式 ②  Agent 驱动                          │
│    Supervisor 统一调度子 Agent                │
└──────┬───────────┬───────────┬───────────────┘
       │           │           │
       ▼           ▼           ▼
  ┌────────┐ ┌─────────┐ ┌────────┐
  │ASR/TTS │ │RAG 检索  │ │ 题库   │
  │ (Qwen) │ │Milvus+ES│ │(PgSQL) │
  └────────┘ └─────────┘ └────────┘
```

**面试阶段流转：** `intro` → `questioning` → `algorithm` → `closing` → `end`

1. 用户通过 WebSocket 发送语音或文本
2. SessionManager 从 Redis 加载面试状态（阶段、历史、上下文）
3. Interview Graph 按当前阶段路由到对应 Agent
4. Agent 调用工具：ASR 转写、RAG 检索题库、TTS 合成语音
5. StageManager 评估是否切换阶段，状态机执行转换
6. 面试结束后 MQ 触发 ReportWorker 生成评估报告

# 联系

- Email：kele3325@gmail.com
- 微信群：扫码加入交流群

<img src="docu/wechat_group.png" alt="微信交流群" width="300" />
