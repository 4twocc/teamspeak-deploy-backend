// collector.go
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
	"time"

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
}

// CollectorOption 收集器配置选项
type CollectorOption func(*Collector)

// WithMaxHistorySize 设置最大历史记录数
func WithMaxHistorySize(size int) CollectorOption {
	return func(c *Collector) {
		c.maxHistorySize = size
	}
}

// WithMinCollectionInterval 设置最小收集间隔
func WithMinCollectionInterval(interval time.Duration) CollectorOption {
	return func(c *Collector) {
		c.minCollectionInterval = interval
	}
}

// WithSystemSampleRate 设置系统指标采样率 (1/n 的频率收集)
func WithSystemSampleRate(rate int) CollectorOption {
	return func(c *Collector) {
		c.systemSampleRate = rate
	}
}

// WithBusinessSampleRate 设置业务指标采样率 (1/n 的频率收集)
func WithBusinessSampleRate(rate int) CollectorOption {
	return func(c *Collector) {
		c.businessSampleRate = rate
	}
}

// NewCollector 创建新的指标收集器
func NewCollector(opts ...CollectorOption) *Collector {
	ctx, cancel := context.WithCancel(context.Background())
	collector := &Collector{
		ctx:                    ctx,
		cancel:                 cancel,
		systemMetricsChan:      make(chan *SystemMetrics, 100),
		businessMetricsChan:    make(chan *BusinessMetrics, 100),
		alertChan:              make(chan *Alert, 1000),
		systemMetricsHistory:   make([]*SystemMetrics, 0),
		businessMetricsHistory: make([]*BusinessMetrics, 0),
		maxHistorySize:         50, // 减少默认历史记录大小
		startTime:              time.Now(),
		minCollectionInterval:  15 * time.Second, // 增加最小收集间隔到15秒
		systemSampleRate:       2, // 系统指标每2次只收集1次
		businessSampleRate:     3, // 业务指标每3次只收集1次
	}

	// 应用选项
	for _, opt := range opts {
		opt(collector)
	}

	return collector
}

// IsRunning 返回收集器是否正在运行
func (c *Collector) IsRunning() bool {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.isRunning
}

// GetUptime 返回收集器运行时间
func (c *Collector) GetUptime() time.Duration {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return time.Since(c.startTime)
}

// GetMetricsCount 返回收集的指标总数
func (c *Collector) GetMetricsCount() int64 {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.metricsCount
}

// GetErrorCount 返回错误计数
func (c *Collector) GetErrorCount() int64 {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.errorCount
}

// 记录错误
func (c *Collector) recordError() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.errorCount++
	promMonitorErrorsTotal.Inc()

	// If we have too many errors in a row, send a critical alert
	if c.errorCount >= 5 {
		c.alertChan <- &Alert{
			Level:   AlertLevelCritical,
			Message: "连续多次收集指标失败，请检查系统状态",
		}
		// Reset counter after sending critical alert
		c.errorCount = 0
	}
}

// 记录指标
func (c *Collector) recordMetrics() {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.metricsCount++
}

// withRetry executes a function with exponential backoff retry
func withRetry(operation func() error, maxRetries int, initialBackoff time.Duration) error {
	var err error
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		if err = operation(); err == nil {
			return nil
		}

		log.Printf("Operation failed (attempt %d/%d): %v", i+1, maxRetries, err)
	}

	return fmt.Errorf("after %d attempts: %w", maxRetries, err)
}

// shouldCollect 判断是否应该执行收集操作（实现缓存和最小间隔控制）
func (c *Collector) shouldCollect(lastCollection time.Time) bool {
	c.collectionMutex.Lock()
	defer c.collectionMutex.Unlock()
	
	// 检查距离上次收集是否已经超过最小间隔
	return time.Since(lastCollection) >= c.minCollectionInterval
}

