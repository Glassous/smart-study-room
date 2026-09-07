package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

// AuthHandler 认证接口
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	pub, err := h.auth.Register(c.Request.Context(), &req)
	if err != nil {
		respondAuthError(c, err)
		return
	}
	ok(c, pub)
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	resp, err := h.auth.Login(c.Request.Context(), &req)
	if err != nil {
		respondAuthError(c, err)
		return
	}
	ok(c, resp)
}

// Logout POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if err := h.auth.Logout(c.Request.Context(), tokenStr); err != nil {
			fail(c, http.StatusServiceUnavailable, "注销服务暂时不可用，请稍后重试")
			return
		}
	}
	ok(c, gin.H{"message": "已安全退出"})
}

// Profile GET /api/auth/profile
func (h *AuthHandler) Profile(c *gin.Context) {
	pub, err := h.auth.Profile(c.Request.Context(), middleware.CurrentUID(c))
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询失败")
		return
	}
	ok(c, pub)
}

func respondAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUsernameTaken):
		fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrBadCredentials):
		fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrUserDisabled):
		fail(c, http.StatusForbidden, err.Error())
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "服务内部错误", "data": nil})
	}
}
