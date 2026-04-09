package api

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/storage"
	_ "github.com/go-sql-driver/mysql"
)

// MySQLUserStorage MySQL 用户存储实现（复用 event-processor 的表结构）
type MySQLUserStorage struct {
	db *sql.DB
}

// NewMySQLUserStorage 创建 MySQL 用户存储
func NewMySQLUserStorage(dsn string) (*MySQLUserStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数以避免空闲连接被服务器关闭
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(4 * time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &MySQLUserStorage{db: db}, nil
}

// Close 关闭数据库连接
func (s *MySQLUserStorage) Close() error {
	return s.db.Close()
}

// GetSessionWithUser 根据 session_id 获取 session 和用户信息
func (s *MySQLUserStorage) GetSessionWithUser(sessionID string) (map[string]interface{}, error) {
	query := `
		SELECT s.id, s.user_id, s.expires_at, s.created_at,
			u.id, u.username, u.role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ? AND s.expires_at > NOW()
	`

	var sessionID2 string
	var userID int
	var expiresAt, createdAt time.Time
	var uid int
	var username, role string

	err := s.db.QueryRow(query, sessionID).Scan(
		&sessionID2, &userID, &expiresAt, &createdAt,
		&uid, &username, &role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session with user: %w", err)
	}

	return map[string]interface{}{
		"id":         sessionID2,
		"user_id":    userID,
		"expires_at": expiresAt,
		"created_at": createdAt,
		"user": map[string]interface{}{
			"id":       uid,
			"username": username,
			"role":     role,
		},
	}, nil
}

// GetUserByUsername 根据用户名获取用户（实现 storage.UserStorage 接口）
func (s *MySQLUserStorage) GetUserByUsername(username string) (*storage.User, error) {
	query := `
		SELECT id, username, password, role
		FROM users
		WHERE username = ?
	`

	user := &storage.User{}

	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser 创建新用户（实现 storage.UserStorage 接口）
func (s *MySQLUserStorage) CreateUser(username, hashedPassword, role string) error {
	query := `
		INSERT INTO users (username, password, role, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	result, err := s.db.Exec(
		query,
		username, hashedPassword, role,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	if id == 0 {
		return fmt.Errorf("failed to create user: no rows affected")
	}

	return nil
}
