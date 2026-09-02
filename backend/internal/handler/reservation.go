package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// ReservationHandler 预约接口
type ReservationHandler struct {
	res  *service.ReservationService
	life *service.LifecycleService
}

func NewReservationHandler(res *service.ReservationService, life *service.LifecycleService) *ReservationHandler {
	return &ReservationHandler{res: res, life: life}
}

// Create POST /api/reservations
func (h *ReservationHandler) Create(c *gin.Context) {
	var req model.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	view, err := h.res.Create(c.Request.Context(), middleware.CurrentUID(c), &req)
	if err != nil {
		respondResError(c, err)
		return
	}
	ok(c, view)
}

// Cancel POST /api/reservations/:id/cancel
func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.res.Cancel(c.Request.Context(), middleware.CurrentUID(c), id); err != nil {
		respondResError(c, err)
		return
	}
	ok(c, gin.H{"cancelled": id})
}

// ListMine GET /api/reservations/mine
func (h *ReservationHandler) ListMine(c *gin.Context) {
	list, err := h.res.ListMine(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询预约列表失败")
		return
	}
	if list == nil {
		list = []*model.ReservationView{}
	}
	ok(c, list)
}

// Checkin POST /api/reservations/:id/checkin
func (h *ReservationHandler) Checkin(c *gin.Context) {
	h.lifecycleAction(c, h.life.Checkin, "checked_in")
}

// Leave POST /api/reservations/:id/leave
func (h *ReservationHandler) Leave(c *gin.Context) {
	h.lifecycleAction(c, h.life.Leave, "temp_leave")
}

// ReturnBack POST /api/reservations/:id/return
func (h *ReservationHandler) ReturnBack(c *gin.Context) {
	h.lifecycleAction(c, h.life.ReturnBack, "checked_in")
}

// Checkout POST /api/reservations/:id/checkout
func (h *ReservationHandler) Checkout(c *gin.Context) {
	h.lifecycleAction(c, h.life.Checkout, "completed")
}

func (h *ReservationHandler) lifecycleAction(
	c *gin.Context,
	fn func(ctx context.Context, uid, id int64) error,
	action string,
) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := fn(c.Request.Context(), middleware.CurrentUID(c), id); err != nil {
		respondResError(c, err)
		return
	}
	ok(c, gin.H{"id": id, "action": action})
}

// 预约错误统一映射
func respondResError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrBadSlot),
		errors.Is(err, service.ErrTooLong):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrOutOfOpenTime),
		errors.Is(err, service.ErrInPast):
		fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrSeatConflict),
		errors.Is(err, service.ErrUserConflict):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrSeatUnavailable):
		fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrCreditBanned):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrResNotFound):
		fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrNotOwner):
		fail(c, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrInvalidState):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrNotInCheckinWindow),
		errors.Is(err, service.ErrLeaveTimeout):
		fail(c, http.StatusBadRequest, err.Error())
	default:
		fail(c, http.StatusInternalServerError, "服务内部错误")
	}
}
