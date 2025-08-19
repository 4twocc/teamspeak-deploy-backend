// database/database.go
package database

import (
	"fmt"
	"os"
	"path/filepath"
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
		// 确保目录存在
		dir := filepath.Dir(config.DSN)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %v", err)
			}
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
