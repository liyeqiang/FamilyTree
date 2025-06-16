package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// Config 应用配置结构
type Config struct {
	// 应用配置
	Port        string `json:"port"`
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`

	// 数据库配置
	Database DatabaseConfig `json:"database"`

	// Redis配置
	Redis RedisConfig `json:"redis"`

	// 服务器配置
	Server ServerConfig `json:"server"`

	// 工作池配置
	WorkerPool WorkerPoolConfig `json:"worker_pool"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string `json:"type"`
	Path            string `json:"path"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"` // 秒
}

// RedisConfig Redis配置
type RedisConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	ReadTimeout  int  `json:"read_timeout"`  // 秒
	WriteTimeout int  `json:"write_timeout"` // 秒
	IdleTimeout  int  `json:"idle_timeout"`  // 秒
	EnableCORS   bool `json:"enable_cors"`
}

// WorkerPoolConfig 工作池配置
type WorkerPoolConfig struct {
	Enabled     bool `json:"enabled"`
	WorkerCount int  `json:"worker_count"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:        "8080",
		Environment: "development",
		LogLevel:    "info",
		Database: DatabaseConfig{
			Type:            "sqlite",
			Path:            "familytree.db",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 3600, // 1小时
		},
		Redis: RedisConfig{
			Enabled:  false,
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Server: ServerConfig{
			ReadTimeout:  15,
			WriteTimeout: 15,
			IdleTimeout:  60,
			EnableCORS:   true,
		},
		WorkerPool: WorkerPoolConfig{
			Enabled:     true,
			WorkerCount: 10,
		},
	}
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	config := DefaultConfig()

	// 从环境变量加载配置
	loadFromEnv(config)

	// 从配置文件加载配置
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("加载配置文件失败: %v", err)
		}
	}

	// 验证配置
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	return config, nil
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(config *Config) {
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	}
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if port, err := strconv.Atoi(redisPort); err == nil {
			config.Redis.Port = port
		}
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if workerCount := os.Getenv("WORKER_COUNT"); workerCount != "" {
		if count, err := strconv.Atoi(workerCount); err == nil {
			config.WorkerPool.WorkerCount = count
		}
	}
}

// loadFromFile 从配置文件加载配置
func loadFromFile(config *Config, configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, config)
}

// validate 验证配置
func validate(config *Config) error {
	if config.Port == "" {
		return fmt.Errorf("端口不能为空")
	}

	if config.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("数据库最大连接数必须大于0")
	}

	if config.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("数据库最大空闲连接数必须大于0")
	}

	if config.WorkerPool.WorkerCount <= 0 {
		return fmt.Errorf("工作池大小必须大于0")
	}

	return nil
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *Config) GetDatabaseDSN() string {
	switch c.Database.Type {
	case "sqlite":
		return c.Database.Path
	default:
		return c.Database.Path
	}
}
