package instance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"teamspeak-one-click-deploy/logs"
	"teamspeak-one-click-deploy/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrInstanceNotFound 实例未找到
	ErrInstanceNotFound = errors.New("instance not found")
	// ErrInvalidStatus 无效的状态
	ErrInvalidStatus = errors.New("invalid status")
	// ErrOperationNotAllowed 不允许的操作
	ErrOperationNotAllowed = errors.New("operation not allowed")
)

// SetLogService 设置日志服务
func (s *Service) SetLogService(logService logs.LogService) {
	s.logService = logService
}

// Service 实例服务
// 提供对 TeamSpeak 实例的 CRUD 操作和状态管理
type Service struct {
	db             *gorm.DB
	healthMonitor  *HealthMonitor
	restartManager *RestartManager
	alertManager   *AlertManager
	logService     logs.LogService
}

// CreateInstanceInput 创建实例的输入参数
type CreateInstanceInput struct {
	Name               string `json:"name" validate:"required,min=3,max=100"`
	Version            string `json:"version" validate:"required"`
	Host               string `json:"host" validate:"required,hostname|ip"`
	ServerName         string `json:"server_name" validate:"required,min=3,max=255"`
	WelcomeMsg         string `json:"welcome_msg"`
	MaxClients         int    `json:"max_clients" validate:"min=1,max=1024"`
	VoicePort          int    `json:"voice_port" validate:"min=1,max=65535"`
	FilePort           int    `json:"file_port" validate:"min=1,max=65535"`
	QueryPort          int    `json:"query_port" validate:"min=1,max=65535"`
	ServerPort         int    `json:"server_port" validate:"min=1,max=65535"`
	QueryAdminPassword string `json:"query_admin_password" validate:"required,min=8"`
	ServerAdminToken   string `json:"server_admin_token"`
	LogQueries         bool   `json:"log_queries"`
	LogClientCmds      bool   `json:"log_client_cmds"`
}

// UpdateInstanceInput 更新实例的输入参数
type UpdateInstanceInput struct {
	Name       string `json:"name" validate:"omitempty,min=3,max=100"`
	Version    string `json:"version" validate:"omitempty"`
	ServerName string `json:"server_name" validate:"omitempty,min=3,max=255"`
	WelcomeMsg string `json:"welcome_msg"`
	MaxClients int    `json:"max_clients" validate:"omitempty,min=1,max=1024"`
}

// InstanceFilter 实例查询过滤器
type InstanceFilter struct {
	Name     string `form:"name"`
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=10"`
}

// RestartConfig 重启配置
type RestartConfig struct {
	ID            string    `gorm:"primary_key"`
	RestartCount  int       `gorm:"default:0"`
	LastRestartAt time.Time `gorm:"default:NULL"`
}

// NewService 创建新的实例服务
func NewService(db *gorm.DB, alertManager *AlertManager) *Service {
	// 创建日志服务
	logConfig := logs.LogConfig{
		Level:         "info",
		EnableDB:      true,
		RetentionDays: 30,
		BatchSize:     100,
		BatchInterval: 60,
		EnableFile:    true,
		FilePath:      "../.logs/instance.log",
	}
	logService, err := logs.NewLogService(db, logConfig)
	if err != nil {
		// 如果日志服务创建失败，使用标准日志记录
		log.Printf("Failed to create log service: %v", err)
	}

	service := &Service{
		db:             db,
		healthMonitor:  NewHealthMonitor(db),
		restartManager: NewRestartManager(db),
		alertManager:   alertManager,
		logService:     logService,
	}

	// 添加默认的健康检查器
	service.healthMonitor.AddChecker(&ProcessHealthChecker{})

	// 添加资源健康检查器（使用默认资源限制）
	resourceLimits := &ResourceLimits{
		MaxCPUPercent:   80.0,
		MaxMemoryMB:     2048,  // 2GB
		MaxDiskUsageMB:  10240, // 10GB
		MaxNetworkInMB:  1024,  // 1GB/hour
		MaxNetworkOutMB: 1024,  // 1GB/hour
	}
	resourceChecker := &ResourceHealthChecker{
		db:           service.db,
		Limits:       resourceLimits,
		AlertManager: service.alertManager,
	}
	service.healthMonitor.AddChecker(resourceChecker)

	// 启动健康监控
	go service.healthMonitor.Start()

	// 加载并监控所有运行中的实例
	go service.monitorAllInstances()

	return service
}

