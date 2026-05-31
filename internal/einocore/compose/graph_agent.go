package compose

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	einocompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"ai_interview/internal/domain"
)

// NewAgentDrivenGraph 创建 Agent Supervisor 驱动的面试流程图。
//
// Graph 结构：START → Supervisor → END
// Supervisor 内部协调子 Agent（question_selector/response_analyzer/stage_manager）
// 和 Tool（ASR/TTS），阶段切换由 Supervisor 的 JSON 输出中 stage_action 字段决定。
func NewAgentDrivenGraph(ctx context.Context, supervisor adk.ResumableAgent) (*InterviewGraph, error) {
	if supervisor == nil {
		return nil, fmt.Errorf("[graph] supervisor is required")
	}

	ig := &InterviewGraph{
		agents: make(map[domain.InterviewStage]StageAgent),
	}

	g := einocompose.NewGraph[GraphInput, GraphOutput]()

	callSupervisor := func(ctx context.Context, input GraphInput) (GraphOutput, error) {
		userMessage := resolveUserMessage(input)
		messages := buildSupervisorMessages(input, userMessage)

		iter := supervisor.Run(ctx, &adk.AgentInput{Messages: messages})
		var responseText string
		for {
			event, hasNext := iter.Next()
			if !hasNext {
				break
			}
			if event.Output != nil && event.Output.MessageOutput != nil {
				if event.Output.MessageOutput.Message != nil {
					responseText += event.Output.MessageOutput.Message.Content
				}
			}
		}

		output := parseSupervisorOutput(responseText, input.Stage)
		output.Context = appendHistory(input.Context, userMessage, output.Text)
		return output, nil
	}

	_ = g.AddLambdaNode("Supervisor", einocompose.InvokableLambda(callSupervisor))
	_ = g.AddEdge(einocompose.START, "Supervisor")
	_ = g.AddEdge("Supervisor", einocompose.END)

	runnable, err := g.Compile(ctx, einocompose.WithMaxRunSteps(20))
	if err != nil {
		return nil, fmt.Errorf("[graph] compile agent-driven graph: %w", err)
	}
	ig.runnable = runnable
	return ig, nil
}

// buildSupervisorMessages 构建传给 Supervisor 的完整消息列表（Agent 驱动模式专用）。
func buildSupervisorMessages(input GraphInput, userMessage string) []*schema.Message {
	messages := []*schema.Message{schema.SystemMessage(getSystemPrompt(input.Stage))}

	if input.Context != nil {
		if history, ok := input.Context["history"].([]map[string]string); ok {
			for _, msg := range history {
				switch msg["role"] {
				case "user":
					messages = append(messages, schema.UserMessage(msg["content"]))
				case "assistant":
					messages = append(messages, schema.AssistantMessage(msg["content"], nil))
				}
			}
		}
	}

	return append(messages, schema.UserMessage(userMessage))
}
