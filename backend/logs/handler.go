// Package logs 提供系统日志管理功能
// 包含日志查询、统计和清理等API接口
// Author: TeamSpeak Deploy System
// Version: 1.0.0
package logs

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LogHandler 日志处理器结构体
// 提供日志相关的HTTP API接口
type LogHandler struct {
	logService LogService
}

// NewLogHandler 创建新的日志处理器实例
// 参数:
//   - logService: 日志服务接口实例
//
// 返回值:
//   - *LogHandler: 日志处理器实例
func NewLogHandler(logService LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// GetLogs 查询系统日志
// GET /api/logs
// 查询参数:
//   - level: 日志级别过滤
//   - module: 模块名过滤
//   - uid: 用户ID过滤
//   - start_time: 开始时间
//   - end_time: 结束时间
//   - page: 页码
//   - page_size: 每页大小
//
// 返回值: 日志列表和分页信息
func (h *LogHandler) GetLogs(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userRole, _ := c.Get("role")

	// 构建查询参数
	params := &LogQueryParams{}

	// 解析查询参数
	if level := c.Query("level"); level != "" {
		params.Level = LogLevel(level)
	}

	if module := c.Query("module"); module != "" {
		params.Module = module
	}

	// 权限控制：普通用户只能查看自己的日志
	if userRole != "admin" {
		params.Uid = userID.(uint)
	} else if userIDStr := c.Query("uid"); userIDStr != "" {
		if uid, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			params.Uid = uint(uid)
		}
	}

	// 解析时间范围
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		params.StartTime = startTimeStr
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		params.EndTime = endTimeStr
	}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}
	if params.Page == 0 {
		params.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			params.PageSize = pageSize
		}
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	// 查询日志
	logs, total, err := h.logService.Query(*params)
	if err != nil {
		h.logService.Error("logs", "查询日志失败", LogField{Key: "error", Value: err.Error()}, LogField{Key: "params", Value: params})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败"})
		return
	}

	// 计算分页信息
	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data": logs,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetLogStats 获取日志统计信息
// GET /api/logs/stats
// 返回值: 各级别日志数量统计
func (h *LogHandler) GetLogStats(c *gin.Context) {
	// 获取当前用户信息
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 获取统计信息
	stats, err := h.logService.GetStats()
	if err != nil {
		h.logService.Error("logs", "获取日志统计失败", LogField{Key: "error", Value: err.Error()}, LogField{Key: "user_id", Value: userID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// CleanupLogs 清理过期日志
// DELETE /api/logs/cleanup
// 仅管理员可用
// 返回值: 清理结果
func (h *LogHandler) CleanupLogs(c *gin.Context) {
	// 权限检查：仅管理员可用
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足，仅管理员可执行此操作"})
		return
	}

	userID, _ := c.Get("uid")

	// 执行清理 (使用默认保留天数90天)
	deleted, err := h.logService.Cleanup(90)
	if err != nil {
		h.logService.Error("logs", "清理日志失败", LogField{Key: "error", Value: err.Error()}, LogField{Key: "operator", Value: userID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清理日志失败"})
		return
	}

	// 记录清理操作
	h.logService.WithUser(userID.(uint)).WithIP(c.ClientIP()).Info("logs", "执行日志清理操作", LogField{Key: "deleted_count", Value: deleted}, LogField{Key: "operator", Value: userID})

	c.JSON(http.StatusOK, gin.H{
		"message":       "日志清理完成",
		"deleted_count": deleted,
	})
}
