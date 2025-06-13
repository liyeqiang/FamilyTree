package database

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生存时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
}

// DefaultPoolConfig 默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 30,
	}
}

// Pool 数据库连接池
type Pool struct {
	db     *sql.DB
	config *PoolConfig
	mu     sync.RWMutex
	stats  *PoolStats
}

// PoolStats 连接池统计信息
type PoolStats struct {
	OpenConnections   int
	InUse             int
	Idle              int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxLifetimeClosed int64
}

// NewPool 创建新的数据库连接池
func NewPool(driverName, dataSourceName string, config *PoolConfig) (*Pool, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	pool := &Pool{
		db:     db,
		config: config,
		stats:  &PoolStats{},
	}

	return pool, nil
}

// DB 获取数据库连接
func (p *Pool) DB() *sql.DB {
	return p.db
}

// Stats 获取连接池统计信息
func (p *Pool) Stats() *PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	dbStats := p.db.Stats()

	return &PoolStats{
		OpenConnections:   dbStats.OpenConnections,
		InUse:             dbStats.InUse,
		Idle:              dbStats.Idle,
		WaitCount:         dbStats.WaitCount,
		WaitDuration:      dbStats.WaitDuration,
		MaxIdleClosed:     dbStats.MaxIdleClosed,
		MaxLifetimeClosed: dbStats.MaxLifetimeClosed,
	}
}

// Health 检查连接池健康状态
func (p *Pool) Health() error {
	return p.db.Ping()
}

// Close 关闭连接池
func (p *Pool) Close() error {
	return p.db.Close()
}

// Transaction 执行事务
func (p *Pool) Transaction(fn func(*sql.Tx) error) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// PreparedStatement 预处理语句管理器
type PreparedStatement struct {
	pool       *Pool
	statements map[string]*sql.Stmt
	mu         sync.RWMutex
}

// NewPreparedStatement 创建预处理语句管理器
func NewPreparedStatement(pool *Pool) *PreparedStatement {
	return &PreparedStatement{
		pool:       pool,
		statements: make(map[string]*sql.Stmt),
	}
}

// Prepare 准备语句
func (ps *PreparedStatement) Prepare(name, query string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	stmt, err := ps.pool.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement %s: %v", name, err)
	}

	// 如果已存在，先关闭旧的
	if oldStmt, exists := ps.statements[name]; exists {
		oldStmt.Close()
	}

	ps.statements[name] = stmt
	return nil
}

// Get 获取预处理语句
func (ps *PreparedStatement) Get(name string) (*sql.Stmt, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	stmt, exists := ps.statements[name]
	if !exists {
		return nil, fmt.Errorf("prepared statement %s not found", name)
	}

	return stmt, nil
}

// Close 关闭所有预处理语句
func (ps *PreparedStatement) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var lastErr error
	for name, stmt := range ps.statements {
		if err := stmt.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close statement %s: %v", name, err)
		}
	}

	ps.statements = make(map[string]*sql.Stmt)
	return lastErr
}

// BatchExecutor 批量执行器
type BatchExecutor struct {
	pool *Pool
	size int
}

// NewBatchExecutor 创建批量执行器
func NewBatchExecutor(pool *Pool, batchSize int) *BatchExecutor {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &BatchExecutor{
		pool: pool,
		size: batchSize,
	}
}

// Execute 批量执行SQL语句
func (be *BatchExecutor) Execute(queries []string) error {
	return be.pool.Transaction(func(tx *sql.Tx) error {
		for i, query := range queries {
			if _, err := tx.Exec(query); err != nil {
				return fmt.Errorf("failed to execute query %d: %v", i, err)
			}
		}
		return nil
	})
}

// ExecuteWithArgs 批量执行带参数的SQL语句
func (be *BatchExecutor) ExecuteWithArgs(query string, args [][]interface{}) error {
	return be.pool.Transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %v", err)
		}
		defer stmt.Close()

		for i, argSet := range args {
			if _, err := stmt.Exec(argSet...); err != nil {
				return fmt.Errorf("failed to execute statement with args %d: %v", i, err)
			}
		}
		return nil
	})
}
