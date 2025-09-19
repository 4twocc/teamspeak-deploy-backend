package v1

import (
	"teamspeak-one-click-deploy/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// create user
type CreateReq struct {
	g.Meta   `path:"/user" method:"post" tags:"User" summary:"Create user"`
	Username string `v:"required|length:2,18" dc:"user username"`
	Email    string `v:"required|email" dc:"user email"`
	Phone    string `v:"required|regex:^\\+?[1-9][0-9]{1,14}$" dc:"user phone (E.164)"`
	Password string `v:"required|min:9|max:64" dc:"user password"`
}
type CreateRes struct {
	Uid string `json:"uid" dc:"user id"`
}

// update user
type UpdateReq struct {
	g.Meta `path:"/user/:uid" method:"put" tags:"User" summary:"Update user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type UpdateRes struct{}

// delete user
type DeleteReq struct {
	g.Meta `path:"/user/:uid" method:"delete" tags:"User" summary:"Delete user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type DeleteRes struct{}

// get user
type GetReq struct {
	g.Meta `path:"/user/:uid" method:"get" tags:"User" summary:"Get user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type GetRes struct {
	User *entity.Users `json:"user" dc:"user"`
}

// get user list
type GetListReq struct {
	g.Meta `path:"/user/list" method:"get" tags:"User" summary:"Get user list"`
	Page   int `in:"query" d:"1" v:"min:1" dc:"page number"`
	Size   int `in:"query" d:"10" v:"min:1|max:100" dc:"page size"`
}
type GetListRes struct {
	Users []*entity.Users `json:"users" dc:"user list"`
}
