package instance

import (
	"context"
	"fmt"
	"net/http"

	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/logs"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册实例管理路由
func RegisterRoutes(router *gin.Engine) {
	// 注册实例管理路由
	router.GET("/api/instances", listInstancesHandler)
	router.POST("/api/instances", createInstanceHandler)
	router.GET("/api/instances/:id", getInstanceHandler)
	router.PUT("/api/instances/:id", updateInstanceHandler)
	router.DELETE("/api/instances/:id", deleteInstanceHandler)
	router.POST("/api/instances/:id/actions/start", startInstanceHandler)
	router.POST("/api/instances/:id/actions/stop", stopInstanceHandler)
	router.POST("/api/instances/:id/actions/restart", restartInstanceHandler)
	router.GET("/api/instances/:id/logs", getInstanceLogsHandler)
	router.GET("/api/instances/:id/resources", getInstanceResourcesHandler)
}

var (
	instanceService *Service
	alertManager    *AlertManager
)

// SetLogService 设置全局实例服务的日志服务
func SetLogService(ls logs.LogService) {
	if instanceService != nil {
		instanceService.SetLogService(ls)
	}
}

// Initialize 初始化实例服务
func Initialize() error {
	// 初始化告警管理器
	alertManager = NewAlertManager(database.DB, nil)

	// 确保数据库表已创建
	if err := ensureTables(database.DB); err != nil {
		return fmt.Errorf("failed to ensure instance tables exist: %w", err)
	}

	// 初始化实例服务
	instanceService = NewService(database.DB, alertManager)

	return nil
}

// ensureTables 确保实例相关的表已创建
func ensureTables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 自动迁移实例相关表
	if err := db.AutoMigrate(&Instance{}, &InstanceLog{}); err != nil {
		return fmt.Errorf("failed to auto migrate instance tables: %w", err)
	}

	return nil
}

func listInstancesHandler(c *gin.Context) {
	// 简化实现 - 实际项目中应该使用分页和过滤
	ctx := context.Background()
	instances, _, err := instanceService.ListInstances(ctx, nil)
	if err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, err.Error())
		return
	}
	utils.OKGin(c, instances)
}

func createInstanceHandler(c *gin.Context) {
	var input CreateInstanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, utils.ErrorMessage(utils.ErrInvalidRequest))
		return
	}

	ctx := context.Background()
	instance, err := instanceService.CreateInstance(ctx, &input)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}

	utils.OKGin(c, instance)
}

func getInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	instance, err := instanceService.GetInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusNotFound, utils.ErrNotFound, utils.ErrorMessage(utils.ErrNotFound))
		return
	}
	utils.OKGin(c, instance)
}

func updateInstanceHandler(c *gin.Context) {
	id := c.Param("id")

	var input UpdateInstanceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, utils.ErrorMessage(utils.ErrInvalidRequest))
		return
	}

	ctx := context.Background()
	instance, err := instanceService.UpdateInstance(ctx, id, &input)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}

	utils.OKGin(c, instance)
}

func deleteInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	err := instanceService.DeleteInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func startInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	err := instanceService.StartInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}
	utils.OKGin(c, map[string]string{"message": "Instance start request submitted"})
}

func stopInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	err := instanceService.StopInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}
	utils.OKGin(c, map[string]string{"message": "Instance stop request submitted"})
}

func restartInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	err := instanceService.RestartInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}
	utils.OKGin(c, map[string]string{"message": "Instance restart request submitted"})
}

func getInstanceLogsHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	logs, err := instanceService.GetInstanceLogs(ctx, id, 100) // 获取最近100条日志
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, err.Error())
		return
	}
	utils.OKGin(c, logs)
}

func getInstanceResourcesHandler(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	// 获取实例
	instance, err := instanceService.GetInstance(ctx, id)
	if err != nil {
		utils.FailGin(c, http.StatusNotFound, utils.ErrNotFound, utils.ErrorMessage(utils.ErrNotFound))
		return
	}

	// 检查实例是否正在运行
	if !instance.IsRunning() {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRequest, utils.ErrorMessage(utils.ErrInvalidRequest))
		return
	}

	// 获取资源使用情况
	usage, err := getProcessResourceUsage(instance.ProcessID)
	if err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	utils.OKGin(c, usage)
}
