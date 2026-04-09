package storage

import (
	"database/sql"
	"fmt"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// TestbedStorage Testbed 存储接口
type TestbedStorage interface {
	// CreateTestbed 创建新 Testbed
	CreateTestbed(testbed *models.Testbed) error

	// GetTestbed 根据 ID 获取 Testbed
	GetTestbed(id int) (*models.Testbed, error)

	// GetTestbedByUUID 根据 UUID 获取 Testbed
	GetTestbedByUUID(uuid string) (*models.Testbed, error)

	// ListTestbeds 列出所有 Testbed
	ListTestbeds() ([]*models.Testbed, error)

	// ListTestbedsByCategory 根据类别 UUID 列出 Testbed
	ListTestbedsByCategory(categoryUUID string) ([]*models.Testbed, error)

	// ListTestbedsByStatus 根据状态列出 Testbed
	ListTestbedsByStatus(status models.TestbedStatus) ([]*models.Testbed, error)

	// ListAvailableTestbeds 列出可用的 Testbed
	ListAvailableTestbeds(categoryUUID string) ([]*models.Testbed, error)

	// ListAvailableTestbedsByServiceTarget 根据服务对象列出可用的 Testbed
	ListAvailableTestbedsByServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) ([]*models.Testbed, error)

	// UpdateTestbed 更新 Testbed
	UpdateTestbed(testbed *models.Testbed) error

	// UpdateTestbedStatus 更新 Testbed 状态
	UpdateTestbedStatus(uuid string, status models.TestbedStatus) error

	// UpdateTestbedAllocation 更新 Testbed 分配信息
	UpdateTestbedAllocation(testbedUUID, allocUUID string, status models.TestbedStatus) error

	// ClearTestbedAllocation 清除 Testbed 分配信息
	ClearTestbedAllocation(testbedUUID string) error

	// UpdateTestbedHealthCheck 更新健康检查时间
	UpdateTestbedHealthCheck(uuid string) error

	// DeleteTestbed 删除 Testbed
	DeleteTestbed(id int) error

	// CountTestbedsByCategory 统计类别的 Testbed 数量
	CountTestbedsByCategory(categoryUUID string) (int, error)

	// CountAvailableTestbedsByCategory 统计类别的可用 Testbed 数量
	CountAvailableTestbedsByCategory(categoryUUID string, serviceTarget models.ServiceTarget) (int, error)

	// CountAllAvailableTestbeds 统计所有可用 Testbed 数量
	CountAllAvailableTestbeds() (int, error)

	// CountAllocatedTestbedsByCategory 统计类别的已分配 Testbed 数量
	CountAllocatedTestbedsByCategory(categoryUUID string) (int, error)

	// CountTestbedsByCategoryAndServiceTarget 统计类别和服务对象的 Testbed 数量
	CountTestbedsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error)

	// DeleteAllTestbeds 删除所有 Testbed
	DeleteAllTestbeds() error

	// ListTestbedsWithPagination 分页列出 Testbed
	ListTestbedsWithPagination(page, pageSize int, status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, int, error)
}

// MySQLTestbedStorage MySQL Testbed 存储实现
type MySQLTestbedStorage struct {
	db *sql.DB
}

// NewMySQLTestbedStorage 创建 MySQL Testbed 存储
func NewMySQLTestbedStorage(db *sql.DB) *MySQLTestbedStorage {
	return &MySQLTestbedStorage{db: db}
}

// CreateTestbed 创建新 Testbed
func (s *MySQLTestbedStorage) CreateTestbed(testbed *models.Testbed) error {
	query := `
		INSERT INTO testbeds (
			uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(
		query,
		testbed.UUID, testbed.Name, testbed.CategoryUUID, testbed.ServiceTarget, testbed.ResourceInstanceUUID,
		testbed.CurrentAllocUUID, testbed.MariaDBPort, testbed.MariaDBUser, testbed.MariaDBPasswd,
		testbed.Status, testbed.LastHealthCheck, testbed.CreatedAt, testbed.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create testbed: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	testbed.ID = int(id)
	return nil
}

// GetTestbed 根据 ID 获取 Testbed
func (s *MySQLTestbedStorage) GetTestbed(id int) (*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE id = ?
	`

	return s.scanTestbed(s.db.QueryRow(query, id))
}

