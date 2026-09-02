package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// NotificationHandler 消息通知接口
type NotificationHandler struct {
	notify *service.NotificationService
}

func NewNotificationHandler(notify *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notify: notify}
}

// List GET /api/notifications
func (h *NotificationHandler) List(c *gin.Context) {
	list, err := h.notify.List(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询通知失败")
		return
	}
	ok(c, list)
}

// UnreadCount GET /api/notifications/unread_count
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	n, err := h.notify.UnreadCount(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询未读数失败")
		return
	}
	ok(c, gin.H{"count": n})
}

// MarkRead POST /api/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.notify.MarkRead(c.Request.Context(), middleware.CurrentUID(c), id); err != nil {
		fail(c, http.StatusNotFound, err.Error())
		return
	}
	ok(c, gin.H{"read": id})
}

// MarkAllRead POST /api/notifications/read_all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	n, err := h.notify.MarkAllRead(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "操作失败")
		return
	}
	ok(c, gin.H{"marked": n})
}
