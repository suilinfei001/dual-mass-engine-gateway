package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ServiceTarget 服务对象类型
type ServiceTarget string

const (
	// ServiceTargetRobot robot用户
	ServiceTargetRobot ServiceTarget = "robot"
	// ServiceTargetNormal 普通用户
	ServiceTargetNormal ServiceTarget = "normal"
)

// ValidServiceTargets 所有有效的服务对象
var ValidServiceTargets = map[string]ServiceTarget{
	"robot":  ServiceTargetRobot,
	"normal": ServiceTargetNormal,
}

// ParseServiceTarget 解析服务对象字符串
func ParseServiceTarget(s string) (ServiceTarget, error) {
	target, ok := ValidServiceTargets[s]
	if !ok {
		return "", fmt.Errorf("invalid service target: %s", s)
	}
	return target, nil
}

// DisplayName 返回服务对象的显示名称
func (s ServiceTarget) DisplayName() string {
	switch s {
	case ServiceTargetRobot:
		return "robot"
	case ServiceTargetNormal:
		return "普通用户"
	default:
		return string(s)
	}
}

// QuotaPolicy 配额策略
type QuotaPolicy struct {
	ID                 int           `json:"id"`
	UUID               string        `json:"uuid"`
	CategoryUUID       string        `json:"category_uuid"`
	MinInstances       int           `json:"min_instances"`
	MaxInstances       int           `json:"max_instances"`
	Priority           int           `json:"priority"`
	ServiceTarget      ServiceTarget `json:"service_target"`
	AutoReplenish      bool          `json:"auto_replenish"`
	ReplenishThreshold int           `json:"replenish_threshold"`
	MaxLifetimeSeconds int           `json:"max_lifetime_seconds"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// QuotaPolicyResponse API 响应格式
type QuotaPolicyResponse struct {
	ID                 int           `json:"id"`
	UUID               string        `json:"uuid"`
	CategoryUUID       string        `json:"category_uuid"`
	CategoryName       *string       `json:"category_name,omitempty"`
	MinInstances       int           `json:"min_instances"`
	MaxInstances       int           `json:"max_instances"`
	Priority           int           `json:"priority"`
	ServiceTarget      ServiceTarget `json:"service_target"`
	AutoReplenish      bool          `json:"auto_replenish"`
	ReplenishThreshold int           `json:"replenish_threshold"`
	MaxLifetimeSeconds int           `json:"max_lifetime_seconds"`
	CreatedAt          string        `json:"created_at"`
	UpdatedAt          string        `json:"updated_at"`
}

// ShouldReplenish 检查是否需要补充
func (q *QuotaPolicy) ShouldReplenish(availableCount int) bool {
	if !q.AutoReplenish {
		return false
	}
	return availableCount < q.ReplenishThreshold
}

// IsOverQuota 检查是否超过配额
func (q *QuotaPolicy) IsOverQuota(currentCount int) bool {
	return currentCount >= q.MaxInstances
}

// IsUnderQuota 检查是否低于最小配额
func (q *QuotaPolicy) IsUnderQuota(currentCount int) bool {
	return currentCount < q.MinInstances
}

// CanAllocate 检查是否可以分配
func (q *QuotaPolicy) CanAllocate(allocatedCount int) bool {
	return allocatedCount < q.MaxInstances
}

// ToResponse 转换为 API 响应格式
func (q *QuotaPolicy) ToResponse() QuotaPolicyResponse {
	return QuotaPolicyResponse{
		ID:                 q.ID,
		UUID:               q.UUID,
		CategoryUUID:       q.CategoryUUID,
		MinInstances:       q.MinInstances,
		MaxInstances:       q.MaxInstances,
		Priority:           q.Priority,
		AutoReplenish:      q.AutoReplenish,
		ReplenishThreshold: q.ReplenishThreshold,
		MaxLifetimeSeconds: q.MaxLifetimeSeconds,
		CreatedAt:          q.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          q.UpdatedAt.Format(time.RFC3339),
	}
}

// NewQuotaPolicy 创建新的配额策略
func NewQuotaPolicy(categoryUUID string, minInstances, maxInstances, priority, maxLifetimeSeconds int) *QuotaPolicy {
	now := time.Now()
	// 默认补充阈值为 minInstances 的 80%
	replenishThreshold := minInstances * 80 / 100
	if replenishThreshold < 1 {
		replenishThreshold = 1
	}

	return &QuotaPolicy{
		UUID:               uuid.New().String(),
		CategoryUUID:       categoryUUID,
		MinInstances:       minInstances,
		MaxInstances:       maxInstances,
		Priority:           priority,
		ServiceTarget:      ServiceTargetNormal, // 默认为普通用户
		AutoReplenish:      true,
		ReplenishThreshold: replenishThreshold,
		MaxLifetimeSeconds: maxLifetimeSeconds,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
