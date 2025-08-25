package instance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	// AlertLevelInfo 信息级别
	AlertLevelInfo AlertLevel = "info"
	// AlertLevelWarning 警告级别
	AlertLevelWarning AlertLevel = "warning"
	// AlertLevelError 错误级别
	AlertLevelError AlertLevel = "error"
	// AlertLevelCritical 严重级别
	AlertLevelCritical AlertLevel = "critical"
)

// AlertType 告警类型
type AlertType string

const (
	// AlertTypeProcess 进程相关告警
	AlertTypeProcess AlertType = "process"
	// AlertTypeResource 资源相关告警
	AlertTypeResource AlertType = "resource"
	// AlertTypeHealth 健康检查告警
	AlertTypeHealth AlertType = "health"
)

// Alert 告警信息
type Alert struct {
	ID             string     `json:"id" gorm:"primaryKey;size:36"`
	InstanceID     string     `json:"instance_id" gorm:"size:36;index"`
	Level          AlertLevel `json:"level" gorm:"size:20;index"`
	Type           AlertType  `json:"type" gorm:"size:20;index"`
	Title          string     `json:"title" gorm:"size:255"`
	Message        string     `json:"message" gorm:"type:text"`
	Details        string     `json:"details,omitempty" gorm:"type:json"`
	Status         string     `json:"status" gorm:"size:20;default:'active'"` // active, resolved, acknowledged
	AcknowledgedBy string     `json:"acknowledged_by" gorm:"size:36;index"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Alert) TableName() string {
	return "ts_alerts"
}

// BeforeCreate 创建前的钩子
func (a *Alert) BeforeCreate(tx *gorm.DB) (err error) {
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	if a.Status == "" {
		a.Status = "active"
	}
	return
}

// BeforeUpdate 更新前的钩子
func (a *Alert) BeforeUpdate(tx *gorm.DB) (err error) {
	a.UpdatedAt = time.Now()
	return
}

// AlertConfig 告警配置
type AlertConfig struct {
	ID         string    `json:"id" gorm:"primaryKey;size:36"`
	InstanceID string    `json:"instance_id" gorm:"size:36;uniqueIndex"`
	Enabled    bool      `json:"enabled" gorm:"default:true"`
	Channels   string    `json:"channels" gorm:"type:json"`   // 告警通道配置
	Thresholds string    `json:"thresholds" gorm:"type:json"` // 告警阈值配置
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AlertConfig) TableName() string {
	return "ts_alert_configs"
}

// BeforeCreate 创建前的钩子
func (ac *AlertConfig) BeforeCreate(tx *gorm.DB) (err error) {
	ac.CreatedAt = time.Now()
	ac.UpdatedAt = time.Now()
	return
}

// BeforeUpdate 更新前的钩子
func (ac *AlertConfig) BeforeUpdate(tx *gorm.DB) (err error) {
	ac.UpdatedAt = time.Now()
	return
}

// AlertChannel 告警通道接口
type AlertChannel interface {
	Send(alert *Alert) error
}

// EmailAlertChannel 邮件告警通道
type EmailAlertChannel struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	From         string
	To           []string
}

// Send 发送邮件告警
func (c *EmailAlertChannel) Send(alert *Alert) error {
	// 实现邮件发送逻辑
	// 这里简化为记录日志
	log.Printf("[Email Alert] %s - %s: %s", alert.Level, alert.Title, alert.Message)
	return nil
}

// WebhookAlertChannel Webhook告警通道
type WebhookAlertChannel struct {
	URL     string
	Headers map[string]string
}

// Send 发送Webhook告警
func (c *WebhookAlertChannel) Send(alert *Alert) error {
	// 准备请求体
	body, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// AlertManager 告警管理器
type AlertManager struct {
	db         *gorm.DB
	channels   map[string]AlertChannel
	logService interface {
		Error(msg string, fields ...any)
		Warn(msg string, fields ...any)
		Info(msg string, fields ...any)
	}
}

// NewAlertManager 创建告警管理器
func NewAlertManager(db *gorm.DB, logService interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Info(msg string, fields ...any)
}) *AlertManager {
	return &AlertManager{
		db:         db,
		channels:   make(map[string]AlertChannel),
		logService: logService,
	}
}

// AddChannel 添加告警通道
func (m *AlertManager) AddChannel(name string, channel AlertChannel) {
	m.channels[name] = channel
}

// Trigger 触发告警
func (m *AlertManager) Trigger(instanceID string, level AlertLevel, alertType AlertType, title, message string, details any) error {
	// 获取实例的告警配置
	var config AlertConfig
	if err := m.db.First(&config, "instance_id = ?", instanceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有配置，使用默认配置
			config = AlertConfig{
				ID:         uuid.New().String(),
				InstanceID: instanceID,
				Enabled:    true,
			}
			if dbCreateErr := m.db.Create(&config).Error; dbCreateErr != nil {
				return fmt.Errorf("failed to create default alert config: %w", dbCreateErr)
			}
		} else {
			return fmt.Errorf("failed to get alert config: %w", err)
		}
	}

	// 如果告警被禁用，直接返回
	if !config.Enabled {
		return nil
	}

	// 创建告警记录
	var detailsStr string
	if details != nil {
		detailsBytes, err := json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
		detailsStr = string(detailsBytes)
	}

	alert := &Alert{
		ID:         uuid.New().String(),
		InstanceID: instanceID,
		Level:      level,
		Type:       alertType,
		Title:      title,
		Message:    message,
		Details:    detailsStr,
		Status:     "active",
	}

	// 保存告警记录
	if err := m.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	// 发送告警通知
	for name, channel := range m.channels {
		if err := channel.Send(alert); err != nil {
			if m.logService != nil {
				m.logService.Error("Failed to send alert via channel", "channel", name, "error", err)
			} else {
				log.Printf("Failed to send alert via channel %s: %v", name, err)
			}
			continue
		}
	}

	return nil
}

// Resolve 解决告警
func (m *AlertManager) Resolve(instanceID, alertType, resolvedBy string) error {
	// 查找未解决的告警
	var alerts []Alert
	if err := m.db.Where("instance_id = ? AND type = ? AND status = ?",
		instanceID, alertType, "active").Find(&alerts).Error; err != nil {
		return fmt.Errorf("failed to find active alerts: %w", err)
	}

	// 更新告警状态
	now := time.Now()
	for i := range alerts {
		alerts[i].Status = "resolved"
		alerts[i].ResolvedAt = &now
		alerts[i].AcknowledgedBy = resolvedBy
		alerts[i].AcknowledgedAt = &now

		if err := m.db.Save(&alerts[i]).Error; err != nil {
			if m.logService != nil {
				m.logService.Error("Failed to resolve alert", "alert_id", alerts[i].ID, "error", err)
			} else {
				log.Printf("Failed to resolve alert %s: %v", alerts[i].ID, err)
			}
		}
	}

	return nil
}

// GetAlerts 获取告警列表
func (m *AlertManager) GetAlerts(filter *AlertFilter) ([]*Alert, int64, error) {
	var alerts []*Alert
	var total int64

	tx := m.db.Model(&Alert{})

	// 应用过滤器
	if filter != nil {
		tx = filter.Apply(tx)
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count alerts: %w", err)
	}

	// 应用分页
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		tx = tx.Offset(offset).Limit(filter.PageSize)
	}

	// 获取数据
	if err := tx.Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find alerts: %w", err)
	}

	return alerts, total, nil
}

// GetAlert 根据ID获取告警
func (m *AlertManager) GetAlert(id string) (*Alert, error) {
	var alert Alert
	if err := m.db.First(&alert, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

// AlertFilter 告警过滤器
type AlertFilter struct {
	InstanceID string     `form:"instance_id"`
	Level      AlertLevel `form:"level"`
	Type       AlertType  `form:"type"`
	Status     string     `form:"status"`
	StartTime  *time.Time `form:"start_time"`
	EndTime    *time.Time `form:"end_time"`
	Page       int        `form:"page,default=1"`
	PageSize   int        `form:"page_size,default=20"`
}

// Apply 应用过滤器
func (f *AlertFilter) Apply(tx *gorm.DB) *gorm.DB {
	if f.InstanceID != "" {
		tx = tx.Where("instance_id = ?", f.InstanceID)
	}

	if f.Level != "" {
		tx = tx.Where("level = ?", f.Level)
	}

	if f.Type != "" {
		tx = tx.Where("type = ?", f.Type)
	}

	if f.Status != "" {
		tx = tx.Where("status = ?", f.Status)
	}

	if f.StartTime != nil {
		tx = tx.Where("created_at >= ?", f.StartTime)
	}

	if f.EndTime != nil {
		tx = tx.Where("created_at <= ?", f.EndTime)
	}

	// 默认按创建时间倒序
	tx = tx.Order("created_at DESC")

	return tx
}
