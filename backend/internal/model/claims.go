package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷: 标识用户身份与角色
type Claims struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewClaims 构造带过期时间的载荷
func NewClaims(u *User, issuedAt, expiresAt time.Time) *Claims {
	return &Claims{
		UID:      u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "studyroom",
		},
	}
}
