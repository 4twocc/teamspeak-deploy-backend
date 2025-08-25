// backend/main.go
package main

import (
	"log"
	"os"

	"teamspeak-one-click-deploy/auth"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/instance"
	"teamspeak-one-click-deploy/monitor"
	"teamspeak-one-click-deploy/server"
	"teamspeak-one-click-deploy/user"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 先尝试加载 .env 文件（如果存在），使得 os.Getenv 能读取这些值
	if loadCfgErr := utils.DiscoverAndLoadDotEnv(); loadCfgErr != nil {
		log.Printf("No .env files loaded: %v", loadCfgErr)
	}

	// Load configuration
	cfg, loadCfgErr := config.Load("config.yaml")
	if loadCfgErr != nil {
		log.Fatalf("Failed to load config: %v", loadCfgErr)
	}

	// 初始化认证模块
	auth.Init(cfg)

	// 初始化数据库
	if dataBaseErr := server.InitDatabase(cfg); dataBaseErr != nil {
		log.Fatalf("Failed to initialize database: %v", dataBaseErr)
	}

	// 确保在程序退出时关闭数据库连接
	defer func() {
		if dataBaseCloseErr := database.Close(); dataBaseCloseErr != nil {
			log.Printf("Error closing database connection: %v", dataBaseCloseErr)
		}
	}()

	// 初始化用户服务
	if userIniterr := user.Initialize(); userIniterr != nil {
		log.Fatalf("Failed to initialize user service: %v", userIniterr)
	}

	// 初始化实例服务
	if instanceInitErr := instance.Initialize(); instanceInitErr != nil {
		log.Fatalf("Failed to initialize instance service: %v", instanceInitErr)
	}

	// 初始化部署模块
	if initDeployErr := server.InitDeployment(cfg); initDeployErr != nil {
		log.Fatalf("Failed to initialize deployment module: %v", initDeployErr)
	}

	// 更新监控模块配置（使用已加载的配置）
	monitor.UpdateConfig(cfg)

	// 启动监控服务（在单独的 goroutine 中启动以避免阻塞主线程）
	go func() {
		if monitorRunErr := monitor.Run(cfg); monitorRunErr != nil {
			log.Printf("Failed to start monitoring service: %v", monitorRunErr)
		}
	}()

	// 创建并配置路由引擎
	// 根据环境变量设置 gin 模式（优先使用 GIN_MODE）
	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		gin.SetMode(ginMode)
		log.Printf("Applied GIN_MODE=%s", ginMode)
	}

	routerEngine, err := server.SetupRouter(cfg)
	if err != nil {
		log.Fatalf("Failed to setup router: %v", err)
	}

	// 启动服务器
	if err := server.StartServer(routerEngine, cfg); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
