package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"ai_interview/internal/config"
)

// loadConfig 向上遍历目录树找到项目根目录的 .env 文件并加载。
// 集成测试需要真实的 QWEN_API_KEY。
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

// skipIfNoAPIKey 没有配置 QWEN_API_KEY 时跳过集成测试。
func skipIfNoAPIKey(t *testing.T) {
	t.Helper()
	if config.Cfg == nil || config.Cfg.QwenAPIKey == "" {
		t.Skip("QWEN_API_KEY not set, skipping integration test")
	}
}

// ========== Mock 单元测试 ==========

func TestASRTool_Mock(t *testing.T) {
	ctx := context.Background()

	asrService := NewMockASRService()
	asrTool, err := NewASRTool(asrService)
	if err != nil {
		t.Fatalf("NewASRTool failed: %v", err)
	}

	info, err := asrTool.Info(ctx)
	if err != nil {
		t.Fatalf("asrTool.Info failed: %v", err)
	}
	t.Logf("ASR Tool: %s - %s", info.Name, info.Desc)

	result, err := asrTool.InvokableRun(ctx, `{"audio_data":"bW9jayBhdWRpbyBkYXRh"}`)
	if err != nil {
		t.Fatalf("asrTool.InvokableRun failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty ASR result")
	}
	t.Logf("ASR Result: %s", result)
}

func TestTTSTool_Mock(t *testing.T) {
	ctx := context.Background()

	ttsService := NewMockTTSService()
	ttsTool, err := NewTTSTool(ttsService)
	if err != nil {
		t.Fatalf("NewTTSTool failed: %v", err)
	}

	info, err := ttsTool.Info(ctx)
	if err != nil {
		t.Fatalf("ttsTool.Info failed: %v", err)
	}
	t.Logf("TTS Tool: %s - %s", info.Name, info.Desc)

	result, err := ttsTool.InvokableRun(ctx, `{"text":"你好，欢迎参加面试"}`)
	if err != nil {
		t.Fatalf("ttsTool.InvokableRun failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty TTS result")
	}
	t.Logf("TTS Result length: %d bytes", len(result))
}

// ========== Qwen 集成测试 ==========

func TestQwenTTS_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	ctx := context.Background()
	svc := NewQwenTTSService()

	cases := []struct {
		name string
		text string
	}{
		{"short_greeting", "你好，欢迎参加面试"},
		{"technical_question", "请介绍一下你对 Go 语言并发模型的理解"},
		{"long_sentence", "我们的面试分为自我介绍、技术问答、算法题和反问四个阶段，请做好准备"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			audio, err := svc.ConvertToAudio(ctx, tc.text)
			if err != nil {
				t.Fatalf("ConvertToAudio failed: %v", err)
			}
			if len(audio) == 0 {
				t.Error("expected non-empty audio bytes")
			}
			t.Logf("TTS [%s]: %d bytes audio generated", tc.name, len(audio))
		})
	}
}

func TestQwenASR_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	ctx := context.Background()
	ttsSvc := NewQwenTTSService()
	asrSvc := NewQwenASRService()

	// 用 TTS 生成真实音频，再用 ASR 识别，形成闭环
	text := "我有三年 Go 后端开发经验，熟悉微服务架构"
	audio, err := ttsSvc.ConvertToAudio(ctx, text)
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}
	t.Logf("TTS generated %d bytes, sending to ASR", len(audio))

	result, err := asrSvc.ConvertToText(ctx, audio)
	if err != nil {
		t.Fatalf("ASR failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty ASR result")
	}
	t.Logf("ASR result: %s", result)
}

func TestQwenASRTTS_RoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	ctx := context.Background()
	ttsSvc := NewQwenTTSService()
	asrSvc := NewQwenASRService()

	original := "请问您对分布式系统有哪些了解"

	// TTS: 文字 → 音频
	audio, err := ttsSvc.ConvertToAudio(ctx, original)
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}
	t.Logf("TTS: %q → %d bytes", original, len(audio))

	// ASR: 音频 → 文字
	recognized, err := asrSvc.ConvertToText(ctx, audio)
	if err != nil {
		t.Fatalf("ASR failed: %v", err)
	}
	t.Logf("ASR: %d bytes → %q", len(audio), recognized)

	if recognized == "" {
		t.Error("round-trip result is empty")
	}
}

func TestQwenTTSTool_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	ctx := context.Background()
	ttsTool, err := NewTTSTool(NewQwenTTSService())
	if err != nil {
		t.Fatalf("NewTTSTool failed: %v", err)
	}

	result, err := ttsTool.InvokableRun(ctx, `{"text":"面试开始，请先做自我介绍"}`)
	if err != nil {
		t.Fatalf("InvokableRun failed: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
	t.Logf("TTS Tool result length: %d", len(result))
}

func TestQwenASRTool_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	loadConfig(t)
	skipIfNoAPIKey(t)

	ctx := context.Background()

	audio, err := NewQwenTTSService().ConvertToAudio(ctx, "你好，我是候选人")
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}

	asrTool, err := NewASRTool(NewQwenASRService())
	if err != nil {
		t.Fatalf("NewASRTool failed: %v", err)
	}

	input, _ := json.Marshal(map[string]string{
		"audio_data": base64.StdEncoding.EncodeToString(audio),
	})
	result, err := asrTool.InvokableRun(ctx, string(input))
	if err != nil {
		t.Fatalf("InvokableRun failed: %v", err)
	}
	t.Logf("ASR Tool result: %s", result)
}
