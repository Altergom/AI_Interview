package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	einocompose "github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"ai_interview/internal/domain"
)

// endedMessage 是面试进入终态后，对后续输入返回的固定提示。
const endedMessage = "面试已结束，无法继续作答。"

// GraphInput 图的输入：支持音频或文本输入。
type GraphInput struct {
	// AudioData 是用户本轮语音输入；是否调用 ASR 由当前阶段 Agent 按需决定。
	AudioData []byte
	// Text 是用户本轮文本输入；当 AudioData 为空时作为主要用户消息。
	Text string
	// InterviewID 是当前面试会话 ID，用于工具、子 Agent 和上下文追踪。
	InterviewID string
	// Stage 是 session 中保存的当前面试阶段，StageRouter 会据此分发到阶段节点。
	Stage domain.InterviewStage
	// Context 是跨轮次共享上下文，包含 history、统计信息、岗位方向和阶段 metadata。
	Context map[string]any
}

// GraphOutput 图的输出：AI 回复 + 新状态。
type GraphOutput struct {
	// AudioData 是 TTS 生成的音频结果；当前阶段化改造先保留字段，具体音频由阶段 Tool 决定。
	AudioData []byte
	// Text 是本轮面试官返回给用户的文本内容。
	Text string
	// NewStage 是 TransitionNode 经过状态机裁决后的新阶段。
	NewStage domain.InterviewStage
	// Context 是合并了阶段 metadata 和本轮 history 后的会话上下文。
	Context map[string]any
}

// StageManagerSuggestion 是 stage_manager Agent 的结构化建议。
type StageManagerSuggestion struct {
	// SuggestedAction 是 stage_manager 对阶段动作的建议，只能是 continue、advance、finish。
	SuggestedAction string `json:"suggested_action"`
	// Reason 解释为什么给出该建议，主要用于日志、调试和报告解释。
	Reason string `json:"reason"`
	// Confidence 是 stage_manager 的主观置信度，不作为状态机最终裁决的唯一依据。
	Confidence float64 `json:"confidence"`
}

// StageResult 是每个阶段 Agent 的统一返回结构。
type StageResult struct {
	// Response 是阶段 Agent 面向候选人的回复文本。
	Response string `json:"response"`
	// NeedTTS 表示当前回复是否希望转成语音输出。
	NeedTTS bool `json:"need_tts"`
	// AgentAction 是阶段 Agent 的动作意图，最终是否切换阶段仍由 TransitionNode 裁决。
	AgentAction string `json:"agent_action"`
	// StageNotes 保存当前阶段内部说明，例如追问依据、阶段观察等，不直接决定转移。
	StageNotes map[string]any `json:"stage_notes"`
	// Metadata 保存需要跨轮次传递的结构化信息，会合并到 session.Context。
	Metadata map[string]any `json:"metadata"`
}

// StageAgent 是阶段节点内部执行单元。生产环境使用 ADK Agent，测试可注入 fake。
type StageAgent interface {
	// Run 执行当前阶段逻辑，并返回统一 StageResult。
	Run(ctx context.Context, input StageInput) (StageResult, error)
}

// StageInput 是阶段 Agent 接收的输入。
type StageInput struct {
	// AudioData 是原始语音输入，阶段 Agent 可按需调用 ASR Tool。
	AudioData []byte
	// Text 是原始文本输入，保留给阶段 Agent 判断输入来源和内容。
	Text string
	// UserMessage 是 Graph 标准化后的用户消息，优先用于构造 LLM 对话。
	UserMessage string
	// InterviewID 是当前面试 ID，便于子 Agent 或 Tool 访问会话相关资源。
	InterviewID string
	// Stage 是当前阶段，调用 question_selector 时也会作为 stage 参数传入语义约束。
	Stage domain.InterviewStage
	// Context 是会话级上下文，包含历史对话、方向、统计数据和之前阶段 metadata。
	Context map[string]any
}

