/**
 * 日志服务适配器
 * 文件作用：提供日志服务适配器，将logs.LogService转换为utils.LogServiceAdapter接口
 * 版本：v1.0.0
 */

package logs

// LogServiceAdapter 日志服务适配器，将logs.LogService转换为utils.LogServiceAdapter
// 实现适配器模式，统一不同模块的日志接口
type LogServiceAdapter struct {
	logService LogService
}

// NewLogServiceAdapter 创建日志服务适配器实例
// 参数：logService - 日志服务实例
// 返回值：*LogServiceAdapter - 适配器实例
func NewLogServiceAdapter(logService LogService) *LogServiceAdapter {
	return &LogServiceAdapter{
		logService: logService,
	}
}

// Info 记录信息级别日志
// 参数：module - 模块名称，message - 日志消息，fields - 键值对字段
func (l *LogServiceAdapter) Info(module, message string, fields ...interface{}) {
	logFields := make([]LogField, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			logFields = append(logFields, LogField{Key: key, Value: fields[i+1]})
		}
	}
	l.logService.Info(module, message, logFields...)
}

// Warn 记录警告级别日志
// 参数：module - 模块名称，message - 日志消息，fields - 键值对字段
func (l *LogServiceAdapter) Warn(module, message string, fields ...interface{}) {
	logFields := make([]LogField, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			logFields = append(logFields, LogField{Key: key, Value: fields[i+1]})
		}
	}
	l.logService.Warn(module, message, logFields...)
}

// Error 记录错误级别日志
// 参数：module - 模块名称，message - 日志消息，fields - 键值对字段
func (l *LogServiceAdapter) Error(module, message string, fields ...interface{}) {
	logFields := make([]LogField, 0, len(fields)/2)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			logFields = append(logFields, LogField{Key: key, Value: fields[i+1]})
		}
	}
	l.logService.Error(module, message, logFields...)
}
