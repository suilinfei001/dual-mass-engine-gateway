// Package service provides unit tests for resource manager service.
package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/quality-gateway/resource-manager/internal/storage"
	"github.com/quality-gateway/shared/pkg/logger"
)

// mockResourceStorage 是 ResourceStorageInterface 的 mock 实现
type mockResourceStorage struct {
	resources     map[string]*storage.ResourceInstance
	forceError    bool
	forceErrorMsg string
	callCount     map[string]int
}

func newMockResourceStorage() *mockResourceStorage {
	return &mockResourceStorage{
		resources: make(map[string]*storage.ResourceInstance),
		callCount: make(map[string]int),
	}
}

func (m *mockResourceStorage) incrementCall(name string) {
	m.callCount[name]++
}

func (m *mockResourceStorage) getCallCount(name string) int {
	return m.callCount[name]
}

func (m *mockResourceStorage) List(ctx context.Context) ([]*storage.ResourceInstance, error) {
	m.incrementCall("List")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	var result []*storage.ResourceInstance
	for _, r := range m.resources {
		result = append(result, r)
	}
	return result, nil
}

func (m *mockResourceStorage) GetByUUID(ctx context.Context, uuid string) (*storage.ResourceInstance, error) {
	m.incrementCall("GetByUUID")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	return m.resources[uuid], nil
}

