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
