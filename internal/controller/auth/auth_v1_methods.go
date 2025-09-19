package auth

// Auth 控制器 V1 版本方法实现。
// 负责登录、刷新与登出能力：
// - Login: 校验凭据并签发 Access/Refresh Token；
// - Refresh: 校验 RefreshToken 并签发新的 Access Token；支持按配置进行刷新令牌旋转；
// - Logout: 预留接口，当前不做状态持久化变更。

import (
	"context"
	"crypto/sha256"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"golang.org/x/crypto/bcrypt"

	v1 "teamspeak-one-click-deploy/api/auth/v1"
	"teamspeak-one-click-deploy/internal/consts"
	"teamspeak-one-click-deploy/internal/dao"
	"teamspeak-one-click-deploy/internal/model/entity"
	"teamspeak-one-click-deploy/utility"
)

// Login 用户登录：
// 参数: ctx 请求上下文；req 包含 identifier(用户名/邮箱/手机号)、password。
// 返回: 登录成功返回 Access/Refresh Token 与过期秒数；失败返回错误码。
// 可能错误: CodeInvalidParameter, CodeNotAuthorized, CodeInternalError。
func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	// 参数校验
	if req == nil || req.Account == "" || req.Password == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter)
	}

	// 1) 查找用户: 支持用户名/邮箱/手机号
	var u entity.Users
	if err = dao.Users.Ctx(ctx).
		Where("username = ? OR email = ? OR phone = ?", req.Account, req.Account, req.Account).
		Limit(1).
		Scan(&u); err != nil {
		g.Log().Errorf(ctx, "Login query user error: %v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	if u.Id == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}

	// 2) 读取密码哈希
	var ua entity.UsersAuth
	if err = dao.UsersAuth.Ctx(ctx).
		Where("user_id", u.Id).
		Limit(1).
		Scan(&ua); err != nil {
		g.Log().Errorf(ctx, "Login query users_auth error: %v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	if ua.UserId == 0 || ua.PasswordHash == "" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}

	// 3) 校验密码，仅支持 bcrypt
	if ua.PasswordAlgo != "bcrypt" {
		g.Log().Warningf(ctx, "Login unsupported password algo: %s, userId=%d", ua.PasswordAlgo, u.Id)
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(ua.PasswordHash), []byte(req.Password)); err != nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}

	// 4) 查询角色代码列表
	var roleIds []uint
	if rs, e := dao.UserRoles.Ctx(ctx).Where("user_id", u.Id).Fields("role_id").All(); e != nil {
		g.Log().Errorf(ctx, "Login query user_roles error: %v", e)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	} else {
		for _, r := range rs {
			roleIds = append(roleIds, gconv.Uint(r["role_id"]))
		}
	}

	var roles []string
	if len(roleIds) > 0 {
		if rr, e := dao.Roles.Ctx(ctx).WhereIn("id", roleIds).Fields("code").All(); e != nil {
			g.Log().Errorf(ctx, "Login query roles error: %v", e)
			return nil, gerror.NewCode(gcode.CodeInternalError)
		} else {
			for _, r := range rr {
				if code := gconv.String(r["code"]); code != "" {
					roles = append(roles, code)
				}
			}
		}
	}

	// 5) 签发 Access/Refresh Token
	pair, err := utility.IssueTokenPair(ctx, gconv.String(u.Id), u.Uid, roles)
	if err != nil {
		g.Log().Errorf(ctx, "Login issue token error: %v", err)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}

	// 6) 存储 Refresh Token 白名单（如启用）
	if pair.RefreshToken != "" {
		sum := sha256.Sum256([]byte(pair.RefreshToken))
		tokenHash := sum[:]

		// 解析 refresh 的过期时间用于存储
		rClaims, perr := utility.ParseToken(ctx, pair.RefreshToken)
		if perr != nil {
			g.Log().Warningf(ctx, "Login parse refresh for store error: %v", perr)
		} else {
			_, ierr := dao.VerificationTokens.Ctx(ctx).Data(g.Map{
				"user_id":    u.Id,
				"token_type": "refresh",
				"token_hash": tokenHash,
				"expires_at": rClaims.ExpiresAt.Time,
				"metadata":   gjson.New("login"),
				"created_at": gtime.Now(),
			}).Insert()
			if ierr != nil {
				g.Log().Warningf(ctx, "Login store refresh token error: %v", ierr)
			}
		}
	}

	// 7) 返回结果
	res = &v1.LoginRes{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}
	return res, nil
}

// Refresh 刷新访问令牌：
// 参数: ctx 请求上下文；req 包含 refreshToken。
// 返回: 新的 Access Token 与有效期秒数；若启用刷新旋转，也返回新的 Refresh Token。
// 可能错误: CodeInvalidParameter, CodeNotAuthorized, CodeInternalError。
func (c *ControllerV1) Refresh(ctx context.Context, req *v1.RefreshReq) (res *v1.RefreshRes, err error) {
	// 进入刷新流程的日志，避免打印令牌原文
	g.Log().Info(ctx, "auth.v1.Refresh: begin")

	if req == nil || req.RefreshToken == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter)
	}

	// 1) 解析并验证刷新令牌（必须有效未过期）
	claims, err := utility.ParseToken(ctx, req.RefreshToken)
	if err != nil {
		g.Log().Warningf(ctx, "Refresh parse token error: %v", err)
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}

	// 2) 校验白名单（存在、未过期、未标记消费）
	sum := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := sum[:]

	rec, qerr := dao.VerificationTokens.Ctx(ctx).
		Where("user_id", gconv.Uint64(claims.Subject)).
		Where("token_type", "refresh").
		Where("token_hash", tokenHash).
		Where("consumed_at IS NULL").
		Where("expires_at > ?", gtime.Now()).
		One()
	if qerr != nil {
		g.Log().Errorf(ctx, "Refresh query whitelist error: %v", qerr)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}
	if rec.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized)
	}

	// 3) 判断是否启用刷新旋转
	rotate := g.Cfg().MustGet(ctx, consts.ConfKeyJWTRotateRef, true).Bool()

	if !rotate {
		// 不旋转，仅签发新的 Access Token
		at, exp, ierr := utility.IssueAccessToken(ctx, claims.Subject, claims.UID, claims.Roles)
		if ierr != nil {
			g.Log().Errorf(ctx, "Refresh issue access token error: %v", ierr)
			return nil, gerror.NewCode(gcode.CodeInternalError)
		}
		return &v1.RefreshRes{AccessToken: at, ExpiresIn: exp - gtime.Now().Timestamp()}, nil
	}

	// 4) 旋转：
	// 4.1 将旧的 refresh 记录标记为已消费
	if _, uerr := dao.VerificationTokens.Ctx(ctx).
		Data(g.Map{"consumed_at": gtime.Now(), "metadata": gjson.New("rotated")}).
		Where("id", rec["id"]).
		Update(); uerr != nil {
		g.Log().Errorf(ctx, "Refresh consume old token error: %v", uerr)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}

	// 4.2 签发新的 Access/Refresh 对，并白名单新 refresh
	pair, perr := utility.IssueTokenPair(ctx, claims.Subject, claims.UID, claims.Roles)
	if perr != nil {
		g.Log().Errorf(ctx, "Refresh rotate issue pair error: %v", perr)
		return nil, gerror.NewCode(gcode.CodeInternalError)
	}

	if pair.RefreshToken != "" {
		nsum := sha256.Sum256([]byte(pair.RefreshToken))
		nHash := nsum[:]
		rClaims, rperr := utility.ParseToken(ctx, pair.RefreshToken)
		if rperr != nil {
			g.Log().Warningf(ctx, "Refresh parse new refresh for store error: %v", rperr)
		} else {
			if _, ierr := dao.VerificationTokens.Ctx(ctx).Data(g.Map{
				"user_id":    gconv.Uint64(claims.Subject),
				"token_type": "refresh",
				"token_hash": nHash,
				"expires_at": rClaims.ExpiresAt.Time,
				"metadata":   gjson.New("rotate"),
				"created_at": gtime.Now(),
			}).Insert(); ierr != nil {
				g.Log().Warningf(ctx, "Refresh store new refresh token error: %v", ierr)
			}
		}
	}

	// 4.3 返回新的 access 与 refresh
	return &v1.RefreshRes{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}

// Logout 登出：
// 当前接口未要求传入刷新令牌，暂不修改白名单状态，仅返回成功。
// 后续如需要支持刷新令牌失效，请在 DTO 中增加 RefreshToken 字段并在此处标记 consumed_at。
func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	return &v1.LogoutRes{}, nil
}
