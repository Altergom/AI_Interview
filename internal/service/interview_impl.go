package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/compose"
	"ai_interview/internal/log"
	"ai_interview/internal/mq/mqclient"
	"ai_interview/internal/mq"
	"ai_interview/internal/storage/postgres"
	redisstorage "ai_interview/internal/storage/redis"
	"ai_interview/internal/storage/s3"
)

// interviewServiceImpl InterviewService 的实现
type interviewServiceImpl struct {
	sessionManager *SessionManager
	graph          *compose.InterviewGraph
	redisCli       *redisstorage.Client
	turnRepo       postgres.InterviewTurnRepository
	s3Client       *s3.Client
	mqClient       *mqclient.Client
	mqTopic        string
	stateTTL       time.Duration
}

// NewInterviewService 创建 InterviewService 实例，所有依赖从外部注入。
// turnRepo / s3Client / mqClient 允许 nil（测试 / 灰度场景），nil 时跳过对应功能。
func NewInterviewService(
	sessionManager *SessionManager,
	graph *compose.InterviewGraph,
	redisCli *redisstorage.Client,
	turnRepo postgres.InterviewTurnRepository,
	s3Client *s3.Client,
	mqClient *mqclient.Client,
	mqTopic string,
	stateTTL time.Duration,
) InterviewService {
	return &interviewServiceImpl{
		sessionManager: sessionManager,
		graph:          graph,
		redisCli:       redisCli,
		turnRepo:       turnRepo,
		s3Client:       s3Client,
		mqClient:       mqClient,
		mqTopic:        mqTopic,
		stateTTL:       stateTTL,
	}
}

// SetConfig 保存面试岗位和方向配置到 Redis，返回 interview_id。
// 前端拿到 interview_id 后再调 Create 创建 session。
func (s *interviewServiceImpl) SetConfig(ctx context.Context, req InterviewConfigRequest) (string, error) {
	if req.Direction == "" {
		return "", fmt.Errorf("[interview] direction is required")
	}

	interviewID := uuid.New().String()

	cfg := redisstorage.InterviewConfig{
		Direction: req.Direction,
		Position:  req.Position,
	}
	if err := s.redisCli.SaveInterviewConfig(ctx, interviewID, cfg, s.stateTTL); err != nil {
		return "", fmt.Errorf("[interview] save config: %w", err)
	}

	return interviewID, nil
}

// Create 创建面试 session，interviewID 由 SetConfig 返回，此处复用。
func (s *interviewServiceImpl) Create(ctx context.Context, interviewID, userID string) (*InterviewCreateResult, error) {
	// 校验 config 存在
	cfg, err := s.redisCli.GetInterviewConfig(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("[interview] config not found for interview_id=%s, call SetConfig first", interviewID)
	}

	// 创建 session
	if err := s.sessionManager.CreateSession(ctx, interviewID, userID); err != nil {
		return nil, fmt.Errorf("[interview] create session: %w", err)
	}

	// 初始化 InterviewState（含 Direction）
	state := &domain.InterviewState{
		InterviewID:  interviewID,
		Stage:        domain.StageIntro,
		Direction:    cfg.Direction,
		Position:     cfg.Position,
		ReportStatus: "pending",
		StartedAt:    time.Now(),
	}
	if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
		return nil, fmt.Errorf("[interview] save state: %w", err)
	}

	return &InterviewCreateResult{
		InterviewID: interviewID,
		Stage:       domain.StageIntro,
		CreatedAt:   state.StartedAt.Format(time.RFC3339),
	}, nil
}

