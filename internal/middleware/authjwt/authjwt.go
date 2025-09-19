package authjwt

// JWT 鉴权中间件，负责从 Authorization 头解析 Access Token 并校验，随后将用户身份信息注入到请求上下文。

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"teamspeak-one-click-deploy/internal/consts"
	"teamspeak-one-click-deploy/utility"
)

// Middleware 是一个 ghttp 中间件函数。
// 参数: r - 当前请求对象。
// 行为:
//   - 解析 "Authorization: Bearer <token>" 头；
//   - 验证 Access Token 的签名与标准声明；
//   - 将 sub/uid/roles 注入到请求上下文，供控制器读取；
//   - 校验失败时，返回 401 并中止后续处理。
func Middleware(r *ghttp.Request) {
	auth := r.Request.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		r.Response.WriteStatusExit(http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimSpace(auth[len("Bearer "):])

	claims, err := utility.ParseToken(r.Context(), tokenStr)
	if err != nil {
		g.Log().Warningf(r.Context(), "JWT parse failed: %v", err)
		r.Response.WriteStatusExit(http.StatusUnauthorized)
		return
	}

	// 将身份信息写入 Context
	ctx := r.Context()
	ctx = context.WithValue(ctx, consts.CtxKeyUserID, claims.Subject)
	ctx = context.WithValue(ctx, consts.CtxKeyUserUID, claims.UID)
	ctx = context.WithValue(ctx, consts.CtxKeyUserRole, claims.Roles)
	r.SetCtx(ctx)

	r.Middleware.Next()
}
