package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// AllocationHandler 智能自动分配接口
type AllocationHandler struct {
	alloc *service.AllocationService
}

func NewAllocationHandler(alloc *service.AllocationService) *AllocationHandler {
	return &AllocationHandler{alloc: alloc}
}

// AutoAllocate POST /api/reservations/auto
// body: {room_id, date, start_time, end_time, zone?, has_power?, near_window?, top_n?, auto_book?}
func (h *AllocationHandler) AutoAllocate(c *gin.Context) {
	var req model.AllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.alloc.AutoAllocate(c.Request.Context(), middleware.CurrentUID(c), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoCandidate):
			fail(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrUserConflict):
			fail(c, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrCreditBanned):
			fail(c, http.StatusForbidden, err.Error())
		default:
			respondResError(c, err)
		}
		return
	}
	ok(c, resp)
}
