# ASR/TTS 实现总结

## ✅ 已完成

### 1. Tool 接口层
- ✅ `asr.go` - ASR Tool 接口定义
- ✅ `tts.go` - TTS Tool 接口定义
- ✅ `register.go` - Tool 注册辅助函数

### 2. Mock 实现（开发测试）
- ✅ `mock.go` - Mock ASR/TTS 服务
- ✅ 单元测试通过

### 3. 千问服务实现（生产环境）
- ✅ `qwen_asr.go` - 千问 ASR 服务框架
- ✅ `qwen_tts.go` - 千问 TTS 服务框架
- ✅ `QWEN_IMPLEMENTATION.md` - 详细实现指南

### 4. 集成到 Supervisor
- ✅ 在 `agent/supervisor.go` 中注册 ASR/TTS Tool
- ✅ 支持切换 Mock/千问服务

---

## 📁 文件结构

```
internal/einocore/tools/
├── asr.go                      # ASR Tool 接口
├── tts.go                      # TTS Tool 接口
├── mock.go                     # Mock 服务实现
├── qwen_asr.go                 # 千问 ASR 服务（待完善）
├── qwen_tts.go                 # 千问 TTS 服务（待完善）
├── register.go                 # Tool 注册辅助
├── tools_test.go               # 单元测试
├── README.md                   # 使用文档
└── QWEN_IMPLEMENTATION.md      # 千问实现指南
```

---

## 🔄 工作流程

### 当前流程（使用 Mock）

```
用户发送音频
  ↓
Graph → Supervisor
  ↓
Supervisor 判断需要 ASR
  ↓
调用 ASR Tool
  ↓
ASR Tool → MockASRService.ConvertToText()
  ↓
返回模拟文字："你好，我想应聘 Go 开发工程师"
  ↓
Supervisor 协调子 Agent 处理
  ↓
Supervisor 判断需要 TTS
  ↓
调用 TTS Tool
  ↓
TTS Tool → MockTTSService.ConvertToAudio()
  ↓
返回模拟音频
```

### 未来流程（使用千问）

```
用户发送音频
  ↓
Graph → Supervisor
  ↓
Supervisor 判断需要 ASR
  ↓
调用 ASR Tool
  ↓
ASR Tool → QwenASRService.ConvertToText()
  ↓
WebSocket 连接千问 ASR API
  ↓
实时识别返回文字
  ↓
Supervisor 协调子 Agent 处理
  ↓
Supervisor 判断需要 TTS
  ↓
调用 TTS Tool
  ↓
TTS Tool → QwenTTSService.ConvertToAudio()
  ↓
WebSocket 连接千问 TTS API
  ↓
实时合成返回音频
```

---

## 🎯 下一步工作

### 1. 实现千问 WebSocket 调用（高优先级）

**ASR 实现**：
- [ ] 实现 WebSocket 连接
- [ ] 实现音频流发送
- [ ] 实现识别结果接收
- [ ] 添加错误处理和重试

**TTS 实现**：
- [ ] 实现 WebSocket 连接
- [ ] 实现文本发送
- [ ] 实现音频流接收
- [ ] 添加错误处理和重试

**依赖**：
```bash
go get github.com/gorilla/websocket
```

### 2. 音频格式处理（中优先级）

- [ ] 添加音频格式检测
- [ ] 实现 PCM/WAV/MP3 格式转换
- [ ] 添加采样率转换（统一为 16kHz）

### 3. 性能优化（低优先级）

- [ ] 添加连接池
- [ ] 实现流式处理（边识别边返回）
- [ ] 添加缓存机制
- [ ] 添加性能监控

### 4. 测试（持续）

- [ ] 添加千问服务的集成测试
- [ ] 添加音频格式测试
- [ ] 添加性能测试
- [ ] 添加错误场景测试

---

## 📝 使用方式

### 开发环境（Mock）

```go
// 在 agent/supervisor.go 中
useQwenService := false // 使用 Mock 服务
```

### 生产环境（千问）

```go
// 在 agent/supervisor.go 中
useQwenService := true // 使用千问服务

// 配置环境变量
QWEN_API_KEY=your_api_key
QWEN_BASE_URL=https://dashscope.aliyuncs.com
```

---

## 📚 参考资料

- [千问实时语音识别文档](https://help.aliyun.com/zh/model-studio/qwen-real-time-speech-recognition)
- [千问实时语音合成文档](https://help.aliyun.com/zh/model-studio/qwen-tts-realtime)
- [Eino Tool 文档](https://www.cloudwego.io/zh/docs/eino/core_modules/components/tool/)

---

## ✅ 验证

```bash
# 编译验证
go build ./internal/einocore/tools/
go build ./internal/einocore/agent/

# 运行测试
go test ./internal/einocore/tools/ -v

# 测试结果
=== RUN   TestASRTool
    tools_test.go:26: ASR Tool Name: ASR
    tools_test.go:27: ASR Tool Description: 将用户的语音输入转换为文字。当收到音频数据时使用此工具。
    tools_test.go:35: ASR Result: [Mock ASR] 收到 15 字节音频，转换为文字：你好，我想应聘 Go 开发工程师
--- PASS: TestASRTool (0.10s)
=== RUN   TestTTSTool
    tools_test.go:56: TTS Tool Name: TTS
    tools_test.go:57: TTS Tool Description: 将文字转换为语音。当需要语音输出时使用此工具。
    tools_test.go:65: TTS Result: {"audio_data":"..."}
--- PASS: TestTTSTool (0.10s)
PASS
```
