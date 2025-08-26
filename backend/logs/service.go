/**
 * 系统日志服务层
 * 文件作用：提供统一的日志记录服务，集成zap高性能日志库
 * 作者：AI Assistant
 * 版本：v1.0.0
 */

package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// LogService 日志服务接口
type LogService interface {
	// Debug 记录调试级别日志
	Debug(module, message string, fields ...LogField)

	// Info 记录信息级别日志
	Info(module, message string, fields ...LogField)

	// Warn 记录警告级别日志
	Warn(module, message string, fields ...LogField)

	// Error 记录错误级别日志
	Error(module, message string, fields ...LogField)

	// WithUser 设置用户上下文
	WithUser(uid uint) LogService

	// WithIP 设置IP地址上下文
	WithIP(ipAddress string) LogService

	// Query 查询日志
	Query(params LogQueryParams) ([]SystemLog, int64, error)

	// GetStats 获取日志统计
	GetStats() (LogStats, error)

	// Cleanup 清理过期日志
	Cleanup(retentionDays int) (int64, error)

	// Close 关闭日志服务
	Close() error
}

// LogField 日志字段
type LogField struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别
	Level string `yaml:"level" json:"level"`

	// 是否启用数据库存储
	EnableDB bool `yaml:"enable_db" json:"enable_db"`

	// 日志保留天数
	RetentionDays int `yaml:"retention_days" json:"retention_days"`

	// 批量写入大小
	BatchSize int `yaml:"batch_size" json:"batch_size"`

	// 批量写入间隔(秒)
	BatchInterval int `yaml:"batch_interval" json:"batch_interval"`

	// 是否启用文件日志
	EnableFile bool `yaml:"enable_file" json:"enable_file"`

	// 日志文件路径
	FilePath string `yaml:"file_path" json:"file_path"`
}

