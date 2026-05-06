# Project CLAUDE.md

> 项目级配置，全局 CLAUDE.md 的规则同样适用。

---

## 项目概览

- **类型**：全栈项目（前端 + 后端）
- **前端**：React<!-- 填写：React / Next.js / Vue 等 -->
- **后端**：Go<!-- 填写：Node.js / Python / Go 等 -->
- **数据库**：PgSql Milvus<!-- 填写：PostgreSQL / MySQL / MongoDB 等 -->
- **包管理**：npm go install<!-- 填写：npm / pnpm / yarn / pip 等 -->

---

## 常用命令

```bash
# 开发
<!-- frontend: npm run dev -->
<!-- backend: npm run dev / python main.py -->

# 测试
<!-- frontend: npm test -->
<!-- backend: pytest / npm test -->

# 构建
<!-- npm run build -->

# Lint / Format
<!-- npm run lint / ruff check . -->
```

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
- 函数保持单一职责，超过 50 行考虑拆分
- 禁止魔法数字，使用命名常量
- 错误必须显式处理，不允许空 catch
- 注释写"为什么"，不写"是什么"

### 前端
- 组件使用函数式写法，禁止 class 组件
- Props 必须有类型定义（TypeScript interface）
- 组件文件名使用 PascalCase：`UserCard.tsx`
- hooks 文件名使用 camelCase：`useAuth.ts`
- 避免超过 2 层的 props drilling，改用 context 或状态管理

### 后端
- 路由只做参数校验和调用 service，业务逻辑放在 service 层
- 数据库操作只在 model/repository 层
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
    - **E2E 测试**：关键用户流程（可选）
- 修改已有逻辑时，先运行相关测试确保不破坏现有功能

---

## Git 规范

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

## 给 Claude 的指令

- 修改代码前，先说明改动范围和思路，等我确认再执行
- 新增依赖前，先告知包名和用途
- 遇到多种实现方案时，列出优劣后由我决策
- 不要自动重构与当前任务无关的代码
- 每次对话结束时，在末尾列出本次所有修改过的文件路径