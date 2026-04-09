package storage

import (
	"database/sql"
	"fmt"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// CategoryStorage Category 存储接口
type CategoryStorage interface {
	// CreateCategory 创建新 Category
	CreateCategory(category *models.Category) error

	// GetCategory 根据 ID 获取 Category
	GetCategory(id int) (*models.Category, error)

	// GetCategoryByUUID 根据 UUID 获取 Category
	GetCategoryByUUID(uuid string) (*models.Category, error)

	// GetCategoryByName 根据名称获取 Category
	GetCategoryByName(name string) (*models.Category, error)

	// ListCategories 列出所有 Category
	ListCategories() ([]*models.Category, error)

	// ListEnabledCategories 列出启用的 Category
	ListEnabledCategories() ([]*models.Category, error)

	// UpdateCategory 更新 Category
	UpdateCategory(category *models.Category) error

	// EnableCategory 启用 Category
	EnableCategory(uuid string) error

	// DisableCategory 禁用 Category
	DisableCategory(uuid string) error

	// DeleteCategory 删除 Category
	DeleteCategory(id int) error

	// DeleteAllCategories 删除所有 Category
	DeleteAllCategories() error
}

// MySQLCategoryStorage MySQL Category 存储实现
type MySQLCategoryStorage struct {
	db *sql.DB
}

// NewMySQLCategoryStorage 创建 MySQL Category 存储
func NewMySQLCategoryStorage(db *sql.DB) *MySQLCategoryStorage {
	return &MySQLCategoryStorage{db: db}
}

// CreateCategory 创建新 Category
func (s *MySQLCategoryStorage) CreateCategory(category *models.Category) error {
	query := `
		INSERT INTO categories (uuid, name, description, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(
		query,
		category.UUID, category.Name, category.Description, category.Enabled,
		category.CreatedAt, category.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	category.ID = int(id)
	return nil
}

// GetCategory 根据 ID 获取 Category
func (s *MySQLCategoryStorage) GetCategory(id int) (*models.Category, error) {
	query := `
		SELECT id, uuid, name, description, enabled, created_at, updated_at
		FROM categories
		WHERE id = ?
	`

	return s.scanCategory(s.db.QueryRow(query, id))
}

// GetCategoryByUUID 根据 UUID 获取 Category
func (s *MySQLCategoryStorage) GetCategoryByUUID(uuid string) (*models.Category, error) {
	query := `
		SELECT id, uuid, name, description, enabled, created_at, updated_at
		FROM categories
		WHERE uuid = ?
	`

	return s.scanCategory(s.db.QueryRow(query, uuid))
}

// GetCategoryByName 根据名称获取 Category
func (s *MySQLCategoryStorage) GetCategoryByName(name string) (*models.Category, error) {
	query := `
		SELECT id, uuid, name, description, enabled, created_at, updated_at
		FROM categories
		WHERE name = ?
	`

	return s.scanCategory(s.db.QueryRow(query, name))
}

// ListCategories 列出所有 Category
func (s *MySQLCategoryStorage) ListCategories() ([]*models.Category, error) {
	query := `
		SELECT id, uuid, name, description, enabled, created_at, updated_at
		FROM categories
		ORDER BY created_at ASC
	`

	return s.listCategoriesByQuery(query)
}

// ListEnabledCategories 列出启用的 Category
func (s *MySQLCategoryStorage) ListEnabledCategories() ([]*models.Category, error) {
	query := `
		SELECT id, uuid, name, description, enabled, created_at, updated_at
		FROM categories
		WHERE enabled = true
		ORDER BY created_at ASC
	`

	return s.listCategoriesByQuery(query)
}

// UpdateCategory 更新 Category
func (s *MySQLCategoryStorage) UpdateCategory(category *models.Category) error {
	query := `
		UPDATE categories SET
			name = ?, description = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(
		query,
		category.Name, category.Description, category.Enabled,
		category.UpdatedAt, category.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// EnableCategory 启用 Category
func (s *MySQLCategoryStorage) EnableCategory(uuid string) error {
	query := `UPDATE categories SET enabled = true, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to enable category: %w", err)
	}
	return nil
}

// DisableCategory 禁用 Category
func (s *MySQLCategoryStorage) DisableCategory(uuid string) error {
	query := `UPDATE categories SET enabled = false, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to disable category: %w", err)
	}
	return nil
}

// DeleteCategory 删除 Category
func (s *MySQLCategoryStorage) DeleteCategory(id int) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}

// DeleteAllCategories 删除所有 Category
func (s *MySQLCategoryStorage) DeleteAllCategories() error {
	query := `DELETE FROM categories`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all categories: %w", err)
	}
	return nil
}

// scanCategory 扫描单行数据到 Category 对象
func (s *MySQLCategoryStorage) scanCategory(row *sql.Row) (*models.Category, error) {
	category := &models.Category{}

	err := row.Scan(
		&category.ID, &category.UUID, &category.Name, &category.Description,
		&category.Enabled, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to scan category: %w", err)
	}

	return category, nil
}

// listCategoriesByQuery 根据查询列出 Category
func (s *MySQLCategoryStorage) listCategoriesByQuery(query string, args ...interface{}) ([]*models.Category, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}

		err := rows.Scan(
			&category.ID, &category.UUID, &category.Name, &category.Description,
			&category.Enabled, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
	}

	return categories, nil
}
