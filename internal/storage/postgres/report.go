package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"ai_interview/internal/domain"

	"gorm.io/gorm"
)

var ErrReportNotFound = errors.New("report not found")

type ReportRepository interface {
	Save(ctx context.Context, report *domain.Report) error
	GetByInterviewID(ctx context.Context, interviewID string) (*domain.Report, error)
	SaveError(ctx context.Context, interviewID, errMsg string) error
}

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

func (r *ReportRepo) Save(ctx context.Context, report *domain.Report) error {
	weakJSON, err := json.Marshal(report.WeakPoints)
	if err != nil {
		return fmt.Errorf("marshal weak_points: %w", err)
	}
	strongJSON, err := json.Marshal(report.StrongPoints)
	if err != nil {
		return fmt.Errorf("marshal strong_points: %w", err)
	}

	row := ReportModel{
		InterviewID:    report.InterviewID,
		KnowledgeDepth: report.KnowledgeDepth,
		Expression:     report.Expression,
		ProblemSolving: report.ProblemSolving,
		CodeQuality:    report.CodeQuality,
		StressResponse: report.StressResponse,
		Summary:        report.Summary,
		WeakPoints:     weakJSON,
		StrongPoints:   strongJSON,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("[ReportRepo] save: %w", err)
	}
	report.ID = row.ID
	report.CreatedAt = row.CreatedAt
	return nil
}

func (r *ReportRepo) GetByInterviewID(ctx context.Context, interviewID string) (*domain.Report, error) {
	var row ReportModel
	err := r.db.WithContext(ctx).
		Where(&ReportModel{InterviewID: interviewID}).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrReportNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("[ReportRepo] get by interview_id: %w", err)
	}
	return row.toDomain()
}

func (r *ReportRepo) SaveError(ctx context.Context, interviewID, errMsg string) error {
	row := ReportModel{
		InterviewID:  interviewID,
		ErrorMessage: errMsg,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("[ReportRepo] save error: %w", err)
	}
	return nil
}
