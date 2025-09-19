package auth

// 文件说明: Auth 控制器构造与版本选择。

import (
	"teamspeak-one-click-deploy/api/auth"
)

// ControllerV1 实现 IAuthV1 接口。
// 负责登录、刷新与登出。
type ControllerV1 struct{}

// NewV1 返回 IAuthV1 的实现，用于路由绑定。
func NewV1() auth.IAuthV1 {
	return &ControllerV1{}
}
