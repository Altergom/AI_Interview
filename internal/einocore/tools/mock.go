package tools

import (
	"context"
	"fmt"
	"time"
)

// MockASRService Mock ASR 服务（用于开发测试）
type MockASRService struct{}

// NewMockASRService 创建 Mock ASR 服务
func NewMockASRService() *MockASRService {
	return &MockASRService{}
}

// ConvertToText 模拟将音频转为文字
func (m *MockASRService) ConvertToText(ctx context.Context, audioData []byte) (string, error) {
	// 模拟 ASR 处理延迟
	time.Sleep(100 * time.Millisecond)

	// 返回模拟的文字结果
	return fmt.Sprintf("[Mock ASR] 收到 %d 字节音频，转换为文字：你好，我想应聘 Go 开发工程师", len(audioData)), nil
}

// MockTTSService Mock TTS 服务（用于开发测试）
type MockTTSService struct{}

// NewMockTTSService 创建 Mock TTS 服务
func NewMockTTSService() *MockTTSService {
	return &MockTTSService{}
}

// ConvertToAudio 模拟将文字转为语音
func (m *MockTTSService) ConvertToAudio(ctx context.Context, text string) ([]byte, error) {
	// 模拟 TTS 处理延迟
	time.Sleep(100 * time.Millisecond)

	// 返回模拟的音频数据（实际应该是音频字节流）
	mockAudio := []byte(fmt.Sprintf("[Mock TTS Audio] %s", text))
	return mockAudio, nil
}
