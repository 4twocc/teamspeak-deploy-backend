// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UserProfiles is the golang structure for table user_profiles.
type UserProfiles struct {
	UserId      uint64      `json:"userId"      orm:"user_id"       description:"ID=users.id"` // ID=users.id
	FullName    string      `json:"fullName"    orm:"full_name"     description:""`            //
	FirstName   string      `json:"firstName"   orm:"first_name"    description:""`            //
	LastName    string      `json:"lastName"    orm:"last_name"     description:""`            //
	AvatarUrl   string      `json:"avatarUrl"   orm:"avatar_url"    description:""`            //
	Locale      string      `json:"locale"      orm:"locale"        description:"/"`           // /
	Timezone    string      `json:"timezone"    orm:"timezone"      description:""`            //
	Gender      uint        `json:"gender"      orm:"gender"        description:"0,1,2,3"`     // 0,1,2,3
	DateOfBirth *gtime.Time `json:"dateOfBirth" orm:"date_of_birth" description:""`            //
	Bio         string      `json:"bio"         orm:"bio"           description:""`            //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"    description:""`            //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"    description:""`            //
}
