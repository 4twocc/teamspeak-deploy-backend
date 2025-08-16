package monitor

import (
	"fmt"
	"math"
	"strings"
	"time"

	"sync"

	"slices"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
)

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

// CPU 采样缓存（用于非阻塞差分计算）
var (
	prevCPUTimes cpu.TimesStat
	hasPrevCPU   bool
	cpuMutex     sync.Mutex
)

// CollectSystemMetrics 收集系统指标
func CollectSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// 并行收集各项指标以提高效率
	var wg sync.WaitGroup
	var cpuErr, memErr, diskErr, netErr, uptimeErr, loadErr error

	wg.Add(6)
	go func() {
		defer wg.Done()
		cpuErr = collectCPUUsage(metrics)
	}()
	
	go func() {
		defer wg.Done()
		memErr = collectMemoryInfo(metrics)
	}()
	
	go func() {
		defer wg.Done()
		diskErr = collectSimpleDiskUsage(metrics)
	}()
	
	go func() {
		defer wg.Done()
		netErr = collectSimpleNetworkStats(metrics)
	}()
	
	go func() {
		defer wg.Done()
		uptimeErr = collectUptime(metrics)
	}()
	
	go func() {
		defer wg.Done()
		// 降低频率采集负载信息
		if time.Now().Unix()%5 == 0 {
			loadErr = collectLoadAvg(metrics)
		}
	}()

	wg.Wait()

	// 处理错误
	if cpuErr != nil {
		return nil, fmt.Errorf("failed to collect CPU usage: %w", cpuErr)
	}
	if memErr != nil {
		return nil, fmt.Errorf("failed to collect memory info: %w", memErr)
	}
	if diskErr != nil {
		return nil, fmt.Errorf("failed to collect disk usage: %w", diskErr)
	}
	if netErr != nil {
		return nil, fmt.Errorf("failed to collect network stats: %w", netErr)
	}
	if uptimeErr != nil {
		return nil, fmt.Errorf("failed to collect uptime: %w", uptimeErr)
	}
	if loadErr != nil {
		return nil, fmt.Errorf("failed to collect load average: %w", loadErr)
	}

	// 检查告警
	checkSystemAlerts(metrics)

	return metrics, nil
}

// collectCPUUsage 收集 CPU 使用率（跨平台，非阻塞差分法）
func collectCPUUsage(metrics *SystemMetrics) error {
	cpuMutex.Lock()
	defer cpuMutex.Unlock()

	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)

	times, err := cpu.Times(false)
	if err != nil || len(times) == 0 {
		// 回退：使用短间隔 Percent（仍为非阻塞，基于上次 Times 缓存计算）
		p, perr := cpu.Percent(0, false)
		if perr != nil || len(p) == 0 {
			if err != nil {
				return err
			}
			return fmt.Errorf("cpu times unavailable")
		}
		metrics.CPU = math.Round(p[0]*100) / 100
		return nil
	}

	curr := times[0]
	if !hasPrevCPU {
		prevCPUTimes = curr
		hasPrevCPU = true
		// 首次无法计算差分，尝试即时百分比回退
		p, _ := cpu.Percent(0, false)
		if len(p) > 0 {
			metrics.CPU = math.Round(p[0]*100) / 100
		}
		return nil
	}

	deltaTotal := (curr.User - prevCPUTimes.User) + (curr.Nice - prevCPUTimes.Nice) + (curr.System - prevCPUTimes.System) + (curr.Idle - prevCPUTimes.Idle) + (curr.Iowait - prevCPUTimes.Iowait) + (curr.Irq - prevCPUTimes.Irq) + (curr.Softirq - prevCPUTimes.Softirq) + (curr.Steal - prevCPUTimes.Steal) + (curr.Guest - prevCPUTimes.Guest) + (curr.GuestNice - prevCPUTimes.GuestNice)
	deltaIdle := (curr.Idle - prevCPUTimes.Idle) + (curr.Iowait - prevCPUTimes.Iowait)
	prevCPUTimes = curr

	if deltaTotal <= 0 {
		metrics.CPU = 0
		return nil
	}
	usage := 100.0 * (1.0 - (deltaIdle / deltaTotal))
	metrics.CPU = math.Round(usage*100) / 100
	return nil
}

// collectMemoryInfo 收集内存信息（跨平台）
func collectMemoryInfo(metrics *SystemMetrics) error {
	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)
	
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	metrics.Memory.Total = vm.Total
	metrics.Memory.Available = vm.Available
	metrics.Memory.Used = vm.Used
	metrics.Memory.Usage = vm.UsedPercent
	return nil
}

// collectSimpleDiskUsage 简化版磁盘使用情况收集（只检查根分区）
func collectSimpleDiskUsage(metrics *SystemMetrics) error {
	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)
	
	// 只检查根分区，避免遍历所有分区
	u, err := disk.Usage("/")
	if err != nil {
		return err
	}
	
	metrics.Disk.Total = u.Total
	metrics.Disk.Used = u.Used
	metrics.Disk.Free = u.Free
	if u.Total > 0 {
		metrics.Disk.Usage = (float64(u.Used) / float64(u.Total)) * 100
	} else {
		metrics.Disk.Usage = 0
	}
	return nil
}