// monitorAllInstances 加载并监控所有运行中的实例
func (s *Service) monitorAllInstances() {
	// 等待数据库连接就绪
	time.Sleep(2 * time.Second)

	var instances []Instance
	if err := s.db.Where("status IN (?)",
		[]InstanceStatus{StatusStarting, StatusRunning}).Find(&instances).Error; err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "Failed to load running instances", logs.LogField{Key: "error", Value: err})
		} else {
			log.Printf("Failed to load running instances: %v", err)
		}
		return
	}

	for _, instance := range instances {
		s.healthMonitor.AddInstance(instance.ID)
		if s.logService != nil {
			s.logService.Info("instance", "Added instance to health monitoring", logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("Added instance %s to health monitoring", instance.ID)
		}
	}
}

// CreateInstance 创建新的 TeamSpeak 实例
func (s *Service) CreateInstance(ctx context.Context, input *CreateInstanceInput) (*Instance, error) {
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user ID not found in context")
	}

	// 验证输入
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 生成唯一ID
	id := uuid.New().String()

	// 创建实例
	instance := &Instance{
		ID:        id,
		Name:      input.Name,
		Status:    StatusStopped,
		Version:   input.Version,
		Host:      input.Host,
		CreatedBy: userID,
		Config: InstanceConfig{
			ServerName:         input.ServerName,
			WelcomeMsg:         input.WelcomeMsg,
			MaxClients:         input.MaxClients,
			VoicePort:          input.VoicePort,
			FilePort:           input.FilePort,
			QueryPort:          input.QueryPort,
			ServerPort:         input.ServerPort,
			QueryAdminPassword: input.QueryAdminPassword,
			ServerAdminToken:   input.ServerAdminToken,
			LogQueries:         input.LogQueries,
			LogClientCmds:      input.LogClientCmds,
		},
	}

	// 开始事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 保存实例
		if err := tx.Create(instance).Error; err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}

		// 添加创建日志
		logMsg := fmt.Sprintf("Instance %s created by user %d", instance.Name, userID)
		return instance.AddLog(tx, "info", logMsg)
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// GetInstance 获取实例详情
func (s *Service) GetInstance(ctx context.Context, id string) (*Instance, error) {
	var instance Instance
	err := s.db.First(&instance, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInstanceNotFound
		}
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	return &instance, nil
}

// ListInstances 获取实例列表
func (s *Service) ListInstances(ctx context.Context, filter *InstanceFilter) ([]*Instance, int64, error) {
	var instances []*Instance
	var total int64

	tx := s.db.Model(&Instance{})

	// 应用过滤器
	if filter != nil {
		tx = filter.Apply(tx)
	}

	// 获取总数
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count instances: %w", err)
	}

	// 应用分页
	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		tx = tx.Offset(offset).Limit(filter.PageSize)
	}

	// 获取数据
	if err := tx.Find(&instances).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list instances: %w", err)
	}

	return instances, total, nil
}

// UpdateInstance 更新实例配置
func (s *Service) UpdateInstance(ctx context.Context, id string, input *UpdateInstanceInput) (*Instance, error) {
	// 获取当前用户ID
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return nil, errors.New("user ID not found in context")
	}

	// 验证输入
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if input.Name != "" {
		instance.Name = input.Name
	}
	if input.Version != "" {
		instance.Version = input.Version
	}
	if input.ServerName != "" {
		instance.Config.ServerName = input.ServerName
	}
	if input.WelcomeMsg != "" {
		instance.Config.WelcomeMsg = input.WelcomeMsg
	}
	if input.MaxClients > 0 {
		instance.Config.MaxClients = input.MaxClients
	}

	// 更新数据库
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if dberr := tx.Save(instance).Error; dberr != nil {
			return fmt.Errorf("failed to update instance: %w", dberr)
		}

		// 添加更新日志
		logMsg := fmt.Sprintf("Instance %s updated by user %d", instance.Name, userID)
		return instance.AddLog(tx, "info", logMsg)
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// DeleteInstance 删除实例
func (s *Service) DeleteInstance(ctx context.Context, id string) error {
	// 获取当前用户ID
	userID, ok := utils.GetUserIDFromContext(ctx)
	if !ok {
		return errors.New("user ID not found in context")
	}

	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	// 检查实例状态
	if instance.IsRunning() {
		return fmt.Errorf("cannot delete running instance: %w", ErrOperationNotAllowed)
	}

	// 执行删除
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除实例
		if err := tx.Delete(instance).Error; err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}

		// 添加删除日志
		logMsg := fmt.Sprintf("Instance %s deleted by user %d", instance.Name, userID)
		return instance.AddLog(tx, "info", logMsg)
	})
}

