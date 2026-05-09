package handler

import "encoding/json"

// ─── 上行消息（前端 → 服务端）────────────────────────────────────────────────

// UpMsgType 上行消息类型。
type UpMsgType string

const (
	// UpMsgAudioChunk 流式 PCM 音频帧（二进制帧，无 JSON 包装）。
	UpMsgAudioChunk UpMsgType = "audio_chunk"
	// UpMsgControl 控制指令（start / pause / resume / stop）。
	UpMsgControl UpMsgType = "control"
	// UpMsgCodeSubmit 代码题提交（携带代码文本和语言）。
	UpMsgCodeSubmit UpMsgType = "code_submit"
)

// UpMsg 上行文本帧的通用包装。
// 音频帧（audio_chunk）直接发二进制帧，不走此结构。
type UpMsg struct {
	Type    UpMsgType       `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// ControlPayload 控制指令负载。
type ControlPayload struct {
	// Action: start | pause | resume | stop
	Action string `json:"action"`
}

// CodeSubmitPayload 代码提交负载。
type CodeSubmitPayload struct {
	Language string `json:"language"` // e.g. "go", "python", "java"
	Code     string `json:"code"`
}

// ─── 下行消息（服务端 → 前端）────────────────────────────────────────────────

// DownMsgType 下行消息类型。
type DownMsgType string

const (
	// DownMsgASRPartial ASR 中间结果（流式识别中）。
	DownMsgASRPartial DownMsgType = "asr_partial"
	// DownMsgASRFinal ASR 最终结果（本轮识别结束）。
	DownMsgASRFinal DownMsgType = "asr_final"
	// DownMsgLLMToken LLM 流式 token（用于显示 AI 回复打字效果）。
	DownMsgLLMToken DownMsgType = "llm_token"
	// DownMsgTTSAudio TTS 音频帧（二进制帧，PCM 16kHz）。
	// 服务端以二进制 WebSocket 帧发送，此常量用于日志标记。
	DownMsgTTSAudio DownMsgType = "tts_audio"
	// DownMsgStageChange 阶段切换事件。
	DownMsgStageChange DownMsgType = "stage_change"
	// DownMsgError 错误通知（不强制断连，前端自行决定）。
	DownMsgError DownMsgType = "error"
	// DownMsgReportReady 报告生成完成通知（Report Worker 异步推送）。
	DownMsgReportReady DownMsgType = "report_ready"
)

// DownMsg 下行文本帧的通用包装。
type DownMsg struct {
	Type    DownMsgType `json:"type"`
	Payload any         `json:"payload,omitempty"`
}

// ASRPartialPayload ASR 中间结果负载。
type ASRPartialPayload struct {
	Text string `json:"text"`
}

// ASRFinalPayload ASR 最终结果负载。
type ASRFinalPayload struct {
	Text   string `json:"text"`
	TurnID string `json:"turn_id"`
}

// LLMTokenPayload LLM 流式 token 负载。
type LLMTokenPayload struct {
	Token  string `json:"token"`
	TurnID string `json:"turn_id"`
}

// StageChangePayload 阶段切换负载。
type StageChangePayload struct {
	// Stage: intro | questioning | closing | end
	Stage          string `json:"stage"`
	QuestionsAsked int    `json:"questions_asked"`
}

// ErrorPayload 错误通知负载。
type ErrorPayload struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
