package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrInterviewConfigNotFound = errors.New("interview config not found")

type InterviewConfigRecord struct {
	ConfigID  string `json:"config_id"`
	UserID    string `json:"user_id"`
	Position  string `json:"position"`
	Direction string `json:"direction"`
	CreatedAt string `json:"created_at"`
}

// SaveInterviewConfig 同时写入完整配置和用户最新配置索引。
// 完整配置用于按 config_id 精确读取；用户索引用于后续按 user_id 创建面试。
func (c *Client) SaveInterviewConfig(ctx context.Context, cfg InterviewConfigRecord, ttl time.Duration) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	configKey := InterviewConfigKey(cfg.ConfigID)
	userConfigKey := UserInterviewConfigKey(cfg.UserID)

	pipe := c.rdb.TxPipeline()

	// 存完整配置
	pipe.Set(ctx, configKey, data, ttl)

	// 关联 user_id 和 config_id
	pipe.Set(ctx, userConfigKey, cfg.ConfigID, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

// GetLatestInterviewConfigID 返回用户最近一次保存的面试配置 ID。
// Create 只接收 user_id 时，可通过该索引找到用户当前配置。
func (c *Client) GetLatestInterviewConfigID(ctx context.Context, userID string) (string, error) {
	configID, err := c.rdb.Get(ctx, UserInterviewConfigKey(userID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrInterviewConfigNotFound
		}
		return "", err
	}
	return configID, nil
}

// GetInterviewConfig 按 config_id 读取完整面试配置记录。
func (c *Client) GetInterviewConfig(ctx context.Context, configID string) (*InterviewConfigRecord, error) {

	configKey := InterviewConfigKey(configID)

	val, err := c.rdb.Get(ctx, configKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, ErrInterviewConfigNotFound
		}
		return nil, err
	}

	var config InterviewConfigRecord
	if err := json.Unmarshal([]byte(val), &config); err != nil {
		return nil, err
	}

	return &config, nil
}
