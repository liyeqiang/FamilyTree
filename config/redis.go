package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// DefaultRedisConfig 返回默认的Redis配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
		PoolSize: 10,
	}
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return client, nil
}

// GetRedisURL 获取Redis连接URL
func (c *RedisConfig) GetRedisURL() string {
	if c.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d", c.Password, c.Host, c.Port, c.DB)
	}
	return fmt.Sprintf("redis://%s:%d/%d", c.Host, c.Port, c.DB)
}

// Validate 验证配置
func (c *RedisConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("Redis host cannot be empty")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid Redis port: %d", c.Port)
	}
	if c.PoolSize <= 0 {
		return fmt.Errorf("invalid pool size: %d", c.PoolSize)
	}
	return nil
}