// StartInstance 启动实例
func (s *Service) StartInstance(ctx context.Context, id string) error {
	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否可以启动
	if !instance.CanStart() {
		return fmt.Errorf("cannot start instance in status %s: %w", instance.Status, ErrOperationNotAllowed)
	}

	// 更新状态为启动中
	if err := instance.SetStatus(StatusStarting); err != nil {
		return fmt.Errorf("failed to set instance status: %w", err)
	}

	// 保存状态
	if err := s.db.Save(instance).Error; err != nil {
		return fmt.Errorf("failed to save instance status: %w", err)
	}

	// 异步启动实例
	go s.startInstanceProcess(instance)

	return nil
}

// StopInstance 停止实例
func (s *Service) StopInstance(ctx context.Context, id string) error {
	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否可以停止
	if !instance.CanStop() {
		return fmt.Errorf("cannot stop instance in status %s: %w", instance.Status, ErrOperationNotAllowed)
	}

	// 更新状态为停止中
	if err := instance.SetStatus(StatusStopping); err != nil {
		return fmt.Errorf("failed to set instance status: %w", err)
	}

	// 保存状态
	if err := s.db.Save(instance).Error; err != nil {
		return fmt.Errorf("failed to save instance status: %w", err)
	}

	// 异步停止实例
	go s.stopInstance(instance)

	return nil
}

// RestartInstance 重启实例
func (s *Service) RestartInstance(ctx context.Context, id string) error {
	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return err
	}

	// 检查是否可以重启
	if !instance.CanRestart() {
		return fmt.Errorf("cannot restart instance in status %s: %w", instance.Status, ErrOperationNotAllowed)
	}

	// 异步重启实例
	go s.restartInstanceProcess(instance)

	return nil
}

// startInstanceProcess 异步启动实例的实际实现
func (s *Service) startInstanceProcess(instance *Instance) {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 准备环境
		if err := s.prepareInstanceEnvironment(instance); err != nil {
			instance.AddLog(tx, "error", fmt.Sprintf("准备环境失败: %v", err))
			instance.SetStatus(StatusError)
			if txInitErr := tx.Save(instance).Error; txInitErr != nil {
				if s.logService != nil {
					s.logService.Error("instance", "更新实例状态失败", logs.LogField{Key: "error", Value: txInitErr}, logs.LogField{Key: "instance_id", Value: instance.ID})
				} else {
					log.Printf("更新实例状态失败: %v", txInitErr)
				}
			}
			return err
		}

		// 2. 构建启动命令
		cmd, err := s.buildStartCommand(instance)
		if err != nil {
			instance.AddLog(tx, "error", fmt.Sprintf("构建启动命令失败: %v", err))
			instance.SetStatus(StatusError)
			if txBuildErr := tx.Save(instance).Error; txBuildErr != nil {
				if s.logService != nil {
					s.logService.Error("instance", "更新实例状态失败", logs.LogField{Key: "error", Value: txBuildErr}, logs.LogField{Key: "instance_id", Value: instance.ID})
				} else {
					log.Printf("更新实例状态失败: %v", txBuildErr)
				}
			}
			return err
		}

		// 3. 启动进程
		if err := cmd.Start(); err != nil {
			instance.AddLog(tx, "error", fmt.Sprintf("启动进程失败: %v", err))
			instance.SetStatus(StatusError)
			if txSaveErr := tx.Save(instance).Error; txSaveErr != nil {
				if s.logService != nil {
					s.logService.Error("instance", "更新实例状态失败", logs.LogField{Key: "error", Value: txSaveErr}, logs.LogField{Key: "instance_id", Value: instance.ID})
				} else {
					log.Printf("更新实例状态失败: %v", txSaveErr)
				}
			}
			return err
		}

		// 4. 更新状态
		if err := instance.SetStatus(StatusRunning); err != nil {
			instance.AddLog(tx, "error", fmt.Sprintf("更新状态失败: %v", err))
			return err
		}

		// 5. 保存进程ID
		instance.ProcessID = int32(cmd.Process.Pid)
		if err := tx.Save(instance).Error; err != nil {
			instance.AddLog(tx, "error", fmt.Sprintf("保存进程ID失败: %v", err))
			return err
		}

		// 6. 添加启动成功日志
		if err := instance.AddLog(tx, "info", "TeamSpeak 服务器启动成功"); err != nil {
			if s.logService != nil {
				s.logService.Error("instance", "添加启动日志失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
			} else {
				log.Printf("添加启动日志失败: %v", err)
			}
		}

		// 7. 启动监控协程
		go s.monitorInstance(instance, cmd)

		return nil
	})

	if err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "启动实例失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("启动实例 %s 失败: %v", instance.ID, err)
		}
	}
}

