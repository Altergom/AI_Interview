package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"ai_interview/internal/config"
	"ai_interview/internal/log"
)

const (
	asrWSURL          = "wss://dashscope.aliyuncs.com/api-ws/v1/realtime"
	asrChunkBytes     = 3200 // 100ms of PCM 16kHz/16bit/mono
	asrDefaultTimeout = 30 * time.Second
)

// QwenASRService 千问实时语音识别服务（DashScope WebSocket Realtime API）
//
// 模型：qwen3-asr-flash-realtime
// 协议：WebSocket，server VAD（400ms 静音阈值）
// 音频格式：PCM 16kHz/16bit/单声道
type QwenASRService struct {
	apiKey  string
	model   string
	timeout time.Duration
}

// NewQwenASRService 从 config.Cfg 读取配置创建 ASR 服务。
func NewQwenASRService() *QwenASRService {
	return &QwenASRService{
		apiKey:  config.Cfg.QwenAPIKey,
		model:   config.Cfg.ASRModel,
		timeout: asrDefaultTimeout,
	}
}

// ─── WebSocket 消息结构 ───────────────────────────────────────────────────────

// asrClientMsg 发送给 DashScope 的客户端消息。
type asrClientMsg struct {
	Type    string         `json:"type"`
	Session *asrSessionCfg `json:"session,omitempty"`
	Item    *asrAudioItem  `json:"item,omitempty"`
	Audio   string         `json:"audio,omitempty"` // base64 PCM
}

type asrSessionCfg struct {
	Modalities              []string          `json:"modalities"`
	InputAudioFormat        string            `json:"input_audio_format"`
	InputAudioSampleRate    int               `json:"input_audio_sample_rate"`
	TurnDetection           *asrTurnDetection `json:"turn_detection"`
	InputAudioTranscription *asrTransCfg      `json:"input_audio_transcription"`
}

type asrTurnDetection struct {
	Type              string  `json:"type"`
	SilenceDurationMs int     `json:"silence_duration_ms"`
	Threshold         float64 `json:"threshold"`
}

type asrTransCfg struct {
	Model    string `json:"model"`
	Language string `json:"language"`
}

type asrAudioItem struct {
	Type    string            `json:"type"`
	Role    string            `json:"role"`
	Content []asrAudioContent `json:"content"`
}

type asrAudioContent struct {
	Type  string `json:"type"`
	Audio string `json:"audio"` // base64 PCM
}

// asrServerMsg 从 DashScope 收到的服务端消息。
type asrServerMsg struct {
	Type       string          `json:"type"`
	Transcript string          `json:"transcript,omitempty"`
	Text       string          `json:"text,omitempty"`
	Stash      string          `json:"stash,omitempty"`
	Delta      json.RawMessage `json:"delta,omitempty"`
	Item       *struct {
		Transcript string `json:"transcript,omitempty"`
	} `json:"item,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ─── ConvertToText ────────────────────────────────────────────────────────────

// ConvertToText 将 PCM 音频数据识别为文字。
// 建立 WebSocket 连接 → 配置 session → 流式推送音频 → 等待最终结果。
func (s *QwenASRService) ConvertToText(ctx context.Context, audioData []byte) (string, error) {
	if len(audioData) == 0 {
		return "", fmt.Errorf("[ASR] audio data is empty")
	}
	if s.apiKey == "" {
		return "", fmt.Errorf("[ASR] QWEN_API_KEY not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 1. 建立 WebSocket 连接
	conn, err := s.dial(ctx)
	if err != nil {
		return "", fmt.Errorf("[ASR] dial: %w", err)
	}
	defer conn.Close()

	// 2. 并发读取服务端消息；结果通过 channel 回传
	resultCh := make(chan string, 1)
	errCh := make(chan error, 1)
	var once sync.Once

	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				select {
				case <-ctx.Done():
				default:
					once.Do(func() { errCh <- fmt.Errorf("[ASR] read: %w", err) })
				}
				return
			}
			var msg asrServerMsg
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			log.Debugf("[ASR] event: %s", msg.Type)

			switch msg.Type {
			case "conversation.item.input_audio_transcription.completed":
				text := msg.Transcript
				if text == "" && msg.Item != nil {
					text = msg.Item.Transcript
				}
				once.Do(func() { resultCh <- text })
				return

			case "error":
				var errMsg string
				if msg.Error != nil {
					errMsg = fmt.Sprintf("%s/%s: %s", msg.Error.Type, msg.Error.Code, msg.Error.Message)
				} else {
					errMsg = string(raw)
				}
				once.Do(func() { errCh <- fmt.Errorf("[ASR] server error: %s", errMsg) })
				return
			}
		}
	}()

	// 3. 发送 session.update（server VAD，400ms 静音）
	sessionMsg := asrClientMsg{
		Type: "session.update",
		Session: &asrSessionCfg{
			Modalities:           []string{"text"},
			InputAudioFormat:     "pcm16",
			InputAudioSampleRate: 16000,
			TurnDetection: &asrTurnDetection{
				Type:              "server_vad",
				SilenceDurationMs: 400,
				Threshold:         0.5,
			},
			InputAudioTranscription: &asrTransCfg{
				Model:    s.model,
				Language: "zh",
			},
		},
	}
	if err := s.writeJSON(conn, sessionMsg); err != nil {
		return "", err
	}

	// 4. 分块流式发送音频
	for offset := 0; offset < len(audioData); offset += asrChunkBytes {
		end := offset + asrChunkBytes
		if end > len(audioData) {
			end = len(audioData)
		}
		chunk := audioData[offset:end]

		appendMsg := asrClientMsg{
			Type:  "input_audio_buffer.append",
			Audio: base64.StdEncoding.EncodeToString(chunk),
		}
		if err := s.writeJSON(conn, appendMsg); err != nil {
			return "", err
		}
		// 小间隔避免塞满发送缓冲
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}

	// 5. 提交：触发服务端 VAD 最终识别
	if err := s.writeJSON(conn, asrClientMsg{Type: "input_audio_buffer.commit"}); err != nil {
		return "", err
	}

	// 6. 等待结果
	select {
	case text := <-resultCh:
		log.Infof("[ASR] transcribed: %q", text)
		return text, nil
	case err := <-errCh:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("[ASR] timeout waiting for transcription result")
	}
}

// ─── 内部辅助 ─────────────────────────────────────────────────────────────────

func (s *QwenASRService) dial(ctx context.Context) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s?model=%s", asrWSURL, s.model)
	hdr := http.Header{"Authorization": []string{"Bearer " + s.apiKey}}

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.DialContext(ctx, url, hdr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *QwenASRService) writeJSON(conn *websocket.Conn, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("[ASR] marshal msg: %w", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("[ASR] write msg: %w", err)
	}
	return nil
}
