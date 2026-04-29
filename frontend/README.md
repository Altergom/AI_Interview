# AI 面试系统 - 前端

基于 React + TypeScript + Vite 构建的 AI 面试系统前端应用。

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **路由**: React Router v6
- **状态管理**: Zustand
- **样式**: Tailwind CSS
- **HTTP 客户端**: Axios
- **代码编辑器**: Monaco Editor
- **图表**: Recharts

## 项目结构

```
frontend/
├── src/
│   ├── components/       # 可复用组件
│   │   ├── common/      # 通用组件（Button, Input, Modal 等）
│   │   ├── layout/      # 布局组件（Header, Footer, Container）
│   │   └── interview/   # 面试相关组件
│   ├── pages/           # 页面组件
│   │   ├── Auth/        # 认证相关页面
│   │   ├── Resume/      # 简历信息页
│   │   ├── Config/      # 岗位方向选择
│   │   ├── Prepare/     # 设备检测
│   │   ├── Interview/   # 面试间
│   │   ├── Report/      # 报告页面
│   │   └── ...
│   ├── hooks/           # 自定义 Hooks
│   ├── services/        # API 调用
│   ├── store/           # Zustand 状态管理
│   ├── types/           # TypeScript 类型定义
│   ├── utils/           # 工具函数
│   ├── router.tsx       # 路由配置
│   └── main.tsx         # 应用入口
├── public/              # 静态资源
└── package.json
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

应用将在 http://localhost:5173 启动。

### 构建生产版本

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

### 预览生产构建

```bash
npm run preview
```

## 环境变量

创建 `.env` 文件配置环境变量：

```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_SSE_BASE_URL=http://localhost:8080
```

## 核心功能

### 1. 用户认证
- 注册/登录
- 游客模式
- Token 持久化

### 2. 简历管理
- 表单填写
- PDF 上传解析
- 简历信息持久化

### 3. 面试配置
- 岗位选择（Golang, Java, Frontend, Test）
- 方向选择（Backend, Cloud, Agent, Server）

### 4. 设备检测
- 麦克风权限和测试
- 摄像头权限和测试
- 音频波形可视化

### 5. 面试间
- 实时语音对话（SSE）
- 代码编辑和提交
- 阶段进度展示
- 对话历史记录

### 6. 面试报告
- 多维度评分雷达图
- 优势点和待改进点
- 报告生成状态轮询

## 状态管理

使用 Zustand 管理全局状态：

- `authStore`: 用户认证状态
- `interviewStore`: 面试状态和对话历史
- `deviceStore`: 设备检测状态
- `resumeStore`: 简历信息

## 路由守卫

- `RequireAuth`: 需要登录才能访问
- `RedirectIfAuth`: 已登录用户重定向

## 代码规范

- 使用 ESLint 进行代码检查
- 使用 TypeScript 严格模式
- 组件使用函数式写法
- Props 必须有类型定义

## 性能优化

- 路由懒加载（React.lazy + Suspense）
- 代码分割
- 图片懒加载
- 状态持久化（localStorage）

## 浏览器支持

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## 开发指南

详见 [核心代码.md](./核心代码.md) 和 [TODO.md](./TODO.md)。

## License

MIT
