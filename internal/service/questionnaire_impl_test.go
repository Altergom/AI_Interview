package service_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"ai_interview/internal/domain"
	biz "ai_interview/internal/errors"
	"ai_interview/internal/service"
	"ai_interview/internal/storage/postgres"
)

// fakeQuestionnaireRepo 内存实现 QuestionnaireRepository，单测专用。
type fakeQuestionnaireRepo struct {
	mu sync.Mutex

	// owners interview_id -> user_id
	owners map[string]string
	// turns interview_id -> turn_ids set
	turns map[string]map[string]struct{}
	// turnRows interview_id -> []InterviewTurn
	turnRows map[string][]domain.InterviewTurn

	// records primary key: interview_id|turn_id -> result（模拟 UPSERT）
	records map[string]domain.QuestionnaireResult

	// failOn 控制方法返回 error，用于异常路径测试
	failOn string
}

func newFakeQuestionnaireRepo() *fakeQuestionnaireRepo {
	return &fakeQuestionnaireRepo{
		owners:   map[string]string{},
		turns:    map[string]map[string]struct{}{},
		turnRows: map[string][]domain.InterviewTurn{},
		records:  map[string]domain.QuestionnaireResult{},
	}
}

func (f *fakeQuestionnaireRepo) seedInterview(interviewID, userID string, turnIDs ...string) {
	f.owners[interviewID] = userID
	set := make(map[string]struct{}, len(turnIDs))
	rows := make([]domain.InterviewTurn, 0, len(turnIDs))
	for _, t := range turnIDs {
		set[t] = struct{}{}
		rows = append(rows, domain.InterviewTurn{InterviewID: interviewID, TurnID: t, Stage: "questioning"})
	}
	f.turns[interviewID] = set
	f.turnRows[interviewID] = rows
}

func (f *fakeQuestionnaireRepo) UpsertBatch(_ context.Context, items []domain.QuestionnaireResult) error {
	if f.failOn == "Upsert" {
		return errors.New("simulated db failure")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, it := range items {
		key := it.InterviewID + "|" + it.TurnID
		f.records[key] = it
	}
	return nil
}

func (f *fakeQuestionnaireRepo) ListByInterview(_ context.Context, interviewID string) ([]domain.QuestionnaireResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []domain.QuestionnaireResult
	for k, v := range f.records {
		if strings.HasPrefix(k, interviewID+"|") {
			out = append(out, v)
		}
	}
	return out, nil
}

func (f *fakeQuestionnaireRepo) InterviewOwner(_ context.Context, interviewID string) (string, error) {
	if f.failOn == "Owner" {
		return "", errors.New("simulated db failure")
	}
	u, ok := f.owners[interviewID]
	if !ok {
		return "", postgres.ErrInterviewNotFound
	}
	return u, nil
}

func (f *fakeQuestionnaireRepo) ExistingTurnIDs(_ context.Context, interviewID string, ids []string) ([]string, error) {
	set := f.turns[interviewID]
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := set[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func (f *fakeQuestionnaireRepo) ListTurns(_ context.Context, interviewID string) ([]domain.InterviewTurn, error) {
	return f.turnRows[interviewID], nil
}

// ---------- 测试 ----------

const (
	tInterview = "interview-001"
	tOwner     = "user-001"
	tOther     = "user-evil"
)

func newSvcWithSeed(t *testing.T) (service.QuestionnaireService, *fakeQuestionnaireRepo) {
	t.Helper()
	repo := newFakeQuestionnaireRepo()
	repo.seedInterview(tInterview, tOwner, "T01", "T02", "T03")
	return service.NewQuestionnaireService(repo), repo
}

func TestSubmit_HappyPath(t *testing.T) {
	svc, repo := newSvcWithSeed(t)

	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID:      tOwner,
		InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{
			{TurnID: "T01", Quality: domain.QualityGood, Feedback: "节奏好"},
			{TurnID: "T02", Quality: domain.QualityBad, Feedback: "追问跑题"},
		},
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if got := len(repo.records); got != 2 {
		t.Fatalf("expected 2 records, got %d", got)
	}
	if repo.records[tInterview+"|T01"].UserID != tOwner {
		t.Errorf("UserID not propagated to record")
	}
}

func TestSubmit_UpsertOverwrites(t *testing.T) {
	svc, repo := newSvcWithSeed(t)
	ctx := context.Background()

	_ = svc.Submit(ctx, service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: domain.QualityGood}},
	})
	_ = svc.Submit(ctx, service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: domain.QualityBad, Feedback: "改主意"}},
	})

	if len(repo.records) != 1 {
		t.Fatalf("expected 1 record after upsert, got %d", len(repo.records))
	}
	if repo.records[tInterview+"|T01"].Quality != domain.QualityBad {
		t.Errorf("quality not overwritten on second submit")
	}
}

