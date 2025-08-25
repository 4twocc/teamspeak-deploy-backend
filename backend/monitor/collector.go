// collector.go
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	configPkg "teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/logs"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus 指标
var (
	promBusinessOnlineUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "teamspeak",
		Subsystem: "business",
		Name:      "online_users",
		Help:      "Number of online users on the selected TeamSpeak virtual server",
	})
	promBusinessChannels = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "teamspeak",
		Subsystem: "business",
		Name:      "channel_count",
		Help:      "Number of channels on the selected TeamSpeak virtual server",
	})
	promBusinessVoiceQuality = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "teamspeak",
		Subsystem: "business",
		Name:      "voice_quality",
		Help:      "Estimated voice quality score (0-100)",
	})
	promBusinessCollectDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "teamspeak",
		Subsystem: "collector",
		Name:      "business_collect_duration_seconds",
		Help:      "Duration of business metrics collection in seconds",
		Buckets:   prometheus.DefBuckets,
	})
	promMonitorErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "teamspeak",
		Subsystem: "monitor",
		Name:      "errors_total",
		Help:      "Total number of monitoring errors",
	})
	// 在重连成功处自增
	promReconnectsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "teamspeak",
		Subsystem: "monitor",
		Name:      "reconnects_total",
		Help:      "Total number of successful TeamSpeak reconnects",
	})
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert 表示一个告警
type Alert struct {
	Level     AlertLevel `json:"level"`     // Alert level (e.g., "critical", "warning", "info")
	Message   string     `json:"message"`   // Alert message
	Timestamp time.Time  `json:"timestamp"` // When the alert was triggered
	Source    string     `json:"source"`    // Source of the alert (e.g., "system", "business")
}

// SystemMetrics 表示系统级指标
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"` // 采集时间
	CPU       float64   `json:"cpu"`       // CPU 使用率（百分比）
	Memory    struct {
		Total     uint64  `json:"total"`     // 总内存（字节）
		Used      uint64  `json:"used"`      // 已用内存（字节）
		Available uint64  `json:"available"` // 可用内存（字节）
		Usage     float64 `json:"usage"`     // 内存使用率（百分比）
	} `json:"memory"`
	Disk struct {
		Total uint64  `json:"total"` // 总磁盘空间（字节）
		Used  uint64  `json:"used"`  // 已用磁盘空间（字节）
		Free  uint64  `json:"free"`  // 可用磁盘空间（字节）
		Usage float64 `json:"usage"` // 磁盘使用率（百分比）
	} `json:"disk"`
	Network struct {
		RxBytes uint64  `json:"rx_bytes"` // 接收字节数
		TxBytes uint64  `json:"tx_bytes"` // 发送字节数
		RxRate  float64 `json:"rx_rate"`  // 接收速率（字节/秒）
		TxRate  float64 `json:"tx_rate"`  // 发送速率（字节/秒）
	} `json:"network"`
	Uptime uint64 `json:"uptime"` // 系统运行时间（秒）
	Load   struct {
		Load1  float64 `json:"load1"`  // 1分钟平均负载
		Load5  float64 `json:"load5"`  // 5分钟平均负载
		Load15 float64 `json:"load15"` // 15分钟平均负载
	} `json:"load"`
	Alert string `json:"alert,omitempty"` // 告警信息
}

// BusinessMetrics 表示业务指标
type BusinessMetrics struct {
	Timestamp    time.Time     `json:"timestamp"`       // 时间戳
	OnlineUsers  int           `json:"onlineUsers"`     // 在线用户数
	ChannelCount int           `json:"channelCount"`    // 频道数量
	Uptime       time.Duration `json:"uptime"`          // 服务器运行时间
	VoiceQuality float64       `json:"voiceQuality"`    // 语音质量 (0-1)
	Alert        string        `json:"alert,omitempty"` // 告警信息
}

