package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/imicola/smart-study-room/backend/internal/middleware"
	"github.com/imicola/smart-study-room/backend/internal/model"
	"github.com/imicola/smart-study-room/backend/internal/repository"
)

// AdminUserHandler 管理端用户管理接口(RBAC: 仅 admin)
type AdminUserHandler struct {
	users *repository.UserRepo
}

func NewAdminUserHandler(users *repository.UserRepo) *AdminUserHandler {
	return &AdminUserHandler{users: users}
}

// ListUsers GET /api/admin/users
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	list, err := h.users.ListAll(c.Request.Context(), 200)
	if err != nil {
		fail(c, http.StatusInternalServerError, "查询用户列表失败")
		return
	}
	pubs := make([]*model.UserPublic, 0, len(list))
	for _, u := range list {
		pubs = append(pubs, u.ToPublic())
	}
	ok(c, pubs)
}

// SetUserStatus PUT /api/admin/users/:id/status  body: {status: "active"|"disabled"}
func (h *AdminUserHandler) SetUserStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status string `json:"status" binding:"required,oneof=active disabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}
	// 防止管理员误禁用自己
	if id == middleware.CurrentUID(c) {
		fail(c, http.StatusBadRequest, "不能禁用当前登录账号")
		return
	}
	if err := h.users.SetStatus(c.Request.Context(), id, req.Status); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			fail(c, http.StatusNotFound, "用户不存在")
			return
		}
		fail(c, http.StatusInternalServerError, "操作失败")
		return
	}
	ok(c, gin.H{"id": id, "status": req.Status})
}
