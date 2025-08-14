package deploy

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

type deployStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	currentStatus = deployStatus{
		Status:  "idle",
		Message: "Ready to deploy",
	}
	statusMutex   sync.RWMutex
	deployManager *DeploymentManager
)

// Initialize 初始化部署模块
func Initialize(cfg *config.Config) error {
	deployConfig, err := LoadConfigFromAppConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to load deployment config: %v", err)
	}

	deployManager = NewDeploymentManager(deployConfig.ScriptDir)
	return nil
}

func RegisterRoutes(router *gin.Engine) {
	router.POST(api.DeployStartPath, startDeployHandler)
	router.GET(api.DeployStatusPath, deployStatusHandler)
	router.POST(api.DeployResetPath, resetDeployStatusHandler)
	router.GET(api.DeployContainerStatusPath, containerStatusHandler)
	router.POST(api.DeployInitEnvPath, initEnvironmentHandler)
	router.POST(api.DeployCleanupPath, cleanupHandler)
}

func startDeployHandler(c *gin.Context) {
	// 检查是否已经在部署中
	statusMutex.RLock()
	if currentStatus.Status == "running" {
		statusMutex.RUnlock()
		utils.FailGin(c, http.StatusConflict, utils.ErrDeployInProgress, utils.ErrorMessage(utils.ErrDeployInProgress))
		return
	}
	statusMutex.RUnlock()

	// 更新状态为运行中
	statusMutex.Lock()
	currentStatus.Status = "running"
	currentStatus.Message = "Deployment started"
	statusMutex.Unlock()

	// 异步执行部署
	go func() {
		err := runDeployScript()
		statusMutex.Lock()
		if err != nil {
			currentStatus.Status = "error"
			currentStatus.Message = fmt.Sprintf("Deployment failed: %v", err)
		} else {
			currentStatus.Status = "completed"
			currentStatus.Message = "Deployment completed successfully"
		}
		statusMutex.Unlock()
	}()

	utils.OKGin(c, map[string]string{"message": "Deployment started"})
}

func deployStatusHandler(c *gin.Context) {
	statusMutex.RLock()
	status := currentStatus
	statusMutex.RUnlock()

	utils.OKGin(c, status)
}

func resetDeployStatusHandler(c *gin.Context) {
	statusMutex.Lock()
	currentStatus.Status = "idle"
	currentStatus.Message = "Ready to deploy"
	statusMutex.Unlock()

	utils.OKGin(c, map[string]string{"message": "Deployment status reset to idle"})
}

func containerStatusHandler(c *gin.Context) {
	if deployManager == nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Deployment manager not initialized")
		return
	}

	status, err := deployManager.GetContainerStatus()
	if err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, err.Error())
		return
	}

	utils.OKGin(c, status)
}

func initEnvironmentHandler(c *gin.Context) {
	if deployManager == nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Deployment manager not initialized")
		return
	}

	go func() {
		statusMutex.Lock()
		currentStatus.Status = "running"
		currentStatus.Message = "Environment initialization started"
		statusMutex.Unlock()

		err := deployManager.InitEnvironment()
		statusMutex.Lock()
		if err != nil {
			currentStatus.Status = "error"
			currentStatus.Message = fmt.Sprintf("Environment initialization failed: %v", err)
		} else {
			currentStatus.Status = "completed"
			currentStatus.Message = "Environment initialization completed successfully"
		}
		statusMutex.Unlock()
	}()

	utils.OKGin(c, map[string]string{"message": "Environment initialization started"})
}

func cleanupHandler(c *gin.Context) {
	if deployManager == nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Deployment manager not initialized")
		return
	}

	go func() {
		statusMutex.Lock()
		currentStatus.Status = "running"
		currentStatus.Message = "Cleanup started"
		statusMutex.Unlock()

		err := deployManager.Cleanup()
		statusMutex.Lock()
		if err != nil {
			currentStatus.Status = "error"
			currentStatus.Message = fmt.Sprintf("Cleanup failed: %v", err)
		} else {
			currentStatus.Status = "completed"
			currentStatus.Message = "Cleanup completed successfully"
		}
		statusMutex.Unlock()
	}()

	utils.OKGin(c, map[string]string{"message": "Cleanup completed successfully"})
}

func runDeployScript() error {
	// 检查部署脚本是否存在
	scriptPath := "deploy-scripts/one-click.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("deploy script not found: %s", scriptPath)
	}

	// 执行部署脚本
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "bash", scriptPath)

	// 设置工作目录为项目根目录
	cmd.Dir = "."

	// 执行命令并等待完成
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("deploy script failed: %v, output: %s", err, string(output))
	}

	return nil
}