// shouldSample 判断是否应该进行采样收集
func (c *Collector) shouldSample(isSystem bool) bool {
	c.sampleMutex.Lock()
	defer c.sampleMutex.Unlock()
	
	if isSystem {
		c.sampleCounter.system++
		if c.systemSampleRate <= 1 {
			return true
		}
		return c.sampleCounter.system % c.systemSampleRate == 1
	} else {
		c.sampleCounter.business++
		if c.businessSampleRate <= 1 {
			return true
		}
		return c.sampleCounter.business % c.businessSampleRate == 1
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

// 收集系统指标
func (c *Collector) collectSystemMetrics() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in collectSystemMetrics: %v", r)
			// Record the panic as an error
			c.recordError()
		}
	}()

	defer c.wg.Done()

	ticker := time.NewTicker(GetConfig().CollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// 检查是否应该采样收集
			if !c.shouldSample(true) {
				continue
			}
			
			// 检查是否应该收集（避免过于频繁的IO操作）
			if !c.shouldCollect(c.lastSystemCollection) {
				log.Printf("Skipping system metrics collection due to minimum interval constraint")
				continue
			}

			// Use retry for collecting metrics
			var metrics *SystemMetrics
			err := withRetry(func() error {
				var err error
				metrics, err = CollectSystemMetrics()
				return err
			}, 2, 2*time.Second) // 减少重试次数和增加初始退避时间

			if err != nil {
				c.recordError()
				c.alertChan <- &Alert{
					Level:   AlertLevelError,
					Message: fmt.Sprintf("收集系统指标失败: %v", err),
				}
				continue
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

			// Send to channel
			select {
			case c.systemMetricsChan <- metrics:
			default:
				log.Println("WARNING: System metrics channel is full, dropping metrics")
			}
			
			// 在每次收集后增加延迟，减轻系统压力
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// 收集业务指标
func (c *Collector) collectBusinessMetrics() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in collectBusinessMetrics: %v", r)
			c.recordError()
		}
	}()

	defer c.wg.Done()

	ticker := time.NewTicker(GetConfig().CollectInterval)
	defer ticker.Stop()

	var lastErr error
	var retryCount int
	const maxRetries = 2 // 减少重试次数

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// 检查是否应该采样收集
			if !c.shouldSample(false) {
				continue
			}
			
			// 检查是否应该收集（避免过于频繁的IO操作）
			if !c.shouldCollect(c.lastBusinessCollection) {
				log.Printf("Skipping business metrics collection due to minimum interval constraint")
				continue
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
				if ts, e := NewTeamSpeakClient(GetConfig().TeamSpeakConfig); e != nil {
					err = fmt.Errorf("init TeamSpeak client failed: %w", e)
				} else {
					c.tsClient = ts
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
					log.Printf("Retrying in %v... (attempt %d/%d)", backoff, attempt+1, maxRetries)
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
				checkBusinessAlerts(m)
				metrics = m
				err = nil
				break
			}

			if err != nil {
				retryCount++
				c.recordError()

				// 首次失败或错误变化时发送告警
				if retryCount == 1 || (lastErr != nil && lastErr.Error() != err.Error()) {
					c.alertChan <- &Alert{
						Level:   AlertLevelError,
						Message: fmt.Sprintf("收集业务指标失败: %v", err),
					}
				}
				lastErr = err
				continue
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

			// Send to channel
			select {
			case c.businessMetricsChan <- metrics:
			default:
				log.Println("Business metrics channel is full, dropping data")
			}
			
			// 在每次收集后增加延迟，减轻系统压力
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// 处理告警
func (c *Collector) processAlerts() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case alert := <-c.alertChan:
			c.handleAlert(alert)
		}
	}
}

// 处理单个告警
func (c *Collector) handleAlert(alert *Alert) {
	// 记录日志
	log.Printf("[%s] %s", alert.Level, alert.Message)

	// 根据配置发送告警通知
	config := GetConfig()
	if !config.AlertConfig.Enabled {
		return
	}

	for _, method := range config.AlertConfig.NotifyMethods {
		switch method {
		case "console":
			// 控制台输出
			fmt.Printf("[%s] %s\n", alert.Level, alert.Message)
		case "email":
			// 发送邮件通知
			go c.sendEmailAlert(alert)
		case "webhook":
			// 调用Webhook
			go c.sendWebhookAlert(alert)
		}
	}
}

// 发送邮件告警
func (c *Collector) sendEmailAlert(alert *Alert) {
	// 实现邮件发送逻辑
	// 注意：需要实现具体的邮件发送逻辑
	log.Printf("Sending email alert: [%s] %s", alert.Level, alert.Message)
}