// GetTestbedByUUID 根据 UUID 获取 Testbed
func (s *MySQLTestbedStorage) GetTestbedByUUID(uuid string) (*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE uuid = ?
	`

	return s.scanTestbed(s.db.QueryRow(query, uuid))
}

// ListTestbeds 列出所有 Testbed
func (s *MySQLTestbedStorage) ListTestbeds() ([]*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		ORDER BY created_at DESC
	`

	return s.listTestbedsByQuery(query)
}

// ListTestbedsByCategory 根据类别 UUID 列出 Testbed
func (s *MySQLTestbedStorage) ListTestbedsByCategory(categoryUUID string) ([]*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE category_uuid = ?
		ORDER BY created_at DESC
	`

	return s.listTestbedsByQuery(query, categoryUUID)
}

// ListTestbedsByStatus 根据状态列出 Testbed
func (s *MySQLTestbedStorage) ListTestbedsByStatus(status models.TestbedStatus) ([]*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE status = ?
		ORDER BY created_at DESC
	`

	return s.listTestbedsByQuery(query, status)
}

// ListAvailableTestbeds 列出可用的 Testbed
func (s *MySQLTestbedStorage) ListAvailableTestbeds(categoryUUID string) ([]*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE status = 'available'
			AND category_uuid = ?
		ORDER BY created_at DESC
	`

	return s.listTestbedsByQuery(query, categoryUUID)
}

// ListAvailableTestbedsByServiceTarget 根据服务对象列出可用的 Testbed
func (s *MySQLTestbedStorage) ListAvailableTestbedsByServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) ([]*models.Testbed, error) {
	query := `
		SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
			current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
			status, last_health_check, created_at, updated_at
		FROM testbeds
		WHERE status = 'available'
			AND category_uuid = ?
			AND service_target = ?
		ORDER BY created_at DESC
	`

	return s.listTestbedsByQuery(query, categoryUUID, serviceTarget)
}

// UpdateTestbed 更新 Testbed
func (s *MySQLTestbedStorage) UpdateTestbed(testbed *models.Testbed) error {
	query := `
		UPDATE testbeds SET
			name = ?, category_uuid = ?, service_target = ?, resource_instance_uuid = ?,
			current_alloc_uuid = ?, mariadb_port = ?, mariadb_user = ?, mariadb_passwd = ?,
			status = ?, last_health_check = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(
		query,
		testbed.Name, testbed.CategoryUUID, testbed.ServiceTarget, testbed.ResourceInstanceUUID,
		testbed.CurrentAllocUUID, testbed.MariaDBPort, testbed.MariaDBUser, testbed.MariaDBPasswd,
		testbed.Status, testbed.LastHealthCheck, testbed.UpdatedAt, testbed.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update testbed: %w", err)
	}

	return nil
}

// UpdateTestbedStatus 更新 Testbed 状态
func (s *MySQLTestbedStorage) UpdateTestbedStatus(uuid string, status models.TestbedStatus) error {
	query := `UPDATE testbeds SET status = ?, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, status, uuid)
	if err != nil {
		return fmt.Errorf("failed to update testbed status: %w", err)
	}
	return nil
}

// UpdateTestbedAllocation 更新 Testbed 分配信息（原子操作）
func (s *MySQLTestbedStorage) UpdateTestbedAllocation(testbedUUID, allocUUID string, status models.TestbedStatus) error {
	query := `
		UPDATE testbeds
		SET current_alloc_uuid = ?, status = ?, updated_at = NOW()
		WHERE uuid = ? AND status = 'available'
	`
	result, err := s.db.Exec(query, allocUUID, status, testbedUUID)
	if err != nil {
		return fmt.Errorf("failed to update testbed allocation: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("testbed not available for allocation")
	}
	return nil
}

// ClearTestbedAllocation 清除 Testbed 分配信息
func (s *MySQLTestbedStorage) ClearTestbedAllocation(testbedUUID string) error {
	query := `
		UPDATE testbeds
		SET current_alloc_uuid = NULL, status = 'available', updated_at = NOW()
		WHERE uuid = ?
	`
	_, err := s.db.Exec(query, testbedUUID)
	if err != nil {
		return fmt.Errorf("failed to clear testbed allocation: %w", err)
	}
	return nil
}

// UpdateTestbedHealthCheck 更新健康检查时间
func (s *MySQLTestbedStorage) UpdateTestbedHealthCheck(uuid string) error {
	query := `UPDATE testbeds SET last_health_check = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(query, uuid)
	if err != nil {
		return fmt.Errorf("failed to update testbed health check: %w", err)
	}
	return nil
}

