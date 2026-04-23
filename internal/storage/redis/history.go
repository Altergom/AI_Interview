package redis

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
)

// AppendHistory 追加一条对话消息到 history。
// key: interview:{interview_id}:history
// 使用 Redis List 结构（RPUSH）。
func (c *Client) AppendHistory(ctx context.Context, interviewID string, msg *schema.Message, ttl time.Duration) error {
	// TODO: 实现追加逻辑
	// 1. 序列化 msg 为 JSON
	// 2. 使用 keys.InterviewHistoryKey() 生成 key
	// 3. RPUSH key value
	// 4. EXPIRE key ttl（刷新过期时间）
	return nil
}

// GetHistory 读取完整对话 history。
// 返回按时间顺序排列的消息列表。
func (c *Client) GetHistory(ctx context.Context, interviewID string) ([]*schema.Message, error) {
	// TODO: 实现读取逻辑
	// 1. 使用 keys.InterviewHistoryKey() 生成 key
	// 2. LRANGE key 0 -1
	// 3. 反序列化每条 JSON 为 schema.Message
	return nil, nil
}

// TrimHistory 裁剪 history，保留最近 N 条消息。
// 用于防止 history 过长导致 LLM context 超限。
func (c *Client) TrimHistory(ctx context.Context, interviewID string, keepLast int) error {
	// TODO: 实现裁剪逻辑
	// 1. 使用 keys.InterviewHistoryKey() 生成 key
	// 2. LTRIM key -keepLast -1
	return nil
}

// DeleteHistory 删除对话 history（面试结束时调用）。
func (c *Client) DeleteHistory(ctx context.Context, interviewID string) error {
	// TODO: 实现删除
	return nil
}
