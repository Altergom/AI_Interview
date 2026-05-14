// Package llm 提供统一的 LLM provider 注册与 ChatModel 创建接口。
//
// v1 设计：编译期静态配置，从 .env 读取各 provider 的 APIKey / BaseURL；
// 不做运行时热切换，不做 provider 路由复杂逻辑。
// 使用方式：
//
//	model, err := llm.Registry.NewChatModel(ctx, llm.RoleSupervisor)
package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/qwen"
	einomodel "github.com/cloudwego/eino/components/model"

	"ai_interview/internal/config"
	"ai_interview/internal/log"
)

// Provider 表示 LLM 服务提供商。
type Provider string

const (
	ProviderQwen     Provider = "qwen"
	ProviderOpenAI   Provider = "openai"
	ProviderDoubao   Provider = "doubao"
	ProviderDeepSeek Provider = "deepseek"
	ProviderClaude   Provider = "claude" // 通过 openai 兼容模式接入
	ProviderGemini   Provider = "gemini" // 通过 openai 兼容模式接入
)

// AgentRole 标识调用方是哪个 Agent，Registry 据此从配置中选模型名。
type AgentRole string

const (
	RoleSupervisor AgentRole = "supervisor"
	RoleSelector   AgentRole = "selector"
	RoleManager    AgentRole = "manager"
	RoleAnalyzer   AgentRole = "analyzer"
	RoleEvaluator  AgentRole = "evaluator"
)

// providerCfg 单个 provider 的连接参数。
type providerCfg struct {
	APIKey  string
	BaseURL string
}

// ProviderRegistry 持有各 provider 配置，提供统一的 ChatModel 构造方法。
//
// v1 限制：
//   - 只支持 Qwen（其余 provider 配置加载但暂不路由）
//   - 所有 Agent 的 provider 均为 qwen，通过 AgentRole 区分模型名
//   - v2 可扩展为按 role 指定不同 provider
type ProviderRegistry struct {
	providers map[Provider]providerCfg
	// roleModel 从 role 映射到模型名，v1 全走 qwen
	roleModel map[AgentRole]string
}

// Registry 是全局单例，在 app 启动时由 Init 初始化。
var Registry *ProviderRegistry

// Init 从 config.Cfg 加载各 provider 配置，构建 Registry 单例。
// 必须在 config.Load() 之后调用。
func Init(cfg *config.Config) {
	providers := map[Provider]providerCfg{
		ProviderQwen: {
			APIKey:  cfg.QwenAPIKey,
			BaseURL: cfg.QwenBaseURL,
		},
		ProviderOpenAI: {
			APIKey:  cfg.OpenAIAPIKey,
			BaseURL: cfg.OpenAIBaseURL,
		},
		ProviderDoubao: {
			APIKey:  cfg.DoubaoAPIKey,
			BaseURL: cfg.DoubaoBaseURL,
		},
		ProviderDeepSeek: {
			APIKey:  cfg.DeepSeekAPIKey,
			BaseURL: cfg.DeepSeekBaseURL,
		},
		ProviderClaude: {
			APIKey:  cfg.ClaudeAPIKey,
			BaseURL: cfg.ClaudeBaseURL,
		},
		ProviderGemini: {
			APIKey:  cfg.GenimiAPIKey,
			BaseURL: cfg.GenimiBaseURL,
		},
	}

	roleModel := map[AgentRole]string{
		RoleSupervisor: cfg.Supervisor,
		RoleSelector:   cfg.Selector,
		RoleManager:    cfg.Manager,
		RoleAnalyzer:   cfg.Analyzer,
		RoleEvaluator:  cfg.Evaluator,
	}

	Registry = &ProviderRegistry{
		providers: providers,
		roleModel: roleModel,
	}

	// 启动时打印已配置的 provider，便于排查
	var configured []string
	for p, c := range providers {
		if c.APIKey != "" {
			configured = append(configured, string(p))
		}
	}
	log.Infof("[LLM] registry initialized, configured providers: [%s]", strings.Join(configured, ", "))
}

// NewChatModel 根据 AgentRole 创建对应的 ChatModel。
//
// v1：所有 role 均使用 Qwen provider；模型名从 config 读取（SUPERVISOR / SELECTOR 等环境变量）。
// 若 Qwen APIKey 未配置，返回错误，启动即失败，避免运行时 panic。
func (r *ProviderRegistry) NewChatModel(ctx context.Context, role AgentRole) (einomodel.ChatModel, error) {
	modelName, ok := r.roleModel[role]
	if !ok {
		return nil, fmt.Errorf("[LLM] unknown agent role: %q", role)
	}
	if modelName == "" {
		return nil, fmt.Errorf("[LLM] model name for role %q is empty", role)
	}

	// v1 固定走 Qwen
	return r.newQwenModel(ctx, modelName)
}

// newQwenModel 创建 Qwen ChatModel。
func (r *ProviderRegistry) newQwenModel(ctx context.Context, modelName string) (einomodel.ChatModel, error) {
	cfg, ok := r.providers[ProviderQwen]
	if !ok || cfg.APIKey == "" {
		return nil, fmt.Errorf("[LLM] qwen provider not configured: QWEN_API_KEY is empty")
	}

	model, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("[LLM] new qwen chat model %q: %w", modelName, err)
	}
	return model, nil
}

// NewChatModelWithProvider 显式指定 provider 和模型名创建 ChatModel。
// 适用于需要跨 provider 调用的场景（如 embedding 用 qwen、eval 用 deepseek）。
// v1 只实现 Qwen，其余 provider 返回 ErrNotImplemented。
func (r *ProviderRegistry) NewChatModelWithProvider(ctx context.Context, provider Provider, modelName string) (einomodel.ChatModel, error) {
	switch provider {
	case ProviderQwen:
		return r.newQwenModel(ctx, modelName)
	default:
		return nil, fmt.Errorf("[LLM] provider %q not implemented in v1; only qwen is supported", provider)
	}
}
