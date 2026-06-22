package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"
	"ai_interview/internal/storage/postgres"
	biz "ai_interview/internal/utils/respx"
)

// 问卷标注业务约束常量（v1 已确认）：
//   - 单次最多 50 条 turn，防止滥用
//   - feedback 上限 1000 字符，留足 SFT 上下文
const (
	maxAnswersPerSubmit   = 50
	maxFeedbackCharacters = 1000
)

type questionnaireService struct {
	repo postgres.QuestionnaireRepository
}

// NewQuestionnaireService 创建 QuestionnaireService 实现。
func NewQuestionnaireService(repo postgres.QuestionnaireRepository) QuestionnaireService {
	return &questionnaireService{repo: repo}
}

// Get 拉取面试的所有 turn 给前端渲染问卷条目。
// 鉴权：仅面试归属者可访问，不匹配返回 CodeInterviewForbidden。
func (s *questionnaireService) Get(ctx context.Context, userID, interviewID string) ([]domain.InterviewTurn, error) {
	if strings.TrimSpace(interviewID) == "" {
		return nil, biz.NewMsg(biz.CodeBadRequest, "interview_id is required")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, biz.New(biz.CodeUnauthorized)
	}

	if err := s.assertOwner(ctx, userID, interviewID); err != nil {
		return nil, err
	}

	turns, err := s.repo.ListTurns(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("list turns: %w", err)
	}
	log.Infof("[QuestionnaireService] get interview=%s turns=%d", interviewID, len(turns))
	return turns, nil
}

// Submit 校验 + 落库；good/bad 都采集，DPO 负样本由 quality=bad 自然形成。
func (s *questionnaireService) Submit(ctx context.Context, req QuestionnaireSubmitRequest) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("submit questionnaire: %w", err)
		}
	}()

	if err = validateSubmitRequest(req); err != nil {
		return err
	}

	if err = s.assertOwner(ctx, req.UserID, req.InterviewID); err != nil {
		return err
	}

	if err = s.assertTurnsExist(ctx, req); err != nil {
		return err
	}

	items := buildResults(req)
	if err = s.repo.UpsertBatch(ctx, items); err != nil {
		return fmt.Errorf("upsert batch: %w", err)
	}

	log.Infof("[QuestionnaireService] submitted interview=%s user=%s turns=%d",
		req.InterviewID, req.UserID, len(items))
	return nil
}

// assertOwner 通过 interviews.user_id 校验归属，复用同一处 BizError 转换。
func (s *questionnaireService) assertOwner(ctx context.Context, userID, interviewID string) (err error) {
	var owner string

	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"assert owner failed: interview_id=%s expected_user=%s actual_owner=%s: %w",
				interviewID,
				userID,
				owner,
				err,
			)
		}
	}()

	owner, err = s.repo.InterviewOwner(ctx, interviewID)
	if err != nil {
		if errors.Is(err, postgres.ErrInterviewNotFound) {
			return biz.New(biz.CodeInterviewSessionNotFound)
		}
		return fmt.Errorf("load interview owner: %w", err)
	}

	if owner != userID {
		return biz.New(biz.CodeInterviewForbidden)
	}

	return nil
}

// assertTurnsExist 校验所有 turn_id 都真实存在，防止伪造 turn 注入 SFT 数据。
func (s *questionnaireService) assertTurnsExist(ctx context.Context, req QuestionnaireSubmitRequest) (err error) {
	var missingTurnID string

	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"assert turns exist failed: interview_id=%s requested_turns=%d missing_turn_id=%s: %w",
				req.InterviewID,
				len(req.Answers),
				missingTurnID,
				err,
			)
		}
	}()

	requested := make([]string, 0, len(req.Answers))
	for _, a := range req.Answers {
		requested = append(requested, a.TurnID)
	}

	existing, err := s.repo.ExistingTurnIDs(ctx, req.InterviewID, requested)
	if err != nil {
		return fmt.Errorf("check turn existence: %w", err)
	}

	have := make(map[string]struct{}, len(existing))
	for _, id := range existing {
		have[id] = struct{}{}
	}

	for _, id := range requested {
		if _, ok := have[id]; !ok {
			missingTurnID = id
			return biz.NewMsg(
				biz.CodeInterviewTurnNotFound,
				fmt.Sprintf("turn_id %q not found in interview", id),
			)
		}
	}

	return nil
}

// validateSubmitRequest 纯函数：所有参数校验集中此处便于单测。
func validateSubmitRequest(req QuestionnaireSubmitRequest) error {
	if strings.TrimSpace(req.UserID) == "" {
		return biz.New(biz.CodeUnauthorized)
	}
	if strings.TrimSpace(req.InterviewID) == "" {
		return biz.NewMsg(biz.CodeBadRequest, "interview_id is required")
	}
	if len(req.Answers) == 0 {
		return biz.NewMsg(biz.CodeBadRequest, "answers is required")
	}
	if len(req.Answers) > maxAnswersPerSubmit {
		return biz.NewMsg(biz.CodeBadRequest,
			fmt.Sprintf("answers exceeds max %d per submit", maxAnswersPerSubmit))
	}

	seen := make(map[string]struct{}, len(req.Answers))
	for i, a := range req.Answers {
		if strings.TrimSpace(a.TurnID) == "" {
			return biz.NewMsg(biz.CodeBadRequest,
				fmt.Sprintf("answers[%d].turn_id is required", i))
		}
		if _, dup := seen[a.TurnID]; dup {
			return biz.NewMsg(biz.CodeBadRequest,
				fmt.Sprintf("duplicate turn_id %q in answers", a.TurnID))
		}
		seen[a.TurnID] = struct{}{}

		if a.Quality != domain.QualityGood && a.Quality != domain.QualityBad {
			return biz.NewMsg(biz.CodeBadRequest,
				fmt.Sprintf("answers[%d].quality must be 'good' or 'bad'", i))
		}
		// 中英文混合按 rune 计长度
		if len([]rune(a.Feedback)) > maxFeedbackCharacters {
			return biz.NewMsg(biz.CodeBadRequest,
				fmt.Sprintf("answers[%d].feedback exceeds %d chars", i, maxFeedbackCharacters))
		}
	}
	return nil
}

func buildResults(req QuestionnaireSubmitRequest) []domain.QuestionnaireResult {
	out := make([]domain.QuestionnaireResult, 0, len(req.Answers))
	for _, a := range req.Answers {
		out = append(out, domain.QuestionnaireResult{
			InterviewID: req.InterviewID,
			TurnID:      a.TurnID,
			Quality:     a.Quality,
			Feedback:    a.Feedback,
			UserID:      req.UserID,
		})
	}
	return out
}