// Collector 指标收集器
type Collector struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 指标通道
	systemMetricsChan   chan *SystemMetrics
	businessMetricsChan chan *BusinessMetrics
	alertChan           chan *Alert

	// 指标缓存
	lastSystemMetrics   *SystemMetrics
	lastBusinessMetrics *BusinessMetrics
	metricsMutex        sync.RWMutex

	// 指标历史记录
	systemMetricsHistory   []*SystemMetrics
	businessMetricsHistory []*BusinessMetrics
	historyMutex           sync.RWMutex
	maxHistorySize         int

	// 状态
	isRunning    bool
	startTime    time.Time
	metricsCount int64
	errorCount   int64
	statusMutex  sync.RWMutex

	// 复用的 TeamSpeak 客户端
	tsClient   *TeamSpeakClient
	tsClientMu sync.Mutex

	// 缓存控制
	lastSystemCollection   time.Time
	lastBusinessCollection time.Time
	collectionMutex        sync.Mutex
	minCollectionInterval  time.Duration

	// 采样率控制
	systemSampleRate   int
	businessSampleRate int
	sampleCounter      struct {
		system   int
		business int
	}
	sampleMutex sync.Mutex

	// 延迟控制
	collectionDelay time.Duration

	// 日志服务
	logService logs.LogService

	// 冷却期控制键
}

// CollectorOption 收集器配置选项
type CollectorOption func(*Collector)

// WithMaxHistorySize 设置最大历史记录数
func WithMaxHistorySize(size int) CollectorOption {
	return func(c *Collector) {
		c.maxHistorySize = size
	}
}

// WithCollectionDelay 设置收集后延迟时间
func WithCollectionDelay(delay time.Duration) CollectorOption {
	return func(c *Collector) {
		c.collectionDelay = delay
	}
}

// WithSystemSampleRate 设置系统指标采样率
func WithSystemSampleRate(rate int) CollectorOption {
	return func(c *Collector) {
		c.systemSampleRate = rate
	}
}

// WithBusinessSampleRate 设置业务指标采样率
func WithBusinessSampleRate(rate int) CollectorOption {
	return func(c *Collector) {
		c.businessSampleRate = rate
	}
}

// WithMinCollectionInterval 设置最小收集间隔
func WithMinCollectionInterval(interval time.Duration) CollectorOption {
	return func(c *Collector) {
		c.minCollectionInterval = interval
	}
}

// NewCollector 创建新的指标收集器
func NewCollector(opts ...CollectorOption) *Collector {
	c := &Collector{
		systemMetricsChan:   make(chan *SystemMetrics, 10),
		businessMetricsChan: make(chan *BusinessMetrics, 10),
		alertChan:           make(chan *Alert, 10),
		maxHistorySize:      100, // 默认值
		startTime:           time.Now(),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(c)
	}

	// 从配置加载参数
	cfg := GetConfig()
	if cfg != nil {
		monitoringCfg := cfg.Monitoring

		// 如果没有通过选项设置，则使用配置文件中的值
		if c.maxHistorySize == 100 && monitoringCfg.Performance.MaxHistorySize > 0 {
			c.maxHistorySize = monitoringCfg.Performance.MaxHistorySize
		}

		if c.collectionDelay == 0 && monitoringCfg.Performance.CollectionDelay > 0 {
			c.collectionDelay = monitoringCfg.Performance.CollectionDelay
		}

		if c.systemSampleRate == 0 && monitoringCfg.Performance.SystemSampleRate > 0 {
			c.systemSampleRate = monitoringCfg.Performance.SystemSampleRate
		}

		if c.businessSampleRate == 0 && monitoringCfg.Performance.BusinessSampleRate > 0 {
			c.businessSampleRate = monitoringCfg.Performance.BusinessSampleRate
		}

		if c.minCollectionInterval == 0 && monitoringCfg.MinCollectionInterval > 0 {
			c.minCollectionInterval = monitoringCfg.MinCollectionInterval
		}
	}

	// 设置默认值
	if c.maxHistorySize <= 0 {
		c.maxHistorySize = 100
	}
	if c.systemSampleRate <= 0 {
		c.systemSampleRate = 1
	}
	if c.businessSampleRate <= 0 {
		c.businessSampleRate = 1
	}
	if c.minCollectionInterval <= 0 {
		c.minCollectionInterval = time.Second * 30
	}

	// 初始化日志服务
	if cfg != nil && database.DB != nil {
		if service, err := logs.NewLogService(database.DB, logs.LogConfig{
			Level:         cfg.Logging.Level,
			EnableDB:      cfg.Logging.EnableDatabase,
			RetentionDays: cfg.Logging.RetentionDays,
			BatchSize:     cfg.Logging.BatchSize,
			BatchInterval: int(cfg.Logging.FlushInterval.Seconds()),
			EnableFile:    true,
			FilePath:      cfg.Logging.OutputFile,
		}); err == nil {
			c.logService = service
		}
	}

	return c
}