// ProcessAudio 处理音频输入
func (s *interviewServiceImpl) ProcessAudio(ctx context.Context, req AudioRequest) error {
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get session: %w", err)
	}

	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get graph context: %w", err)
	}

	// 注入面试方向到 graph context，供 question_selector 使用
	state, err := s.redisCli.GetInterviewState(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get state: %w", err)
	}
	if state != nil {
		graphContext["direction"] = state.Direction
		graphContext["interview_id"] = req.InterviewID
	}

	// 异步上传音频到 S3，失败不阻塞面试流程
	if s.s3Client != nil && len(req.AudioData) > 0 {
		audioData := make([]byte, len(req.AudioData))
		copy(audioData, req.AudioData)
		go func() {
			if err := s.s3Client.UploadAudio(context.Background(), req.InterviewID, req.TurnID, bytes.NewReader(audioData)); err != nil {
				log.Warnf("[InterviewService] upload audio failed interview_id=%s turn_id=%s: %v", req.InterviewID, req.TurnID, err)
			}
		}()
	}

	// 在调用 graph 前抓取「上一句 AI 提问」——也就是用户本轮在回答的问题。
	// 必须在 UpdateFromGraphOutput 之前读取，否则 history 会被本轮新消息污染。
	priorQuestion := lastAssistantMessage(session.History)
	priorTotalRounds := session.Stats.TotalRounds

	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		AudioData:   req.AudioData,
		Text:        "",
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("[interview] invoke graph: %w", err)
	}

	userInput := "[音频输入]"
	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		userInput,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("[interview] update session: %w", err)
	}

	// GraphOutput.NewStage 是状态机裁决后的结果，需要同步到 Redis InterviewState。
	if state != nil && output.NewStage != "" && output.NewStage != state.Stage {
		state.Stage = output.NewStage
		if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
			log.Warnf("[InterviewService] save state after stage change interview_id=%s: %v", req.InterviewID, err)
		}
	}

	// 落 turn 到 PG（SFT 数据采集起点，失败不阻塞面试流程）
	s.recordTurn(ctx, req.InterviewID, priorTotalRounds+1,
		effectiveStage(output.NewStage, session.Stage),
		priorQuestion, userInput, userInput)

	return nil
}

// Finish 结束面试
func (s *interviewServiceImpl) Finish(ctx context.Context, interviewID string) (*InterviewFinishResult, error) {
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get session: %w", err)
	}

	duration := time.Since(session.CreatedAt)

	// 手动结束和 workflow 的 closing + finish 保持一致，都落到显式终态 StageEnd。
	if err := s.sessionManager.UpdateStage(ctx, interviewID, domain.StageEnd); err != nil {
		return nil, fmt.Errorf("[interview] update stage: %w", err)
	}

	// Redis state 也必须进入 StageEnd，否则前端和后续报告流程会看到不一致状态。
	state, err := s.redisCli.GetInterviewState(ctx, interviewID)
	if err == nil && state != nil {
		state.Stage = domain.StageEnd
		state.ReportStatus = "pending"
		if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
			log.Warnf("[InterviewService] save state on finish interview_id=%s: %v", interviewID, err)
		}
	}

	// 发布 interview_finished 事件到 MQ
	if s.mqClient != nil {
		event := mq.NewInterviewFinishedEvent(interviewID, session.UserID, time.Now())
		if err := s.mqClient.Publish(ctx, s.mqTopic, event); err != nil {
			log.Errorf("[InterviewService] publish interview_finished interview_id=%s: %v", interviewID, err)
		}
	}

	return &InterviewFinishResult{
		InterviewID:     interviewID,
		FinishedAt:      time.Now().Format(time.RFC3339),
		DurationSeconds: int64(duration.Seconds()),
	}, nil
}

// GetState 查询面试当前状态
func (s *interviewServiceImpl) GetState(ctx context.Context, interviewID string) (*domain.InterviewState, error) {
	state, err := s.redisCli.GetInterviewState(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get state: %w", err)
	}
	if state != nil {
		return state, nil
	}

	// 降级：从 session 构建
	session, err := s.sessionManager.GetSession(ctx, interviewID)
	if err != nil {
		return nil, fmt.Errorf("[interview] get session: %w", err)
	}

	return &domain.InterviewState{
		InterviewID:    session.InterviewID,
		Stage:          session.Stage,
		QuestionsAsked: session.Stats.QuestionCount,
		StartedAt:      session.CreatedAt,
	}, nil
}

