package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"ai_interview/internal/config"
	"ai_interview/internal/log"
	"ai_interview/internal/utils/wsx"
)

const (
	ttsWSURL          = "wss://dashscope.aliyuncs.com/api-ws/v1/realtime"
	ttsDefaultTimeout = 30 * time.Second
	ttsDefaultVoice   = "Cherry"
)

// QwenTTSService 千问实时语音合成服务（DashScope WebSocket Realtime API）
//
// 模型：qwen-tts-flash-realtime（默认）/ qwen3-tts-flash-realtime
// 协议：WebSocket，commit 模式（手动提交文本触发合成）
// 音频格式：PCM 16kHz/16bit/单声道
type QwenTTSService struct {
	apiKey  string
	model   string
	voice   string
	timeout time.Duration
}

// NewQwenTTSService 从 config.Cfg 读取配置创建 TTS 服务。
func NewQwenTTSService() *QwenTTSService {
	voice := config.Cfg.TTSVoice
	if voice == "" {
		voice = ttsDefaultVoice
	}
	return &QwenTTSService{
		apiKey:  config.Cfg.QwenAPIKey,
		model:   config.Cfg.TTSModel,
		voice:   voice,
		timeout: ttsDefaultTimeout,
	}
}

// ─── WebSocket 消息结构 ───────────────────────────────────────────────────────

// ttsClientMsg 发送给 DashScope 的客户端消息。
type ttsClientMsg struct {
	Type    string         `json:"type"`
	Session *ttsSessionCfg `json:"session,omitempty"`
	Text    string         `json:"text,omitempty"` // input_text_buffer.append
}

type ttsSessionCfg struct {
	Modalities            []string `json:"modalities"`
	Voice                 string   `json:"voice"`
	OutputAudioFormat     string   `json:"output_audio_format"`
	OutputAudioSampleRate int      `json:"output_audio_sample_rate"`
	Mode                  string   `json:"mode"` // "commit"
}

// ttsServerMsg 从 DashScope 收到的服务端消息。
type ttsServerMsg struct {
	Type  string `json:"type"`
	Delta string `json:"delta,omitempty"` // base64 PCM audio chunk
	Error *struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ─── ConvertToAudio ───────────────────────────────────────────────────────────

// ConvertToAudio 将文字合成为 PCM 音频。
// 建立 WebSocket 连接 → 配置 session → 发送文本并 commit → 收集音频 delta → 返回完整 PCM。
func (s *QwenTTSService) ConvertToAudio(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("[TTS] text is empty")
	}
	if s.apiKey == "" {
		return nil, fmt.Errorf("[TTS] QWEN_API_KEY not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 1. 建立 WebSocket 连接
	conn, err := s.dial(ctx)
	if err != nil {
		return nil, fmt.Errorf("[TTS] dial: %w", err)
	}
	defer conn.Close()

	// 2. 并发读取音频 delta
	type result struct {
		audio []byte
		err   error
	}
	resultCh := make(chan result, 1)
	var (
		audioBuf []byte
		mu       sync.Mutex
		once     sync.Once
	)

	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					once.Do(func() { resultCh <- result{err: fmt.Errorf("[TTS] read: %w", err)} })
				}
				return
			}
			var msg ttsServerMsg
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			log.Debugf("[TTS] event: %s", msg.Type)

			switch msg.Type {
			case "response.audio.delta":
				if msg.Delta != "" {
					chunk, decErr := base64.StdEncoding.DecodeString(msg.Delta)
					if decErr != nil {
						log.Warnf("[TTS] base64 decode delta: %v", decErr)
						continue
					}
					mu.Lock()
					audioBuf = append(audioBuf, chunk...)
					mu.Unlock()
				}

			case "response.done":
				mu.Lock()
				final := make([]byte, len(audioBuf))
				copy(final, audioBuf)
				mu.Unlock()
				log.Infof("[TTS] synthesis done, %d bytes", len(final))
				once.Do(func() { resultCh <- result{audio: final} })
				return

			case "error":
				var errMsg string
				if msg.Error != nil {
					errMsg = fmt.Sprintf("%s/%s: %s", msg.Error.Type, msg.Error.Code, msg.Error.Message)
				} else {
					errMsg = string(raw)
				}
				once.Do(func() { resultCh <- result{err: fmt.Errorf("[TTS] server error: %s", errMsg)} })
				return
			}
		}
	}()

	// 3. session.update（commit 模式，PCM 16kHz）
	sessionMsg := ttsClientMsg{
		Type: "session.update",
		Session: &ttsSessionCfg{
			Modalities:            []string{"audio"},
			Voice:                 s.voice,
			OutputAudioFormat:     "pcm16",
			OutputAudioSampleRate: 16000,
			Mode:                  "commit",
		},
	}
	if err := s.writeJSON(conn, sessionMsg); err != nil {
		return nil, err
	}

	// 4. 发送文本
	if err := s.writeJSON(conn, ttsClientMsg{Type: "input_text_buffer.append", Text: text}); err != nil {
		return nil, err
	}

	// 5. commit：触发合成
	if err := s.writeJSON(conn, ttsClientMsg{Type: "input_text_buffer.commit"}); err != nil {
		return nil, err
	}

	// 6. 等待完成
	select {
	case res := <-resultCh:
		if res.err != nil {
			return nil, res.err
		}
		return res.audio, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("[TTS] timeout waiting for synthesis result")
	}
}

// SetVoice 动态切换音色（单次合成生效）。
func (s *QwenTTSService) SetVoice(voice string) {
	s.voice = voice
}

// ─── 内部辅助 ─────────────────────────────────────────────────────────────────

func (s *QwenTTSService) dial(ctx context.Context) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s?model=%s", ttsWSURL, s.model)
	return wsx.DialBearer(ctx, url, s.apiKey)
}

func (s *QwenTTSService) writeJSON(conn *websocket.Conn, v any) error {
	return wsx.WriteJSON(conn, v, "TTS")
}
