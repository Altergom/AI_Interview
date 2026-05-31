package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"

	"ai_interview/internal/llm"
)

// NewManager 创建阶段管理建议 Agent。
// 它只给出结构化建议，最终阶段切换由 compose.StateMachine/TransitionNode 裁决。
func NewManager() (*adk.ChatModelAgent, error) {
	ctx := context.Background()

	model, err := llm.Registry.NewChatModel(ctx, llm.RoleManager)
	if err != nil {
		return nil, fmt.Errorf("[manager]new chat model: %w", err)
	}

	manager, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "stage_manager",
		Description: "面试状态管理者",
		Instruction: `你的职责是分析面试状态，并给出阶段动作建议。

你只提供建议，不做最终裁决；最终阶段切换由外部 StateMachine 决定。

必须只返回 JSON：
{
  "suggested_action": "continue",
  "reason": "候选人尚未介绍主要项目经验",
  "confidence": 0.8
}

suggested_action 只能是 continue、advance、finish。
confidence 是你对建议的主观置信度，范围 0 到 1。`,
		Model: model,
	})
	if err != nil {
		return nil, fmt.Errorf("[manager]new chat model agent: %w", err)
	}

	return manager, nil
}
