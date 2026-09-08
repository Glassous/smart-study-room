// Package router 路由注册
package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/imicola/smart-study-room/backend/internal/config"
	"github.com/imicola/smart-study-room/backend/internal/handler"
	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/pkg/rediscache"
	"github.com/imicola/smart-study-room/backend/internal/repository"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// Setup 组装数据层/服务层/路由树, 并返回待启动的调度器
func Setup(pool *pgxpool.Pool, rdb *redis.Client, cfg *config.Config) (*gin.Engine, *service.Scheduler) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// 缓存辅助器
	cacheHelper := rediscache.NewHelper(rdb)

	// 数据层
	userRepo := repository.NewUserRepo(pool)
	roomRepo := repository.NewRoomRepo(pool)
	seatRepo := repository.NewSeatRepo(pool)
	reservationRepo := repository.NewReservationRepo(pool)
	creditRepo := repository.NewCreditRepo(pool)
	notificationRepo := repository.NewNotificationRepo(pool)
	waitlistRepo := repository.NewWaitlistRepo(pool)
	statsRepo := repository.NewStatsRepo(pool)
	aiRepo := repository.NewAIRepo(pool)

	// 服务层
	authService := service.NewAuthService(userRepo, cfg, rdb)
	seatService := service.NewSeatService(roomRepo, seatRepo)
	seatService.SetCache(cacheHelper)
	creditService := service.NewCreditService(creditRepo, userRepo)
	notificationService := service.NewNotificationService(notificationRepo)
	reservationService := service.NewReservationService(reservationRepo, roomRepo, seatRepo, userRepo, creditService, notificationService)
	reservationService.SetRedis(rdb, cacheHelper)
	lifecycleService := service.NewLifecycleService(reservationRepo)
	lifecycleService.SetHooks(service.NewCompositeHooks(creditService, notificationService)) // 违约扣分+警告通知 / 履约加分
	lifecycleService.SetCache(cacheHelper)
	allocationService := service.NewAllocationService(seatRepo, roomRepo, userRepo, reservationRepo, reservationService)
	waitlistService := service.NewWaitlistService(waitlistRepo, seatRepo, reservationRepo, reservationService, notificationService)
	statsService := service.NewStatsService(statsRepo, seatRepo, roomRepo)
	statsService.SetCache(cacheHelper)
	aiClient := service.NewAIClient(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, cfg.AIConnectTimeout, cfg.AIResponseTimeout)
	aiService := service.NewAIService(aiRepo, userRepo, roomRepo, seatRepo, reservationRepo, notificationRepo, waitlistRepo, aiClient, cfg.AIContextMessages)

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
	adminUser := handler.NewAdminUserHandler(userRepo)
	stats := handler.NewStatsHandler(statsService)
	ai := handler.NewAIHandler(aiService)

	api := r.Group("/api")
	{
		api.GET("/health", health.Ping)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", middleware.RateLimitByIP(rdb, "login", 10, time.Minute), auth.Login)
			authGroup.POST("/logout", middleware.AuthRequired(authService), auth.Logout)
			authGroup.GET("/profile", middleware.AuthRequired(authService), auth.Profile)
		}

		// 学生侧: 房间/座位/预约(需登录)
		authorized := api.Group("", middleware.AuthRequired(authService))
		{
			authorized.GET("/rooms", room.ListRooms)
			authorized.GET("/rooms/:id/seats", room.GetSeatMap)
			authorized.POST("/reservations", middleware.RateLimitByUser(rdb, "reserve", 2, time.Second), reservation.Create)
			authorized.POST("/reservations/auto", middleware.RateLimitByUser(rdb, "reserve_auto", 2, time.Second), allocation.AutoAllocate)
			authorized.GET("/reservations/mine", reservation.ListMine)
			authorized.POST("/reservations/:id/cancel", reservation.Cancel)
			authorized.POST("/reservations/:id/checkin", reservation.Checkin)
			authorized.POST("/reservations/:id/leave", reservation.Leave)
			authorized.POST("/reservations/:id/return", reservation.ReturnBack)
			authorized.POST("/reservations/:id/checkout", reservation.Checkout)
			authorized.GET("/credit", credit.Overview)
			aiGroup := authorized.Group("/ai", middleware.RequireStudent())
			{
				aiGroup.GET("/conversations", ai.ListConversations)
				aiGroup.POST("/conversations", ai.CreateConversation)
				aiGroup.GET("/conversations/:id/messages", ai.ListMessages)
				aiGroup.DELETE("/conversations/:id", ai.DeleteConversation)
				aiGroup.POST("/chat/stream", middleware.RateLimitByUser(rdb, "ai_chat", 10, time.Minute), ai.Stream)
			}
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
			statsGroup := authorized.Group("/stats")
			{
				statsGroup.GET("/heatmap", stats.Heatmap)
				statsGroup.GET("/trend", stats.Trend)
				statsGroup.GET("/peak", stats.Peak)
				statsGroup.GET("/top-seats", stats.TopSeats)
				statsGroup.GET("/overview", stats.Overview)
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
			adminGroup.GET("/users", adminUser.ListUsers)
			adminGroup.PUT("/users/:id/status", adminUser.SetUserStatus)
		}
	}

	scheduler := service.NewScheduler(lifecycleService, waitlistService, time.Minute)
	scheduler.SetRedis(rdb)
	return r, scheduler
}
