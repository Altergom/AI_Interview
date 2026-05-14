# Project CLAUDE.md


---

## ⚠️ AI 协作红线（最高优先级）

- 修改代码前，先说明改动范围和思路，等确认再执行
- 新增依赖前，先告知包名和用途，等确认再安装
- 遇到多种实现方案时，列出优劣后由开发者决策，不要自行选择
- 不要自动重构与当前任务无关的代码
- 每次对话结束时，在末尾列出本次所有修改过的文件路径

---

## 项目概览

- **类型**：全栈项目（前端 + 后端）
- **前端**：React<!-- 填写：React / Next.js / Vue 等 -->
- **后端**：Go<!-- 填写：Node.js / Python / Go 等 -->
- **数据库**：PgSql Milvus<!-- 填写：PostgreSQL / MySQL / MongoDB 等 -->
- **包管理**：npm go install<!-- 填写：npm / pnpm / yarn / pip 等 -->

---

## 文件结构

```
project-root/
├── frontend/          # 前端代码
│   ├── src/
│   │   ├── components/   # 可复用组件
│   │   ├── pages/        # 页面/路由
│   │   ├── hooks/        # 自定义 hooks
│   │   ├── utils/        # 工具函数
│   │   └── types/        # 类型定义
│   └── tests/
├── backend/           # 后端代码
│   ├── src/
│   │   ├── routes/       # 路由/控制器
│   │   ├── services/     # 业务逻辑
│   │   ├── models/       # 数据模型
│   │   └── utils/        # 工具函数
│   └── tests/
└── CLAUDE.md
```

> 新文件必须放在对应的目录下，不要在根目录堆文件。

---

## 代码风格

### 通用原则
- 函数保持单一职责，函数体不超过 80 行 / 60 语句
- 函数圈复杂度不超过 15，参数不超过 4 个
- 文件行数不超过 1000 行，新增文件同样适用
- 禁止魔法数字，使用命名常量
- 错误必须显式处理，不允许空 catch，Go 中禁止用 `_` 丢弃 error
- 注释写"为什么"，不写"是什么"
- 禁止硬编码凭据、SQL 拼接、弱加密（MD5/SHA1）

### 前端
- 组件使用函数式写法，禁止 class 组件
- Props 必须有类型定义（TypeScript interface）
- 组件文件名使用 PascalCase：`UserCard.tsx`
- hooks 文件名使用 camelCase：`useAuth.ts`
- 避免超过 2 层的 props drilling，改用 context 或状态管理

### 后端
- 路由只做参数校验和调用 service，业务逻辑放在 service 层
- 数据库操作只在 model/repository 层
- API 层不能直接使用 model 定义的结构体
- service 层函数的输入和返回类型必须在 model 中定义
- 不要跨 service 文件调用私有函数或直接操作数据表
- 数据库操作统一使用 GORM，禁止裸写 SQL 字符串
- 所有 API 返回统一的响应格式：
  ```json
  { "success": true, "data": {}, "error": null }
  ```
- 敏感信息（密钥、密码）只从环境变量读取，禁止硬编码

---

## 测试要求

- 新增功能必须附带测试，不允许无测试的 PR
- 核心业务逻辑（service 层）覆盖率目标：**80%+**
- 测试文件与源文件同目录，命名加 `.test` / `.spec`：
  ```
  userService.ts → userService.test.ts
  ```
- 测试分层：
    - **单元测试**：service、utils、纯函数
    - **集成测试**：API 路由（使用真实数据库或 mock）
- 修改已有逻辑时，先运行相关测试确保不破坏现有功能

---

## Git 规范

### 分支管理
- 禁止直接向 `main` 和 `dev` 分支提交代码，所有新增/修改功能必须单开分支
- 分支名以功能命名，例如：`feat/resume-parse`、`fix/interview-stage-bug`
- 分支合并后删除

### 提交信息格式
```
<type>(<scope>): <简短描述>

type: feat | fix | refactor | test | docs | chore
```

示例：
```
feat(auth): 添加 JWT 刷新 token 逻辑
fix(frontend): 修复登录页表单校验不触发的问题
test(user): 补充 userService 单元测试
```

---

## 禁止事项

- 禁止直接修改数据库 schema，必须通过 migration 文件
- 禁止在前端代码中硬编码 API 地址，统一用环境变量
- 禁止提交 `.env` 文件
- 禁止跳过测试直接合并到主分支
- 修改公共 utils / 共享组件前，先确认影响范围

---

## 日志规范

- 统一使用 `ai_interview/internal/log` 包，禁止直接使用 `hlog` 或 `slog`
- 格式：`[组件名] 动作: 具体信息`
  ```go
  log.Infof("[Redis] connected")
  log.Errorf("[Postgres] ping failed: %v", err)
  log.Infof("[InterviewService] creating interview for user %s", userID)
  ```
- 级别使用规范：
  - `Debug`：开发调试，详细中间状态（LLM 输入输出、Redis 读写 key）
  - `Info`：正常业务关键节点（服务启动、面试创建、阶段切换）
  - `Warn`：可继续运行但值得关注（重试、Redis miss 回源）
  - `Error`：操作失败需处理（数据库写入失败、LLM 调用失败）
- 业务层（handler/service）使用 `hlog.CtxInfof(ctx, ...)` 注入请求上下文