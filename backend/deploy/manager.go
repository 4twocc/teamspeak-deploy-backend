package deploy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DeploymentManager 部署管理器
type DeploymentManager struct {
	scriptDir string
	timeout   time.Duration
}

// NewDeploymentManager 创建新的部署管理器
func NewDeploymentManager(scriptDir string) *DeploymentManager {
	return &DeploymentManager{
		scriptDir: scriptDir,
		timeout:   10 * time.Minute, // 默认超时时间
	}
}

// SetTimeout 设置超时时间
func (dm *DeploymentManager) SetTimeout(timeout time.Duration) {
	dm.timeout = timeout
}

// IsDockerAvailable 检查Docker是否可用
func (dm *DeploymentManager) IsDockerAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version")
	err := cmd.Run()
	return err == nil
}

// IsDockerComposeAvailable 检查Docker Compose是否可用
func (dm *DeploymentManager) IsDockerComposeAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "compose", "version")
	err := cmd.Run()
	return err == nil
}

// GetContainerStatus 获取容器状态
func (dm *DeploymentManager) GetContainerStatus() (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 docker compose ps 命令获取容器状态
	cmd := exec.CommandContext(ctx, "docker", "compose", "ps", "--format", "table {{.Names}}\\t{{.Status}}\\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container status: %w", err)
	}

	return map[string]any{
		"container_info":   output,
		"docker_available": dm.IsDockerAvailable(),
		"compose_available": dm.IsDockerComposeAvailable(),
	}, nil
}

// GetDeploymentStatus 获取部署状态
func (dm *DeploymentManager) GetDeploymentStatus() (map[string]any, error) {
	return map[string]any{
		"docker_available": dm.IsDockerAvailable(),
		"compose_available": dm.IsDockerComposeAvailable(),
		"script_dir":       dm.scriptDir,
	}, nil
}

// ExecuteScript 执行部署脚本
func (dm *DeploymentManager) ExecuteScript(scriptName string) error {
	// 构建脚本路径
	scriptPath := fmt.Sprintf("%s/%s", dm.scriptDir, scriptName)

	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script %s not found", scriptPath)
	}

	// 设置执行上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), dm.timeout)
	defer cancel()

	// 创建命令
	cmd := exec.CommandContext(ctx, "bash", scriptPath)

	// 设置工作目录为脚本目录
	cmd.Dir = dm.scriptDir

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute script %s: %w\nOutput: %s", scriptName, err, string(output))
	}

	// 记录成功执行日志
	log.Printf("Successfully executed script %s", scriptName)
	log.Printf("Script output: %s", strings.TrimSpace(string(output)))

	return nil
}

func (dm *DeploymentManager) Deploy() error {
	// 检查Docker和Docker Compose是否可用
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}
	
	if !dm.IsDockerComposeAvailable() {
		return fmt.Errorf("docker compose is not available")
	}

	// 构建脚本路径
	scriptPath := fmt.Sprintf("%s/deploy.sh", dm.scriptDir)
	
	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script %s not found", scriptPath)
	}

	// 设置执行上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), dm.timeout)
	defer cancel()

	// 创建命令，执行deploy.sh脚本的deploy命令
	cmd := exec.CommandContext(ctx, "bash", scriptPath, "deploy")

	// 设置工作目录为脚本目录
	cmd.Dir = dm.scriptDir

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute deploy script: %w\nOutput: %s", err, string(output))
	}

	// 记录成功执行日志
	log.Printf("Successfully executed deploy script")
	log.Printf("Script output: %s", strings.TrimSpace(string(output)))

	return nil
}

// InitEnvironment 初始化环境
func (dm *DeploymentManager) InitEnvironment() error {
	// 检查Docker和Docker Compose是否都已可用
	if dm.IsDockerAvailable() && dm.IsDockerComposeAvailable() {
		log.Printf("Docker and Docker Compose are already available")
		return nil
	}

	// 构建脚本路径
	scriptPath := fmt.Sprintf("%s/init-env.sh", dm.scriptDir)
	
	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script %s not found", scriptPath)
	}

	// 设置执行上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), dm.timeout)
	defer cancel()

	// 创建命令，执行init-env.sh脚本
	cmd := exec.CommandContext(ctx, "bash", scriptPath)

	// 设置工作目录为脚本目录
	cmd.Dir = dm.scriptDir

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute init-env script: %w\nOutput: %s", err, string(output))
	}

	// 记录成功执行日志
	log.Printf("Successfully executed init-env script")
	log.Printf("Script output: %s", strings.TrimSpace(string(output)))

	// 检查Docker和Docker Compose现在是否可用
	if !dm.IsDockerAvailable() || !dm.IsDockerComposeAvailable() {
		return fmt.Errorf("failed to install Docker or Docker Compose")
	}

	return nil
}

// Cleanup 清理部署
func (dm *DeploymentManager) Cleanup() error {
	// 检查Docker是否可用
	if !dm.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	// 构建脚本路径
	scriptPath := fmt.Sprintf("%s/cleanup.sh", dm.scriptDir)
	
	// 检查脚本文件是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script %s not found", scriptPath)
	}

	// 设置执行上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), dm.timeout)
	defer cancel()

	// 创建命令，执行cleanup.sh脚本
	cmd := exec.CommandContext(ctx, "bash", scriptPath)

	// 设置工作目录为脚本目录
	cmd.Dir = dm.scriptDir

	// 捕获输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute cleanup script: %w\nOutput: %s", err, string(output))
	}

	// 记录成功执行日志
	log.Printf("Successfully executed cleanup script")
	log.Printf("Script output: %s", strings.TrimSpace(string(output)))

	return nil
}
