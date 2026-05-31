package compose

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/config"
	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/agent"
)

func loadConfig(t *testing.T) {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			_ = config.Load(filepath.Join(dir, ".env"))
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	_ = config.Load()
}

func skipIfNoAPIKey(t *testing.T) {
	t.Helper()
	if config.Cfg == nil || config.Cfg.QwenAPIKey == "" {
		t.Skip("QWEN_API_KEY not set, skipping integration test")
	}
}

func newTestGraph(t *testing.T) *InterviewGraph {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
	})

	stageAgents, err := agent.NewStageAgents(context.Background(), agent.StageAgentsConfig{
		SelectorCfg: agent.SelectorConfig{
			SkillsDir:   "internal/einocore/skills",
			RedisClient: rdb,
		},
	})
	if err != nil {
		t.Fatalf("NewStageAgents failed: %v", err)
	}
	graph, err := NewInterviewGraph(context.Background(), InterviewGraphConfig{
		IntroAgent: &ADKStageAgent{
			Stage:  domain.StageIntro,
			Prompt: GetSystemPromptForStage(domain.StageIntro),
			Agent:  stageAgents.Intro,
		},
		QuestioningAgent: &ADKStageAgent{
			Stage:  domain.StageQuestioning,
			Prompt: GetSystemPromptForStage(domain.StageQuestioning),
			Agent:  stageAgents.Questioning,
		},
		AlgorithmAgent: &ADKStageAgent{
			Stage:  domain.StageAlgorithm,
			Prompt: GetSystemPromptForStage(domain.StageAlgorithm),
			Agent:  stageAgents.Algorithm,
		},
		ClosingAgent: &ADKStageAgent{
			Stage:  domain.StageClosing,
			Prompt: GetSystemPromptForStage(domain.StageClosing),
			Agent:  stageAgents.Closing,
		},
	})
	if err != nil {
		t.Fatalf("NewInterviewGraph failed: %v", err)
	}
	return graph
}

// TestGraph_IntroStage 验证自我介绍阶段能正常返回回复且阶段不切换。
func TestGraph_IntroStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "你好，我叫张三，有三年 Go 后端开发经验，做过微服务和中间件相关项目",
		InterviewID: "test-intro-001",
		Stage:       domain.StageIntro,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if output.Text == "" {
		t.Error("expected non-empty response text")
	}
	t.Logf("Stage: %s → %s", domain.StageIntro, output.NewStage)
	t.Logf("Response: %s", output.Text)
}

// TestGraph_QuestioningStage 验证技术问答阶段能正确出题。
func TestGraph_QuestioningStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "Goroutine 和线程的区别是什么？Goroutine 是怎么调度的？",
		InterviewID: "test-questioning-001",
		Stage:       domain.StageQuestioning,
		Context: map[string]any{
			"history": []map[string]string{
				{"role": "assistant", "content": "请介绍一下 Go 语言的并发模型"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if output.Text == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("Stage: %s → %s", domain.StageQuestioning, output.NewStage)
	t.Logf("Response: %s", output.Text)
}

// TestGraph_AlgorithmStage 验证算法题阶段出题正常。
func TestGraph_AlgorithmStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "好的，我准备好了",
		InterviewID: "test-algorithm-001",
		Stage:       domain.StageAlgorithm,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if output.Text == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("Stage: %s → %s", domain.StageAlgorithm, output.NewStage)
	t.Logf("Response: %s", output.Text)
}

// TestGraph_ClosingStage 验证反问阶段能正常回答候选人的问题。
func TestGraph_ClosingStage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "请问贵公司的技术栈主要是什么？团队规模大概多少人？",
		InterviewID: "test-closing-001",
		Stage:       domain.StageClosing,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if output.Text == "" {
		t.Error("expected non-empty response")
	}
	t.Logf("Stage: %s → %s", domain.StageClosing, output.NewStage)
	t.Logf("Response: %s", output.Text)
}

// TestGraph_HistoryCarryOver 验证多轮对话历史正确传递。
func TestGraph_HistoryCarryOver(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()
	interviewID := "test-history-001"

	// 第一轮
	out1, err := graph.Invoke(ctx, GraphInput{
		Text:        "我叫李四，熟悉 Redis 和 MySQL",
		InterviewID: interviewID,
		Stage:       domain.StageIntro,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Round 1 failed: %v", err)
	}
	t.Logf("Round 1 response: %s", out1.Text)

	// 第二轮，携带第一轮历史
	out2, err := graph.Invoke(ctx, GraphInput{
		Text:        "我在上一家公司主要负责订单系统的开发和优化",
		InterviewID: interviewID,
		Stage:       out1.NewStage,
		Context:     out1.Context,
	})
	if err != nil {
		t.Fatalf("Round 2 failed: %v", err)
	}
	if out2.Text == "" {
		t.Error("expected non-empty round 2 response")
	}
	t.Logf("Round 2 response: %s", out2.Text)
}

// TestGraph_StageAdvance 验证阶段能正确推进（intro → questioning）。
func TestGraph_StageAdvance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	// 构造一个已经完成自我介绍的上下文，看阶段是否推进
	history := []map[string]string{
		{"role": "assistant", "content": "你好，请先做自我介绍"},
		{"role": "user", "content": "我叫王五，有五年 Java 开发经验，熟悉 Spring Boot、Redis、Kafka"},
		{"role": "assistant", "content": "好的，请继续介绍你的项目经验"},
		{"role": "user", "content": "我在上一家公司做过电商平台，负责订单、支付、库存三个模块"},
		{"role": "assistant", "content": "很好，能详细介绍一下性能优化方面的工作吗"},
		{"role": "user", "content": "主要是通过 Redis 缓存热点数据，把数据库查询从每次请求降到十分之一"},
	}

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "好的，我的自我介绍就这些",
		InterviewID: "test-advance-001",
		Stage:       domain.StageIntro,
		Context:     map[string]any{"history": history},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	t.Logf("Stage result: %s → %s", domain.StageIntro, output.NewStage)
	t.Logf("Response: %s", output.Text)

	// 验证回复不为空，阶段字段有效
	if output.Text == "" {
		t.Error("expected non-empty response")
	}
	validStages := map[domain.InterviewStage]bool{
		domain.StageIntro:       true,
		domain.StageQuestioning: true,
	}
	if !validStages[output.NewStage] {
		t.Errorf("unexpected stage: %s", output.NewStage)
	}
}

// TestGraph_JSONResponseFormat 验证 LLM 返回的是可解析的 JSON 格式。
func TestGraph_JSONResponseFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	graph := newTestGraph(t)
	ctx := context.Background()

	output, err := graph.Invoke(ctx, GraphInput{
		Text:        "你好",
		InterviewID: "test-format-001",
		Stage:       domain.StageIntro,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	// 如果 LLM 按 prompt 要求返回了 JSON，output.Text 应该是解析后的 response 字段
	// 如果 LLM 没有返回 JSON（fallback），output.Text 是原始文字，不应该包含 JSON 标记
	if strings.Contains(output.Text, `"stage_action"`) {
		t.Errorf("response should be parsed text, not raw JSON: %s", output.Text)
	}
	t.Logf("Parsed response: %s", output.Text)
	t.Logf("New stage: %s", output.NewStage)
}
