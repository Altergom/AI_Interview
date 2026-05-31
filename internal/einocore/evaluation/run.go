package evaluation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"ai_interview/internal/domain"
	"ai_interview/internal/log"
)

const summarizeSystemPrompt = `你是一位专业的技术面试评估官。根据以下面试对话记录和各维度评分，生成面试总结。

请严格按照以下 JSON 格式输出，不要包含任何其他文字：
{"summary":"<100字以内的整体评价>","strong_points":["<优势1>","<优势2>"],"weak_points":["<不足1>","<不足2>"]}`

// Run 对完整对话记录执行评估，返回报告。
// 内部按 BatchSize 分批调用 LLM，再聚合为最终分数。
// LLM 全部失败时返回 Fallback 保底报告，不返回 error。
func (e *Evaluator) Run(ctx context.Context, interviewID string, turns []Turn) (*domain.Report, error) {
	if len(turns) == 0 {
		return Fallback(interviewID, fmt.Errorf("no turns to evaluate")), nil
	}

	batches := splitBatches(turns, BatchSize)
	var (
		results []*domain.ReportDimensions
		lastErr error
	)
	for i, batch := range batches {
		dims, err := e.BatchScore(ctx, batch)
		if err != nil {
			log.Warnf("[Evaluation] batch %d/%d failed for interview %s: %v", i+1, len(batches), interviewID, err)
			lastErr = err
			continue
		}
		results = append(results, dims)
	}

	if len(results) == 0 {
		return Fallback(interviewID, lastErr), nil
	}

	dims := Aggregate(results)
	summary, strongPoints, weakPoints := e.summarize(ctx, turns, dims)

	report := &domain.Report{
		ID:             uuid.New().String(),
		InterviewID:    interviewID,
		KnowledgeDepth: dims.KnowledgeDepth,
		Expression:     dims.Expression,
		ProblemSolving: dims.ProblemSolving,
		CodeQuality:    dims.CodeQuality,
		StressResponse: dims.StressResponse,
		Summary:        summary,
		StrongPoints:   strongPoints,
		WeakPoints:     weakPoints,
		CreatedAt:      time.Now(),
	}

	log.Infof("[Evaluation] report generated for interview %s", interviewID)
	return report, nil
}

// summarize 生成总结文字、优势和不足。LLM 失败时返回默认文案，不阻断主流程。
func (e *Evaluator) summarize(ctx context.Context, turns []Turn, dims *domain.ReportDimensions) (summary string, strongPoints, weakPoints []string) {
	type summaryOutput struct {
		Summary      string   `json:"summary"`
		StrongPoints []string `json:"strong_points"`
		WeakPoints   []string `json:"weak_points"`
	}

	var out summaryOutput
	if err := e.invoker.Invoke(ctx, summarizeSystemPrompt, buildSummarizeContent(turns, dims), &out); err != nil {
		log.Warnf("[Evaluation] summarize failed, using default: %v", err)
		return "面试已完成，详情请参考各维度评分。", []string{}, []string{}
	}
	return out.Summary, out.StrongPoints, out.WeakPoints
}

// buildSummarizeContent 构建总结 LLM 的输入文本，最多取最近 16 轮避免超长。
func buildSummarizeContent(turns []Turn, dims *domain.ReportDimensions) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "各维度评分：知识深度=%d 表达能力=%d 解决问题=%d 代码质量=%d 抗压能力=%d\n\n",
		dims.KnowledgeDepth, dims.Expression, dims.ProblemSolving, dims.CodeQuality, dims.StressResponse)
	sb.WriteString("面试对话摘要：\n\n")
	start := 0
	if len(turns) > 16 {
		start = len(turns) - 16
	}
	for i, t := range turns[start:] {
		answer := t.AudioTranscript
		if answer == "" {
			answer = t.Text
		}
		fmt.Fprintf(&sb, "第 %d 轮\n问题：%s\n回答：%s\n\n", i+1, t.Question, answer)
	}
	return sb.String()
}

// splitBatches 将 turns 按 size 切分为多个批次。
func splitBatches(turns []Turn, size int) [][]Turn {
	var batches [][]Turn
	for i := 0; i < len(turns); i += size {
		end := i + size
		if end > len(turns) {
			end = len(turns)
		}
		batches = append(batches, turns[i:end])
	}
	return batches
}
