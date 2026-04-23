package postgres

import (
	"context"
	"database/sql"
)

// RunMigrations 执行指定目录下的 SQL migration 文件。
// migrationsDir: migration 文件所在目录路径（如 "./migrations"）
// 按文件名字典序执行（001_init.sql, 002_xxx.sql...）。
//
// 设计说明：
// - 不使用 embed，支持独立 migration 工具（如 golang-migrate）
// - 根目录 migrations/ 可被 Docker 挂载、CI/CD 独立执行
// - 生产环境建议使用专业 migration 工具，此函数仅供开发/测试
func RunMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	// TODO: 实现 migration 执行逻辑
	// 1. 读取 migrationsDir 目录下的 .sql 文件
	// 2. 按文件名排序
	// 3. 创建 schema_migrations 表记录已执行的版本
	// 4. 跳过已执行的，逐个执行新的 migration
	// 5. 每个 migration 在事务中执行，失败则回滚
	return nil
}
