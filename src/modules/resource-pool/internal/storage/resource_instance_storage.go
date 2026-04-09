package storage

import (
	"database/sql"
	"fmt"

	"github.com/hugoh/go-designs/resource-pool/internal/models"
)

// ResourceInstanceStorage ResourceInstance 存储接口
type ResourceInstanceStorage interface {
	// CreateResourceInstance 创建新 ResourceInstance
	CreateResourceInstance(instance *models.ResourceInstance) error

	// GetResourceInstance 根据 ID 获取 ResourceInstance
	GetResourceInstance(id int) (*models.ResourceInstance, error)

	// GetResourceInstanceByUUID 根据 UUID 获取 ResourceInstance
	GetResourceInstanceByUUID(uuid string) (*models.ResourceInstance, error)

	// GetResourceInstanceByIPAddress 根据 IP 地址获取 ResourceInstance
	GetResourceInstanceByIPAddress(ipAddress string) (*models.ResourceInstance, error)

	// ListResourceInstances 列出所有 ResourceInstance
	ListResourceInstances() ([]*models.ResourceInstance, error)

	// ListPublicResourceInstances 列出公开的 ResourceInstance
	ListPublicResourceInstances() ([]*models.ResourceInstance, error)

	// ListPublicResourceInstancesByType 列出公开的指定类型的 ResourceInstance
	ListPublicResourceInstancesByType(instanceType models.InstanceType) ([]*models.ResourceInstance, error)

	// ListResourceInstancesByCreatedBy 根据创建者列出 ResourceInstance
	ListResourceInstancesByCreatedBy(createdBy string) ([]*models.ResourceInstance, error)

	// ListAvailableResourceInstances 列出可用的 ResourceInstance（未被 Testbed 使用）
	ListAvailableResourceInstances() ([]*models.ResourceInstance, error)

	// UpdateResourceInstance 更新 ResourceInstance
	UpdateResourceInstance(instance *models.ResourceInstance) error

	// UpdateResourceInstanceStatus 更新 ResourceInstance 状态
	UpdateResourceInstanceStatus(uuid string, status models.ResourceInstanceStatus) error

	// DeleteResourceInstance 删除 ResourceInstance
	DeleteResourceInstance(id int) error

	// DeleteAllResourceInstances 删除所有 ResourceInstance
	DeleteAllResourceInstances() error
}

// MySQLResourceInstanceStorage MySQL ResourceInstance 存储实现
type MySQLResourceInstanceStorage struct {
	db *sql.DB
}

// NewMySQLResourceInstanceStorage 创建 MySQL ResourceInstance 存储
func NewMySQLResourceInstanceStorage(db *sql.DB) *MySQLResourceInstanceStorage {
	return &MySQLResourceInstanceStorage{db: db}
}

// CreateResourceInstance 创建新 ResourceInstance
func (s *MySQLResourceInstanceStorage) CreateResourceInstance(instance *models.ResourceInstance) error {
	query := `
		INSERT INTO resource_instances (
			uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var terminatedAt sql.NullTime
	if instance.TerminatedAt != nil {
		terminatedAt = sql.NullTime{Time: *instance.TerminatedAt, Valid: true}
	}

	var description sql.NullString
	if instance.Description != nil {
		description = sql.NullString{String: *instance.Description, Valid: true}
	}

	result, err := s.db.Exec(
		query,
		instance.UUID, instance.InstanceType, instance.SnapshotID, instance.SnapshotInstanceUUID,
		instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd, description, instance.IsPublic,
		instance.CreatedBy, instance.Status, instance.CreatedAt, terminatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create resource instance: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	instance.ID = int(id)
	return nil
}

// GetResourceInstance 根据 ID 获取 ResourceInstance
func (s *MySQLResourceInstanceStorage) GetResourceInstance(id int) (*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE id = ?
	`

	return s.scanResourceInstance(s.db.QueryRow(query, id))
}

