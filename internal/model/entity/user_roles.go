// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserRoles is the golang structure for table user_roles.
type UserRoles struct {
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"ID=users.id"` // ID=users.id
	RoleId    uint        `json:"roleId"    orm:"role_id"    description:"ID=roles.id"` // ID=roles.id
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`            //
}
