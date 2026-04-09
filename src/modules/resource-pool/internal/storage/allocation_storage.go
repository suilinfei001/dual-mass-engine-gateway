package storage

import (
	"database/sql"
	"fmt"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// AllocationStorage Allocation 存储接口
type AllocationStorage interface {
	// CreateAllocation 创建新 Allocation
	CreateAllocation(allocation *models.Allocation) error

	// GetAllocation 根据 ID 获取 Allocation
	GetAllocation(id int) (*models.Allocation, error)

	// GetAllocationByUUID 根据 UUID 获取 Allocation
	GetAllocationByUUID(uuid string) (*models.Allocation, error)

	// ListAllocations 列出所有 Allocation
	ListAllocations() ([]*models.Allocation, error)

	// ListAllocationsByRequester 根据请求者列出 Allocation
	ListAllocationsByRequester(requester string) ([]*models.Allocation, error)

	// ListAllocationsByTestbed 根据 Testbed UUID 列出 Allocation
	ListAllocationsByTestbed(testbedUUID string) ([]*models.Allocation, error)

	// ListAllocationsByStatus 根据状态列出 Allocation
	ListAllocationsByStatus(status models.AllocationStatus) ([]*models.Allocation, error)

	// ListActiveAllocations 列出活跃的 Allocation
	ListActiveAllocations() ([]*models.Allocation, error)

	// ListExpiredAllocations 列出已过期的 Allocation
	ListExpiredAllocations() ([]*models.Allocation, error)

	// UpdateAllocation 更新 Allocation
	UpdateAllocation(allocation *models.Allocation) error

	// UpdateAllocationStatus 更新 Allocation 状态
	UpdateAllocationStatus(uuid string, status models.AllocationStatus) error

	// MarkAllocationReleased 标记 Allocation 为已释放
	MarkAllocationReleased(uuid string) error

	// MarkAllocationExpired 标记 Allocation 为已过期
	MarkAllocationExpired(uuid string) error

	// DeleteAllocation 删除 Allocation
	DeleteAllocation(id int) error

	// DeleteAllocationByUUID 通过 UUID 删除 Allocation
	DeleteAllocationByUUID(uuid string) error

	// CountActiveAllocationsByCategory 统计类别的活跃 Allocation 数量
	CountActiveAllocationsByCategory(categoryUUID string) (int, error)

	// CountActiveAllocationsByCategoryAndServiceTarget 统计类别和服务对象的活跃 Allocation 数量
	CountActiveAllocationsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error)

	// CountActiveAllocationsByRequester 统计请求者的活跃 Allocation 数量
	CountActiveAllocationsByRequester(requester string) (int, error)

	// DeleteAllAllocations 删除所有 Allocation
	DeleteAllAllocations() error
}

// MySQLAllocationStorage MySQL Allocation 存储实现
type MySQLAllocationStorage struct {
	db *sql.DB
}

// NewMySQLAllocationStorage 创建 MySQL Allocation 存储
func NewMySQLAllocationStorage(db *sql.DB) *MySQLAllocationStorage {
	return &MySQLAllocationStorage{db: db}
}

// CreateAllocation 创建新 Allocation
func (s *MySQLAllocationStorage) CreateAllocation(allocation *models.Allocation) error {
	query := `
		INSERT INTO allocations (
			uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var requesterComment sql.NullString
	if allocation.RequesterComment != nil {
		requesterComment = sql.NullString{String: *allocation.RequesterComment, Valid: true}
	}

	var expiresAt sql.NullTime
	if allocation.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *allocation.ExpiresAt, Valid: true}
	}

	result, err := s.db.Exec(
		query,
		allocation.UUID, allocation.TestbedUUID, allocation.CategoryUUID,
		allocation.Requester, requesterComment, allocation.Status,
		expiresAt, allocation.ReleasedAt, allocation.CreatedAt, allocation.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create allocation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	allocation.ID = int(id)
	return nil
}

// GetAllocation 根据 ID 获取 Allocation
func (s *MySQLAllocationStorage) GetAllocation(id int) (*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE id = ?
	`

	return s.scanAllocation(s.db.QueryRow(query, id))
}

