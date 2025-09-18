package user

import (
	"context"

	v1 "teamspeak-one-click-deploy/api/user/v1"
	"teamspeak-one-click-deploy/internal/dao"
)

func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	res = &v1.GetOneRes{}
	err = dao.User.Ctx(ctx).WherePri(req.Uid).Scan(&res.User)
	return
}
