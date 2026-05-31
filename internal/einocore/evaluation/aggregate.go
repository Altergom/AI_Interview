package evaluation

import "ai_interview/internal/domain"

// Aggregate 对多批次分数做简单平均，返回最终维度分数。
// v1 简单平均；v2 可按批次权重（如后期批次权重更高）增强。
func Aggregate(batches []*domain.ReportDimensions) *domain.ReportDimensions {
	if len(batches) == 0 {
		return &domain.ReportDimensions{}
	}
	var total domain.ReportDimensions
	for _, b := range batches {
		total.KnowledgeDepth += b.KnowledgeDepth
		total.Expression += b.Expression
		total.ProblemSolving += b.ProblemSolving
		total.CodeQuality += b.CodeQuality
		total.StressResponse += b.StressResponse
	}
	n := len(batches)
	return &domain.ReportDimensions{
		KnowledgeDepth: total.KnowledgeDepth / n,
		Expression:     total.Expression / n,
		ProblemSolving: total.ProblemSolving / n,
		CodeQuality:    total.CodeQuality / n,
		StressResponse: total.StressResponse / n,
	}
}
