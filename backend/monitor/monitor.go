package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/utils"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

// 告警级别
const (
	AlertLevelInfo     = "info"
	AlertLevelWarning  = "warning"
	AlertLevelError    = "error"
	AlertLevelCritical = "critical"
)

// Alert 表示一个告警
type Alert struct {
	Level     string    `json:"level"`     // Alert level (e.g., "critical", "warning", "info")
	Message   string    `json:"message"`   // Alert message
	Timestamp time.Time `json:"timestamp"` // When the alert was triggered
	Source    string    `json:"source"`    // Source of the alert (e.g., "system", "business")
}

var (
	lastNetwork struct {
		rxBytes uint64
		txBytes uint64
		time    time.Time
	}
	networkMutex sync.Mutex
)

// 全局收集器实例
var (
	collector *Collector
	once      sync.Once
)

// RegisterRoutes sets up the HTTP routes for monitoring endpoints
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/monitor/system", systemMonitorHandler)
	mux.HandleFunc("/api/monitor/business", businessMonitorHandler)
	mux.HandleFunc("/api/monitor/stats", statsHandler)   // 新增统计接口
	mux.HandleFunc("/api/monitor/health", healthHandler) // 新增健康检查接口
	// Prometheus 指标
	mux.Handle("/metrics", promhttp.Handler())
}

// CollectBusinessMetrics 收集业务指标
func CollectBusinessMetrics() (*BusinessMetrics, error) {
	metrics := &BusinessMetrics{
		Timestamp: time.Now(),
	}

	// 初始化 TeamSpeak 客户端
	tsClient, err := NewTeamSpeakClient(GetConfig().TeamSpeakConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create TeamSpeak client: %w", err)
	}
	defer tsClient.Close()

	// 获取服务器信息
	serverInfo, err := tsClient.GetServerInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	// 设置业务指标（ServerInfo 已包含在线用户、频道数、语音质量与运行时间）
	metrics.OnlineUsers = serverInfo.OnlineUsers
	metrics.ChannelCount = serverInfo.ChannelCount
	metrics.Uptime = serverInfo.Uptime
	// Convert 0-100 quality to 0-1 scale
	metrics.VoiceQuality = max(serverInfo.VoiceQuality/100.0, 0.7)

	// 检查告警
	checkBusinessAlerts(metrics)

	return metrics, nil
}

// systemMonitorHandler 处理系统指标请求
func systemMonitorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Fail(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed, utils.ErrorMessage(utils.ErrMethodNotAllowed))
		return
	}

	// 获取收集器实例
	collector := GetCollector()
	if collector == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrCollectorInit, utils.ErrorMessage(utils.ErrCollectorInit))
		return
	}

	// 获取系统指标
	metrics := collector.GetLastSystemMetrics()
	if metrics == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrNoSystemMetrics, utils.ErrorMessage(utils.ErrNoSystemMetrics))
		return
	}

	// 返回 JSON 响应
	utils.OK(w, metrics)
}

// businessMonitorHandler 处理业务指标请求
func businessMonitorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Fail(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed, utils.ErrorMessage(utils.ErrMethodNotAllowed))
		return
	}

	// 获取收集器实例
	collector := GetCollector()
	if collector == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrCollectorInit, utils.ErrorMessage(utils.ErrCollectorInit))
		return
	}

	// 获取业务指标
	metrics := collector.GetLastBusinessMetrics()
	if metrics == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrNoBusinessMetrics, utils.ErrorMessage(utils.ErrNoBusinessMetrics))
		return
	}

	// 返回 JSON 响应
	utils.OK(w, metrics)
}

// 获取监控统计数据
func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Fail(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed, utils.ErrorMessage(utils.ErrMethodNotAllowed))
		return
	}

	// 添加限流
	if !rateLimiter.Allow() {
		utils.Fail(w, http.StatusTooManyRequests, utils.ErrTooManyRequests, utils.ErrorMessage(utils.ErrTooManyRequests))
		return
	}

	// 获取时间范围参数
	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h" // 默认1小时
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		utils.Fail(w, http.StatusBadRequest, utils.ErrInvalidDuration, utils.ErrorMessage(utils.ErrInvalidDuration))
		return
	}

	// 验证时间范围
	maxDuration := 24 * time.Hour
	if duration > maxDuration {
		utils.Fail(w, http.StatusBadRequest, utils.ErrDurationTooLong, fmt.Sprintf("%s: %s", utils.ErrorMessage(utils.ErrDurationTooLong), maxDuration))
		return
	}

	// 获取收集器实例
	collector := GetCollector()
	if collector == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrCollectorInit, utils.ErrorMessage(utils.ErrCollectorInit))
		return
	}

	// 创建带有超时的上下文
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// 使用通道来收集结果
	type result struct {
		system   *SystemMetrics
		business *BusinessMetrics
		err      error
	}

	resultCh := make(chan result, 1)

	// 在goroutine中获取统计数据
	go func() {
		systemStats := collector.GetSystemStats(duration)
		businessStats := collector.GetBusinessStats(duration)
		resultCh <- result{system: systemStats, business: businessStats}
	}()

	// 等待结果或超时
	select {
	case <-ctx.Done():
		utils.Fail(w, http.StatusRequestTimeout, utils.ErrRequestTimeout, utils.ErrorMessage(utils.ErrRequestTimeout))
		return
	case res := <-resultCh:
		if res.err != nil {
			utils.Fail(w, http.StatusInternalServerError, utils.ErrCollectStatsFailed, utils.ErrorMessage(utils.ErrCollectStatsFailed))
			return
		}
		utils.OK(w, map[string]any{
			"system":   res.system,
			"business": res.business,
		})
	}
}

