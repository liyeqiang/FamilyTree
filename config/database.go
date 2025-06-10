package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	DBPath string
}

// LoadConfig 加载数据库配置
func LoadConfig() *DatabaseConfig {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// 默认数据库路径
		dbPath = "familytree.db"
	}

	return &DatabaseConfig{
		DBPath: dbPath,
	}
}

// Connect 连接到SQLite数据库
func (c *DatabaseConfig) Connect() (*sql.DB, error) {
	// 确保数据库目录存在
	dbDir := filepath.Dir(c.DBPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("创建数据库目录失败: %v", err)
		}
	}

	db, err := sql.Open("sqlite", c.DBPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	return db, nil
} 