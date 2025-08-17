package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	configPkg "teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/utils"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RegisterRoutes 注册监控路由
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/monitor/system", systemMonitorHandler)
	mux.HandleFunc("/api/v1/monitor/business", businessMonitorHandler)
	mux.HandleFunc("/api/v1/monitor/history", historyMonitorHandler)
	mux.HandleFunc("/api/v1/monitor/status", statusHandler)
	mux.Handle("/metrics", metricsHandler())
}

var (
	collector     *Collector
	collectorOnce sync.Once

	// Prometheus 指标
	promSystemCPU = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "system",
		Subsystem: "host",
		Name:      "cpu_percent",
		Help:      "Current CPU usage percentage",
	})

	promSystemMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "system",
		Subsystem: "host",
		Name:      "memory_usage_percent",
		Help:      "Current memory usage percentage",
	})

	promSystemDiskUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "system",
		Subsystem: "host",
		Name:      "disk_usage_percent",
		Help:      "Current disk usage percentage",
	})
)

// Initialize 初始化监控模块
func Initialize(cfg *configPkg.Config) error {
	// 更新监控配置
	UpdateConfig(cfg)

	// 初始化Redis缓存
	if err := InitRedisCache(); err != nil {
		log.Printf("Warning: failed to initialize Redis cache: %v", err)
	}

	// 创建指标收集器
	collector = NewCollector(
		WithMaxHistorySize(cfg.Monitoring.Performance.MaxHistorySize),
		WithCollectionDelay(cfg.Monitoring.Performance.CollectionDelay),
		WithSystemSampleRate(cfg.Monitoring.Performance.SystemSampleRate),
		WithBusinessSampleRate(cfg.Monitoring.Performance.BusinessSampleRate),
		WithMinCollectionInterval(cfg.Monitoring.MinCollectionInterval),
	)

	// 启动指标收集器
	if err := collector.Start(); err != nil {
		return fmt.Errorf("failed to start collector: %w", err)
	}

	// 启动处理协程
	go handleMetrics()

	log.Println("Monitoring module initialized successfully")
	return nil
}

// GetCollector 获取指标收集器实例
func GetCollector() *Collector {
	return collector
}

// Close 关闭监控模块
func Close() error {
	if collector != nil {
		collector.Stop()
	}

	CloseRedisCache()
	return nil
}

// handleMetrics 处理收集到的指标
func handleMetrics() {
	if collector == nil {
		return
	}

	systemChan := collector.GetSystemMetricsChan()
	businessChan := collector.GetBusinessMetricsChan()

	for {
		select {
		case metrics := <-systemChan:
			if metrics != nil {
				// 更新 Prometheus 指标
				promSystemCPU.Set(metrics.CPU)
				promSystemMemoryUsage.Set(metrics.Memory.Usage)
				promSystemDiskUsage.Set(metrics.Disk.Usage)

				// 记录日志
				if metrics.Alert != "" {
					log.Printf("SYSTEM ALERT: %s", metrics.Alert)
				}
			}

		case metrics := <-businessChan:
			if metrics != nil {
				// 记录日志
				if metrics.Alert != "" {
					log.Printf("BUSINESS ALERT: %s", metrics.Alert)
				}
			}

		case <-context.Background().Done():
			return
		}
	}
}

// CollectBusinessMetrics 收集业务指标
func CollectBusinessMetrics() (*BusinessMetrics, error) {
	metrics := &BusinessMetrics{
		Timestamp: time.Now(),
	}

	// 初始化 TeamSpeak 客户端
	cfg := GetConfig()
	if cfg == nil {
		return nil, fmt.Errorf("monitor config not initialized")
	}

	tsClient, err := NewTeamSpeakClient(cfg.Teamspeak)
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
	checkBusinessAlerts(metrics, cfg)

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

	// 尝试从缓存获取系统指标
	var metrics *SystemMetrics
	if cfg := GetConfig(); cfg != nil && cfg.Monitoring.Redis.Enabled {
		cachedMetrics, err := getCachedSystemMetrics()
		if err != nil {
			log.Printf("Failed to get cached system metrics: %v", err)
		} else if cachedMetrics != nil {
			metrics = cachedMetrics
		}
	}

	// 如果缓存未命中，则从收集器获取
	if metrics == nil {
		metrics = collector.GetLastSystemMetrics()
	}

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

	// 尝试从缓存获取业务指标
	var metrics *BusinessMetrics
	if cfg := GetConfig(); cfg != nil && cfg.Monitoring.Redis.Enabled {
		cachedMetrics, err := getCachedBusinessMetrics()
		if err != nil {
			log.Printf("Failed to get cached business metrics: %v", err)
		} else if cachedMetrics != nil {
			metrics = cachedMetrics
		}
	}

	// 如果缓存未命中，则从收集器获取
	if metrics == nil {
		metrics = collector.GetLastBusinessMetrics()
	}

	if metrics == nil {
		utils.Fail(w, http.StatusServiceUnavailable, utils.ErrNoBusinessMetrics, utils.ErrorMessage(utils.ErrNoBusinessMetrics))
		return
	}

	// 返回 JSON 响应
	utils.OK(w, metrics)
}

// historyMonitorHandler 处理历史指标请求
func historyMonitorHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取查询参数
	metricType := r.URL.Query().Get("type")
	if metricType == "" {
		metricType = "system" // 默认返回系统指标
	}

	var history any
	switch metricType {
	case "system":
		history = collector.GetSystemMetricsHistory()
	case "business":
		history = collector.GetBusinessMetricsHistory()
	default:
		utils.Fail(w, http.StatusBadRequest, utils.ErrInvalidRequest, "Invalid metric type")
		return
	}

	// 返回 JSON 响应
	utils.OK(w, history)
}

// statusHandler 处理状态请求
func statusHandler(w http.ResponseWriter, r *http.Request) {
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

	// 获取收集器状态
	status := collector.GetStatus()

	// 返回 JSON 响应
	utils.OK(w, status)
}

// metricsHandler 处理 Prometheus 指标请求
func metricsHandler() http.Handler {
	return promhttp.Handler()
}

// Run 运行监控服务
func Run(cfg *configPkg.Config) error {
	// 初始化监控模块
	if err := Initialize(cfg); err != nil {
		return fmt.Errorf("failed to initialize monitoring module: %w", err)
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":9090", // 监控服务端口
		Handler: mux,
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在 goroutine 中启动服务器
	go func() {
		log.Println("Monitoring server starting on :9090")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Monitoring server error: %v", err)
		}
	}()

	// 等待中断信号
	<-sigChan
	log.Println("Shutting down monitoring server...")

	// 关闭监控模块
	if err := Close(); err != nil {
		log.Printf("Error closing monitoring module: %v", err)
	}

	// 关闭 HTTP 服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown monitoring server: %w", err)
	}

	log.Println("Monitoring server stopped")
	return nil
}