// 判断分区是否为可聚合的常规文件系统
func isUsableFS(fstype string, opts []string, mount string) bool {
	fs := strings.ToLower(fstype)
	// 常见需排除的虚拟/内核伪文件系统
	excludes := []string{"tmpfs", "devtmpfs", "devfs", "proc", "sysfs", "cgroup", "cgroup2", "overlay", "aufs", "squashfs", "ramfs", "zfs", "autofs", "debugfs", "tracefs", "configfs", "selinuxfs", "fusectl"}
	if slices.Contains(excludes, fs) {
		return false
	}
	// Windows: 仅聚合固定磁盘
	optstr := strings.ToLower(strings.Join(opts, ","))
	if strings.Contains(optstr, "cdrom") || strings.Contains(optstr, "removable") {
		return false
	}
	// 排除只读内核挂载或空挂载点
	if mount == "" || mount == "/proc" || mount == "/sys" || mount == "/dev" {
		return false
	}
	return true
}

// collectSimpleNetworkStats 简化版网络统计信息收集（只收集主要接口）
func collectSimpleNetworkStats(metrics *SystemMetrics) error {
	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)
	
	counters, err := gnet.IOCounters(true)
	if err != nil {
		return err
	}
	
	var rx, tx uint64
	for _, c := range counters {
		name := strings.ToLower(c.Name)
		// 只收集主要网络接口，跳过回环和其他虚拟接口
		if strings.HasPrefix(name, "lo") || strings.HasPrefix(name, "docker") || 
		   strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth") {
			continue
		}
		rx += c.BytesRecv
		tx += c.BytesSent
	}

	networkMutex.Lock()
	defer networkMutex.Unlock()

	now := time.Now()
	if !lastNetwork.time.IsZero() {
		duration := now.Sub(lastNetwork.time).Seconds()
		if duration > 0 {
			metrics.Network.RxRate = float64(rx-lastNetwork.rxBytes) / duration
			metrics.Network.TxRate = float64(tx-lastNetwork.txBytes) / duration
		}
	}

	metrics.Network.RxBytes = rx
	metrics.Network.TxBytes = tx
	lastNetwork.rxBytes = rx
	lastNetwork.txBytes = tx
	lastNetwork.time = now
	return nil
}

// collectUptime 收集系统运行时间（跨平台）
func collectUptime(metrics *SystemMetrics) error {
	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)
	
	uptime, err := host.Uptime()
	if err != nil {
		return err
	}
	metrics.Uptime = uptime
	return nil
}

// collectLoadAvg 收集系统负载（跨平台，若不支持则回退为 0）
func collectLoadAvg(metrics *SystemMetrics) error {
	// 添加延迟以减少系统负载
	time.Sleep(20 * time.Millisecond)
	
	avg, err := load.Avg()
	if err != nil {
		// 某些平台不支持，忽略错误
		metrics.Load.Load1 = 0
		metrics.Load.Load5 = 0
		metrics.Load.Load15 = 0
		return nil
	}
	metrics.Load.Load1 = avg.Load1
	metrics.Load.Load5 = avg.Load5
	metrics.Load.Load15 = avg.Load15
	return nil
}

// checkSystemAlerts 检查系统指标告警
func checkSystemAlerts(metrics *SystemMetrics) {
	if metrics == nil {
		return
	}

	config := GetConfig()
	var alerts []string

	// CPU 使用率告警
	if metrics.CPU > config.AlertConfig.Thresholds.CPU {
		alerts = append(alerts, fmt.Sprintf("CPU 使用率过高: %.2f%%", metrics.CPU))
	}

	// 内存使用率告警
	if metrics.Memory.Usage > config.AlertConfig.Thresholds.Memory {
		alerts = append(alerts, fmt.Sprintf("内存使用率过高: %.2f%%", metrics.Memory.Usage))
	}

	// 磁盘使用率告警
	if metrics.Disk.Usage > config.AlertConfig.Thresholds.Disk {
		alerts = append(alerts, fmt.Sprintf("磁盘使用率过高: %.2f%%", metrics.Disk.Usage))
	}

	// 如果有告警，设置告警信息
	if len(alerts) > 0 {
		metrics.Alert = strings.Join(alerts, "; ")
	}
}

// checkBusinessAlerts 检查业务指标告警
func checkBusinessAlerts(metrics *BusinessMetrics) {
	if metrics == nil {
		return
	}

	config := GetConfig()
	var alerts []string

	// 语音质量告警
	if metrics.VoiceQuality < config.AlertConfig.Thresholds.VoiceQuality {
		alerts = append(alerts, fmt.Sprintf("语音质量下降: %.2f", metrics.VoiceQuality))
	}

	// 在线用户数为零告警
	if metrics.OnlineUsers == 0 {
		alerts = append(alerts, "当前没有在线用户")
	}

	// 如果有告警，设置告警信息
	if len(alerts) > 0 {
		metrics.Alert = strings.Join(alerts, "; ")
	}
}