package llm

import (
	"context"
	"testing"

	"ai_interview/internal/config"
)

func TestResolveProvider(t *testing.T) {
	tests := []struct {
		modelName string
		want      Provider
	}{
		// qwen
		{"qwen-plus", ProviderQwen},
		{"qwen-max", ProviderQwen},
		{"qwen-turbo", ProviderQwen},
		{"Qwen-Plus", ProviderQwen},

		// openai: gpt 系列 + o 系列
		{"gpt-4o", ProviderOpenAI},
		{"gpt-4-turbo", ProviderOpenAI},
		{"gpt-3.5-turbo", ProviderOpenAI},
		{"o1-preview", ProviderOpenAI},
		{"o1-mini", ProviderOpenAI},
		{"o3-mini", ProviderOpenAI},
		{"o4-mini", ProviderOpenAI},

		// deepseek
		{"deepseek-chat", ProviderDeepSeek},
		{"deepseek-reasoner", ProviderDeepSeek},
		{"DeepSeek-V3", ProviderDeepSeek},

		// doubao
		{"doubao-pro-32k", ProviderDoubao},
		{"doubao-lite-4k", ProviderDoubao},

		// claude
		{"claude-sonnet-4-20250514", ProviderClaude},
		{"claude-3-opus", ProviderClaude},

		// gemini
		{"gemini-2.5-pro", ProviderGemini},
		{"gemini-flash", ProviderGemini},

		// 未匹配 → Qwen
		{"llama-3-70b", ProviderQwen},
		{"mixtral-8x7b", ProviderQwen},
		{"", ProviderQwen},
	}

	for _, tt := range tests {
		t.Run(tt.modelName, func(t *testing.T) {
			got := resolveProvider(tt.modelName)
			if got != tt.want {
				t.Errorf("resolveProvider(%q) = %q, want %q", tt.modelName, got, tt.want)
			}
		})
	}
}

func TestInit(t *testing.T) {
	cfg := &config.Config{
		QwenAPIKey:      "qwen-key",
		QwenBaseURL:     "https://dashscope.aliyuncs.com/compatible-mode/v1",
		OpenAIAPIKey:    "openai-key",
		OpenAIBaseURL:   "https://api.openai.com/v1",
		DeepSeekAPIKey:  "deepseek-key",
		DeepSeekBaseURL: "https://api.deepseek.com/v1",
		ClaudeAPIKey:    "claude-key",
		ClaudeBaseURL:   "https://api.anthropic.com",
		// Doubao 和 Gemini 不设 key，验证未配置的不出现在已配置列表中

		Supervisor: "qwen-plus",
		Selector:   "gpt-4o",
		Manager:    "deepseek-chat",
		Analyzer:   "claude-sonnet-4-20250514",
		Evaluator:  "doubao-pro-32k",
	}

	Init(cfg)

	if Registry == nil {
		t.Fatal("Registry should not be nil after Init")
	}

	// 验证 provider 配置正确加载
	checks := []struct {
		provider Provider
		wantKey  string
		wantURL  string
	}{
		{ProviderQwen, "qwen-key", "https://dashscope.aliyuncs.com/compatible-mode/v1"},
		{ProviderOpenAI, "openai-key", "https://api.openai.com/v1"},
		{ProviderDeepSeek, "deepseek-key", "https://api.deepseek.com/v1"},
		{ProviderClaude, "claude-key", "https://api.anthropic.com"},
	}

	for _, c := range checks {
		got, ok := Registry.providers[c.provider]
		if !ok {
			t.Errorf("provider %q not found in registry", c.provider)
			continue
		}
		if got.APIKey != c.wantKey {
			t.Errorf("provider %q APIKey = %q, want %q", c.provider, got.APIKey, c.wantKey)
		}
		if got.BaseURL != c.wantURL {
			t.Errorf("provider %q BaseURL = %q, want %q", c.provider, got.BaseURL, c.wantURL)
		}
	}

	// 验证角色模型名映射正确
	roleChecks := []struct {
		role      AgentRole
		wantModel string
	}{
		{RoleSupervisor, "qwen-plus"},
		{RoleSelector, "gpt-4o"},
		{RoleManager, "deepseek-chat"},
		{RoleAnalyzer, "claude-sonnet-4-20250514"},
		{RoleEvaluator, "doubao-pro-32k"},
	}
	for _, rc := range roleChecks {
		got := Registry.roleModel[rc.role]
		if got != rc.wantModel {
			t.Errorf("role %q model = %q, want %q", rc.role, got, rc.wantModel)
		}
	}

	// 验证 API key 为空的 provider 不会出现在已配置列表（日志不报但 config 有值）
	_, doubaoExists := Registry.providers[ProviderDoubao]
	if !doubaoExists {
		t.Error("ProviderDoubao should exist in providers map even if APIKey is empty")
	}
	_, geminiExists := Registry.providers[ProviderGemini]
	if !geminiExists {
		t.Error("ProviderGemini should exist in providers map even if APIKey is empty")
	}
}

