package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InstanceType 实例类型
type InstanceType string

const (
	InstanceTypeVirtualMachine InstanceType = "VirtualMachine"
	InstanceTypeMachine        InstanceType = "Machine"
)

// ResourceInstanceStatus 资源实例状态
type ResourceInstanceStatus string

const (
	ResourceInstanceStatusPending    ResourceInstanceStatus = "pending"
	ResourceInstanceStatusActive     ResourceInstanceStatus = "active"
	ResourceInstanceStatusUnreachable ResourceInstanceStatus = "unreachable"
)

// ResourceInstance 资源实例 - 裸的虚拟机或实体机
type ResourceInstance struct {
	ID                   int                        `json:"id"`
	UUID                 string                     `json:"uuid"`
	InstanceType         InstanceType               `json:"instance_type"`
	SnapshotID           *string                    `json:"snapshot_id,omitempty"`
	SnapshotInstanceUUID *string                    `json:"snapshot_instance_uuid,omitempty"`
	IPAddress            string                     `json:"ip_address"`
	Port                 int                        `json:"port"`
	SSHUser              string                     `json:"ssh_user"`
	Passwd               string                     `json:"passwd"`
	Description          *string                    `json:"description,omitempty"`
	IsPublic             bool                       `json:"is_public"`
	CreatedBy            string                     `json:"created_by"`
	Status               ResourceInstanceStatus     `json:"status"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
	TerminatedAt         *time.Time                 `json:"terminated_at,omitempty"`
}

// ResourceInstanceResponse API 响应格式
type ResourceInstanceResponse struct {
	ID                   int                        `json:"id"`
	UUID                 string                     `json:"uuid"`
	InstanceType         InstanceType               `json:"instance_type"`
	SnapshotID           *string                    `json:"snapshot_id,omitempty"`
	SnapshotInstanceUUID *string                    `json:"snapshot_instance_uuid,omitempty"`
	IPAddress            string                     `json:"ip_address"`
	Port                 int                        `json:"port"`
	SSHUser              string                     `json:"ssh_user"`
	Passwd               string                     `json:"passwd"`
	Description          *string                    `json:"description,omitempty"`
	IsPublic             bool                       `json:"is_public"`
	CreatedBy            string                     `json:"created_by"`
	Status               ResourceInstanceStatus     `json:"status"`
	CreatedAt            string                     `json:"created_at"`
	UpdatedAt            string                     `json:"updated_at"`
	TerminatedAt         *string                    `json:"terminated_at,omitempty"`
}

// IsActive 检查资源实例是否活跃
func (r *ResourceInstance) IsActive() bool {
	return r.Status == ResourceInstanceStatusActive
}

// IsVirtualMachine 检查是否为虚拟机
func (r *ResourceInstance) IsVirtualMachine() bool {
	return r.InstanceType == InstanceTypeVirtualMachine
}

// IsMachine 检查是否为实体机
func (r *ResourceInstance) IsMachine() bool {
	return r.InstanceType == InstanceTypeMachine
}

// CanParticipateInPool 检查是否可以参与资源池
func (r *ResourceInstance) CanParticipateInPool() bool {
	return r.IsActive() && r.IsVirtualMachine()
}

// MarkUnreachable 标记为不可达
func (r *ResourceInstance) MarkUnreachable() {
	now := time.Now()
	r.Status = ResourceInstanceStatusUnreachable
	r.UpdatedAt = now
}

// MarkActive 标记为可用
func (r *ResourceInstance) MarkActive() {
	now := time.Now()
	r.Status = ResourceInstanceStatusActive
	r.UpdatedAt = now
}

// ToResponse 转换为 API 响应格式
func (r *ResourceInstance) ToResponse() ResourceInstanceResponse {
	resp := ResourceInstanceResponse{
		ID:                   r.ID,
		UUID:                 r.UUID,
		InstanceType:         r.InstanceType,
		SnapshotID:           r.SnapshotID,
		SnapshotInstanceUUID: r.SnapshotInstanceUUID,
		IPAddress:            r.IPAddress,
		Port:                 r.Port,
		SSHUser:              r.SSHUser,
		Passwd:               r.Passwd,
		Description:          r.Description,
		IsPublic:             r.IsPublic,
		CreatedBy:            r.CreatedBy,
		Status:               r.Status,
		CreatedAt:            r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            r.UpdatedAt.Format(time.RFC3339),
	}
	if r.TerminatedAt != nil {
		terminatedAt := r.TerminatedAt.Format(time.RFC3339)
		resp.TerminatedAt = &terminatedAt
	}
	return resp
}

// ToResponseWithMaskedPassword 转换为 API 响应格式（密码掩码）
func (r *ResourceInstance) ToResponseWithMaskedPassword(showPassword bool) ResourceInstanceResponse {
	resp := r.ToResponse()
	if !showPassword {
		resp.Passwd = "****"
	}
	return resp
}

// NewVirtualMachine 创建新的虚拟机实例
func NewVirtualMachine(ipAddress string, port int, passwd, snapshotID, createdBy string) *ResourceInstance {
	now := time.Now()
	return &ResourceInstance{
		UUID:         uuid.New().String(),
		InstanceType: InstanceTypeVirtualMachine,
		SnapshotID:   &snapshotID,
		IPAddress:    ipAddress,
		Port:         port,
		SSHUser:      "root", // 默认 SSH 用户
		Passwd:       passwd,
		IsPublic:     true, // 虚拟机强制公开
		CreatedBy:    createdBy,
		Status:       ResourceInstanceStatusPending, // 新创建的实例默认为 pending
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewMachine 创建新的实体机实例
func NewMachine(ipAddress string, port int, passwd, createdBy string, isPublic bool) *ResourceInstance {
	now := time.Now()
	return &ResourceInstance{
		UUID:         uuid.New().String(),
		InstanceType: InstanceTypeMachine,
		SnapshotID:   nil, // 实体机没有快照
		IPAddress:    ipAddress,
		Port:         port,
		SSHUser:      "root", // 默认 SSH 用户
		Passwd:       passwd,
		IsPublic:     isPublic, // 实体机由用户选择
		CreatedBy:    createdBy,
		Status:       ResourceInstanceStatusPending, // 新创建的实例默认为 pending
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ValidateVirtualMachine 验证虚拟机实例配置
func (r *ResourceInstance) ValidateVirtualMachine() error {
	if r.InstanceType != InstanceTypeVirtualMachine {
		return fmt.Errorf("instance type is not VirtualMachine")
	}
	if r.SnapshotID == nil || *r.SnapshotID == "" {
		return fmt.Errorf("VirtualMachine must have a SnapshotID")
	}
	if !r.IsPublic {
		return fmt.Errorf("VirtualMachine must be public")
	}
	return nil
}

// ValidateMachine 验证实体机实例配置
func (r *ResourceInstance) ValidateMachine() error {
	if r.InstanceType != InstanceTypeMachine {
		return fmt.Errorf("instance type is not Machine")
	}
	if r.SnapshotID != nil {
		return fmt.Errorf("Machine should not have a SnapshotID")
	}
	return nil
}

// ParseInstanceType 解析实例类型字符串
func ParseInstanceType(s string) (InstanceType, error) {
	switch s {
	case string(InstanceTypeVirtualMachine):
		return InstanceTypeVirtualMachine, nil
	case string(InstanceTypeMachine):
		return InstanceTypeMachine, nil
	default:
		return "", fmt.Errorf("invalid instance type: %s", s)
	}
}

// Value 实现 driver.Valuer 接口
func (it InstanceType) Value() (driver.Value, error) {
	return string(it), nil
}

// Scan 实现 sql.Scanner 接口
func (it *InstanceType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*it = InstanceType(v)
	case string:
		*it = InstanceType(v)
	default:
		return fmt.Errorf("cannot scan %T into InstanceType", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口
func (ris ResourceInstanceStatus) Value() (driver.Value, error) {
	return string(ris), nil
}

// Scan 实现 sql.Scanner 接口
func (ris *ResourceInstanceStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*ris = ResourceInstanceStatus(v)
	case string:
		*ris = ResourceInstanceStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into ResourceInstanceStatus", value)
	}
	return nil
}

// UnmarshalJSON 自定义 JSON 反序列化（支持 terminated_at 为 null）
func (r *ResourceInstance) UnmarshalJSON(data []byte) error {
	type Alias ResourceInstance
	aux := &struct {
		TerminatedAt *string `json:"terminated_at"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.TerminatedAt != nil {
		t, err := time.Parse(time.RFC3339, *aux.TerminatedAt)
		if err == nil {
			r.TerminatedAt = &t
		}
	}
	return nil
}
