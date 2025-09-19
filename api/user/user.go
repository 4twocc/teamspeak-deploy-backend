// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package user

import (
	"context"

	"teamspeak-one-click-deploy/api/user/v1"
)

type IUserV1 interface {
	CreateUser(ctx context.Context, req *v1.CreateUserReq) (res *v1.CreateUserRes, err error)
	UpdateUser(ctx context.Context, req *v1.UpdateUserReq) (res *v1.UpdateUserRes, err error)
	DeleteUser(ctx context.Context, req *v1.DeleteUserReq) (res *v1.DeleteUserRes, err error)
	GetUser(ctx context.Context, req *v1.GetUserReq) (res *v1.GetUserRes, err error)
	GetUserList(ctx context.Context, req *v1.GetUserListReq) (res *v1.GetUserListRes, err error)
}
