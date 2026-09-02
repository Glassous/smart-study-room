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

// WaitlistHandler 候补队列接口
type WaitlistHandler struct {
	waitlist *service.WaitlistService
}

func NewWaitlistHandler(waitlist *service.WaitlistService) *WaitlistHandler {
	return &WaitlistHandler{waitlist: waitlist}
}

// Join POST /api/waitlist
func (h *WaitlistHandler) Join(c *gin.Context) {
	var req model.JoinWaitlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	view, err := h.waitlist.Join(c.Request.Context(), middleware.CurrentUID(c), &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyWaiting):
			fail(c, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrBadSlot):
			fail(c, http.StatusBadRequest, err.Error())
		default:
			fail(c, http.StatusInternalServerError, "加入候补失败")
		}
		return
	}
	ok(c, view)
}

// ListMine GET /api/waitlist/mine
func (h *WaitlistHandler) ListMine(c *gin.Context) {
	list, err := h.waitlist.ListMine(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询候补失败")
		return
	}
	ok(c, list)
}

// Cancel POST /api/waitlist/:id/cancel
func (h *WaitlistHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.waitlist.Cancel(c.Request.Context(), middleware.CurrentUID(c), id); err != nil {
		switch {
		case errors.Is(err, service.ErrWaitNotFound):
			fail(c, http.StatusNotFound, err.Error())
		case errors.Is(err, service.ErrNotOwner):
			fail(c, http.StatusForbidden, err.Error())
		case errors.Is(err, service.ErrWaitInvalidState):
			fail(c, http.StatusConflict, err.Error())
		default:
			fail(c, http.StatusInternalServerError, "操作失败")
		}
		return
	}
	ok(c, gin.H{"cancelled": id})
}
