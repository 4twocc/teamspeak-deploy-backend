package v1

import (
	"teamspeak-one-click-deploy/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// create user
type CreateUserReq struct {
	g.Meta   `path:"/user" method:"post" tags:"User" summary:"Create user"`
	Email    string `v:"required|email" dc:"user email"`
	Phone    string `v:"required|regex:^\\+?[1-9][0-9]{1,14}$" dc:"user phone (E.164)"`
	Password string `v:"required|min:9|max:64" dc:"user password"`
}
type CreateUserRes struct {
	Uid string `json:"uid" dc:"user id"`
}

// update user
type UpdateUserReq struct {
	g.Meta `path:"/user/:uid" method:"put" tags:"User" summary:"Update user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type UpdateUserRes struct{}

// delete user
type DeleteUserReq struct {
	g.Meta `path:"/user/:uid" method:"delete" tags:"User" summary:"Delete user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type DeleteUserRes struct{}

// get user
type GetUserReq struct {
	g.Meta `path:"/user/:uid" method:"get" tags:"User" summary:"Get user"`
	Uid    string `in:"path" v:"required" dc:"user id"`
}
type GetUserRes struct {
	User *entity.Users `json:"user" dc:"user"`
}

// get user list
type GetUserListReq struct {
	g.Meta `path:"/user/list" method:"get" tags:"User" summary:"Get user list"`
	Page   int `in:"query" d:"1" v:"min:1" dc:"page number"`
	Size   int `in:"query" d:"10" v:"min:1|max:100" dc:"page size"`
}
type GetUserListRes struct {
	Users []*entity.Users `json:"users" dc:"user list"`
}
