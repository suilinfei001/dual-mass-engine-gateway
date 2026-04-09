package models

import (
	"time"

	"github.com/google/uuid"
)

// Category 类别 - 业务层的分类概念
type Category struct {
	ID          int       `json:"id"`
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CategoryResponse API 响应格式
type CategoryResponse struct {
	ID            int    `json:"id"`
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
	AvailableCount int   `json:"available_count"`
	AllocatedCount int   `json:"allocated_count"`
	TotalCount     int   `json:"total_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// IsEnabled 检查类别是否启用
func (c *Category) IsEnabled() bool {
	return c.Enabled
}

// Enable 启用类别
func (c *Category) Enable() {
	c.Enabled = true
	c.UpdatedAt = time.Now()
}

// Disable 禁用类别
func (c *Category) Disable() {
	c.Enabled = false
	c.UpdatedAt = time.Now()
}

// ToResponse 转换为 API 响应格式
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		UUID:        c.UUID,
		Name:        c.Name,
		Description: c.Description,
		Enabled:     c.Enabled,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

// NewCategory 创建新的类别
func NewCategory(name, description string) *Category {
	now := time.Now()
	return &Category{
		UUID:        uuid.New().String(),
		Name:        name,
		Description: description,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
