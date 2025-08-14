package deploy

import (
	"time"

	"teamspeak-one-click-deploy/config"
)

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	ScriptDir string        `mapstructure:"script_dir"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// LoadConfigFromAppConfig 从应用配置加载部署配置
func LoadConfigFromAppConfig(appConfig *config.Config) (*DeploymentConfig, error) {
	deployConfig := &DeploymentConfig{
		ScriptDir: appConfig.Deployment.ScriptDir,
		Timeout:   appConfig.Deployment.Timeout,
	}

	// 设置默认值
	if deployConfig.ScriptDir == "" {
		deployConfig.ScriptDir = "deploy-scripts"
	}

	if deployConfig.Timeout == 0 {
		deployConfig.Timeout = 10 * time.Minute
	}

	return deployConfig, nil
}