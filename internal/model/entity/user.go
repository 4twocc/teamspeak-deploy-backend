// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Uid          int64       `json:"uid"          orm:"uid"            description:"userid"`              // userid
	Password     string      `json:"password"     orm:"password"       description:"password"`            // password
	Email        string      `json:"email"        orm:"email"          description:"email"`               // email
	Username     string      `json:"username"     orm:"username"       description:"username"`            // username
	Nickname     string      `json:"nickname"     orm:"nickname"       description:"nickname"`            // nickname
	Avatar       string      `json:"avatar"       orm:"avatar"         description:"avatar"`              // avatar
	Introduction string      `json:"introduction" orm:"introduction"   description:"user introduction"`   // user introduction
	Role         string      `json:"role"         orm:"role"           description:"user role"`           // user role
	Status       int         `json:"status"       orm:"status"         description:"user status"`         // user status
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"     description:"user create_at"`      // user create_at
	LastOnlineAt *gtime.Time `json:"lastOnlineAt" orm:"last_online_at" description:"user last_online_at"` // user last_online_at
}
