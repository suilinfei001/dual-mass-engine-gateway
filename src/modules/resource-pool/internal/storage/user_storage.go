package storage

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptHashPassword 使用bcrypt哈希密码
func BcryptHashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// BcryptCompareHashAndPassword 比较密码和哈希值
func BcryptCompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// UserStorage 用户存储接口（用于访问event-processor的用户数据库）
type UserStorage interface {
	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(username string) (*User, error)
	// CreateUser 创建新用户
	CreateUser(username, hashedPassword, role string) error
}

// User 用户模型（简化版，兼容event-processor的User模型）
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"` // bcrypt哈希后的密码
	Role     string `json:"role"`
	Email    string `json:"email,omitempty"`
}

// GetID 获取用户ID
func (u *User) GetID() int {
	return u.ID
}

// GetUsername 获取用户名
func (u *User) GetUsername() string {
	return u.Username
}

// GetPassword 获取密码
func (u *User) GetPassword() string {
	return u.Password
}

// GetRole 获取角色
func (u *User) GetRole() string {
	return u.Role
}

// MySQLUserStorage MySQL 用户存储实现
type MySQLUserStorage struct {
	db *sql.DB
}

// NewMySQLUserStorage 创建 MySQL 用户存储
func NewMySQLUserStorage(db *sql.DB) *MySQLUserStorage {
	return &MySQLUserStorage{db: db}
}

// GetUserByUsername 根据用户名获取用户
func (s *MySQLUserStorage) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, password, role
		FROM users
		WHERE username = ?
	`

	user := &User{}

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

// CreateUser 创建新用户
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

// ValidatePassword 验证密码
func (s *MySQLUserStorage) ValidatePassword(username, password string) (*User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, nil
	}

	return user, nil
}

// ListUsers 列出所有用户
func (s *MySQLUserStorage) ListUsers() ([]*User, error) {
	query := `
		SELECT id, username, password, role
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}
