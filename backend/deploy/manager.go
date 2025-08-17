package deploy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// DeploymentManager 部署管理器
type DeploymentManager struct {
	scriptDir string
}

// NewDeploymentManager 创建新的部署管理器
func NewDeploymentManager(scriptDir string) *DeploymentManager {
	return &DeploymentManager{
		scriptDir: scriptDir,
	}
}

// Deploy 执行一键部署
func (dm *DeploymentManager) Deploy() error {
	scriptPath := fmt.Sprintf("%s/one-click.sh", dm.scriptDir)
	return dm.executeScript(scriptPath)
}

// InitEnvironment 初始化环境
func (dm *DeploymentManager) InitEnvironment() error {
	scriptPath := fmt.Sprintf("%s/init-env.sh", dm.scriptDir)
	return dm.executeScript(scriptPath)
}

// Cleanup 清理部署
func (dm *DeploymentManager) Cleanup() error {
	scriptPath := fmt.Sprintf("%s/cleanup.sh", dm.scriptDir)
	return dm.executeScript(scriptPath)
}

// executeScript 执行脚本
func (dm *DeploymentManager) executeScript(scriptPath string) error {
	// 检查脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
	}

	// 执行脚本
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("script execution failed: %v, output: %s", err, string(output))
	}

	return nil
}

// GetContainerStatus 获取容器状态
func (dm *DeploymentManager) GetContainerStatus() (map[string]any, error) {
	// 检查 Docker 是否可用
	if !dm.isDockerAvailable() {
		return map[string]any{
			"docker_available": false,
			"status":           "unavailable",
		}, nil
	}

	// 检查 TeamSpeak 容器状态
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %v", err)
	}

	return map[string]any{
		"docker_available": true,
		"container_info":   output,
	}, nil
}

// isDockerAvailable 检查 Docker 是否可用
func (dm *DeploymentManager) isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}
