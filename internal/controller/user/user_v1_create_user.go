package user

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "teamspeak-one-click-deploy/api/user/v1"
)

func (c *ControllerV1) CreateUser(ctx context.Context, req *v1.CreateUserReq) (res *v1.CreateUserRes, err error) {

	return nil, gerror.NewCode(gcode.CodeNotImplemented)
}
