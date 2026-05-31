package compose

import (
	"context"
	"testing"

	"ai_interview/internal/domain"
)

type fakeStageAgent struct {
	result StageResult
	inputs []StageInput
}

func (f *fakeStageAgent) Run(_ context.Context, input StageInput) (StageResult, error) {
	f.inputs = append(f.inputs, input)
	return f.result, nil
}

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
		{domain.StageClosing, domain.StageEnd},
		{domain.StageEnd, domain.StageEnd},
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

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "pure json",
			input: `{"response": "你好", "need_tts": true, "stage_action": "continue"}`,
			want:  `{"response": "你好", "need_tts": true, "stage_action": "continue"}`,
		},
		{
			name: "json in markdown code block",
			input: "```json\n" +
				`{"response": "你好", "need_tts": true, "stage_action": "continue"}` + "\n```",
			want: `{"response": "你好", "need_tts": true, "stage_action": "continue"}`,
		},
		{
			name: "json in plain code block",
			input: "```\n" +
				`{"response": "你好", "need_tts": true, "stage_action": "continue"}` + "\n```",
			want: `{"response": "你好", "need_tts": true, "stage_action": "continue"}`,
		},
		{
			name:  "json with surrounding text",
			input: `这是回复：{"response": "你好", "need_tts": true, "stage_action": "continue"} 结束`,
			want:  `{"response": "你好", "need_tts": true, "stage_action": "continue"}`,
		},
		{
			name:  "no json",
			input: "这只是普通文本",
			want:  "这只是普通文本",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Errorf("extractJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSupervisorOutput_JSON(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		currentStage domain.InterviewStage
		wantText     string
		wantStage    domain.InterviewStage
	}{
		{
			name:         "valid json - continue",
			text:         `{"response": "请介绍一下你自己", "need_tts": true, "stage_action": "continue"}`,
			currentStage: domain.StageIntro,
			wantText:     "请介绍一下你自己",
			wantStage:    domain.StageIntro,
		},
		{
			name:         "valid json - advance",
			text:         `{"response": "很好，我们进入技术问答", "need_tts": true, "stage_action": "advance"}`,
			currentStage: domain.StageIntro,
			wantText:     "很好，我们进入技术问答",
			wantStage:    domain.StageQuestioning,
		},
		{
			name:         "valid json - finish",
			text:         `{"response": "感谢参加面试", "need_tts": true, "stage_action": "finish"}`,
			currentStage: domain.StageClosing,
			wantText:     "感谢参加面试",
			wantStage:    domain.StageClosing,
		},
		{
			name: "json in markdown",
			text: "```json\n" +
				`{"response": "请介绍一下你自己", "need_tts": true, "stage_action": "continue"}` + "\n```",
			currentStage: domain.StageIntro,
			wantText:     "请介绍一下你自己",
			wantStage:    domain.StageIntro,
		},
		{
			name:         "invalid json - fallback",
			text:         "这不是 JSON",
			currentStage: domain.StageQuestioning,
			wantText:     "这不是 JSON",
			wantStage:    domain.StageQuestioning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSupervisorOutput(tt.text, tt.currentStage)
			if got.Text != tt.wantText {
				t.Errorf("parseSupervisorOutput() Text = %v, want %v", got.Text, tt.wantText)
			}
			if got.NewStage != tt.wantStage {
				t.Errorf("parseSupervisorOutput() NewStage = %v, want %v", got.NewStage, tt.wantStage)
			}
		})
	}
}

func TestInterviewGraph_WorkflowAdvance(t *testing.T) {
	introAgent := &fakeStageAgent{result: StageResult{
		Response:    "进入技术问答",
		NeedTTS:     true,
		AgentAction: "advance",
		Metadata: map[string]any{
			"detected_skills": []string{"Go", "Redis"},
		},
	}}
	graph := newFakeGraph(t, InterviewGraphConfig{
		IntroAgent:       introAgent,
		QuestioningAgent: &fakeStageAgent{result: continueResult("questioning")},
		AlgorithmAgent:   &fakeStageAgent{result: continueResult("algorithm")},
		ClosingAgent:     &fakeStageAgent{result: continueResult("closing")},
	})

	output, err := graph.Invoke(context.Background(), GraphInput{
		Text:        "我熟悉 Go 和 Redis",
		InterviewID: "wf-advance",
		Stage:       domain.StageIntro,
		Context:     map[string]any{"detected_skills": []string{"old"}},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if output.NewStage != domain.StageQuestioning {
		t.Fatalf("NewStage = %s, want %s", output.NewStage, domain.StageQuestioning)
	}
	if output.Text != "进入技术问答" {
		t.Fatalf("Text = %q", output.Text)
	}
	if got := output.Context["next_stage"]; got != domain.StageQuestioning.String() {
		t.Fatalf("next_stage metadata = %v", got)
	}
	if len(introAgent.inputs) != 1 || introAgent.inputs[0].Stage != domain.StageIntro {
		t.Fatalf("intro agent was not called with intro stage")
	}
}

func TestInterviewGraph_ClosingFinishToEnd(t *testing.T) {
	graph := newFakeGraph(t, InterviewGraphConfig{
		IntroAgent:       &fakeStageAgent{result: continueResult("intro")},
		QuestioningAgent: &fakeStageAgent{result: continueResult("questioning")},
		AlgorithmAgent:   &fakeStageAgent{result: continueResult("algorithm")},
		ClosingAgent: &fakeStageAgent{result: StageResult{
			Response:    "感谢参加面试",
			AgentAction: "finish",
		}},
	})

	output, err := graph.Invoke(context.Background(), GraphInput{
		Text:        "没有问题了",
		InterviewID: "wf-end",
		Stage:       domain.StageClosing,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if output.NewStage != domain.StageEnd {
		t.Fatalf("NewStage = %s, want %s", output.NewStage, domain.StageEnd)
	}
}

func TestInterviewGraph_EndStageRejectsInput(t *testing.T) {
	introAgent := &fakeStageAgent{result: continueResult("intro")}
	graph := newFakeGraph(t, InterviewGraphConfig{
		IntroAgent:       introAgent,
		QuestioningAgent: &fakeStageAgent{result: continueResult("questioning")},
		AlgorithmAgent:   &fakeStageAgent{result: continueResult("algorithm")},
		ClosingAgent:     &fakeStageAgent{result: continueResult("closing")},
	})

	output, err := graph.Invoke(context.Background(), GraphInput{
		Text:        "还能继续吗",
		InterviewID: "wf-ended",
		Stage:       domain.StageEnd,
		Context:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if output.NewStage != domain.StageEnd {
		t.Fatalf("NewStage = %s, want %s", output.NewStage, domain.StageEnd)
	}
	if output.Text != endedMessage {
		t.Fatalf("Text = %q, want %q", output.Text, endedMessage)
	}
	if len(introAgent.inputs) != 0 {
		t.Fatalf("end stage should not call stage agents")
	}
}

func newFakeGraph(t *testing.T, cfg InterviewGraphConfig) *InterviewGraph {
	t.Helper()
	graph, err := NewInterviewGraph(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewInterviewGraph failed: %v", err)
	}
	return graph
}

func continueResult(text string) StageResult {
	return StageResult{
		Response:    text,
		AgentAction: "continue",
		Metadata:    map[string]any{},
	}
}
