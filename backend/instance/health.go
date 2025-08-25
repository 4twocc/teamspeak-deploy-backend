package instance

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"teamspeak-one-click-deploy/logs"
	"teamspeak-one-click-deploy/utils"
	"time"

	"gorm.io/gorm"
)

// 健康检查相关常量
const (
	// 默认健康检查间隔
	defaultHealthCheckInterval = 30 * time.Second
	// 资源检查间隔
	resourceCheckInterval = 60 * time.Second
)

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`    // CPU使用百分比
	MemoryMB      uint64  `json:"memory_mb"`      // 内存使用量(MB)
	MemoryPercent float64 `json:"memory_percent"` // 内存使用百分比
	DiskUsageMB   uint64  `json:"disk_usage_mb"`  // 磁盘使用量(MB)
	NetworkIn     uint64  `json:"network_in"`     // 网络接收字节数
	NetworkOut    uint64  `json:"network_out"`    // 网络发送字节数
	NetworkInMB   uint64  `json:"network_in_mb"`  // 网络接收量(MB)
	NetworkOutMB  uint64  `json:"network_out_mb"` // 网络发送量(MB)
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	MaxCPUPercent   float64 `json:"max_cpu_percent"`    // 最大CPU使用率(百分比)
	MaxMemoryMB     uint64  `json:"max_memory_mb"`      // 最大内存使用量(MB)
	MaxDiskUsageMB  uint64  `json:"max_disk_usage_mb"`  // 最大磁盘使用量(MB)
	MaxNetworkInMB  uint64  `json:"max_network_in_mb"`  // 最大网络接收量(MB)
	MaxNetworkOutMB uint64  `json:"max_network_out_mb"` // 最大网络发送量(MB)
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	IsHealthy     bool           `json:"is_healthy"`
	Message       string         `json:"message"`
	Timestamp     time.Time      `json:"timestamp"`
	ResourceUsage *ResourceUsage `json:"resource_usage,omitempty"`
}

// HealthChecker 健康检查接口
type HealthChecker interface {
	Check(instance *Instance) (*HealthCheckResult, error)
}

// ProcessHealthChecker 进程健康检查
type ProcessHealthChecker struct{}

// Check 检查进程是否健康
func (c *ProcessHealthChecker) Check(instance *Instance) (*HealthCheckResult, error) {
	if instance.ProcessID == 0 {
		return &HealthCheckResult{
			IsHealthy: false,
			Message:   "process not running",
			Timestamp: time.Now(),
		}, nil
	}

	// 检查进程是否存在
	cmd := exec.Command("ps", "-p", strconv.Itoa(int(instance.ProcessID)))
	if err := cmd.Run(); err != nil {
		return &HealthCheckResult{
			IsHealthy: false,
			Message:   fmt.Sprintf("process not found: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	return &HealthCheckResult{
		IsHealthy: true,
		Message:   "process is running",
		Timestamp: time.Now(),
	}, nil
}

// ResourceHealthChecker 资源健康检查
type ResourceHealthChecker struct {
	db           *gorm.DB
	Limits       *ResourceLimits
	AlertManager *AlertManager
}

// Check 检查资源使用情况
func (c *ResourceHealthChecker) Check(instance *Instance) (*HealthCheckResult, error) {
	if instance.ProcessID == 0 {
		return &HealthCheckResult{
			IsHealthy: false,
			Message:   "process not running, cannot check resources",
			Timestamp: time.Now(),
		}, nil
	}

	// 获取进程资源使用情况
	usage, err := getProcessResourceUsage(instance.ProcessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get process resource usage: %w", err)
	}

	// 检查资源使用是否超过限制
	var issues []string

	if c.Limits != nil {
		if usage.CPUPercent > c.Limits.MaxCPUPercent && c.Limits.MaxCPUPercent > 0 {
			issues = append(issues, fmt.Sprintf("CPU使用率 %.2f%% 超过限制 %.2f%%",
				usage.CPUPercent, c.Limits.MaxCPUPercent))
			// 触发CPU超限处理
			c.handleResourceExceeded(instance, "CPU", usage.CPUPercent, c.Limits.MaxCPUPercent)
		}

		if usage.MemoryMB > c.Limits.MaxMemoryMB && c.Limits.MaxMemoryMB > 0 {
			issues = append(issues, fmt.Sprintf("内存使用 %dMB 超过限制 %dMB",
				usage.MemoryMB, c.Limits.MaxMemoryMB))
			// 触发内存超限处理
			c.handleResourceExceeded(instance, "内存", float64(usage.MemoryMB), float64(c.Limits.MaxMemoryMB))
		}

		if usage.DiskUsageMB > c.Limits.MaxDiskUsageMB && c.Limits.MaxDiskUsageMB > 0 {
			issues = append(issues, fmt.Sprintf("磁盘使用 %dMB 超过限制 %dMB",
				usage.DiskUsageMB, c.Limits.MaxDiskUsageMB))
			// 触发磁盘超限处理
			c.handleResourceExceeded(instance, "磁盘", float64(usage.DiskUsageMB), float64(c.Limits.MaxDiskUsageMB))
		}

		if usage.NetworkIn/1024/1024 > c.Limits.MaxNetworkInMB && c.Limits.MaxNetworkInMB > 0 {
			issues = append(issues, fmt.Sprintf("网络入流量 %.2fMB 超过限制 %dMB",
				float64(usage.NetworkIn)/1024/1024, c.Limits.MaxNetworkInMB))
		}

		if usage.NetworkOut/1024/1024 > c.Limits.MaxNetworkOutMB && c.Limits.MaxNetworkOutMB > 0 {
			issues = append(issues, fmt.Sprintf("网络出流量 %.2fMB 超过限制 %dMB",
				float64(usage.NetworkOut)/1024/1024, c.Limits.MaxNetworkOutMB))
		}
	}

	if len(issues) > 0 {
		return &HealthCheckResult{
			IsHealthy:     false,
			Message:       strings.Join(issues, "; "),
			Timestamp:     time.Now(),
			ResourceUsage: usage,
		}, nil
	}

	return &HealthCheckResult{
		IsHealthy:     true,
		Message:       "resource usage is within limits",
		Timestamp:     time.Now(),
		ResourceUsage: usage,
	}, nil
}

// handleResourceExceeded 处理资源超限
func (c *ResourceHealthChecker) handleResourceExceeded(instance *Instance, resourceType string, current, limit float64) {
	// 记录资源超限日志
	logMsg := fmt.Sprintf("%s 使用量 %.2f 超过限制 %.2f", resourceType, current, limit)
	_ = instance.AddLog(c.db, "warning", logMsg)

	// 触发告警
	if c.AlertManager != nil {
		_ = c.AlertManager.Trigger(
			instance.ID,
			AlertLevelWarning,
			AlertTypeResource,
			fmt.Sprintf("%s 使用量超限", resourceType),
			logMsg,
			map[string]any{
				"resource_type": resourceType,
				"current":       current,
				"limit":         limit,
			},
		)
	}

	// 根据资源类型执行相应的处理逻辑
	switch resourceType {
	case "CPU":
		usage := &ResourceUsage{
			CPUPercent: current,
		}
		c.handleCPUOverload(instance, usage)
	case "内存":
		c.handleMemoryOverload(instance, current, limit)
	case "磁盘":
		c.handleDiskOverload(instance, current, limit)
	}
}

// handleCPUOverload 处理CPU过载
func (c *ResourceHealthChecker) handleCPUOverload(instance *Instance, usage *ResourceUsage) {
	// 记录CPU过载日志
	logMsg := fmt.Sprintf("CPU使用率 %.2f%% 超过限制 %.2f%%，正在尝试降低负载...",
		usage.CPUPercent, c.Limits.MaxCPUPercent)

	// 记录日志
	_ = instance.AddLog(c.db, "warning", logMsg)

	// 触发告警
	if c.AlertManager != nil {
		_ = c.AlertManager.Trigger(
			instance.ID,
			AlertLevelWarning,
			AlertTypeResource,
			"CPU 使用率超限",
			logMsg,
			map[string]any{
				"resource": "CPU",
				"current":  usage.CPUPercent,
				"limit":    c.Limits.MaxCPUPercent,
			},
		)
	}

	// 自动恢复机制
	// 1. 降低进程优先级
	if err := utils.SetProcessPriority(int(instance.ProcessID), 10); err != nil {
		_ = instance.AddLog(c.db, "error", fmt.Sprintf("降低进程优先级失败: %v", err))
	}

	time.Sleep(time.Second * 5)
	// 2. 限制CPU使用率
	if usage.CPUPercent > c.Limits.MaxCPUPercent {
		cpuLimit := c.Limits.MaxCPUPercent * 0.9 // 设置为限制的90%
		if err := utils.SetCPULimit(int(instance.ProcessID), int(cpuLimit)); err != nil {
			_ = instance.AddLog(c.db, "error",
				fmt.Sprintf("设置CPU限制失败: %v", err))
		} else {
			_ = instance.AddLog(c.db, "info",
				fmt.Sprintf("已限制CPU使用率不超过 %.2f%%", cpuLimit))
		}
	}
}

// handleMemoryOverload 处理内存过载
func (c *ResourceHealthChecker) handleMemoryOverload(instance *Instance, current, limit float64) {
	// 记录内存过载日志
	logMsg := fmt.Sprintf("内存使用 %.2fMB 超过限制 %.2fMB，正在尝试释放内存...", current, limit)
	_ = instance.AddLog(c.db, "warning", logMsg)

	// 这里可以添加释放内存的逻辑，例如：
	// 1. 清理缓存
	// 2. 重启实例（最后手段）
}

// handleDiskOverload 处理磁盘空间不足
func (c *ResourceHealthChecker) handleDiskOverload(instance *Instance, current, limit float64) {
	// 记录磁盘空间不足日志
	logMsg := fmt.Sprintf("磁盘使用 %.2fMB 超过限制 %.2fMB，正在尝试清理空间...", current, limit)
	_ = instance.AddLog(c.db, "warning", logMsg)

	// 这里可以添加清理磁盘空间的逻辑，例如：
	// 1. 清理日志文件
	// 2. 清理临时文件
	// 3. 如果启用了备份，清理旧的备份文件
}

// 获取进程资源使用情况的函数变量，便于测试时mock
var getProcessResourceUsage = func(pid int32) (*ResourceUsage, error) {
	usage := &ResourceUsage{}

	// 获取CPU和内存使用情况
	// 使用/proc/[pid]/stat文件 (Linux系统)
	pidStr := strconv.Itoa(int(pid))

	// 读取/proc/[pid]/stat文件
	statPath := filepath.Join("/proc", pidStr, "stat")
	statContent, err := os.ReadFile(statPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read process stat file: %w", err)
	}

	// 解析stat文件内容
	fields := strings.Fields(string(statContent))
	if len(fields) < 24 {
		return nil, fmt.Errorf("unexpected stat file format")
	}

	// 获取系统启动时间
	statFilePath := "/proc/stat"
	statFile, err := os.Open(statFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open stat file: %w", err)
	}
	defer statFile.Close()

	scanner := bufio.NewScanner(statFile)
	var btime int64
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "btime") {
			fmt.Sscanf(line, "btime %d", &btime)
			break
		}
	}

	// 解析进程启动时间
	_, err = strconv.ParseInt(fields[21], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse process start time: %w", err)
	}

	// 计算进程运行时间(时钟滴答数)
	uptimePath := "/proc/uptime"
	uptimeContent, err := os.ReadFile(uptimePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read uptime file: %w", err)
	}

	var uptimeSec float64
	fmt.Sscanf(string(uptimeContent), "%f", &uptimeSec)

	// 获取页面大小
	pageSize := os.Getpagesize()

	// 计算内存使用量(MB)
	rss, err := strconv.ParseUint(fields[23], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS value: %w", err)
	}
	usage.MemoryMB = rss * uint64(pageSize) / 1024 / 1024

	// 获取更准确的内存信息
	statusPath := filepath.Join("/proc", pidStr, "status")
	statusFile, err := os.Open(statusPath)
	if err == nil {
		defer statusFile.Close()
		statusScanner := bufio.NewScanner(statusFile)
		for statusScanner.Scan() {
			line := statusScanner.Text()
			if strings.HasPrefix(line, "VmRSS:") {
				var vmRSS uint64
				fmt.Sscanf(line, "VmRSS: %d kB", &vmRSS)
				usage.MemoryMB = vmRSS / 1024
				break
			}
		}
	}

	// 获取CPU使用率
	// 读取两次/proc/[pid]/stat来计算CPU使用率
	prevUTime, err := strconv.ParseInt(fields[13], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse utime: %w", err)
	}

	prevSTime, err := strconv.ParseInt(fields[14], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stime: %w", err)
	}

	// 等待一段时间再次读取
	time.Sleep(100 * time.Millisecond)

	statContent2, err := os.ReadFile(statPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read process stat file second time: %w", err)
	}

	fields2 := strings.Fields(string(statContent2))
	if len(fields2) < 24 {
		return nil, fmt.Errorf("unexpected stat file format on second read")
	}

	currUTime, err := strconv.ParseInt(fields2[13], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse utime on second read: %w", err)
	}

	currSTime, err := strconv.ParseInt(fields2[14], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stime on second read: %w", err)
	}

	// 计算CPU使用率
	totalTime := float64((currUTime+currSTime)-(prevUTime+prevSTime)) / 100.0
	usage.CPUPercent = totalTime * 100

	// 获取磁盘使用情况
	// 这里简化处理，使用实例目录大小作为磁盘使用量
	instanceDir := filepath.Join("/var/lib/teamspeak", pidStr)
	if info, err := os.Stat(instanceDir); err == nil && info.IsDir() {
		var size int64
		filepath.Walk(instanceDir, func(_ string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				size += info.Size()
			}
			return nil
		})
		usage.DiskUsageMB = uint64(size) / 1024 / 1024
	}

	// 获取网络使用情况
	// 这里简化处理，暂时设置为固定值
	usage.NetworkIn = 0
	usage.NetworkOut = 0
	usage.NetworkInMB = 0
	usage.NetworkOutMB = 0

	return usage, nil
}

// HealthMonitor 健康监控器
type HealthMonitor struct {
	db                *gorm.DB
	checkers          []HealthChecker
	checkInterval     time.Duration
	instances         map[string]time.Time
	instancesMutex    sync.RWMutex
	stopChan          chan struct{}
	resourceCheckChan chan string
	logService        logs.LogService
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(db *gorm.DB) *HealthMonitor {
	hm := &HealthMonitor{
		db:                db,
		checkers:          make([]HealthChecker, 0),
		checkInterval:     defaultHealthCheckInterval,
		instances:         make(map[string]time.Time),
		stopChan:          make(chan struct{}),
		resourceCheckChan: make(chan string, 100),
	}

	// 尝试初始化日志服务
	if ls, err := logs.NewLogService(nil, logs.LogConfig{
		Level:      "info",
		EnableFile: true,
		FilePath:   "logs/health.log",
	}); err == nil {
		hm.logService = ls
	}

	return hm
}

// AddChecker 添加健康检查器
func (m *HealthMonitor) AddChecker(checker HealthChecker) {
	m.checkers = append(m.checkers, checker)
}

// Start 启动健康监控
func (m *HealthMonitor) Start() {
	if m.logService != nil {
		m.logService.Info("health", "Starting health monitor", logs.LogField{Key: "component", Value: "HealthMonitor"})
	} else {
		log.Println("Starting health monitor...")
	}

	// 启动健康检查循环
	go m.healthCheckLoop()

	// 启动资源检查循环
	go m.resourceCheckLoop()
}

// Stop 停止健康监控
func (m *HealthMonitor) Stop() {
	close(m.stopChan)
}

// AddInstance 添加要监控的实例
func (m *HealthMonitor) AddInstance(instanceID string) {
	m.instancesMutex.Lock()
	defer m.instancesMutex.Unlock()
	m.instances[instanceID] = time.Now()
}

// RemoveInstance 移除监控的实例
func (m *HealthMonitor) RemoveInstance(instanceID string) {
	m.instancesMutex.Lock()
	defer m.instancesMutex.Unlock()
	delete(m.instances, instanceID)
}

// healthCheckLoop 健康检查循环
func (m *HealthMonitor) healthCheckLoop() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllInstances()
		case <-m.stopChan:
			return
		}
	}
}

// resourceCheckLoop 资源检查循环
func (m *HealthMonitor) resourceCheckLoop() {
	ticker := time.NewTicker(resourceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkAllResources()
		case instanceID := <-m.resourceCheckChan:
			m.checkInstanceResources(instanceID)
		case <-m.stopChan:
			return
		}
	}
}

// checkAllInstances 检查所有实例的健康状态
func (m *HealthMonitor) checkAllInstances() {
	m.instancesMutex.RLock()
	defer m.instancesMutex.RUnlock()

	for instanceID := range m.instances {
		m.checkInstance(instanceID)
	}
}

// checkInstance 检查单个实例的健康状态
func (m *HealthMonitor) checkInstance(instanceID string) {
	// 从数据库获取实例信息
	var instance Instance
	if err := m.db.First(&instance, "id = ?", instanceID).Error; err != nil {
		if m.logService != nil {
			m.logService.Error("health", "Failed to get instance", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "error", Value: err})
		} else {
			log.Printf("Failed to get instance %s: %v", instanceID, err)
		}
		return
	}

	// 执行健康检查
	for _, checker := range m.checkers {
		result, err := checker.Check(&instance)
		if err != nil {
			if m.logService != nil {
				m.logService.Error("health", "Health check failed", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "error", Value: err})
			} else {
				log.Printf("Health check failed for instance %s: %v", instanceID, err)
			}
			continue
		}

		// 记录健康检查结果
		if m.logService != nil {
			m.logService.Info("health", "Health check completed", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "message", Value: result.Message})
		} else {
			log.Printf("Health check for instance %s: %s", instanceID, result.Message)
		}

		// 如果检查不健康，触发告警
		if !result.IsHealthy {
			m.triggerAlert(&instance, result)
		}
	}
}

// checkAllResources 检查所有实例的资源使用情况
func (m *HealthMonitor) checkAllResources() {
	m.instancesMutex.RLock()
	defer m.instancesMutex.RUnlock()

	for instanceID := range m.instances {
		m.resourceCheckChan <- instanceID
	}
}

// checkInstanceResources 检查单个实例的资源使用情况
func (m *HealthMonitor) checkInstanceResources(instanceID string) {
	// 从数据库获取实例信息
	var instance Instance
	if err := m.db.First(&instance, "id = ?", instanceID).Error; err != nil {
		if m.logService != nil {
			m.logService.Error("health", "Failed to get instance for resource check", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "error", Value: err})
		} else {
			log.Printf("Failed to get instance %s for resource check: %v", instanceID, err)
		}
		return
	}

	// 执行资源检查
	for _, checker := range m.checkers {
		if resourceChecker, ok := checker.(*ResourceHealthChecker); ok {
			result, err := resourceChecker.Check(&instance)
			if err != nil {
				if m.logService != nil {
					m.logService.Error("health", "Resource check failed", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "error", Value: err})
				} else {
					log.Printf("Resource check failed for instance %s: %v", instanceID, err)
				}
				continue
			}

			// 记录资源检查结果
			if m.logService != nil {
				m.logService.Info("health", "Resource check completed", logs.LogField{Key: "instance_id", Value: instanceID}, logs.LogField{Key: "message", Value: result.Message})
			} else {
				log.Printf("Resource check for instance %s: %s", instanceID, result.Message)
			}

			// 如果资源使用超过限制，触发告警
			if !result.IsHealthy {
				m.triggerAlert(&instance, result)
			}
		}
	}
}

// triggerAlert 触发告警
func (m *HealthMonitor) triggerAlert(instance *Instance, result *HealthCheckResult) {
	// 记录告警日志
	if m.logService != nil {
		m.logService.Error("health", "ALERT: Instance health check failed", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "message", Value: result.Message})
	} else {
		log.Printf("ALERT: Instance %s - %s", instance.ID, result.Message)
	}

	// 更新实例状态为错误
	if instance.Status != StatusError {
		instance.Status = StatusError
		if err := m.db.Save(instance).Error; err != nil {
			if m.logService != nil {
				m.logService.Error("health", "Failed to update instance status to error", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "error", Value: err})
			} else {
				log.Printf("Failed to update instance status to error: %v", err)
			}
		}
	}

	// 添加错误日志
	logMsg := fmt.Sprintf("Health check failed: %s", result.Message)
	if result.ResourceUsage != nil {
		logMsg += fmt.Sprintf(" (CPU: %.2f%%, Memory: %dMB, Disk: %dMB)",
			result.ResourceUsage.CPUPercent,
			result.ResourceUsage.MemoryMB,
			result.ResourceUsage.DiskUsageMB)
	}

	// 记录到实例日志
	if err := instance.AddLog(m.db, "error", logMsg); err != nil {
		if m.logService != nil {
			m.logService.Error("health", "Failed to add error log", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "error", Value: err})
		} else {
			log.Printf("Failed to add error log: %v", err)
		}
	}

	// TODO: 发送告警通知（邮件、Webhook等）
}

// AutoRestartConfig 自动重启配置
type AutoRestartConfig struct {
	Enabled       bool          `json:"enabled"`        // 是否启用自动重启
	MaxRestarts   int           `json:"max_restarts"`   // 最大重启次数
	RestartDelay  time.Duration `json:"restart_delay"`  // 重启延迟
	RestartWindow time.Duration `json:"restart_window"` // 重启窗口（用于计算重启频率）
	ResetAfter    time.Duration `json:"reset_after"`    // 重置计数器的时间窗口
	BackoffFactor float64       `json:"backoff_factor"` // 退避因子
	MaxBackoff    time.Duration `json:"max_backoff"`    // 最大退避时间
}

// DefaultAutoRestartConfig 返回默认的自动重启配置
func DefaultAutoRestartConfig() *AutoRestartConfig {
	return &AutoRestartConfig{
		Enabled:       true,
		MaxRestarts:   5,
		RestartDelay:  5 * time.Second,
		RestartWindow: time.Hour,
		ResetAfter:    24 * time.Hour,
		BackoffFactor: 2.0,
		MaxBackoff:    5 * time.Minute,
	}
}

// RestartManager 重启管理器
type RestartManager struct {
	db         *gorm.DB
	instances  map[string]*restartTracker
	instanceMu sync.RWMutex
}

// restartTracker 跟踪实例的重启状态
type restartTracker struct {
	count       int
	lastReset   time.Time
	lastRestart time.Time
}

// NewRestartManager 创建重启管理器
func NewRestartManager(db *gorm.DB) *RestartManager {
	return &RestartManager{
		db:        db,
		instances: make(map[string]*restartTracker),
	}
}

// CanRestart 检查是否可以重启实例
func (m *RestartManager) CanRestart(instanceID string, config *AutoRestartConfig) (bool, time.Duration) {
	if !config.Enabled {
		return false, 0
	}

	m.instanceMu.Lock()
	defer m.instanceMu.Unlock()

	now := time.Now()
	tracker, exists := m.instances[instanceID]

	// 如果跟踪器不存在或已过重置时间，则重置计数器
	if !exists || now.Sub(tracker.lastReset) >= config.ResetAfter {
		tracker = &restartTracker{
			count:     0,
			lastReset: now,
		}
		m.instances[instanceID] = tracker
	}

	// 如果重启次数超过限制，返回false
	if tracker.count >= config.MaxRestarts {
		return false, 0
	}

	// 计算退避时间
	backoff := config.RestartDelay
	if tracker.count > 0 {
		backoff = time.Duration(float64(backoff) * config.BackoffFactor)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	// 如果距离上次重启时间不足退避时间，返回false
	if !tracker.lastRestart.IsZero() && now.Sub(tracker.lastRestart) < backoff {
		return false, backoff - now.Sub(tracker.lastRestart)
	}

	// 更新重启计数器和最后重启时间
	tracker.count++
	tracker.lastRestart = now

	return true, 0
}

// Reset 重置实例的重启计数器
func (m *RestartManager) Reset(instanceID string) {
	m.instanceMu.Lock()
	defer m.instanceMu.Unlock()

	if tracker, exists := m.instances[instanceID]; exists {
		tracker.count = 0
		tracker.lastReset = time.Now()
	}
}
