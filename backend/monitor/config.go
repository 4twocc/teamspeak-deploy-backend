package monitor

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 监控系统配置
type Config struct {
	// 收集间隔
	CollectInterval time.Duration `yaml:"collect_interval"`

	// 告警相关配置
	AlertConfig AlertConfig `yaml:"alert_config"`

	// TeamSpeak 服务器配置
	TeamSpeakConfig TeamSpeakConfig `yaml:"teamspeak_config"`

	// 系统监控配置
	SystemConfig SystemConfig `yaml:"system_config"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	// 是否启用告警
	Enabled bool `yaml:"enabled"`

	// 告警通知方式：console, email, webhook
	NotifyMethods []string `yaml:"notify_methods"`

	// 告警级别阈值
	Thresholds struct {
		CPU          float64 `yaml:"cpu"`
		Memory       float64 `yaml:"memory"`
		Disk         float64 `yaml:"disk"`
		VoiceQuality float64 `yaml:"voice_quality"`
	} `yaml:"thresholds"`
}

// TeamSpeakConfig TeamSpeak 服务器配置
type TeamSpeakConfig struct {
	Host                    string `yaml:"host"`
	Port                    int    `yaml:"port"`
	Username                string `yaml:"username"`
	Password                string `yaml:"password"`
	Timeout                 int    `yaml:"timeout"`
	VirtualServerPort       int    `yaml:"virtual_server_port"`
	VirtualServerID         int    `yaml:"virtual_server_id"`
	Nickname                string `yaml:"nickname"`
	ReconnectMaxRetries     int    `yaml:"reconnect_max_retries"`
	ReconnectInitialBackoff int    `yaml:"reconnect_initial_backoff"` // seconds
	ReconnectMaxBackoff     int    `yaml:"reconnect_max_backoff"`     // seconds
}

// SystemConfig 系统监控配置
type SystemConfig struct {
	// 监控的挂载点
	MountPoints []string `yaml:"mount_points"`

	// 网络接口
	NetworkInterfaces []string `yaml:"network_interfaces"`
}

var (
	config     *Config
	configOnce sync.Once
)

// GetConfig 获取配置实例（单例模式）
func GetConfig() *Config {
	configOnce.Do(func() {
		// 默认配置
		config = &Config{
			CollectInterval: 60 * time.Second,
			AlertConfig: AlertConfig{
				Enabled:       true,
				NotifyMethods: []string{"console"},
				Thresholds: struct {
					CPU          float64 `yaml:"cpu"`
					Memory       float64 `yaml:"memory"`
					Disk         float64 `yaml:"disk"`
					VoiceQuality float64 `yaml:"voice_quality"`
				}{
					CPU:          90,
					Memory:       90,
					Disk:         90,
					VoiceQuality: 0.9,
				},
			},
			TeamSpeakConfig: TeamSpeakConfig{
				Host:     "localhost",
				Port:     10011,
				Username: "serveradmin",
				Timeout:  10,
				// VirtualServerPort/ID/Nickname 默认留空
				ReconnectMaxRetries:     5,
				ReconnectInitialBackoff: 1,
				ReconnectMaxBackoff:     30,
			},
			SystemConfig: SystemConfig{
				MountPoints:       []string{"/"},
				NetworkInterfaces: []string{"eth0"},
			},
		}
	})
	return config
}

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(path string) error {
	// 读取 YAML 文件
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 定义与文件结构匹配的临时结构体
	type fileCfg struct {
		Monitoring struct {
			CollectInterval string `yaml:"collect_interval"`
			Alert           struct {
				Enabled       bool     `yaml:"enabled"`
				NotifyMethods []string `yaml:"notify_methods"`
				Thresholds    struct {
					CPU          float64 `yaml:"cpu"`
					Memory       float64 `yaml:"memory"`
					Disk         float64 `yaml:"disk"`
					VoiceQuality float64 `yaml:"voice_quality"`
				} `yaml:"thresholds"`
			} `yaml:"alert"`
			System struct {
				MountPoints       []string `yaml:"mount_points"`
				NetworkInterfaces []string `yaml:"network_interfaces"`
			} `yaml:"system"`
		} `yaml:"monitoring"`
		TeamSpeak struct {
			Host                    string `yaml:"host"`
			QueryPort               int    `yaml:"query_port"`
			Username                string `yaml:"username"`
			Password                string `yaml:"password"`
			Timeout                 int    `yaml:"timeout"`
			VirtualServerPort       int    `yaml:"virtual_server_port"`
			VirtualServerID         int    `yaml:"virtual_server_id"`
			Nickname                string `yaml:"nickname"`
			ReconnectMaxRetries     int    `yaml:"reconnect_max_retries"`
			ReconnectInitialBackoff int    `yaml:"reconnect_initial_backoff"`
			ReconnectMaxBackoff     int    `yaml:"reconnect_max_backoff"`
		} `yaml:"teamspeak"`
	}

	var fc fileCfg
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return err
	}

	cfg := GetConfig()

	// 监控采集间隔
	if fc.Monitoring.CollectInterval != "" {
		if d, err := time.ParseDuration(fc.Monitoring.CollectInterval); err == nil {
			cfg.CollectInterval = d
		}
	}

	// 告警配置
	cfg.AlertConfig.Enabled = fc.Monitoring.Alert.Enabled
	if len(fc.Monitoring.Alert.NotifyMethods) > 0 {
		cfg.AlertConfig.NotifyMethods = fc.Monitoring.Alert.NotifyMethods
	}
	cfg.AlertConfig.Thresholds.CPU = fc.Monitoring.Alert.Thresholds.CPU
	cfg.AlertConfig.Thresholds.Memory = fc.Monitoring.Alert.Thresholds.Memory
	cfg.AlertConfig.Thresholds.Disk = fc.Monitoring.Alert.Thresholds.Disk
	if fc.Monitoring.Alert.Thresholds.VoiceQuality > 0 {
		cfg.AlertConfig.Thresholds.VoiceQuality = fc.Monitoring.Alert.Thresholds.VoiceQuality
	}

	// TeamSpeak 配置
	if fc.TeamSpeak.Host != "" {
		cfg.TeamSpeakConfig.Host = fc.TeamSpeak.Host
	}
	if fc.TeamSpeak.QueryPort != 0 {
		cfg.TeamSpeakConfig.Port = fc.TeamSpeak.QueryPort
	}
	if fc.TeamSpeak.Username != "" {
		cfg.TeamSpeakConfig.Username = fc.TeamSpeak.Username
	}
	if fc.TeamSpeak.Password != "" {
		cfg.TeamSpeakConfig.Password = fc.TeamSpeak.Password
	}
	if fc.TeamSpeak.Timeout > 0 {
		cfg.TeamSpeakConfig.Timeout = fc.TeamSpeak.Timeout
	}
	if fc.TeamSpeak.VirtualServerPort != 0 {
		cfg.TeamSpeakConfig.VirtualServerPort = fc.TeamSpeak.VirtualServerPort
	}
	if fc.TeamSpeak.VirtualServerID > 0 {
		cfg.TeamSpeakConfig.VirtualServerID = fc.TeamSpeak.VirtualServerID
	}
	if fc.TeamSpeak.Nickname != "" {
		cfg.TeamSpeakConfig.Nickname = fc.TeamSpeak.Nickname
	}
	if fc.TeamSpeak.ReconnectMaxRetries > 0 {
		cfg.TeamSpeakConfig.ReconnectMaxRetries = fc.TeamSpeak.ReconnectMaxRetries
	}
	if fc.TeamSpeak.ReconnectInitialBackoff > 0 {
		cfg.TeamSpeakConfig.ReconnectInitialBackoff = fc.TeamSpeak.ReconnectInitialBackoff
	}
	if fc.TeamSpeak.ReconnectMaxBackoff > 0 {
		cfg.TeamSpeakConfig.ReconnectMaxBackoff = fc.TeamSpeak.ReconnectMaxBackoff
	}

	// 系统监控配置（可选）
	if len(fc.Monitoring.System.MountPoints) > 0 {
		cfg.SystemConfig.MountPoints = fc.Monitoring.System.MountPoints
	}
	if len(fc.Monitoring.System.NetworkInterfaces) > 0 {
		cfg.SystemConfig.NetworkInterfaces = fc.Monitoring.System.NetworkInterfaces
	}

	return nil
}
