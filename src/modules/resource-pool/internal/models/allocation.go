package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AllocationStatus 分配状态
type AllocationStatus string

const (
	AllocationStatusPending  AllocationStatus = "pending"
	AllocationStatusActive   AllocationStatus = "active"
	AllocationStatusReleased AllocationStatus = "released"
	AllocationStatusExpired  AllocationStatus = "expired"
)

// Allocation 分配记录
type Allocation struct {
	ID               int                `json:"id"`
	UUID             string             `json:"uuid"`
	TestbedUUID      string             `json:"testbed_uuid"`
	CategoryUUID     string             `json:"category_uuid"`
	Requester        string             `json:"requester"`
	RequesterComment *string            `json:"requester_comment,omitempty"`
	Status           AllocationStatus   `json:"status"`
	ExpiresAt        *time.Time         `json:"expires_at,omitempty"`
	ReleasedAt       *time.Time         `json:"released_at,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// AllocationResponse API 响应格式
type AllocationResponse struct {
	ID               int                `json:"id"`
	UUID             string             `json:"uuid"`
	TestbedUUID      string             `json:"testbed_uuid"`
	TestbedName      *string            `json:"testbed_name,omitempty"`
	CategoryUUID     string             `json:"category_uuid"`
	CategoryName     *string            `json:"category_name,omitempty"`
	Requester        string             `json:"requester"`
	RequesterComment *string            `json:"requester_comment,omitempty"`
	Status           AllocationStatus   `json:"status"`
	ExpiresAt        *string            `json:"expires_at,omitempty"`
	ReleasedAt       *string            `json:"released_at,omitempty"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
	RemainingSeconds *int64             `json:"remaining_seconds,omitempty"`
}

// IsActive 检查分配是否活跃
func (a *Allocation) IsActive() bool {
	return a.Status == AllocationStatusActive
}

// IsExpired 检查分配是否过期
func (a *Allocation) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// IsReleased 检查分配是否已释放
func (a *Allocation) IsReleased() bool {
	return a.Status == AllocationStatusReleased || a.Status == AllocationStatusExpired
}

// MarkActive 标记为活跃状态
func (a *Allocation) MarkActive(expiresAt *time.Time) {
	now := time.Now()
	a.Status = AllocationStatusActive
	a.ExpiresAt = expiresAt
	a.UpdatedAt = now
}

// MarkReleased 标记为已释放状态
func (a *Allocation) MarkReleased() {
	now := time.Now()
	a.Status = AllocationStatusReleased
	a.ReleasedAt = &now
	a.UpdatedAt = now
}

// MarkExpired 标记为过期状态
func (a *Allocation) MarkExpired() {
	now := time.Now()
	a.Status = AllocationStatusExpired
	a.ReleasedAt = &now
	a.UpdatedAt = now
}

// GetRemainingSeconds 获取剩余秒数
func (a *Allocation) GetRemainingSeconds() int64 {
	if a.ExpiresAt == nil {
		return 0
	}
	remaining := time.Until(*a.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return int64(remaining.Seconds())
}

// ToResponse 转换为 API 响应格式
func (a *Allocation) ToResponse() AllocationResponse {
	resp := AllocationResponse{
		ID:               a.ID,
		UUID:             a.UUID,
		TestbedUUID:      a.TestbedUUID,
		CategoryUUID:     a.CategoryUUID,
		Requester:        a.Requester,
		RequesterComment: a.RequesterComment,
		Status:           a.Status,
		CreatedAt:        a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        a.UpdatedAt.Format(time.RFC3339),
	}

	if a.ExpiresAt != nil {
		expiresAt := a.ExpiresAt.Format(time.RFC3339)
		resp.ExpiresAt = &expiresAt
	}

	if a.ReleasedAt != nil {
		releasedAt := a.ReleasedAt.Format(time.RFC3339)
		resp.ReleasedAt = &releasedAt
	}

	if a.IsActive() && a.ExpiresAt != nil {
		remaining := a.GetRemainingSeconds()
		resp.RemainingSeconds = &remaining
	}

	return resp
}

// NewAllocation 创建新的分配记录
func NewAllocation(testbedUUID, categoryUUID, requester string, maxLifetimeSeconds int) *Allocation {
	now := time.Now()
	var expiresAt *time.Time
	if maxLifetimeSeconds > 0 {
		exp := now.Add(time.Duration(maxLifetimeSeconds) * time.Second)
		expiresAt = &exp
	}

	return &Allocation{
		UUID:         uuid.New().String(),
		TestbedUUID:  testbedUUID,
		CategoryUUID: categoryUUID,
		Requester:    requester,
		Status:       AllocationStatusPending,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ParseAllocationStatus 解析分配状态字符串
func ParseAllocationStatus(s string) (AllocationStatus, error) {
	switch s {
	case string(AllocationStatusPending):
		return AllocationStatusPending, nil
	case string(AllocationStatusActive):
		return AllocationStatusActive, nil
	case string(AllocationStatusReleased):
		return AllocationStatusReleased, nil
	case string(AllocationStatusExpired):
		return AllocationStatusExpired, nil
	default:
		return "", fmt.Errorf("invalid allocation status: %s", s)
	}
}

// Value 实现 driver.Valuer 接口
func (as AllocationStatus) Value() (driver.Value, error) {
	return string(as), nil
}

// Scan 实现 sql.Scanner 接口
func (as *AllocationStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*as = AllocationStatus(v)
	case string:
		*as = AllocationStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into AllocationStatus", value)
	}
	return nil
}