// GetAllocationByUUID 根据 UUID 获取 Allocation
func (s *MySQLAllocationStorage) GetAllocationByUUID(uuid string) (*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE uuid = ?
	`

	return s.scanAllocation(s.db.QueryRow(query, uuid))
}

// ListAllocations 列出所有 Allocation
func (s *MySQLAllocationStorage) ListAllocations() ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		ORDER BY created_at DESC
	`

	return s.listAllocationsByQuery(query)
}

// ListAllocationsByRequester 根据请求者列出 Allocation
func (s *MySQLAllocationStorage) ListAllocationsByRequester(requester string) ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE requester = ?
		ORDER BY created_at DESC
	`

	return s.listAllocationsByQuery(query, requester)
}

// ListAllocationsByTestbed 根据 Testbed UUID 列出 Allocation
func (s *MySQLAllocationStorage) ListAllocationsByTestbed(testbedUUID string) ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE testbed_uuid = ?
		ORDER BY created_at DESC
	`

	return s.listAllocationsByQuery(query, testbedUUID)
}

// ListAllocationsByStatus 根据状态列出 Allocation
func (s *MySQLAllocationStorage) ListAllocationsByStatus(status models.AllocationStatus) ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE status = ?
		ORDER BY created_at DESC
	`

	return s.listAllocationsByQuery(query, status)
}

// ListActiveAllocations 列出活跃的 Allocation
func (s *MySQLAllocationStorage) ListActiveAllocations() ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE status = 'active'
		ORDER BY created_at DESC
	`

	return s.listAllocationsByQuery(query)
}

// ListExpiredAllocations 列出已过期的 Allocation
func (s *MySQLAllocationStorage) ListExpiredAllocations() ([]*models.Allocation, error) {
	query := `
		SELECT id, uuid, testbed_uuid, category_uuid, requester, requester_comment,
			status, expires_at, released_at, created_at, updated_at
		FROM allocations
		WHERE status = 'active'
			AND expires_at IS NOT NULL
			AND expires_at < NOW()
		ORDER BY expires_at ASC
	`

	return s.listAllocationsByQuery(query)
}

// UpdateAllocation 更新 Allocation
func (s *MySQLAllocationStorage) UpdateAllocation(allocation *models.Allocation) error {
	query := `
		UPDATE allocations SET
			testbed_uuid = ?, category_uuid = ?, requester = ?, requester_comment = ?,
			status = ?, expires_at = ?, released_at = ?, updated_at = ?
		WHERE id = ?
	`

	var requesterComment sql.NullString
	if allocation.RequesterComment != nil {
		requesterComment = sql.NullString{String: *allocation.RequesterComment, Valid: true}
	}

	var expiresAt sql.NullTime
	if allocation.ExpiresAt != nil {
		expiresAt = sql.NullTime{Time: *allocation.ExpiresAt, Valid: true}
	}

	var releasedAt sql.NullTime
	if allocation.ReleasedAt != nil {
		releasedAt = sql.NullTime{Time: *allocation.ReleasedAt, Valid: true}
	}

	_, err := s.db.Exec(
		query,
		allocation.TestbedUUID, allocation.CategoryUUID, allocation.Requester,
		requesterComment, allocation.Status, expiresAt, releasedAt,
		allocation.UpdatedAt, allocation.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update allocation: %w", err)
	}

	return nil
}

// UpdateAllocationStatus 更新 Allocation 状态
func (s *MySQLAllocationStorage) UpdateAllocationStatus(uuid string, status models.AllocationStatus) error {
	query := `UPDATE allocations SET status = ?, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, status, uuid)
	if err != nil {
		return fmt.Errorf("failed to update allocation status: %w", err)
	}
	return nil
}

// MarkAllocationReleased 标记 Allocation 为已释放
func (s *MySQLAllocationStorage) MarkAllocationReleased(uuid string) error {
	query := `
		UPDATE allocations
		SET status = 'released', released_at = NOW(), updated_at = NOW()
		WHERE uuid = ?
	`
	_, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to mark allocation released: %w", err)
	}
	return nil
}

// MarkAllocationExpired 标记 Allocation 为已过期
func (s *MySQLAllocationStorage) MarkAllocationExpired(uuid string) error {
	query := `
		UPDATE allocations
		SET status = 'expired', released_at = NOW(), updated_at = NOW()
		WHERE uuid = ?
	`
	_, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to mark allocation expired: %w", err)
	}
	return nil
}

// DeleteAllocation 删除 Allocation
func (s *MySQLAllocationStorage) DeleteAllocation(id int) error {
	query := `DELETE FROM allocations WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete allocation: %w", err)
	}
	return nil
}