// InterviewGraphConfig 显式 workflow 的依赖配置。
type InterviewGraphConfig struct {
	// IntroAgent 负责自我介绍阶段的引导、背景/项目追问和阶段建议收集。
	IntroAgent StageAgent
	// QuestioningAgent 负责技术问答阶段的回答分析、追问和下一题选择。
	QuestioningAgent StageAgent
	// AlgorithmAgent 负责算法阶段的题目选择和代码/思路分析；本轮不实现 CodeJudge。
	AlgorithmAgent StageAgent
	// ClosingAgent 负责反问、收尾反馈和结束动作建议。
	ClosingAgent StageAgent
}

// InterviewGraph 面试显式 workflow。
//
// Graph 结构：
// START -> StageRouter -> StageBranch(Intro/Questioning/Algorithm/Closing/End)
//
//	-> TransitionNode -> END
type InterviewGraph struct {
	// runnable 是编译后的 Eino Graph，承载实际节点编排和执行。
	runnable einocompose.Runnable[GraphInput, GraphOutput]
	// agents 按阶段保存阶段 Agent，StageRouter 命中节点后由对应节点取用。
	agents map[domain.InterviewStage]StageAgent
}

// workflowState 是 Graph 内部节点之间传递的中间状态。
type workflowState struct {
	// Input 是原始 GraphInput，TransitionNode 需要用它读取当前 stage 和旧 context。
	Input GraphInput
	// UserMessage 是标准化后的用户输入，避免每个阶段节点重复处理音频/文本分支。
	UserMessage string
	// Result 是阶段节点产出的统一结果，供 TransitionNode 做状态机裁决。
	Result StageResult
	// NewStage 是状态机裁决后的阶段，主要用于 Graph 内部传递和调试。
	NewStage domain.InterviewStage
}

// NewInterviewGraph 创建显式 workflow 面试流程图。
func NewInterviewGraph(ctx context.Context, cfg InterviewGraphConfig) (*InterviewGraph, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	ig := &InterviewGraph{
		agents: map[domain.InterviewStage]StageAgent{
			domain.StageIntro:       cfg.IntroAgent,
			domain.StageQuestioning: cfg.QuestioningAgent,
			domain.StageAlgorithm:   cfg.AlgorithmAgent,
			domain.StageClosing:     cfg.ClosingAgent,
		},
	}

	// 这里保留 Eino Graph 作为 workflow 编排器，但不再把所有逻辑塞进单个 Supervisor。
	g := einocompose.NewGraph[GraphInput, GraphOutput]()
	_ = g.AddLambdaNode("StageRouter", einocompose.InvokableLambda(ig.stageRouter))
	_ = g.AddLambdaNode("IntroNode", einocompose.InvokableLambda(ig.introNode))
	_ = g.AddLambdaNode("QuestioningNode", einocompose.InvokableLambda(ig.questioningNode))
	_ = g.AddLambdaNode("AlgorithmNode", einocompose.InvokableLambda(ig.algorithmNode))
	_ = g.AddLambdaNode("ClosingNode", einocompose.InvokableLambda(ig.closingNode))
	_ = g.AddLambdaNode("EndNode", einocompose.InvokableLambda(ig.endNode))
	_ = g.AddLambdaNode("TransitionNode", einocompose.InvokableLambda(ig.transitionNode))

	_ = g.AddEdge(einocompose.START, "StageRouter")
	// StageRouter 后的 branch 让 Graph 结构显式表达阶段节点，而不是只靠 prompt 判断阶段。
	_ = g.AddBranch("StageRouter", einocompose.NewGraphBranch(
		routeStageNode,
		map[string]bool{
			"IntroNode":       true,
			"QuestioningNode": true,
			"AlgorithmNode":   true,
			"ClosingNode":     true,
			"EndNode":         true,
		},
	))
	// 所有阶段节点都汇聚到 TransitionNode，由确定性状态机统一裁决 nextStage。
	_ = g.AddEdge("IntroNode", "TransitionNode")
	_ = g.AddEdge("QuestioningNode", "TransitionNode")
	_ = g.AddEdge("AlgorithmNode", "TransitionNode")
	_ = g.AddEdge("ClosingNode", "TransitionNode")
	_ = g.AddEdge("EndNode", "TransitionNode")
	_ = g.AddEdge("TransitionNode", einocompose.END)

	runnable, err := g.Compile(ctx, einocompose.WithMaxRunSteps(20))
	if err != nil {
		return nil, err
	}
	ig.runnable = runnable
	return ig, nil
}

