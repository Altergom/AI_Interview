package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ASRService ASR 服务接口（需要具体实现）
type ASRService interface {
	// ConvertToText 将音频转为文字
	ConvertToText(ctx context.Context, audioData []byte) (string, error)
}

// ASRInput ASR Tool 的输入
type ASRInput struct {
	AudioData []byte `json:"audio_data" jsonschema:"description=用户的语音数据（音频字节流）"`
}

// NewASRTool 创建 ASR Tool
//
// ASR Tool 的作用：
// - 将用户的语音输入转换为文字
// - Supervisor 在收到音频输入时会调用此 Tool
//
// 使用示例：
//
//	asrService := NewAliyunASRService(...)
//	asrTool, err := NewASRTool(asrService)
//	supervisor.RegisterTool(asrTool)
func NewASRTool(asrService ASRService) (tool.InvokableTool, error) {
	asrFunc := func(ctx context.Context, input ASRInput) (string, error) {
		if len(input.AudioData) == 0 {
			return "", fmt.Errorf("audio data is empty")
		}

		// 调用 ASR 服务
		text, err := asrService.ConvertToText(ctx, input.AudioData)
		if err != nil {
			return "", fmt.Errorf("ASR conversion failed: %w", err)
		}

		return text, nil
	}

	return utils.InferTool(
		"ASR",
		"将用户的语音输入转换为文字。当收到音频数据时使用此工具。",
		asrFunc,
	)
}
