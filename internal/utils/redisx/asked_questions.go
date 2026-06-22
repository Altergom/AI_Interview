package redisx

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// QuestionHash calculates a stable short hash used for question de-duplication.
// It uses SHA-256 and keeps the first 16 bytes as hex string, balancing collision risk and storage size.
func QuestionHash(question string) string {
	sum := sha256.Sum256([]byte(question))
	return fmt.Sprintf("%x", sum[:16])
}

// MarkQuestionAsked writes the question hash into a Redis Set and keeps it aligned with interview TTL.
func MarkQuestionAsked(ctx context.Context, rdb *redis.Client, key, question string, ttl time.Duration) error {
	hash := QuestionHash(question)
	pipe := rdb.Pipeline()
	pipe.SAdd(ctx, key, hash)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// IsQuestionAsked checks whether the question is already in the asked-questions Redis Set.
func IsQuestionAsked(ctx context.Context, rdb *redis.Client, key, question string) (bool, error) {
	hash := QuestionHash(question)
	return rdb.SIsMember(ctx, key, hash).Result()
}