func (m *mockResourceStorage) ListByCategory(ctx context.Context, categoryID int64) ([]*storage.ResourceInstance, error) {
	m.incrementCall("ListByCategory")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	var result []*storage.ResourceInstance
	for _, r := range m.resources {
		if r.CategoryID == categoryID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockResourceStorage) ListAvailable(ctx context.Context, categoryID int64) ([]*storage.ResourceInstance, error) {
	m.incrementCall("ListAvailable")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	var result []*storage.ResourceInstance
	for _, r := range m.resources {
		if r.CategoryID == categoryID && r.Status == storage.ResourceStatusActive {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockResourceStorage) Create(ctx context.Context, r *storage.ResourceInstance) error {
	m.incrementCall("Create")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	r.ID = int64(len(m.resources) + 1)
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	m.resources[r.UUID] = r
	return nil
}

func (m *mockResourceStorage) Update(ctx context.Context, r *storage.ResourceInstance) error {
	m.incrementCall("Update")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	if _, exists := m.resources[r.UUID]; !exists {
		return errors.New("resource not found")
	}
	r.UpdatedAt = time.Now()
	m.resources[r.UUID] = r
	return nil
}

func (m *mockResourceStorage) Delete(ctx context.Context, uuid string) error {
	m.incrementCall("Delete")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	delete(m.resources, uuid)
	return nil
}

// mockCategoryStorage 是 CategoryStorageInterface 的 mock 实现
type mockCategoryStorage struct {
	categories    map[string]*storage.Category
	categoriesID  map[int64]*storage.Category
	forceError    bool
	forceErrorMsg string
	callCount     map[string]int
}

func newMockCategoryStorage() *mockCategoryStorage {
	return &mockCategoryStorage{
		categories:   make(map[string]*storage.Category),
		categoriesID: make(map[int64]*storage.Category),
		callCount:    make(map[string]int),
	}
}

func (m *mockCategoryStorage) incrementCall(name string) {
	m.callCount[name]++
}

func (m *mockCategoryStorage) getCallCount(name string) int {
	return m.callCount[name]
}

func (m *mockCategoryStorage) List(ctx context.Context) ([]*storage.Category, error) {
	m.incrementCall("List")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	var result []*storage.Category
	for _, c := range m.categories {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockCategoryStorage) GetByName(ctx context.Context, name string) (*storage.Category, error) {
	m.incrementCall("GetByName")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	return m.categories[name], nil
}

func (m *mockCategoryStorage) GetByID(ctx context.Context, id int64) (*storage.Category, error) {
	m.incrementCall("GetByID")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	return m.categoriesID[id], nil
}

func (m *mockCategoryStorage) Create(ctx context.Context, c *storage.Category) error {
	m.incrementCall("Create")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	c.ID = int64(len(m.categories) + 1)
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	m.categories[c.Name] = c
	m.categoriesID[c.ID] = c
	return nil
}

func (m *mockCategoryStorage) Update(ctx context.Context, c *storage.Category) error {
	m.incrementCall("Update")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	if _, exists := m.categories[c.Name]; !exists {
		return errors.New("category not found")
	}
	c.UpdatedAt = time.Now()
	m.categories[c.Name] = c
	return nil
}

func (m *mockCategoryStorage) Delete(ctx context.Context, id int64) error {
	m.incrementCall("Delete")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	if c, exists := m.categoriesID[id]; exists {
		delete(m.categories, c.Name)
		delete(m.categoriesID, id)
	}
	return nil
}

// mockQuotaPolicyStorage 是 QuotaPolicyStorageInterface 的 mock 实现
type mockQuotaPolicyStorage struct {
	quotas        map[int64][]*storage.QuotaPolicy
	forceError    bool
	forceErrorMsg string
	callCount     map[string]int
}

func newMockQuotaPolicyStorage() *mockQuotaPolicyStorage {
	return &mockQuotaPolicyStorage{
		quotas:    make(map[int64][]*storage.QuotaPolicy),
		callCount: make(map[string]int),
	}
}

func (m *mockQuotaPolicyStorage) incrementCall(name string) {
	m.callCount[name]++
}

func (m *mockQuotaPolicyStorage) getCallCount(name string) int {
	return m.callCount[name]
}

func (m *mockQuotaPolicyStorage) GetByCategoryID(ctx context.Context, categoryID int64) ([]*storage.QuotaPolicy, error) {
	m.incrementCall("GetByCategoryID")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	return m.quotas[categoryID], nil
}

// mockAllocationStorage 是 AllocationStorageInterface 的 mock 实现
type mockAllocationStorage struct {
	allocations   map[string]*storage.Allocation
	forceError    bool
	forceErrorMsg string
	callCount     map[string]int
}

func newMockAllocationStorage() *mockAllocationStorage {
	return &mockAllocationStorage{
		allocations: make(map[string]*storage.Allocation),
		callCount:   make(map[string]int),
	}
}

func (m *mockAllocationStorage) incrementCall(name string) {
	m.callCount[name]++
}

func (m *mockAllocationStorage) getCallCount(name string) int {
	return m.callCount[name]
}

func (m *mockAllocationStorage) Create(ctx context.Context, alloc *storage.Allocation) error {
	m.incrementCall("Create")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	alloc.ID = int64(len(m.allocations) + 1)
	alloc.AllocatedAt = time.Now()
	m.allocations[alloc.ResourceUUID] = alloc
	return nil
}

func (m *mockAllocationStorage) Release(ctx context.Context, resourceUUID string) error {
	m.incrementCall("Release")
	if m.forceError {
		return errors.New(m.forceErrorMsg)
	}
	if alloc, exists := m.allocations[resourceUUID]; exists {
		now := time.Now()
		alloc.ReleasedAt = &now
		alloc.Status = "released"
	}
	return nil
}

func (m *mockAllocationStorage) GetActiveByResourceUUID(ctx context.Context, resourceUUID string) (*storage.Allocation, error) {
	m.incrementCall("GetActiveByResourceUUID")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	if alloc, exists := m.allocations[resourceUUID]; exists && alloc.Status == "active" {
		return alloc, nil
	}
	return nil, nil
}

// mockTestbedStorage 是 TestbedStorageInterface 的 mock 实现
type mockTestbedStorage struct {
	testbeds      []*storage.Testbed
	forceError    bool
	forceErrorMsg string
	callCount     map[string]int
}

func newMockTestbedStorage() *mockTestbedStorage {
	return &mockTestbedStorage{
		testbeds:  make([]*storage.Testbed, 0),
		callCount: make(map[string]int),
	}
}

func (m *mockTestbedStorage) incrementCall(name string) {
	m.callCount[name]++
}

func (m *mockTestbedStorage) getCallCount(name string) int {
	return m.callCount[name]
}

func (m *mockTestbedStorage) List(ctx context.Context) ([]*storage.Testbed, error) {
	m.incrementCall("List")
	if m.forceError {
		return nil, errors.New(m.forceErrorMsg)
	}
	return m.testbeds, nil
}

// testHelper 创建测试用的辅助对象
func testHelper(t *testing.T) (*ResourceManagerService, *mockResourceStorage, *mockCategoryStorage, *mockQuotaPolicyStorage, *mockAllocationStorage, *mockTestbedStorage) {
	t.Helper()

	log := logger.New(logger.Config{
		Level:       logger.InfoLevel,
		DisableTime: true,
	})

	mockResource := newMockResourceStorage()
	mockCategory := newMockCategoryStorage()
	mockQuota := newMockQuotaPolicyStorage()
	mockAllocation := newMockAllocationStorage()
	mockTestbed := newMockTestbedStorage()

	service := &ResourceManagerService{
		resourceStorage:     mockResource,
		categoryStorage:     mockCategory,
		quotaPolicyStorage:  mockQuota,
		allocationStorage:   mockAllocation,
		testbedStorage:      mockTestbed,
		logger:              log,
	}

	return service, mockResource, mockCategory, mockQuota, mockAllocation, mockTestbed
}

// TestResourceManagerService_ListResources 测试列出资源
func TestResourceManagerService_ListResources(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		resources, err := service.ListResources(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(resources) != 0 {
			t.Errorf("expected empty list, got %d items", len(resources))
		}
		if rs.getCallCount("List") != 1 {
			t.Errorf("expected 1 call to List, got %d", rs.getCallCount("List"))
		}
	})

	t.Run("with resources", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		// 添加测试资源
		resource := &storage.ResourceInstance{
			UUID:        "test-uuid-1",
			Name:        "Test Resource",
			Description: "Test Description",
			IPAddress:   "192.168.1.1",
			SSHPort:     22,
			SSHUser:     "root",
			CategoryID:  1,
			Status:      storage.ResourceStatusActive,
		}
		rs.resources[resource.UUID] = resource

		resources, err := service.ListResources(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(resources) != 1 {
			t.Errorf("expected 1 resource, got %d", len(resources))
		}
		if resources[0].Name != "Test Resource" {
			t.Errorf("expected name 'Test Resource', got '%s'", resources[0].Name)
		}
	})

	t.Run("storage error", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)
		rs.forceError = true
		rs.forceErrorMsg = "database connection failed"

		_, err := service.ListResources(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "database connection failed" {
			t.Errorf("expected specific error, got: %v", err)
		}
	})
}

// TestResourceManagerService_GetResource 测试获取单个资源
func TestResourceManagerService_GetResource(t *testing.T) {
	t.Run("existing resource", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		resource := &storage.ResourceInstance{
			UUID:        "test-uuid-2",
			Name:        "Test Resource 2",
			Description: "Test Description 2",
			IPAddress:   "192.168.1.2",
			CategoryID:  1,
			Status:      storage.ResourceStatusActive,
		}
		rs.resources[resource.UUID] = resource

		result, err := service.GetResource(context.Background(), "test-uuid-2")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected resource, got nil")
		}
		if result.Name != "Test Resource 2" {
			t.Errorf("expected name 'Test Resource 2', got '%s'", result.Name)
		}
		if rs.getCallCount("GetByUUID") != 1 {
			t.Errorf("expected 1 call to GetByUUID, got %d", rs.getCallCount("GetByUUID"))
		}
	})

	t.Run("non-existing resource", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		result, err := service.GetResource(context.Background(), "non-existing")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

// TestResourceManagerService_ListResourcesByCategory 测试按类别列出资源
func TestResourceManagerService_ListResourcesByCategory(t *testing.T) {
	t.Run("resources in category", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		// 添加不同类别的资源
		resource1 := &storage.ResourceInstance{
			UUID:       "test-uuid-1",
			Name:       "Resource 1",
			CategoryID: 1,
			Status:     storage.ResourceStatusActive,
		}
		resource2 := &storage.ResourceInstance{
			UUID:       "test-uuid-2",
			Name:       "Resource 2",
			CategoryID: 2,
			Status:     storage.ResourceStatusActive,
		}
		resource3 := &storage.ResourceInstance{
			UUID:       "test-uuid-3",
			Name:       "Resource 3",
			CategoryID: 1,
			Status:     storage.ResourceStatusActive,
		}
		rs.resources[resource1.UUID] = resource1
		rs.resources[resource2.UUID] = resource2
		rs.resources[resource3.UUID] = resource3

		resources, err := service.ListResourcesByCategory(context.Background(), 1)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(resources) != 2 {
			t.Errorf("expected 2 resources, got %d", len(resources))
		}
	})
}

// TestResourceManagerService_ListCategories 测试列出类别
func TestResourceManagerService_ListCategories(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		categories, err := service.ListCategories(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(categories) != 0 {
			t.Errorf("expected empty list, got %d items", len(categories))
		}
		if cs.getCallCount("List") != 1 {
			t.Errorf("expected 1 call to List, got %d", cs.getCallCount("List"))
		}
	})

	t.Run("with categories", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category1 := &storage.Category{
			Name:        "category-1",
			Description: "Category 1",
		}
		category2 := &storage.Category{
			Name:        "category-2",
			Description: "Category 2",
		}
		cs.categories[category1.Name] = category1
		cs.categories[category2.Name] = category2

		categories, err := service.ListCategories(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(categories) != 2 {
			t.Errorf("expected 2 categories, got %d", len(categories))
		}
	})
}

