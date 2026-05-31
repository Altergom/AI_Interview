package evaluation

import (
	"fmt"
	"time"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"
)

const fallbackScore = 50

// Fallback 在 LLM 全部失败时返回保底报告。
// 所有维度给中等分数（50），ErrorMessage 记录失败原因。
func Fallback(interviewID string, cause error) *domain.Report {
	log.Warnf("[Evaluation] fallback triggered for interview %s: %v", interviewID, cause)
	return &domain.Report{
		InterviewID:    interviewID,
		KnowledgeDepth: fallbackScore,
		Expression:     fallbackScore,
		ProblemSolving: fallbackScore,
		CodeQuality:    fallbackScore,
		StressResponse: fallbackScore,
		Summary:        "评估服务暂时不可用，已生成保底报告，各维度均为中等分数。",
		StrongPoints:   []string{},
		WeakPoints:     []string{},
		ErrorMessage:   fmt.Sprintf("evaluation failed: %v", cause),
		CreatedAt:      time.Now(),
	}
}
