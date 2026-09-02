// Package model 领域模型与传输对象
package model

import "time"

// Role 角色枚举
const (
	RoleStudent = "student"
	RoleAdmin   = "admin"
)

// UserStatus 账号状态
const (
	UserActive   = "active"
	UserDisabled = "disabled"
)

// User 用户表模型
type User struct {
	ID                int64      `json:"id"`
	Username          string     `json:"username"`
	PasswordHash      string     `json:"-"`
	RealName          string     `json:"real_name"`
	StudentNo         *string    `json:"student_no"`
	Role              string     `json:"role"`
	CreditScore       int        `json:"credit_score"`
	CreditBannedUntil *time.Time `json:"credit_banned_until"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UserPublic 用户公开信息(不含口令哈希)
type UserPublic struct {
	ID                int64      `json:"id"`
	Username          string     `json:"username"`
	RealName          string     `json:"real_name"`
	StudentNo         *string    `json:"student_no"`
	Role              string     `json:"role"`
	CreditScore       int        `json:"credit_score"`
	CreditBannedUntil *time.Time `json:"credit_banned_until"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ToPublic 转为公开信息
func (u *User) ToPublic() *UserPublic {
	return &UserPublic{
		ID:                u.ID,
		Username:          u.Username,
		RealName:          u.RealName,
		StudentNo:         u.StudentNo,
		Role:              u.Role,
		CreditScore:       u.CreditScore,
		CreditBannedUntil: u.CreditBannedUntil,
		Status:            u.Status,
		CreatedAt:         u.CreatedAt,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=6,max=64"`
	RealName  string `json:"real_name" binding:"required,max=50"`
	StudentNo string `json:"student_no" binding:"omitempty,max=20"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      *UserPublic `json:"user"`
}
