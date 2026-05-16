package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"ai_interview/internal/log"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Options configures the PostgreSQL connection pool.
type Options struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
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

// DB wraps the GORM handle and its underlying connection pool.
type DB struct {
	gorm *gorm.DB
	conn *sql.DB
}

// New initializes PostgreSQL.
// DSN format: postgres://user:pass@host:port/dbname?sslmode=disable
func New(ctx context.Context, opts Options) (*DB, error) {
	opts.withDefaults()

	gormDB, err := gorm.Open(gormpg.Open(opts.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	conn, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("get postgres sql db: %w", err)
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

	return &DB{gorm: gormDB, conn: conn}, nil
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	log.Infof("[Postgres] closing connection pool")
	return db.conn.Close()
}

// Ping checks database connectivity.
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Conn returns the underlying *sql.DB for migrations and connection pool control.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Gorm returns the GORM handle for runtime business repositories.
func (db *DB) Gorm() *gorm.DB {
	return db.gorm
}
