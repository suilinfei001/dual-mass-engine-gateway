package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TestbedStatus Testbed 状态
type TestbedStatus string

const (
	TestbedStatusAvailable  TestbedStatus = "available"
	TestbedStatusAllocated  TestbedStatus = "allocated"
	TestbedStatusInUse      TestbedStatus = "in_use"
	TestbedStatusReleasing  TestbedStatus = "releasing"
	TestbedStatusDeleted    TestbedStatus = "deleted"
)

// Testbed 测试床 - 部署了产品的可用测试环境
type Testbed struct {
	ID                   int            `json:"id"`
	UUID                 string         `json:"uuid"`
	Name                 string         `json:"name"`
	CategoryUUID         string         `json:"category_uuid"`
	ServiceTarget        ServiceTarget  `json:"service_target"`
	ResourceInstanceUUID string         `json:"resource_instance_uuid"`
	CurrentAllocUUID     *string        `json:"current_alloc_uuid,omitempty"`
	MariaDBPort          int            `json:"mariadb_port"`
	MariaDBUser          string         `json:"mariadb_user"`
	MariaDBPasswd        string         `json:"mariadb_passwd"`
	Status               TestbedStatus  `json:"status"`
	LastHealthCheck      time.Time      `json:"last_health_check"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// TestbedResponse API 响应格式（支持密码掩码）
type TestbedResponse struct {
	ID                   int                    `json:"id"`
	UUID                 string                 `json:"uuid"`
	Name                 string                 `json:"name"`
	CategoryUUID         string                 `json:"category_uuid"`
	CategoryName         *string                `json:"category_name,omitempty"`
	ServiceTarget        ServiceTarget         `json:"service_target"`
	ResourceInstanceUUID string                 `json:"resource_instance_uuid"`
	ResourceInstance     *ResourceInstanceInfo  `json:"resource_instance,omitempty"`
	CurrentAllocUUID     *string                `json:"current_alloc_uuid,omitempty"`
	// 兼容前端的字段名
	MariaDBPort          int    `json:"mariadb_port"`
	DBPort               int    `json:"db_port"`
	MariaDBUser          string `json:"mariadb_user"`
	DBUser               string `json:"db_user"`
	MariaDBPasswd        string `json:"mariadb_passwd"`
	DBPassword           string `json:"db_password"`
	Host                 string `json:"host,omitempty"`
	IPAddress            string `json:"ip_address,omitempty"`
	SSHPort              int    `json:"ssh_port,omitempty"`
	SSHUser              string `json:"ssh_user,omitempty"`
	SSHPassword          string `json:"ssh_password,omitempty"`
	Status               TestbedStatus `json:"status"`
	LastHealthCheck      string `json:"last_health_check"`
	LastHealthCheckAt    string `json:"last_health_check_at,omitempty"`
	ExpiresAt            *string `json:"expires_at,omitempty"` // 分配过期时间（仅在已分配时有值）
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// ResourceInstanceInfo 资源实例简要信息
type ResourceInstanceInfo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	Port        int    `json:"ssh_port,omitempty"`
	InstanceType string `json:"resource_type,omitempty"`
	SnapshotID  string `json:"snapshot_id,omitempty"`
	Status      string `json:"status,omitempty"`
}

// IsAvailable 检查 Testbed 是否可用
func (t *Testbed) IsAvailable() bool {
	return t.Status == TestbedStatusAvailable
}

// IsAllocated 检查 Testbed 是否已分配
func (t *Testbed) IsAllocated() bool {
	return t.Status == TestbedStatusAllocated || t.Status == TestbedStatusInUse
}

// MarkAvailable 标记为可用状态
func (t *Testbed) MarkAvailable() {
	now := time.Now()
	t.Status = TestbedStatusAvailable
	t.CurrentAllocUUID = nil
	t.UpdatedAt = now
}

// MarkAllocated 标记为已分配状态
func (t *Testbed) MarkAllocated(allocUUID string) {
	now := time.Now()
	t.Status = TestbedStatusAllocated
	t.CurrentAllocUUID = &allocUUID
	t.UpdatedAt = now
}

// MarkInUse 标记为使用中状态
func (t *Testbed) MarkInUse() {
	now := time.Now()
	t.Status = TestbedStatusInUse
	t.UpdatedAt = now
}

// MarkReleasing 标记为释放中状态
func (t *Testbed) MarkReleasing() {
	now := time.Now()
	t.Status = TestbedStatusReleasing
	t.UpdatedAt = now
}

// MarkDeleted 标记为已删除状态
// Testbed 是一次性的，释放后应标记为 deleted 而不是回到 available
func (t *Testbed) MarkDeleted() {
	now := time.Now()
	t.Status = TestbedStatusDeleted
	t.CurrentAllocUUID = nil
	t.UpdatedAt = now
}

// ToResponse 转换为 API 响应格式
func (t *Testbed) ToResponse() TestbedResponse {
	lastHealthCheck := t.LastHealthCheck.Format(time.RFC3339)
	return TestbedResponse{
		ID:                   t.ID,
		UUID:                 t.UUID,
		Name:                 t.Name,
		CategoryUUID:         t.CategoryUUID,
		ServiceTarget:        t.ServiceTarget,
		ResourceInstanceUUID: t.ResourceInstanceUUID,
		CurrentAllocUUID:     t.CurrentAllocUUID,
		// 原始字段名
		MariaDBPort:          t.MariaDBPort,
		MariaDBUser:          t.MariaDBUser,
		MariaDBPasswd:        t.MariaDBPasswd,
		// 兼容前端的字段名
		DBPort:               t.MariaDBPort,
		DBUser:               t.MariaDBUser,
		DBPassword:           t.MariaDBPasswd,
		Status:               t.Status,
		LastHealthCheck:      lastHealthCheck,
		LastHealthCheckAt:    lastHealthCheck,
		CreatedAt:            t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            t.UpdatedAt.Format(time.RFC3339),
	}
}

// ToResponseWithMaskedPassword 转换为 API 响应格式（密码掩码）
func (t *Testbed) ToResponseWithMaskedPassword(showPassword bool) TestbedResponse {
	resp := t.ToResponse()
	if !showPassword {
		resp.MariaDBPasswd = "****"
		resp.DBPassword = "****"
	}
	return resp
}

// NewTestbed 创建新的 Testbed
func NewTestbed(name, categoryUUID string, serviceTarget ServiceTarget, resourceInstanceUUID string, mariaDBPort int, mariaDBUser, mariaDBPasswd string) *Testbed {
	now := time.Now()
	return &Testbed{
		UUID:                 uuid.New().String(),
		Name:                 name,
		CategoryUUID:         categoryUUID,
		ServiceTarget:        serviceTarget,
		ResourceInstanceUUID: resourceInstanceUUID,
		MariaDBPort:          mariaDBPort,
		MariaDBUser:          mariaDBUser,
		MariaDBPasswd:        mariaDBPasswd,
		Status:               TestbedStatusAvailable,
		LastHealthCheck:      now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// GenerateTestbedName 生成 Testbed 名称
// 格式: testbed_[category name]_[testbed uuid]_[timestamp]
func GenerateTestbedName(categoryName, testbedUUID string) string {
	// 对 category name 进行清理，替换特殊字符为下划线
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}, categoryName)

	// 使用 UUID 的前 8 位
	shortUUID := testbedUUID
	if len(testbedUUID) > 8 {
		shortUUID = testbedUUID[:8]
	}

	timestamp := time.Now().Unix()
	return fmt.Sprintf("testbed_%s_%s_%d", cleanName, shortUUID, timestamp)
}

// ParseTestbedStatus 解析 Testbed 状态字符串
func ParseTestbedStatus(s string) (TestbedStatus, error) {
	switch s {
	case string(TestbedStatusAvailable):
		return TestbedStatusAvailable, nil
	case string(TestbedStatusAllocated):
		return TestbedStatusAllocated, nil
	case string(TestbedStatusInUse):
		return TestbedStatusInUse, nil
	case string(TestbedStatusReleasing):
		return TestbedStatusReleasing, nil
	case string(TestbedStatusDeleted):
		return TestbedStatusDeleted, nil
	default:
		return "", fmt.Errorf("invalid testbed status: %s", s)
	}
}

// Value 实现 driver.Valuer 接口
func (ts TestbedStatus) Value() (driver.Value, error) {
	return string(ts), nil
}

// Scan 实现 sql.Scanner 接口
func (ts *TestbedStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*ts = TestbedStatus(v)
	case string:
		*ts = TestbedStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into TestbedStatus", value)
	}
	return nil
}
