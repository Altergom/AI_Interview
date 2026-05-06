package service

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/domain"
)

// 注意：这些测试需要真实的配置和 LLM
// 暂时跳过需要 Graph 的测试

func setupTestSessionManager(t *testing.T) (*SessionManager, func()) {
	// 创建 miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// 创建 SessionManager
	sm := NewSessionManager(client, time.Hour)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return sm, cleanup
}

func TestInterviewService_BasicFlow(t *testing.T) {
	t.Skip("Skipping test that requires real config and LLM")

	// 这个测试需要：
	// 1. 加载配置（config.Load()）
	// 2. 真实的 Supervisor 和 Graph
	// 3. 真实的 LLM API Key

	// 实际使用时的流程：
	// 1. 加载配置
	// 2. 创建 SessionManager
	// 3. 创建 InterviewService
	// 4. 调用 Create 创建面试
	// 5. 调用 ProcessText 处理输入
	// 6. 调用 GetState 查询状态
	// 7. 调用 Finish 结束面试
}

// 测试基本的会话管理功能（不需要 Graph）
func TestInterviewService_SessionManagement(t *testing.T) {
	sm, cleanup := setupTestSessionManager(t)
	defer cleanup()

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

	if session.InterviewID != interviewID {
		t.Errorf("Expected InterviewID %s, got %s", interviewID, session.InterviewID)
	}

	if session.Stage != domain.StageIntro {
		t.Errorf("Expected Stage %s, got %s", domain.StageIntro, session.Stage)
	}

	// 模拟对话
	sm.AddMessage(ctx, interviewID, "user", "你好")
	sm.AddMessage(ctx, interviewID, "assistant", "你好，请介绍一下你自己")

	// 获取历史
	history, err := sm.GetHistory(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(history))
	}

	// 更新阶段
	sm.UpdateStage(ctx, interviewID, domain.StageQuestioning)

	// 验证阶段
	stage, err := sm.GetStage(ctx, interviewID)
	if err != nil {
		t.Fatalf("GetStage failed: %v", err)
	}

	if stage != domain.StageQuestioning {
		t.Errorf("Expected Stage %s, got %s", domain.StageQuestioning, stage)
	}
}
