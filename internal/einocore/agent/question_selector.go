package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"ai_interview/internal/llm"
)

func NewSelector() (*adk.ChatModelAgent, error) {
	ctx := context.Background()

	model, err := llm.Registry.NewChatModel(ctx, llm.RoleSelector)
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
