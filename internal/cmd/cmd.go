package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"teamspeak-one-click-deploy/internal/controller/auth"
	"teamspeak-one-click-deploy/internal/controller/user"
	"teamspeak-one-click-deploy/internal/middleware/authjwt"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)

				// 开放路由（如 /auth/*）
				group.Group("/auth", func(open *ghttp.RouterGroup) {
					open.Middleware(ghttp.MiddlewareHandlerResponse)
					open.Bind(
						auth.NewV1(),
					)
				})

				// 受保护路由，挂载 JWT 中间件
				group.Group("/", func(priv *ghttp.RouterGroup) {
					priv.Middleware(authjwt.Middleware)
					priv.Bind(
						user.NewV1(),
					)
				})
			})
			s.Run()
			return nil
		},
	}
)
