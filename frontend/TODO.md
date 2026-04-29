# 前端开发 TODO

> 任务状态说明：
> - 无标记：未开始
> - ○ 进行中
> - ✓ 已完成

---

## Phase 1: 项目初始化

- ✓ 创建 frontend 目录
- ✓ 初始化 Vite + React + TypeScript 项目
- ✓ 安装核心依赖（react-router-dom, zustand, axios, monaco-editor, recharts）
- ✓ 配置 Tailwind CSS（深蓝色主题）
- ✓ 创建项目目录结构
- ✓ 创建环境变量配置文件

---

## Phase 2: 基础设施层

### 类型定义 (types/)
- ✓ 创建 api.ts（API 响应基础类型）
- ✓ 创建 user.ts（用户相关类型）
- ✓ 创建 interview.ts（面试相关类型）
- ✓ 创建 resume.ts（简历相关类型）
- ✓ 创建 report.ts（报告相关类型）

### API 服务层 (services/)
- ✓ 创建 api.ts（Axios 实例配置、拦截器、错误处理）
- ✓ 创建 auth.ts（认证接口：注册、登录、游客模式）
- ✓ 创建 resume.ts（简历接口：解析、提交）
- ✓ 创建 interview.ts（面试接口：创建、推进、结束）
- ✓ 创建 device.ts（设备检测接口）
- ✓ 创建 report.ts（报告接口：查询状态、获取报告）
- ✓ 创建 questionnaire.ts（问卷接口）

### 状态管理 (store/)
- ✓ 创建 authStore.ts（用户信息、token、登录状态）
- ✓ 创建 interviewStore.ts（面试状态、当前阶段、对话历史）
- ✓ 创建 deviceStore.ts（设备检测状态、权限状态）
- ✓ 创建 resumeStore.ts（简历信息）

### 工具函数 (utils/)
- ✓ 创建 constants.ts（常量定义：岗位、方向、阶段等）
- ✓ 创建 validators.ts（表单验证函数）
- ✓ 创建 formatters.ts（数据格式化函数）
- ✓ 创建 storage.ts（localStorage 封装）

### 路由配置
- ✓ 创建 router.tsx（路由定义、路由守卫）

---

## Phase 3: 通用组件库

### 基础组件 (components/common/)
- ✓ Button.tsx（主按钮、次按钮、文字按钮）
- ✓ Input.tsx（文本输入、密码输入）
- ✓ Select.tsx（下拉选择）
- ✓ FileUpload.tsx（文件上传）
- ✓ Modal.tsx（弹窗）
- ✓ Loading.tsx（加载动画）
- ✓ Toast.tsx（消息提示）
- ✓ Card.tsx（卡片容器）

### 布局组件 (components/layout/)
- ✓ Header.tsx（顶部导航）
- ✓ Footer.tsx（底部信息）
- ✓ Container.tsx（页面容器）

---

## Phase 4: 核心页面实现

### Index 首页
- ✓ Index.tsx（欢迎页面、入口按钮）

### 登录/注册 (pages/Auth/)
- ✓ Login.tsx（登录表单）
- ✓ Register.tsx（注册表单）
- ✓ GuestEntry.tsx（游客模式入口）

### 简历信息页 (pages/Resume/)
- ✓ ResumeForm.tsx（表单主体）
- ✓ BasicInfo.tsx（基本信息表单）
- ✓ SkillsInput.tsx（技能输入）
- ✓ ProjectForm.tsx（项目经验表单）
- ✓ EducationForm.tsx（教育背景表单）
- ✓ PDFUpload.tsx（PDF 上传和解析）

### 岗位方向选择 (pages/Config/)
- ✓ PositionSelect.tsx（岗位选择）
- ✓ DirectionSelect.tsx（方向选择）

### 准备页面 (pages/Prepare/)
- ✓ DeviceCheck.tsx（设备检测主页面）
- ✓ MicrophoneTest.tsx（麦克风测试）
- ✓ CameraTest.tsx（摄像头测试）
- ✓ AudioVisualizer.tsx（音频波形可视化）

### 面试间 (pages/Interview/)
- ✓ InterviewRoom.tsx（面试间主页面）
- ✓ ChatPanel.tsx（对话记录面板）
- ✓ StageProgress.tsx（阶段进度条）
- ✓ AudioIndicator.tsx（音频输入状态指示）

### 报告相关 (pages/Report/)
- ✓ ReportGenerating.tsx（报告生成中页面）
- ✓ ReportView.tsx（报告展示页面）
- ✓ RadarChart.tsx（雷达图组件）

### 问卷 (pages/Questionnaire/)
- ✓ QuestionnairePage.tsx（问卷主页面）
- ✓ TurnItem.tsx（单轮对话展示）

### 结束页 (pages/End/)
- ✓ EndPage.tsx（结束页面）

---

## Phase 5: 核心功能实现

### 自定义 Hooks (hooks/)
- ✓ useAuth.ts（认证相关：登录、登出、token 管理）
- ✓ useDevice.ts（设备检测：权限请求、设备测试）
- ✓ useAudio.ts（音频采集：getUserMedia、VAD、分片上传）
- ✓ useVideo.ts（视频采集：getUserMedia、视频流）
- ✓ useSSE.ts（SSE 连接：EventSource、断线重连、事件监听）
- ✓ useInterview.ts（面试状态管理）

### 面试相关组件 (components/interview/)
- ✓ CodeEditor.tsx（Monaco Editor 集成、语言切换、代码提交）
- ✓ AudioPlayer.tsx（音频流播放）
- ✓ MessageBubble.tsx（对话气泡）

---

## Phase 6: 样式和交互优化

- ✓ 实现页面过渡动画
- ✓ 添加加载状态
- ✓ 添加错误提示
- ✓ 响应式布局适配
- ✓ 优化表单交互体验

---

## Phase 7: 测试和优化

- ✓ 编写关键 hooks 的单元测试
- ✓ 编写 utils 函数的单元测试
- ✓ 集成测试（完整用户流程）
- ✓ 性能优化（代码分割、懒加载）
- ✓ 错误边界处理
- ✓ 浏览器兼容性测试

---

## Phase 8: 文档和部署

- ✓ 编写前端 README.md
- ✓ 编写组件使用文档
- ✓ 配置生产环境构建
- ✓ 配置 Docker 部署
