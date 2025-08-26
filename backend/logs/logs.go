package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
)

func Init(config *config.Config) (LogService, error) {
	// 获取项目根目录路径（backend的上级目录）
	backendDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// 获取项目根目录（backend的父目录）
	projectRoot := filepath.Dir(backendDir)

	// 构建日志文件的绝对路径，指向项目根目录下的.logs文件夹
	logFilePath := filepath.Join(projectRoot, ".logs", "system.log")

	// 初始化日志服务
	logService, err := NewLogService(database.DB, LogConfig{
		Level:         config.Logging.Level,
		EnableDB:      config.Logging.EnableDatabase,
		RetentionDays: config.Logging.RetentionDays,
		BatchSize:     config.Logging.BatchSize,
		BatchInterval: int(config.Logging.FlushInterval.Seconds()),
		EnableFile:    true,
		FilePath:      logFilePath,
	})
	if err != nil {
		return nil, err
	}
	return logService, nil
}
