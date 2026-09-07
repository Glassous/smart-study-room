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
	roomRepo := repository.NewRoomRepo(pool)
	seatRepo := repository.NewSeatRepo(pool)

	// 服务层
	authService := service.NewAuthService(userRepo, cfg)
	seatService := service.NewSeatService(roomRepo, seatRepo)

	// 处理层
	health := handler.NewHealthHandler(pool)
	auth := handler.NewAuthHandler(authService)
	room := handler.NewRoomHandler(seatService)
	adminRoom := handler.NewAdminRoomHandler(seatService)

	api := r.Group("/api")
	{
		api.GET("/health", health.Ping)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.GET("/profile", middleware.AuthRequired(authService), auth.Profile)
		}

		// 学生侧: 房间与座位平面图(需登录)
		authorized := api.Group("", middleware.AuthRequired(authService))
		{
			authorized.GET("/rooms", room.ListRooms)
			authorized.GET("/rooms/:id/seats", room.GetSeatMap)
		}

		// 管理端: 房间/座位维护(仅 admin)
		adminGroup := api.Group("/admin",
			middleware.AuthRequired(authService), middleware.RequireAdmin())
		{
			adminGroup.POST("/rooms", adminRoom.CreateRoom)
			adminGroup.PUT("/rooms/:id", adminRoom.UpdateRoom)
			adminGroup.DELETE("/rooms/:id", adminRoom.DeleteRoom)
			adminGroup.POST("/rooms/:id/seats/batch", adminRoom.BatchGenSeats)
			adminGroup.PUT("/seats/:id", adminRoom.UpdateSeat)
			adminGroup.DELETE("/seats/:id", adminRoom.DeleteSeat)
		}
	}

	return r
}
