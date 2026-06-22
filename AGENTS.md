# AI Interview - Agent Guide

> 面向 AI 编程助手的项目规范，与 CLAUDE.md 保持同步。

---

## ⚠️ 最高优先级规则

- 修改代码前，先说明改动范围和思路，等开发者确认再执行
- 新增依赖前，先告知包名和用途，等确认再安装
- 遇到多种实现方案时，列出优劣后由开发者决策，不要自行选择
- 不要自动重构与当前任务无关的代码
- 每次对话结束时，在末尾列出本次所有修改过的文件路径

---

## 项目概览

- **类型**：全栈项目（前端 + 后端）
- **前端**：React + TypeScript + Vite，状态管理用 Zustand
- **后端**：Go，HTTP 框架 Hertz，AI 编排框架 Eino
- **数据库**：PostgreSQL + Redis + Milvus（向量库）+ Elasticsearch
- **对象存储**：S3 兼容 API
- **包管理**：前端 npm，后端 go mod

## 目录结构

```
project-root/
├── frontend/src/
│   ├── components/     # 可复用组件
│   ├── pages/          # 页面/路由
│   ├── hooks/          # 自定义 hooks
│   ├── services/       # API 调用层
│   ├── store/          # Zustand 状态
│   ├── types/          # TypeScript 类型
│   └── utils/          # 工具函数
├── internal/
│   ├── handler/        # HTTP 路由层（仅参数校验 + 调用 service）
│   ├── router/         # 路由注册层（集中记录并注册所有 HTTP/WS 路由）
│   ├── service/        # 业务逻辑层
│   ├── domain/         # 领域模型（所有共享类型定义在此）
│   ├── storage/        # 存储层（postgres / redis / s3 / milvus / es）
│   ├── einocore/       # AI 编排（Eino Graph、Agent、Tools）
│   ├── middleware/     # HTTP 中间件
│   ├── auth/           # JWT 鉴权
│   ├── config/         # 配置加载
│   └── log/            # 统一日志包
├── cmd/main.go
└── CLAUDE.md
```

---

## 代码规范

### 通用
- 函数体不超过 80 行 / 60 语句，圈复杂度不超过 15，参数不超过 4 个
- 文件行数不超过 1000 行
- 禁止魔法数字，使用命名常量
- 错误必须显式处理；Go 中禁止用 `_` 丢弃 error
- 注释写"为什么"，不写"是什么"
- 禁止硬编码凭据、SQL 拼接、弱加密（MD5/SHA1）

### 后端分层规则
- `handler` 层只做参数校验和调用 service，不写业务逻辑
- `handler` 层使用独立 DTO，不直接引用 `domain` 类型
- `service` 层函数的输入和返回类型必须定义在 `domain` 包
- 数据库操作只在 `storage` 层，不在 service / handler 中直接操作
- 数据库操作统一使用 GORM，禁止裸写 SQL 字符串
- 不要跨 service 文件调用私有函数
- 所有 API 响应格式：`{ "success": true, "data": {}, "error": null }`
- 敏感信息只从环境变量读取，禁止硬编码

### 后端路由规范
- 路由注册统一收敛在 `internal/router` 包：只在这里出现 `Group/GET/POST/...` 等路由声明
- `handler.NewServer` 只负责初始化 Hertz + 组装依赖并调用 `router.Register(...)`，不要在 `handler/handler.go` 内直接写路由
- 模块拆分规则：
  - `internal/router/router.go`：总入口（创建 `/v1` 分组、串联各模块注册）
  - `internal/router/<module>.go`：按模块注册路由（例如 `auth.go`、`resume.go`、`interview.go`）
  - `internal/handler/<module>.go`：实现具体 handler 方法（参数校验 + 调用 service），不做路由声明
- 路由分层约定：
  - 公开路由：`/`、`/health`、`/healthz`、`/v1/auth/*`、`/v1/device/check`
  - 受保护路由：`/v1/resume/*`、`/v1/interview/*`（HTTP）、`/v1/report/*`、`/v1/questionnaire/*` 统一挂 JWT 中间件
  - WebSocket 路由：`/v1/interview/ws/:interview_id` 不挂 JWT 中间件（见下文原因），鉴权在握手阶段完成
- WebSocket 鉴权与安全要求：
  - 浏览器原生 WebSocket API 不支持自定义 header，因此 WS 路由鉴权逻辑必须在握手阶段实现
  - token 获取顺序：优先 `Authorization: Bearer <token>`，为空再读取 query 参数 `token=<token>`
  - `CheckOrigin` 不允许长期保持“全放行”，生产环境需按配置白名单校验来源
- 限流约定（保持 key 与维度一致，便于观测与压测）：
  - `/v1/resume/parse`：IP + USER 双维度限流，key 为 `resume.parse`
  - `/v1/questionnaire/submit`：IP + USER 双维度限流，key 为 `questionnaire.submit`
  - WS 握手：IP + USER 双维度限流，连接建立相比普通 HTTP 可更宽松，但必须保留限流
- 新增/调整接口的必做项：
  - 在 `internal/router` 增补或修改路由声明，并按模块归档到对应文件
  - 同步更新 `docu/api.md` 的接口定义与 curl 示例，避免文档与实现漂移
  - 变更接口语义/响应结构时同步更新前端调用与相关测试
  - 避免引入 import cycle：`internal/router` 不应依赖 `internal/handler`（路由层通过 `router.Deps` 以接口方式接收 handler 实例）

### 前端规范
- 组件使用函数式写法，禁止 class 组件
- Props 必须有 TypeScript interface 类型定义
- 组件文件名 PascalCase：`UserCard.tsx`；hooks 文件名 camelCase：`useAuth.ts`
- 避免超过 2 层 props drilling，改用 context 或 Zustand

### 日志规范（后端）
- 统一使用 `ai_interview/internal/log` 包，禁止直接使用 `hlog` 或 `slog`
- 格式：`[组件名] 动作: 具体信息`
- 级别：`Debug` 调试中间状态 / `Info` 业务关键节点 / `Warn` 可继续但需关注 / `Error` 操作失败
- 业务层使用 `hlog.CtxInfof(ctx, ...)` 注入请求上下文

---

## 测试要求

- 新增功能必须附带测试，不允许无测试的 PR
- service 层覆盖率目标 80%+
- Go 测试文件与源文件同目录，命名加 `_test.go`
- 前端测试文件命名加 `.test.ts` / `.spec.ts`
- 修改已有逻辑前，先运行相关测试确认不破坏现有功能

---

## Git 规范

- 禁止直接向 `main` 和 `dev` 分支提交，所有变更必须单开分支
- 分支名以功能命名：`feat/resume-parse`、`fix/interview-stage-bug`、`docs/update-readme`
- 分支合并后删除
- Commit 格式：`<type>(<scope>): <简短描述>`
  - type：`feat | fix | refactor | test | docs | chore`

---

## 禁止事项

- 禁止直接修改数据库 schema，必须通过 migration 文件
- 禁止在前端硬编码 API 地址，统一用环境变量
- 禁止提交 `.env` 文件
- 禁止跳过测试直接合并到主分支
- 修改公共 utils / 共享组件前，先确认影响范围