// TestResourceManagerService_GetCategory 测试获取类别
func TestResourceManagerService_GetCategory(t *testing.T) {
	t.Run("existing category", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category := &storage.Category{
			Name:        "test-category",
			Description: "Test category description",
		}
		cs.categories[category.Name] = category

		result, err := service.GetCategory(context.Background(), "test-category")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected category, got nil")
		}
		if result.Name != "test-category" {
			t.Errorf("expected name 'test-category', got '%s'", result.Name)
		}
	})

	t.Run("non-existing category", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		result, err := service.GetCategory(context.Background(), "non-existing")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})
}

// TestResourceManagerService_GetCategoryByID 测试通过 ID 获取类别
func TestResourceManagerService_GetCategoryByID(t *testing.T) {
	t.Run("existing category by ID", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category := &storage.Category{
			ID:          123,
			Name:        "test-category-id",
			Description: "Test category by ID",
		}
		cs.categoriesID[category.ID] = category

		result, err := service.GetCategoryByID(context.Background(), 123)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected category, got nil")
		}
		if result.ID != 123 {
			t.Errorf("expected ID 123, got %d", result.ID)
		}
	})
}

// TestResourceManagerService_CreateCategory 测试创建类别
func TestResourceManagerService_CreateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category := &storage.Category{
			Name:        "new-category",
			Description: "New category description",
		}

		err := service.CreateCategory(context.Background(), category)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if category.ID == 0 {
			t.Error("expected ID to be set")
		}
		if cs.getCallCount("Create") != 1 {
			t.Errorf("expected 1 call to Create, got %d", cs.getCallCount("Create"))
		}
	})

	t.Run("storage error", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)
		cs.forceError = true

		category := &storage.Category{Name: "error-category"}
		err := service.CreateCategory(context.Background(), category)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestResourceManagerService_UpdateCategory 测试更新类别