// prepareInstanceEnvironment 准备实例运行环境
func (s *Service) prepareInstanceEnvironment(instance *Instance) error {
	// 0. 验证输入
	if instance == nil {
		if s.logService != nil {
			s.logService.Error("instance", "错误: 实例对象为空")
		} else {
			log.Printf("错误: 实例对象为空")
		}
		return &utils.ValidationError{
			Code:    utils.ErrTSInstanceIsNil,
			Message: "实例对象不能为空",
		}
	}

	if s.logService != nil {
		s.logService.Info("instance", "开始准备实例运行环境", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("开始准备实例 %s 的运行环境", instance.ID)
	}

	// 1. 创建实例目录
	instanceDir := filepath.Join("/var/lib/teamspeak", instance.ID)
	if s.logService != nil {
		s.logService.Info("instance", "创建实例目录", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "directory", Value: instanceDir})
	} else {
		log.Printf("创建实例目录: %s", instanceDir)
	}

	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "创建实例目录失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "directory", Value: instanceDir})
		} else {
			log.Printf("创建实例目录失败: %v", err)
		}
		return &utils.ValidationError{
			Code:    utils.ErrTSInstanceDirEmpty,
			Message: fmt.Sprintf("创建实例目录失败: %v", err),
		}
	}

	// 2. 验证端口是否被占用
	if s.logService != nil {
		s.logService.Info("instance", "检查端口可用性", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("检查端口可用性...")
	}
	if err := s.checkPortsAvailability(instance); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "端口检查失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("端口检查失败: %v", err)
		}
		return &utils.ValidationError{
			Code:    utils.ErrTSInstancePortConflict,
			Message: err.Error(),
		}
	}

	// 3. 生成配置文件
	configPath := filepath.Join(instanceDir, "ts3server.ini")
	if s.logService != nil {
		s.logService.Info("instance", "生成配置文件", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "config_path", Value: configPath})
	} else {
		log.Printf("生成配置文件: %s", configPath)
	}

	if err := s.generateConfigFile(instance, configPath); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "生成配置文件失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("生成配置文件失败: %v", err)
		}
		return fmt.Errorf("生成配置文件失败: %w", err)
	}

	// 4. 设置目录权限和所有者
	if s.logService != nil {
		s.logService.Info("instance", "设置目录权限和所有者", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("设置目录权限和所有者...")
	}

	if err := os.Chmod(instanceDir, 0755); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置目录权限失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("警告: 设置目录权限失败: %v", err)
		}
	}

	if err := os.Chown(instanceDir, 1000, 1000); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置目录所有者失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("警告: 设置目录所有者失败: %v", err)
		}
	}

	// 5. 创建数据目录
	dataDir := filepath.Join(instanceDir, "data")
	if s.logService != nil {
		s.logService.Info("instance", "创建数据目录", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "directory", Value: dataDir})
	} else {
		log.Printf("创建数据目录: %s", dataDir)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "创建数据目录失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "directory", Value: dataDir})
		} else {
			log.Printf("创建数据目录失败: %v", err)
		}
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 6. 设置数据目录权限和所有者
	if err := os.Chmod(dataDir, 0755); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置数据目录权限失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("警告: 设置数据目录权限失败: %v", err)
		}
	}

	if err := os.Chown(dataDir, 1000, 1000); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置数据目录所有者失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("警告: 设置数据目录所有者失败: %v", err)
		}
	}

	if s.logService != nil {
		s.logService.Info("instance", "实例运行环境准备完成", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("实例 %s 运行环境准备完成", instance.ID)
	}
	return nil
}

