package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"ai_interview/internal/domain"

	"gorm.io/datatypes"
)

type UserModel struct {
	ID           string    `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"column:email"`
	Username     string    `gorm:"column:username"`
	PasswordHash string    `gorm:"column:password_hash"`
	IsGuest      bool      `gorm:"column:is_guest"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (UserModel) TableName() string {
	return "users"
}

type ResumeModel struct {
	ID          string         `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	UserID      string         `gorm:"column:user_id"`
	ContentHash string         `gorm:"column:content_hash"`
	S3Key       string         `gorm:"column:s3_key"`
	ParsedData  datatypes.JSON `gorm:"column:parsed_data;type:jsonb"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (ResumeModel) TableName() string {
	return "resumes"
}

type BankQuestionModel struct {
	ID                  string         `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	Question            string         `gorm:"column:question"`
	StandardAnswer      string         `gorm:"column:standard_answer"`
	Tags                datatypes.JSON `gorm:"column:tags;type:jsonb;default:'[]'"`
	RelatedConcepts     datatypes.JSON `gorm:"column:related_concepts;type:jsonb;default:'[]'"`
	FollowupQuestionIDs datatypes.JSON `gorm:"column:followup_question_ids;type:jsonb;default:'[]'"`
	Difficulty          string         `gorm:"column:difficulty;default:medium"`
	VecStatus           string         `gorm:"column:vec_status;default:pending"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
}

func (BankQuestionModel) TableName() string {
	return "bank_questions"
}

type InterviewModel struct {
	ID        string     `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string     `gorm:"column:user_id"`
	StartedAt time.Time  `gorm:"column:started_at"`
	EndedAt   *time.Time `gorm:"column:ended_at"`
	Status    string     `gorm:"column:status"`
}

func (InterviewModel) TableName() string {
	return "interviews"
}

func newResumeModel(userID, hash, s3Key string, resume *domain.StructuredResume) (*ResumeModel, error) {
	data, err := json.Marshal(resume)
	if err != nil {
		return nil, fmt.Errorf("marshal resume: %w", err)
	}
	return &ResumeModel{
		UserID:      userID,
		ContentHash: hash,
		S3Key:       s3Key,
		ParsedData:  datatypes.JSON(data),
	}, nil
}

func (m ResumeModel) toDomain() (*domain.StructuredResume, error) {
	var resume domain.StructuredResume
	if err := json.Unmarshal(m.ParsedData, &resume); err != nil {
		return nil, fmt.Errorf("unmarshal resume: %w", err)
	}
	return &resume, nil
}

func newBankQuestionModel(q *domain.BankQuestionRecord) (*BankQuestionModel, error) {
	tagsJSON, err := json.Marshal(q.Tags)
	if err != nil {
		return nil, fmt.Errorf("marshal tags: %w", err)
	}
	conceptsJSON, err := json.Marshal(q.RelatedConcepts)
	if err != nil {
		return nil, fmt.Errorf("marshal related_concepts: %w", err)
	}
	followupJSON, err := json.Marshal(q.FollowupQuestionIDs)
	if err != nil {
		return nil, fmt.Errorf("marshal followup_question_ids: %w", err)
	}
	return &BankQuestionModel{
		Question:            q.Question,
		StandardAnswer:      q.StandardAnswer,
		Tags:                datatypes.JSON(tagsJSON),
		RelatedConcepts:     datatypes.JSON(conceptsJSON),
		FollowupQuestionIDs: datatypes.JSON(followupJSON),
		Difficulty:          string(q.Difficulty),
	}, nil
}

func (m BankQuestionModel) toDomain() (*domain.BankQuestionRecord, error) {
	rec := &domain.BankQuestionRecord{
		ID:             m.ID,
		Question:       m.Question,
		StandardAnswer: m.StandardAnswer,
		Difficulty:     domain.Difficulty(m.Difficulty),
		VecStatus:      domain.VecStatus(m.VecStatus),
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if err := json.Unmarshal(m.Tags, &rec.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal tags: %w", err)
	}
	if err := json.Unmarshal(m.RelatedConcepts, &rec.RelatedConcepts); err != nil {
		return nil, fmt.Errorf("unmarshal related_concepts: %w", err)
	}
	if err := json.Unmarshal(m.FollowupQuestionIDs, &rec.FollowupQuestionIDs); err != nil {
		return nil, fmt.Errorf("unmarshal followup_question_ids: %w", err)
	}
	return rec, nil
}

func bankQuestionModelsToDomain(rows []BankQuestionModel) ([]*domain.BankQuestionRecord, error) {
	records := make([]*domain.BankQuestionRecord, 0, len(rows))
	for _, row := range rows {
		rec, err := row.toDomain()
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}
