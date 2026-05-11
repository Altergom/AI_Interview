# 千问 ASR/TTS 实现说明

## 实现方式

使用千问 **DashScope WebSocket Realtime API**，与 Java 版 interview-guide 的实现一致。

| 服务 | 模型 | 协议 | 端点 |
|------|------|------|------|
| ASR  | `qwen3-asr-flash-realtime` | WebSocket | `wss://dashscope.aliyuncs.com/api-ws/v1/realtime?model=<model>` |
| TTS  | `qwen-tts-flash-realtime` | WebSocket | `wss://dashscope.aliyuncs.com/api-ws/v1/realtime?model=<model>` |

---

## ASR（qwen_asr.go）

**流程：**
1. 建立 WebSocket 连接（Bearer token 鉴权）
2. 发送 `session.update`：配置 server VAD（400ms 静音阈值）、PCM 16kHz、中文
3. 分块流式发送音频（每块 3200 字节 = 100ms），base64 编码
4. 发送 `input_audio_buffer.commit` 触发最终识别
5. 等待 `conversation.item.input_audio_transcription.completed` 事件，取 `transcript` 字段

**音频规格：** PCM 16kHz / 16bit / 单声道

**超时：** 30s（可通过 `QwenASRService.timeout` 调整）

---

## TTS（qwen_tts.go）

**流程：**
1. 建立 WebSocket 连接
2. 发送 `session.update`：配置音色（默认 Cherry）、PCM 16kHz、commit 模式
3. 发送 `input_text_buffer.append` 推送文本
4. 发送 `input_text_buffer.commit` 触发合成
5. 收集 `response.audio.delta` 事件（base64 PCM chunk）
6. 等待 `response.done`，返回拼接后的完整音频

**输出规格：** PCM 16kHz / 16bit / 单声道

**默认音色：** Cherry（通过 `TTSVoice` 环境变量或 `SetVoice()` 覆盖）

---

## Mock 服务（mock.go）

保留 `MockASRService` / `MockTTSService`，用于：
- 单元测试（不依赖真实 API Key）
- `APP_ENV=test` 或 `QWEN_API_KEY` 未配置时自动降级

---

## 配置

```bash
# .env
QWEN_API_KEY=sk-xxx
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1  # LLM 用
ASR_MODEL=qwen3-asr-flash-realtime
TTS_MODEL=qwen-tts-flash-realtime
TTS_VOICE=Cherry  # 可选：Cherry / Serena / Ethan 等
APP_ENV=development  # test 时自动使用 Mock
```

---

## 测试

```bash
# Mock 单元测试（无需 API Key）
go test ./internal/einocore/tools/... -run "TestASRTool_Mock|TestTTSTool_Mock" -v

# Qwen 集成测试（需要真实 QWEN_API_KEY）
go test ./internal/einocore/tools/... -run "TestQwen" -v
```
