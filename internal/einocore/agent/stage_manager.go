package agent

import (
	"ai_interview/internal/config"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
)

func NewManager() (*adk.ChatModelAgent, error) {
	ctx := context.Background()
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  config.Cfg.QwenAPIKey,
		BaseURL: config.Cfg.QwenBaseURL,
		Model:   config.Cfg.Manager,
	})
	if err != nil {
		return nil, fmt.Errorf("[manager]new chat model: %w", err)
	}

	manager, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "stage_manager",
		Description: "面试状态管理者",
		Instruction: "你的职责是分析面试状态判断是否进入下一阶段",
		Model:       model,
	})
	if err != nil {
		return nil, fmt.Errorf("[manager]new chat model agent: %w", err)
	}

	return manager, nil
}