// validate 在编译 Graph 前检查必需阶段 Agent，避免运行期才发现节点无法执行。
func (cfg InterviewGraphConfig) validate() error {
	required := map[string]StageAgent{
		"IntroAgent":       cfg.IntroAgent,
		"QuestioningAgent": cfg.QuestioningAgent,
		"AlgorithmAgent":   cfg.AlgorithmAgent,
		"ClosingAgent":     cfg.ClosingAgent,
	}
	for name, agent := range required {
		if agent == nil {
			return fmt.Errorf("[graph] %s is required", name)
		}
	}
	return nil
}

// routeStageNode 根据当前 session stage 选择要执行的阶段节点。
func routeStageNode(_ context.Context, state workflowState) (string, error) {
	switch state.Input.Stage {
	case domain.StageIntro:
		return "IntroNode", nil
	case domain.StageQuestioning:
		return "QuestioningNode", nil
	case domain.StageAlgorithm:
		return "AlgorithmNode", nil
	case domain.StageClosing:
		return "ClosingNode", nil
	case domain.StageEnd:
		return "EndNode", nil
	default:
		// 未知阶段直接报错。
		return "", fmt.Errorf("[graph] unknown interview stage: %q", state.Input.Stage)
	}
}

// stageRouter 标准化用户输入，并把原始 GraphInput 包装成 workflowState。
func (ig *InterviewGraph) stageRouter(_ context.Context, input GraphInput) (workflowState, error) {
	return workflowState{
		Input:       input,
		UserMessage: resolveUserMessage(input),
		NewStage:    input.Stage,
	}, nil
}

// introNode 执行自我介绍阶段 Agent。
func (ig *InterviewGraph) introNode(ctx context.Context, state workflowState) (workflowState, error) {
	return ig.callStageAgent(ctx, state, domain.StageIntro)
}

// questioningNode 执行技术问答阶段 Agent。
func (ig *InterviewGraph) questioningNode(ctx context.Context, state workflowState) (workflowState, error) {
	return ig.callStageAgent(ctx, state, domain.StageQuestioning)
}

// algorithmNode 执行算法题阶段 Agent。
func (ig *InterviewGraph) algorithmNode(ctx context.Context, state workflowState) (workflowState, error) {
	return ig.callStageAgent(ctx, state, domain.StageAlgorithm)
}

// closingNode 执行反问/收尾阶段 Agent。
func (ig *InterviewGraph) closingNode(ctx context.Context, state workflowState) (workflowState, error) {
	return ig.callStageAgent(ctx, state, domain.StageClosing)
}

// endNode 处理终态输入：面试结束后不再调用任何 LLM/Agent，避免继续产生面试对话。
func (ig *InterviewGraph) endNode(_ context.Context, state workflowState) (workflowState, error) {
	state.Result = StageResult{
		Response:    endedMessage,
		NeedTTS:     false,
		AgentAction: "continue",
		StageNotes:  map[string]any{"stage": domain.StageEnd.String()},
		Metadata:    map[string]any{"transition_reason": "input received after interview ended"},
	}
	state.NewStage = domain.StageEnd
	return state, nil
}

// callStageAgent 统一调用阶段 Agent，并规范化其输出，避免各阶段重复校验返回结构。
func (ig *InterviewGraph) callStageAgent(ctx context.Context, state workflowState, stage domain.InterviewStage) (workflowState, error) {
	stageAgent, ok := ig.agents[stage]
	if !ok || stageAgent == nil {
		return state, fmt.Errorf("[graph] stage agent not found: %s", stage)
	}
	result, err := stageAgent.Run(ctx, StageInput{
		AudioData:   state.Input.AudioData,
		Text:        state.Input.Text,
		UserMessage: state.UserMessage,
		InterviewID: state.Input.InterviewID,
		Stage:       stage,
		Context:     state.Input.Context,
	})
	if err != nil {
		return state, err
	}
	state.Result = normalizeStageResult(result)
	return state, nil
}