// checkPortsAvailability 检查端口是否可用
func (s *Service) checkPortsAvailability(instance *Instance) error {
	ports := map[string]int{
		"语音端口":   instance.Config.VoicePort,
		"文件传输端口": instance.Config.FilePort,
		"查询端口":   instance.Config.QueryPort,
		"服务器端口":  instance.Config.ServerPort,
	}

	for name, port := range ports {
		if port <= 1024 || port > 65535 {
			return fmt.Errorf("%s %d 超出允许范围 (1025-65535)", name, port)
		}

		// 检查端口是否被占用
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("%s %d 已被占用: %v", name, port, err)
		}
		listener.Close()

		// 检查UDP端口
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return fmt.Errorf("解析UDP地址失败: %v", err)
		}
		udpConn, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			return fmt.Errorf("%s %d (UDP) 已被占用: %v", name, port, err)
		}
		udpConn.Close()
	}

	return nil
}

// generateConfigFile 生成 TeamSpeak 服务器配置文件
func (s *Service) generateConfigFile(instance *Instance, configPath string) error {
	// 创建配置文件内容
	configContent := fmt.Sprintf(`# TeamSpeak 3 服务器配置文件
# 生成时间: %s

# 服务器设置
server_name=%s
server_password=
server_welcome_message=%s
max_clients=%d
default_server=1

# 网络设置
default_voice_port=%d
filetransfer_port=%d
query_port=%d
server_port=%d

# 管理员设置
serveradmin_password=%s
serveradmin_token=%s

# 日志设置
log_queries=%t
log_client_cmds=%t
`,
		time.Now().Format(time.RFC3339),
		escapeIniValue(instance.Config.ServerName),
		escapeIniValue(instance.Config.WelcomeMsg),
		instance.Config.MaxClients,
		instance.Config.VoicePort,
		instance.Config.FilePort,
		instance.Config.QueryPort,
		instance.Config.ServerPort,
		escapeIniValue(instance.Config.QueryAdminPassword),
		escapeIniValue(instance.Config.ServerAdminToken),
		instance.Config.LogQueries,
		instance.Config.LogClientCmds,
	)

	// 写入配置文件
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	// 设置文件权限
	if err := os.Chmod(configPath, 0600); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置配置文件权限失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "config_path", Value: configPath})
		} else {
			log.Printf("警告: 设置配置文件权限失败: %v", err)
		}
	}

	// 设置文件所有者
	if err := os.Chown(configPath, 1000, 1000); err != nil {
		if s.logService != nil {
			s.logService.Warn("instance", "设置配置文件所有者失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "config_path", Value: configPath})
		} else {
			log.Printf("警告: 设置配置文件所有者失败: %v", err)
		}
	}

	return nil
}

// escapeIniValue 转义 INI 文件中的特殊字符
func escapeIniValue(value string) string {
	// 转义反斜杠、引号等特殊字符
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\r", "\\r",
		"\n", "\\n",
		"\t", "\\t",
		"\"", "\\\"",
	)
	return replacer.Replace(value)
}

// buildStartCommand 构建启动命令
func (s *Service) buildStartCommand(instance *Instance) (*exec.Cmd, error) {
	// 1. 准备命令参数
	args := []string{
		"run", "-d",
		"--name", fmt.Sprintf("teamspeak-%s", instance.ID),
		"--service-ports",
		"teamspeak",
	}

	// 2. 添加服务器参数
	args = append(args, "ts3server")
	args = append(args, fmt.Sprintf("default_voice_port=%d", instance.Config.VoicePort))
	args = append(args, fmt.Sprintf("filetransfer_port=%d", instance.Config.FilePort))
	args = append(args, fmt.Sprintf("query_port=%d", instance.Config.QueryPort))
	args = append(args, fmt.Sprintf("serveradmin_password=%s", instance.Config.QueryAdminPassword))

	// 3. 添加可选参数
	if instance.Config.ServerName != "" {
		args = append(args, fmt.Sprintf("server_name=%s", instance.Config.ServerName))
	}
	if instance.Config.WelcomeMsg != "" {
		args = append(args, fmt.Sprintf("welcome_message=%s", instance.Config.WelcomeMsg))
	}

	// 4. 创建命令 (使用docker compose而不是docker-compose)
	projectRoot, _ := os.Getwd()
	for range 3 {
		if _, err := os.Stat(filepath.Join(projectRoot, "docker-compose.yml")); err == nil {
			break
		}
		projectRoot = filepath.Dir(projectRoot)
	}

	// 使用docker compose而不是docker-compose
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)

	// 5. 设置工作目录为项目根目录以使用docker-compose.yml
	cmd.Dir = projectRoot

	return cmd, nil
}

