package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"ai_interview/internal/llm"
)

func NewManager() (*adk.ChatModelAgent, error) {
	ctx := context.Background()

	model, err := llm.Registry.NewChatModel(ctx, llm.RoleManager)
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
