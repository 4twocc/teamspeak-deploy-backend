// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// OauthAccounts is the golang structure for table oauth_accounts.
type OauthAccounts struct {
	Id                uint64      `json:"id"                orm:"id"                  description:""`                     //
	UserId            uint64      `json:"userId"            orm:"user_id"             description:"ID=users.id"`          // ID=users.id
	Provider          string      `json:"provider"          orm:"provider"            description:"github/google/wechat"` // github/google/wechat
	ProviderAccountId string      `json:"providerAccountId" orm:"provider_account_id" description:"ID"`                   // ID
	AccessToken       string      `json:"accessToken"       orm:"access_token"        description:""`                     //
	RefreshToken      string      `json:"refreshToken"      orm:"refresh_token"       description:""`                     //
	TokenType         string      `json:"tokenType"         orm:"token_type"          description:""`                     //
	ExpiresAt         *gtime.Time `json:"expiresAt"         orm:"expires_at"          description:""`                     //
	Scope             string      `json:"scope"             orm:"scope"               description:"Scope"`                // Scope
	ProfileJson       string      `json:"profileJson"       orm:"profile_json"        description:""`                     //
	CreatedAt         *gtime.Time `json:"createdAt"         orm:"created_at"          description:""`                     //
	UpdatedAt         *gtime.Time `json:"updatedAt"         orm:"updated_at"          description:""`                     //
}
