/**
 * 系统日志数据模型
 * 文件作用：定义系统日志的数据结构和数据库操作方法
 * 作者：AI Assistant
 * 版本：v1.0.0
 */

package logs

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// LogLevel 日志级别枚举
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// SystemLog 系统日志模型
// 用于存储系统运行过程中的各种日志信息
type SystemLog struct {
	// 主键ID
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// 日志级别 (debug, info, warn, error)
	Level LogLevel `json:"level" gorm:"type:varchar(10);not null;index"`

	// 模块名称 (如: auth, instance, deploy, monitor等)
	Module string `json:"module" gorm:"type:varchar(50);not null;index"`

	// 用户ID (可为空，系统日志时为0)
	Uid uint `json:"uid" gorm:"default:0;index"`

	// 客户端IP地址
	IPAddress string `json:"ip_address" gorm:"type:varchar(45)"`

	// 日志消息内容
	Message string `json:"message" gorm:"type:text;not null"`

	// 扩展元数据 (JSON格式存储额外信息)
	Metadata string `json:"metadata" gorm:"type:json"`

	// 创建时间
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// TableName 指定表名
func (SystemLog) TableName() string {
	return "system_logs"
}

// LogQueryParams 日志查询参数
type LogQueryParams struct {
	// 日志级别过滤
	Level LogLevel `json:"level" form:"level"`

	// 模块过滤
	Module string `json:"module" form:"module"`

	// 用户ID过滤
	Uid uint `json:"uid" form:"uid"`

	// 开始时间
	StartTime string `json:"start_time" form:"start_time"`

	// 结束时间
	EndTime string `json:"end_time" form:"end_time"`

	// 关键词搜索
	Keyword string `json:"keyword" form:"keyword"`

	// 分页参数
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// LogStats 日志统计信息
type LogStats struct {
	// 总日志数
	Total int64 `json:"total"`

	// 各级别日志数量
	DebugCount int64 `json:"debug_count"`
	InfoCount  int64 `json:"info_count"`
	WarnCount  int64 `json:"warn_count"`
	ErrorCount int64 `json:"error_count"`

	// 今日新增日志数
	TodayCount int64 `json:"today_count"`

	// 最近错误日志数 (24小时内)
	RecentErrorCount int64 `json:"recent_error_count"`
}

// SetMetadata 设置元数据
// 参数：data - 要序列化为JSON的数据
// 返回值：error - 序列化错误
func (log *SystemLog) SetMetadata(data interface{}) error {
	if data == nil {
		log.Metadata = ""
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	log.Metadata = string(jsonData)
	return nil
}

// GetMetadata 获取元数据
// 参数：result - 用于接收反序列化结果的指针
// 返回值：error - 反序列化错误
func (log *SystemLog) GetMetadata(result interface{}) error {
	if log.Metadata == "" {
		return nil
	}

	return json.Unmarshal([]byte(log.Metadata), result)
}

// AddLog 添加系统日志
// 参数：db - 数据库连接，log - 日志对象
// 返回值：error - 数据库操作错误
func AddLog(db *gorm.DB, log *SystemLog) error {
	log.CreatedAt = time.Now()
	return db.Create(log).Error
}

// GetLogs 查询系统日志
// 参数：db - 数据库连接，params - 查询参数
// 返回值：[]SystemLog - 日志列表，int64 - 总数，error - 查询错误
func GetLogs(db *gorm.DB, params LogQueryParams) ([]SystemLog, int64, error) {
	var logs []SystemLog
	var total int64

	// 构建查询条件
	query := db.Model(&SystemLog{})

	// 级别过滤
	if params.Level != "" {
		query = query.Where("level = ?", params.Level)
	}

	// 模块过滤
	if params.Module != "" {
		query = query.Where("module = ?", params.Module)
	}

	// 用户过滤
	if params.Uid > 0 {
		query = query.Where("uid = ?", params.Uid)
	}

	// 时间范围过滤
	if params.StartTime != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", params.StartTime); err == nil {
			query = query.Where("created_at >= ?", startTime)
		}
	}

	if params.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", params.EndTime); err == nil {
			query = query.Where("created_at <= ?", endTime)
		}
	}

	// 关键词搜索
	if params.Keyword != "" {
		query = query.Where("message LIKE ?", "%"+params.Keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页参数处理
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// 执行查询
	offset := (params.Page - 1) * params.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&logs).Error

	return logs, total, err
}

// GetLogStats 获取日志统计信息
// 参数：db - 数据库连接
// 返回值：LogStats - 统计信息，error - 查询错误
func GetLogStats(db *gorm.DB) (LogStats, error) {
	var stats LogStats

	// 获取总日志数
	if err := db.Model(&SystemLog{}).Count(&stats.Total).Error; err != nil {
		return stats, err
	}

	// 获取各级别日志数量
	db.Model(&SystemLog{}).Where("level = ?", LogLevelDebug).Count(&stats.DebugCount)
	db.Model(&SystemLog{}).Where("level = ?", LogLevelInfo).Count(&stats.InfoCount)
	db.Model(&SystemLog{}).Where("level = ?", LogLevelWarn).Count(&stats.WarnCount)
	db.Model(&SystemLog{}).Where("level = ?", LogLevelError).Count(&stats.ErrorCount)

	// 获取今日新增日志数
	today := time.Now().Format("2006-01-02")
	todayStart, _ := time.Parse("2006-01-02", today)
	todayEnd := todayStart.Add(24 * time.Hour)
	db.Model(&SystemLog{}).Where("created_at >= ? AND created_at < ?", todayStart, todayEnd).Count(&stats.TodayCount)

	// 获取最近24小时错误日志数
	recentTime := time.Now().Add(-24 * time.Hour)
	db.Model(&SystemLog{}).Where("level = ? AND created_at >= ?", LogLevelError, recentTime).Count(&stats.RecentErrorCount)

	return stats, nil
}

// CleanupLogs 清理过期日志
// 参数：db - 数据库连接，retentionDays - 保留天数
// 返回值：int64 - 删除的日志数量，error - 操作错误
func CleanupLogs(db *gorm.DB, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30 // 默认保留30天
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := db.Where("created_at < ?", cutoffTime).Delete(&SystemLog{})
	return result.RowsAffected, result.Error
}
