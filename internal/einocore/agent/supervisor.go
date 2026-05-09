package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"ai_interview/internal/einocore/tools"
	"ai_interview/internal/llm"
)

// SupervisorConfig Supervisor 构造参数，外部依赖从 app 层注入。
type SupervisorConfig struct {
	SelectorCfg SelectorConfig
}

// NewSupervisor 创建面试 Supervisor Agent。
// 所有外部依赖（Redis、SkillsDir）通过 cfg 注入，避免内部硬编码。
func NewSupervisor(ctx context.Context, cfg SupervisorConfig) (adk.ResumableAgent, error) {
	model, err := llm.Registry.NewChatModel(ctx, llm.RoleSupervisor)
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new chat model: %w", err)
	}

	// ASR/TTS Tool
	var asrService tools.ASRService
	var ttsService tools.TTSService
	useQwenService := true
	if useQwenService {
		asrService = tools.NewQwenASRService()
		ttsService = tools.NewQwenTTSService()
	} else {
		asrService = tools.NewMockASRService()
		ttsService = tools.NewMockTTSService()
	}

	asrTool, err := tools.NewASRTool(asrService)
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new asr tool: %w", err)
	}
	ttsTool, err := tools.NewTTSTool(ttsService)
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new tts tool: %w", err)
	}

	supervisorAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "面试流程监管者，负责协调各个子 Agent 和工具",
		Instruction: `你是面试流程的监管者。

你的职责：
1. 理解当前面试阶段和候选人的输入
2. 决定需要调用哪个子 Agent 或工具
3. 汇总结果并返回

可用的工具：
- ASR: 将语音转为文字（当收到音频输入时使用）
- TTS: 将文字转为语音（当需要语音输出时使用）

可用的子 Agent：
- question_selector: 根据候选人方向加载 Skill 规则，生成技术面试题
- response_analyzer: 分析候选人的回答质量
- stage_manager: 判断是否应该切换面试阶段

重要：你是协调者，不是执行者。不要自己生成问题或分析回答，而是调用对应的子 Agent。`,
		Model: model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{asrTool, ttsTool},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new chat model agent: %w", err)
	}

	// 子 Agent：selector 接受外部注入的 SelectorConfig
	selector, err := NewSelector(ctx, cfg.SelectorCfg)
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new selector: %w", err)
	}
	mgr, err := NewManager()
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new manager: %w", err)
	}
	analyzer, err := NewAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("[supervisor] new analyzer: %w", err)
	}

	su, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: supervisorAgent,
		SubAgents:  []adk.Agent{selector, mgr, analyzer},
	})
	if err != nil {
		return nil, fmt.Errorf("[supervisor] supervisor.New: %w", err)
	}

	return su, nil
}