func TestResourceManagerService_UpdateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category := &storage.Category{
			Name:        "update-category",
			Description: "Original description",
		}
		cs.categories[category.Name] = category

		category.Description = "Updated description"
		err := service.UpdateCategory(context.Background(), category)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if cs.getCallCount("Update") != 1 {
			t.Errorf("expected 1 call to Update, got %d", cs.getCallCount("Update"))
		}
	})

	t.Run("non-existing category", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		category := &storage.Category{
			Name:        "non-existing",
			Description: "Does not exist",
		}

		err := service.UpdateCategory(context.Background(), category)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestResourceManagerService_DeleteCategory 测试删除类别
func TestResourceManagerService_DeleteCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, _, cs, _, _, _ := testHelper(t)

		category := &storage.Category{
			ID:   999,
			Name: "delete-category",
		}
		cs.categories[category.Name] = category
		cs.categoriesID[category.ID] = category

		err := service.DeleteCategory(context.Background(), 999)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if _, exists := cs.categories[category.Name]; exists {
			t.Error("category should be deleted")
		}
	})
}

// TestResourceManagerService_CreateResource 测试创建资源
func TestResourceManagerService_CreateResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		resource := &storage.ResourceInstance{
			UUID:        "new-resource-uuid",
			Name:        "New Resource",
			Description: "New resource description",
			IPAddress:   "192.168.1.100",
			SSHPort:     22,
			SSHUser:     "admin",
			CategoryID:  1,
			Status:      storage.ResourceStatusPending,
		}

		err := service.CreateResource(context.Background(), resource)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if resource.ID == 0 {
			t.Error("expected ID to be set")
		}
		if rs.getCallCount("Create") != 1 {
			t.Errorf("expected 1 call to Create, got %d", rs.getCallCount("Create"))
		}
	})
}

