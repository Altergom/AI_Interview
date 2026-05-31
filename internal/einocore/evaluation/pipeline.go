// Package evaluation 定义面试评估管道的统一接口。
// v1 只实现语音输入分支（audio_transcript），文字分支留接口给 v2。
package evaluation

import (
	"context"

	"ai_interview/internal/domain"
)

// Turn 表示一轮面试对话，同时支持文字和语音转写两种输入来源。
// v1 只填 AudioTranscript；Text 留给 v2 文字面试分支。
type Turn struct {
	TurnID          string `json:"turn_id"`
	Question        string `json:"question"`
	Text            string `json:"text,omitempty"`            // v2 文字分支
	AudioTranscript string `json:"audio_transcript,omitempty"` // v1 语音转写
}

// EvaluationPipeline 评估管道统一接口。
// 调用方只需关心 Run，内部分批、聚合、降级对外透明。
type EvaluationPipeline interface {
	// Run 对完整对话记录执行评估，返回报告。
	// 内部按 BatchSize 分批调用 LLM，再聚合为最终分数。
	// LLM 全部失败时返回 Fallback 保底报告，不返回 error。
	Run(ctx context.Context, interviewID string, turns []Turn) (*domain.Report, error)
}
