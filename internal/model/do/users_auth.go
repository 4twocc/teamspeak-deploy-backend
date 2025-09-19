// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// UsersAuth is the golang structure of table users_auth for DAO operations like Where/Data.
type UsersAuth struct {
	g.Meta                `orm:"table:users_auth, do:true"`
	UserId                any         // ID=users.id
	PasswordHash          any         // argon2id/bcrypt
	PasswordAlgo          any         //
	MfaEnabled            any         // MFA
	MfaSecret             any         // TOTPKMS
	LastPasswordChangedAt *gtime.Time //
	CreatedAt             *gtime.Time //
	UpdatedAt             *gtime.Time //
}