// TestResourceManagerService_UpdateResource 测试更新资源
func TestResourceManagerService_UpdateResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		resource := &storage.ResourceInstance{
			UUID:        "update-resource-uuid",
			Name:        "Original Name",
			Description: "Original description",
			IPAddress:   "192.168.1.50",
			CategoryID:  1,
			Status:      storage.ResourceStatusActive,
		}
		rs.resources[resource.UUID] = resource

		resource.Name = "Updated Name"
		resource.Status = storage.ResourceStatusUnreachable

		err := service.UpdateResource(context.Background(), resource)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if rs.getCallCount("Update") != 1 {
			t.Errorf("expected 1 call to Update, got %d", rs.getCallCount("Update"))
		}
	})
}

// TestResourceManagerService_DeleteResource 测试删除资源
func TestResourceManagerService_DeleteResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		resource := &storage.ResourceInstance{
			UUID:       "delete-resource-uuid",
			Name:       "Delete Resource",
			CategoryID: 1,
			Status:     storage.ResourceStatusActive,
		}
		rs.resources[resource.UUID] = resource

		err := service.DeleteResource(context.Background(), "delete-resource-uuid")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if _, exists := rs.resources[resource.UUID]; exists {
			t.Error("resource should be deleted")
		}
		if rs.getCallCount("Delete") != 1 {
			t.Errorf("expected 1 call to Delete, got %d", rs.getCallCount("Delete"))
		}
	})
}

