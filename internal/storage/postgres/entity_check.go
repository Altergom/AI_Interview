package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

// EntityChecker Worker 处理消息前的实体预校验。
//
// 使用场景：Worker 从 MQ 消费到消息后，先调用对应的 Exists 方法确认实体仍然存在。
// 若已被删除则直接 ACK 丢弃，不执行后续业务逻辑，避免无效重试。
//
// Worker 标准用法：
//
//	ok, err := checker.InterviewExists(ctx, interviewID)
//	if err != nil { /* 查询失败，NACK 重试 */ }
//	if !ok { /* 实体已删除，ACK 丢弃 */ return nil }
//	// 正常处理...
type EntityChecker struct {
	db *sql.DB
}

// NewEntityChecker 创建 EntityChecker。
func NewEntityChecker(db *sql.DB) *EntityChecker {
	return &EntityChecker{db: db}
}

// InterviewExists 检查面试会话是否存在。
func (c *EntityChecker) InterviewExists(ctx context.Context, interviewID string) (bool, error) {
	return c.exists(ctx, "interviews", interviewID)
}

// BankQuestionExists 检查题库题目是否存在（向量化 Worker 使用）。
func (c *EntityChecker) BankQuestionExists(ctx context.Context, questionID string) (bool, error) {
	return c.exists(ctx, "bank_questions", questionID)
}

// exists 通用存在性查询，仅对已知表名使用，避免拼接用户输入导致 SQL 注入。
func (c *EntityChecker) exists(ctx context.Context, table, id string) (bool, error) {
	var ok bool
	// table 为硬编码字符串，不来自外部输入，无注入风险
	q := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", table)
	if err := c.db.QueryRowContext(ctx, q, id).Scan(&ok); err != nil {
		return false, fmt.Errorf("check %s exists id=%s: %w", table, id, err)
	}
	return ok, nil
}
