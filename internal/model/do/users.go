// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Users is the golang structure of table users for DAO operations like Where/Data.
type Users struct {
	g.Meta          `orm:"table:users, do:true"`
	Id              any         // ID
	Uid             any         // UIDUUIDv48.0.13-
	Username        any         //
	Email           any         //
	Phone           any         // E.164
	Status          any         // 123
	IsEmailVerified any         //
	IsPhoneVerified any         //
	CreatedAt       *gtime.Time //
	UpdatedAt       *gtime.Time //
	DeletedAt       *gtime.Time // NULL
}
