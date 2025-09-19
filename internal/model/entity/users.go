// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure for table users.
type Users struct {
	Id              uint64      `json:"id"              orm:"id"                description:"ID"`               // ID
	Uid             string      `json:"uid"             orm:"uid"               description:"UIDUUIDv48.0.13-"` // UIDUUIDv48.0.13-
	Username        string      `json:"username"        orm:"username"          description:""`                 //
	Email           string      `json:"email"           orm:"email"             description:""`                 //
	Phone           string      `json:"phone"           orm:"phone"             description:"E.164"`            // E.164
	Status          uint        `json:"status"          orm:"status"            description:"123"`              // 123
	IsEmailVerified int         `json:"isEmailVerified" orm:"is_email_verified" description:""`                 //
	IsPhoneVerified int         `json:"isPhoneVerified" orm:"is_phone_verified" description:""`                 //
	CreatedAt       *gtime.Time `json:"createdAt"       orm:"created_at"        description:""`                 //
	UpdatedAt       *gtime.Time `json:"updatedAt"       orm:"updated_at"        description:""`                 //
	DeletedAt       *gtime.Time `json:"deletedAt"       orm:"deleted_at"        description:"NULL"`             // NULL
}
