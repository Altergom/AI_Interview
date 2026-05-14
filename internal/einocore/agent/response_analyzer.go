package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"ai_interview/internal/llm"
)

func NewAnalyzer() (*adk.ChatModelAgent, error) {
	ctx := context.Background()

	model, err := llm.Registry.NewChatModel(ctx, llm.RoleAnalyzer)
	if err != nil {
		return nil, fmt.Errorf("[analyzer]new chat model: %w", err)
	}

	analyzer, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "response_analyzer",
		Description: "回答结果分析者",
		Instruction: "你的职责是分析面试者对问题的回答质量判断是否需要追问深挖或者开启下一题",
		Model:       model,
	})
	if err != nil {
		return nil, fmt.Errorf("[analyzer]new chat model agent: %w", err)
	}

	return analyzer, nil
}
