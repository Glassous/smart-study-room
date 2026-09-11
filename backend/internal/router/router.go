// Package router 路由注册
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/handler"
	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/repository"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// Setup 组装数据层/服务层/路由树, 并返回待启动的调度器
func Setup(pool *pgxpool.Pool, cfg *config.Config) (*gin.Engine, *service.Scheduler) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// 数据层
	userRepo := repository.NewUserRepo(pool)
	roomRepo := repository.NewRoomRepo(pool)
	seatRepo := repository.NewSeatRepo(pool)
	reservationRepo := repository.NewReservationRepo(pool)
	creditRepo := repository.NewCreditRepo(pool)
	notificationRepo := repository.NewNotificationRepo(pool)
	waitlistRepo := repository.NewWaitlistRepo(pool)

	// 服务层
	authService := service.NewAuthService(userRepo, cfg)
	seatService := service.NewSeatService(roomRepo, seatRepo)
	creditService := service.NewCreditService(creditRepo, userRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	reservationService := service.NewReservationService(reservationRepo, roomRepo, seatRepo, userRepo, creditService, notificationService)
	lifecycleService := service.NewLifecycleService(reservationRepo)
	lifecycleService.SetHooks(service.NewCompositeHooks(creditService, notificationService)) // 违约扣分+警告通知 / 履约加分
	allocationService := service.NewAllocationService(seatRepo, roomRepo, userRepo, reservationRepo, reservationService)
	waitlistService := service.NewWaitlistService(waitlistRepo, seatRepo, reservationRepo, reservationService, notificationService)

	// 处理层
	health := handler.NewHealthHandler(pool)
	auth := handler.NewAuthHandler(authService)
	room := handler.NewRoomHandler(seatService)
	adminRoom := handler.NewAdminRoomHandler(seatService)
	reservation := handler.NewReservationHandler(reservationService, lifecycleService)
	allocation := handler.NewAllocationHandler(allocationService)
	credit := handler.NewCreditHandler(creditService)
	notify := handler.NewNotificationHandler(notificationService)
	waitlist := handler.NewWaitlistHandler(waitlistService)

	api := r.Group("/api")
	{
		api.GET("/health", health.Ping)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.GET("/profile", middleware.AuthRequired(authService), auth.Profile)
		}

		// 学生侧: 房间/座位/预约(需登录)
		authorized := api.Group("", middleware.AuthRequired(authService))
		{
			authorized.GET("/rooms", room.ListRooms)
			authorized.GET("/rooms/:id/seats", room.GetSeatMap)
			authorized.POST("/reservations", reservation.Create)
			authorized.POST("/reservations/auto", allocation.AutoAllocate)
			authorized.GET("/reservations/mine", reservation.ListMine)
			authorized.POST("/reservations/:id/cancel", reservation.Cancel)
			authorized.POST("/reservations/:id/checkin", reservation.Checkin)
			authorized.POST("/reservations/:id/leave", reservation.Leave)
			authorized.POST("/reservations/:id/return", reservation.ReturnBack)
			authorized.POST("/reservations/:id/checkout", reservation.Checkout)
			authorized.GET("/credit", credit.Overview)
			notifyGroup := authorized.Group("/notifications")
			{
				notifyGroup.GET("", notify.List)
				notifyGroup.GET("/unread_count", notify.UnreadCount)
				notifyGroup.POST("/:id/read", notify.MarkRead)
				notifyGroup.POST("/read_all", notify.MarkAllRead)
			}
			waitlistGroup := authorized.Group("/waitlist")
			{
				waitlistGroup.POST("", waitlist.Join)
				waitlistGroup.GET("/mine", waitlist.ListMine)
				waitlistGroup.POST("/:id/cancel", waitlist.Cancel)
			}
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

	return r, service.NewScheduler(lifecycleService, waitlistService, time.Minute)
}
