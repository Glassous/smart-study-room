package handler

import (
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
	res *service.ReservationService
}

func NewReservationHandler(res *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{res: res}
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
	default:
		fail(c, http.StatusInternalServerError, "服务内部错误")
	}
}
