package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ai_interview/internal/log"
)

// Options PostgreSQL 连接池配置，全部字段均有合理默认值。
type Options struct {
	DSN             string
	MaxOpenConns    int           // 默认 25
	MaxIdleConns    int           // 默认 5
	ConnMaxLifetime time.Duration // 默认 30m
	ConnMaxIdleTime time.Duration // 默认 5m
}

func (o *Options) withDefaults() {
	if o.MaxOpenConns <= 0 {
		o.MaxOpenConns = 25
	}
	if o.MaxIdleConns <= 0 {
		o.MaxIdleConns = 5
	}
	if o.ConnMaxLifetime <= 0 {
		o.ConnMaxLifetime = 30 * time.Minute
	}
	if o.ConnMaxIdleTime <= 0 {
		o.ConnMaxIdleTime = 5 * time.Minute
	}
}

// DB 封装 PostgreSQL 连接池。
type DB struct {
	conn *sql.DB
}

// New 初始化 PostgreSQL 连接池。
// DSN 格式: postgres://user:pass@host:port/dbname?sslmode=disable
func New(ctx context.Context, opts Options) (*DB, error) {
	opts.withDefaults()

	conn, err := sql.Open("pgx", opts.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	conn.SetMaxOpenConns(opts.MaxOpenConns)
	conn.SetMaxIdleConns(opts.MaxIdleConns)
	conn.SetConnMaxLifetime(opts.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(opts.ConnMaxIdleTime)

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	log.Infof("[Postgres] connected, maxOpen=%d maxIdle=%d lifetime=%s idleTime=%s",
		opts.MaxOpenConns, opts.MaxIdleConns, opts.ConnMaxLifetime, opts.ConnMaxIdleTime)

	return &DB{conn: conn}, nil
}

// Close 关闭连接池。
func (db *DB) Close() error {
	log.Infof("[Postgres] closing connection pool")
	return db.conn.Close()
}

// Ping 检查数据库连通性。
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Conn 返回底层 *sql.DB，供业务层执行查询。
func (db *DB) Conn() *sql.DB {
	return db.conn
}