// SubmitCode 提交代码
func (s *interviewServiceImpl) SubmitCode(ctx context.Context, req CodeSubmitRequest) error {
	session, err := s.sessionManager.GetSession(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get session: %w", err)
	}

	if session.Stage != domain.StageAlgorithm {
		return fmt.Errorf("[interview] invalid stage for code submission: %s", session.Stage)
	}

	graphContext, err := s.sessionManager.GetGraphContext(ctx, req.InterviewID)
	if err != nil {
		return fmt.Errorf("[interview] get graph context: %w", err)
	}

	priorQuestion := lastAssistantMessage(session.History)
	priorTotalRounds := session.Stats.TotalRounds

	codeSubmitText := fmt.Sprintf("我提交了代码：\n```%s\n%s\n```", req.Language, req.Code)
	output, err := s.graph.Invoke(ctx, compose.GraphInput{
		Text:        codeSubmitText,
		InterviewID: req.InterviewID,
		Stage:       session.Stage,
		Context:     graphContext,
	})
	if err != nil {
		return fmt.Errorf("[interview] invoke graph: %w", err)
	}

	if err := s.sessionManager.UpdateFromGraphOutput(
		ctx,
		req.InterviewID,
		codeSubmitText,
		output.Text,
		output.NewStage,
		output.Context,
	); err != nil {
		return fmt.Errorf("[interview] update session: %w", err)
	}

	if err := s.sessionManager.IncrementAlgorithmCount(ctx, req.InterviewID); err != nil {
		return fmt.Errorf("[interview] increment algorithm count: %w", err)
	}

	// 代码提交也可能触发 algorithm -> closing，需要和音频流程一样同步 Redis state。
	if output.NewStage != "" && output.NewStage != session.Stage {
		state, err := s.redisCli.GetInterviewState(ctx, req.InterviewID)
		if err == nil && state != nil {
			state.Stage = output.NewStage
			if err := s.redisCli.SaveInterviewState(ctx, state, s.stateTTL); err != nil {
				log.Warnf("[InterviewService] save state after code stage change interview_id=%s: %v", req.InterviewID, err)
			}
		}
	}

	// 代码提交也算一个 turn，asr_raw 保持为空（非语音输入）
	s.recordTurn(ctx, req.InterviewID, priorTotalRounds+1,
		effectiveStage(output.NewStage, session.Stage),
		priorQuestion, codeSubmitText, "")

	return nil
}

// recordTurn 异步语义的落库点：失败只记 warn，不影响面试主流程。
// turnNumber 从 1 起，VARCHAR(10) 上限远超任意单场面试规模。
func (s *interviewServiceImpl) recordTurn(
	ctx context.Context,
	interviewID string,
	turnNumber int,
	stage domain.InterviewStage,
	question, userAnswer, asrRaw string,
) {
	if s.turnRepo == nil {
		return
	}
	turn := domain.InterviewTurn{
		InterviewID: interviewID,
		TurnID:      fmt.Sprintf("T%02d", turnNumber),
		Stage:       string(stage),
		Question:    question,
		UserAnswer:  userAnswer,
		ASRRaw:      asrRaw,
	}
	if err := s.turnRepo.SaveTurn(ctx, turn); err != nil {
		log.Warnf("[InterviewService] save turn failed interview_id=%s turn_id=%s: %v",
			interviewID, turn.TurnID, err)
	}
}

// lastAssistantMessage 返回 history 中最后一条 assistant 消息内容；找不到返回 ""。
func lastAssistantMessage(history []domain.SessionMessage) string {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "assistant" {
			return history[i].Content
		}
	}
	return ""
}

// effectiveStage 选取本轮 turn 实际所处阶段：优先 graph 新阶段，否则保留旧值。
func effectiveStage(newStage, fallback domain.InterviewStage) domain.InterviewStage {
	if newStage != "" {
		return newStage
	}
	return fallback
}
