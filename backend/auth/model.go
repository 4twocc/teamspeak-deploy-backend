package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 定义 JWT 声明
// 包含标准声明和自定义的用户信息
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
// 用于解析登录请求的 JSON 数据
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	LastLogin time.Time `json:"last_login,omitempty"`
}

// Config 认证配置
type Config struct {
	JWTSecret   string        `yaml:"jwt_secret"`
	ExpiresIn   time.Duration `yaml:"expires_in"`
	TokenPrefix string        `yaml:"token_prefix"`
}

// DefaultConfig 返回默认的认证配置
func DefaultConfig() *Config {
	return &Config{
		JWTSecret:   "your-secret-key-change-in-production",
		ExpiresIn:   24 * time.Hour,
		TokenPrefix: "Bearer ",
	}
}
