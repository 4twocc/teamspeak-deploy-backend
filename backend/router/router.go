package router

import (
	"net/http"

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
)

// RegisterRoutes 统一注册所有模块的路由
func RegisterRoutes(router *gin.Engine) {
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

	// ---------------- Swagger 文档（仅非生产环境启用）----------------
	// 为避免在生产环境暴露 API 文档，仅当 gin.Mode() 不是 ReleaseMode 时才挂载。
	if gin.Mode() != gin.ReleaseMode {
		// 可选：根据实际路由前缀设置 BasePath，默认即可
		docs.SwaggerInfo.BasePath = "/"
		// 访问地址示例：/docs/index.html
		router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
