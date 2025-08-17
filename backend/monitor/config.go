package monitor

import (
	"sync"
	
	configPkg "teamspeak-one-click-deploy/config"
)

var (
	monitorConfig *configPkg.Config
	once    sync.Once
)

// GetConfig 获取配置实例（单例模式）
// 现在从主配置获取监控配置
func GetConfig() *configPkg.Config {
	return monitorConfig
}

// LoadConfigFromFile 从主配置加载监控配置
func LoadConfigFromFile(path string) error {
	once.Do(func() {
		// 加载主配置
		cfg, err := configPkg.Load(path)
		if err != nil {
			// 如果加载失败，使用默认配置
			monitorConfig = configPkg.DefaultConfig()
			return
		}
		monitorConfig = cfg
	})
	
	return nil
}

// UpdateConfig 更新监控配置
func UpdateConfig(cfg *configPkg.Config) {
	once.Do(func() {
		monitorConfig = cfg
	})
}

