// 文件说明：路由集中注册与 API 文档（Swagger）按配置开关控制
// 作者：Trae AI
// 版本：v1.0
// 作用：统一注册各模块路由；根据运行环境与配置决定是否暴露 Swagger 文档，并支持基于 IP/CIDR 的访问白名单。

package router

import (
    "net/http"
    "log"
    "net"
    "strings"

    "teamspeak-one-click-deploy/api"
    "teamspeak-one-click-deploy/auth"
    "teamspeak-one-click-deploy/deploy"
    "teamspeak-one-click-deploy/instance"
    "teamspeak-one-click-deploy/logs"
    "teamspeak-one-click-deploy/monitor"
    "teamspeak-one-click-deploy/user"

    "github.com/gin-gonic/gin"

    // Swagger 相关依赖，仅在非生产环境启用
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    "teamspeak-one-click-deploy/docs"
    "teamspeak-one-click-deploy/config"
)

// RegisterRoutes 统一注册所有模块的路由（兼容入口）
// param router *gin.Engine 路由引擎
// return void
// throws 无
// author Trae AI
func RegisterRoutes(router *gin.Engine) {
    // 兼容旧调用，默认不携带配置；文档开关将根据 gin.Mode() 的默认策略
    RegisterRoutesWithConfig(router, nil)
}

// RegisterRoutesWithConfig 统一注册所有模块的路由，并按配置控制 Swagger 文档暴露
// param router *gin.Engine 路由引擎
// param cfg *config.Config 全量配置；当为 nil 时退化为仅按 gin.Mode() 控制
// return void
// throws 无
// author Trae AI
func RegisterRoutesWithConfig(router *gin.Engine, cfg *config.Config) {
    // 注册API基础路由
    api.RegisterRoutes(router)

    // 注册认证路由
    auth.RegisterRoutes(router)

    // 注册实例管理路由
    instanceHandler := instance.NewHandler()
    instanceHandler.RegisterRoutes(router)

    // 注册部署路由
    deploy.RegisterRoutes(router)

    // 注册用户路由
    user.RegisterRoutes(router)

    // 注册日志路由
    logs.RegisterRoutes(router)

    // 监控路由需要 http.ServeMux，所以我们创建一个并挂载到 Gin
    monitorMux := http.NewServeMux()
    monitor.RegisterRoutes(monitorMux)
    // 将监控路由挂载到 Gin
    router.Any("/api/monitor/*any", gin.WrapH(monitorMux))

    // ---------------- Swagger 文档开关与白名单 ----------------
    // 规则：
    // - 非生产环境（!= ReleaseMode）：默认开启；若 cfg!=nil 则以 cfg.Server.Docs.Enabled 为准
    // - 生产环境（ReleaseMode）：默认关闭；仅当 cfg!=nil 且 cfg.Server.Docs.Enabled=true 时开启
    docsEnabled := gin.Mode() != gin.ReleaseMode
    if gin.Mode() == gin.ReleaseMode {
        // 生产环境默认关闭
        docsEnabled = false
        if cfg != nil {
            docsEnabled = cfg.Server.Docs.Enabled
        }
    } else {
        // 非生产环境默认开启，可通过配置显式关闭
        if cfg != nil {
            docsEnabled = cfg.Server.Docs.Enabled
        }
    }

    if !docsEnabled {
        return
    }

    // 启用文档：设置基础路径
    docs.SwaggerInfo.BasePath = "/"

    // 若存在白名单，启用白名单中间件
    var whitelist []string
    if cfg != nil {
        whitelist = cfg.Server.Docs.Whitelist
    }

    if len(whitelist) > 0 {
        // 使用分组并挂载中间件
        group := router.Group("/docs").Use(docsWhitelistMiddleware(whitelist))
        group.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    } else {
        // 无白名单限制
        router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }
}

// docsWhitelistMiddleware Swagger 文档访问白名单中间件
// 支持精确 IP 和 CIDR 段（形如 192.168.1.0/24）
// param whitelist []string 允许访问的 IP 或网段列表
// return gin.HandlerFunc Gin 中间件处理函数
// throws 无
// author Trae AI
func docsWhitelistMiddleware(whitelist []string) gin.HandlerFunc {
    // 预解析白名单，提升匹配性能
    var (
        ipList  []net.IP
        cidrNets []*net.IPNet
    )

    for _, item := range whitelist {
        s := strings.TrimSpace(item)
        if s == "" {
            continue
        }
        if strings.Contains(s, "/") {
            if _, ipnet, err := net.ParseCIDR(s); err == nil {
                cidrNets = append(cidrNets, ipnet)
            } else {
                log.Printf("[docsWhitelist] 无效CIDR：%s，错误：%v", s, err)
            }
        } else {
            if ip := net.ParseIP(s); ip != nil {
                ipList = append(ipList, ip)
            } else {
                log.Printf("[docsWhitelist] 无效IP：%s", s)
            }
        }
    }

    return func(c *gin.Context) {
        clientIPStr := c.ClientIP()
        if clientIPStr == "" {
            log.Printf("[docsWhitelist] 获取客户端IP失败，拒绝访问")
            c.AbortWithStatusJSON(403, gin.H{"code": 403, "message": "forbidden"})
            return
        }

        clientIP := net.ParseIP(clientIPStr)
        if clientIP == nil {
            log.Printf("[docsWhitelist] 解析客户端IP失败：%s，拒绝访问", clientIPStr)
            c.AbortWithStatusJSON(403, gin.H{"code": 403, "message": "forbidden"})
            return
        }

        // 精确 IP 匹配
        for _, allowed := range ipList {
            if allowed.Equal(clientIP) {
                c.Next()
                return
            }
        }

        // CIDR 网段匹配
        for _, netw := range cidrNets {
            if netw.Contains(clientIP) {
                c.Next()
                return
            }
        }

        // 均未命中，拒绝访问
        log.Printf("[docsWhitelist] 未命中白名单，客户端IP：%s，拒绝访问", clientIPStr)
        c.AbortWithStatusJSON(403, gin.H{"code": 403, "message": "forbidden"})
    }
}