// DeleteAllocationByUUID 通过 UUID 删除 Allocation
func (s *MySQLAllocationStorage) DeleteAllocationByUUID(uuid string) error {
	query := `DELETE FROM allocations WHERE uuid = ?`
	result, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete allocation by uuid: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("allocation not found with uuid: %s", uuid)
	}
	return nil
}

// CountActiveAllocationsByCategory 统计类别的活跃 Allocation 数量
func (s *MySQLAllocationStorage) CountActiveAllocationsByCategory(categoryUUID string) (int, error) {
	query := `SELECT COUNT(*) FROM allocations WHERE category_uuid = ? AND status = 'active'`
	var count int
	err := s.db.QueryRow(query, categoryUUID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active allocations: %w", err)
	}
	return count, nil
}

// CountActiveAllocationsByCategoryAndServiceTarget 统计类别和服务对象的活跃 Allocation 数量
// 通过 JOIN testbeds 表获取 service_target
func (s *MySQLAllocationStorage) CountActiveAllocationsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	query := `
		SELECT COUNT(*) FROM allocations
		INNER JOIN testbeds ON allocations.testbed_uuid = testbeds.uuid
		WHERE allocations.category_uuid = ?
			AND allocations.status = 'active'
			AND testbeds.service_target = ?
	`
	var count int
	err := s.db.QueryRow(query, categoryUUID, serviceTarget).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active allocations by service target: %w", err)
	}
	return count, nil
}

// CountActiveAllocationsByRequester 统计请求者的活跃 Allocation 数量
func (s *MySQLAllocationStorage) CountActiveAllocationsByRequester(requester string) (int, error) {
	query := `SELECT COUNT(*) FROM allocations WHERE requester = ? AND status = 'active'`
	var count int
	err := s.db.QueryRow(query, requester).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active allocations by requester: %w", err)
	}
	return count, nil
}

// DeleteAllAllocations 删除所有 Allocation
func (s *MySQLAllocationStorage) DeleteAllAllocations() error {
	query := `DELETE FROM allocations`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all allocations: %w", err)
	}
	return nil
}

// scanAllocation 扫描单行数据到 Allocation 对象
func (s *MySQLAllocationStorage) scanAllocation(row *sql.Row) (*models.Allocation, error) {
	allocation := &models.Allocation{}
	var requesterComment sql.NullString
	var expiresAt sql.NullTime
	var releasedAt sql.NullTime

	err := row.Scan(
		&allocation.ID, &allocation.UUID, &allocation.TestbedUUID, &allocation.CategoryUUID,
		&allocation.Requester, &requesterComment, &allocation.Status, &expiresAt,
		&releasedAt, &allocation.CreatedAt, &allocation.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("allocation not found")
		}
		return nil, fmt.Errorf("failed to scan allocation: %w", err)
	}

	if requesterComment.Valid {
		allocation.RequesterComment = &requesterComment.String
	}

	if expiresAt.Valid {
		allocation.ExpiresAt = &expiresAt.Time
	}

	if releasedAt.Valid {
		allocation.ReleasedAt = &releasedAt.Time
	}

	return allocation, nil
}

// listAllocationsByQuery 根据查询列出 Allocation
func (s *MySQLAllocationStorage) listAllocationsByQuery(query string, args ...interface{}) ([]*models.Allocation, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query allocations: %w", err)
	}
	defer rows.Close()

	var allocations []*models.Allocation
	for rows.Next() {
		allocation := &models.Allocation{}
		var requesterComment sql.NullString
		var expiresAt sql.NullTime
		var releasedAt sql.NullTime

		err := rows.Scan(
			&allocation.ID, &allocation.UUID, &allocation.TestbedUUID, &allocation.CategoryUUID,
			&allocation.Requester, &requesterComment, &allocation.Status, &expiresAt,
			&releasedAt, &allocation.CreatedAt, &allocation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan allocation: %w", err)
		}

		if requesterComment.Valid {
			allocation.RequesterComment = &requesterComment.String
		}

		if expiresAt.Valid {
			allocation.ExpiresAt = &expiresAt.Time
		}

		if releasedAt.Valid {
			allocation.ReleasedAt = &releasedAt.Time
		}

		allocations = append(allocations, allocation)
	}

	return allocations, nil
}
