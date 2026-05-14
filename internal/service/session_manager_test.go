package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/domain"
	redistore "ai_interview/internal/storage/redis"
)

func setupTestRedis(t *testing.T) (*redistore.Client, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb, err := redistore.New(context.Background(), redistore.Options{Addr: mr.Addr()})
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to create redis client: %v", err)
	}

	cleanup := func() {
		rdb.Close()
		mr.Close()
	}
	return rdb, cleanup
}

// newTestRedisClient 供 interview_impl_test.go 使用的裸客户端（miniredis）
func newTestRedisClient(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return client, func() { client.Close(); mr.Close() }
}

func TestSessionManager_CreateAndGetSession(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	// 创建会话
	interviewID := "test-interview-123"
	userID := "user-456"

	err := sm.CreateSession(ctx, interviewID, userID)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// 获取会话
	session, err := sm.GetSession(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	// 验证
	if session.InterviewID != interviewID {
		t.Errorf("Expected InterviewID %s, got %s", interviewID, session.InterviewID)
	}
	if session.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, session.UserID)
	}
	if session.Stage != domain.StageIntro {
		t.Errorf("Expected Stage %s, got %s", domain.StageIntro, session.Stage)
	}
	if len(session.History) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(session.History))
	}
}

func TestSessionManager_AddMessage(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 添加消息
	err := sm.AddMessage(ctx, interviewID, "user", "你好")
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	err = sm.AddMessage(ctx, interviewID, "assistant", "你好，请介绍一下你自己")
	if err != nil {
		t.Fatalf("AddMessage failed: %v", err)
	}

	// 获取历史
	history, err := sm.GetHistory(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	if history[0].Role != "user" || history[0].Content != "你好" {
		t.Errorf("First message mismatch")
	}

	if history[1].Role != "assistant" {
		t.Errorf("Second message role mismatch")
	}
}

func TestSessionManager_UpdateStage(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 更新阶段
	err := sm.UpdateStage(ctx, interviewID, domain.StageQuestioning)
	if err != nil {
		t.Fatalf("UpdateStage failed: %v", err)
	}

	// 获取阶段
	stage, err := sm.GetStage(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetStage failed: %v", err)
	}

	if stage != domain.StageQuestioning {
		t.Errorf("Expected stage %s, got %s", domain.StageQuestioning, stage)
	}
}

func TestSessionManager_Stats(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 增加计数
	sm.IncrementQuestionCount(ctx, interviewID)
	sm.IncrementQuestionCount(ctx, interviewID)
	sm.IncrementAlgorithmCount(ctx, interviewID)

	// 获取统计
	stats, err := sm.GetStats(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.QuestionCount != 2 {
		t.Errorf("Expected QuestionCount 2, got %d", stats.QuestionCount)
	}

	if stats.AlgorithmCount != 1 {
		t.Errorf("Expected AlgorithmCount 1, got %d", stats.AlgorithmCount)
	}
}

func TestSessionManager_Context(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 更新上下文
	err := sm.UpdateContext(ctx, interviewID, "tech_stack", []string{"Go", "Redis", "PostgreSQL"})
	if err != nil {
		t.Fatalf("UpdateContext failed: %v", err)
	}

	// 获取上下文
	context, err := sm.GetContext(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if context["tech_stack"] == nil {
		t.Error("Expected tech_stack in context")
	}
}

func TestSessionManager_DeleteSession(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 删除会话
	err := sm.DeleteSession(ctx, interviewID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// 尝试获取（应该失败）
	_, err = sm.GetSession(ctx, interviewID)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestSessionManager_GetGraphContext(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 添加一些消息
	sm.AddMessage(ctx, interviewID, "user", "你好")
	sm.AddMessage(ctx, interviewID, "assistant", "你好，请介绍一下你自己")

	// 更新统计
	sm.IncrementQuestionCount(ctx, interviewID)

	// 获取 Graph 上下文
	graphCtx, err := sm.GetGraphContext(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetGraphContext failed: %v", err)
	}

	// 验证历史对话
	history, ok := graphCtx["history"].([]map[string]string)
	if !ok {
		t.Fatal("Expected history in graph context")
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 messages in history, got %d", len(history))
	}

	// 验证统计信息
	if graphCtx["question_count"] != 1 {
		t.Errorf("Expected question_count 1, got %v", graphCtx["question_count"])
	}
}

func TestSessionManager_UpdateFromGraphOutput(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	interviewID := "test-interview-123"
	sm.CreateSession(ctx, interviewID, "user-456")

	// 模拟 Graph 输出
	err := sm.UpdateFromGraphOutput(
		ctx,
		interviewID,
		"我有3年Go开发经验",
		"很好，能详细说说你的项目经验吗？",
		domain.StageIntro,
		map[string]any{
			"tech_stack": []string{"Go", "Redis"},
		},
	)
	if err != nil {
		t.Fatalf("UpdateFromGraphOutput failed: %v", err)
	}

	// 验证历史对话
	history, err := sm.GetHistory(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	// 验证上下文
	context, err := sm.GetContext(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetContext failed: %v", err)
	}

	if context["tech_stack"] == nil {
		t.Error("Expected tech_stack in context")
	}
}

func TestSessionManager_ShouldAdvanceStage(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	sm := NewSessionManager(client, time.Hour)
	ctx := context.Background()

	tests := []struct {
		name          string
		stage         domain.InterviewStage
		setupFunc     func(string)
		shouldAdvance bool
	}{
		{
			name:  "intro stage - not enough rounds",
			stage: domain.StageIntro,
			setupFunc: func(id string) {
				sm.AddMessage(ctx, id, "user", "test")
			},
			shouldAdvance: false,
		},
		{
			name:  "intro stage - enough rounds",
			stage: domain.StageIntro,
			setupFunc: func(id string) {
				for i := 0; i < 3; i++ {
					sm.AddMessage(ctx, id, "user", "test")
				}
			},
			shouldAdvance: true,
		},
		{
			name:  "questioning stage - enough questions",
			stage: domain.StageQuestioning,
			setupFunc: func(id string) {
				for i := 0; i < 5; i++ {
					sm.IncrementQuestionCount(ctx, id)
				}
			},
			shouldAdvance: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interviewID := "test-" + tt.name
			sm.CreateSession(ctx, interviewID, "user-456")
			sm.UpdateStage(ctx, interviewID, tt.stage)

			if tt.setupFunc != nil {
				tt.setupFunc(interviewID)
			}

			shouldAdvance, err := sm.ShouldAdvanceStage(ctx, interviewID)
			if err != nil {
				t.Fatalf("ShouldAdvanceStage failed: %v", err)
			}

			if shouldAdvance != tt.shouldAdvance {
				t.Errorf("Expected shouldAdvance %v, got %v", tt.shouldAdvance, shouldAdvance)
			}
		})
	}
}
