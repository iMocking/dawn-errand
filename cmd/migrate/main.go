// cmd/migrate 执行手工 SQL 迁移（up/down），以 schema_migrations 表记录已执行迁移。
// 用法：go run ./cmd/migrate up|down
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"dawn-errand/internal/config"
)

func main() {
	if len(os.Args) < 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		fmt.Println("用法: go run ./cmd/migrate up|down")
		os.Exit(1)
	}
	dir := os.Args[1]
	cfg := config.Load()
	if err := run(context.Background(), cfg.PostgresDSN, dir); err != nil {
		fmt.Fprintln(os.Stderr, "迁移失败:", err)
		os.Exit(1)
	}
	fmt.Println("迁移完成:", dir)
}

func run(ctx context.Context, dsn, dir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version varchar(64) PRIMARY KEY,
		applied_at bigint NOT NULL
	)`); err != nil {
		return err
	}

	if dir == "up" {
		return applyUp(ctx, db)
	}
	return applyDown(ctx, db)
}

func applyUp(ctx context.Context, db *sql.DB) error {
	var exists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version='001_init'`).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		fmt.Println("001_init 已应用，跳过")
		return nil
	}
	script, err := os.ReadFile("migrations/001_init.up.sql")
	if err != nil {
		return fmt.Errorf("读取迁移文件: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, string(script)); err != nil {
		return fmt.Errorf("执行 001_init.up.sql: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, applied_at) VALUES ('001_init', $1)`,
		time.Now().Unix()); err != nil {
		return err
	}
	return tx.Commit()
}

func applyDown(ctx context.Context, db *sql.DB) error {
	var exists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version='001_init'`).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		fmt.Println("schema_migrations 无 001_init 记录，跳过回滚")
		return nil
	}
	script, err := os.ReadFile("migrations/001_init.down.sql")
	if err != nil {
		return fmt.Errorf("读取回滚文件: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, string(script)); err != nil {
		return fmt.Errorf("执行 001_init.down.sql: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version='001_init'`); err != nil {
		return err
	}
	return tx.Commit()
}
