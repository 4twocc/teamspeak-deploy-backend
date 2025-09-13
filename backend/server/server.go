package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"teamspeak-one-click-deploy/auth"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/deploy"
	"teamspeak-one-click-deploy/instance"
	"teamspeak-one-click-deploy/logs"
	"teamspeak-one-click-deploy/router"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// 全局日志服务
var logService logs.LogService

// SetLogService 设置日志服务
func SetLogService(ls logs.LogService) {
	logService = ls
}

// configWrapper 包装配置以适配认证模块的接口
type configWrapper struct {
	config *config.Config
}

func (cw *configWrapper) GetServerConfig() interface {
	GetAuthConfig() interface{ GetPublicPaths() []string }
} {
	return &serverConfigWrapper{config: cw.config}
}

type serverConfigWrapper struct {
	config *config.Config
}

func (scw *serverConfigWrapper) GetAuthConfig() interface{ GetPublicPaths() []string } {
	return &authConfigWrapper{config: scw.config}
}

type authConfigWrapper struct {
	config *config.Config
}

func (acw *authConfigWrapper) GetPublicPaths() []string {
	return acw.config.Server.Auth.PublicPaths
}

// initDatabase 初始化数据库连接
func InitDatabase(config *config.Config) error {
	// 创建数据库配置
	dbConfig := &database.Config{
		Driver:          config.Database.Driver,
		DSN:             config.Database.DSN,
		MaxIdleConns:    config.Database.MaxIdleConns,
		MaxOpenConns:    config.Database.MaxOpenConns,
		ConnMaxLifetime: config.Database.ConnMaxLifetime,
		AutoMigrate:     config.Database.AutoMigrate,
	}

	// 初始化数据库连接
	if err := database.Init(dbConfig); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// 测试数据库连接
	db, err := database.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// 如果配置了自动迁移，则执行迁移
	if config.Database.AutoMigrate {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			// 用户相关表
			if err := tx.AutoMigrate(&user.User{}); err != nil {
				return fmt.Errorf("failed to migrate users table: %v", err)
			}

			// 实例相关表
			if err := tx.AutoMigrate(&instance.Instance{}); err != nil {
				return fmt.Errorf("failed to migrate instances table: %v", err)
			}

			// 实例日志表
			if err := tx.AutoMigrate(&instance.InstanceLog{}); err != nil {
				return fmt.Errorf("failed to migrate instance_logs table: %v", err)
			}

			// 系统日志表
			if err := tx.AutoMigrate(&logs.SystemLog{}); err != nil {
				return fmt.Errorf("failed to migrate system_logs table: %v", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to auto migrate database: %v", err)
		}

		if logService != nil {
			logService.Info("server", "Database migration completed successfully")
		} else {
			log.Println("Database migration completed successfully")
		}
	}

	if logService != nil {
		logService.Info("server", "Database connection established successfully")
	} else {
		log.Println("Database connection established successfully")
	}
	return nil
}

// InitDeployment 初始化部署模块
func InitDeployment(config *config.Config) error {
	if err := deploy.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize deployment module: %v", err)
	}

	if logService != nil {
		logService.Info("server", "Deployment module initialized successfully")
	} else {
		log.Println("Deployment module initialized successfully")
	}
	return nil
}

// SetupRouter 创建并配置 Gin 路由引擎
func SetupRouter(config *config.Config) (*gin.Engine, error) {
	// 创建Gin引擎
	routerEngine := gin.New()

	// 设置认证模块的配置
	auth.SetServerConfig(&configWrapper{config: config})

	// 添加中间件（必须在注册路由之前）
	// 设置请求ID中间件（最先设置，确保所有请求都有ID）
	routerEngine.Use(utils.RequestIDMiddlewareWithGin())

	// 设置错误处理中间件（在其他中间件之前，确保能捕获所有错误）
	routerEngine.Use(utils.ErrorHandlerMiddlewareWithGin())

	// 注意：Gin的CORS中间件需要单独配置
	if len(config.Server.CORS.AllowedOrigins) > 0 {
		routerEngine.Use(utils.NewCORSWithGin(utils.CORSConfig{
			AllowedOrigins:   config.Server.CORS.AllowedOrigins,
			AllowedMethods:   config.Server.CORS.AllowedMethods,
			AllowedHeaders:   config.Server.CORS.AllowedHeaders,
			ExposeHeaders:    config.Server.CORS.ExposeHeaders,
			AllowCredentials: config.Server.CORS.AllowCredentials,
		}))
	}

	// Add other common middleware
	routerEngine.Use(
		gin.Recovery(), // 保护服务不被panic打垮
		utils.LoggingMiddlewareWithGin(config.Server.Middleware.EnableAccessLog), // 可关访问日志
	)

	// Add rate limiting if enabled
	if config.Server.RateLimit.Enabled {
		routerEngine.Use(utils.RateLimitMiddlewareWithGin(
			rate.Limit(config.Server.RateLimit.RPS),
			config.Server.RateLimit.Burst,
		))
	}

	// Add authentication middleware if required
	if config.Server.Auth.RequireAuth {
		routerEngine.Use(auth.AuthMiddlewareWithGin())
	}

	// 注册路由（必须在中间件之后）
	// 使用带配置的注册方法，以启用 Swagger 文档的开关与白名单控制
	router.RegisterRoutesWithConfig(routerEngine, config)

	return routerEngine, nil
}

// StartServer 启动HTTP服务器
func StartServer(routerEngine *gin.Engine, config *config.Config) error {
	port := config.Server.Port
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: routerEngine,
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if logService != nil {
			logService.Info("server", "Server starting", logs.LogField{Key: "port", Value: port}, logs.LogField{Key: "mode", Value: config.Server.Env})
		} else {
			log.Printf("Server starting on port %s in %s mode", port, config.Server.Env)
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			if logService != nil {
				logService.Error("server", "Error starting server", logs.LogField{Key: "error", Value: err.Error()})
			} else {
				log.Fatalf("Error starting server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal
	<-sigChan

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if logService != nil {
		logService.Info("server", "Shutting down server...")
	} else {
		log.Println("Shutting down server...")
	}
	if err := server.Shutdown(ctx); err != nil {
		if logService != nil {
			logService.Error("server", "Server shutdown error", logs.LogField{Key: "error", Value: err.Error()})
		} else {
			log.Printf("Server shutdown error: %v", err)
		}
		return err
	}

	if logService != nil {
		logService.Info("server", "Server exited properly")
	} else {
		log.Println("Server exited properly")
	}
	return nil
}
