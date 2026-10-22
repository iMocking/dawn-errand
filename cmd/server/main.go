// cmd/server 曙光跑腿 SaaS 后端服务入口：装配依赖、初始化 schema（dev）、启动 HTTP 服务并优雅退出。
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"

	"dawn-errand/internal/assembly"
	"dawn-errand/internal/config"
	apihandlers "dawn-errand/internal/interfaces/http"
)

func main() {
	cfg := config.Load()
	app, err := assembly.Build(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "装配失败:", err)
		os.Exit(1)
	}
	defer app.Client.Close()
	defer app.Rdb.Close()

	// 开发环境便捷迁移：ent 自动迁移（幂等）。生产环境请使用 go run ./cmd/migrate up 执行 SQL 迁移。
	if err := app.Client.Schema.Create(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "schema 迁移失败:", err)
		os.Exit(1)
	}
	if cfg.AppEnv != "dev" {
		app.Logger.Warn("非 dev 环境：请使用 go run ./cmd/migrate up 执行手工 SQL 迁移以保证可审查/可回滚")
	}

	router := apihandlers.NewRouter(app.Handler, app.MW)
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		app.Logger.Info("server started", zap.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger.Error("server error", zap.Error(err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	app.Logger.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		app.Logger.Error("graceful shutdown failed", zap.Error(err))
	}
	app.Logger.Info("server stopped")
}
