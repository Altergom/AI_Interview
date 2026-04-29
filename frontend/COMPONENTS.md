# 组件使用文档

本文档介绍前端通用组件的使用方法和示例。

## 基础组件

### Button

按钮组件，支持多种样式和状态。

**Props:**
- `variant`: 按钮样式 - `'primary' | 'secondary' | 'text'`（默认 `'primary'`）
- `size`: 按钮尺寸 - `'sm' | 'md' | 'lg'`（默认 `'md'`）
- `loading`: 加载状态 - `boolean`（默认 `false`）
- `disabled`: 禁用状态 - `boolean`
- 其他原生 button 属性

**示例:**
```tsx
import { Button } from '@/components/common/Button';

// 主按钮
<Button onClick={handleClick}>提交</Button>

// 次按钮
<Button variant="secondary">取消</Button>

// 文字按钮
<Button variant="text">了解更多</Button>

// 加载状态
<Button loading>提交中...</Button>

// 小尺寸
<Button size="sm">小按钮</Button>
```

---

### Input

输入框组件，支持密码显示/隐藏切换。

**Props:**
- `label`: 标签文本 - `string`
- `error`: 错误提示 - `string`
- `helperText`: 辅助文本 - `string`
- `type`: 输入类型 - `string`（默认 `'text'`）
- 其他原生 input 属性

**示例:**
```tsx
import { Input } from '@/components/common/Input';

// 基础输入框
<Input
  label="用户名"
  placeholder="请输入用户名"
  value={username}
  onChange={(e) => setUsername(e.target.value)}
/>

// 密码输入框（带显示/隐藏切换）
<Input
  type="password"
  label="密码"
  placeholder="请输入密码"
  value={password}
  onChange={(e) => setPassword(e.target.value)}
/>

// 带错误提示
<Input
  label="邮箱"
  error="邮箱格式不正确"
  value={email}
  onChange={(e) => setEmail(e.target.value)}
/>

// 带辅助文本
<Input
  label="验证码"
  helperText="验证码已发送到您的邮箱"
  value={code}
  onChange={(e) => setCode(e.target.value)}
/>
```

---

### Select

下拉选择组件。

**Props:**
- `label`: 标签文本 - `string`
- `error`: 错误提示 - `string`
- `options`: 选项列表 - `{ value: string; label: string }[]`
- `placeholder`: 占位文本 - `string`
- 其他原生 select 属性

**示例:**
```tsx
import { Select } from '@/components/common/Select';

const positionOptions = [
  { value: 'golang', label: 'Golang 工程师' },
  { value: 'java', label: 'Java 工程师' },
  { value: 'frontend', label: '前端工程师' },
];

<Select
  label="选择岗位"
  placeholder="请选择岗位"
  options={positionOptions}
  value={position}
  onChange={(e) => setPosition(e.target.value)}
/>
```

---

### FileUpload

文件上传组件，支持拖拽上传。

**Props:**
- `accept`: 接受的文件类型 - `string`（默认 `'.pdf'`）
- `maxSize`: 最大文件大小（字节）- `number`（默认 10MB）
- `onFileSelect`: 文件选择回调 - `(file: File) => void`
- `error`: 错误提示 - `string`
- `label`: 标签文本 - `string`

**示例:**
```tsx
import { FileUpload } from '@/components/common/FileUpload';

<FileUpload
  label="上传简历"
  accept=".pdf,.doc,.docx"
  maxSize={10 * 1024 * 1024}
  onFileSelect={(file) => {
    console.log('选择的文件:', file);
  }}
  error={uploadError}
/>
```

---

### Modal

弹窗组件。

**Props:**
- `isOpen`: 是否打开 - `boolean`
- `onClose`: 关闭回调 - `() => void`
- `title`: 标题 - `string`
- `footer`: 底部内容 - `React.ReactNode`
- `size`: 尺寸 - `'sm' | 'md' | 'lg'`（默认 `'md'`）
- `children`: 弹窗内容 - `React.ReactNode`

