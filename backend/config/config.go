package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Teamspeak  TeamspeakConfig  `yaml:"teamspeak"`
	Security   SecurityConfig   `yaml:"security"`
	Deployment DeploymentConfig `yaml:"deployment"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port       string           `yaml:"port"`
	Env        string           `yaml:"env"`
	LogLevel   string           `yaml:"log_level"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit"`
	Auth       AuthConfig       `yaml:"auth"`
	CORS       CORSConfig       `yaml:"cors"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool    `yaml:"enabled"`
	RPS     float64 `yaml:"rps"`
	Burst   int     `yaml:"burst"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	RequireAuth bool     `yaml:"require_auth"`
	PublicPaths []string `yaml:"public_paths"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	EnableAccessLog bool `yaml:"enable_access_log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	DSN             string        `yaml:"dsn"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	AutoMigrate     bool          `yaml:"auto_migrate"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	Addr                  string            `yaml:"addr"`
	CollectInterval       time.Duration     `yaml:"collect_interval"`
	MinCollectionInterval time.Duration     `yaml:"min_collection_interval"`
	Alert                 AlertConfig       `yaml:"alert"`
	System                SystemConfig      `yaml:"system"`
	Performance           PerformanceConfig `yaml:"performance"`
	Redis                 RedisConfig       `yaml:"redis"`
}

// SystemConfig 系统监控配置
type SystemConfig struct {
	// 监控的挂载点
	MountPoints []string `yaml:"mount_points"`

	// 网络接口
	NetworkInterfaces []string `yaml:"network_interfaces"`
}

// PerformanceConfig 性能优化配置
type PerformanceConfig struct {
	// 系统指标采样率 (1/N 的频率收集)
	SystemSampleRate int `yaml:"system_sample_rate"`

	// 业务指标采样率 (1/N 的频率收集)
	BusinessSampleRate int `yaml:"business_sample_rate"`

	// 历史记录最大数量
	MaxHistorySize int `yaml:"max_history_size"`

	// 收集后延迟时间
	CollectionDelay time.Duration `yaml:"collection_delay"`

	// 函数间延迟时间
	InterFuncDelay time.Duration `yaml:"inter_func_delay"`

	// 函数内延迟时间
	InnerFuncDelay time.Duration `yaml:"inner_func_delay"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled       bool            `yaml:"enabled"`
	NotifyMethods []string        `yaml:"notify_methods"`
	Thresholds    ThresholdConfig `yaml:"thresholds"`
}

// ThresholdConfig 阈值配置
type ThresholdConfig struct {
	CPU          float64 `yaml:"cpu"`
	Memory       float64 `yaml:"memory"`
	Disk         float64 `yaml:"disk"`
	VoiceQuality float64 `yaml:"voice_quality"`
}

// TeamspeakConfig TeamSpeak配置
type TeamspeakConfig struct {
	Host                    string        `yaml:"host"`
	QueryPort               int           `yaml:"query_port"`
	VirtualServerPort       int           `yaml:"virtual_server_port"`
	VirtualServerID         int           `yaml:"virtual_server_id"`
	Username                string        `yaml:"username"`
	Password                string        `yaml:"password"`
	Nickname                string        `yaml:"nickname"`
	ReconnectMaxRetries     int           `yaml:"reconnect_max_retries"`
	ReconnectInitialBackoff time.Duration `yaml:"reconnect_initial_backoff"`
	ReconnectMaxBackoff     time.Duration `yaml:"reconnect_max_backoff"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret          string `yaml:"jwt_secret"`
	PasswordSaltRounds int    `yaml:"password_salt_rounds"`
}

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	ScriptDir string        `yaml:"script_dir"`
	Timeout   time.Duration `yaml:"timeout"`
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

	// 从环境变量加载 TeamSpeak username
	if tsUsername := os.Getenv("TEAMSPEAK_USERNAME"); tsUsername != "" {
		config.Teamspeak.Username = tsUsername
	}

	// 从环境变量加载 TeamSpeak 主机和端口
	if tsHost := os.Getenv("TEAMSPEAK_HOST"); tsHost != "" {
		config.Teamspeak.Host = tsHost
	}

	if tsPort := os.Getenv("TEAMSPEAK_QUERY_PORT"); tsPort != "" {
		if p, err := strconv.Atoi(tsPort); err == nil {
			config.Teamspeak.QueryPort = p
		}
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
					"/api/v1/auth/login",
					"/api/v1/auth/register",
					"/health",
				},
			},
			CORS: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{
					"GET", "POST", "PUT", "DELETE", "OPTIONS",
				},
				AllowedHeaders: []string{
					"Accept", "Authorization", "Content-Type", "X-CSRF-Token",
				},
				ExposeHeaders:    []string{},
				AllowCredentials: false,
			},
			Middleware: MiddlewareConfig{
				EnableAccessLog: true,
			},
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			DSN:             "data/teamspeak.db",
			MaxIdleConns:    5,
			MaxOpenConns:    10,
			ConnMaxLifetime: 1 * time.Hour,
			AutoMigrate:     true,
		},
		Monitoring: MonitoringConfig{
			Addr:                  ":9090",
			CollectInterval:       1 * time.Minute,
			MinCollectionInterval: 1 * time.Hour,
			Alert: AlertConfig{
				Enabled:       true,
				NotifyMethods: []string{"log"},
				Thresholds: ThresholdConfig{
					CPU:          80.0,
					Memory:       80.0,
					Disk:         90.0,
					VoiceQuality: 70.0,
				},
			},
			System: SystemConfig{
				MountPoints:       []string{"/"},
				NetworkInterfaces: []string{"eth0", "en0"},
			},
			Performance: PerformanceConfig{
				SystemSampleRate:   1,
				BusinessSampleRate: 1,
				MaxHistorySize:     24,
				CollectionDelay:    2 * time.Second,
				InterFuncDelay:     10 * time.Minute,
				InnerFuncDelay:     5 * time.Second,
			},
			Redis: RedisConfig{
				Enabled:  true,
				Addr:     "redis:6379",
				Password: "",
				DB:       0,
			},
		},
		Teamspeak: TeamspeakConfig{
			Host:                    "127.0.0.1",
			QueryPort:               10011,
			VirtualServerPort:       9987,
			VirtualServerID:         1,
			Username:                "serveradmin",
			Password:                "",
			Nickname:                "Monitoring Bot",
			ReconnectMaxRetries:     3,
			ReconnectInitialBackoff: 2 * time.Second,
			ReconnectMaxBackoff:     60 * time.Second,
		},
		Security: SecurityConfig{
			JWTSecret:          "",
			PasswordSaltRounds: 10,
		},
		Deployment: DeploymentConfig{
			ScriptDir: "deploy-scripts",
			Timeout:   10 * time.Minute,
		},
	}
}