// 发送Webhook告警
func (c *Collector) sendWebhookAlert(alert *Alert) {
	// 实现Webhook调用逻辑
	// 注意：需要实现具体的Webhook调用逻辑
	log.Printf("Sending webhook alert: [%s] %s", alert.Level, alert.Message)
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
	return slices.Clone(c.systemMetricsHistory)
}

// GetBusinessMetricsHistory 获取业务指标历史记录
func (c *Collector) GetBusinessMetricsHistory() []*BusinessMetrics {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()
	return slices.Clone(c.businessMetricsHistory)
}

// GetSystemStats gets system statistics for the given duration
func (c *Collector) GetSystemStats(duration time.Duration) *SystemMetrics {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	now := time.Now()
	var stats SystemMetrics
	var count int

	// Initialize stats timestamp
	stats.Timestamp = now

	for _, m := range c.systemMetricsHistory {
		if now.Sub(m.Timestamp) <= duration {
			// Sum up metrics for averaging
			stats.CPU += m.CPU
			stats.Memory.Used += m.Memory.Used
			stats.Memory.Total += m.Memory.Total
			stats.Disk.Used += m.Disk.Used
			stats.Disk.Total += m.Disk.Total
			stats.Network.RxBytes += m.Network.RxBytes
			stats.Network.TxBytes += m.Network.TxBytes
			stats.Network.RxRate += m.Network.RxRate
			stats.Network.TxRate += m.Network.TxRate
			stats.Load.Load1 += m.Load.Load1
			stats.Load.Load5 += m.Load.Load5
			stats.Load.Load15 += m.Load.Load15
			count++
		}
	}

	if count == 0 {
		return nil
	}

	// Calculate averages
	stats.CPU /= float64(count)
	stats.Memory.Used /= uint64(count)
	stats.Memory.Total /= uint64(count)
	stats.Disk.Used /= uint64(count)
	stats.Disk.Total /= uint64(count)
	stats.Network.RxBytes /= uint64(count)
	stats.Network.TxBytes /= uint64(count)
	stats.Network.RxRate /= float64(count)
	stats.Network.TxRate /= float64(count)
	stats.Load.Load1 /= float64(count)
	stats.Load.Load5 /= float64(count)
	stats.Load.Load15 /= float64(count)

	// Derive usage percentages
	if stats.Memory.Total > 0 {
		stats.Memory.Usage = (float64(stats.Memory.Used) / float64(stats.Memory.Total)) * 100
	}
	if stats.Disk.Total > 0 {
		stats.Disk.Usage = (float64(stats.Disk.Used) / float64(stats.Disk.Total)) * 100
	}

	return &stats
}

// GetBusinessStats 获取业务指标统计数据
func (c *Collector) GetBusinessStats(duration time.Duration) *BusinessMetrics {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	now := time.Now()
	var totalUsers, totalChannels int
	var totalUptime time.Duration
	var totalVoiceQuality float64
	var count int

	for _, m := range c.businessMetricsHistory {
		if now.Sub(m.Timestamp) <= duration {
			totalUsers += m.OnlineUsers
			totalChannels += m.ChannelCount
			totalUptime += m.Uptime
			totalVoiceQuality += m.VoiceQuality
			count++
		}
	}

	if count == 0 {
		return nil
	}

	return &BusinessMetrics{
		Timestamp:    now,
		OnlineUsers:  totalUsers / count,
		ChannelCount: totalChannels / count,
		Uptime:       time.Duration(int64(totalUptime) / int64(count)),
		VoiceQuality: totalVoiceQuality / float64(count),
	}
}

// 保存监控数据到文件
func (c *Collector) SaveToFile(filename string) error {
	c.metricsMutex.RLock()
	defer c.metricsMutex.RUnlock()

	data := struct {
		SystemMetrics   *SystemMetrics
		BusinessMetrics *BusinessMetrics
		Timestamp       time.Time
	}{
		SystemMetrics:   c.lastSystemMetrics,
		BusinessMetrics: c.lastBusinessMetrics,
		Timestamp:       time.Now(),
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("编码数据失败: %v", err)
	}

	return nil
}

// 从文件加载监控数据
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
		if c.tsClient == nil {
			if ts, err := NewTeamSpeakClient(GetConfig().TeamSpeakConfig); err != nil {
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