package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"ai_interview/internal/domain"
	"ai_interview/internal/storage/redis"

	"github.com/google/uuid"
)


type interviewService struct {
	rdb       *redis.Client
	configTTL time.Duration
}

func NewInterviewService(client *redis.Client, configTTL time.Duration) InterviewService {
	return &interviewService{
		rdb:       client,
		configTTL: configTTL,
	}
}

func (s *interviewService) SetConfig(ctx context.Context, req InterviewConfigRequest) (string, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Position = strings.TrimSpace(req.Position)
	req.Direction = strings.TrimSpace(req.Direction)

	if req.UserID == "" {
		return "", errors.New("user id is empty")
	}
	if req.Position == "" {
		return "", errors.New("position is empty")
	}
	if req.Direction == "" {
		return "", errors.New("direction is empty")
	}

	configID := uuid.NewString()

	now := time.Now().UTC().Format(time.RFC3339)

	config := redis.InterviewConfigRecord{
		ConfigID:  configID,
		UserID:    req.UserID,
		Position:  req.Position,
		Direction: req.Direction,
		CreatedAt: now,
	}

	if err := s.rdb.SaveInterviewConfig(ctx, config, s.configTTL); err != nil {
		return "", err
	}

	return configID, nil
}


func (s *interviewService) Create(ctx context.Context, userID string) (*InterviewCreateResult, error) {
	panic("not implemented")
}

func (s *interviewService) ProcessAudio(ctx context.Context, req AudioRequest) error {
	panic("not implemented")
}

func (s *interviewService) Finish(ctx context.Context, interviewID string) (*InterviewFinishResult, error) {
	panic("not implemented")
}

func (s *interviewService) GetState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	panic("not implemented")
}

func (s *interviewService) SubmitCode(ctx context.Context, req CodeSubmitRequest) error {
	panic("not implemented")
}
