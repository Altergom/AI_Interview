package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai_interview/internal/config"
)

// QwenTTSService 千问语音合成服务（使用 DashScope HTTP API）
//
// 文档：https://help.aliyun.com/zh/model-studio/qwen-tts-realtime
// API 参考：https://help.aliyun.com/zh/model-studio/dashscope-api-reference/
type QwenTTSService struct {
	apiKey  string
	baseURL string
	model   string
	voice   string
	client  *http.Client
}

// NewQwenTTSService 创建千问 TTS 服务
func NewQwenTTSService() *QwenTTSService {
	return &QwenTTSService{
		apiKey:  config.Cfg.QwenAPIKey,
		baseURL: config.Cfg.QwenBaseURL,
		model:   config.Cfg.TTSModel,
		voice:   config.Cfg.TTSVoice,
		client:  &http.Client{},
	}
}

// ConvertToAudio 将文字转为语音（使用 HTTP 异步调用）
//
// 支持的输出格式：
// - PCM（16kHz, 16bit, 单声道）
// - WAV
// - MP3
func (s *QwenTTSService) ConvertToAudio(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is empty")
	}

	// 构建请求
	// 参考：https://help.aliyun.com/zh/model-studio/dashscope-api-reference/
	url := fmt.Sprintf("%s/v1/audio/speech", s.baseURL)

	requestBody := map[string]interface{}{
		"model": s.model,
		"voice": s.voice,
		"input": text,
		// 可选参数
		"response_format": "pcm",
		"speed":           1.0, // 语速
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 读取音频数据
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	return audioData, nil
}

// SetVoice 设置音色
//
// 可选音色：
// - zhifeng_emo: 知风（支持情感）
// - zhiyan_emo: 知言（支持情感）
// - zhiyu_emo: 知语（支持情感）
func (s *QwenTTSService) SetVoice(voice string) {
	s.voice = voice
}