// GetSystemMetricsChan 返回系统指标通道
func (c *Collector) GetSystemMetricsChan() <-chan *SystemMetrics {
	return c.systemMetricsChan
}

// GetBusinessMetricsChan 返回业务指标通道
func (c *Collector) GetBusinessMetricsChan() <-chan *BusinessMetrics {
	return c.businessMetricsChan
}

// GetAlertChan 返回告警通道
func (c *Collector) GetAlertChan() <-chan *Alert {
	return c.alertChan
}

// GetLastSystemMetrics 获取最新的系统指标
func (c *Collector) GetLastSystemMetrics() *SystemMetrics {
	c.metricsMutex.RLock()
	defer c.metricsMutex.RUnlock()
	return c.lastSystemMetrics
}

// GetLastBusinessMetrics 获取最新的业务指标
func (c *Collector) GetLastBusinessMetrics() *BusinessMetrics {
	c.metricsMutex.RLock()
	defer c.metricsMutex.RUnlock()
	return c.lastBusinessMetrics
}

// GetSystemMetricsHistory 获取系统指标历史记录
func (c *Collector) GetSystemMetricsHistory() []*SystemMetrics {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	// 返回历史记录的副本
	history := make([]*SystemMetrics, len(c.systemMetricsHistory))
	copy(history, c.systemMetricsHistory)
	return history
}

// GetBusinessMetricsHistory 获取业务指标历史记录
func (c *Collector) GetBusinessMetricsHistory() []*BusinessMetrics {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	// 返回历史记录的副本
	history := make([]*BusinessMetrics, len(c.businessMetricsHistory))
	copy(history, c.businessMetricsHistory)
	return history
}

// recordMetrics 记录指标收集次数
func (c *Collector) recordMetrics() {
	c.statusMutex.Lock()
	c.metricsCount++
	c.statusMutex.Unlock()
}

// recordError 记录错误次数
func (c *Collector) recordError() {
	c.statusMutex.Lock()
	c.errorCount++
	c.statusMutex.Unlock()
	promMonitorErrorsTotal.Inc()
}

// IsRunning 检查收集器是否正在运行
func (c *Collector) IsRunning() bool {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.isRunning
}

// GetStatus 获取收集器状态
func (c *Collector) GetStatus() map[string]any {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()

	return map[string]any{
		"running":       c.isRunning,
		"start_time":    c.startTime,
		"metrics_count": c.metricsCount,
		"error_count":   c.errorCount,
		"uptime":        time.Since(c.startTime).String(),
	}
}

// updateLastCollection 更新上次收集时间
func (c *Collector) updateLastCollection(isSystem bool) {
	c.collectionMutex.Lock()
	defer c.collectionMutex.Unlock()

	if isSystem {
		c.lastSystemCollection = time.Now()
	} else {
		c.lastBusinessCollection = time.Now()
	}
}

// shouldCollect 检查是否应该收集指标（基于采样率和最小间隔）
func (c *Collector) shouldCollect(isSystem bool) bool {
	// 检查最小收集间隔
	c.collectionMutex.Lock()
	lastCollection := c.lastSystemCollection
	if !isSystem {
		lastCollection = c.lastBusinessCollection
	}
	c.collectionMutex.Unlock()

	if time.Since(lastCollection) < c.minCollectionInterval {
		return false
	}

	// 检查采样率
	c.sampleMutex.Lock()
	defer c.sampleMutex.Unlock()

	if isSystem {
		c.sampleCounter.system++
		if c.sampleCounter.system%c.systemSampleRate != 0 {
			return false
		}
	} else {
		c.sampleCounter.business++
		if c.sampleCounter.business%c.businessSampleRate != 0 {
			return false
		}
	}

	return true
}

