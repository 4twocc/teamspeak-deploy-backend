package user

import (
	"context"

	v1 "teamspeak-one-click-deploy/api/user/v1"
	"teamspeak-one-click-deploy/internal/dao"
	"teamspeak-one-click-deploy/internal/model/do"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	_, err = dao.User.Ctx(ctx).WherePri(req.Uid).Data(do.User{
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Introduction: req.Introduction,
		Status:       req.Status,
	}).Update()
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}
