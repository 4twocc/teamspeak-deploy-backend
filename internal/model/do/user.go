// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure of table user for DAO operations like Where/Data.
type User struct {
	g.Meta       `orm:"table:user, do:true"`
	Uid          any         // userid
	Password     any         // password
	Email        any         // email
	Username     any         // username
	Nickname     any         // nickname
	Avatar       any         // avatar
	Introduction any         // user introduction
	Role         any         // user role
	Status       any         // user status
	CreatedAt    *gtime.Time // user create_at
	LastOnlineAt *gtime.Time // user last_online_at
}
