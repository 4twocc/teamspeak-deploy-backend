// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Roles is the golang structure for table roles.
type Roles struct {
	Id          uint        `json:"id"          orm:"id"          description:"ID"`         // ID
	Code        string      `json:"code"        orm:"code"        description:"admin/user"` // admin/user
	Name        string      `json:"name"        orm:"name"        description:""`           //
	Description string      `json:"description" orm:"description" description:""`           //
	IsSystem    int         `json:"isSystem"    orm:"is_system"   description:""`           //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"  description:""`           //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"  description:""`           //
}
