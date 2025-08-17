package monitor

import (
	"os"
	"sync"
	
	configPkg "teamspeak-one-click-deploy/config"
)

var (
	monitorConfig *configPkg.Config
	configOnce    sync.Once
)

// GetConfig 获取配置实例（单例模式）
// 现在从主配置获取监控配置
func GetConfig() *configPkg.Config {
	return monitorConfig
}

// UpdateConfig 更新监控配置
func UpdateConfig(cfg *configPkg.Config) {
	// 更新监控配置
	monitorConfig = cfg
	
	// 从环境变量加载TeamSpeak密码
	if tsPassword := os.Getenv("TEAMSPEAK_PASSWORD"); tsPassword != "" {
		monitorConfig.Teamspeak.Password = tsPassword
	}
}

// LoadConfigFromFile 从主配置加载监控配置
func LoadConfigFromFile(cfg *configPkg.Config) error {
	// 使用单例模式确保配置只初始化一次
	configOnce.Do(func() {
		UpdateConfig(cfg)
	})
	return nil
}