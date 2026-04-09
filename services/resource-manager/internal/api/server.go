// Package api provides HTTP handlers for resource manager service.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/quality-gateway/resource-manager/internal/service"
	"github.com/quality-gateway/resource-manager/internal/storage"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
	"github.com/quality-gateway/shared/pkg/logger"
)

// Server 资源管理服务器
type Server struct {
	config    sharedapi.Config
	service   *service.ResourceManagerService
	logger    *logger.Logger
	apiServer *sharedapi.Server
}

// NewServer 创建资源管理服务器
func NewServer(cfg sharedapi.Config, svc *service.ResourceManagerService, log *logger.Logger) *Server {
	srv := sharedapi.New(cfg, log)

	return &Server{
		config:    cfg,
		service:   svc,
		logger:    log,
		apiServer: srv,
	}
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes() {
	// API 路由
	router := s.apiServer.Router()

	// 健康检查
	router.GET("/health", s.healthCheck)

	// === 原有路由（保持兼容） ===
	// 资源相关
	router.GET("/api/resources", s.listResources)
	router.POST("/api/resources", s.createResource)
	router.GET("/api/resources/{uuid}", s.getResource)
	router.PUT("/api/resources/{uuid}", s.updateResource)
	router.DELETE("/api/resources/{uuid}", s.deleteResource)
	router.GET("/api/categories/{id}/resources", s.listResourcesByCategory)

	// 类别相关
	router.GET("/api/categories", s.listCategories)
	router.POST("/api/categories", s.createCategory)
	router.GET("/api/categories/{id}", s.getCategory)
	router.PUT("/api/categories/{id}", s.updateCategory)
	router.DELETE("/api/categories/{id}", s.deleteCategory)

	// 配额策略相关
	router.GET("/api/categories/{id}/quota-policies", s.listQuotaPolicies)

	// 分配相关
	router.POST("/api/resources/match", s.matchResource)
	router.POST("/api/resources/{uuid}/release", s.releaseResource)
	router.GET("/api/resources/{uuid}/allocation", s.getAllocation)

	// 测试床相关
	router.GET("/api/testbeds", s.listTestbeds)

	// === Resource Pool API 路由（前端使用） ===
	// /api/resource-pool/admin/* - 管理员 API
	router.GET("/api/resource-pool/admin/categories", s.listCategories)
	router.POST("/api/resource-pool/admin/categories", s.createCategory)
	router.GET("/api/resource-pool/admin/categories/{id}", s.getCategory)
	router.PUT("/api/resource-pool/admin/categories/{id}", s.updateCategory)
	router.DELETE("/api/resource-pool/admin/categories/{id}", s.deleteCategory)

	router.GET("/api/resource-pool/admin/testbeds", s.listTestbeds)

	// /api/resource-pool/internal/* - 内部 API
	router.GET("/api/resource-pool/internal/testbeds", s.listTestbeds)
	router.GET("/api/resource-pool/internal/testbeds/{uuid}", s.getTestbedByUUID)
	router.GET("/api/resource-pool/internal/testbeds/available", s.listAvailableTestbeds)

	// /api/resource-pool/external/* - 外部 API
	router.GET("/api/resource-pool/external/categories", s.listCategories)
	router.GET("/api/resource-pool/external/categories/{id}", s.getCategory)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting Resource Manager API server",
		logger.String("address", s.config.Address()),
	)
	return s.apiServer.Start()
}

// Shutdown 关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.apiServer.Shutdown()
}

// 健康检查
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	sharedapi.OK(w, map[string]string{
		"status": "ok",
		"service": "resource-manager",
	})
}

