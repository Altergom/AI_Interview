# 千问 ASR/TTS 实现说明

## ✅ 已实现（HTTP API）

你说得对！千问支持 **DashScope HTTP API**，不需要 WebSocket。

### 实现方式

使用千问的 **OpenAI 兼容 API**：
- **ASR**：`POST /v1/audio/transcriptions`
- **TTS**：`POST /v1/audio/speech`

### 代码实现

#### ASR（qwen_asr.go）

```go
// 简单的 HTTP POST 请求
url := fmt.Sprintf("%s/v1/audio/transcriptions", s.baseURL)

requestBody := map[string]any{
    "model": "qwen3-asr-flash-realtime",
    "audio": base64.StdEncoding.EncodeToString(audioData),
    "language": "zh",
    "format": "pcm",
}

// 发送请求，获取文字结果
```

#### TTS（qwen_tts.go）

```go
// 简单的 HTTP POST 请求
url := fmt.Sprintf("%s/v1/audio/speech", s.baseURL)

requestBody := map[string]any{
    "model": "qwen-tts-flash-realtime",
    "voice": "zhifeng_emo",
    "input": text,
    "response_format": "pcm",
}

// 发送请求，获取音频数据
```

---

## 配置

### 环境变量

```bash
# .env 文件
QWEN_API_KEY=your_api_key_here
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
```

### 使用方式

```go
// 在 agent/supervisor.go 中
useQwenService := true // 使用千问服务

// 创建服务
asrService := tools.NewQwenASRService()
ttsService := tools.NewQwenTTSService()

// 创建 Tool
asrTool, _ := tools.NewASRTool(asrService)
ttsTool, _ := tools.NewTTSTool(ttsService)
```

---

## API 参考

### ASR API

**请求**：
```json
POST /v1/audio/transcriptions
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "model": "qwen3-asr-flash-realtime",
  "audio": "base64_encoded_audio_data",
  "language": "zh",
  "format": "pcm"
}
```

**响应**：
```json
{
  "text": "识别的文字内容"
}
```

### TTS API

**请求**：
```json
POST /v1/audio/speech
Authorization: Bearer {api_key}
Content-Type: application/json

{
  "model": "qwen-tts-flash-realtime",
  "voice": "zhifeng_emo",
  "input": "要转换的文字",
  "response_format": "pcm",
  "speed": 1.0
}
```

**响应**：
```
Content-Type: audio/pcm
[音频二进制数据]
```

---

## 测试

```bash
# 编译
go build ./internal/einocore/tools/

# 测试（需要配置真实的 API Key）
go test ./internal/einocore/tools/ -run TestQwen -v
```

---

## 参考资料

- [千问 DashScope API 参考](https://help.aliyun.com/zh/model-studio/dashscope-api-reference/)
- [千问 OpenAI 兼容性](https://help.aliyun.com/zh/model-studio/developer-reference/compatibility-of-openai-with-dashscope)
- [千问语音识别文档](https://help.aliyun.com/zh/model-studio/qwen-real-time-speech-recognition)
- [千问语音合成文档](https://help.aliyun.com/zh/model-studio/qwen-tts-realtime)

---

## 总结

✅ **不需要 WebSocket**
- 千问提供了简单的 HTTP API
- 使用 OpenAI 兼容的接口
- 异步调用，等待结果返回

✅ **实现简单**
- 只需要 HTTP POST 请求
- 不需要额外的 WebSocket 库
- 代码更简洁，易于维护

✅ **适合面试场景**
- 用户说完一段话
- 上传音频，等待识别
- 获取结果，继续对话