// logServiceImpl 日志服务实现
type logServiceImpl struct {
	db        *gorm.DB
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	config    LogConfig

	// 上下文信息
	uid       uint
	ipAddress string

	// 批量写入相关
	logBuffer []SystemLog
	bufferMu  sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewLogService 创建日志服务实例
// 参数：db - 数据库连接，config - 日志配置
// 返回值：LogService - 日志服务实例，error - 创建错误
func NewLogService(db *gorm.DB, config LogConfig) (LogService, error) {
	// 设置默认配置
	if config.Level == "" {
		config.Level = "info"
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = 30
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchInterval <= 0 {
		config.BatchInterval = 5
	}

	// 创建zap日志配置
	zapConfig := zap.NewProductionConfig()

	// 配置时间戳格式为毫秒级
	zapConfig.EncoderConfig.TimeKey = "ts"
	zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(t.UnixMilli()) // 使用毫秒级时间戳
	}

	// 设置日志级别
	switch config.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// 配置输出路径
	if config.EnableFile && config.FilePath != "" {
		// 确保日志文件目录存在
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		zapConfig.OutputPaths = []string{"stdout", config.FilePath}
		zapConfig.ErrorOutputPaths = []string{"stderr", config.FilePath}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	// 创建zap logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &logServiceImpl{
		db:        db,
		zapLogger: zapLogger,
		sugar:     zapLogger.Sugar(),
		config:    config,
		logBuffer: make([]SystemLog, 0, config.BatchSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	// 启动批量写入协程
	if config.EnableDB {
		go service.batchWriter()
	}

	return service, nil
}

// Debug 记录调试级别日志
func (s *logServiceImpl) Debug(module, message string, fields ...LogField) {
	s.log(LogLevelDebug, module, message, fields...)
}

// Info 记录信息级别日志
func (s *logServiceImpl) Info(module, message string, fields ...LogField) {
	s.log(LogLevelInfo, module, message, fields...)
}

// Warn 记录警告级别日志
func (s *logServiceImpl) Warn(module, message string, fields ...LogField) {
	s.log(LogLevelWarn, module, message, fields...)
}

// Error 记录错误级别日志
func (s *logServiceImpl) Error(module, message string, fields ...LogField) {
	s.log(LogLevelError, module, message, fields...)
}

// WithUser 设置用户上下文
func (s *logServiceImpl) WithUser(uid uint) LogService {
	newService := &logServiceImpl{
		db:        s.db,
		zapLogger: s.zapLogger,
		sugar:     s.sugar,
		config:    s.config,
		logBuffer: make([]SystemLog, 0, s.config.BatchSize),
		ctx:       s.ctx,
		cancel:    s.cancel,
	}
	newService.uid = uid
	return newService
}

// WithIP 设置IP地址上下文
func (s *logServiceImpl) WithIP(ipAddress string) LogService {
	newService := &logServiceImpl{
		db:        s.db,
		zapLogger: s.zapLogger,
		sugar:     s.sugar,
		config:    s.config,
		logBuffer: make([]SystemLog, 0, s.config.BatchSize),
		ctx:       s.ctx,
		cancel:    s.cancel,
	}
	newService.ipAddress = ipAddress
	return newService
}

// Query 查询日志
func (s *logServiceImpl) Query(params LogQueryParams) ([]SystemLog, int64, error) {
	return GetLogs(s.db, params)
}

// GetStats 获取日志统计
func (s *logServiceImpl) GetStats() (LogStats, error) {
	return GetLogStats(s.db)
}

// Cleanup 清理过期日志
func (s *logServiceImpl) Cleanup(retentionDays int) (int64, error) {
	return CleanupLogs(s.db, retentionDays)
}

// Close 关闭日志服务
func (s *logServiceImpl) Close() error {
	s.cancel()

	// 等待批量写入完成
	time.Sleep(time.Duration(s.config.BatchInterval+1) * time.Second)

	// 写入剩余日志
	s.flushBuffer()

	return s.zapLogger.Sync()
}

// log 内部日志记录方法
func (s *logServiceImpl) log(level LogLevel, module, message string, fields ...LogField) {
	// 构建zap字段
	zapFields := make([]zap.Field, 0, len(fields)+3)
	zapFields = append(zapFields, zap.String("module", module))

	if s.uid > 0 {
		zapFields = append(zapFields, zap.Uint("user_id", s.uid))
	}

	if s.ipAddress != "" {
		zapFields = append(zapFields, zap.String("ip_address", s.ipAddress))
	}

	// 添加自定义字段
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}

	// 记录到zap日志
	switch level {
	case LogLevelDebug:
		s.zapLogger.Debug(message, zapFields...)
	case LogLevelInfo:
		s.zapLogger.Info(message, zapFields...)
	case LogLevelWarn:
		s.zapLogger.Warn(message, zapFields...)
	case LogLevelError:
		s.zapLogger.Error(message, zapFields...)
	}

	// 如果启用数据库存储，添加到缓冲区
	if s.config.EnableDB {
		systemLog := SystemLog{
			Level:     level,
			Module:    module,
			Uid:       s.uid,
			IPAddress: s.ipAddress,
			Message:   message,
			CreatedAt: time.Now(),
		}

		// 设置元数据
		if len(fields) > 0 {
			metadata := make(map[string]any)
			for _, field := range fields {
				metadata[field.Key] = field.Value
			}
			systemLog.SetMetadata(metadata)
		}

		s.addToBuffer(systemLog)
	}
}

// addToBuffer 添加日志到缓冲区
func (s *logServiceImpl) addToBuffer(log SystemLog) {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()

	s.logBuffer = append(s.logBuffer, log)

	// 如果缓冲区满了，立即写入（不使用goroutine避免并发问题）
	if len(s.logBuffer) >= s.config.BatchSize {
		s.flushBufferUnsafe()
	}
}

// batchWriter 批量写入协程
func (s *logServiceImpl) batchWriter() {
	ticker := time.NewTicker(time.Duration(s.config.BatchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.flushBuffer()
		}
	}
}

// flushBuffer 刷新缓冲区到数据库（带锁）
func (s *logServiceImpl) flushBuffer() {
	s.bufferMu.Lock()
	defer s.bufferMu.Unlock()
	s.flushBufferUnsafe()
}

// flushBufferUnsafe 刷新缓冲区到数据库（不加锁，调用方需要持有锁）
func (s *logServiceImpl) flushBufferUnsafe() {
	if len(s.logBuffer) == 0 {
		return
	}

	// 检查数据库连接是否有效
	if s.db == nil {
		s.sugar.Error("Database connection is nil, cannot flush logs")
		return
	}

	// 创建缓冲区副本以避免并发修改
	logs := make([]SystemLog, len(s.logBuffer))
	copy(logs, s.logBuffer)

	// 清空缓冲区
	s.logBuffer = s.logBuffer[:0]

	// 批量插入到数据库
	if err := s.db.CreateInBatches(logs, s.config.BatchSize).Error; err != nil {
		// 如果批量插入失败，记录到zap日志
		s.sugar.Errorf("Failed to batch insert logs: %v", err)
	}
}

// 全局日志服务实例
var (
	globalLogService LogService
	globalLogMutex   sync.RWMutex
)

// SetGlobalLogService 设置全局日志服务
func SetGlobalLogService(service LogService) {
	globalLogMutex.Lock()
	defer globalLogMutex.Unlock()
	globalLogService = service
}

// GetGlobalLogService 获取全局日志服务
func GetGlobalLogService() LogService {
	globalLogMutex.RLock()
	defer globalLogMutex.RUnlock()
	return globalLogService
}

// 便捷方法

// Debug 全局调试日志
func Debug(module, message string, fields ...LogField) {
	if service := GetGlobalLogService(); service != nil {
		service.Debug(module, message, fields...)
	}
}

// Info 全局信息日志
func Info(module, message string, fields ...LogField) {
	if service := GetGlobalLogService(); service != nil {
		service.Info(module, message, fields...)
	}
}

// Warn 全局警告日志
func Warn(module, message string, fields ...LogField) {
	if service := GetGlobalLogService(); service != nil {
		service.Warn(module, message, fields...)
	}
}

// Error 全局错误日志
func Error(module, message string, fields ...LogField) {
	if service := GetGlobalLogService(); service != nil {
		service.Error(module, message, fields...)
	}
}

// WithUser 全局用户上下文日志
func WithUser(uid uint) LogService {
	if service := GetGlobalLogService(); service != nil {
		return service.WithUser(uid)
	}
	return nil
}

// WithIP 全局IP上下文日志
func WithIP(ipAddress string) LogService {
	if service := GetGlobalLogService(); service != nil {
		return service.WithIP(ipAddress)
	}
	return nil
}