// GetResourceInstanceByUUID 根据 UUID 获取 ResourceInstance
func (s *MySQLResourceInstanceStorage) GetResourceInstanceByUUID(uuid string) (*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE uuid = ?
	`

	return s.scanResourceInstance(s.db.QueryRow(query, uuid))
}

// GetResourceInstanceByIPAddress 根据 IP 地址获取 ResourceInstance
func (s *MySQLResourceInstanceStorage) GetResourceInstanceByIPAddress(ipAddress string) (*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE ip_address = ?
	`

	return s.scanResourceInstance(s.db.QueryRow(query, ipAddress))
}

// ListResourceInstances 列出所有 ResourceInstance
func (s *MySQLResourceInstanceStorage) ListResourceInstances() ([]*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		ORDER BY created_at DESC
	`

	return s.listResourceInstancesByQuery(query)
}

// ListPublicResourceInstances 列出公开的 ResourceInstance
func (s *MySQLResourceInstanceStorage) ListPublicResourceInstances() ([]*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE is_public = true AND status = 'active'
		ORDER BY created_at DESC
	`

	return s.listResourceInstancesByQuery(query)
}

// ListPublicResourceInstancesByType 列出公开的指定类型的 ResourceInstance
func (s *MySQLResourceInstanceStorage) ListPublicResourceInstancesByType(instanceType models.InstanceType) ([]*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE is_public = true AND status = 'active' AND instance_type = ?
		ORDER BY created_at DESC
	`

	return s.listResourceInstancesByQuery(query, instanceType)
}

// ListResourceInstancesByCreatedBy 根据创建者列出 ResourceInstance
func (s *MySQLResourceInstanceStorage) ListResourceInstancesByCreatedBy(createdBy string) ([]*models.ResourceInstance, error) {
	query := `
		SELECT id, uuid, instance_type, snapshot_id, snapshot_instance_uuid, ip_address, port, ssh_user, passwd,
			description, is_public, created_by, status, created_at, terminated_at
		FROM resource_instances
		WHERE created_by = ?
		ORDER BY created_at DESC
	`

	return s.listResourceInstancesByQuery(query, createdBy)
}

// ListAvailableResourceInstances 列出可用的 ResourceInstance
// 只返回没有被活跃 Testbed (available/allocated/in_use) 关联的 ResourceInstance
// 被已删除 (deleted) 或释放中 (releasing) 的 Testbed 关联的 ResourceInstance 可以再次使用
func (s *MySQLResourceInstanceStorage) ListAvailableResourceInstances() ([]*models.ResourceInstance, error) {
	query := `
		SELECT ri.id, ri.uuid, ri.instance_type, ri.snapshot_id, ri.snapshot_instance_uuid, ri.ip_address, ri.port, ri.ssh_user, ri.passwd,
			ri.description, ri.is_public, ri.created_by, ri.status, ri.created_at, ri.terminated_at
		FROM resource_instances ri
		WHERE ri.status = 'active'
			AND ri.instance_type = 'VirtualMachine'
			AND NOT EXISTS (
				SELECT 1 FROM testbeds t
				WHERE t.resource_instance_uuid = ri.uuid
				AND t.status IN ('available', 'allocated', 'in_use')
			)
		ORDER BY ri.created_at DESC
	`

	return s.listResourceInstancesByQuery(query)
}

// UpdateResourceInstance 更新 ResourceInstance
func (s *MySQLResourceInstanceStorage) UpdateResourceInstance(instance *models.ResourceInstance) error {
	query := `
		UPDATE resource_instances SET
			snapshot_id = ?, snapshot_instance_uuid = ?, ip_address = ?, port = ?, ssh_user = ?, passwd = ?, description = ?,
			is_public = ?, status = ?, terminated_at = ?
		WHERE id = ?
	`

	var terminatedAt sql.NullTime
	if instance.TerminatedAt != nil {
		terminatedAt = sql.NullTime{Time: *instance.TerminatedAt, Valid: true}
	}

	var description sql.NullString
	if instance.Description != nil {
		description = sql.NullString{String: *instance.Description, Valid: true}
	}

	_, err := s.db.Exec(
		query,
		instance.SnapshotID, instance.SnapshotInstanceUUID, instance.IPAddress, instance.Port,
		instance.SSHUser, instance.Passwd, description, instance.IsPublic, instance.Status, terminatedAt,
		instance.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update resource instance: %w", err)
	}

	return nil
}

// UpdateResourceInstanceStatus 更新 ResourceInstance 状态
func (s *MySQLResourceInstanceStorage) UpdateResourceInstanceStatus(uuid string, status models.ResourceInstanceStatus) error {
	query := `UPDATE resource_instances SET status = ? WHERE uuid = ?`
	_, err := s.db.Exec(query, status, uuid)
	if err != nil {
		return fmt.Errorf("failed to update resource instance status: %w", err)
	}
	return nil
}

// DeleteResourceInstance 删除 ResourceInstance
func (s *MySQLResourceInstanceStorage) DeleteResourceInstance(id int) error {
	query := `DELETE FROM resource_instances WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete resource instance: %w", err)
	}
	return nil
}

