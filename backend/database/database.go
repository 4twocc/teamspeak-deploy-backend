// database/database.go
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 全局数据库实例
var DB *gorm.DB

// Config 数据库配置
// 对应 config.yaml 中的 database 部分
type Config struct {
	Driver          string        `yaml:"driver"`
	DSN             string        `yaml:"dsn"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	AutoMigrate     bool          `yaml:"auto_migrate"`
}

// Init 初始化数据库连接
func Init(config *Config) error {
	var dialector gorm.Dialector

	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(config.DSN)
	case "postgres":
		dialector = postgres.Open(config.DSN)
	case "sqlite":
		// 支持 URL 风格的 DSN（例如 sqlite:///data/teamspeak.db 或 file:/data/teamspeak.db）
		// 将其规范化为文件系统路径，避免把整个 URL 当作路径拼接导致错误（如 "/app/sqlite:/data"）
		config.DSN = strings.TrimPrefix(config.DSN, "sqlite://")
		if after, ok := strings.CutPrefix(config.DSN, "file:"); ok {
			config.DSN = after
		}

		// 确保目录存在
		dir := filepath.Dir(config.DSN)

		// 如果是相对路径，转换为绝对路径
		if !filepath.IsAbs(config.DSN) {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current working directory: %v", err)
			}
			absPath := filepath.Join(wd, config.DSN)
			dir = filepath.Dir(absPath)

			// 确保目录存在，使用更宽松的权限
			if err := os.MkdirAll(dir, 0777); err != nil {
				return fmt.Errorf("failed to create database directory '%s': %v", dir, err)
			}

			// 更新DSN为绝对路径
			config.DSN = absPath
		} else {
			// 对于绝对路径，确保目录存在，使用更宽松的权限
			if err := os.MkdirAll(dir, 0777); err != nil {
				return fmt.Errorf("failed to create database directory '%s': %v", dir, err)
			}
		}

		// 确保数据库文件存在或者可以被创建
		if _, err := os.Stat(config.DSN); os.IsNotExist(err) {
			// 直接创建数据库文件
			file, err := os.OpenFile(config.DSN, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				// 获取更多错误信息
				stat, _ := os.Stat(dir)
				return fmt.Errorf("failed to create database file '%s': %v. Directory info: %+v", config.DSN, err, stat)
			}
			file.Close()
		}

		dialector = sqlite.Open(config.DSN)
	default:
		return fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// 配置 GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// 解析连接最大生命周期
	if config.ConnMaxLifetime != 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// 测试数据库连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	DB = db
	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %v", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate 自动迁移模型
func AutoMigrate(models ...any) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	return DB.AutoMigrate(models...)
}
