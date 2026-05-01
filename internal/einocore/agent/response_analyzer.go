package agent

import (
	"ai_interview/internal/config"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
)

func NewAnalyzer() (*adk.ChatModelAgent, error) {
	ctx := context.Background()
	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  config.Cfg.QwenAPIKey,
		BaseURL: config.Cfg.QwenBaseURL,
		Model:   config.Cfg.Analyzer,
	})
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