// monitorInstance 监控实例运行状态
func (s *Service) monitorInstance(instance *Instance, cmd *exec.Cmd) {
	if s.logService != nil {
		s.logService.Info("instance", "开始监控实例", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "pid", Value: cmd.Process.Pid})
	} else {
		log.Printf("开始监控实例 %s (PID: %d)", instance.ID, cmd.Process.Pid)
	}

	// 1. 等待进程退出
	processState, err := cmd.Process.Wait()
	if err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "等待进程退出时出错", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "pid", Value: cmd.Process.Pid})
		} else {
			log.Printf("等待进程 %d 退出时出错: %v", cmd.Process.Pid, err)
		}
	}

	// 2. 更新实例状态并处理自动重启
	autoRestart := true
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 获取最新状态的实例
		var currentInstance Instance
		if txFirstErr := tx.First(&currentInstance, "id = ?", instance.ID).Error; txFirstErr != nil {
			return fmt.Errorf("获取实例状态失败: %w", txFirstErr)
		}

		// 如果实例已经被手动停止，则不更新状态
		if currentInstance.Status == StatusStopped {
			if s.logService != nil {
				s.logService.Info("instance", "实例已被手动停止，不更新状态", logs.LogField{Key: "instance_id", Value: instance.ID})
			} else {
				log.Printf("实例 %s 已被手动停止，不更新状态", instance.ID)
			}
			autoRestart = false
			return nil
		}

		// 更新实例状态
		exitCode := -1
		if processState != nil {
			exitCode = processState.ExitCode()
		}

		updateData := map[string]any{
			"status":      StatusStopped,
			"process_id":  0,
			"exit_code":   exitCode,
			"finished_at": time.Now(),
		}

		// 检查退出状态
		if exitCode != 0 {
			updateData["status"] = StatusError
			logMsg := fmt.Sprintf("TeamSpeak 服务器异常退出，退出码: %d", exitCode)
			currentInstance.AddLog(tx, "error", logMsg)
		} else {
			autoRestart = false // 正常退出时不自动重启
			logMsg := "TeamSpeak 服务器已正常退出"
			currentInstance.AddLog(tx, "info", logMsg)
		}

		// 更新实例状态
		if txModelErr := tx.Model(&currentInstance).Updates(updateData).Error; txModelErr != nil {
			return fmt.Errorf("更新实例状态失败: %w", txModelErr)
		}

		return nil
	})

	if err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "更新实例状态失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("更新实例 %s 状态失败: %v", instance.ID, err)
		}
	}

	// 3. 处理自动重启
	if autoRestart {
		// 获取自动重启配置
		config := DefaultAutoRestartConfig()

		// 检查是否可以重启
		if canRestart, delay := s.restartManager.CanRestart(instance.ID, config); canRestart {
			if s.logService != nil {
				s.logService.Info("instance", "准备自动重启实例", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "delay", Value: delay})
			} else {
				log.Printf("准备自动重启实例 %s (延迟: %v)", instance.ID, delay)
			}

			// 延迟重启
			time.Sleep(delay)

			// 获取最新状态的实例
			var currentInstance Instance
			if err := s.db.First(&currentInstance, "id = ?", instance.ID).Error; err != nil {
				if s.logService != nil {
					s.logService.Error("instance", "获取实例状态失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
				} else {
					log.Printf("获取实例 %s 状态失败: %v", instance.ID, err)
				}
				return
			}

			// 再次检查状态
			if currentInstance.Status == StatusStopped || currentInstance.Status == StatusError {
				// 更新状态为启动中
				if err := currentInstance.SetStatus(StatusStarting); err != nil {
					if s.logService != nil {
						s.logService.Error("instance", "更新实例状态为starting失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
					} else {
						log.Printf("更新实例 %s 状态为 starting 失败: %v", instance.ID, err)
					}
					return
				}

				if err := s.db.Save(&currentInstance).Error; err != nil {
					if s.logService != nil {
						s.logService.Error("instance", "保存实例状态失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
					} else {
						log.Printf("保存实例 %s 状态失败: %v", instance.ID, err)
					}
					return
				}

				// 记录重启日志
				_ = currentInstance.AddLog(s.db, "info", "正在自动重启实例...")

				// 启动实例
				if err := s.startInstance(&currentInstance); err != nil {
					if s.logService != nil {
						s.logService.Error("instance", "自动重启实例失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
					} else {
						log.Printf("自动重启实例 %s 失败: %v", instance.ID, err)
					}
					_ = currentInstance.AddLog(s.db, "error",
						fmt.Sprintf("自动重启失败: %v", err))
				} else {
					_ = currentInstance.AddLog(s.db, "info", "实例已自动重启")
				}
			}
		} else if delay > 0 {
			if s.logService != nil {
				s.logService.Warn("instance", "实例重启被限流", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "retry_delay", Value: delay})
			} else {
				log.Printf("实例 %s 的重启被限流，将在 %v 后重试", instance.ID, delay)
			}
		} else {
			if s.logService != nil {
				s.logService.Error("instance", "实例已达到最大重启次数，停止自动重启", logs.LogField{Key: "instance_id", Value: instance.ID})
			} else {
				log.Printf("实例 %s 已达到最大重启次数，停止自动重启", instance.ID)
			}
			_ = instance.AddLog(s.db, "error",
				"实例已达到最大自动重启次数，请检查问题后手动重启")
		}
	}

	// 4. 从健康监控中移除实例
	s.healthMonitor.RemoveInstance(instance.ID)

	if s.logService != nil {
		s.logService.Info("instance", "实例监控结束", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("实例 %s 监控结束", instance.ID)
	}
}

// buildStopCommand 构建停止命令
func (s *Service) buildStopCommand(instance *Instance) (*exec.Cmd, error) {
	// 使用 docker compose stop 命令停止容器
	containerName := fmt.Sprintf("teamspeak-%s", instance.ID)

	// 确定项目根目录
	projectRoot, _ := os.Getwd()
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(filepath.Join(projectRoot, "docker-compose.yml")); err == nil {
			break
		}
		projectRoot = filepath.Dir(projectRoot)
	}

	// 使用docker compose而不是docker-compose
	cmd := exec.Command("docker", "compose", "stop", "--timeout", "10", containerName)
	cmd.Dir = projectRoot
	return cmd, nil
}

// restartInstanceProcess 异步重启实例的实际实现
func (s *Service) restartInstanceProcess(instance *Instance) {
	if s.logService != nil {
		s.logService.Info("instance", "开始重启实例", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("开始重启实例 %s", instance.ID)
	}

	// 1. 停止实例
	if err := s.stopInstance(instance); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "停止实例失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("停止实例 %s 失败: %v", instance.ID, err)
		}
		return
	}

	// 2. 准备环境
	if err := s.prepareInstanceEnvironment(instance); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "准备实例环境失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("准备实例 %s 环境失败: %v", instance.ID, err)
		}
		return
	}

	// 3. 启动实例
	if err := s.startInstance(instance); err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "启动实例失败", logs.LogField{Key: "error", Value: err}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("启动实例 %s 失败: %v", instance.ID, err)
		}
		return
	}

	if s.logService != nil {
		s.logService.Info("instance", "实例重启完成", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("实例 %s 重启完成", instance.ID)
	}
}