// TestResourceManagerService_MatchResource 测试资源匹配
func TestResourceManagerService_MatchResource(t *testing.T) {
	t.Run("no available resources", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		req := &MatchResourceRequest{
			CategoryID:    1,
			TaskUUID:      "task-123",
			RequiredCount: 1,
		}

		result, err := service.MatchResource(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result.Matched {
			t.Error("expected not matched")
		}
		if result.Reason != "no available resources in this category" {
			t.Errorf("unexpected reason: %s", result.Reason)
		}
	})

	t.Run("not enough resources", func(t *testing.T) {
		service, rs, _, _, _, _ := testHelper(t)

		// 添加 2 个可用资源
		for i := 1; i <= 2; i++ {
			resource := &storage.ResourceInstance{
				UUID:       "resource-" + string(rune('0'+i)),
				Name:       "Resource " + string(rune('0'+i)),
				CategoryID: 1,
				Status:     storage.ResourceStatusActive,
			}
			rs.resources[resource.UUID] = resource
		}

		req := &MatchResourceRequest{
			CategoryID:    1,
			TaskUUID:      "task-123",
			RequiredCount: 5,
		}

		result, err := service.MatchResource(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result.Matched {
			t.Error("expected not matched")
		}
		if result.Reason != "not enough resources: available=2, required=5" {
			t.Errorf("unexpected reason: %s", result.Reason)
		}
	})

	t.Run("successful match", func(t *testing.T) {
		service, rs, _, _, as, _ := testHelper(t)

		// 添加 5 个可用资源
		for i := 1; i <= 5; i++ {
			resource := &storage.ResourceInstance{
				UUID:       "resource-matched-" + string(rune('0'+i)),
				Name:       "Matched Resource " + string(rune('0'+i)),
				CategoryID: 1,
				Status:     storage.ResourceStatusActive,
			}
			rs.resources[resource.UUID] = resource
		}

		req := &MatchResourceRequest{
			CategoryID:    1,
			TaskUUID:      "task-456",
			RequiredCount: 3,
		}

		result, err := service.MatchResource(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if !result.Matched {
			t.Errorf("expected matched, reason: %s", result.Reason)
		}
		if len(result.Resources) != 3 {
			t.Errorf("expected 3 resources, got %d", len(result.Resources))
		}
		if result.Allocation == nil {
			t.Error("expected allocation to be created")
		}
		if result.Allocation.TaskUUID != "task-456" {
			t.Errorf("expected task UUID 'task-456', got '%s'", result.Allocation.TaskUUID)
		}
		if result.Allocation.Status != "active" {
			t.Errorf("expected allocation status 'active', got '%s'", result.Allocation.Status)
		}
		if rs.getCallCount("ListAvailable") != 1 {
			t.Errorf("expected 1 call to ListAvailable, got %d", rs.getCallCount("ListAvailable"))
		}
		if as.getCallCount("Create") != 1 {
			t.Errorf("expected 1 call to Create allocation, got %d", as.getCallCount("Create"))
		}
	})

	t.Run("allocation creation failure", func(t *testing.T) {
		service, rs, _, _, as, _ := testHelper(t)

		resource := &storage.ResourceInstance{
			UUID:       "alloc-fail-resource",
			Name:       "Allocation Fail Resource",
			CategoryID: 1,
			Status:     storage.ResourceStatusActive,
		}
		rs.resources[resource.UUID] = resource

		// 设置分配创建时的错误
		as.forceError = true

		req := &MatchResourceRequest{
			CategoryID:    1,
			TaskUUID:      "task-fail",
			RequiredCount: 1,
		}

		_, err := service.MatchResource(context.Background(), req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// TestResourceManagerService_ReleaseResource 测试释放资源
func TestResourceManagerService_ReleaseResource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, _, _, _, as, _ := testHelper(t)

		resourceUUID := "release-resource-uuid"
		allocation := &storage.Allocation{
			ResourceUUID: resourceUUID,
			TaskUUID:     "task-release",
			Status:       "active",
		}
		as.allocations[resourceUUID] = allocation

		err := service.ReleaseResource(context.Background(), resourceUUID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if as.getCallCount("Release") != 1 {
			t.Errorf("expected 1 call to Release, got %d", as.getCallCount("Release"))
		}
		if allocation.Status != "released" {
			t.Errorf("expected status 'released', got '%s'", allocation.Status)
		}
		if allocation.ReleasedAt == nil {
			t.Error("expected ReleasedAt to be set")
		}
	})
}

// TestResourceManagerService_GetActiveAllocation 测试获取活跃分配
func TestResourceManagerService_GetActiveAllocation(t *testing.T) {
	t.Run("existing active allocation", func(t *testing.T) {
		service, _, _, _, as, _ := testHelper(t)

		resourceUUID := "active-allocation-uuid"
		allocation := &storage.Allocation{
			ResourceUUID: resourceUUID,
			TaskUUID:     "task-active",
			Status:       "active",
		}
		as.allocations[resourceUUID] = allocation

		result, err := service.GetActiveAllocation(context.Background(), resourceUUID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected allocation, got nil")
		}
		if result.Status != "active" {
			t.Errorf("expected status 'active', got '%s'", result.Status)
		}
		if as.getCallCount("GetActiveByResourceUUID") != 1 {
			t.Errorf("expected 1 call to GetActiveByResourceUUID, got %d", as.getCallCount("GetActiveByResourceUUID"))
		}
	})

	t.Run("no active allocation", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		resourceUUID := "no-allocation-uuid"

		result, err := service.GetActiveAllocation(context.Background(), resourceUUID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %+v", result)
		}
	})

	t.Run("released allocation", func(t *testing.T) {
		service, _, _, _, as, _ := testHelper(t)

		resourceUUID := "released-allocation-uuid"
		now := time.Now()
		allocation := &storage.Allocation{
			ResourceUUID: resourceUUID,
			TaskUUID:     "task-released",
			Status:       "released",
			ReleasedAt:   &now,
		}
		as.allocations[resourceUUID] = allocation

		result, err := service.GetActiveAllocation(context.Background(), resourceUUID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil for released allocation, got %+v", result)
		}
	})
}

// TestResourceManagerService_ListQuotaPolicies 测试列出配额策略
func TestResourceManagerService_ListQuotaPolicies(t *testing.T) {
	t.Run("with policies", func(t *testing.T) {
		service, _, _, qs, _, _ := testHelper(t)

		categoryID := int64(1)
		policies := []*storage.QuotaPolicy{
			{
				ID:            1,
				Name:          "policy-1",
				CategoryID:    categoryID,
				MaxCount:      10,
				ReplenishRate: 1,
				ReplenishUnit: "hour",
			},
			{
				ID:            2,
				Name:          "policy-2",
				CategoryID:    categoryID,
				MaxCount:      20,
				ReplenishRate: 2,
				ReplenishUnit: "day",
			},
		}
		qs.quotas[categoryID] = policies

		result, err := service.ListQuotaPolicies(context.Background(), categoryID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 policies, got %d", len(result))
		}
		if qs.getCallCount("GetByCategoryID") != 1 {
			t.Errorf("expected 1 call to GetByCategoryID, got %d", qs.getCallCount("GetByCategoryID"))
		}
	})

	t.Run("empty policies", func(t *testing.T) {
		service, _, _, _, _, _ := testHelper(t)

		result, err := service.ListQuotaPolicies(context.Background(), 999)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty list, got %d items", len(result))
		}
	})
}

// TestResourceManagerService_ListTestbeds 测试列出测试床
func TestResourceManagerService_ListTestbeds(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		service, _, _, _, _, ts := testHelper(t)

		testbeds, err := service.ListTestbeds(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(testbeds) != 0 {
			t.Errorf("expected empty list, got %d items", len(testbeds))
		}
		if ts.getCallCount("List") != 1 {
			t.Errorf("expected 1 call to List, got %d", ts.getCallCount("List"))
		}
	})

	t.Run("with testbeds", func(t *testing.T) {
		service, _, _, _, _, ts := testHelper(t)

		testbeds := []*storage.Testbed{
			{
				ID:        1,
				Name:      "testbed-1",
				IPAddress: "192.168.1.1",
				SSHPort:   22,
				Capacity:  10,
			},
			{
				ID:        2,
				Name:      "testbed-2",
				IPAddress: "192.168.1.2",
				SSHPort:   22,
				Capacity:  20,
			},
		}
		ts.testbeds = testbeds

		result, err := service.ListTestbeds(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 testbeds, got %d", len(result))
		}
	})
}

// TestResourceManagerService_CreateAllocation 测试创建分配
func TestResourceManagerService_CreateAllocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service, _, _, _, as, _ := testHelper(t)

		allocation := &storage.Allocation{
			ResourceUUID: "resource-alloc-uuid",
			TaskUUID:     "task-alloc",
			Status:       "active",
		}

		err := service.CreateAllocation(context.Background(), allocation)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if allocation.ID == 0 {
			t.Error("expected ID to be set")
		}
		if as.getCallCount("Create") != 1 {
			t.Errorf("expected 1 call to Create, got %d", as.getCallCount("Create"))
		}
	})
}
