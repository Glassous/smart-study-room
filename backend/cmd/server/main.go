// 智能共享自习室预约系统 后端服务入口
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/repository"
	"github.com/imicola/smart-study-room/backend/internal/router"
)

func main() {
	cfg := config.Load()

	pool, err := repository.NewPool(context.Background(), cfg.DBUrl)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer pool.Close()

	// 初始化 Redis 客户端(支持平稳降级)
	rdb, err := repository.NewRedisClient(context.Background(), cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("提示: 连接 Redis 失败(%v), 系统已降级运行(纯 PostgreSQL 兜底模式)", err)
		rdb = nil
	} else {
		defer rdb.Close()
		log.Printf("Redis 连接成功: %s (db=%d)", cfg.RedisAddr, cfg.RedisDB)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r, scheduler := router.Setup(pool, rdb, cfg)

	// 启动后台调度器(违约扫描/到时完成/临时离开超时)
	go scheduler.Run(ctx)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() {
		log.Printf("studyroom-server 监听 :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务退出: %v", err)
		}
	}()

	// 优雅关闭: 等待终止信号 → 停调度器 → 关 HTTP → 释放连接池
	<-ctx.Done()
	log.Printf("收到终止信号, 正在优雅关闭 ...")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP 关闭异常: %v", err)
	}
	os.Exit(0)
}