// collectSystemMetrics 收集系统指标
func (c *Collector) collectSystemMetrics() {
	defer c.wg.Done()

	cfg := GetConfig()
	var collectInterval time.Duration
	if cfg != nil && cfg.Monitoring.CollectInterval > 0 {
		collectInterval = cfg.Monitoring.CollectInterval
	} else {
		collectInterval = time.Hour // 默认值
	}

	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()

	// 立即收集一次
	c.collectSystemMetricsOnce()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectSystemMetricsOnce()
		}
	}
}

// collectSystemMetricsOnce 执行一次系统指标收集
func (c *Collector) collectSystemMetricsOnce() {
	// 检查是否应该收集
	if !c.shouldCollect(true) {
		return
	}

	// 延迟收集（避免系统负载过高）
	if c.collectionDelay > 0 {
		time.Sleep(c.collectionDelay)
	}

	var metrics *SystemMetrics
	var err error

	start := time.Now()
	defer func() {
		promBusinessCollectDuration.Observe(time.Since(start).Seconds())
	}()

	metrics, err = CollectSystemMetrics()
	if err != nil {
		c.recordError()
		c.alertChan <- &Alert{
			Level:   AlertLevelError,
			Message: fmt.Sprintf("收集系统指标失败: %v", err),
		}
		return
	}

	// 更新上次收集时间
	c.updateLastCollection(true)

	// Update latest metrics
	c.metricsMutex.Lock()
	c.lastSystemMetrics = metrics
	c.metricsMutex.Unlock()

	// Add to history
	c.historyMutex.Lock()
	c.systemMetricsHistory = append(c.systemMetricsHistory, metrics)
	// Limit history size
	if len(c.systemMetricsHistory) > c.maxHistorySize {
		c.systemMetricsHistory = c.systemMetricsHistory[1:]
	}
	c.historyMutex.Unlock()

	// Record metrics
	c.recordMetrics()

	// 缓存到Redis
	if cfg := GetConfig(); cfg != nil && cfg.Monitoring.Redis.Enabled {
		if err := cacheSystemMetrics(metrics); err != nil {
			if c.logService != nil {
				c.logService.Error("monitor", "Failed to cache system metrics", logs.LogField{Key: "error", Value: err})
			} else {
				log.Printf("Failed to cache system metrics: %v", err)
			}
		}
	}

	// Send to channel
	select {
	case c.systemMetricsChan <- metrics:
	default:
		log.Println("System metrics channel is full, dropping data")
	}
}

const maxRetries = 3

// collectBusinessMetrics 收集业务指标
func (c *Collector) collectBusinessMetrics() {
	defer c.wg.Done()

	cfg := GetConfig()
	var collectInterval time.Duration
	if cfg != nil && cfg.Monitoring.CollectInterval > 0 {
		collectInterval = cfg.Monitoring.CollectInterval
	} else {
		collectInterval = time.Hour // 默认值
	}

	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()

	// 立即收集一次
	c.collectBusinessMetricsOnce()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectBusinessMetricsOnce()
		}
	}
}

