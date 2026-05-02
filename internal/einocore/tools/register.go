package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
)

// RegisterASRTTSTools 将 ASR 和 TTS Tool 注册到 Agent
//
// 使用示例：
//
//	// 1. 创建 ASR/TTS 服务（Mock 或真实服务）
//	asrService := tools.NewMockASRService()
//	ttsService := tools.NewMockTTSService()
//
//	// 2. 创建 Tool
//	asrTool, _ := tools.NewASRTool(asrService)
//	ttsTool, _ := tools.NewTTSTool(ttsService)
//
//	// 3. 注册到 Supervisor
//	tools := []tool.BaseTool{asrTool, ttsTool}
//	supervisor := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
//	    Tools: tools,
//	})
func RegisterASRTTSTools(ctx context.Context, asrService ASRService, ttsService TTSService) ([]tool.BaseTool, error) {
	// 创建 ASR Tool
	asrTool, err := NewASRTool(asrService)
	if err != nil {
		return nil, err
	}

	// 创建 TTS Tool
	ttsTool, err := NewTTSTool(ttsService)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{asrTool, ttsTool}, nil
}
