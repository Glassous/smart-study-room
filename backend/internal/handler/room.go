package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/service"
)

// RoomHandler 学生侧房间/座位查询接口
type RoomHandler struct {
	seats *service.SeatService
}

func NewRoomHandler(seats *service.SeatService) *RoomHandler {
	return &RoomHandler{seats: seats}
}

// ListRooms GET /api/rooms
func (h *RoomHandler) ListRooms(c *gin.Context) {
	rooms, err := h.seats.ListRooms(c.Request.Context())
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询房间列表失败")
		return
	}
	ok(c, rooms)
}

// GetSeatMap GET /api/rooms/:id/seats?date=&start=&end=
func (h *RoomHandler) GetSeatMap(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	resp, err := h.seats.GetSeatMap(c.Request.Context(), id,
		c.Query("date"), c.Query("start"), c.Query("end"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			fail(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrBadSeatMapArg):
			fail(c, http.StatusBadRequest, err.Error())
		default:
			fail(c, http.StatusInternalServerError, "查询座位平面图失败")
		}
		return
	}
	ok(c, resp)
}
