package tools

import (
	"context"
	"testing"
)

func TestASRTool(t *testing.T) {
	ctx := context.Background()

	// 创建 Mock ASR 服务
	asrService := NewMockASRService()

	// 创建 ASR Tool
	asrTool, err := NewASRTool(asrService)
	if err != nil {
		t.Fatalf("NewASRTool failed: %v", err)
	}

	// 获取 Tool 信息
	info, err := asrTool.Info(ctx)
	if err != nil {
		t.Fatalf("asrTool.Info failed: %v", err)
	}

	t.Logf("ASR Tool Name: %s", info.Name)
	t.Logf("ASR Tool Description: %s", info.Desc)

	// 测试调用 ASR Tool
	result, err := asrTool.InvokableRun(ctx, `{"audio_data":"bW9jayBhdWRpbyBkYXRh"}`)
	if err != nil {
		t.Fatalf("asrTool.InvokableRun failed: %v", err)
	}

	t.Logf("ASR Result: %s", result)
}

func TestTTSTool(t *testing.T) {
	ctx := context.Background()

	// 创建 Mock TTS 服务
	ttsService := NewMockTTSService()

	// 创建 TTS Tool
	ttsTool, err := NewTTSTool(ttsService)
	if err != nil {
		t.Fatalf("NewTTSTool failed: %v", err)
	}

	// 获取 Tool 信息
	info, err := ttsTool.Info(ctx)
	if err != nil {
		t.Fatalf("ttsTool.Info failed: %v", err)
	}

	t.Logf("TTS Tool Name: %s", info.Name)
	t.Logf("TTS Tool Description: %s", info.Desc)

	// 测试调用 TTS Tool
	result, err := ttsTool.InvokableRun(ctx, `{"text":"你好，欢迎参加面试"}`)
	if err != nil {
		t.Fatalf("ttsTool.InvokableRun failed: %v", err)
	}

	t.Logf("TTS Result: %s", result)
}
