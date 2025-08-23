package utils

// 用户账号状态
const (
	AccountStatusActive = iota // 账号活跃
	AccountStatusBanned        // 账号禁止
	AccountStatusLocked        // 账号锁定
)

// 用户账号角色
const (
	AccountRoleVisitor  = 8  // 普通用户
	AccountRoleOperator = 16 // 运营者
	AccountRoleAdmin    = 32 // 管理员
)
