package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Teamspeak  TeamspeakConfig  `mapstructure:"teamspeak"`
	Security   SecurityConfig   `mapstructure:"security"`
	Deployment DeploymentConfig `mapstructure:"deployment"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port      string `mapstructure:"port"`
	Env       string `mapstructure:"env"`
	LogLevel  string `mapstructure:"log_level"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Auth      AuthConfig      `mapstructure:"auth"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"`
	RPS     float64 `mapstructure:"rps"`
	Burst   int     `mapstructure:"burst"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	RequireAuth  bool     `mapstructure:"require_auth"`
	PublicPaths  []string `mapstructure:"public_paths"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	EnableAccessLog bool `mapstructure:"enable_access_log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	CollectInterval time.Duration    `mapstructure:"collect_interval"`
	Alert           AlertConfig      `mapstructure:"alert"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	NotifyMethods  []string `mapstructure:"notify_methods"`
	Thresholds     ThresholdConfig `mapstructure:"thresholds"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	CPU     float64 `mapstructure:"cpu"`
	Memory  float64 `mapstructure:"memory"`
	Disk    float64 `mapstructure:"disk"`
	VoiceQuality float64 `mapstructure:"voice_quality"`
}

// TeamspeakConfig TeamSpeak配置
type TeamspeakConfig struct {
	Host                string        `mapstructure:"host"`
	QueryPort           int           `mapstructure:"query_port"`
	VirtualServerPort   int           `mapstructure:"virtual_server_port"`
	VirtualServerID     int           `mapstructure:"virtual_server_id"`
	Username            string        `mapstructure:"username"`
	Password            string        `mapstructure:"password"`
	Nickname            string        `mapstructure:"nickname"`
	ReconnectMaxRetries int           `mapstructure:"reconnect_max_retries"`
	ReconnectInitialBackoff time.Duration `mapstructure:"reconnect_initial_backoff"`
	ReconnectMaxBackoff time.Duration `mapstructure:"reconnect_max_backoff"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret         string `mapstructure:"jwt_secret"`
	PasswordSaltRounds int    `mapstructure:"password_salt_rounds"`
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	ScriptDir string        `mapstructure:"script_dir"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:     "8080",
			Env:      "development",
			LogLevel: "info",
			RateLimit: RateLimitConfig{
				Enabled: true,
				RPS:     10.0,
				Burst:   20,
			},
			Auth: AuthConfig{
				RequireAuth: true,
				PublicPaths: []string{
					"/api/auth/login",
					"/api/auth/me",
					"/api/health",
					"/metrics",
				},
			},
			CORS: CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
				AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
				ExposeHeaders:    []string{"X-Request-ID"},
				AllowCredentials: true,
			},
			Middleware: MiddlewareConfig{
				EnableAccessLog: true,
			},
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             "data/teamspeak.db",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
			AutoMigrate:     true,
		},
		Monitoring: MonitoringConfig{
			CollectInterval: 60 * time.Second,
			Alert: AlertConfig{
				Enabled:       true,
				NotifyMethods: []string{"console", "email"},
				Thresholds: ThresholdConfig{
					CPU:          90,
					Memory:       85,
					Disk:         80,
					VoiceQuality: 0.7,
				},
			},
		},
		Teamspeak: TeamspeakConfig{
			Host:                    "127.0.0.1",
			QueryPort:               10011,
			VirtualServerPort:       30033,
			VirtualServerID:         0,
			Username:                "serveradmin",
			Password:                "your_password_here",
			Nickname:                "Monitoring Bot",
			ReconnectMaxRetries:     5,
			ReconnectInitialBackoff: 1 * time.Second,
			ReconnectMaxBackoff:     30 * time.Second,
		},
		Security: SecurityConfig{
			JWTSecret:         "your_jwt_secret_here",
			PasswordSaltRounds: 10,
		},
		Deployment: DeploymentConfig{
			ScriptDir: "deploy-scripts",
			Timeout:   10 * time.Minute,
		},
	}
}