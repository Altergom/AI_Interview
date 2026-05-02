package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// TTSService TTS 服务接口（需要具体实现）
type TTSService interface {
	// ConvertToAudio 将文字转为语音
	ConvertToAudio(ctx context.Context, text string) ([]byte, error)
}

// TTSInput TTS Tool 的输入
type TTSInput struct {
	Text string `json:"text" jsonschema:"description=需要转换为语音的文字内容"`
}

// TTSOutput TTS Tool 的输出
type TTSOutput struct {
	AudioData []byte `json:"audio_data"`
}

// NewTTSTool 创建 TTS Tool
//
// TTS Tool 的作用：
// - 将 AI 的文字回复转换为语音
// - Supervisor 在需要语音输出时会调用此 Tool
//
// 使用示例：
//
//	ttsService := NewAliyunTTSService(...)
//	ttsTool, err := NewTTSTool(ttsService)
//	supervisor.RegisterTool(ttsTool)
func NewTTSTool(ttsService TTSService) (tool.InvokableTool, error) {
	ttsFunc := func(ctx context.Context, input TTSInput) (*TTSOutput, error) {
		if input.Text == "" {
			return nil, fmt.Errorf("text is empty")
		}

		// 调用 TTS 服务
		audioData, err := ttsService.ConvertToAudio(ctx, input.Text)
		if err != nil {
			return nil, fmt.Errorf("TTS conversion failed: %w", err)
		}

		return &TTSOutput{
			AudioData: audioData,
		}, nil
	}

	return utils.InferTool(
		"TTS",
		"将文字转换为语音。当需要语音输出时使用此工具。",
		ttsFunc,
	)
}
