package agent

import (
	"ai_interview/internal/config"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
)

func NewEvaluator() (*adk.ChatModelAgent, error) {
	ctx := context.Background()
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  config.Cfg.QwenAPIKey,
		BaseURL: config.Cfg.QwenBaseURL,
		Model:   config.Cfg.Evaluator,
	})
	if err != nil {
		return nil, fmt.Errorf("[evaluator]new chat model: %w", err)
	}

	evaluator, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "evaluator",
		Description: "面试结果评估者",
		Instruction: "你的职责是根据用户在面试的整体表现(语气、神态、专业能力掌握程度)产出最终评估报告",
		Model:       model,
	})
	if err != nil {
		return nil, fmt.Errorf("[evaluator]new chat model agent: %w", err)
	}

	return evaluator, nil
}
