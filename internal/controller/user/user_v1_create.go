package user

import (
	"context"

	v1 "teamspeak-one-click-deploy/api/user/v1"
	"teamspeak-one-click-deploy/internal/dao"
	"teamspeak-one-click-deploy/internal/model/do"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	insertId, err := dao.User.Ctx(ctx).Data(do.User{
		Username: req.Account,
		Password: req.Password,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{
		Uid: insertId,
	}, nil
}
