// Package storage provides data models for resource manager service.
package storage

import "time"

// ResourceInstanceStatus 资源实例状态
type ResourceInstanceStatus string

const (
	ResourceStatusPending     ResourceInstanceStatus = "pending"
	ResourceStatusActive      ResourceInstanceStatus = "active"
	ResourceStatusUnreachable ResourceInstanceStatus = "unreachable"
)

// ResourceInstance 资源实例
type ResourceInstance struct {
	ID          int64                   `json:"id" db:"id"`
	UUID        string                  `json:"uuid" db:"uuid"`
	Name        string                  `json:"name" db:"name"`
	Description string                  `json:"description" db:"description"`
	IPAddress   string                  `json:"ip_address" db:"ip_address"`
	SSHPort     int                     `json:"ssh_port" db:"ssh_port"`
	SSHUser     string                  `json:"ssh_user" db:"ssh_user"`
	SSHPassword string                  `json:"ssh_password" db:"ssh_password"`
	CategoryID  int64                   `json:"category_id" db:"category_id"`
	IsPublic    bool                    `json:"is_public" db:"is_public"`
	CreatedBy   string                  `json:"created_by" db:"created_by"`
	Status      ResourceInstanceStatus  `json:"status" db:"status"`
	CreatedAt   time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at" db:"updated_at"`
}

// Category 资源类别
type Category struct {
	ID          int64     `json:"id" db:"id"`
	UUID        string    `json:"uuid" db:"uuid"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// QuotaPolicy 配额策略
type QuotaPolicy struct {
	ID              int64     `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	CategoryID      int64     `json:"category_id" db:"category_id"`
	MaxCount        int       `json:"max_count" db:"max_count"`
	ReplenishRate   int       `json:"replenish_rate" db:"replenish_rate"`
	ReplenishUnit   string    `json:"replenish_unit" db:"replenish_unit"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Allocation 资源分配记录
type Allocation struct {
	ID            int64     `json:"id" db:"id"`
	ResourceUUID  string    `json:"resource_uuid" db:"resource_uuid"`
	PolicyUUID    string    `json:"policy_uuid" db:"policy_uuid"`
	TaskUUID      string    `json:"task_uuid" db:"task_uuid"`
	AllocatedAt   time.Time `json:"allocated_at" db:"allocated_at"`
	ReleasedAt    *time.Time `json:"released_at" db:"released_at"`
	Status        string    `json:"status" db:"status"` // active, released
}

// Testbed 测试床
type Testbed struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	SSHPort     int       `json:"ssh_port" db:"ssh_port"`
	SSHUser     string    `json:"ssh_user" db:"ssh_user"`
	SSHPassword string    `json:"ssh_password" db:"ssh_password"`
	Capacity    int       `json:"capacity" db:"capacity"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DeploymentTask 部署任务
type DeploymentTask struct {
	ID              int64     `json:"id" db:"id"`
	InstanceUUID    string    `json:"instance_uuid" db:"instance_uuid"`
	PolicyUUID      string    `json:"policy_uuid" db:"policy_uuid"`
	CategoryUUID    string    `json:"category_uuid" db:"category_uuid"`
	ChartURL        string    `json:"chart_url" db:"chart_url"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