// stopInstance 停止实例
func (s *Service) stopInstance(instance *Instance) error {
	if s.logService != nil {
		s.logService.Info("instance", "正在停止实例", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("正在停止实例 %s", instance.ID)
	}

	// 1. 构建停止命令
	cmd, err := s.buildStopCommand(instance)
	if err != nil {
		return fmt.Errorf("构建停止命令失败: %w", err)
	}

	// 2. 执行停止命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行停止命令失败: %v, 输出: %s", err, string(output))
	}

	// 3. 更新状态
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新实例状态
		if ModelStatusErr := tx.Model(instance).Update("status", StatusStopped).Error; ModelStatusErr != nil {
			return fmt.Errorf("更新实例状态失败: %w", ModelStatusErr)
		}

		// 清除进程ID
		if ModelProcessIDErr := tx.Model(instance).Update("process_id", 0).Error; ModelProcessIDErr != nil {
			return fmt.Errorf("清除进程ID失败: %w", ModelProcessIDErr)
		}

		// 添加日志
		instance.AddLog(tx, "info", "TeamSpeak 服务器已成功停止")
		return nil
	})

	if err != nil {
		log.Printf("更新实例 %s 状态失败: %v", instance.ID, err)
		return err
	}

	log.Printf("实例 %s 已成功停止", instance.ID)
	return nil
}

