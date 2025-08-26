// backend/main.go
// TeamSpeak一键部署工具主程序入口
// 版本: v1.0.0
// 功能: 初始化各模块服务，启动HTTP服务器，提供TeamSpeak服务器一键部署功能
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"teamspeak-one-click-deploy/auth"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/instance"
	"teamspeak-one-click-deploy/logs"
	"teamspeak-one-click-deploy/monitor"
	"teamspeak-one-click-deploy/server"
	"teamspeak-one-click-deploy/user"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

// main 主函数，程序入口点
// 功能: 初始化所有服务模块，启动HTTP服务器，处理优雅关闭
func main() {
	// 先尝试加载 .env 文件（如果存在），使得 os.Getenv 能读取这些值
	if loadCfgErr := utils.DiscoverAndLoadDotEnv(); loadCfgErr != nil {
		log.Printf("No .env files loaded: %v", loadCfgErr)
	}

	// 加载配置文件
	cfg, loadCfgErr := config.Load("config.yaml")
	if loadCfgErr != nil {
		log.Fatalf("Failed to load config: %v", loadCfgErr)
	}

	log.Println("TeamSpeak一键部署工具启动中...")

	// 初始化认证模块
	auth.Init(cfg)
	log.Println("认证模块初始化完成")

	// 初始化数据库
	if dataBaseErr := server.InitDatabase(cfg); dataBaseErr != nil {
		log.Fatalf("Failed to initialize database: %v", dataBaseErr)
	}
	log.Println("数据库初始化完成")

	// 初始化日志服务（在数据库初始化之后）
	logService, logErr := logs.Init(cfg)
	if logErr != nil {
		log.Fatalf("Failed to initialize log service: %v", logErr)
	}
	defer logService.Close()

	// 创建日志适配器
	logAdapter := logs.NewLogServiceAdapter(logService)

	// 设置各模块的日志服务
	utils.SetLogService(logAdapter)
	server.SetLogService(logService)
	instance.SetLogService(logService)
	auth.SetLogService(logService)
	monitor.SetLogService(logService)

	logService.Info("main", "日志服务初始化完成")

	// 确保在程序退出时关闭数据库连接
	defer func() {
		if dataBaseCloseErr := database.Close(); dataBaseCloseErr != nil {
			logService.Error("main", "Error closing database connection", logs.LogField{Key: "error", Value: dataBaseCloseErr})
		}
	}()

	// 初始化用户服务
	if userIniterr := user.Initialize(); userIniterr != nil {
		logService.Error("main", "Failed to initialize user service", logs.LogField{Key: "error", Value: userIniterr})
		log.Fatalf("Failed to initialize user service: %v", userIniterr)
	}
	logService.Info("main", "用户服务初始化完成")

	// 初始化实例服务
	if instanceInitErr := instance.Initialize(); instanceInitErr != nil {
		logService.Error("main", "Failed to initialize instance service", logs.LogField{Key: "error", Value: instanceInitErr})
		log.Fatalf("Failed to initialize instance service: %v", instanceInitErr)
	}
	logService.Info("main", "实例服务初始化完成")

	// 初始化部署模块
	if initDeployErr := server.InitDeployment(cfg); initDeployErr != nil {
		logService.Error("main", "Failed to initialize deployment module", logs.LogField{Key: "error", Value: initDeployErr})
		log.Fatalf("Failed to initialize deployment module: %v", initDeployErr)
	}
	logService.Info("main", "部署模块初始化完成")

	// 更新监控模块配置（使用已加载的配置）
	monitor.UpdateConfig(cfg)

	// 启动监控服务（在单独的 goroutine 中启动以避免阻塞主线程）
	go func() {
		if monitorRunErr := monitor.Run(cfg); monitorRunErr != nil {
			logService.Error("main", "Failed to start monitoring service", logs.LogField{Key: "error", Value: monitorRunErr})
		}
	}()
	logService.Info("main", "监控服务启动完成")

	// 创建并配置路由引擎
	// 根据环境变量设置 gin 模式（优先使用 GIN_MODE）
	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		gin.SetMode(ginMode)
		logService.Info("main", "Applied GIN_MODE", logs.LogField{Key: "mode", Value: ginMode})
	}

	routerEngine, err := server.SetupRouter(cfg)
	if err != nil {
		logService.Error("main", "Failed to setup router", logs.LogField{Key: "error", Value: err})
		log.Fatalf("Failed to setup router: %v", err)
	}
	logService.Info("main", "路由引擎配置完成")

	// 构建服务器地址
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         addr,
		Handler:      routerEngine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器（在单独的 goroutine 中）
	go func() {
		logService.Info("main", "HTTP服务器启动", logs.LogField{Key: "address", Value: addr})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logService.Error("main", "HTTP服务器启动失败", logs.LogField{Key: "error", Value: err})
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logService.Info("main", "TeamSpeak一键部署工具启动完成", logs.LogField{Key: "version", Value: "v1.0.0"})

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logService.Info("main", "收到关闭信号，开始关闭服务器...")

	// 设置5秒的超时时间来关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭监控模块
	if err := monitor.Close(); err != nil {
		logService.Error("main", "监控模块关闭失败", logs.LogField{Key: "error", Value: err})
	}

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		logService.Error("main", "服务器强制关闭", logs.LogField{Key: "error", Value: err})
		log.Fatal("Server forced to shutdown:", err)
	}

	logService.Info("main", "服务器关闭")
}
