package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// AdminRoomHandler 管理端房间/座位维护接口(RBAC: 仅 admin)
type AdminRoomHandler struct {
	seats *service.SeatService
}

func NewAdminRoomHandler(seats *service.SeatService) *AdminRoomHandler {
	return &AdminRoomHandler{seats: seats}
}

// CreateRoom POST /api/admin/rooms
func (h *AdminRoomHandler) CreateRoom(c *gin.Context) {
	var req model.RoomUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	room, err := h.seats.CreateRoom(c.Request.Context(), &req)
	h.respondRoom(c, room, err)
}

// UpdateRoom PUT /api/admin/rooms/:id
func (h *AdminRoomHandler) UpdateRoom(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.RoomUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	room, err := h.seats.UpdateRoom(c.Request.Context(), id, &req)
	h.respondRoom(c, room, err)
}

// DeleteRoom DELETE /api/admin/rooms/:id
func (h *AdminRoomHandler) DeleteRoom(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.seats.DeleteRoom(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrRoomNotFound) {
			fail(c, http.StatusNotFound, err.Error())
			return
		}
		fail(c, http.StatusInternalServerError, "删除房间失败")
		return
	}
	ok(c, gin.H{"deleted": id})
}

// BatchGenSeats POST /api/admin/rooms/:id/seats/batch
func (h *AdminRoomHandler) BatchGenSeats(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.BatchGenSeatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	n, err := h.seats.BatchGenSeats(c.Request.Context(), id, req.SeatRows, req.SeatCols)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			fail(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrRoomHasSeats):
			fail(c, http.StatusConflict, err.Error())
		default:
			fail(c, http.StatusInternalServerError, "批量生成座位失败")
		}
		return
	}
	ok(c, gin.H{"room_id": id, "created": n})
}

// UpdateSeat PUT /api/admin/seats/:id
func (h *AdminRoomHandler) UpdateSeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.SeatUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	if err := h.seats.UpdateSeat(c.Request.Context(), id, &req); err != nil {
		if errors.Is(err, service.ErrSeatNotFound) {
			fail(c, http.StatusNotFound, err.Error())
			return
		}
		fail(c, http.StatusInternalServerError, "更新座位失败")
		return
	}
	ok(c, gin.H{"updated": id})
}

// DeleteSeat DELETE /api/admin/seats/:id
func (h *AdminRoomHandler) DeleteSeat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.seats.DeleteSeat(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrSeatNotFound) {
			fail(c, http.StatusNotFound, err.Error())
			return
		}
		fail(c, http.StatusInternalServerError, "删除座位失败")
		return
	}
	ok(c, gin.H{"deleted": id})
}

func (h *AdminRoomHandler) respondRoom(c *gin.Context, room any, err error) {
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			fail(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrBadTimeRange):
			fail(c, http.StatusBadRequest, err.Error())
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			fail(c, http.StatusConflict, "房间名称已存在")
		default:
			fail(c, http.StatusInternalServerError, "保存房间失败")
		}
		return
	}
	ok(c, room)
}
