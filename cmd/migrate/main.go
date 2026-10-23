// cmd/migrate 执行手工 SQL 迁移（up/down），以 schema_migrations 表记录已执行迁移。
// 按 migrations/ 目录内 *.up.sql / *.down.sql 文件名排序逐版本执行（如 001_init、002_xxx）。
// 用法：go run ./cmd/migrate up|down
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
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
		return apply(ctx, db, "up")
	}
	return apply(ctx, db, "down")
}

// listVersions 读取 migrations/ 下全部迁移版本（去掉 .up.sql/.down.sql 后缀），升序。
func listVersions() ([]string, error) {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("读取 migrations 目录: %w", err)
	}
	seen := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(name, ".up.sql") {
			seen[strings.TrimSuffix(name, ".up.sql")] = true
		} else if strings.HasSuffix(name, ".down.sql") {
			seen[strings.TrimSuffix(name, ".down.sql")] = true
		}
	}
	vers := make([]string, 0, len(seen))
	for v := range seen {
		vers = append(vers, v)
	}
	sort.Strings(vers)
	return vers, nil
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

func apply(ctx context.Context, db *sql.DB, dir string) error {
	vers, err := listVersions()
	if err != nil {
		return err
	}
	done, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}
	// down 反向执行。
	if dir == "down" {
		for i, j := 0, len(vers)-1; i < j; i, j = i+1, j-1 {
			vers[i], vers[j] = vers[j], vers[i]
		}
	}
	for _, v := range vers {
		if dir == "up" && done[v] {
			fmt.Println(v, "已应用，跳过")
			continue
		}
		if dir == "down" && !done[v] {
			continue
		}
		suffix := ".up.sql"
		if dir == "down" {
			suffix = ".down.sql"
		}
		script, err := os.ReadFile("migrations/" + v + suffix)
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s: %w", v+suffix, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("执行 %s: %w", v+suffix, err)
		}
		if dir == "up" {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`,
				v, time.Now().Unix()); err != nil {
				_ = tx.Rollback()
				return err
			}
		} else {
			if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version=$1`, v); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		fmt.Println("已执行:", v, suffix)
	}
	return nil
}
