package evaluation

import (
	"context"
	"fmt"
	"strings"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore"
	"ai_interview/internal/log"
)

const (
	// BatchSize 每批最多评估的对话轮数，控制单次 LLM 上下文大小。
	BatchSize = 8
)

const batchScoreSystemPrompt = `你是一位专业的技术面试评估官。请根据以下面试对话记录，对候选人进行评估。

评估维度（每项 0-100 分）：
- knowledge_depth：技术知识深度，考察对核心概念的理解程度
- expression：表达能力，考察语言组织和沟通清晰度
- problem_solving：解决问题能力，考察分析和拆解问题的思路
- code_quality：代码质量意识，考察对代码规范、性能、安全的关注
- stress_response：抗压能力，考察在追问和压力下的表现

请严格按照以下 JSON 格式输出，不要包含任何其他文字：
{"knowledge_depth":<0-100>,"expression":<0-100>,"problem_solving":<0-100>,"code_quality":<0-100>,"stress_response":<0-100>}`

// Evaluator 实现 EvaluationPipeline 接口。
type Evaluator struct {
	invoker *einocore.StructuredOutputInvoker
}

// NewEvaluator 创建 Evaluator，invoker 由调用方注入，便于测试替换。
func NewEvaluator(invoker *einocore.StructuredOutputInvoker) *Evaluator {
	return &Evaluator{invoker: invoker}
}

// BatchScore 对一批 turns 调用 LLM 评估，返回该批次的维度分数。
// turns 长度不超过 BatchSize，由 Run 负责切分后调用。
func (e *Evaluator) BatchScore(ctx context.Context, turns []Turn) (*domain.ReportDimensions, error) {
	content := buildBatchContent(turns)

	var dims domain.ReportDimensions
	if err := e.invoker.Invoke(ctx, batchScoreSystemPrompt, content, &dims); err != nil {
		return nil, fmt.Errorf("batch score: %w", err)
	}

	log.Infof("[Evaluation] batch scored %d turns: knowledge=%d expression=%d problem=%d code=%d stress=%d",
		len(turns), dims.KnowledgeDepth, dims.Expression, dims.ProblemSolving, dims.CodeQuality, dims.StressResponse)

	return &dims, nil
}

// buildBatchContent 将一批 turns 序列化为 LLM 输入文本。
func buildBatchContent(turns []Turn) string {
	var sb strings.Builder
	sb.WriteString("以下是本批次的面试对话记录：\n\n")
	for i, t := range turns {
		answer := t.AudioTranscript
		if answer == "" {
			answer = t.Text
		}
		fmt.Fprintf(&sb, "第 %d 轮\n问题：%s\n回答：%s\n\n", i+1, t.Question, answer)
	}
	return sb.String()
}
