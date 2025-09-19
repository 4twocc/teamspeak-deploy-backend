// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// VerificationTokens is the golang structure of table verification_tokens for DAO operations like Where/Data.
type VerificationTokens struct {
	g.Meta     `orm:"table:verification_tokens, do:true"`
	Id         any         //
	UserId     any         // ID=users.id
	TokenType  any         //
	TokenHash  []byte      // SHA-256
	ExpiresAt  *gtime.Time //
	ConsumedAt *gtime.Time // NULL
	Metadata   any         //
	CreatedAt  *gtime.Time //
}
