package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore"
	"ai_interview/internal/log"
	"ai_interview/internal/mq"
	"ai_interview/internal/storage/postgres"
)

type ReportWorker struct {
	turnRepo   *postgres.InterviewTurnRepo
	reportRepo *postgres.ReportRepo
	invoker    *einocore.StructuredOutputInvoker
	msgs       <-chan amqp.Delivery
}

func NewReportWorker(
	turnRepo *postgres.InterviewTurnRepo,
	reportRepo *postgres.ReportRepo,
	invoker *einocore.StructuredOutputInvoker,
	msgs <-chan amqp.Delivery,
) *ReportWorker {
	return &ReportWorker{
		turnRepo:   turnRepo,
		reportRepo: reportRepo,
		invoker:    invoker,
		msgs:       msgs,
	}
}

func (w *ReportWorker) Run(ctx context.Context) {
	log.Infof("[ReportWorker] started, waiting for tasks on queue=%s", mq.TopicInterviewFinished)
	for {
		select {
		case <-ctx.Done():
			log.Infof("[ReportWorker] context done, shutting down")
			return
		case d, ok := <-w.msgs:
			if !ok {
				log.Warnf("[ReportWorker] delivery channel closed, stopping")
				return
			}
			w.handle(ctx, d)
		}
	}
}

func (w *ReportWorker) handle(ctx context.Context, d amqp.Delivery) {
	var event mq.InterviewFinished
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Errorf("[ReportWorker] unmarshal event: %v, body=%s", err, d.Body)
		d.Ack(false)
		return
	}

	log.Infof("[ReportWorker] processing interview_id=%s", event.InterviewID)

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := w.process(ctx, event.InterviewID); err != nil {
			lastErr = err
			log.Warnf("[ReportWorker] attempt %d/3 failed interview_id=%s: %v", attempt, event.InterviewID, err)
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}
		d.Ack(false)
		log.Infof("[ReportWorker] done interview_id=%s", event.InterviewID)
		return
	}

	log.Errorf("[ReportWorker] all retries failed interview_id=%s: %v", event.InterviewID, lastErr)
	_ = w.reportRepo.SaveError(ctx, event.InterviewID, lastErr.Error())
	d.Ack(false)
}

func (w *ReportWorker) process(ctx context.Context, interviewID string) error {
	turns, err := w.turnRepo.ListByInterview(ctx, interviewID)
	if err != nil {
		return fmt.Errorf("list turns: %w", err)
	}
	if len(turns) == 0 {
		return fmt.Errorf("no turns found for interview_id=%s", interviewID)
	}

	userContent := buildTranscript(turns)

	var report domain.Report
	if err := w.invoker.Invoke(ctx, reportSystemPrompt, userContent, &report); err != nil {
		return fmt.Errorf("invoke LLM: %w", err)
	}

	report.InterviewID = interviewID
	if err := w.reportRepo.Save(ctx, &report); err != nil {
		return fmt.Errorf("save report: %w", err)
	}
	return nil
}

func buildTranscript(turns []domain.InterviewTurn) string {
	var sb strings.Builder
	for _, t := range turns {
		if t.Question != "" {
			sb.WriteString(fmt.Sprintf("[面试官] %s\n", t.Question))
		}
		if t.UserAnswer != "" {
			sb.WriteString(fmt.Sprintf("[候选人] %s\n", t.UserAnswer))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}