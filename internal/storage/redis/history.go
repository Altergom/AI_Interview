package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
)

// AppendHistory 追加一条对话消息到 history。
// key: interview:{interview_id}:history，使用 Redis List 结构。
func (c *Client) AppendHistory(ctx context.Context, interviewID string, msg *schema.Message, ttl time.Duration) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	key := InterviewHistoryKey(interviewID)
	if err := c.rdb.RPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("rpush history: %w", err)
	}
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// GetHistory 读取完整对话 history，按时间顺序排列。
func (c *Client) GetHistory(ctx context.Context, interviewID string) ([]*schema.Message, error) {
	items, err := c.rdb.LRange(ctx, InterviewHistoryKey(interviewID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange history: %w", err)
	}

	msgs := make([]*schema.Message, 0, len(items))
	for _, item := range items {
		var msg schema.Message
		if err := json.Unmarshal([]byte(item), &msg); err != nil {
			return nil, fmt.Errorf("unmarshal message: %w", err)
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

// TrimHistory 裁剪 history，只保留最近 keepLast 条消息。
// 用于防止 history 过长导致 LLM context 超限。
func (c *Client) TrimHistory(ctx context.Context, interviewID string, keepLast int) error {
	return c.rdb.LTrim(ctx, InterviewHistoryKey(interviewID), int64(-keepLast), -1).Err()
}

// DeleteHistory 删除对话 history（面试结束时调用）。
func (c *Client) DeleteHistory(ctx context.Context, interviewID string) error {
	return c.rdb.Del(ctx, InterviewHistoryKey(interviewID)).Err()
}
