// backend/monitor/system_collector_helpers.go
package monitor

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// getCPUUsage gets the current CPU usage percentage
func getCPUUsage() (float64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxCPUUsage()
	case "darwin":
		return getDarwinCPUUsage()
	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// getLinuxCPUUsage gets CPU usage percentage on Linux systems
func getLinuxCPUUsage() (float64, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, fmt.Errorf("failed to open /proc/stat: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, fmt.Errorf("failed to read CPU stats")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) < 8 {
		return 0, fmt.Errorf("invalid CPU stats format")
	}

	var total, idle uint64
	for i, field := range fields[1:] {
		val, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid CPU stat value: %v", err)
		}
		total += val
		if i == 3 { // idle time is the 4th field (0-based index 3)
			idle = val
		}
	}

	if total == 0 {
		return 0, nil
	}

	return 100.0 * (1 - float64(idle)/float64(total)), nil
}

// getDarwinCPUUsage gets CPU usage percentage on macOS systems
func getDarwinCPUUsage() (float64, error) {
	// Using sysctl to get CPU usage on macOS
	cmd := exec.Command("sysctl", "-n", "kern.cp_time")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU usage: %v", err)
	}

	fields := strings.Fields(string(output))
	if len(fields) < 5 {
		return 0, fmt.Errorf("invalid CPU stats format")
	}

	var total, idle uint64
	for i, field := range fields {
		val, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid CPU stat value: %v", err)
		}
		total += val
		if i == 3 { // idle time is the 4th field (0-based index 3)
			idle = val
		}
	}

	if total == 0 {
		return 0, nil
	}

	return 100.0 * (1 - float64(idle)/float64(total)), nil
}

// getMemoryInfo gets memory information
func getMemoryInfo(metrics *SystemMetrics) error {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	memInfo := make(map[string]uint64)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		memInfo[key] = value * 1024 // Convert from KB to bytes
	}

	metrics.Memory.Total = memInfo["MemTotal"]
	metrics.Memory.Available = memInfo["MemAvailable"]
	metrics.Memory.Used = metrics.Memory.Total - memInfo["MemFree"] - memInfo["Buffers"] - memInfo["Cached"]
	metrics.Memory.Usage = 100 * float64(metrics.Memory.Used) / float64(metrics.Memory.Total)

	return nil
}

// getDiskInfo gets disk usage information
func getDiskInfo(metrics *SystemMetrics) error {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return err
	}

	// Calculate disk space in bytes
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	metrics.Disk.Total = total
	metrics.Disk.Used = used
	metrics.Disk.Free = free
	metrics.Disk.Usage = 100 * float64(used) / float64(total)

	return nil
}

// getNetworkInfo gets network statistics
func getNetworkInfo(metrics *SystemMetrics) error {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return err
	}
	defer file.Close()

	var rx, tx uint64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 10 || !strings.Contains(parts[0], ":") {
			continue
		}

		// Skip loopback interface
		if strings.Trim(parts[0], ":") == "lo" {
			continue
		}

		// Parse receive and transmit bytes
		r, _ := strconv.ParseUint(parts[1], 10, 64)
		t, _ := strconv.ParseUint(parts[9], 10, 64)
		rx += r
		tx += t
	}

	metrics.Network.RxBytes = rx
	metrics.Network.TxBytes = tx

	return nil
}

// getUptime gets system uptime in seconds
func getUptime() (uint64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid uptime format")
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	return uint64(uptime), nil
}

// getLoadAvg gets system load averages
func getLoadAvg(metrics *SystemMetrics) error {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return fmt.Errorf("invalid loadavg format")
	}

	loads := make([]float64, 3)
	for i := range 3 {
		loads[i], _ = strconv.ParseFloat(fields[i], 64)
	}

	metrics.Load.Load1 = loads[0]
	metrics.Load.Load5 = loads[1]
	metrics.Load.Load15 = loads[2]

	return nil
}
