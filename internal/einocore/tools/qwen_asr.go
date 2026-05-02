package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ai_interview/internal/config"
)

// QwenASRService 千问语音识别服务（使用 DashScope HTTP API）
//
// 文档：https://help.aliyun.com/zh/model-studio/qwen-real-time-speech-recognition
// API 参考：https://help.aliyun.com/zh/model-studio/dashscope-api-reference/
type QwenASRService struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewQwenASRService 创建千问 ASR 服务
func NewQwenASRService() *QwenASRService {
	return &QwenASRService{
		apiKey:  config.Cfg.QwenAPIKey,
		baseURL: config.Cfg.QwenBaseURL,
		model:   config.Cfg.ASRModel,
		client:  &http.Client{},
	}
}

// ConvertToText 将音频转为文字（使用 HTTP 异步调用）
//
// 支持的音频格式：
// - PCM（16kHz, 16bit, 单声道）
// - WAV
// - MP3
// - FLAC
func (s *QwenASRService) ConvertToText(ctx context.Context, audioData []byte) (string, error) {
	if len(audioData) == 0 {
		return "", fmt.Errorf("audio data is empty")
	}

	// 构建请求
	// 参考：https://help.aliyun.com/zh/model-studio/dashscope-api-reference/
	url := fmt.Sprintf("%s/v1/audio/transcriptions", s.baseURL)

	// 将音频数据编码为 base64
	audioBase64 := base64.StdEncoding.EncodeToString(audioData)

	requestBody := map[string]interface{}{
		"model": s.model,
		"audio": audioBase64,
		// 可选参数
		"language": "zh", // 中文
		"format":   "pcm",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	return result.Text, nil
}
