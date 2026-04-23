package health

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	"ai_interview/internal/config"
)

type checkResult struct {
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func runReadiness(ctx context.Context, cfg *config.App) (map[string]checkResult, bool) {
	out := make(map[string]checkResult)
	allOK := true

	if cfg.PostgresDSN != "" {
		cr := checkPostgres(ctx, cfg.PostgresDSN)
		out["postgres"] = cr
		if !cr.OK {
			allOK = false
		}
	} else {
		out["postgres"] = checkResult{OK: true, Detail: "skipped (POSTGRES_DSN empty)"}
	}

	cr := checkRedis(ctx, cfg)
	out["redis"] = cr
	if !cr.OK {
		allOK = false
	}

	return out, allOK
}

func checkPostgres(ctx context.Context, dsn string) checkResult {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return checkResult{OK: false, Error: "open", Detail: err.Error()}
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return checkResult{OK: false, Error: "ping", Detail: err.Error()}
	}
	return checkResult{OK: true, Detail: "pong"}
}

func checkRedis(ctx context.Context, cfg *config.App) checkResult {
	if cfg.RedisAddr == "" {
		return checkResult{OK: true, Detail: "skipped (REDIS_ADDR empty)"}
	}

	opts := &redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return checkResult{OK: false, Error: "ping", Detail: err.Error()}
	}
	return checkResult{OK: true, Detail: "PONG"}
}