// listResources 列出所有资源
func (s *Server) listResources(w http.ResponseWriter, r *http.Request) {
	resources, err := s.service.ListResources(r.Context())
	if err != nil {
		s.logger.Error("Failed to list resources", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list resources")
		return
	}
	sharedapi.OK(w, resources)
}

// getResource 获取单个资源
func (s *Server) getResource(w http.ResponseWriter, r *http.Request) {
	uuid := sharedapi.GetPathParam(r, "uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	resource, err := s.service.GetResource(r.Context(), uuid)
	if err != nil {
		s.logger.Error("Failed to get resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get resource")
		return
	}
	if resource == nil {
		sharedapi.NotFound(w, "Resource not found")
		return
	}
	sharedapi.OK(w, resource)
}

// listResourcesByCategory 按类别列出资源
func (s *Server) listResourcesByCategory(w http.ResponseWriter, r *http.Request) {
	idStr := sharedapi.GetPathParam(r, "id")
	if idStr == "" {
		sharedapi.BadRequest(w, "Category ID is required")
		return
	}

	var categoryID int64
	if _, err := fmt.Sscanf(idStr, "%d", &categoryID); err != nil {
		sharedapi.BadRequest(w, "Invalid category ID")
		return
	}

	resources, err := s.service.ListResourcesByCategory(r.Context(), categoryID)
	if err != nil {
		s.logger.Error("Failed to list resources by category", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list resources")
		return
	}
	sharedapi.OK(w, resources)
}

// listCategories 列出所有类别
func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.service.ListCategories(r.Context())
	if err != nil {
		s.logger.Error("Failed to list categories", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list categories")
		return
	}
	sharedapi.OK(w, categories)
}

// listQuotaPolicies 列出配额策略
func (s *Server) listQuotaPolicies(w http.ResponseWriter, r *http.Request) {
	idStr := sharedapi.GetPathParam(r, "id")
	if idStr == "" {
		sharedapi.BadRequest(w, "Category ID is required")
		return
	}

	var categoryID int64
	if _, err := fmt.Sscanf(idStr, "%d", &categoryID); err != nil {
		sharedapi.BadRequest(w, "Invalid category ID")
		return
	}

	policies, err := s.service.ListQuotaPolicies(r.Context(), categoryID)
	if err != nil {
		s.logger.Error("Failed to list quota policies", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list quota policies")
		return
	}
	sharedapi.OK(w, policies)
}

// listTestbeds 列出测试床
func (s *Server) listTestbeds(w http.ResponseWriter, r *http.Request) {
	testbeds, err := s.service.ListTestbeds(r.Context())
	if err != nil {
		s.logger.Error("Failed to list testbeds", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list testbeds")
		return
	}
	sharedapi.OK(w, testbeds)
}

// createResource 创建资源
func (s *Server) createResource(w http.ResponseWriter, r *http.Request) {
	var req storage.ResourceInstance
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if req.UUID == "" {
		req.UUID = uuid.New().String()
	}

	if req.Status == "" {
		req.Status = "available"
	}

	if err := s.service.CreateResource(r.Context(), &req); err != nil {
		s.logger.Error("Failed to create resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to create resource")
		return
	}

	sharedapi.OK(w, req)
}

// updateResource 更新资源
func (s *Server) updateResource(w http.ResponseWriter, r *http.Request) {
	uuid := sharedapi.GetPathParam(r, "uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	var req storage.ResourceInstance
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	req.UUID = uuid
	if err := s.service.UpdateResource(r.Context(), &req); err != nil {
		s.logger.Error("Failed to update resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to update resource")
		return
	}

	sharedapi.OK(w, req)
}

// deleteResource 删除资源
func (s *Server) deleteResource(w http.ResponseWriter, r *http.Request) {
	uuid := sharedapi.GetPathParam(r, "uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	if err := s.service.DeleteResource(r.Context(), uuid); err != nil {
		s.logger.Error("Failed to delete resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to delete resource")
		return
	}

	sharedapi.NoContent(w)
}

// createCategory 创建类别
func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	var req storage.Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if req.UUID == "" {
		req.UUID = uuid.New().String()
	}

	if err := s.service.CreateCategory(r.Context(), &req); err != nil {
		s.logger.Error("Failed to create category", logger.Err(err))
		sharedapi.InternalError(w, "Failed to create category")
		return
	}

	sharedapi.OK(w, req)
}

// getCategory 获取类别
func (s *Server) getCategory(w http.ResponseWriter, r *http.Request) {
	idStr := sharedapi.GetPathParam(r, "id")
	if idStr == "" {
		sharedapi.BadRequest(w, "Category ID is required")
		return
	}

	var categoryID int64
	if _, err := fmt.Sscanf(idStr, "%d", &categoryID); err != nil {
		sharedapi.BadRequest(w, "Invalid category ID")
		return
	}

	category, err := s.service.GetCategoryByID(r.Context(), categoryID)
	if err != nil {
		s.logger.Error("Failed to get category", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get category")
		return
	}
	if category == nil {
		sharedapi.NotFound(w, "Category not found")
		return
	}

	sharedapi.OK(w, category)
}

// updateCategory 更新类别
func (s *Server) updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := sharedapi.GetPathParam(r, "id")
	if idStr == "" {
		sharedapi.BadRequest(w, "Category ID is required")
		return
	}

	var req storage.Category
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	var categoryID int64
	if _, err := fmt.Sscanf(idStr, "%d", &categoryID); err != nil {
		sharedapi.BadRequest(w, "Invalid category ID")
		return
	}

	req.ID = categoryID
	if err := s.service.UpdateCategory(r.Context(), &req); err != nil {
		s.logger.Error("Failed to update category", logger.Err(err))
		sharedapi.InternalError(w, "Failed to update category")
		return
	}

	sharedapi.OK(w, req)
}

// deleteCategory 删除类别
func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := sharedapi.GetPathParam(r, "id")
	if idStr == "" {
		sharedapi.BadRequest(w, "Category ID is required")
		return
	}

	var categoryID int64
	if _, err := fmt.Sscanf(idStr, "%d", &categoryID); err != nil {
		sharedapi.BadRequest(w, "Invalid category ID")
		return
	}

	if err := s.service.DeleteCategory(r.Context(), categoryID); err != nil {
		s.logger.Error("Failed to delete category", logger.Err(err))
		sharedapi.InternalError(w, "Failed to delete category")
		return
	}

	sharedapi.NoContent(w)
}

// matchResource 匹配资源
func (s *Server) matchResource(w http.ResponseWriter, r *http.Request) {
	var req service.MatchResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	result, err := s.service.MatchResource(r.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to match resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to match resource")
		return
	}

	sharedapi.OK(w, result)
}

// releaseResource 释放资源
func (s *Server) releaseResource(w http.ResponseWriter, r *http.Request) {
	uuid := sharedapi.GetPathParam(r, "uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	if err := s.service.ReleaseResource(r.Context(), uuid); err != nil {
		s.logger.Error("Failed to release resource", logger.Err(err))
		sharedapi.InternalError(w, "Failed to release resource")
		return
	}

	sharedapi.OK(w, map[string]string{"status": "released"})
}

// getAllocation 获取资源分配信息
func (s *Server) getAllocation(w http.ResponseWriter, r *http.Request) {
	uuid := sharedapi.GetPathParam(r, "uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	allocation, err := s.service.GetActiveAllocation(r.Context(), uuid)
	if err != nil {
		s.logger.Error("Failed to get allocation", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get allocation")
		return
	}

	if allocation == nil {
		sharedapi.OK(w, map[string]string{"status": "not_allocated"})
		return
	}

	sharedapi.OK(w, allocation)
}

// getTestbedByUUID 获取测试床详情（通过 UUID 或 ID）
func (s *Server) getTestbedByUUID(w http.ResponseWriter, r *http.Request) {
	idOrUUID := sharedapi.GetPathParam(r, "uuid")
	if idOrUUID == "" {
		sharedapi.BadRequest(w, "UUID or ID is required")
		return
	}

	testbeds, err := s.service.ListTestbeds(r.Context())
	if err != nil {
		s.logger.Error("Failed to list testbeds", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get testbed")
		return
	}

	// 查找匹配的 testbed（通过 ID）
	var id int64
	if _, err := fmt.Sscanf(idOrUUID, "%d", &id); err == nil {
		for _, tb := range testbeds {
			if tb.ID == id {
				sharedapi.OK(w, tb)
				return
			}
		}
	}

	sharedapi.NotFound(w, "Testbed not found")
}

// listAvailableTestbeds 列出可用的测试床
func (s *Server) listAvailableTestbeds(w http.ResponseWriter, r *http.Request) {
	testbeds, err := s.service.ListTestbeds(r.Context())
	if err != nil {
		s.logger.Error("Failed to list testbeds", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list testbeds")
		return
	}

	// 简单实现：返回所有测试床
	// 实际应该根据 allocation status 过滤
	sharedapi.OK(w, testbeds)
}
