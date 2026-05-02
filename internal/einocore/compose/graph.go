package compose

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"ai_interview/internal/domain"
)

// GraphInput 图的输入：支持音频或文本输入
type GraphInput struct {
	AudioData   []byte                // 用户语音数据（可选，如果提供则 Supervisor 会调用 ASR Tool）
	Text        string                // 用户文本输入（可选，如果没有 AudioData 则使用此字段）
	InterviewID string                // 面试会话ID
	Stage       domain.InterviewStage // 当前阶段
	Context     map[string]any        // 上下文信息（历史对话、当前问题等）
}

// GraphOutput 图的输出：AI回复 + 新状态
type GraphOutput struct {
	AudioData []byte                // TTS生成的语音（可选，取决于 Supervisor 是否调用 TTS Tool）
	Text      string                // AI回复文本
	NewStage  domain.InterviewStage // 更新后的阶段
	Context   map[string]any        // 更新后的上下文
}

// InterviewGraph 面试流程图（薄包装层）
//
// 设计理念：
// - Graph 只包含一个 Supervisor 节点
// - ASR/TTS 不再是节点，而是 Supervisor 的 Tool（按需调用）
// - Graph 提供流式输出能力和未来的扩展性
type InterviewGraph struct {
	runnable   compose.Runnable[GraphInput, GraphOutput]
	supervisor adk.ResumableAgent
}

// NewInterviewGraph 创建面试流程图
//
// Graph 结构：START → Supervisor → END
// Supervisor 内部会：
// 1. 判断是否需要调用 ASR Tool（如果输入是音频）
// 2. 协调子 Agent 完成决策（question_selector/response_analyzer/stage_manager）
// 3. 判断是否需要调用 TTS Tool（如果需要语音输出）
func NewInterviewGraph(ctx context.Context, supervisor adk.ResumableAgent) (*InterviewGraph, error) {
	ig := &InterviewGraph{
		supervisor: supervisor,
	}

	// 创建图
	g := compose.NewGraph[GraphInput, GraphOutput]()

	// 只添加一个节点：Supervisor
	supervisorLambda := compose.InvokableLambda(ig.callSupervisor)
	_ = g.AddLambdaNode("Supervisor", supervisorLambda)

	// 连接：START → Supervisor → END
	_ = g.AddEdge(compose.START, "Supervisor")
	_ = g.AddEdge("Supervisor", compose.END)

	// 编译图
	runnable, err := g.Compile(ctx,
		compose.WithMaxRunSteps(20), // 最多 20 轮对话（防止无限循环）
	)
	if err != nil {
		return nil, err
	}

	ig.runnable = runnable
	return ig, nil
}

// callSupervisor 调用 Supervisor（Graph 节点的实现）
func (ig *InterviewGraph) callSupervisor(ctx context.Context, input GraphInput) (GraphOutput, error) {
	// 1. 构建系统提示词（根据当前阶段）
	systemPrompt := getSystemPrompt(input.Stage)

	// 2. 构建消息列表（包含历史对话）
	messages := make([]*schema.Message, 0)

	// 添加系统提示词
	messages = append(messages, schema.SystemMessage(systemPrompt))

	// 添加历史对话（从 Context 中获取）
	if input.Context != nil {
		if history, ok := input.Context["history"].([]map[string]string); ok {
			for _, msg := range history {
				role := msg["role"]
				content := msg["content"]

				if role == "user" {
					messages = append(messages, schema.UserMessage(content))
				} else if role == "assistant" {
					messages = append(messages, schema.AssistantMessage(content, nil))
				}
			}
		}
	}

	// 3. 添加当前用户输入
	var userMessage string
	if len(input.AudioData) > 0 {
		// 如果有音频数据，Supervisor 会调用 ASR Tool
		userMessage = "[音频输入，Supervisor 将调用 ASR Tool 处理]"
	} else {
		userMessage = input.Text
	}
	messages = append(messages, schema.UserMessage(userMessage))

	// 4. 调用 Supervisor
	iter := ig.supervisor.Run(ctx, &adk.AgentInput{
		Messages: messages,
	})

	// 5. 收集 Supervisor 的输出
	var responseText string
	for {
		event, hasNext := iter.Next()
		if !hasNext {
			break
		}

		// 提取消息内容
		if event.Output != nil && event.Output.MessageOutput != nil {
			if event.Output.MessageOutput.Message != nil {
				responseText += event.Output.MessageOutput.Message.Content
			}
		}
	}

	// 6. 解析 Supervisor 的输出
	// Supervisor 应该返回 JSON 格式：
	// {
	//   "response": "回复内容",
	//   "need_tts": true/false,
	//   "stage_action": "continue" | "advance" | "finish"
	// }
	output := parseSupervisorOutput(responseText, input.Stage)

	// 7. 更新上下文（添加本轮对话到历史）
	newContext := input.Context
	if newContext == nil {
		newContext = make(map[string]any)
	}

	history, _ := newContext["history"].([]map[string]string)
	history = append(history, map[string]string{
		"role":    "user",
		"content": userMessage,
	})
	history = append(history, map[string]string{
		"role":    "assistant",
		"content": output.Text,
	})
	newContext["history"] = history

	return output, nil
}

// parseSupervisorOutput 解析 Supervisor 的输出
func parseSupervisorOutput(text string, currentStage domain.InterviewStage) GraphOutput {
	// TODO: 实现 JSON 解析
	// 当前简化实现：直接返回文本

	// 判断是否需要切换阶段
	// 简化逻辑：如果输出包含特定关键词，则切换阶段
	newStage := currentStage
	if containsStageAdvanceSignal(text) {
		newStage = getNextStage(currentStage)
	}

	return GraphOutput{
		Text:     text,
		NewStage: newStage,
		Context:  make(map[string]any),
	}
}

// containsStageAdvanceSignal 检查是否包含阶段切换信号
func containsStageAdvanceSignal(text string) bool {
	// TODO: 实现更智能的判断逻辑
	// 当前简化实现：检查关键词
	keywords := []string{
		"进入下一阶段",
		"开始技术问答",
		"开始算法题",
		"面试结束",
	}

	for _, keyword := range keywords {
		if len(text) > 0 && len(keyword) > 0 {
			// 简化判断
			return false
		}
	}
	return false
}

// getNextStage 获取下一个阶段
func getNextStage(current domain.InterviewStage) domain.InterviewStage {
	stageOrder := []domain.InterviewStage{
		domain.StageIntro,
		domain.StageQuestioning,
		domain.StageAlgorithm,
		domain.StageClosing,
	}

	for i, stage := range stageOrder {
		if stage == current && i < len(stageOrder)-1 {
			return stageOrder[i+1]
		}
	}

	return current
}

// Invoke 同步执行图
func (ig *InterviewGraph) Invoke(ctx context.Context, input GraphInput) (GraphOutput, error) {
	return ig.runnable.Invoke(ctx, input)
}

// Stream 流式执行图
func (ig *InterviewGraph) Stream(ctx context.Context, input GraphInput) (*schema.StreamReader[GraphOutput], error) {
	return ig.runnable.Stream(ctx, input)
}
