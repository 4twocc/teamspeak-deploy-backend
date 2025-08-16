package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
)

// InitRedisCache 初始化Redis缓存
func InitRedisCache() error {
	config := GetConfig()
	
	// 如果未启用Redis，则不初始化
	if !config.RedisConfig.Enabled {
		log.Println("Redis cache is disabled")
		return nil
	}
	
	// 创建Redis客户端
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisConfig.Addr,
		Password: config.RedisConfig.Password,
		DB:       config.RedisConfig.DB,
	})
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		redisClient = nil
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	
	log.Println("Redis cache initialized successfully")
	return nil
}

// CloseRedisCache 关闭Redis缓存连接
func CloseRedisCache() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
		redisClient = nil
	}
}

// getCachedSystemMetrics 从Redis缓存获取系统指标
func getCachedSystemMetrics() (*SystemMetrics, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	val, err := redisClient.Get(ctx, "system_metrics").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get system metrics from cache: %w", err)
	}
	
	var metrics SystemMetrics
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal system metrics: %w", err)
	}
	
	return &metrics, nil
}

// cacheSystemMetrics 将系统指标缓存到Redis
func cacheSystemMetrics(metrics *SystemMetrics) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal system metrics: %w", err)
	}
	
	// 缓存1小时
	err = redisClient.Set(ctx, "system_metrics", data, 1*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache system metrics: %w", err)
	}
	
	return nil
}

// getCachedBusinessMetrics 从Redis缓存获取业务指标
func getCachedBusinessMetrics() (*BusinessMetrics, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	val, err := redisClient.Get(ctx, "business_metrics").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get business metrics from cache: %w", err)
	}
	
	var metrics BusinessMetrics
	if err := json.Unmarshal([]byte(val), &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal business metrics: %w", err)
	}
	
	return &metrics, nil
}

// cacheBusinessMetrics 将业务指标缓存到Redis
func cacheBusinessMetrics(metrics *BusinessMetrics) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal business metrics: %w", err)
	}
	
	// 缓存1小时
	err = redisClient.Set(ctx, "business_metrics", data, 1*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to cache business metrics: %w", err)
	}
	
	return nil
}

// isInCooldown 检查是否在冷却期内
func isInCooldown(key string, cooldown time.Duration) (bool, error) {
	if redisClient == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cooldown key: %w", err)
	}
	
	return exists > 0, nil
}

// setCooldown 设置冷却期
func setCooldown(key string, duration time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := redisClient.Set(ctx, key, "1", duration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cooldown: %w", err)
	}
	
	return nil
}