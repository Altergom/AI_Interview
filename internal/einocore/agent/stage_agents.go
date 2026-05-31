package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/tools"
	"ai_interview/internal/llm"
)

// StageAgentsConfig 是显式 workflow 阶段 Agent 的构造参数。
type StageAgentsConfig struct {
	// SelectorCfg 是 question_selector 的构造配置，包含 skills 目录、Redis 去重客户端和 TTL。
	SelectorCfg SelectorConfig
}

// StageAgents 按面试阶段装配 ADK Agent。
type StageAgents struct {
	// Intro 是自我介绍阶段 Agent。
	Intro adk.Agent
	// Questioning 是技术问答阶段 Agent。
	Questioning adk.Agent
	// Algorithm 是算法题阶段 Agent。
	Algorithm adk.Agent
	// Closing 是反问/收尾阶段 Agent。
	Closing adk.Agent
}

// NewStageAgents 创建 intro/questioning/algorithm/closing 四个阶段 Agent。
func NewStageAgents(ctx context.Context, cfg StageAgentsConfig) (*StageAgents, error) {
	tools, err := buildVoiceTools()
	if err != nil {
		return nil, err
	}

	// 子 Agent 仍可跨阶段复用，但由阶段 Agent 决定是否持有和如何调用。
	selector, err := NewSelector(ctx, cfg.SelectorCfg)
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new selector: %w", err)
	}
	mgr, err := NewManager()
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new manager: %w", err)
	}
	analyzer, err := NewAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new analyzer: %w", err)
	}

	// intro 阶段需要语音交互、背景/项目追问和阶段建议。
	intro, err := newStageAgent(ctx, stageAgentSpec{
		Stage:       domain.StageIntro,
		Name:        "intro_agent",
		Description: "自我介绍阶段 Agent，负责背景和项目追问",
		Tools:       tools,
		SubAgents:   []adk.Agent{selector, mgr},
	})
	if err != nil {
		return nil, err
	}
	// questioning 阶段需要回答分析、技术追问、下一题选择和阶段建议。
	questioning, err := newStageAgent(ctx, stageAgentSpec{
		Stage:       domain.StageQuestioning,
		Name:        "questioning_agent",
		Description: "技术问答阶段 Agent，负责技术问题追问和下一题选择",
		Tools:       tools,
		SubAgents:   []adk.Agent{selector, analyzer, mgr},
	})
	if err != nil {
		return nil, err
	}
	// algorithm 阶段本次不接 CodeJudge，仅保留题目选择、代码/思路分析和阶段建议。
	algorithm, err := newStageAgent(ctx, stageAgentSpec{
		Stage:       domain.StageAlgorithm,
		Name:        "algorithm_agent",
		Description: "算法阶段 Agent，负责算法题选择和代码思路分析",
		Tools:       nil,
		SubAgents:   []adk.Agent{selector, analyzer, mgr},
	})
	if err != nil {
		return nil, err
	}
	// closing 阶段只保留收尾所需能力，避免继续调用技术问答类分析器。
	closing, err := newStageAgent(ctx, stageAgentSpec{
		Stage:       domain.StageClosing,
		Name:        "closing_agent",
		Description: "反问和收尾阶段 Agent，负责回答候选人问题并确认结束",
		Tools:       tools,
		SubAgents:   []adk.Agent{mgr},
	})
	if err != nil {
		return nil, err
	}

	return &StageAgents{
		Intro:       intro,
		Questioning: questioning,
		Algorithm:   algorithm,
		Closing:     closing,
	}, nil
}

type stageAgentSpec struct {
	// Stage 是阶段 Agent 绑定的面试阶段。
	Stage domain.InterviewStage
	// Name 是 ADK Agent 名称，便于事件追踪和子 Agent 识别。
	Name string
	// Description 描述该阶段 Agent 能力，会被 ADK 用于 Agent 说明。
	Description string
	// Tools 是该阶段可调用的 Tool，例如 ASR/TTS。
	Tools []tool.BaseTool
	// SubAgents 是该阶段可协调的子 Agent。
	SubAgents []adk.Agent
}

// newStageAgent 根据阶段配置创建 ADK ChatModelAgent。
func newStageAgent(ctx context.Context, spec stageAgentSpec) (*adk.ChatModelAgent, error) {
	model, err := llm.Registry.NewChatModel(ctx, llm.RoleSupervisor)
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new chat model %s: %w", spec.Name, err)
	}

	// 这里的统一 instruction 约束所有阶段 Agent 返回同一 StageResult JSON，
	// 这样 Graph 的 TransitionNode 可以用一套状态机逻辑处理所有阶段。
	instruction := fmt.Sprintf(`你是 %s。

当前阶段：%s

你只负责当前阶段的面试执行，可以按需调用可用 Tool 和 SubAgent。

协作规则：
- 如果输入是音频，按需调用 ASR Tool。
- 如果需要语音输出，按需调用 TTS Tool，并设置 need_tts。
- 调用 question_selector 时必须传入 stage: %s。
- 调用 stage_manager 后，只把它的结构化建议写入 metadata.stage_manager_suggestion；最终阶段裁决由外部 StateMachine 完成。
- 不要输出 JSON 以外的任何内容。

统一输出格式：
{
  "response": "回复内容",
  "need_tts": true,
  "agent_action": "continue",
  "stage_notes": {},
  "metadata": {
    "stage_manager_suggestion": {
      "suggested_action": "continue",
      "reason": "原因",
      "confidence": 0.8
    }
  }
}`, spec.Description, spec.Stage, spec.Stage)

	cfg := &adk.ChatModelAgentConfig{
		Name:        spec.Name,
		Description: spec.Description,
		Instruction: instruction,
		Model:       model,
	}
	if len(spec.Tools) > 0 {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: spec.Tools},
		}
	}
	agent, err := adk.NewChatModelAgent(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new %s: %w", spec.Name, err)
	}
	if len(spec.SubAgents) > 0 {
		// ADK 的 ChatModelAgent 通过 OnSetSubAgents 挂载子 Agent，不在 Config 中直接声明。
		if err := agent.OnSetSubAgents(ctx, spec.SubAgents); err != nil {
			return nil, fmt.Errorf("[stage_agents] set sub agents for %s: %w", spec.Name, err)
		}
	}
	return agent, nil
}

// buildVoiceTools 创建 ASR/TTS Tool，供需要语音交互的阶段 Agent 按需调用。
func buildVoiceTools() ([]tool.BaseTool, error) {
	var asrService tools.ASRService
	var ttsService tools.TTSService
	if needsMock() {
		asrService = tools.NewMockASRService()
		ttsService = tools.NewMockTTSService()
	} else {
		asrService = tools.NewQwenASRService()
		ttsService = tools.NewQwenTTSService()
	}

	asrTool, err := tools.NewASRTool(asrService)
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new asr tool: %w", err)
	}
	ttsTool, err := tools.NewTTSTool(ttsService)
	if err != nil {
		return nil, fmt.Errorf("[stage_agents] new tts tool: %w", err)
	}
	return []tool.BaseTool{asrTool, ttsTool}, nil
}
