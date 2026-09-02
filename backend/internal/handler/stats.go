package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/service"
)

// StatsHandler 统计分析接口
type StatsHandler struct {
	stats *service.StatsService
}

func NewStatsHandler(stats *service.StatsService) *StatsHandler {
	return &StatsHandler{stats: stats}
}

// Heatmap GET /api/stats/heatmap?room_id=&date=
func (h *StatsHandler) Heatmap(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Query("room_id"), 10, 64)
	if roomID <= 0 {
		fail(c, http.StatusBadRequest, "缺少 room_id 参数")
		return
	}
	resp, err := h.stats.Heatmap(c.Request.Context(), roomID, c.Query("date"))
	if err != nil {
		if errors.Is(err, service.ErrStatsBadArg) || errors.Is(err, service.ErrRoomNotFound) {
			fail(c, http.StatusBadRequest, err.Error())
			return
		}
		fail(c, http.StatusInternalServerError, "生成热力图失败")
		return
	}
	ok(c, resp)
}

// Trend GET /api/stats/trend?days=14
func (h *StatsHandler) Trend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	points, err := h.stats.Trend(c.Request.Context(), days)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询趋势失败")
		return
	}
	ok(c, points)
}

// Peak GET /api/stats/peak?days=14
func (h *StatsHandler) Peak(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	points, err := h.stats.Peak(c.Request.Context(), days)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询高峰时段失败")
		return
	}
	ok(c, points)
}

// TopSeats GET /api/stats/top-seats?days=14&limit=10
func (h *StatsHandler) TopSeats(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "14"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	list, err := h.stats.TopSeats(c.Request.Context(), days, limit)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询热门座位失败")
		return
	}
	ok(c, list)
}

// Overview GET /api/stats/overview
func (h *StatsHandler) Overview(c *gin.Context) {
	ov, err := h.stats.Overview(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询统计总览失败")
		return
	}
	ok(c, ov)
}
