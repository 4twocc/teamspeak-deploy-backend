package monitor

import (
	"fmt"
	"strings"
	"time"

	configPkg "teamspeak-one-click-deploy/config"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"
)

// CollectSystemMetrics 收集系统指标
func CollectSystemMetrics() (*SystemMetrics, error) {
	// 获取性能配置
	cfg := GetConfig()

	// 如果配置未加载，使用默认值
	var systemConfig configPkg.SystemConfig
	var perfConfig configPkg.PerformanceConfig

	if cfg != nil {
		systemConfig = cfg.Monitoring.System
		perfConfig = cfg.Monitoring.Performance
	} else {
		// 默认配置
		systemConfig = configPkg.SystemConfig{
			MountPoints:       []string{"/"},
			NetworkInterfaces: []string{"eth0"},
		}
		perfConfig = configPkg.PerformanceConfig{
			InterFuncDelay: 50 * time.Millisecond,
		}
	}

	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// 收集 CPU 使用率
	if err := collectCPUUsage(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect CPU usage: %w", err)
	}

	// 短暂延迟，避免资源竞争
	if perfConfig.InterFuncDelay > 0 {
		time.Sleep(perfConfig.InterFuncDelay)
	}

	// 收集内存使用情况
	if err := collectMemoryInfo(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect memory info: %w", err)
	}

	// 短暂延迟，避免资源竞争
	if perfConfig.InterFuncDelay > 0 {
		time.Sleep(perfConfig.InterFuncDelay)
	}

	// 收集磁盘使用情况
	if err := collectDiskUsage(metrics, systemConfig.MountPoints); err != nil {
		return nil, fmt.Errorf("failed to collect disk usage: %w", err)
	}

	// 短暂延迟，避免资源竞争
	if perfConfig.InterFuncDelay > 0 {
		time.Sleep(perfConfig.InterFuncDelay)
	}

	// 收集网络使用情况
	if err := collectNetworkUsage(metrics, systemConfig.NetworkInterfaces); err != nil {
		return nil, fmt.Errorf("failed to collect network usage: %w", err)
	}

	// 短暂延迟，避免资源竞争
	if perfConfig.InterFuncDelay > 0 {
		time.Sleep(perfConfig.InterFuncDelay)
	}

	// 收集系统负载和运行时间
	if err := collectSystemLoadAndUptime(metrics); err != nil {
		return nil, fmt.Errorf("failed to collect system load and uptime: %w", err)
	}

	// 检查告警
	if cfg != nil {
		checkSystemAlerts(metrics, cfg)
	}

	return metrics, nil
}

// collectCPUUsage 收集 CPU 使用率
func collectCPUUsage(metrics *SystemMetrics) error {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("failed to get CPU percent: %w", err)
	}
	if len(percent) > 0 {
		metrics.CPU = percent[0]
	}
	return nil
}

// collectMemoryInfo 收集内存信息
func collectMemoryInfo(metrics *SystemMetrics) error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("failed to get virtual memory info: %w", err)
	}

	metrics.Memory.Total = vm.Total
	metrics.Memory.Used = vm.Used
	metrics.Memory.Available = vm.Available
	metrics.Memory.Usage = vm.UsedPercent

	return nil
}

// collectDiskUsage 收集磁盘使用情况
func collectDiskUsage(metrics *SystemMetrics, mountPoints []string) error {
	if len(mountPoints) == 0 {
		mountPoints = []string{"/"}
	}

	var total, used, free uint64
	for _, mountPoint := range mountPoints {
		usage, err := disk.Usage(mountPoint)
		if err != nil {
			// 忽略单个挂载点错误
			continue
		}
		total += usage.Total
		used += usage.Used
		free += usage.Free
	}

	if total > 0 {
		metrics.Disk.Total = total
		metrics.Disk.Used = used
		metrics.Disk.Free = free
		metrics.Disk.Usage = float64(used) / float64(total) * 100
	}

	return nil
}

// collectNetworkUsage 收集网络使用情况
func collectNetworkUsage(metrics *SystemMetrics, interfaces []string) error {
	stats, err := gnet.IOCounters(true)
	if err != nil {
		return fmt.Errorf("failed to get network IO counters: %w", err)
	}

	var rxBytes, txBytes uint64
	if len(interfaces) == 0 {
		// 如果没有指定接口，汇总所有接口
		for _, stat := range stats {
			rxBytes += stat.BytesRecv
			txBytes += stat.BytesSent
		}
	} else {
		// 只汇总指定的接口
		interfaceMap := make(map[string]bool)
		for _, iface := range interfaces {
			interfaceMap[iface] = true
		}

		for _, stat := range stats {
			if interfaceMap[stat.Name] {
				rxBytes += stat.BytesRecv
				txBytes += stat.BytesSent
			}
		}
	}

	metrics.Network.RxBytes = rxBytes
	metrics.Network.TxBytes = txBytes

	return nil
}

// collectSystemLoadAndUptime 收集系统负载和运行时间
func collectSystemLoadAndUptime(metrics *SystemMetrics) error {
	// 获取系统负载
	loadAvg, err := load.Avg()
	if err != nil {
		// 在某些系统上可能不支持，忽略错误
		loadAvg = &load.AvgStat{}
	}

	metrics.Load.Load1 = loadAvg.Load1
	metrics.Load.Load5 = loadAvg.Load5
	metrics.Load.Load15 = loadAvg.Load15

	// 获取系统运行时间
	info, err := host.Info()
	if err != nil {
		return fmt.Errorf("failed to get host info: %w", err)
	}

	metrics.Uptime = info.Uptime

	return nil
}

// checkSystemAlerts 检查系统告警
func checkSystemAlerts(metrics *SystemMetrics, cfg *configPkg.Config) {
	if cfg == nil {
		return
	}

	thresholds := cfg.Monitoring.Alert.Thresholds

	var alerts []string

	// 检查 CPU 使用率
	if metrics.CPU > thresholds.CPU {
		alerts = append(alerts, fmt.Sprintf("CPU使用率过高: %.2f%%", metrics.CPU))
	}

	// 检查内存使用率
	if metrics.Memory.Usage > thresholds.Memory {
		alerts = append(alerts, fmt.Sprintf("内存使用率过高: %.2f%%", metrics.Memory.Usage))
	}

	// 检查磁盘使用率
	if metrics.Disk.Usage > thresholds.Disk {
		alerts = append(alerts, fmt.Sprintf("磁盘使用率过高: %.2f%%", metrics.Disk.Usage))
	}

	if len(alerts) > 0 {
		metrics.Alert = strings.Join(alerts, "; ")
	}
}

// min 返回两个 float64 中的最小值
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
