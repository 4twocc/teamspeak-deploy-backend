// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// OauthAccounts is the golang structure of table oauth_accounts for DAO operations like Where/Data.
type OauthAccounts struct {
	g.Meta            `orm:"table:oauth_accounts, do:true"`
	Id                any         //
	UserId            any         // ID=users.id
	Provider          any         // github/google/wechat
	ProviderAccountId any         // ID
	AccessToken       any         //
	RefreshToken      any         //
	TokenType         any         //
	ExpiresAt         *gtime.Time //
	Scope             any         // Scope
	ProfileJson       any         //
	CreatedAt         *gtime.Time //
	UpdatedAt         *gtime.Time //
}
