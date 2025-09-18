package user

import (
	"context"

	v1 "teamspeak-one-click-deploy/api/user/v1"
	"teamspeak-one-click-deploy/internal/dao"
	"teamspeak-one-click-deploy/internal/model/do"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	res = &v1.GetListRes{}
	err = dao.User.Ctx(ctx).Where(do.User{
		Uid:          req.Uid,
		Nickname:     req.Nickname,
		Status:       req.Status,
		CreatedAt:    req.CreatedAt,
		LastOnlineAt: req.LastOnlineAt,
	}).Scan(&res.List)
	return
}
