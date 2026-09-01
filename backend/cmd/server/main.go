// 智能共享自习室预约系统 后端服务入口
package main

import (
	"context"
	"log"
	"net/http"

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

	r := router.Setup(pool, cfg)
	log.Printf("studyroom-server 监听 :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}
