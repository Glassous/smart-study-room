package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// CreditHandler 信用分接口
type CreditHandler struct {
	credit *service.CreditService
}

func NewCreditHandler(credit *service.CreditService) *CreditHandler {
	return &CreditHandler{credit: credit}
}

// Overview GET /api/credit
func (h *CreditHandler) Overview(c *gin.Context) {
	ov, err := h.credit.GetOverview(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询信用信息失败")
		return
	}
	ok(c, ov)
}