**示例:**
```tsx
import { Modal } from '@/components/common/Modal';
import { Button } from '@/components/common/Button';

const [isOpen, setIsOpen] = useState(false);

<Modal
  isOpen={isOpen}
  onClose={() => setIsOpen(false)}
  title="确认操作"
  footer={
    <div className="flex gap-3 justify-end">
      <Button variant="secondary" onClick={() => setIsOpen(false)}>
        取消
      </Button>
      <Button onClick={handleConfirm}>
        确认
      </Button>
    </div>
  }
>
  <p>确定要执行此操作吗？</p>
</Modal>
```

---

### Loading

加载动画组件。

**Props:**
- `size`: 尺寸 - `'sm' | 'md' | 'lg'`（默认 `'md'`）
- `text`: 加载文本 - `string`
- `fullScreen`: 全屏显示 - `boolean`（默认 `false`）

**示例:**
```tsx
import { Loading } from '@/components/common/Loading';

// 局部加载
<Loading size="md" text="加载中..." />

// 全屏加载
<Loading size="lg" text="正在处理..." fullScreen />
```

---

### Toast

消息提示组件。

**Props:**
- `message`: 提示消息 - `string`
- `type`: 提示类型 - `'success' | 'error' | 'warning' | 'info'`（默认 `'info'`）
- `duration`: 显示时长（毫秒）- `number`（默认 3000）
- `onClose`: 关闭回调 - `() => void`

**示例:**
```tsx
import { Toast } from '@/components/common/Toast';
import { useState } from 'react';

const [toast, setToast] = useState<{ message: string; type: ToastType } | null>(null);

// 显示提示
setToast({ message: '操作成功', type: 'success' });

// 渲染
{toast && (
  <Toast
    message={toast.message}
    type={toast.type}
    onClose={() => setToast(null)}
  />
)}
```

---

### Card

卡片容器组件。

**Props:**
- `title`: 标题 - `string`
- `padding`: 内边距 - `'none' | 'sm' | 'md' | 'lg'`（默认 `'md'`）
- `className`: 自定义类名 - `string`
- `children`: 卡片内容 - `React.ReactNode`

**示例:**
```tsx
import { Card } from '@/components/common/Card';

<Card title="用户信息" padding="lg">
  <p>姓名: 张三</p>
  <p>邮箱: zhangsan@example.com</p>
</Card>
```

---

## 布局组件

### Header

顶部导航组件，自动显示用户信息和退出按钮。

**示例:**
```tsx
import { Header } from '@/components/layout/Header';

<Header />
```

---

### Footer

底部信息组件。

**示例:**
```tsx
import { Footer } from '@/components/layout/Footer';

<Footer />
```

---

### Container

页面容器组件，统一页面布局。

**Props:**
- `showHeader`: 是否显示 Header - `boolean`（默认 `true`）
- `showFooter`: 是否显示 Footer - `boolean`（默认 `true`）
- `maxWidth`: 最大宽度 - `'sm' | 'md' | 'lg' | 'xl' | 'full'`（默认 `'xl'`）
- `className`: 自定义类名 - `string`
- `children`: 页面内容 - `React.ReactNode`

**示例:**
```tsx
import { Container } from '@/components/layout/Container';

<Container maxWidth="lg">
  <h1>页面标题</h1>
  <p>页面内容...</p>
</Container>

// 不显示 Header 和 Footer
<Container showHeader={false} showFooter={false}>
  <div>全屏内容</div>
</Container>
```

---

## 错误边界

### ErrorBoundary

错误边界组件，捕获子组件错误。

**Props:**
- `children`: 子组件 - `React.ReactNode`
- `fallback`: 自定义错误 UI - `React.ReactNode`

**示例:**
```tsx
import { ErrorBoundary } from '@/components/common/ErrorBoundary';

<ErrorBoundary>
  <App />
</ErrorBoundary>

// 自定义错误 UI
<ErrorBoundary fallback={<div>出错了，请刷新页面</div>}>
  <SomeComponent />
</ErrorBoundary>
```

---

## 样式规范

所有组件使用 Tailwind CSS，遵循以下规范：

- **主题色**: `primary-500` (#3b82f6)
- **圆角**: `rounded-lg`
- **阴影**: `shadow-md` / `shadow-lg`
- **过渡**: `transition-colors duration-200`
- **间距**: 使用 Tailwind 的间距系统（p-4, m-2 等）

## 自定义样式

所有组件都支持通过 `className` prop 添加自定义样式：

```tsx
<Button className="w-full mt-4">
  全宽按钮
</Button>
```
