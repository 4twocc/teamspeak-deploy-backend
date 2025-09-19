// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// LoginAudit is the golang structure of table login_audit for DAO operations like Where/Data.
type LoginAudit struct {
	g.Meta    `orm:"table:login_audit, do:true"`
	Id        any         //
	UserId    any         // ID
	Identity  any         // //
	Provider  any         //
	IpAddress []byte      // IPIPv4/IPv6
	UserAgent any         // UA
	Success   any         //
	ErrorCode any         //
	CreatedAt *gtime.Time //
}
