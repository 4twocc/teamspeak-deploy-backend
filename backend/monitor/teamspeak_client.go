package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	configPkg "teamspeak-one-click-deploy/config"

	"github.com/multiplay/go-ts3"
)

// promReconnectsTotal is defined in collector.go

type TeamSpeakClient struct {
	client *ts3.Client
	config configPkg.TeamspeakConfig
	ctx    context.Context
	cancel context.CancelFunc
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

	// 创建 TeamSpeak 客户端
	client, err := ts3.NewClient(addr)
	if err != nil {
		return fmt.Errorf("failed to create TeamSpeak client: %w", err)
	}

	// 设置超时
	// client.Cmd.SetTimeout(time.Duration(c.config.ReconnectInitialBackoff) * time.Second)

	// 登录
	if err := client.Login(c.config.Username, c.config.Password); err != nil {
		_ = client.Close()
		return fmt.Errorf("failed to login to TeamSpeak server: %w", err)
	}

	// 选择虚拟服务器
	if c.config.VirtualServerID > 0 {
		// if err := client.Cmd.Execute("use", fmt.Sprintf("sid=%d", c.config.VirtualServerID)); err != nil {
		// 	_ = client.Close()
		// 	return fmt.Errorf("failed to select virtual server: %w", err)
		// }
	} else if c.config.VirtualServerPort > 0 {
		if err := client.UsePort(c.config.VirtualServerPort); err != nil {
			_ = client.Close()
			return fmt.Errorf("failed to select virtual server by port: %w", err)
		}
	}

	// 设置昵称
	if c.config.Nickname != "" {
		// if err := client.Cmd.Execute("clientupdate", fmt.Sprintf("client_nickname=%s", c.config.Nickname)); err != nil {
		// 	log.Printf("Warning: failed to set nickname: %v", err)
		// }
	}

	c.client = client
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

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := c.connect(); err != nil {
			lastErr = err
			log.Printf("Reconnect attempt %d failed: %v", attempt+1, err)

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
		log.Println("Successfully reconnected to TeamSpeak server")
		// promReconnectsTotal.Inc() // Increment reconnect counter
		return nil
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %w", maxRetries, lastErr)
}

// GetServerInfo retrieves information about the TeamSpeak server
func (c *TeamSpeakClient) GetServerInfo() (*ServerInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected to TeamSpeak server")
	}

	// Get server info
	si, err := c.client.Server.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	// Get server instance info
	// sii, err := c.client.Server.InstanceInfo()
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get server instance info: %w", err)
	// }

	info := &ServerInfo{
		OnlineUsers:  si.ClientsOnline,
		ChannelCount: 0, // Placeholder - need to find proper way to get channel count
		Uptime:       time.Duration(0) * time.Second, // Placeholder for uptime
		// VoiceQuality is a synthetic metric based on packet loss, ping, etc.
		// For now, we'll use a placeholder value
		VoiceQuality: 80.0,
	}

	// Calculate voice quality based on server info
	// packetLoss := si.PacketLossTotal
	// ping := si.Ping

	// Simple voice quality calculation:
	// Start with 100%, subtract penalty for packet loss and ping
	quality := 100.0
	// if packetLoss > 0 {
	// 	quality -= packetLoss * 100 * 2 // 2x penalty for packet loss
	// }
	// if ping > 50 {
	// 	quality -= float64(ping-50) * 0.5 // 0.5% penalty for each ms over 50
	// }

	// Ensure quality is between 0 and 100
	if quality < 0 {
		quality = 0
	} else if quality > 100 {
		quality = 100
	}

	info.VoiceQuality = quality

	return info, nil
}

// ServerInfo contains information about the TeamSpeak server
type ServerInfo struct {
	OnlineUsers  int           `json:"online_users"`
	ChannelCount int           `json:"channel_count"`
	Uptime       time.Duration `json:"uptime"`
	VoiceQuality float64       `json:"voice_quality"` // 0-100 scale
}