package agent

import (
	"ai_interview/internal/config"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
)

func NewSelector() (*adk.ChatModelAgent, error) {
	ctx := context.Background()
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  config.Cfg.QwenAPIKey,
		BaseURL: config.Cfg.QwenBaseURL,
		Model:   config.Cfg.Selector,
	})
	if err != nil {
		return nil, fmt.Errorf("[selector]new chat model: %w", err)
	}

	selector, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "question_selector",
		Description: "题库问题选择者",
		Instruction: "你的职责是从知识库抽选合适的问题",
		Model:       model,
	})
	if err != nil {
		return nil, fmt.Errorf("[selector]new chat model agent: %w", err)
	}

	return selector, nil
}