func TestSubmit_ForbiddenWhenNotOwner(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOther, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: domain.QualityGood}},
	})
	assertBizCode(t, err, biz.CodeInterviewForbidden)
}

func TestSubmit_InterviewNotFound(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: "nonexistent",
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: domain.QualityGood}},
	})
	assertBizCode(t, err, biz.CodeInterviewSessionNotFound)
}

func TestSubmit_TurnNotFound(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T99", Quality: domain.QualityGood}},
	})
	assertBizCode(t, err, biz.CodeInterviewTurnNotFound)
}

func TestSubmit_InvalidQuality(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: "neutral"}},
	})
	assertBizCode(t, err, biz.CodeBadRequest)
}

func TestSubmit_DuplicateTurnID(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{
			{TurnID: "T01", Quality: domain.QualityGood},
			{TurnID: "T01", Quality: domain.QualityBad},
		},
	})
	assertBizCode(t, err, biz.CodeBadRequest)
}

func TestSubmit_EmptyAnswers(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview, Answers: nil,
	})
	assertBizCode(t, err, biz.CodeBadRequest)
}

func TestSubmit_TooManyAnswers(t *testing.T) {
	svc, repo := newSvcWithSeed(t)
	// 准备 51 个真实 turn 以排除 turn-not-found 影响
	turnIDs := make([]string, 51)
	for i := range turnIDs {
		turnIDs[i] = "T" + itoa(i)
	}
	repo.seedInterview("big", tOwner, turnIDs...)

	answers := make([]service.QuestionnaireAnswer, 51)
	for i := range answers {
		answers[i] = service.QuestionnaireAnswer{TurnID: turnIDs[i], Quality: domain.QualityGood}
	}
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: "big", Answers: answers,
	})
	assertBizCode(t, err, biz.CodeBadRequest)
}

func TestSubmit_FeedbackTooLong(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	longFeedback := strings.Repeat("a", 1001)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: tOwner, InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{
			{TurnID: "T01", Quality: domain.QualityGood, Feedback: longFeedback},
		},
	})
	assertBizCode(t, err, biz.CodeBadRequest)
}

func TestSubmit_MissingUserID(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	err := svc.Submit(context.Background(), service.QuestionnaireSubmitRequest{
		UserID: "", InterviewID: tInterview,
		Answers: []service.QuestionnaireAnswer{{TurnID: "T01", Quality: domain.QualityGood}},
	})
	assertBizCode(t, err, biz.CodeUnauthorized)
}

func TestGet_HappyPath(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	turns, err := svc.Get(context.Background(), tOwner, tInterview)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(turns) != 3 {
		t.Errorf("expected 3 turns, got %d", len(turns))
	}
}

func TestGet_Forbidden(t *testing.T) {
	svc, _ := newSvcWithSeed(t)
	_, err := svc.Get(context.Background(), tOther, tInterview)
	assertBizCode(t, err, biz.CodeInterviewForbidden)
}

// ---------- helpers ----------

func assertBizCode(t *testing.T, err error, want biz.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %d, got nil", want)
	}
	be, ok := biz.IsBizError(err)
	if !ok {
		t.Fatalf("expected *BizError, got %T: %v", err, err)
	}
	if be.Code != want {
		t.Errorf("expected code %d, got %d (msg=%q)", want, be.Code, be.Message)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}
