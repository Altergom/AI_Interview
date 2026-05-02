package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"

	"ai_interview/internal/config"
	"ai_interview/internal/einocore/tools"
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

	// 创建 ASR/TTS Tool
	var asrService tools.ASRService
	var ttsService tools.TTSService

	// 开发环境：使用 Mock 服务
	// 生产环境：使用千问服务
	//useQwenService := false // 临时使用 Mock，实现千问后改为 true
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
		return nil, fmt.Errorf("[tools]NewASRTool: %v", err)
	}

	ttsTool, err := tools.NewTTSTool(ttsService)
	if err != nil {
		return nil, fmt.Errorf("[tools]NewTTSTool: %v", err)
	}

	// 创建 Supervisor Agent（配置 ASR/TTS Tool）
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
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
- question_selector: 从题库选择合适的问题
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
		return nil, fmt.Errorf("[adk]NewChatModelAgent: %v", err)
	}

	// 创建子 Agent
	selector, err := NewSelector()
	if err != nil {
		return nil, err
	}
	manager, err := NewManager()
	if err != nil {
		return nil, err
	}
	analyzer, err := NewAnalyzer()
	if err != nil {
		return nil, err
	}

	// 创建 Supervisor（配置子 Agent）
	su, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: agent,
		SubAgents: []adk.Agent{
			selector,
			manager,
			analyzer,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("[supervisor]New: %v", err)
	}

	return su, nil
}
