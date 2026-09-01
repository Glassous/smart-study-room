// Package router 路由注册
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/handler"
	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Setup 组装路由树
func Setup(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	api := r.Group("/api")
	{
		health := handler.NewHealthHandler(pool)
		api.GET("/health", health.Ping)
	}

	return r
}