// scanResourceInstance 扫描单行数据到 ResourceInstance 对象
func (s *MySQLResourceInstanceStorage) scanResourceInstance(row *sql.Row) (*models.ResourceInstance, error) {
	instance := &models.ResourceInstance{}
	var snapshotID sql.NullString
	var snapshotInstanceUUID sql.NullString
	var description sql.NullString
	var terminatedAt sql.NullTime

	err := row.Scan(
		&instance.ID, &instance.UUID, &instance.InstanceType, &snapshotID, &snapshotInstanceUUID,
		&instance.IPAddress, &instance.Port, &instance.SSHUser, &instance.Passwd, &description,
		&instance.IsPublic, &instance.CreatedBy, &instance.Status,
		&instance.CreatedAt, &terminatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resource instance not found")
		}
		return nil, fmt.Errorf("failed to scan resource instance: %w", err)
	}

	if snapshotID.Valid {
		instance.SnapshotID = &snapshotID.String
	}

	if snapshotInstanceUUID.Valid {
		instance.SnapshotInstanceUUID = &snapshotInstanceUUID.String
	}

	if description.Valid {
		instance.Description = &description.String
	}

	if terminatedAt.Valid {
		instance.TerminatedAt = &terminatedAt.Time
	}

	return instance, nil
}

// listResourceInstancesByQuery 根据查询列出 ResourceInstance
func (s *MySQLResourceInstanceStorage) listResourceInstancesByQuery(query string, args ...interface{}) ([]*models.ResourceInstance, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query resource instances: %w", err)
	}
	defer rows.Close()

	var instances []*models.ResourceInstance
	for rows.Next() {
		instance := &models.ResourceInstance{}
		var snapshotID sql.NullString
		var snapshotInstanceUUID sql.NullString
		var description sql.NullString
		var terminatedAt sql.NullTime

		err := rows.Scan(
			&instance.ID, &instance.UUID, &instance.InstanceType, &snapshotID, &snapshotInstanceUUID,
			&instance.IPAddress, &instance.Port, &instance.SSHUser, &instance.Passwd, &description,
			&instance.IsPublic, &instance.CreatedBy, &instance.Status,
			&instance.CreatedAt, &terminatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource instance: %w", err)
		}

		if snapshotID.Valid {
			instance.SnapshotID = &snapshotID.String
		}

		if snapshotInstanceUUID.Valid {
			instance.SnapshotInstanceUUID = &snapshotInstanceUUID.String
		}

		if description.Valid {
			instance.Description = &description.String
		}

		if terminatedAt.Valid {
			instance.TerminatedAt = &terminatedAt.Time
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// DeleteAllResourceInstances 删除所有 ResourceInstance
func (s *MySQLResourceInstanceStorage) DeleteAllResourceInstances() error {
	query := `DELETE FROM resource_instances`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all resource instances: %w", err)
	}
	return nil
}