func TestNewChatModel_UnknownRole(t *testing.T) {
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{},
		roleModel: map[AgentRole]string{
			RoleSupervisor: "qwen-plus",
		},
	}

	_, err := r.NewChatModel(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown role, got nil")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestNewChatModel_EmptyModelName(t *testing.T) {
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{},
		roleModel: map[AgentRole]string{
			RoleEvaluator: "",
		},
	}

	_, err := r.NewChatModel(context.Background(), RoleEvaluator)
	if err == nil {
		t.Fatal("expected error for empty model name, got nil")
	}
}

func TestNewChatModel_MissingAPIKey(t *testing.T) {
	// 所有 provider 的 APIKey 均为空 → 无论路由到哪个都报错
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{
			ProviderQwen:     {APIKey: "", BaseURL: ""},
			ProviderOpenAI:   {APIKey: "", BaseURL: ""},
			ProviderDeepSeek: {APIKey: "", BaseURL: ""},
			ProviderDoubao:   {APIKey: "", BaseURL: ""},
			ProviderClaude:   {APIKey: "", BaseURL: ""},
			ProviderGemini:   {APIKey: "", BaseURL: ""},
		},
		roleModel: map[AgentRole]string{
			RoleSupervisor: "qwen-plus",
			RoleSelector:   "gpt-4o",
			RoleManager:    "deepseek-chat",
			RoleAnalyzer:   "claude-sonnet-4-20250514",
			RoleEvaluator:  "doubao-pro-32k",
		},
	}

	ctx := context.Background()
	roles := []AgentRole{RoleSupervisor, RoleSelector, RoleManager, RoleAnalyzer, RoleEvaluator}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			_, err := r.NewChatModel(ctx, role)
			if err == nil {
				t.Errorf("expected error for role %q with missing API key, got nil", role)
			}
		})
	}
}

func TestNewChatModelWithProvider_UnsupportedProvider(t *testing.T) {
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{},
		roleModel: map[AgentRole]string{},
	}

	_, err := r.NewChatModelWithProvider(context.Background(), "unknown-provider", "model")
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

func TestNewChatModelWithProvider_MissingAPIKey(t *testing.T) {
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{
			ProviderQwen:     {APIKey: "", BaseURL: ""},
			ProviderOpenAI:   {APIKey: "", BaseURL: ""},
			ProviderDeepSeek: {APIKey: "", BaseURL: ""},
			ProviderDoubao:   {APIKey: "", BaseURL: ""},
			ProviderClaude:   {APIKey: "", BaseURL: ""},
			ProviderGemini:   {APIKey: "", BaseURL: ""},
		},
		roleModel: map[AgentRole]string{},
	}

	ctx := context.Background()
	allProviders := []Provider{ProviderQwen, ProviderOpenAI, ProviderDeepSeek, ProviderDoubao, ProviderClaude, ProviderGemini}

	for _, p := range allProviders {
		t.Run(string(p), func(t *testing.T) {
			_, err := r.NewChatModelWithProvider(ctx, p, "test-model")
			if err == nil {
				t.Errorf("expected error for provider %q with missing API key, got nil", p)
			}
		})
	}
}

func TestResolveProvider_AllRolesRouteCorrectly(t *testing.T) {
	// 端到端验证：角色 → 模型名 → provider 的完整链路
	r := &ProviderRegistry{
		providers: map[Provider]providerCfg{
			ProviderQwen:     {APIKey: "k1", BaseURL: "u1"},
			ProviderOpenAI:   {APIKey: "k2", BaseURL: "u2"},
			ProviderDeepSeek: {APIKey: "k3", BaseURL: "u3"},
			ProviderDoubao:   {APIKey: "k4", BaseURL: "u4"},
			ProviderClaude:   {APIKey: "k5", BaseURL: "u5"},
			ProviderGemini:   {APIKey: "k6", BaseURL: "u6"},
		},
		roleModel: map[AgentRole]string{
			RoleSupervisor: "qwen-plus",
			RoleSelector:   "gpt-4o",
			RoleManager:    "deepseek-chat",
			RoleAnalyzer:   "claude-sonnet-4-20250514",
			RoleEvaluator:  "doubao-pro-32k",
		},
	}

	tests := []struct {
		role     AgentRole
		wantProv Provider
	}{
		{RoleSupervisor, ProviderQwen},
		{RoleSelector, ProviderOpenAI},
		{RoleManager, ProviderDeepSeek},
		{RoleAnalyzer, ProviderClaude},
		{RoleEvaluator, ProviderDoubao},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			modelName := r.roleModel[tt.role]
			got := resolveProvider(modelName)
			if got != tt.wantProv {
				t.Errorf("role %q model=%q resolved to %q, want %q", tt.role, modelName, got, tt.wantProv)
			}
		})
	}
}
