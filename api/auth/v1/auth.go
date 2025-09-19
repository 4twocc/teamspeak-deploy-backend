package v1

// Auth V1 版本的 DTO 定义，包含登录、刷新、登出请求与响应结构体。

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LoginReq 用户名/邮箱/手机号 + 密码登录。
type LoginReq struct {
	g.Meta `path:"/login" method:"post" tags:"Auth" summary:"用户登录"`
	// 允许用户名/邮箱/手机号三选一
	Account  string `json:"account" v:"required" dc:"用户名/邮箱/手机号"`
	Password string `json:"password" v:"required|length:6,128" dc:"密码，长度6-128字符"`
}

type LoginRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// RefreshReq 使用刷新令牌获取新的访问令牌。
type RefreshReq struct {
	g.Meta       `path:"/refresh" method:"post" tags:"Auth" summary:"刷新访问令牌"`
	RefreshToken string `json:"refreshToken" v:"required"`
}

type RefreshRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// LogoutReq 登出。
type LogoutReq struct {
	g.Meta `path:"/logout" method:"post" tags:"Auth" summary:"登出"`
}

type LogoutRes struct{}
