# ASR/TTS Tools

ASR（Automatic Speech Recognition）和 TTS（Text-to-Speech）工具封装，用于 Supervisor Agent。

## 架构设计

```
Supervisor Agent
  ↓
Tools（按需调用）:
  - ASR Tool: 语音 → 文字
  - TTS Tool: 文字 → 语音
```

## 文件说明

- `asr.go` - ASR Tool 实现
- `tts.go` - TTS Tool 实现
- `mock.go` - Mock 服务实现（用于开发测试）
- `register.go` - Tool 注册辅助函数
- `tools_test.go` - 单元测试

## 使用方式

### 1. 使用 Mock 服务（开发测试）

```go
import "ai_interview/internal/einocore/tools"

// 创建 Mock 服务
asrService := tools.NewMockASRService()
ttsService := tools.NewMockTTSService()

// 创建 Tool
asrTool, _ := tools.NewASRTool(asrService)
ttsTool, _ := tools.NewTTSTool(ttsService)

// 注册到 Supervisor
supervisor := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    ToolsConfig: adk.ToolsConfig{
        ToolsNodeConfig: compose.ToolsNodeConfig{
            Tools: []tool.BaseTool{asrTool, ttsTool},
        },
    },
})
```

### 2. 使用真实服务（生产环境）

```go
// 实现 ASRService 接口
type AliyunASRService struct {
    client *aliyun.Client
}

func (s *AliyunASRService) ConvertToText(ctx context.Context, audioData []byte) (string, error) {
    // 调用阿里云 ASR API
    return s.client.RecognizeSpeech(audioData)
}

// 实现 TTSService 接口
type AliyunTTSService struct {
    client *aliyun.Client
}

func (s *AliyunTTSService) ConvertToAudio(ctx context.Context, text string) ([]byte, error) {
    // 调用阿里云 TTS API
    return s.client.SynthesizeSpeech(text)
}

// 使用真实服务
asrService := NewAliyunASRService(...)
ttsService := NewAliyunTTSService(...)

asrTool, _ := tools.NewASRTool(asrService)
ttsTool, _ := tools.NewTTSTool(ttsService)
```

## Tool 调用流程

### ASR Tool

```
用户发送音频
  ↓
Supervisor 判断需要 ASR
  ↓
调用 ASR Tool
  ↓
ASR Tool 调用 ASRService.ConvertToText()
  ↓
返回文字给 Supervisor
```

### TTS Tool

```
Supervisor 生成回复文本
  ↓
Supervisor 判断需要语音输出
  ↓
调用 TTS Tool
  ↓
TTS Tool 调用 TTSService.ConvertToAudio()
  ↓
返回音频给用户
```

## 接口定义

### ASRService

```go
type ASRService interface {
    ConvertToText(ctx context.Context, audioData []byte) (string, error)
}
```

### TTSService

```go
type TTSService interface {
    ConvertToAudio(ctx context.Context, text string) ([]byte, error)
}
```

## 测试

```bash
go test ./internal/einocore/tools/ -v
```

## TODO

- [ ] 实现阿里云 ASR 服务
- [ ] 实现阿里云 TTS 服务
- [ ] 实现腾讯云 ASR 服务
- [ ] 实现腾讯云 TTS 服务
- [ ] 添加音频格式转换
- [ ] 添加错误重试机制
- [ ] 添加性能监控