// collectBusinessMetricsOnce 执行一次业务指标收集
func (c *Collector) collectBusinessMetricsOnce() {
	// 检查是否应该收集
	if !c.shouldCollect(false) {
		return
	}

	// 延迟收集（避免系统负载过高）
	if c.collectionDelay > 0 {
		time.Sleep(c.collectionDelay)
	}

	var metrics *BusinessMetrics
	var err error

	start := time.Now()
	defer func() {
		promBusinessCollectDuration.Observe(time.Since(start).Seconds())
	}()

	// 确保复用客户端存在
	c.tsClientMu.Lock()
	if c.tsClient == nil {
		cfg := GetConfig()
		if cfg != nil {
			if ts, e := NewTeamSpeakClient(cfg.Teamspeak); e != nil {
				err = fmt.Errorf("init TeamSpeak client failed: %w", e)
			} else {
				c.tsClient = ts
			}
		}
	}
	ts := c.tsClient
	c.tsClientMu.Unlock()

	if err == nil && ts == nil {
		err = fmt.Errorf("teamspeak client not available")
	}

	// Try to collect metrics with retry on the shared client
	for attempt := 0; err == nil && attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Second * time.Duration(attempt*2)
			if c.logService != nil {
				c.logService.Info("monitor", "Retrying business metrics collection",
					logs.LogField{Key: "backoff", Value: backoff},
					logs.LogField{Key: "attempt", Value: attempt + 1},
					logs.LogField{Key: "maxRetries", Value: maxRetries})
			} else {
				log.Printf("Retrying in %v... (attempt %d/%d)", backoff, attempt+1, maxRetries)
			}
			time.Sleep(backoff)
		}

		// 自愈连接
		if e := ts.ensureConnected(); e != nil {
			err = fmt.Errorf("ensureConnected failed: %w", e)
			continue
		}

		// 获取服务器信息并构造业务指标
		srv, e := ts.GetServerInfo()
		if e != nil {
			err = e
			continue
		}

		m := &BusinessMetrics{Timestamp: time.Now()}
		m.OnlineUsers = srv.OnlineUsers
		m.ChannelCount = srv.ChannelCount
		m.Uptime = srv.Uptime
		m.VoiceQuality = srv.VoiceQuality / 100.0
		if m.VoiceQuality < 0.7 {
			m.VoiceQuality = 0.7
		}

		cfg := GetConfig()
		if cfg != nil {
			checkBusinessAlerts(m, cfg)
		}
		metrics = m
		err = nil
		break
	}

	if err != nil {
		c.recordError()

		// 发送告警
		c.alertChan <- &Alert{
			Level:   AlertLevelError,
			Message: fmt.Sprintf("收集业务指标失败: %v", err),
		}
		return
	}

	// 更新上次收集时间
	c.updateLastCollection(false)

	// Update latest metrics
	c.metricsMutex.Lock()
	c.lastBusinessMetrics = metrics
	c.metricsMutex.Unlock()

	// Prometheus: 更新业务指标
	promBusinessOnlineUsers.Set(float64(metrics.OnlineUsers))
	promBusinessChannels.Set(float64(metrics.ChannelCount))
	promBusinessVoiceQuality.Set(metrics.VoiceQuality * 100.0)

	// Add to history
	c.historyMutex.Lock()
	c.businessMetricsHistory = append(c.businessMetricsHistory, metrics)
	// Limit history size
	if len(c.businessMetricsHistory) > c.maxHistorySize {
		c.businessMetricsHistory = c.businessMetricsHistory[1:]
	}
	c.historyMutex.Unlock()

	// Record metrics
	c.recordMetrics()

	// 缓存到Redis
	if cfg := GetConfig(); cfg != nil && cfg.Monitoring.Redis.Enabled {
		if err := cacheBusinessMetrics(metrics); err != nil {
			if c.logService != nil {
				c.logService.Error("monitor", "Failed to cache business metrics", logs.LogField{Key: "error", Value: err.Error()})
			} else {
				log.Printf("Failed to cache business metrics: %v", err)
			}
		}
	}

	// Send to channel
	select {
	case c.businessMetricsChan <- metrics:
	default:
		if logService != nil {
			logService.Warn("monitor", "Business metrics channel is full, dropping data")
		} else {
			log.Println("Business metrics channel is full, dropping data")
		}
	}
}

// processAlerts 处理告警
func (c *Collector) processAlerts() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case alert := <-c.alertChan:
			// 处理告警（例如：记录日志、发送邮件等）
			if logService != nil {
				logService.Warn("monitor", "ALERT", logs.LogField{Key: "level", Value: string(alert.Level)}, logs.LogField{Key: "message", Value: alert.Message})
			} else {
				log.Printf("ALERT [%s]: %s", alert.Level, alert.Message)
			}

			// 这里可以添加更多的告警处理逻辑
			// 例如发送邮件、短信、调用 webhook 等
		}
	}
}

