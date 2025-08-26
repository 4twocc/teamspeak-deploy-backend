/**
 * 日志模块路由注册
 * 作者: AI Assistant
 * 版本: 1.0.0
 * 功能: 注册日志管理相关的API路由
 */

package logs

import (
	"log"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/database"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册日志管理路由
// 参数:
//   - router: Gin路由引擎实例
//
// 功能: 注册所有日志相关的API端点
func RegisterRoutes(router *gin.Engine) {
	// 创建日志服务实例
	logService, err := NewLogService(database.DB, LogConfig{
		Level:         "info",
		EnableDB:      true,
		RetentionDays: 90,
		BatchSize:     100,
		BatchInterval: 5,
		EnableFile:    true,
		FilePath:      "logs/app.log",
	})
	if err != nil {
		log.Printf("Failed to create log service: %v", err)
		return
	}

	// 创建日志处理器实例
	logHandler := NewLogHandler(logService)

	// 注册日志查询路由 (GET /api/logs)
	router.GET(api.LogsQueryPath, logHandler.GetLogs)

	// 注册日志统计路由 (GET /api/logs/stats)
	router.GET(api.LogsStatsPath, logHandler.GetLogStats)

	// 注册日志清理路由 (DELETE /api/logs/cleanup)
	router.DELETE(api.LogsCleanupPath, logHandler.CleanupLogs)
}
