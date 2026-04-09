// Package storage provides database access for resource manager service.
package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/shared/pkg/storage"
)

// Config extends shared config with additional fields
type Config struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// DSN generates data source name
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Open creates a new database connection
func Open(cfg Config, log *logger.Logger) (*storage.DB, error) {
	sharedCfg := storage.Config{
		Driver:          cfg.Driver,
		DSN:             cfg.DSN(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}
	return storage.Open(sharedCfg, log)
}

// ResourceStorage 资源实例存储
type ResourceStorage struct {
	db     *storage.DB
	logger *logger.Logger
}

// NewResourceStorage 创建资源存储
func NewResourceStorage(db *storage.DB) *ResourceStorage {
	return &ResourceStorage{db: db}
}

// List 列出资源
func (s *ResourceStorage) List(ctx context.Context) ([]*ResourceInstance, error) {
	query := `
		SELECT id, uuid, name, description, ip_address, ssh_port, ssh_user, ssh_password,
		       category_id, is_public, created_by, status, created_at, updated_at
		FROM resource_instances
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*ResourceInstance
	for rows.Next() {
		var r ResourceInstance
		err := rows.Scan(
			&r.ID, &r.UUID, &r.Name, &r.Description, &r.IPAddress, &r.SSHPort,
			&r.SSHUser, &r.SSHPassword, &r.CategoryID, &r.IsPublic, &r.CreatedBy,
			&r.Status, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &r)
	}
	return resources, nil
}

// GetByUUID 根据 UUID 获取资源
func (s *ResourceStorage) GetByUUID(ctx context.Context, uuid string) (*ResourceInstance, error) {
	query := `
		SELECT id, uuid, name, description, ip_address, ssh_port, ssh_user, ssh_password,
		       category_id, is_public, created_by, status, created_at, updated_at
		FROM resource_instances
		WHERE uuid = ?
	`
	var r ResourceInstance
	err := s.db.QueryRow(ctx, query, uuid).Scan(
		&r.ID, &r.UUID, &r.Name, &r.Description, &r.IPAddress, &r.SSHPort,
		&r.SSHUser, &r.SSHPassword, &r.CategoryID, &r.IsPublic, &r.CreatedBy,
		&r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// ListByCategory 根据类别列出资源
func (s *ResourceStorage) ListByCategory(ctx context.Context, categoryID int64) ([]*ResourceInstance, error) {
	query := `
		SELECT id, uuid, name, description, ip_address, ssh_port, ssh_user, ssh_password,
		       category_id, is_public, created_by, status, created_at, updated_at
		FROM resource_instances
		WHERE category_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*ResourceInstance
	for rows.Next() {
		var r ResourceInstance
		err := rows.Scan(
			&r.ID, &r.UUID, &r.Name, &r.Description, &r.IPAddress, &r.SSHPort,
			&r.SSHUser, &r.SSHPassword, &r.CategoryID, &r.IsPublic, &r.CreatedBy,
			&r.Status, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &r)
	}
	return resources, nil
}

// Create 创建资源
func (s *ResourceStorage) Create(ctx context.Context, r *ResourceInstance) error {
	query := `
		INSERT INTO resource_instances
		(uuid, name, description, ip_address, ssh_port, ssh_user, ssh_password,
		 category_id, is_public, created_by, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(ctx, query,
		r.UUID, r.Name, r.Description, r.IPAddress, r.SSHPort, r.SSHUser, r.SSHPassword,
		r.CategoryID, r.IsPublic, r.CreatedBy, r.Status,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

// Update 更新资源
func (s *ResourceStorage) Update(ctx context.Context, r *ResourceInstance) error {
	query := `
		UPDATE resource_instances
		SET name = ?, description = ?, ip_address = ?, ssh_port = ?, ssh_user = ?,
		    ssh_password = ?, category_id = ?, is_public = ?, status = ?
		WHERE uuid = ?
	`
	_, err := s.db.Exec(ctx, query,
		r.Name, r.Description, r.IPAddress, r.SSHPort, r.SSHUser, r.SSHPassword,
		r.CategoryID, r.IsPublic, r.Status, r.UUID,
	)
	return err
}

// Delete 删除资源
func (s *ResourceStorage) Delete(ctx context.Context, uuid string) error {
	query := `DELETE FROM resource_instances WHERE uuid = ?`
	_, err := s.db.Exec(ctx, query, uuid)
	return err
}

// ListAvailable 列出可用资源（状态为 active 且未被分配）
func (s *ResourceStorage) ListAvailable(ctx context.Context, categoryID int64) ([]*ResourceInstance, error) {
	query := `
		SELECT r.id, r.uuid, r.name, r.description, r.ip_address, r.ssh_port, r.ssh_user, r.ssh_password,
		       r.category_id, r.is_public, r.created_by, r.status, r.created_at, r.updated_at
		FROM resource_instances r
		WHERE r.category_id = ? AND r.status = ?
		  AND NOT EXISTS (
		    SELECT 1 FROM allocations a WHERE a.resource_uuid = r.uuid AND a.status = 'active'
		  )
		ORDER BY r.created_at ASC
	`
	rows, err := s.db.Query(ctx, query, categoryID, ResourceStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*ResourceInstance
	for rows.Next() {
		var r ResourceInstance
		err := rows.Scan(
			&r.ID, &r.UUID, &r.Name, &r.Description, &r.IPAddress, &r.SSHPort,
			&r.SSHUser, &r.SSHPassword, &r.CategoryID, &r.IsPublic, &r.CreatedBy,
			&r.Status, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &r)
	}
	return resources, nil
}

// CategoryStorage 类别存储
type CategoryStorage struct {
	db *storage.DB
}

// NewCategoryStorage 创建类别存储
func NewCategoryStorage(db *storage.DB) *CategoryStorage {
	return &CategoryStorage{db: db}
}

// List 列出所有类别
func (s *CategoryStorage) List(ctx context.Context) ([]*Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		var c Category
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

// GetByName 根据名称获取类别
func (s *CategoryStorage) GetByName(ctx context.Context, name string) (*Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE name = ?`
	var c Category
	err := s.db.QueryRow(ctx, query, name).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

// GetByID 根据 ID 获取类别
func (s *CategoryStorage) GetByID(ctx context.Context, id int64) (*Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = ?`
	var c Category
	err := s.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

// Create 创建类别
func (s *CategoryStorage) Create(ctx context.Context, c *Category) error {
	query := `INSERT INTO categories (uuid, name, description) VALUES (?, ?, ?)`
	result, err := s.db.Exec(ctx, query, c.UUID, c.Name, c.Description)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

// Update 更新类别
func (s *CategoryStorage) Update(ctx context.Context, c *Category) error {
	query := `UPDATE categories SET name = ?, description = ? WHERE id = ?`
	_, err := s.db.Exec(ctx, query, c.Name, c.Description, c.ID)
	return err
}

// Delete 删除类别
func (s *CategoryStorage) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

// QuotaPolicyStorage 配额策略存储
type QuotaPolicyStorage struct {
	db *storage.DB
}

// NewQuotaPolicyStorage 创建配额策略存储
func NewQuotaPolicyStorage(db *storage.DB) *QuotaPolicyStorage {
	return &QuotaPolicyStorage{db: db}
}

// GetByCategoryID 获取类别的配额策略
func (s *QuotaPolicyStorage) GetByCategoryID(ctx context.Context, categoryID int64) ([]*QuotaPolicy, error) {
	query := `
		SELECT id, name, category_id, max_count, replenish_rate, replenish_unit, created_at, updated_at
		FROM quota_policies
		WHERE category_id = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*QuotaPolicy
	for rows.Next() {
		var p QuotaPolicy
		err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.MaxCount, &p.ReplenishRate,
			&p.ReplenishUnit, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		policies = append(policies, &p)
	}
	return policies, nil
}

// AllocationStorage 分配记录存储
type AllocationStorage struct {
	db *storage.DB
}

// NewAllocationStorage 创建分配存储
func NewAllocationStorage(db *storage.DB) *AllocationStorage {
	return &AllocationStorage{db: db}
}

// Create 创建分配记录
func (s *AllocationStorage) Create(ctx context.Context, alloc *Allocation) error {
	query := `
		INSERT INTO allocations (resource_uuid, policy_uuid, task_uuid, allocated_at, status)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(ctx, query, alloc.ResourceUUID, alloc.PolicyUUID, alloc.TaskUUID,
		alloc.AllocatedAt, alloc.Status)
	return err
}

// Release 释放分配记录
func (s *AllocationStorage) Release(ctx context.Context, resourceUUID string) error {
	query := `
		UPDATE allocations SET released_at = NOW(), status = 'released'
		WHERE resource_uuid = ? AND status = 'active'
	`
	_, err := s.db.Exec(ctx, query, resourceUUID)
	return err
}

// GetActiveByResourceUUID 获取资源的活跃分配
func (s *AllocationStorage) GetActiveByResourceUUID(ctx context.Context, resourceUUID string) (*Allocation, error) {
	query := `
		SELECT id, resource_uuid, policy_uuid, task_uuid, allocated_at, released_at, status
		FROM allocations
		WHERE resource_uuid = ? AND status = 'active'
		ORDER BY allocated_at DESC
		LIMIT 1
	`
	var a Allocation
	err := s.db.QueryRow(ctx, query, resourceUUID).Scan(
		&a.ID, &a.ResourceUUID, &a.PolicyUUID, &a.TaskUUID,
		&a.AllocatedAt, &a.ReleasedAt, &a.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

// TestbedStorage 测试床存储
type TestbedStorage struct {
	db *storage.DB
}

// NewTestbedStorage 创建测试床存储
func NewTestbedStorage(db *storage.DB) *TestbedStorage {
	return &TestbedStorage{db: db}
}

// List 列出所有测试床
func (s *TestbedStorage) List(ctx context.Context) ([]*Testbed, error) {
	query := `SELECT id, name, ip_address, ssh_port, ssh_user, ssh_password, capacity, created_at, updated_at
	            FROM testbeds ORDER BY name`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testbeds []*Testbed
	for rows.Next() {
		var t Testbed
		err := rows.Scan(&t.ID, &t.Name, &t.IPAddress, &t.SSHPort, &t.SSHUser, &t.SSHPassword,
			&t.Capacity, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		testbeds = append(testbeds, &t)
	}
	return testbeds, nil
}
