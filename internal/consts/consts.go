package consts

// 定义全局常量与上下文键名，供全局复用。

// Context Keys（用于 ghttp.Request Ctx 中携带身份信息）
const (
	CtxKeyUserID   = "auth.user_id" // string, 来自 claims.sub（users.id）
	CtxKeyUserUID  = "auth.uid"     // string, 来自 claims.uid（users.uid）
	CtxKeyUserRole = "auth.roles"   // []string, 来自 claims.roles
)

// JWT 配置键名（从 env 或配置文件读取）
const (
	ConfKeyJWTSecret     = "jwt.secret"
	ConfKeyJWTIssuer     = "jwt.issuer"
	ConfKeyJWTAccessTTL  = "jwt.access_ttl"  // 例如: 15m
	ConfKeyJWTRefreshTTL = "jwt.refresh_ttl" // 例如: 168h
	ConfKeyJWTEnableRef  = "jwt.enable_refresh"
	ConfKeyJWTRotateRef  = "jwt.refresh_rotate"
	ConfKeyJWTClockSkew  = "jwt.clock_skew" // 例如: 2s
)
