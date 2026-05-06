package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultConnMaxIdleTime = 5 * time.Minute
)

// DB 封装 PostgreSQL 连接池。
type DB struct {
	conn *sql.DB
}

// New 初始化 PostgreSQL 连接池。
// dsn 格式: postgres://user:pass@host:port/dbname?sslmode=disable
func New(ctx context.Context, dsn string) (*DB, error) {
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	conn.SetMaxOpenConns(defaultMaxOpenConns)
	conn.SetMaxIdleConns(defaultMaxIdleConns)
	conn.SetConnMaxLifetime(defaultConnMaxLifetime)
	conn.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close 关闭连接池。
func (db *DB) Close() error {
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