// 健康检查，验证数据库连接和 TeamSpeak 连接
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// 添加恢复机制
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in health check: %v", r)
			utils.Fail(w, http.StatusInternalServerError, utils.ErrInternalServer, "Internal server error during health check")
		}
	}()

	if r.Method != http.MethodGet {
		utils.Fail(w, http.StatusMethodNotAllowed, utils.ErrMethodNotAllowed, "Method not allowed")
		return
	}

	if !rateLimiter.Allow() {
		utils.Fail(w, http.StatusTooManyRequests, utils.ErrTooManyRequests, "Too many requests")
		return
	}

	// 检查数据库连接
	dbStatus := checkDatabaseHealth()

	// 检查 TeamSpeak 连接
	tsStatus := checkTeamSpeakHealth()

	// 组合状态
	status := map[string]any{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  dbStatus,
		"teamspeak": tsStatus,
	}

	// 如果有任何服务不可用，返回 503
	if dbStatus["status"] != "ok" || tsStatus["status"] != "ok" {
		status["status"] = "error"
		log.Printf("Health check failed - Database: %v, TeamSpeak: %v", dbStatus["message"], tsStatus["message"])
		utils.WriteJSON(w, http.StatusServiceUnavailable, utils.ErrServiceUnavailable, "Service Unavailable", status)
		return
	}

	utils.WriteJSON(w, http.StatusOK, 0, "ok", status)
}

// 检查数据库健康状态
func checkDatabaseHealth() map[string]any {
	result := map[string]any{
		"status":  "ok",
		"message": "Database connection is healthy",
	}

	// 检查数据库连接是否初始化
	if database.DB == nil {
		err := fmt.Errorf("database connection is not initialized")
		log.Printf("Database error: %v", err)
		result["status"] = "error"
		result["message"] = "Database not initialized"
		result["error"] = err.Error()
		return result
	}

	// 获取底层 sql.DB 连接
	sqlDB, err := database.DB.DB()
	if err != nil {
		log.Printf("Failed to get database instance: %v", err)
		result["status"] = "error"
		result["message"] = "Failed to get database instance"
		result["error"] = err.Error()
		return result
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行带超时的 Ping 测试
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Printf("Database ping failed: %v", err)
		result["status"] = "error"
		result["message"] = "Database connection failed"
		result["error"] = err.Error()
		return result
	}

	// 检查 users 表是否存在
	var tableExists bool
	err = database.DB.WithContext(ctx).Raw(
		"SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)",
		"users",
	).Scan(&tableExists).Error

	if err != nil {
		log.Printf("Failed to check users table: %v", err)
		result["status"] = "warning"
		result["message"] = "Unable to verify users table"
		result["error"] = err.Error()
	} else if !tableExists {
		log.Printf("Users table does not exist")
		result["status"] = "warning"
		result["message"] = "Users table does not exist"
	}

	return result
}

// 检查 TeamSpeak 连接健康状态
func checkTeamSpeakHealth() map[string]any {
	result := map[string]any{
		"status":  "ok",
		"message": "TeamSpeak connection is healthy",
	}

	// 获取配置
	cfg := GetConfig()
	if cfg == nil {
		err := fmt.Errorf("failed to get configuration")
		log.Printf("TeamSpeak config error: %v", err)
		result["status"] = "error"
		result["message"] = "Configuration error"
		result["error"] = err.Error()
		return result
	}

	// 创建临时客户端以验证连接与上下文
	tsClient, err := NewTeamSpeakClient(cfg.TeamSpeakConfig)
	if err != nil {
		log.Printf("Failed to create TeamSpeak client: %v", err)
		result["status"] = "error"
		result["message"] = "Failed to create TeamSpeak client"
		result["error"] = err.Error()
		return result
	}
	defer func() {
		if err := tsClient.Close(); err != nil {
			log.Printf("Error closing TeamSpeak client: %v", err)
		}
	}()

	// 测试连接（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	var versionErr error

	go func() {
		_, versionErr = tsClient.client.Version()
		close(done)
	}()

	select {
	case <-ctx.Done():
		err := fmt.Errorf("connection timeout: %v", ctx.Err())
		log.Printf("TeamSpeak connection timeout: %v", err)
		result["status"] = "error"
		result["message"] = "Connection to TeamSpeak server timed out"
		result["error"] = err.Error()
	case <-done:
		if versionErr != nil {
			log.Printf("Failed to get TeamSpeak version: %v", versionErr)
			result["status"] = "error"
			result["message"] = "Failed to communicate with TeamSpeak server"
			result["error"] = versionErr.Error()
		}
	}

	return result
}

// 全局速率限制器
var rateLimiter = rate.NewLimiter(rate.Every(time.Second), 10) // 每秒最多10个请求

// 在 init 函数中启动收集器
func init() {
	once.Do(func() {
		collector = NewCollector(WithMaxHistorySize(1000)) // 增加默认历史记录大小
		if err := collector.Start(); err != nil {
			log.Printf("Failed to start metrics collector: %v", err)
		}
	})
}

// 在 main 函数退出时停止收集器
func Cleanup() {
	if collector != nil {
		collector.Stop()
	}
}

// GetCollector 获取收集器实例
func GetCollector() *Collector {
	return collector
}
