// backend/main.go
package main

import (
	"log"
	"os"
	"time"

	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/instance"
	"teamspeak-one-click-deploy/monitor"
	"teamspeak-one-click-deploy/server"
	"teamspeak-one-click-deploy/users"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置时区为北京时间 UTC+8
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("Warning: failed to load Asia/Shanghai timezone: %v", err)
		// 如果加载时区失败，尝试使用UTC+8的固定时区偏移
		location = time.FixedZone("UTC+8", 8*60*60)
		log.Println("Using fixed UTC+8 timezone")
	}
	time.Local = location
	log.Println("Set timezone to Beijing (UTC+8)")

	// 先尝试加载 .env 文件（如果存在），使得 os.Getenv 能读取这些值
	if err := utils.DiscoverAndLoadDotEnv(); err != nil {
		log.Printf("No .env files loaded: %v", err)
	}

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := server.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 确保在程序退出时关闭数据库连接
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// 初始化用户服务
	if err := users.Initialize(); err != nil {
		log.Fatalf("Failed to initialize user service: %v", err)
	}

	// 初始化实例服务
	if err := instance.Initialize(); err != nil {
		log.Fatalf("Failed to initialize instance service: %v", err)
	}

	// 初始化部署模块
	if err := server.InitDeployment(cfg); err != nil {
		log.Fatalf("Failed to initialize deployment module: %v", err)
	}

	// 更新监控模块配置（使用已加载的配置）
	monitor.UpdateConfig(cfg)

	// 启动监控服务（在单独的 goroutine 中启动以避免阻塞主线程）
	go func() {
		if err := monitor.Run(cfg); err != nil {
			log.Printf("Failed to start monitoring service: %v", err)
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