// startInstance 启动实例
func (s *Service) startInstance(instance *Instance) error {
	if s.logService != nil {
		s.logService.Info("instance", "正在启动实例", logs.LogField{Key: "instance_id", Value: instance.ID})
	} else {
		log.Printf("正在启动实例 %s", instance.ID)
	}

	// 1. 构建启动命令
	cmd, err := s.buildStartCommand(instance)
	if err != nil {
		return fmt.Errorf("构建启动命令失败: %w", err)
	}

	// 2. 启动进程
	if CmdStartErr := cmd.Start(); CmdStartErr != nil {
		return fmt.Errorf("启动进程失败: %w", CmdStartErr)
	}

	// 3. 更新状态
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 更新实例状态
		if ModelInstanceErr := tx.Model(instance).Updates(map[string]any{
			"status":     StatusRunning,
			"process_id": cmd.Process.Pid,
			"started_at": time.Now(),
		}).Error; ModelInstanceErr != nil {
			return fmt.Errorf("更新实例状态失败: %w", ModelInstanceErr)
		}

		// 添加日志
		instance.AddLog(tx, "info", "TeamSpeak 服务器已成功启动")
		return nil
	})

	if err != nil {
		if s.logService != nil {
			s.logService.Error("instance", "更新实例状态失败", logs.LogField{Key: "error", Value: err.Error()}, logs.LogField{Key: "instance_id", Value: instance.ID})
		} else {
			log.Printf("更新实例 %s 状态失败: %v", instance.ID, err)
		}
		return err
	}

	// 4. 启动监控协程
	go s.monitorInstance(instance, cmd)

	if s.logService != nil {
		s.logService.Info("instance", "实例已成功启动", logs.LogField{Key: "instance_id", Value: instance.ID}, logs.LogField{Key: "pid", Value: cmd.Process.Pid})
	} else {
		log.Printf("实例 %s 已成功启动 (PID: %d)", instance.ID, cmd.Process.Pid)
	}
	return nil
}

// GetInstanceLogs 获取实例日志
func (s *Service) GetInstanceLogs(ctx context.Context, id string, limit int) ([]InstanceLog, error) {
	// 检查实例是否存在
	if _, err := s.GetInstance(ctx, id); err != nil {
		return nil, err
	}

	var logs []InstanceLog
	err := s.db.Where("instance_id = ?", id).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get instance logs: %w", err)
	}

	return logs, nil
}

// GetInstanceResources 获取实例资源使用情况
func (s *Service) GetInstanceResources(ctx context.Context, id string) (*ResourceUsage, error) {
	// 获取实例
	instance, err := s.GetInstance(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查实例是否正在运行
	if !instance.IsRunning() {
		return nil, fmt.Errorf("instance is not running")
	}

	// 获取资源使用情况
	usage, err := getProcessResourceUsage(instance.ProcessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get process resource usage: %w", err)
	}

	return usage, nil
}

// Validate 验证输入参数
func (i *CreateInstanceInput) Validate() error {
	// 设置默认值
	if i.MaxClients == 0 {
		i.MaxClients = 32
	}
	if i.VoicePort == 0 {
		i.VoicePort = 9987
	}
	if i.FilePort == 0 {
		i.FilePort = 30033
	}
	if i.QueryPort == 0 {
		i.QueryPort = 10011
	}
	if i.ServerPort == 0 {
		i.ServerPort = 2010
	}

	// 验证端口是否冲突
	ports := map[int]string{
		i.VoicePort: "voice_port",
		i.FilePort:  "file_port",
		i.QueryPort: "query_port",
	}

	portMap := make(map[int]string)
	for port, name := range ports {
		if other, exists := portMap[port]; exists {
			return fmt.Errorf("port %d is used by both %s and %s", port, other, name)
		}
		portMap[port] = name
	}

	return nil
}

// Validate 验证更新参数
func (i *UpdateInstanceInput) Validate() error {
	// 至少需要一个更新字段
	if i.Name == "" && i.Version == "" && i.ServerName == "" && i.WelcomeMsg == "" && i.MaxClients == 0 {
		return errors.New("at least one field must be provided for update")
	}
	return nil
}

// Apply 应用过滤器
func (f *InstanceFilter) Apply(tx *gorm.DB) *gorm.DB {
	if f.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+f.Name+"%")
	}

	if f.Status != "" {
		tx = tx.Where("status = ?", f.Status)
	}

	// 应用分页
	if f.PageSize > 0 {
		offset := (f.Page - 1) * f.PageSize
		tx = tx.Offset(offset).Limit(f.PageSize)
	}

	return tx
}
