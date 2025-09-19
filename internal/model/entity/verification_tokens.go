// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// VerificationTokens is the golang structure for table verification_tokens.
type VerificationTokens struct {
	Id         uint64      `json:"id"         orm:"id"          description:""`            //
	UserId     uint64      `json:"userId"     orm:"user_id"     description:"ID=users.id"` // ID=users.id
	TokenType  string      `json:"tokenType"  orm:"token_type"  description:""`            //
	TokenHash  []byte      `json:"tokenHash"  orm:"token_hash"  description:"SHA-256"`     // SHA-256
	ExpiresAt  *gtime.Time `json:"expiresAt"  orm:"expires_at"  description:""`            //
	ConsumedAt *gtime.Time `json:"consumedAt" orm:"consumed_at" description:"NULL"`        // NULL
	Metadata   string      `json:"metadata"   orm:"metadata"    description:""`            //
	CreatedAt  *gtime.Time `json:"createdAt"  orm:"created_at"  description:""`            //
}
