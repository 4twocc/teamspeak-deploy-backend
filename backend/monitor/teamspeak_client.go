package monitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/multiplay/go-ts3"
)

// promReconnectsTotal is defined in collector.go

type TeamSpeakClient struct {
	client *ts3.Client
	config TeamSpeakConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTeamSpeakClient creates a new TeamSpeak client with the given configuration
func NewTeamSpeakClient(config TeamSpeakConfig) (*TeamSpeakClient, error) {
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
	if c.client != nil {
		c.client.Close()
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	var client *ts3.Client
	var err error

	// Try to connect with a timeout
	_, cancel := context.WithTimeout(c.ctx, time.Duration(c.config.Timeout)*time.Second)
	defer cancel()

	// Connect to the server
	client, err = ts3.NewClient(addr,
		ts3.Timeout(time.Duration(c.config.Timeout)*time.Second),
	)

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	// Authenticate after connection
	if err := client.Login(c.config.Username, c.config.Password); err != nil {
		client.Close()
		return fmt.Errorf("login failed: %w", err)
	}

	// Prefer selecting virtual server by ID if configured; otherwise by port
	if c.config.VirtualServerID > 0 {
		if _, err := client.Exec(fmt.Sprintf("use sid=%d", c.config.VirtualServerID)); err != nil {
			client.Close()
			return fmt.Errorf("select virtual server by id failed: %w", err)
		}
		log.Printf("Selected TeamSpeak virtual server by id %d", c.config.VirtualServerID)
	} else if c.config.VirtualServerPort > 0 {
		if _, err := client.Exec(fmt.Sprintf("use port=%d", c.config.VirtualServerPort)); err != nil {
			client.Close()
			return fmt.Errorf("select virtual server by port failed: %w", err)
		}
		log.Printf("Selected TeamSpeak virtual server on port %d", c.config.VirtualServerPort)
	}

	// Optionally set ServerQuery nickname
	if c.config.Nickname != "" {
		if err := client.ClientUpdate(ts3.NewArg("client_nickname", c.config.Nickname)); err != nil {
			client.Close()
			return fmt.Errorf("set ServerQuery nickname failed: %w", err)
		}
		log.Printf("Set ServerQuery nickname to '%s'", c.config.Nickname)
	}

	c.client = client
	return nil
}

// reconnect attempts to reestablish the connection with exponential backoff
func (c *TeamSpeakClient) reconnect() error {
	// 从配置读取重连参数，提供合理默认值
	retries := c.config.ReconnectMaxRetries
	if retries <= 0 {
		retries = 5
	}
	initial := time.Duration(c.config.ReconnectInitialBackoff) * time.Second
	if initial <= 0 {
		initial = 1 * time.Second
	}
	maxBackoff := time.Duration(c.config.ReconnectMaxBackoff) * time.Second
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}

	var lastErr error
	backoff := initial

	for i := range retries {
		if i > 0 {
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			log.Printf("TeamSpeak reconnection attempt %d/%d in %v...", i+1, retries, backoff)
			select {
			case <-time.After(backoff):
				// Continue with retry
			case <-c.ctx.Done():
				return context.Canceled
			}
			// 退避翻倍
			if backoff < maxBackoff {
				backoff *= 2
			}
		}

		if err := c.connect(); err == nil {
			log.Println("Successfully reconnected to TeamSpeak server")
			promReconnectsTotal.Inc()
			return nil
		} else {
			lastErr = err
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %w", retries, lastErr)
}

// ensureConnected checks if the client is connected and reconnects if necessary
func (c *TeamSpeakClient) ensureConnected() error {
	if c.client == nil {
		return c.reconnect()
	}

	// Simple ping to check if connection is alive
	_, err := c.client.Version()
	if err != nil {
		log.Printf("TeamSpeak connection lost, attempting to reconnect...")
		return c.reconnect()
	}

	return nil
}

// GetServerInfo retrieves server information including online users and channels
func (c *TeamSpeakClient) GetServerInfo() (*ServerInfo, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}

	server, err := c.client.Server.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	clients, err := c.client.Server.ClientList()
	if err != nil {
		return nil, fmt.Errorf("failed to get client list: %w", err)
	}

	channels, err := c.client.Server.ChannelList()
	if err != nil {
		return nil, fmt.Errorf("failed to get channel list: %w", err)
	}

	// Filter out query clients
	var onlineUsers int
	for _, client := range clients {
		if client.Type == 0 { // Regular client, not a query
			onlineUsers++
		}
	}

	return &ServerInfo{
		Uptime:       time.Duration(server.Uptime) * time.Second,
		OnlineUsers:  onlineUsers,
		ChannelCount: len(channels),
		VoiceQuality: calculateVoiceQuality(server.TotalPacketLossTotal),
		LastUpdated:  time.Now(),
	}, nil
}

// Close closes the TeamSpeak client connection
func (c *TeamSpeakClient) Close() error {
	c.cancel()
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// calculateVoiceQuality calculates voice quality based on packet loss
func calculateVoiceQuality(packetLoss float64) float64 {
	// Convert packet loss percentage to quality score (0-100)
	// Lower packet loss = higher quality
	quality := 100.0 - (packetLoss * 2) // Simple linear scaling

	// Ensure quality is within bounds
	if quality < 0 {
		return 0
	}
	if quality > 100 {
		return 100
	}
	return quality
}

// ServerInfo contains information about the TeamSpeak server
type ServerInfo struct {
	Uptime       time.Duration `json:"uptime"`
	OnlineUsers  int           `json:"online_users"`
	ChannelCount int           `json:"channel_count"`
	VoiceQuality float64       `json:"voice_quality"`
	LastUpdated  time.Time     `json:"last_updated"`
}