// transitionNode 将阶段 Agent 的动作意图交给状态机裁决，并合并上下文。
func (ig *InterviewGraph) transitionNode(_ context.Context, state workflowState) (GraphOutput, error) {
	nextStage, reason := applyStageTransition(state.Input.Stage, state.Result.AgentAction)
	state.NewStage = nextStage

	// metadata 进入 session.Context 后，后续阶段可以读取结构化面试信息。
	metadata := cloneMap(state.Result.Metadata)
	metadata["transition_reason"] = reason
	metadata["agent_action"] = state.Result.AgentAction
	metadata["current_stage"] = state.Input.Stage.String()
	metadata["next_stage"] = nextStage.String()

	nextContext := mergeContext(state.Input.Context, metadata)
	nextContext = appendHistory(nextContext, state.UserMessage, state.Result.Response)

	return GraphOutput{
		Text:      state.Result.Response,
		AudioData: nil,
		NewStage:  nextStage,
		Context:   nextContext,
	}, nil
}

// applyStageTransition 是确定性状态机的核心规则。
// Agent 只能提出动作意图，真正 nextStage 必须由这里统一裁决。
func applyStageTransition(current domain.InterviewStage, action string) (domain.InterviewStage, string) {
	switch current {
	case domain.StageEnd:
		return domain.StageEnd, "interview already ended"
	case domain.StageClosing:
		if action == "finish" {
			return domain.StageEnd, "closing finished"
		}
		return domain.StageClosing, "closing continues"
	}

	switch action {
	case "advance":
		return getNextStage(current), "agent requested advance"
	case "continue", "finish":
		return current, "agent action keeps current stage"
	default:
		return current, "unknown agent action keeps current stage"
	}
}

// normalizeStageResult 给阶段 Agent 输出补默认值，并把非法 agent_action 收敛为 continue。
func normalizeStageResult(result StageResult) StageResult {
	if result.StageNotes == nil {
		result.StageNotes = make(map[string]any)
	}
	if result.Metadata == nil {
		result.Metadata = make(map[string]any)
	}
	switch result.AgentAction {
	case "continue", "advance", "finish":
	default:
		result.AgentAction = "continue"
	}
	return result
}

// resolveUserMessage 根据输入类型返回用户消息文本。
func resolveUserMessage(input GraphInput) string {
	if len(input.AudioData) > 0 {
		return "[音频输入，阶段节点将按需调用 ASR Tool 处理]"
	}
	return input.Text
}

