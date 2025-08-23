package auth

import (
	"time"

	"teamspeak-one-click-deploy/user"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 定义 JWT 声明
// 包含标准声明和自定义的用户信息
type Claims struct {
	UID      uint   `json:"uid"`
	Username string `json:"username"`
	Role     uint8  `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
// 用于解析登录请求的 JSON 数据
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
// 用于解析用户注册请求的 JSON 数据
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Nickname string `json:"nickname" binding:"max=50"`
}

// LoginResponse 登录响应
// 返回给客户端的登录成功信息
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *UserInfo `json:"user"`
}

// UserInfo 用户信息
// 返回给客户端的用户信息
type UserInfo struct {
	UID       uint       `json:"uid"`
	Username  string     `json:"username"`
	Nickname  string     `json:"nickname"`
	Email     string     `json:"email"`
	Role      uint8      `json:"role"`
	Status    uint8      `json:"status"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// 将User模型转换为UserInfo模型
func NewUserInfo(user *user.User) *UserInfo {
	return &UserInfo{
		UID:       user.UID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		LastLogin: user.LastLogin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
