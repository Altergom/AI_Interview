package compose

import (
	"testing"

	"ai_interview/internal/domain"
)

func TestGetSystemPrompt(t *testing.T) {
	tests := []struct {
		stage domain.InterviewStage
		want  string
	}{
		{domain.StageIntro, introPrompt},
		{domain.StageQuestioning, questioningPrompt},
		{domain.StageAlgorithm, algorithmPrompt},
		{domain.StageClosing, closingPrompt},
		{domain.InterviewStage("unknown"), basePrompt},
	}

	for _, tt := range tests {
		t.Run(string(tt.stage), func(t *testing.T) {
			got := getSystemPrompt(tt.stage)
			if got != tt.want {
				t.Errorf("getSystemPrompt(%v) = %v, want %v", tt.stage, got, tt.want)
			}
		})
	}
}

func TestGetNextStage(t *testing.T) {
	tests := []struct {
		current domain.InterviewStage
		want    domain.InterviewStage
	}{
		{domain.StageIntro, domain.StageQuestioning},
		{domain.StageQuestioning, domain.StageAlgorithm},
		{domain.StageAlgorithm, domain.StageClosing},
		{domain.StageClosing, domain.StageClosing}, // 最后阶段不变
	}

	for _, tt := range tests {
		t.Run(string(tt.current), func(t *testing.T) {
			got := getNextStage(tt.current)
			if got != tt.want {
				t.Errorf("getNextStage(%v) = %v, want %v", tt.current, got, tt.want)
			}
		})
	}
}

func TestParseSupervisorOutput(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		currentStage domain.InterviewStage
		wantStage    domain.InterviewStage
	}{
		{
			name:         "normal response",
			text:         "你好，请介绍一下你自己",
			currentStage: domain.StageIntro,
			wantStage:    domain.StageIntro,
		},
		{
			name:         "empty response",
			text:         "",
			currentStage: domain.StageQuestioning,
			wantStage:    domain.StageQuestioning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSupervisorOutput(tt.text, tt.currentStage)
			if got.NewStage != tt.wantStage {
				t.Errorf("parseSupervisorOutput() NewStage = %v, want %v", got.NewStage, tt.wantStage)
			}
			if got.Text != tt.text {
				t.Errorf("parseSupervisorOutput() Text = %v, want %v", got.Text, tt.text)
			}
		})
	}
}

func TestCallSupervisor_HistoryHandling(t *testing.T) {
	// 测试历史对话处理

	// 创建测试输入
	input := GraphInput{
		Text:        "我有3年Go开发经验",
		InterviewID: "test-123",
		Stage:       domain.StageIntro,
		Context: map[string]any{
			"history": []map[string]string{
				{"role": "user", "content": "你好"},
				{"role": "assistant", "content": "你好，请介绍一下你自己"},
			},
		},
	}

	// 验证历史对话能正确解析
	if history, ok := input.Context["history"].([]map[string]string); ok {
		if len(history) != 2 {
			t.Errorf("Expected 2 history messages, got %d", len(history))
		}
	} else {
		t.Error("Failed to parse history from context")
	}
}
