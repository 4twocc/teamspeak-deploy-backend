package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
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
	Port       string           `mapstructure:"port"`
	Env        string           `mapstructure:"env"`
	LogLevel   string           `mapstructure:"log_level"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Auth       AuthConfig       `mapstructure:"auth"`
	CORS       CORSConfig       `mapstructure:"cors"`
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
	RequireAuth bool     `mapstructure:"require_auth"`
	PublicPaths []string `mapstructure:"public_paths"`
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
	Addr                  string        `mapstructure:"addr"`
	CollectInterval       time.Duration `mapstructure:"collect_interval"`
	MinCollectionInterval time.Duration `mapstructure:"min_collection_interval"`
	Alert                 AlertConfig   `mapstructure:"alert"`
	System                SystemConfig  `mapstructure:"system"`
	Performance           PerformanceConfig `mapstructure:"performance"`
	Redis                 RedisConfig   `mapstructure:"redis"`
}

// SystemConfig 系统监控配置
type SystemConfig struct {
	// 监控的挂载点
	MountPoints []string `mapstructure:"mount_points"`

	// 网络接口
	NetworkInterfaces []string `mapstructure:"network_interfaces"`
}

// PerformanceConfig 性能优化配置
type PerformanceConfig struct {
	// 系统指标采样率 (1/N 的频率收集)
	SystemSampleRate int `mapstructure:"system_sample_rate"`
	
	// 业务指标采样率 (1/N 的频率收集)
	BusinessSampleRate int `mapstructure:"business_sample_rate"`
	
	// 历史记录最大数量
	MaxHistorySize int `mapstructure:"max_history_size"`
	
	// 收集后延迟时间
	CollectionDelay time.Duration `mapstructure:"collection_delay"`
	
	// 函数间延迟时间
	InterFuncDelay time.Duration `mapstructure:"inter_func_delay"`
	
	// 函数内延迟时间
	InnerFuncDelay time.Duration `mapstructure:"inner_func_delay"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled       bool            `mapstructure:"enabled"`
	NotifyMethods []string        `mapstructure:"notify_methods"`
	Thresholds    ThresholdConfig `mapstructure:"thresholds"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	CPU          float64 `mapstructure:"cpu"`
	Memory       float64 `mapstructure:"memory"`
	Disk         float64 `mapstructure:"disk"`
	VoiceQuality float64 `mapstructure:"voice_quality"`
}

// TeamspeakConfig TeamSpeak配置
type TeamspeakConfig struct {
	Host                    string        `mapstructure:"host"`
	QueryPort               int           `mapstructure:"query_port"`
	VirtualServerPort       int           `mapstructure:"virtual_server_port"`
	VirtualServerID         int           `mapstructure:"virtual_server_id"`
	Username                string        `mapstructure:"username"`
	Password                string        `mapstructure:"password"`
	Nickname                string        `mapstructure:"nickname"`
	ReconnectMaxRetries     int           `mapstructure:"reconnect_max_retries"`
	ReconnectInitialBackoff time.Duration `mapstructure:"reconnect_initial_backoff"`
	ReconnectMaxBackoff     time.Duration `mapstructure:"reconnect_max_backoff"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret"`
	PasswordSaltRounds int    `mapstructure:"password_salt_rounds"`
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	ScriptDir string        `mapstructure:"script_dir"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

// Load 从文件加载配置
func Load(filename string) (*Config, error) {
	configFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		return nil, err
	}

	// 从环境变量加载敏感配置（如果存在）
	loadSensitiveConfigFromEnv(&config)

	return &config, nil
}

// loadSensitiveConfigFromEnv 从环境变量加载敏感配置
func loadSensitiveConfigFromEnv(config *Config) {
	// 从环境变量加载 TeamSpeak 密码
	if tsPassword := os.Getenv("TEAMSPEAK_PASSWORD"); tsPassword != "" {
		config.Teamspeak.Password = tsPassword
	}

	// 从环境变量加载 JWT secret
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}

	// 从环境变量加载数据库配置
	if dbDSN := os.Getenv("DATABASE_DSN"); dbDSN != "" {
		config.Database.DSN = dbDSN
	}

	if dbDriver := os.Getenv("DATABASE_DRIVER"); dbDriver != "" {
		config.Database.Driver = dbDriver
	}

	// 从环境变量加载服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}

	if env := os.Getenv("SERVER_ENV"); env != "" {
		config.Server.Env = env
	}

	if logLevel := os.Getenv("SERVER_LOG_LEVEL"); logLevel != "" {
		config.Server.LogLevel = logLevel
	}

	// 从环境变量加载监控配置
	if interval := os.Getenv("MONITORING_COLLECT_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.Monitoring.CollectInterval = d
		}
	}

	// 从环境变量加载部署配置
	if scriptDir := os.Getenv("DEPLOYMENT_SCRIPT_DIR"); scriptDir != "" {
		config.Deployment.ScriptDir = scriptDir
	}

	if timeout := os.Getenv("DEPLOYMENT_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.Deployment.Timeout = d
		}
	}
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
			Addr:            ":9090",
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
			System: SystemConfig{
				MountPoints:       []string{"/"},
				NetworkInterfaces: []string{"eth0"},
			},
			Performance: PerformanceConfig{
				SystemSampleRate:      3,
				BusinessSampleRate:    4,
				MaxHistorySize:        50,
				CollectionDelay:       500 * time.Millisecond,
				InterFuncDelay:        50 * time.Millisecond,
				InnerFuncDelay:        20 * time.Millisecond,
			},
			Redis: RedisConfig{
				Enabled:  false,
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
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
			JWTSecret:          "your_jwt_secret_here",
			PasswordSaltRounds: 10,
		},
		Deployment: DeploymentConfig{
			ScriptDir: "deploy-scripts",
			Timeout:   10 * time.Minute,
		},
	}
}