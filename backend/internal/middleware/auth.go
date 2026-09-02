package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/service"
)

const (
	ctxUID      = "uid"
	ctxUsername = "username"
	ctxRole     = "role"
)

// AuthRequired JWT 认证中间件: 解析 Bearer 令牌并注入用户身份
func AuthRequired(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abort(c, http.StatusUnauthorized, "缺少 Bearer 令牌")
			return
		}
		claims, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			abort(c, http.StatusUnauthorized, "令牌无效或已过期")
			return
		}
		c.Set(ctxUID, claims.UID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireAdmin RBAC: 仅管理员可访问
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if CurrentRole(c) != model.RoleAdmin {
			abort(c, http.StatusForbidden, "需要管理员权限")
			return
		}
		c.Next()
	}
}

// CurrentUID 从上下文取当前用户 ID
func CurrentUID(c *gin.Context) int64 {
	v, _ := c.Get(ctxUID)
	uid, _ := v.(int64)
	return uid
}

// CurrentRole 从上下文取当前角色
func CurrentRole(c *gin.Context) string {
	v, _ := c.Get(ctxRole)
	role, _ := v.(string)
	return role
}

func abort(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"code": status, "message": msg, "data": nil})
}
