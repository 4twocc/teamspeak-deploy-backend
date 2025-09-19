// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UserProfiles is the golang structure of table user_profiles for DAO operations like Where/Data.
type UserProfiles struct {
	g.Meta      `orm:"table:user_profiles, do:true"`
	UserId      any         // ID=users.id
	FullName    any         //
	FirstName   any         //
	LastName    any         //
	AvatarUrl   any         //
	Locale      any         // /
	Timezone    any         //
	Gender      any         // 0,1,2,3
	DateOfBirth *gtime.Time //
	Bio         any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
