// Package llm 提供统一的 LLM provider 注册与 ChatModel 创建接口。
//
// v2 设计：通过模型名前缀自动匹配 provider，无需显式配置 provider 名。
// 配置只需设置各 provider 的 APIKey/BaseURL 和每角色的模型名。
package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
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
	ProviderClaude   Provider = "claude"
	ProviderGemini   Provider = "gemini"
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

// modelProviderMapping 定义模型名前缀到 provider 的映射。
type modelProviderMapping struct {
	prefix   string
	provider Provider
}

var modelPrefixes = []modelProviderMapping{
	{"qwen", ProviderQwen},
	{"gpt", ProviderOpenAI},
	{"o1", ProviderOpenAI},
	{"o3", ProviderOpenAI},
	{"o4", ProviderOpenAI},
	{"deepseek", ProviderDeepSeek},
	{"doubao", ProviderDoubao},
	{"claude", ProviderClaude},
	{"gemini", ProviderGemini},
}

// resolveProvider 根据模型名匹配 provider。匹配失败时回退到 Qwen。
func resolveProvider(modelName string) Provider {
	lower := strings.ToLower(modelName)
	for _, m := range modelPrefixes {
		if strings.HasPrefix(lower, m.prefix) {
			return m.provider
		}
	}
	return ProviderQwen
}

// ProviderRegistry 持有各 provider 配置，提供统一的 ChatModel 构造方法。
type ProviderRegistry struct {
	providers map[Provider]providerCfg
	roleModel map[AgentRole]string
}

// Registry 是全局单例，在 app 启动时由 Init 初始化。
var Registry *ProviderRegistry

// Init 从 config.Cfg 加载各 provider 配置，构建 Registry 单例。
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
// 模型名前缀自动匹配 provider：
//   - qwen-* → Qwen
//   - gpt-*/o1/o3/o4 → OpenAI
//   - deepseek-* → DeepSeek
//   - doubao-* → Doubao
//   - claude-* → Claude
//   - gemini-* → Gemini
//
// 未匹配时回退到 Qwen。
func (r *ProviderRegistry) NewChatModel(ctx context.Context, role AgentRole) (einomodel.ChatModel, error) {
	modelName, ok := r.roleModel[role]
	if !ok {
		return nil, fmt.Errorf("[LLM] unknown agent role: %q", role)
	}
	if modelName == "" {
		return nil, fmt.Errorf("[LLM] model name for role %q is empty", role)
	}

	return r.NewChatModelWithProvider(ctx, resolveProvider(modelName), modelName)
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

// newOpenAICompatModel 为 OpenAI 兼容 provider 创建 ChatModel。
// 适用于 OpenAI、DeepSeek、Doubao、Gemini。
func (r *ProviderRegistry) newOpenAICompatModel(ctx context.Context, provider Provider, modelName string) (einomodel.ChatModel, error) {
	cfg, ok := r.providers[provider]
	if !ok || cfg.APIKey == "" {
		return nil, fmt.Errorf("[LLM] %s provider not configured: %s_API_KEY is empty", provider, strings.ToUpper(string(provider)))
	}

	model, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   modelName,
	})
	if err != nil {
		return nil, fmt.Errorf("[LLM] new %s chat model %q: %w", provider, modelName, err)
	}
	return model, nil
}

// newClaudeModel 为 Anthropic Claude 创建 ChatModel。
func (r *ProviderRegistry) newClaudeModel(ctx context.Context, modelName string) (einomodel.ChatModel, error) {
	cfg, ok := r.providers[ProviderClaude]
	if !ok || cfg.APIKey == "" {
		return nil, fmt.Errorf("[LLM] claude provider not configured: CLAUDE_API_KEY is empty")
	}

	baseURL := cfg.BaseURL
	model, err := claude.NewChatModel(ctx, &claude.Config{
		APIKey:    cfg.APIKey,
		BaseURL:   &baseURL,
		Model:     modelName,
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("[LLM] new claude chat model %q: %w", modelName, err)
	}
	return model, nil
}

// NewChatModelWithProvider 显式指定 provider 和模型名创建 ChatModel。
func (r *ProviderRegistry) NewChatModelWithProvider(ctx context.Context, provider Provider, modelName string) (einomodel.ChatModel, error) {
	switch provider {
	case ProviderQwen:
		return r.newQwenModel(ctx, modelName)
	case ProviderClaude:
		return r.newClaudeModel(ctx, modelName)
	case ProviderOpenAI, ProviderDeepSeek, ProviderDoubao, ProviderGemini:
		return r.newOpenAICompatModel(ctx, provider, modelName)
	default:
		return nil, fmt.Errorf("[LLM] unsupported provider %q", provider)
	}
}
