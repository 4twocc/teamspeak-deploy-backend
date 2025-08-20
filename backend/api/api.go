package api

import (
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

// API路径常量定义
const (
	// 基础路径
	APIBasePath = "api"

	// 基础API路径
	PingPath = APIBasePath + "/ping"

	// 认证相关路径
	AuthBasePath = APIBasePath + "/auth"
	// 认证登录路径
	LoginPath = AuthBasePath + "/login"
	// 认证登出路径
	LogoutPath = AuthBasePath + "/logout"
	// 认证用户信息路径
	UserInfoPath = AuthBasePath + "/info"

	// 用户管理路径
	UsersBasePath = APIBasePath + "/users"
	// 用户列表路径
	UsersListPath = UsersBasePath + "/list"
	// 用户分页列表路径
	UsersPagePath = UsersBasePath + "/page"
	// 用户添加路径
	UsersAddPath = UsersBasePath + "/add"
	// 用户删除路径
	UsersRemovePath = UsersBasePath + "/remove"

	// 实例管理路径
	InstancesBasePath = APIBasePath + "/instances"
	// 实例列表路径
	InstancesListPath = InstancesBasePath
	// 实例创建路径
	InstancesCreatePath = InstancesBasePath
	// 实例详情路径
	InstancesDetailPath = InstancesBasePath + "/:id"
	// 实例更新路径
	InstancesUpdatePath = InstancesBasePath + "/:id"
	// 实例删除路径
	InstancesDeletePath = InstancesBasePath + "/:id"
	// 实例启动路径 - 修改为符合RESTful规范的action
	InstancesStartPath = InstancesBasePath + "/:id/actions/start"
	// 实例停止路径 - 修改为符合RESTful规范的action
	InstancesStopPath = InstancesBasePath + "/:id/actions/stop"
	// 实例重启路径 - 修改为符合RESTful规范的action
	InstancesRestartPath = InstancesBasePath + "/:id/actions/restart"
	// 实例日志路径
	InstancesLogsPath = InstancesBasePath + "/:id/logs"
	// 实例资源路径
	InstancesResourcesPath = InstancesBasePath + "/:id/resources"

	// 部署相关路径
	DeployBasePath = APIBasePath + "/v1/deploy"
	// 部署启动路径
	DeployStartPath = DeployBasePath + "/start"
	// 部署状态路径
	DeployStatusPath = DeployBasePath + "/status"
	// 部署重置路径
	DeployResetPath = DeployBasePath + "/reset"
	// 部署容器状态路径
	DeployContainerStatusPath = DeployBasePath + "/container-status"
	// 部署初始化环境路径
	DeployInitEnvPath = DeployBasePath + "/init-env"
	// 部署清理路径
	DeployCleanupPath = DeployBasePath + "/cleanup"

	// 监控相关路径 (注意：监控使用的是http.ServeMux，不是gin)
	MonitorBasePath = APIBasePath + "/monitor"
	// 系统监控路径
	MonitorSystemPath = MonitorBasePath + "/system"
	// 业务监控路径
	MonitorBusinessPath = MonitorBasePath + "/business"
	// 状态监控路径
	MonitorStatsPath = MonitorBasePath + "/stats"
	// 健康监控路径
	MonitorHealthPath = MonitorBasePath + "/health"
	// 指标监控路径
	MonitorMetricsPath = MonitorBasePath + "/metrics"
	// 历史监控路径
	MonitorHistoryPath = MonitorBasePath + "/history"
	// Redis健康检查路径
	RedisHealthPath = MonitorBasePath + "/redis/health"
)

// RegisterRoutes registers API routes
func RegisterRoutes(router *gin.Engine) {
	// 注册各模块路由
	router.GET(PingPath, pingHandler)
}

func pingHandler(c *gin.Context) {
	utils.OKGin(c, map[string]string{"message": "pong"})
}