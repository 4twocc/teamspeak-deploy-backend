package v1

import (
	"teamspeak-one-click-deploy/internal/model/entity"
	"teamspeak-one-click-deploy/utility"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// create user
type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"User" summary:"Create user"`
	Account  string `v:"required|length:6,18" dc:"user account"`
	Password string `v:"required|length:11,18" dc:"user password"`
}
type CreateRes struct {
	Uid int64 `json:"uid" dc:"user id"`
}

// delete user
type DeleteReq struct {
	g.Meta `path:"/user/{uid}" method:"delete" tags:"User" summary:"Delete user"`
	Uid    int64 `v:"required" dc:"user id"`
}
type DeleteRes struct{}

// update user
type UpdateReq struct {
	g.Meta       `path:"/user/{uid}" method:"put" tags:"User" summary:"Update user"`
	Uid          int64           `v:"required" dc:"user id"`
	Nickname     *string         `v:"length:3,10" dc:"user nickname"`
	Avatar       *string         `v:"length:3,10" dc:"user avatar"`
	Introduction *string         `v:"length:30,100" dc:"user introduction"`
	Status       *utility.Status `v:"in:0,1,2,3" dc:"user status"`
}
type UpdateRes struct{}

// get one user
type GetOneReq struct {
	g.Meta `path:"/user/{uid}" method:"get" tags:"User" summary:"Get one user"`
	Uid    int64 `v:"required" dc:"user id"`
}
type GetOneRes struct {
	*entity.User `dc:"user"`
}

// get user list
type GetListReq struct {
	g.Meta       `path:"/user" method:"get" tags:"User" summary:"Get users"`
	Uid          *int64          `v:"required" dc:"user id"`
	Nickname     *string         `v:"length:3,10" dc:"user nickname"`
	Status       *utility.Status `v:"in:0,1,2,3" dc:"user status"`
	CreatedAt    *gtime.Time     `v:"required" dc:"user created at"`
	LastOnlineAt *gtime.Time     `v:"required" dc:"user last online at"`
}
type GetListRes struct {
	List []*entity.User `json:"list" dc:"user list"`
}
