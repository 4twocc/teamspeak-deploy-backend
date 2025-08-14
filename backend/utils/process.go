//go:build linux || darwin
// +build linux darwin

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// SetProcessPriority sets the priority of a process on Unix-like systems
// priority should be between -20 (highest priority) and 19 (lowest priority)
func SetProcessPriority(pid int, priority int) error {
	// Convert priority to the correct range for the system
	if priority < -20 || priority > 19 {
		return fmt.Errorf("priority must be between -20 and 19")
	}

	// Use syscall to set the process priority
	// Note: This requires appropriate permissions (usually root for negative priorities)
	err := syscall.Setpriority(syscall.PRIO_PROCESS, pid, priority)
	if err != nil {
		return fmt.Errorf("failed to set process priority: %v", err)
	}

	return nil
}

// SetCPULimit sets CPU usage limit for a process on Unix-like systems
// percentage should be between 1-100
func SetCPULimit(pid int, percentage int) error {
	if percentage < 1 || percentage > 100 {
		return fmt.Errorf("CPU percentage must be between 1 and 100")
	}

	// Use cgroups v2 if available, otherwise fall back to cpulimit
	_, err := exec.LookPath("systemd-cgtop")
	if err == nil {
		// Using systemd cgroups
		cgroupPath := fmt.Sprintf("/sys/fs/cgroup/cpu/teamspeak/instance-%d", pid)
		os.MkdirAll(cgroupPath, 0755)

		// Set CPU quota (in microseconds per second)
		quota := percentage * 1000 // Convert percentage to millicores
		return os.WriteFile(
			fmt.Sprintf("%s/cpu.max", cgroupPath),
			fmt.Appendf(nil, "%d 100000", quota),
			0644,
		)
	}

	// Fall back to cpulimit if available
	_, err = exec.LookPath("cpulimit")
	if err == nil {
		cmd := exec.Command("cpulimit", "-p", fmt.Sprintf("%d", pid), "-l", fmt.Sprintf("%d", percentage))
		return cmd.Start() // Run in background
	}

	return fmt.Errorf("no supported CPU limiting mechanism found (tried cgroups and cpulimit)")
}
