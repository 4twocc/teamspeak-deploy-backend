package utility

// JWT 工具封装，提供签发与解析 Token 的能力。

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/grand"
	jwt "github.com/golang-jwt/jwt/v5"

	"teamspeak-one-click-deploy/internal/consts"
)

// TokenPair 表示一对 Access/Refresh Token。
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn"` // Access Token 剩余秒数
}

// Claims 自定义声明，包含用户ID、UID与角色列表。
// 注意：sub 用于承载 users.id 的字符串形式；uid 为外显UID；roles 为角色代码列表。
type Claims struct {
	UID   string   `json:"uid"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

// getConfig 获取 JWT 相关配置。
// 返回: secret, issuer, accessTTL, refreshTTL, enableRefresh, rotateRefresh, clockSkew
func getConfig(ctx context.Context) (string, string, time.Duration, time.Duration, bool, bool, time.Duration) {
	cfg := g.Cfg()
	secret := cfg.MustGet(ctx, consts.ConfKeyJWTSecret, "change_me").String()
	issuer := cfg.MustGet(ctx, consts.ConfKeyJWTIssuer, "teamspeak-one-click-deploy").String()
	accessTTL := cfg.MustGet(ctx, consts.ConfKeyJWTAccessTTL, "15m").Duration()
	refreshTTL := cfg.MustGet(ctx, consts.ConfKeyJWTRefreshTTL, "168h").Duration()
	enableRefresh := cfg.MustGet(ctx, consts.ConfKeyJWTEnableRef, true).Bool()
	rotateRefresh := cfg.MustGet(ctx, consts.ConfKeyJWTRotateRef, true).Bool()
	clockSkew := cfg.MustGet(ctx, consts.ConfKeyJWTClockSkew, "2s").Duration()
	return secret, issuer, accessTTL, refreshTTL, enableRefresh, rotateRefresh, clockSkew
}

// IssueAccessToken 签发 Access Token。
// 参数: ctx 上下文, sub 用户ID（users.id，字符串）, uid 外显UID（users.uid）与 roles 角色代码列表。
// 返回: 签发的 token 字符串 与 过期时间戳。
func IssueAccessToken(ctx context.Context, sub string, uid string, roles []string) (string, int64, error) {
	secret, issuer, accessTTL, _, _, _, _ := getConfig(ctx)
	now := time.Now()
	exp := now.Add(accessTTL)

	claims := &Claims{
		UID:   uid,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        grand.Letters(20),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		g.Log().Errorf(ctx, "IssueAccessToken sign error: %v", err)
		return "", 0, err
	}
	return s, exp.Unix(), nil
}

// IssueTokenPair 签发 Access/Refresh Token 对。
// 参数: 同 IssueAccessToken。
// 返回: TokenPair，包含 AccessToken、RefreshToken（可选）、ExpiresIn。
func IssueTokenPair(ctx context.Context, sub string, uid string, roles []string) (*TokenPair, error) {
	secret, issuer, _, refreshTTL, enableRefresh, _, _ := getConfig(ctx)
	access, exp, err := IssueAccessToken(ctx, sub, uid, roles)
	if err != nil {
		return nil, err
	}

	pair := &TokenPair{AccessToken: access, ExpiresIn: exp - time.Now().Unix()}

	if !enableRefresh {
		return pair, nil
	}

	now := time.Now()
	rclaims := &Claims{
		UID:   uid,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
			ID:        "r-" + grand.Letters(20),
		},
	}

	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, rclaims)
	rs, err := rtoken.SignedString([]byte(secret))
	if err != nil {
		g.Log().Errorf(ctx, "IssueTokenPair sign refresh error: %v", err)
		return nil, err
	}

	pair.RefreshToken = rs
	return pair, nil
}

// ParseToken 解析并验证 Token，返回自定义 Claims。
// 参数: tokenStr 令牌字符串，allowExpired 表示是否允许过期（用于刷新流程提前读取）。
// 返回: *Claims。
func ParseToken(ctx context.Context, tokenStr string) (*Claims, error) {
	secret, issuer, _, _, _, _, clockSkew := getConfig(ctx)
	parser := jwt.NewParser(
		jwt.WithIssuer(issuer),
		jwt.WithLeeway(clockSkew),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		g.Log().Warningf(ctx, "ParseToken error: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// ParseTokenAllowExpired 解析 Token，若仅因过期报错则返回 claims（用于刷新流程）。
func ParseTokenAllowExpired(ctx context.Context, tokenStr string) (*Claims, error) {
	secret, issuer, _, _, _, _, clockSkew := getConfig(ctx)
	parser := jwt.NewParser(
		jwt.WithIssuer(issuer),
		jwt.WithLeeway(clockSkew),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)

	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		// 仅因过期导致的错误，直接返回 claims 以便刷新
		if token != nil && errors.Is(err, jwt.ErrTokenExpired) {
			if claims, ok := token.Claims.(*Claims); ok {
				return claims, nil
			}
		}
		g.Log().Warningf(ctx, "ParseTokenAllowExpired error: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// HashToken 计算 Token 的 SHA-256 哈希，用于存入数据库白名单。
func HashToken(_ context.Context, token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CtxWithAuth 将 claims 写入上下文，便于控制器读取。
func CtxWithAuth(ctx context.Context, sub, uid string, roles []string) context.Context {
	nctx := gctx.New()
	nctx = context.WithValue(nctx, consts.CtxKeyUserID, sub)
	nctx = context.WithValue(nctx, consts.CtxKeyUserUID, uid)
	nctx = context.WithValue(nctx, consts.CtxKeyUserRole, roles)
	return nctx
}