// DeleteTestbed 删除 Testbed
func (s *MySQLTestbedStorage) DeleteTestbed(id int) error {
	query := `DELETE FROM testbeds WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete testbed: %w", err)
	}
	return nil
}

// CountTestbedsByCategory 统计类别的 Testbed 数量（不包含 deleted 状态）
func (s *MySQLTestbedStorage) CountTestbedsByCategory(categoryUUID string) (int, error) {
	query := `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ? AND status != 'deleted'`
	var count int
	err := s.db.QueryRow(query, categoryUUID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count testbeds: %w", err)
	}
	return count, nil
}

// CountAvailableTestbedsByCategory 统计类别的可用 Testbed 数量
func (s *MySQLTestbedStorage) CountAvailableTestbedsByCategory(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	query := `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ? AND service_target = ? AND status = 'available'`
	var count int
	err := s.db.QueryRow(query, categoryUUID, serviceTarget).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count available testbeds: %w", err)
	}
	return count, nil
}

// CountAllAvailableTestbeds 统计所有可用 Testbed 数量
func (s *MySQLTestbedStorage) CountAllAvailableTestbeds() (int, error) {
	query := `SELECT COUNT(*) FROM testbeds WHERE status = 'available' AND status != 'deleted'`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count available testbeds: %w", err)
	}
	return count, nil
}

// CountAllocatedTestbedsByCategory 统计类别的已分配 Testbed 数量
func (s *MySQLTestbedStorage) CountAllocatedTestbedsByCategory(categoryUUID string) (int, error) {
	query := `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ? AND status IN ('allocated', 'in_use')`
	var count int
	err := s.db.QueryRow(query, categoryUUID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count allocated testbeds: %w", err)
	}
	return count, nil
}

// CountTestbedsByCategoryAndServiceTarget 统计类别和服务对象的 Testbed 数量
func (s *MySQLTestbedStorage) CountTestbedsByCategoryAndServiceTarget(categoryUUID string, serviceTarget models.ServiceTarget) (int, error) {
	query := `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ? AND service_target = ?`
	var count int
	err := s.db.QueryRow(query, categoryUUID, serviceTarget).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count testbeds by category and service target: %w", err)
	}
	return count, nil
}

// scanTestbed 扫描单行数据到 Testbed 对象
func (s *MySQLTestbedStorage) scanTestbed(row *sql.Row) (*models.Testbed, error) {
	testbed := &models.Testbed{}
	var currentAllocUUID sql.NullString
	var lastHealthCheck sql.NullTime

	err := row.Scan(
		&testbed.ID, &testbed.UUID, &testbed.Name, &testbed.CategoryUUID, &testbed.ServiceTarget, &testbed.ResourceInstanceUUID,
		&currentAllocUUID, &testbed.MariaDBPort, &testbed.MariaDBUser, &testbed.MariaDBPasswd,
		&testbed.Status, &lastHealthCheck, &testbed.CreatedAt, &testbed.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("testbed not found")
		}
		return nil, fmt.Errorf("failed to scan testbed: %w", err)
	}

	if currentAllocUUID.Valid {
		testbed.CurrentAllocUUID = &currentAllocUUID.String
	}

	if lastHealthCheck.Valid {
		testbed.LastHealthCheck = lastHealthCheck.Time
	} else {
		testbed.LastHealthCheck = testbed.CreatedAt
	}

	return testbed, nil
}

// listTestbedsByQuery 根据查询列出 Testbed
func (s *MySQLTestbedStorage) listTestbedsByQuery(query string, args ...interface{}) ([]*models.Testbed, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query testbeds: %w", err)
	}
	defer rows.Close()

	var testbeds []*models.Testbed
	for rows.Next() {
		testbed := &models.Testbed{}
		var currentAllocUUID sql.NullString
		var lastHealthCheck sql.NullTime

		err := rows.Scan(
			&testbed.ID, &testbed.UUID, &testbed.Name, &testbed.CategoryUUID, &testbed.ServiceTarget, &testbed.ResourceInstanceUUID,
			&currentAllocUUID, &testbed.MariaDBPort, &testbed.MariaDBUser, &testbed.MariaDBPasswd,
			&testbed.Status, &lastHealthCheck, &testbed.CreatedAt, &testbed.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan testbed: %w", err)
		}

		if currentAllocUUID.Valid {
			testbed.CurrentAllocUUID = &currentAllocUUID.String
		}

		if lastHealthCheck.Valid {
			testbed.LastHealthCheck = lastHealthCheck.Time
		} else {
			testbed.LastHealthCheck = testbed.CreatedAt
		}

		testbeds = append(testbeds, testbed)
	}

	return testbeds, nil
}

// TryAllocateTestbed 尝试分配 Testbed（原子操作）
// 只有当状态为 'available' 时才会更新为 'allocated'，并返回 true 表示成功
func (s *MySQLTestbedStorage) TryAllocateTestbed(testbedUUID, allocUUID string) (bool, error) {
	query := `
		UPDATE testbeds
		SET current_alloc_uuid = ?, status = 'allocated', updated_at = NOW()
		WHERE uuid = ? AND status = 'available'
	`
	result, err := s.db.Exec(query, allocUUID, testbedUUID)
	if err != nil {
		return false, fmt.Errorf("failed to try allocate testbed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

// DeleteAllTestbeds 删除所有 Testbed
func (s *MySQLTestbedStorage) DeleteAllTestbeds() error {
	query := `DELETE FROM testbeds`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all testbeds: %w", err)
	}
	return nil
}

// ListTestbedsWithPagination 分页列出 Testbed
func (s *MySQLTestbedStorage) ListTestbedsWithPagination(page, pageSize int, status *models.TestbedStatus, categoryUUID *string) ([]*models.Testbed, int, error) {
	var query string
	var countQuery string
	var args []interface{}

	if status != nil && categoryUUID != nil {
		query = `
			SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
				current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
				status, last_health_check, created_at, updated_at
			FROM testbeds
			WHERE status = ? AND category_uuid = ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = `SELECT COUNT(*) FROM testbeds WHERE status = ? AND category_uuid = ?`
		args = []interface{}{*status, *categoryUUID, pageSize, (page - 1) * pageSize}
	} else if status != nil {
		query = `
			SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
				current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
				status, last_health_check, created_at, updated_at
			FROM testbeds
			WHERE status = ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = `SELECT COUNT(*) FROM testbeds WHERE status = ?`
		args = []interface{}{*status, pageSize, (page - 1) * pageSize}
	} else if categoryUUID != nil {
		query = `
			SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
				current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
				status, last_health_check, created_at, updated_at
			FROM testbeds
			WHERE category_uuid = ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = `SELECT COUNT(*) FROM testbeds WHERE category_uuid = ?`
		args = []interface{}{*categoryUUID, pageSize, (page - 1) * pageSize}
	} else {
		query = `
			SELECT id, uuid, name, category_uuid, service_target, resource_instance_uuid,
				current_alloc_uuid, mariadb_port, mariadb_user, mariadb_passwd,
				status, last_health_check, created_at, updated_at
			FROM testbeds
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = `SELECT COUNT(*) FROM testbeds`
		args = []interface{}{pageSize, (page - 1) * pageSize}
	}

	var total int
	err := s.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count testbeds: %w", err)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query testbeds: %w", err)
	}
	defer rows.Close()

	var testbeds []*models.Testbed
	for rows.Next() {
		testbed := &models.Testbed{}
		var currentAllocUUID sql.NullString
		var lastHealthCheck sql.NullTime

		err := rows.Scan(
			&testbed.ID, &testbed.UUID, &testbed.Name, &testbed.CategoryUUID, &testbed.ServiceTarget, &testbed.ResourceInstanceUUID,
			&currentAllocUUID, &testbed.MariaDBPort, &testbed.MariaDBUser, &testbed.MariaDBPasswd,
			&testbed.Status, &lastHealthCheck, &testbed.CreatedAt, &testbed.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan testbed: %w", err)
		}

		if currentAllocUUID.Valid {
			testbed.CurrentAllocUUID = &currentAllocUUID.String
		}

		if lastHealthCheck.Valid {
			testbed.LastHealthCheck = lastHealthCheck.Time
		} else {
			testbed.LastHealthCheck = testbed.CreatedAt
		}

		testbeds = append(testbeds, testbed)
	}

	return testbeds, total, nil
}
