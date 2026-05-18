package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	"ai_interview/internal/domain"
)

// stubTurnRepo 收集所有 SaveTurn 调用，便于断言。
type stubTurnRepo struct {
	mu      sync.Mutex
	saved   []domain.InterviewTurn
	failNow bool
}

func (s *stubTurnRepo) SaveTurn(_ context.Context, t domain.InterviewTurn) error {
	if s.failNow {
		return errors.New("simulated db failure")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saved = append(s.saved, t)
	return nil
}

func TestLastAssistantMessage(t *testing.T) {
	cases := []struct {
		name string
		in   []domain.SessionMessage
		want string
	}{
		{"empty", nil, ""},
		{"only user msgs", []domain.SessionMessage{
			{Role: "user", Content: "hi"},
		}, ""},
		{"single assistant", []domain.SessionMessage{
			{Role: "assistant", Content: "Q1"},
		}, "Q1"},
		{"picks last assistant", []domain.SessionMessage{
			{Role: "assistant", Content: "Q1"},
			{Role: "user", Content: "A1"},
			{Role: "assistant", Content: "Q2"},
			{Role: "user", Content: "A2"},
		}, "Q2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lastAssistantMessage(tc.in); got != tc.want {
				t.Errorf("want %q got %q", tc.want, got)
			}
		})
	}
}

func TestEffectiveStage(t *testing.T) {
	cases := []struct {
		name     string
		newStage domain.InterviewStage
		fallback domain.InterviewStage
		want     domain.InterviewStage
	}{
		{"prefers new stage", domain.StageQuestioning, domain.StageIntro, domain.StageQuestioning},
		{"falls back when empty", "", domain.StageIntro, domain.StageIntro},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := effectiveStage(tc.newStage, tc.fallback); got != tc.want {
				t.Errorf("want %q got %q", tc.want, got)
			}
		})
	}
}

func TestRecordTurn_WritesExpectedFields(t *testing.T) {
	stub := &stubTurnRepo{}
	svc := &interviewServiceImpl{turnRepo: stub}

	svc.recordTurn(context.Background(), "iv-1", 3,
		domain.StageQuestioning, "Q3", "我的回答", "我的 回答")

	if len(stub.saved) != 1 {
		t.Fatalf("expected 1 SaveTurn call, got %d", len(stub.saved))
	}
	got := stub.saved[0]
	if got.InterviewID != "iv-1" {
		t.Errorf("InterviewID: got %q", got.InterviewID)
	}
	if got.TurnID != "T03" {
		t.Errorf("TurnID: want T03 got %q", got.TurnID)
	}
	if got.Stage != string(domain.StageQuestioning) {
		t.Errorf("Stage: got %q", got.Stage)
	}
	if got.Question != "Q3" || got.UserAnswer != "我的回答" || got.ASRRaw != "我的 回答" {
		t.Errorf("content not propagated: %+v", got)
	}
}

func TestRecordTurn_NilRepoIsNoop(t *testing.T) {
	// 灰度场景：turnRepo 未注入时不应 panic
	svc := &interviewServiceImpl{turnRepo: nil}
	svc.recordTurn(context.Background(), "iv-1", 1, domain.StageIntro, "", "", "")
}

func TestRecordTurn_RepoFailureDoesNotPanic(t *testing.T) {
	// SFT 数据写库失败不应影响面试主流程，service 层只记 warn 不返回错误
	stub := &stubTurnRepo{failNow: true}
	svc := &interviewServiceImpl{turnRepo: stub}
	svc.recordTurn(context.Background(), "iv-1", 1, domain.StageIntro, "Q", "A", "A")
}
