// Package service provides business logic for resource manager service.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/quality-gateway/resource-manager/internal/storage"
	"github.com/quality-gateway/shared/pkg/logger"
)

// ResourceManagerService 资源管理服务
type ResourceManagerService struct {
	resourceStorage     storage.ResourceStorageInterface
	categoryStorage     storage.CategoryStorageInterface
	quotaPolicyStorage  storage.QuotaPolicyStorageInterface
	allocationStorage   storage.AllocationStorageInterface
	testbedStorage      storage.TestbedStorageInterface
	logger              *logger.Logger
}

// NewResourceManagerService 创建资源管理服务
func NewResourceManagerService(
	rs storage.ResourceStorageInterface,
	cs storage.CategoryStorageInterface,
	qs storage.QuotaPolicyStorageInterface,
	as storage.AllocationStorageInterface,
	ts storage.TestbedStorageInterface,
	log *logger.Logger,
) *ResourceManagerService {
	return &ResourceManagerService{
		resourceStorage:    rs,
		categoryStorage:    cs,
		quotaPolicyStorage: qs,
		allocationStorage:  as,
		testbedStorage:     ts,
		logger:            log,
	}
}

// ListResources 列出所有资源
func (s *ResourceManagerService) ListResources(ctx context.Context) ([]*storage.ResourceInstance, error) {
	return s.resourceStorage.List(ctx)
}

// GetResource 获取单个资源
func (s *ResourceManagerService) GetResource(ctx context.Context, uuid string) (*storage.ResourceInstance, error) {
	return s.resourceStorage.GetByUUID(ctx, uuid)
}

// ListResourcesByCategory 按类别列出资源
func (s *ResourceManagerService) ListResourcesByCategory(ctx context.Context, categoryID int64) ([]*storage.ResourceInstance, error) {
	return s.resourceStorage.ListByCategory(ctx, categoryID)
}

// ListCategories 列出所有类别
func (s *ResourceManagerService) ListCategories(ctx context.Context) ([]*storage.Category, error) {
	return s.categoryStorage.List(ctx)
}

// GetCategory 获取单个类别
func (s *ResourceManagerService) GetCategory(ctx context.Context, name string) (*storage.Category, error) {
	return s.categoryStorage.GetByName(ctx, name)
}

// GetCategoryByID 根据 ID 获取类别
func (s *ResourceManagerService) GetCategoryByID(ctx context.Context, id int64) (*storage.Category, error) {
	return s.categoryStorage.GetByID(ctx, id)
}

// ListQuotaPolicies 列出配额策略
func (s *ResourceManagerService) ListQuotaPolicies(ctx context.Context, categoryID int64) ([]*storage.QuotaPolicy, error) {
	return s.quotaPolicyStorage.GetByCategoryID(ctx, categoryID)
}

// ListTestbeds 列出所有测试床
func (s *ResourceManagerService) ListTestbeds(ctx context.Context) ([]*storage.Testbed, error) {
	return s.testbedStorage.List(ctx)
}

// CreateAllocation 创建分配记录
func (s *ResourceManagerService) CreateAllocation(ctx context.Context, alloc *storage.Allocation) error {
	return s.allocationStorage.Create(ctx, alloc)
}

// CreateResource 创建资源
func (s *ResourceManagerService) CreateResource(ctx context.Context, r *storage.ResourceInstance) error {
	return s.resourceStorage.Create(ctx, r)
}

// UpdateResource 更新资源
func (s *ResourceManagerService) UpdateResource(ctx context.Context, r *storage.ResourceInstance) error {
	return s.resourceStorage.Update(ctx, r)
}

// DeleteResource 删除资源
func (s *ResourceManagerService) DeleteResource(ctx context.Context, uuid string) error {
	return s.resourceStorage.Delete(ctx, uuid)
}

// CreateCategory 创建类别
func (s *ResourceManagerService) CreateCategory(ctx context.Context, c *storage.Category) error {
	return s.categoryStorage.Create(ctx, c)
}

// UpdateCategory 更新类别
func (s *ResourceManagerService) UpdateCategory(ctx context.Context, c *storage.Category) error {
	return s.categoryStorage.Update(ctx, c)
}

// DeleteCategory 删除类别
func (s *ResourceManagerService) DeleteCategory(ctx context.Context, id int64) error {
	return s.categoryStorage.Delete(ctx, id)
}

// MatchResourceRequest 资源匹配请求
type MatchResourceRequest struct {
	CategoryID    int64  `json:"category_id"`
	TaskUUID      string `json:"task_uuid"`
	RequiredCount int    `json:"required_count"`
}

// MatchResourceResult 资源匹配结果
type MatchResourceResult struct {
	Matched    bool                       `json:"matched"`
	Resources  []*storage.ResourceInstance `json:"resources,omitempty"`
	Reason     string                     `json:"reason,omitempty"`
	Allocation *storage.Allocation        `json:"allocation,omitempty"`
}

// MatchResource 匹配可用资源
func (s *ResourceManagerService) MatchResource(ctx context.Context, req *MatchResourceRequest) (*MatchResourceResult, error) {
	// 获取可用资源
	available, err := s.resourceStorage.ListAvailable(ctx, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list available resources: %w", err)
	}

	if len(available) == 0 {
		return &MatchResourceResult{
			Matched: false,
			Reason:  "no available resources in this category",
		}, nil
	}

	if len(available) < req.RequiredCount {
		return &MatchResourceResult{
			Matched: false,
			Reason:  fmt.Sprintf("not enough resources: available=%d, required=%d", len(available), req.RequiredCount),
		}, nil
	}

	// 选择前 N 个资源
	selected := available[:req.RequiredCount]

	// 创建分配记录
	alloc := &storage.Allocation{
		ResourceUUID: selected[0].UUID, // 主资源
		PolicyUUID:   "",               // TODO: 从配额策略获取
		TaskUUID:     req.TaskUUID,
		AllocatedAt:  time.Now(),
		Status:       "active",
	}

	if err := s.allocationStorage.Create(ctx, alloc); err != nil {
		return nil, fmt.Errorf("failed to create allocation: %w", err)
	}

	return &MatchResourceResult{
		Matched:    true,
		Resources:  selected,
		Allocation: alloc,
	}, nil
}

// ReleaseResource 释放资源
func (s *ResourceManagerService) ReleaseResource(ctx context.Context, resourceUUID string) error {
	return s.allocationStorage.Release(ctx, resourceUUID)
}

// GetActiveAllocation 获取资源的活跃分配
func (s *ResourceManagerService) GetActiveAllocation(ctx context.Context, resourceUUID string) (*storage.Allocation, error) {
	return s.allocationStorage.GetActiveByResourceUUID(ctx, resourceUUID)
}
