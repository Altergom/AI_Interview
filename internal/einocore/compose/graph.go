package compose

import (
	"context"
	"encoding/json"
	"strings"

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
	userMessage := resolveUserMessage(input)
	messages := buildMessages(input, userMessage)

	iter := ig.supervisor.Run(ctx, &adk.AgentInput{Messages: messages})
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

// resolveUserMessage 根据输入类型返回用户消息文本。
func resolveUserMessage(input GraphInput) string {
	if len(input.AudioData) > 0 {
		return "[音频输入，Supervisor 将调用 ASR Tool 处理]"
	}
	return input.Text
}

// buildMessages 构建传给 Supervisor 的完整消息列表。
func buildMessages(input GraphInput, userMessage string) []*schema.Message {
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

// appendHistory 将本轮对话追加到上下文历史并返回新上下文。
func appendHistory(ctx map[string]any, userMessage, aiResponse string) map[string]any {
	if ctx == nil {
		ctx = make(map[string]any)
	}
	history, _ := ctx["history"].([]map[string]string)
	history = append(history,
		map[string]string{"role": "user", "content": userMessage},
		map[string]string{"role": "assistant", "content": aiResponse},
	)
	ctx["history"] = history
	return ctx
}

// SupervisorResponse Supervisor 的 JSON 响应格式
type SupervisorResponse struct {
	Response    string `json:"response"`     // 回复内容
	NeedTTS     bool   `json:"need_tts"`     // 是否需要语音输出
	StageAction string `json:"stage_action"` // 阶段动作：continue | advance | finish
}

// parseSupervisorOutput 解析 Supervisor 的输出
func parseSupervisorOutput(text string, currentStage domain.InterviewStage) GraphOutput {
	// 1. 尝试提取 JSON（可能包含在 markdown 代码块中）
	jsonText := extractJSON(text)

	// 2. 尝试解析 JSON
	var resp SupervisorResponse
	if err := json.Unmarshal([]byte(jsonText), &resp); err != nil {
		// 解析失败，回退到简单实现
		// 这样即使 LLM 偶尔不返回 JSON，系统也能继续工作
		return GraphOutput{
			Text:     text,
			NewStage: currentStage,
			Context:  make(map[string]any),
		}
	}

	// 3. 根据 stage_action 决定新阶段
	newStage := currentStage
	switch resp.StageAction {
	case "advance":
		// 进入下一阶段
		newStage = getNextStage(currentStage)
	case "finish":
		// 面试结束（保持在 closing 阶段）
		newStage = domain.StageClosing
	case "continue":
		// 继续当前阶段
		newStage = currentStage
	default:
		// 未知动作，保持当前阶段
		newStage = currentStage
	}

	return GraphOutput{
		Text:      resp.Response,
		AudioData: nil, // TTS 由 Supervisor 调用 Tool 处理
		NewStage:  newStage,
		Context:   make(map[string]any),
	}
}

// extractJSON 从文本中提取 JSON
// 处理以下情况：
// 1. 纯 JSON：{"response": "..."}
// 2. Markdown 代码块：```json\n{...}\n```
// 3. 普通代码块：```\n{...}\n```
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	// 情况 1：如果以 { 开头，可能是纯 JSON
	if strings.HasPrefix(text, "{") {
		return text
	}

	// 情况 2：包含 ```json ... ```
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.Index(text[start:], "```")
		if end > 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// 情况 3：包含 ``` ... ```
	if strings.Contains(text, "```") {
		start := strings.Index(text, "```") + 3
		// 跳过可能的语言标识符（如 json, JSON）
		if start < len(text) && text[start] != '\n' {
			newlineIdx := strings.Index(text[start:], "\n")
			if newlineIdx > 0 {
				start += newlineIdx + 1
			}
		}
		end := strings.Index(text[start:], "```")
		if end > 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// 情况 4：尝试查找 JSON 对象
	// 查找第一个 { 和最后一个 }
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		return strings.TrimSpace(text[startIdx : endIdx+1])
	}

	// 无法提取，返回原文本
	return text
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
