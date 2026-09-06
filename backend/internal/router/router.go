// Package router 路由注册
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/handler"
	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/repository"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// Setup 组装数据层/服务层/路由树
func Setup(pool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// 数据层
	userRepo := repository.NewUserRepo(pool)

	// 服务层
	authService := service.NewAuthService(userRepo, cfg)

	// 处理层
	health := handler.NewHealthHandler(pool)
	auth := handler.NewAuthHandler(authService)

	api := r.Group("/api")
	{
		api.GET("/health", health.Ping)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.GET("/profile", middleware.AuthRequired(authService), auth.Profile)
		}
	}

	return r
}
