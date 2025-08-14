package instance

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InstanceStatus 实例状态类型
type InstanceStatus string

const (
	// StatusStopped 实例已停止
	StatusStopped InstanceStatus = "stopped"
	// StatusStarting 实例启动中
	StatusStarting InstanceStatus = "starting"
	// StatusRunning 实例运行中
	StatusRunning InstanceStatus = "running"
	// StatusStopping 实例停止中
	StatusStopping InstanceStatus = "stopping"
	// StatusError 实例错误
	StatusError InstanceStatus = "error"
)

// InstanceConfig 实例配置
// 包含 TeamSpeak 服务器的配置选项
type InstanceConfig struct {
	// 服务器配置
	ServerName    string `json:"server_name" gorm:"size:255"`
	WelcomeMsg    string `json:"welcome_msg" gorm:"type:text"`
	MaxClients    int    `json:"max_clients" gorm:"default:32"`
	DefaultServer int    `json:"default_server" gorm:"default:1"`

	// 网络配置
	VoicePort  int `json:"voice_port" gorm:"default:9987"`
	FilePort   int `json:"file_port" gorm:"default:30033"`
	QueryPort  int `json:"query_port" gorm:"default:10011"`
	ServerPort int `json:"server_port" gorm:"default:2010"`

	// 认证配置
	QueryAdminPassword string `json:"query_admin_password" gorm:"size:255"`
	ServerAdminToken   string `json:"server_admin_token" gorm:"size:255"`

	// 高级配置
	LogQueries    bool `json:"log_queries" gorm:"default:false"`
	LogClientCmds bool `json:"log_client_cmds" gorm:"default:false"`
}

// Instance 表示一个 TeamSpeak 服务器实例
type Instance struct {
	ID        string         `json:"id" gorm:"primaryKey;size:36"`
	ProcessID int32          `json:"process_id" gorm:"default:0"`
	Name      string         `json:"name" gorm:"size:100;not null;index"`
	Status    InstanceStatus `json:"status" gorm:"size:20;index"`
	Version   string         `json:"version" gorm:"size:50"`
	Config    InstanceConfig `json:"config" gorm:"serializer:json"`
	Host      string         `json:"host" gorm:"size:255"`
	CreatedBy uint           `json:"created_by" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Instance) TableName() string {
	return "ts_instances"
}

// BeforeCreate 创建前的钩子
func (i *Instance) BeforeCreate(tx *gorm.DB) (err error) {
	i.CreatedAt = time.Now()
	i.UpdatedAt = time.Now()
	if i.Status == "" {
		i.Status = StatusStopped
	}
	return
}

// BeforeUpdate 更新前的钩子
func (i *Instance) BeforeUpdate(tx *gorm.DB) (err error) {
	i.UpdatedAt = time.Now()
	return
}

// SetStatus 设置实例状态
func (i *Instance) SetStatus(status InstanceStatus) error {
	switch status {
	case StatusStopped, StatusStarting, StatusRunning, StatusStopping, StatusError:
		i.Status = status
		i.UpdatedAt = time.Now()
		return nil
	default:
		return fmt.Errorf("invalid status: %s", status)
	}
}

// GetConnectionString 获取实例连接字符串
func (i *Instance) GetConnectionString() string {
	return fmt.Sprintf("%s:%d", i.Host, i.Config.VoicePort)
}

// IsRunning 检查实例是否在运行
func (i *Instance) IsRunning() bool {
	return i.Status == StatusRunning
}

// CanStart 检查实例是否可以启动
func (i *Instance) CanStart() bool {
	return i.Status == StatusStopped || i.Status == StatusError
}

// CanStop 检查实例是否可以停止
func (i *Instance) CanStop() bool {
	return i.Status == StatusRunning
}

// CanRestart 检查实例是否可以重启
func (i *Instance) CanRestart() bool {
	return i.Status == StatusRunning || i.Status == StatusError
}

// InstanceLog 实例日志记录
type InstanceLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	InstanceID string    `json:"instance_id" gorm:"size:36;index"`
	Level      string    `json:"level" gorm:"size:20;index"` // info, warning, error
	Message    string    `json:"message" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`
}

// TableName 指定日志表名
func (InstanceLog) TableName() string {
	return "ts_instance_logs"
}

// AddLog 添加日志记录
func (i *Instance) AddLog(tx *gorm.DB, level, message string) error {
	logEntry := InstanceLog{
		InstanceID: i.ID,
		Level:      level,
		Message:    message,
		CreatedAt:  time.Now(),
	}
	return tx.Create(&logEntry).Error
}

// GetLogs 获取实例日志
func (i *Instance) GetLogs(tx *gorm.DB, limit int) ([]InstanceLog, error) {
	var logs []InstanceLog
	err := tx.Where("instance_id = ?", i.ID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
