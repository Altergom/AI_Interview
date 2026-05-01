package agent

import (
	"ai_interview/internal/config"
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
)

func NewSupervisor() (adk.ResumableAgent, error) {
	ctx := context.Background()

	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  config.Cfg.QwenAPIKey,
		Model:   config.Cfg.Supervisor,
		BaseURL: config.Cfg.QwenBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("[qwen]NewChatModel: %v", err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "监管范式负责审批全局的监管者",
		Instruction: "你的职责是根据程序运行中各接口的数据返回判断我们是否要进行下一步",
		Model:       model,
	})

	selector, _ := NewSelector()
	manager, _ := NewManager()
	analyzer, _ := NewAnalyzer()
	su, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: agent,
		SubAgents: []adk.Agent{
			selector,
			manager,
			analyzer,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("[supervisor]NewChatModelAgent: %v", err)
	}

	return su, nil
}
