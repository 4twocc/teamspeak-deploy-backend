package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	configPkg "teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/logs"

	"github.com/multiplay/go-ts3"
)

type TeamSpeakClient struct {
	client *ts3.Client
	config configPkg.TeamspeakConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// ServerInfo contains information about the TeamSpeak server
type ServerInfo struct {
	OnlineUsers  int           `json:"online_users"`
	ChannelCount int           `json:"channel_count"`
	Uptime       time.Duration `json:"uptime"`
	VoiceQuality float64       `json:"voice_quality"` // 0-100 scale
}

// NewTeamSpeakClient creates a new TeamSpeak client with the given configuration
func NewTeamSpeakClient(config configPkg.TeamspeakConfig) (*TeamSpeakClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &TeamSpeakClient{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to TeamSpeak server: %w", err)
	}

	return client, nil
}

// connect establishes a new connection to the TeamSpeak server
func (c *TeamSpeakClient) connect() error {
	// 构造连接地址
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.QueryPort)
	if logService != nil {
		logService.Info("monitor", "尝试连接TeamSpeak服务器", logs.LogField{Key: "address", Value: addr})
	} else {
		log.Printf("Attempting to connect to TeamSpeak server at %s", addr)
	}

	// 创建 TeamSpeak 客户端
	client, err := ts3.NewClient(addr)
	if err != nil {
		return fmt.Errorf("failed to create TeamSpeak client: %w", err)
	}

	// 登录
	if err := client.Login(c.config.Username, c.config.Password); err != nil {
		_ = client.Close()
		return fmt.Errorf("failed to login to TeamSpeak server: %w", err)
	}

	// 选择虚拟服务器
	if c.config.VirtualServerID > 0 {
		// 使用 ExecCmd 方法执行 use 命令
		cmd := ts3.NewCmd(fmt.Sprintf("use sid=%d", c.config.VirtualServerID))
		if _, err := client.ExecCmd(cmd); err != nil {
			_ = client.Close()
			return fmt.Errorf("failed to select virtual server: %w", err)
		}
	} else if c.config.VirtualServerPort > 0 {
		if err := client.UsePort(c.config.VirtualServerPort); err != nil {
			_ = client.Close()
			return fmt.Errorf("failed to select virtual server by port: %w", err)
		}
	}

	// 设置昵称
	if c.config.Nickname != "" {
		cmd := ts3.NewCmd(fmt.Sprintf("clientupdate client_nickname=%s", c.config.Nickname))
		if _, err := client.ExecCmd(cmd); err != nil {
			if logService != nil {
				logService.Warn("monitor", "设置昵称失败", logs.LogField{Key: "error", Value: err.Error()})
			} else {
				log.Printf("Warning: failed to set nickname: %v", err)
			}
		}
	}

	c.client = client
	if logService != nil {
		logService.Info("monitor", "成功连接到TeamSpeak服务器")
	} else {
		log.Println("Successfully connected to TeamSpeak server")
	}
	return nil
}

// Close closes the TeamSpeak client connection
func (c *TeamSpeakClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ensureConnected ensures the client is connected, reconnecting if necessary
func (c *TeamSpeakClient) ensureConnected() error {
	// Check if we're already connected
	if c.client != nil {
		if _, err := c.client.Whoami(); err == nil {
			return nil // Already connected
		}
		// Connection lost, clean up
		_ = c.client.Close()
		c.client = nil
	}

	// Reconnect with exponential backoff
	maxRetries := c.config.ReconnectMaxRetries
	initialBackoff := c.config.ReconnectInitialBackoff
	maxBackoff := c.config.ReconnectMaxBackoff

	var lastErr error
	backoff := initialBackoff

	for attempt := range maxRetries {
		if err := c.connect(); err != nil {
			lastErr = err
			if logService != nil {
				logService.Warn("monitor", "重连尝试失败", logs.LogField{Key: "attempt", Value: attempt + 1}, logs.LogField{Key: "error", Value: err.Error()})
			} else {
				log.Printf("Reconnect attempt %d failed: %v", attempt+1, err)
			}

			if attempt < maxRetries-1 { // Don't sleep after the last attempt
				time.Sleep(backoff)
				// Exponential backoff, capped at maxBackoff
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}

		// Success
		if logService != nil {
			logService.Info("monitor", "成功重连到TeamSpeak服务器")
		} else {
			log.Println("Successfully reconnected to TeamSpeak server")
		}
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %w", maxRetries, lastErr)
}

// GetServerInfo retrieves information about the TeamSpeak server
func (c *TeamSpeakClient) GetServerInfo() (*ServerInfo, error) {
	// 确保连接可用
	if err := c.ensureConnected(); err != nil {
		return nil, fmt.Errorf("not connected to TeamSpeak server: %w", err)
	}

	// Get server info
	si, err := c.client.Server.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	info := &ServerInfo{
		OnlineUsers:  si.ClientsOnline,
		ChannelCount: si.ChannelsOnline,
		Uptime:       time.Duration(si.Uptime) * time.Second,
		VoiceQuality: 80.0, // Placeholder for now
	}

	return info, nil
}

// UseID is a helper function to select a virtual server by ID
func UseID(client *ts3.Client, serverID int) error {
	cmd := ts3.NewCmd(fmt.Sprintf("use sid=%d", serverID))
	_, err := client.ExecCmd(cmd)
	return err
}
