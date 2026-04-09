package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/go-sql-driver/mysql"
)

// Storage 存储接口聚合
type Storage interface {
	Close() error
	DB() *sql.DB
}

// MySQLStorage MySQL 存储实现
type MySQLStorage struct {
	db *sql.DB
}

// NewMySQLStorage 创建 MySQL 存储，带重试逻辑
func NewMySQLStorage(dsn string) (*MySQLStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	// 连接最大存活时间（小于 MySQL wait_timeout，默认 8 小时）
	// 设置为 4 小时，让连接在服务器超时前主动回收
	db.SetConnMaxLifetime(4 * time.Hour)
	// 空闲连接最大存活时间，10 分钟后回收空闲连接
	db.SetConnMaxIdleTime(10 * time.Minute)

	// 带重试的连接测试，最多重试 30 次，每次间隔 2 秒
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(); err != nil {
			if i < maxRetries-1 {
				log.Printf("[Storage] Database connection failed (attempt %d/%d): %v, retrying in %v...", i+1, maxRetries, err, retryInterval)
				time.Sleep(retryInterval)
				continue
			}
			return nil, fmt.Errorf("failed to ping database after %d attempts: %w", maxRetries, err)
		}
		break
	}

	log.Printf("[Storage] Database connected successfully")
	return &MySQLStorage{db: db}, nil
}

// Close 关闭数据库连接
func (s *MySQLStorage) Close() error {
	return s.db.Close()
}

// DB 返回底层数据库连接
func (s *MySQLStorage) DB() *sql.DB {
	return s.db
}

// Transaction 在事务中执行操作
func (s *MySQLStorage) Transaction(fn func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// WithContext 创建带超时的上下文
func (s *MySQLStorage) WithContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultTimeout)
}

// DefaultTimeout 默认操作超时时间
const DefaultTimeout = 30 * time.Second
