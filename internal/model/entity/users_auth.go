// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// UsersAuth is the golang structure for table users_auth.
type UsersAuth struct {
	UserId                uint64      `json:"userId"                orm:"user_id"                  description:"ID=users.id"`     // ID=users.id
	PasswordHash          string      `json:"passwordHash"          orm:"password_hash"            description:"argon2id/bcrypt"` // argon2id/bcrypt
	PasswordAlgo          string      `json:"passwordAlgo"          orm:"password_algo"            description:""`                //
	MfaEnabled            int         `json:"mfaEnabled"            orm:"mfa_enabled"              description:"MFA"`             // MFA
	MfaSecret             string      `json:"mfaSecret"             orm:"mfa_secret"               description:"TOTPKMS"`         // TOTPKMS
	LastPasswordChangedAt *gtime.Time `json:"lastPasswordChangedAt" orm:"last_password_changed_at" description:""`                //
	CreatedAt             *gtime.Time `json:"createdAt"             orm:"created_at"               description:""`                //
	UpdatedAt             *gtime.Time `json:"updatedAt"             orm:"updated_at"               description:""`                //
}
