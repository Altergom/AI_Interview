package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB 封装 PostgreSQL 连接池。
type DB struct {
	conn *sql.DB
}

// New 初始化 PostgreSQL 连接池。
// dsn 格式: postgres://user:pass@host:port/dbname?sslmode=disable
func New(ctx context.Context, dsn string) (*DB, error) {
	// TODO: 实现连接池初始化
	return nil, nil
}

// Close 关闭连接池。
func (db *DB) Close() error {
	// TODO: 实现关闭逻辑
	return nil
}

// Ping 检查数据库连通性。
func (db *DB) Ping(ctx context.Context) error {
	// TODO: 实现 ping
	return nil
}

// Conn 返回底层 *sql.DB，供业务层执行查询。
func (db *DB) Conn() *sql.DB {
	return db.conn
}
