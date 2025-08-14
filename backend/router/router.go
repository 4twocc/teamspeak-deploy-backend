package router

import (
	"net/http"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/auth"
	"teamspeak-one-click-deploy/deploy"
	"teamspeak-one-click-deploy/instance"
	"teamspeak-one-click-deploy/monitor"
	"teamspeak-one-click-deploy/users"

	"github.com/gin-gonic/gin"
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
	users.RegisterRoutes(router)

	// 监控路由需要 http.ServeMux，所以我们创建一个并挂载到 Gin
	monitorMux := http.NewServeMux()
	monitor.RegisterRoutes(monitorMux)
	// 将监控路由挂载到 Gin
	router.Any("/api/monitor/*any", gin.WrapH(monitorMux))
}