// SaveToFile 保存指标到文件
func (c *Collector) SaveToFile(filename string) error {
	c.metricsMutex.RLock()
	data := struct {
		SystemMetrics   *SystemMetrics   `json:"system_metrics,omitempty"`
		BusinessMetrics *BusinessMetrics `json:"business_metrics,omitempty"`
	}{
		SystemMetrics:   c.lastSystemMetrics,
		BusinessMetrics: c.lastBusinessMetrics,
	}
	c.metricsMutex.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("编码数据失败: %v", err)
	}

	return nil
}

// LoadFromFile 从文件加载指标
func (c *Collector) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	var data struct {
		SystemMetrics   *SystemMetrics
		BusinessMetrics *BusinessMetrics
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("解码数据失败: %v", err)
	}

	c.metricsMutex.Lock()
	defer c.metricsMutex.Unlock()

	if data.SystemMetrics != nil {
		c.lastSystemMetrics = data.SystemMetrics
	}
	if data.BusinessMetrics != nil {
		c.lastBusinessMetrics = data.BusinessMetrics
	}

	return nil
}

// 清理过期的历史数据
func (c *Collector) CleanupHistory(maxAge time.Duration) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()

	now := time.Now()
	var newSystemHistory []*SystemMetrics
	var newBusinessHistory []*BusinessMetrics

	// 清理系统指标历史
	for _, m := range c.systemMetricsHistory {
		if now.Sub(m.Timestamp) <= maxAge {
			newSystemHistory = append(newSystemHistory, m)
		}
	}
	c.systemMetricsHistory = newSystemHistory

	// 清理业务指标历史
	for _, m := range c.businessMetricsHistory {
		if now.Sub(m.Timestamp) <= maxAge {
			newBusinessHistory = append(newBusinessHistory, m)
		}
	}
	c.businessMetricsHistory = newBusinessHistory
}

// Start 启动收集器
func (c *Collector) Start() error {
	if c.IsRunning() {
		return fmt.Errorf("collector already started")
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.statusMutex.Lock()
	c.isRunning = true
	c.statusMutex.Unlock()
	c.wg.Add(4) // 三个收集/处理 + 一个定时清理

	// 预创建 TeamSpeak 客户端（失败不阻塞启动，采集路径会自愈）
	go func() {
		c.tsClientMu.Lock()
		defer c.tsClientMu.Unlock()

		cfg := GetConfig()
		if cfg != nil && c.tsClient == nil {
			if ts, err := NewTeamSpeakClient(cfg.Teamspeak); err != nil {
				log.Printf("Init TeamSpeak client failed: %v", err)
			} else {
				c.tsClient = ts
			}
		}
	}()

	// 启动系统指标收集
	go c.collectSystemMetrics()

	// 启动业务指标收集
	go c.collectBusinessMetrics()

	// 启动告警处理
	go c.processAlerts()

	// 启动定时清理任务
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				c.CleanupHistory(24 * time.Hour) // 保留24小时数据
			}
		}
	}()

	return nil
}

// Stop 停止指标收集
func (c *Collector) Stop() {
	c.statusMutex.Lock()
	if !c.isRunning {
		c.statusMutex.Unlock()
		return
	}
	c.isRunning = false
	c.statusMutex.Unlock()

	log.Println("Stopping metrics collector...")

	// 取消上下文，通知所有goroutine退出
	c.cancel()

	// 等待所有goroutine退出
	c.wg.Wait()

	// 关闭 TeamSpeak 客户端
	c.tsClientMu.Lock()
	if c.tsClient != nil {
		_ = c.tsClient.Close()
		c.tsClient = nil
	}
	c.tsClientMu.Unlock()

	log.Println("Metrics collector stopped")
}

// checkBusinessAlerts 检查业务指标告警
func checkBusinessAlerts(metrics *BusinessMetrics, cfg *configPkg.Config) {
	if cfg == nil {
		return
	}

	thresholds := cfg.Monitoring.Alert.Thresholds
	if metrics.VoiceQuality < thresholds.VoiceQuality {
		metrics.Alert = fmt.Sprintf("语音质量较低: %.2f", metrics.VoiceQuality)
	}
}
