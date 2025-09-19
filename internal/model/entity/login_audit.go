// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// LoginAudit is the golang structure for table login_audit.
type LoginAudit struct {
	Id        uint64      `json:"id"        orm:"id"         description:""`            //
	UserId    uint64      `json:"userId"    orm:"user_id"    description:"ID"`          // ID
	Identity  string      `json:"identity"  orm:"identity"   description:"//"`          // //
	Provider  string      `json:"provider"  orm:"provider"   description:""`            //
	IpAddress []byte      `json:"ipAddress" orm:"ip_address" description:"IPIPv4/IPv6"` // IPIPv4/IPv6
	UserAgent string      `json:"userAgent" orm:"user_agent" description:"UA"`          // UA
	Success   int         `json:"success"   orm:"success"    description:""`            //
	ErrorCode string      `json:"errorCode" orm:"error_code" description:""`            //
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:""`            //
}
