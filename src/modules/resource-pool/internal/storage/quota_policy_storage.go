package storage

import (
	"database/sql"
	"fmt"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// QuotaPolicyStorage QuotaPolicy 存储接口
type QuotaPolicyStorage interface {
	// CreateQuotaPolicy 创建新 QuotaPolicy
	CreateQuotaPolicy(policy *models.QuotaPolicy) error

	// GetQuotaPolicy 根据 ID 获取 QuotaPolicy
	GetQuotaPolicy(id int) (*models.QuotaPolicy, error)

	// GetQuotaPolicyByUUID 根据 UUID 获取 QuotaPolicy
	GetQuotaPolicyByUUID(uuid string) (*models.QuotaPolicy, error)

	// GetQuotaPolicyByCategory 根据类别 UUID 获取 QuotaPolicy
	GetQuotaPolicyByCategory(categoryUUID string) (*models.QuotaPolicy, error)

	// GetQuotaPolicyByCategoryAndServiceTarget 根据类别 UUID 和服务对象获取 QuotaPolicy
	GetQuotaPolicyByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (*models.QuotaPolicy, error)

	// ListQuotaPolicies 列出所有 QuotaPolicy
	ListQuotaPolicies() ([]*models.QuotaPolicy, error)

	// UpdateQuotaPolicy 更新 QuotaPolicy
	UpdateQuotaPolicy(policy *models.QuotaPolicy) error

	// DeleteQuotaPolicy 删除 QuotaPolicy
	DeleteQuotaPolicy(id int) error

	// ListPoliciesByPriority 按优先级列出 QuotaPolicy
	ListPoliciesByPriority() ([]*models.QuotaPolicy, error)

	// DeleteAllQuotaPolicies 删除所有 QuotaPolicy
	DeleteAllQuotaPolicies() error
}

// MySQLQuotaPolicyStorage MySQL QuotaPolicy 存储实现
type MySQLQuotaPolicyStorage struct {
	db *sql.DB
}

// NewMySQLQuotaPolicyStorage 创建 MySQL QuotaPolicy 存储
func NewMySQLQuotaPolicyStorage(db *sql.DB) *MySQLQuotaPolicyStorage {
	return &MySQLQuotaPolicyStorage{db: db}
}

// CreateQuotaPolicy 创建新 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) CreateQuotaPolicy(policy *models.QuotaPolicy) error {
	query := `
		INSERT INTO quota_policies (
			uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(
		query,
		policy.UUID, policy.CategoryUUID, policy.MinInstances, policy.MaxInstances,
		policy.Priority, policy.ServiceTarget, policy.AutoReplenish, policy.ReplenishThreshold,
		policy.MaxLifetimeSeconds, policy.CreatedAt, policy.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create quota policy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	policy.ID = int(id)
	return nil
}

// GetQuotaPolicy 根据 ID 获取 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) GetQuotaPolicy(id int) (*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		WHERE id = ?
	`

	return s.scanQuotaPolicy(s.db.QueryRow(query, id))
}

// GetQuotaPolicyByUUID 根据 UUID 获取 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) GetQuotaPolicyByUUID(uuid string) (*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		WHERE uuid = ?
	`

	return s.scanQuotaPolicy(s.db.QueryRow(query, uuid))
}

// GetQuotaPolicyByCategory 根据类别 UUID 获取 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) GetQuotaPolicyByCategory(categoryUUID string) (*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		WHERE category_uuid = ?
	`

	return s.scanQuotaPolicy(s.db.QueryRow(query, categoryUUID))
}

// GetQuotaPolicyByCategoryAndServiceTarget 根据类别 UUID 和服务对象获取 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) GetQuotaPolicyByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		WHERE category_uuid = ? AND service_target = ?
	`

	return s.scanQuotaPolicy(s.db.QueryRow(query, categoryUUID, serviceTarget))
}

// ListQuotaPolicies 列出所有 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) ListQuotaPolicies() ([]*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		ORDER BY priority DESC, created_at ASC
	`

	return s.listQuotaPoliciesByQuery(query)
}

// UpdateQuotaPolicy 更新 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) UpdateQuotaPolicy(policy *models.QuotaPolicy) error {
	query := `
		UPDATE quota_policies SET
			min_instances = ?, max_instances = ?, priority = ?, service_target = ?,
			auto_replenish = ?, replenish_threshold = ?, max_lifetime_seconds = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(
		query,
		policy.MinInstances, policy.MaxInstances, policy.Priority, policy.ServiceTarget,
		policy.AutoReplenish, policy.ReplenishThreshold,
		policy.MaxLifetimeSeconds, policy.UpdatedAt, policy.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update quota policy: %w", err)
	}

	return nil
}

// DeleteQuotaPolicy 删除 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) DeleteQuotaPolicy(id int) error {
	query := `DELETE FROM quota_policies WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete quota policy: %w", err)
	}
	return nil
}

// ListPoliciesByPriority 按优先级列出 QuotaPolicy
// priority 数值越小，优先级越高
func (s *MySQLQuotaPolicyStorage) ListPoliciesByPriority() ([]*models.QuotaPolicy, error) {
	query := `
		SELECT id, uuid, category_uuid, min_instances, max_instances, priority, service_target,
			auto_replenish, replenish_threshold, max_lifetime_seconds, created_at, updated_at
		FROM quota_policies
		ORDER BY priority ASC
	`

	return s.listQuotaPoliciesByQuery(query)
}

// scanQuotaPolicy 扫描单行数据到 QuotaPolicy 对象
func (s *MySQLQuotaPolicyStorage) scanQuotaPolicy(row *sql.Row) (*models.QuotaPolicy, error) {
	policy := &models.QuotaPolicy{}

	err := row.Scan(
		&policy.ID, &policy.UUID, &policy.CategoryUUID, &policy.MinInstances,
		&policy.MaxInstances, &policy.Priority, &policy.ServiceTarget,
		&policy.AutoReplenish, &policy.ReplenishThreshold, &policy.MaxLifetimeSeconds,
		&policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quota policy not found")
		}
		return nil, fmt.Errorf("failed to scan quota policy: %w", err)
	}

	return policy, nil
}

// listQuotaPoliciesByQuery 根据查询列出 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) listQuotaPoliciesByQuery(query string, args ...interface{}) ([]*models.QuotaPolicy, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query quota policies: %w", err)
	}
	defer rows.Close()

	var policies []*models.QuotaPolicy
	for rows.Next() {
		policy := &models.QuotaPolicy{}

		err := rows.Scan(
			&policy.ID, &policy.UUID, &policy.CategoryUUID, &policy.MinInstances,
			&policy.MaxInstances, &policy.Priority, &policy.ServiceTarget,
			&policy.AutoReplenish, &policy.ReplenishThreshold, &policy.MaxLifetimeSeconds,
			&policy.CreatedAt, &policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quota policy: %w", err)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}

// DeleteAllQuotaPolicies 删除所有 QuotaPolicy
func (s *MySQLQuotaPolicyStorage) DeleteAllQuotaPolicies() error {
	query := `DELETE FROM quota_policies`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all quota policies: %w", err)
	}
	return nil
}