// buildMessages 构建传给阶段 Agent 的完整消息列表。
func buildMessages(input StageInput, prompt string) []*schema.Message {
	messages := []*schema.Message{schema.SystemMessage(prompt)}
	messages = append(messages, schema.SystemMessage(fmt.Sprintf(
		"当前 stage: %s。调用 question_selector 时必须传入 stage: %s。",
		input.Stage,
		input.Stage,
	)))

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

	return append(messages, schema.UserMessage(input.UserMessage))
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

// mergeContext 将阶段 metadata 合并到会话上下文；同名 key 以新值覆盖旧值。
func mergeContext(ctx map[string]any, metadata map[string]any) map[string]any {
	next := cloneMap(ctx)
	for k, v := range metadata {
		next[k] = v
	}
	return next
}

// cloneMap 复制 map，避免直接修改调用方传入的 Context。
func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// SupervisorResponse 是遗留 Supervisor 的 JSON 响应格式。
type SupervisorResponse struct {
	Response    string `json:"response"`
	NeedTTS     bool   `json:"need_tts"`
	StageAction string `json:"stage_action"`
}

// parseSupervisorOutput 解析遗留 Supervisor 的输出，保留给兼容测试使用。
func parseSupervisorOutput(text string, currentStage domain.InterviewStage) GraphOutput {
	jsonText := extractJSON(text)

	var resp SupervisorResponse
	if err := json.Unmarshal([]byte(jsonText), &resp); err != nil {
		return GraphOutput{
			Text:     text,
			NewStage: currentStage,
			Context:  make(map[string]any),
		}
	}

	newStage := currentStage
	switch resp.StageAction {
	case "advance":
		newStage = getNextStage(currentStage)
	case "finish":
		newStage = domain.StageClosing
	case "continue":
		newStage = currentStage
	default:
		newStage = currentStage
	}

	return GraphOutput{
		Text:      resp.Response,
		AudioData: nil,
		NewStage:  newStage,
		Context:   make(map[string]any),
	}
}

// parseStageResult 解析阶段 Agent 的统一 JSON 输出，失败时降级为继续当前阶段。
func parseStageResult(text string) StageResult {
	jsonText := extractJSON(text)

	var result StageResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return StageResult{
			Response:    text,
			NeedTTS:     false,
			AgentAction: "continue",
			StageNotes:  map[string]any{},
			Metadata:    map[string]any{},
		}
	}
	return normalizeStageResult(result)
}

// extractJSON 从文本中提取 JSON。
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "{") {
		return text
	}

	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.Index(text[start:], "```")
		if end > 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	if strings.Contains(text, "```") {
		start := strings.Index(text, "```") + 3
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

	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")
	if startIdx >= 0 && endIdx > startIdx {
		return strings.TrimSpace(text[startIdx : endIdx+1])
	}
	return text
}

// getNextStage 获取下一个阶段。
func getNextStage(current domain.InterviewStage) domain.InterviewStage {
	stageOrder := []domain.InterviewStage{
		domain.StageIntro,
		domain.StageQuestioning,
		domain.StageAlgorithm,
		domain.StageClosing,
		domain.StageEnd,
	}

	for i, stage := range stageOrder {
		if stage == current && i < len(stageOrder)-1 {
			return stageOrder[i+1]
		}
	}
	return current
}

// Invoke 同步执行图。
func (ig *InterviewGraph) Invoke(ctx context.Context, input GraphInput) (GraphOutput, error) {
	return ig.runnable.Invoke(ctx, input)
}

// Stream 流式执行图。
func (ig *InterviewGraph) Stream(ctx context.Context, input GraphInput) (*schema.StreamReader[GraphOutput], error) {
	return ig.runnable.Stream(ctx, input)
}

// ADKStageAgent 将 ADK Agent 适配为阶段 Agent。
type ADKStageAgent struct {
	// Stage 是该适配器绑定的固定阶段，防止外部误传导致 Agent 阶段语义漂移。
	Stage domain.InterviewStage
	// Prompt 是该阶段的系统提示词，例如 introPrompt、questioningPrompt。
	Prompt string
	// Agent 是底层 ADK ChatModelAgent 或其他 ADK Agent 实现。
	Agent adk.Agent
}

// Run 构造阶段消息、调用底层 ADK Agent，并解析为 StageResult。
func (a *ADKStageAgent) Run(ctx context.Context, input StageInput) (StageResult, error) {
	if a == nil || a.Agent == nil {
		return StageResult{}, fmt.Errorf("[stage_agent] ADK agent is nil")
	}

	input.Stage = a.Stage
	messages := buildMessages(input, a.Prompt)
	iter := a.Agent.Run(ctx, &adk.AgentInput{Messages: messages})
	responseText := collectAgentText(iter)
	return parseStageResult(responseText), nil
}

// collectAgentText 汇总 ADK 事件流里的文本输出。
func collectAgentText(iter *adk.AsyncIterator[*adk.AgentEvent]) string {
	var responseText string
	for {
		event, hasNext := iter.Next()
		if !hasNext {
			break
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		if event.Output.MessageOutput.Message != nil {
			responseText += event.Output.MessageOutput.Message.Content
		}
	}
	return responseText
}
