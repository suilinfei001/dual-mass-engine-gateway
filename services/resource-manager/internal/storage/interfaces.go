// Package storage provides interfaces for mocking in tests.
package storage

import "context"

// ResourceStorageInterface defines the interface for resource storage.
type ResourceStorageInterface interface {
	List(ctx context.Context) ([]*ResourceInstance, error)
	GetByUUID(ctx context.Context, uuid string) (*ResourceInstance, error)
	ListByCategory(ctx context.Context, categoryID int64) ([]*ResourceInstance, error)
	ListAvailable(ctx context.Context, categoryID int64) ([]*ResourceInstance, error)
	Create(ctx context.Context, r *ResourceInstance) error
	Update(ctx context.Context, r *ResourceInstance) error
	Delete(ctx context.Context, uuid string) error
}

// CategoryStorageInterface defines the interface for category storage.
type CategoryStorageInterface interface {
	List(ctx context.Context) ([]*Category, error)
	GetByName(ctx context.Context, name string) (*Category, error)
	GetByID(ctx context.Context, id int64) (*Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id int64) error
}

// QuotaPolicyStorageInterface defines the interface for quota policy storage.
type QuotaPolicyStorageInterface interface {
	GetByCategoryID(ctx context.Context, categoryID int64) ([]*QuotaPolicy, error)
}

// AllocationStorageInterface defines the interface for allocation storage.
type AllocationStorageInterface interface {
	Create(ctx context.Context, alloc *Allocation) error
	Release(ctx context.Context, resourceUUID string) error
	GetActiveByResourceUUID(ctx context.Context, resourceUUID string) (*Allocation, error)
}

// TestbedStorageInterface defines the interface for testbed storage.
type TestbedStorageInterface interface {
	List(ctx context.Context) ([]*Testbed, error)
}

// Ensure the concrete types implement the interfaces.
var (
	_ ResourceStorageInterface     = (*ResourceStorage)(nil)
	_ CategoryStorageInterface     = (*CategoryStorage)(nil)
	_ QuotaPolicyStorageInterface  = (*QuotaPolicyStorage)(nil)
	_ AllocationStorageInterface   = (*AllocationStorage)(nil)
	_ TestbedStorageInterface      = (*TestbedStorage)(nil)
)
