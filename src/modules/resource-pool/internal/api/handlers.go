package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hugoh/go-designs/resource-pool/internal/deployer"
	"github.com/hugoh/go-designs/resource-pool/internal/models"
	"github.com/hugoh/go-designs/resource-pool/internal/service"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// InternalAPIHandler 内部 API 处理器（无认证，供 event-processor 调用）
type InternalAPIHandler struct {
	service                 service.ResourcePoolService
	resourceInstanceStorage storage.ResourceInstanceStorage
}

// NewInternalAPIHandler 创建内部 API 处理器
func NewInternalAPIHandler(service service.ResourcePoolService, resourceInstanceStorage storage.ResourceInstanceStorage) *InternalAPIHandler {
	return &InternalAPIHandler{
		service:                 service,
		resourceInstanceStorage: resourceInstanceStorage,
	}
}

// RegisterRoutes 注册路由
func (h *InternalAPIHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/internal/testbeds/acquire", h.handleAcquireTestbed).Methods("POST")
	router.HandleFunc("/internal/testbeds/acquire-robot", h.handleAcquireTestbedForRobot).Methods("POST")
	router.HandleFunc("/internal/testbeds/{uuid}/release", h.handleReleaseTestbed).Methods("POST")
	router.HandleFunc("/internal/testbeds", h.handleListTestbeds).Methods("GET")
	router.HandleFunc("/internal/testbeds/{uuid}", h.handleGetTestbed).Methods("GET")
	router.HandleFunc("/internal/health", h.handleHealth).Methods("GET")
}

// handleAcquireTestbed 处理获取 Testbed 请求
func (h *InternalAPIHandler) handleAcquireTestbed(w http.ResponseWriter, r *http.Request) {
	var req AcquireTestbedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CategoryUUID == "" || req.Requester == "" {
		h.respondError(w, "category_uuid and requester are required", http.StatusBadRequest)
		return
	}

	allocation, testbed, err := h.service.AcquireTestbed(context.Background(), req.CategoryUUID, req.Requester)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"allocation": allocation.ToResponse(),
		"testbed":    testbed.ToResponse(),
	})
}

// handleAcquireTestbedForRobot 处理获取 Robot 专用 Testbed 请求（无需 category）
func (h *InternalAPIHandler) handleAcquireTestbedForRobot(w http.ResponseWriter, r *http.Request) {
	allocation, testbed, err := h.service.AcquireTestbedForRobot(context.Background())
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	// 获取关联的 ResourceInstance 以获取 IP、SSH 用户密码
	resourceInstance, err := h.resourceInstanceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err != nil {
		log.Printf("[handleAcquireTestbedForRobot] Warning: failed to get resource instance: %v", err)
	}

	// 构建响应，包含 Host、SSHUser、SSHPassword
	resp := testbed.ToResponse()
	if resourceInstance != nil {
		resp.Host = resourceInstance.IPAddress
		resp.IPAddress = resourceInstance.IPAddress // 添加 ip_address 字段
		resp.SSHUser = resourceInstance.SSHUser
		resp.SSHPassword = resourceInstance.Passwd
		// 同时填充 ResourceInstance 以便前端使用
		resp.ResourceInstance = &models.ResourceInstanceInfo{
			UUID:      resourceInstance.UUID,
			IPAddress: resourceInstance.IPAddress,
			Port:      resourceInstance.Port,
		}
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"allocation": allocation.ToResponse(),
		"testbed":    resp,
	})
}

// handleReleaseTestbed 处理释放 Testbed 请求
func (h *InternalAPIHandler) handleReleaseTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		h.respondError(w, "allocation uuid is required", http.StatusBadRequest)
		return
	}

	err := h.service.ReleaseTestbed(context.Background(), uuid)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Testbed released successfully",
	})
}

// handleListTestbeds 处理列出 Testbed 请求
func (h *InternalAPIHandler) handleListTestbeds(w http.ResponseWriter, r *http.Request) {
	categoryUUID := r.URL.Query().Get("category")
	statusStr := r.URL.Query().Get("status")

	var status *models.TestbedStatus
	if statusStr != "" {
		s, err := models.ParseTestbedStatus(statusStr)
		if err != nil {
			h.respondError(w, fmt.Sprintf("invalid status: %s", statusStr), http.StatusBadRequest)
			return
		}
		status = &s
	}

	var categoryUUIDPtr *string
	if categoryUUID != "" {
		categoryUUIDPtr = &categoryUUID
	}

	// 支持分页参数
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}

	pageSize := 20
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}

	testbeds, total, err := h.service.ListTestbedsWithPagination(page, pageSize, status, categoryUUIDPtr)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.TestbedResponse, len(testbeds))
	for i, t := range testbeds {
		responses[i] = t.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"testbeds":  responses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handleGetTestbed 获取 Testbed 详情（内部接口，不包含敏感信息）
func (h *InternalAPIHandler) handleGetTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	// 使用掩码密码版本的响应
	resp := testbed.ToResponseWithMaskedPassword(false)

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"testbed": resp,
	})
}

// handleHealth 健康检查
func (h *InternalAPIHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, map[string]interface{}{
		"status":  "ok",
		"service": "resource-pool",
	})
}

// ExternalAPIHandler 外部 API 处理器（需要 session 认证）
type ExternalAPIHandler struct {
	service           service.ResourcePoolService
	userStorage       UserStorage
	resourceStorage   storage.ResourceInstanceStorage
	categoryStorage   storage.CategoryStorage
	testbedStorage    storage.TestbedStorage
	allocationStorage storage.AllocationStorage
}

// NewExternalAPIHandler 创建外部 API 处理器
func NewExternalAPIHandler(service service.ResourcePoolService, userStorage UserStorage, resourceStorage storage.ResourceInstanceStorage, categoryStorage storage.CategoryStorage, testbedStorage storage.TestbedStorage, allocationStorage storage.AllocationStorage) *ExternalAPIHandler {
	return &ExternalAPIHandler{
		service:           service,
		userStorage:       userStorage,
		resourceStorage:   resourceStorage,
		categoryStorage:   categoryStorage,
		testbedStorage:    testbedStorage,
		allocationStorage: allocationStorage,
	}
}

// RegisterRoutes 注册路由
func (h *ExternalAPIHandler) RegisterRoutes(router *mux.Router) {
	// 申请 testbed
	router.HandleFunc("/external/allocations", h.withAuth(h.handleAcquireTestbed)).Methods("POST")
	// 我的分配列表
	router.HandleFunc("/external/allocations", h.withAuth(h.handleListMyAllocations)).Methods("GET")
	// 获取分配详情
	router.HandleFunc("/external/allocations/{uuid}", h.withAuth(h.handleGetMyAllocation)).Methods("GET")
	// 延长分配
	router.HandleFunc("/external/allocations/{uuid}/extend", h.withAuth(h.handleExtendMyAllocation)).Methods("POST")
	// 释放 testbed
	router.HandleFunc("/external/allocations/{uuid}", h.withAuth(h.handleReleaseMyAllocation)).Methods("DELETE")

	// 类别列表
	router.HandleFunc("/external/categories", h.withAuth(h.handleListCategories)).Methods("GET")
	// 获取类别详情
	router.HandleFunc("/external/categories/{uuid}", h.withAuth(h.handleGetCategory)).Methods("GET")
	// 获取类别配额
	router.HandleFunc("/external/categories/{uuid}/quota", h.withAuth(h.handleGetCategoryQuota)).Methods("GET")

	// Testbed 详情（所有用户可以查看）
	router.HandleFunc("/external/testbeds/{uuid}", h.withAuth(h.handleGetTestbed)).Methods("GET")

	// 我的资源实例
	router.HandleFunc("/external/resource-instances/my", h.withAuth(h.handleListMyResourceInstances)).Methods("GET")
	// 公开的资源实例（所有用户可见）
	router.HandleFunc("/external/resource-instances/public", h.withAuth(h.handleListPublicResourceInstances)).Methods("GET")
}

// withAuth 包装需要认证的处理器
func (h *ExternalAPIHandler) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 cookie 获取 session
		cookie, err := r.Cookie("session_id")
		if err != nil {
			h.respondError(w, "Unauthorized: No session", http.StatusUnauthorized)
			return
		}

		// 使用 UserStorage 验证 session
		if h.userStorage == nil {
			// 如果没有 userStorage，尝试从 query 参数获取 username（用于测试）
			username := r.URL.Query().Get("username")
			if username == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "username", username)
			handler(w, r.WithContext(ctx))
			return
		}

		session, err := h.userStorage.GetSessionWithUser(cookie.Value)
		if err != nil || session == nil {
			h.respondError(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
			return
		}

		// 从 session 中获取用户信息
		userData, ok := session["user"].(map[string]interface{})
		if !ok {
			h.respondError(w, "Unauthorized: Invalid user data", http.StatusUnauthorized)
			return
		}

		username, ok := userData["username"].(string)
		if !ok {
			h.respondError(w, "Unauthorized: Invalid username", http.StatusUnauthorized)
			return
		}

		role, _ := userData["role"].(string)

		// 将 username 和 role 添加到 context
		ctx := context.WithValue(r.Context(), "username", username)
		ctx = context.WithValue(ctx, "userRole", role)
		handler(w, r.WithContext(ctx))
	}
}

// handleListMyAllocations 处理列出我的分配请求
func (h *ExternalAPIHandler) handleListMyAllocations(w http.ResponseWriter, r *http.Request) {
	requester := r.Context().Value("username").(string)

	allocations, err := h.service.ListMyAllocations(requester)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建包含 testbed 信息的响应
	responses := make([]map[string]interface{}, len(allocations))
	for i, a := range allocations {
		allocationResp := a.ToResponse()

		// 获取关联的 testbed 信息
		testbed, err := h.testbedStorage.GetTestbedByUUID(a.TestbedUUID)
		var testbedInfo map[string]interface{}
		var categoryName string
		if err == nil && testbed != nil {
			// 获取资源实例信息（用于 SSH 连接）
			resourceInstance, _ := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)

			// 安全地获取 SSH 连接信息
			var host string
			var sshPort int
			var sshUser, sshPasswd string
			if resourceInstance != nil {
				host = resourceInstance.IPAddress
				sshPort = resourceInstance.Port
				sshUser = resourceInstance.SSHUser
				sshPasswd = resourceInstance.Passwd
			}

			testbedInfo = map[string]interface{}{
				"uuid":           testbed.UUID,
				"name":           testbed.Name,
				"category_uuid":  testbed.CategoryUUID,
				"service_target": testbed.ServiceTarget,
				"status":         testbed.Status,
				// 数据库连接信息
				"db_port":     testbed.MariaDBPort,
				"db_user":     testbed.MariaDBUser,
				"db_password": testbed.MariaDBPasswd,
				// SSH 连接信息
				"host":       host,
				"ssh_port":   sshPort,
				"ssh_user":   sshUser,
				"ssh_passwd": sshPasswd,
				"created_at": testbed.CreatedAt.Format(time.RFC3339),
			}

			// 获取类别名称
			category, _ := h.categoryStorage.GetCategoryByUUID(testbed.CategoryUUID)
			if category != nil {
				categoryName = category.Name
			}
		}

		responses[i] = map[string]interface{}{
			"id":                allocationResp.ID,
			"uuid":              allocationResp.UUID,
			"testbed_uuid":      allocationResp.TestbedUUID,
			"testbed":           testbedInfo,
			"category_uuid":     allocationResp.CategoryUUID,
			"category_name":     categoryName,
			"requester":         allocationResp.Requester,
			"status":            allocationResp.Status,
			"expires_at":        allocationResp.ExpiresAt,
			"remaining_seconds": allocationResp.RemainingSeconds,
			"created_at":        allocationResp.CreatedAt,
			"updated_at":        allocationResp.UpdatedAt,
		}
	}

	h.respondJSON(w, map[string]interface{}{
		"success":     true,
		"allocations": responses,
	})
}

// handleGetMyAllocation 处理获取我的分配详情请求
func (h *ExternalAPIHandler) handleGetMyAllocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	requester := r.Context().Value("username").(string)

	allocation, err := h.service.GetAllocation(uuid)
	if err != nil {
		h.respondError(w, "Allocation not found", http.StatusNotFound)
		return
	}

	if allocation.Requester != requester {
		h.respondError(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 获取关联的 testbed 信息
	testbed, err := h.testbedStorage.GetTestbedByUUID(allocation.TestbedUUID)
	var testbedInfo map[string]interface{}
	if err == nil && testbed != nil {
		// 获取资源实例信息（用于 SSH 连接）
		resourceInstance, _ := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)

		// 安全地获取 SSH 连接信息
		var host string
		var sshPort int
		var sshUser, sshPasswd string
		if resourceInstance != nil {
			host = resourceInstance.IPAddress
			sshPort = resourceInstance.Port
			sshUser = resourceInstance.SSHUser
			sshPasswd = resourceInstance.Passwd
		}

		testbedInfo = map[string]interface{}{
			"uuid":           testbed.UUID,
			"name":           testbed.Name,
			"category_uuid":  testbed.CategoryUUID,
			"service_target": testbed.ServiceTarget,
			"status":         testbed.Status,
			// 数据库连接信息
			"db_port":     testbed.MariaDBPort,
			"db_user":     testbed.MariaDBUser,
			"db_password": testbed.MariaDBPasswd,
			// SSH 连接信息
			"host":       host,
			"ssh_port":   sshPort,
			"ssh_user":   sshUser,
			"ssh_passwd": sshPasswd,
			"created_at": testbed.CreatedAt.Format(time.RFC3339),
		}
	}

	allocationResp := allocation.ToResponse()

	// 获取类别名称
	categoryName := ""
	if testbedInfo != nil {
		if category, err := h.categoryStorage.GetCategoryByUUID(allocation.CategoryUUID); err == nil && category != nil {
			categoryName = category.Name
		}
	}

	response := map[string]interface{}{
		"id":                allocationResp.ID,
		"uuid":              allocationResp.UUID,
		"testbed_uuid":      allocationResp.TestbedUUID,
		"testbed":           testbedInfo,
		"category_uuid":     allocationResp.CategoryUUID,
		"category_name":     categoryName,
		"requester":         allocationResp.Requester,
		"status":            allocationResp.Status,
		"expires_at":        allocationResp.ExpiresAt,
		"remaining_seconds": allocationResp.RemainingSeconds,
		"created_at":        allocationResp.CreatedAt,
		"updated_at":        allocationResp.UpdatedAt,
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"allocation": response,
	})
}

// handleReleaseMyAllocation 处理释放我的分配请求
func (h *ExternalAPIHandler) handleReleaseMyAllocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	requester := r.Context().Value("username").(string)

	allocation, err := h.service.GetAllocation(uuid)
	if err != nil {
		h.respondError(w, "Allocation not found", http.StatusNotFound)
		return
	}

	if allocation.Requester != requester {
		h.respondError(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = h.service.ReleaseTestbed(context.Background(), uuid)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Allocation released successfully",
	})
}

// handleExtendMyAllocation 处理延长我的分配请求
func (h *ExternalAPIHandler) handleExtendMyAllocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	requester := r.Context().Value("username").(string)

	allocation, err := h.service.GetAllocation(uuid)
	if err != nil {
		h.respondError(w, "Allocation not found", http.StatusNotFound)
		return
	}

	if allocation.Requester != requester {
		h.respondError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req ExtendAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AdditionalSeconds <= 0 {
		h.respondError(w, "additional_seconds must be positive", http.StatusBadRequest)
		return
	}

	err = h.service.ExtendAllocation(context.Background(), uuid, req.AdditionalSeconds)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Allocation extended successfully",
	})
}

// handleAcquireTestbed 处理申请 Testbed 请求
func (h *ExternalAPIHandler) handleAcquireTestbed(w http.ResponseWriter, r *http.Request) {
	var req AcquireTestbedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	requester := r.Context().Value("username").(string)
	if requester == "" {
		// 尝试从 query 参数获取（用于测试）
		requester = r.URL.Query().Get("username")
		if requester == "" {
			h.respondError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	allocation, testbed, err := h.service.AcquireTestbed(context.Background(), req.CategoryUUID, requester)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"allocation": allocation.ToResponse(),
		"testbed":    testbed.ToResponse(),
	})
}

// handleListCategories 处理获取类别列表请求
func (h *ExternalAPIHandler) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"categories": categories,
	})
}

// handleGetCategory 处理获取类别详情请求
func (h *ExternalAPIHandler) handleGetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	category, err := h.service.GetCategory(uuid)
	if err != nil {
		h.respondError(w, "Category not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"category": category,
	})
}

// handleGetCategoryQuota 处理获取类别配额请求
func (h *ExternalAPIHandler) handleGetCategoryQuota(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryUUID := vars["uuid"]

	quota, err := h.service.GetQuotaPolicy(categoryUUID)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"quota":   quota,
	})
}

// handleGetTestbed 处理获取 Testbed 详情请求（所有用户可查看）
func (h *ExternalAPIHandler) handleGetTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 从 context 获取用户信息
	username := r.Context().Value("username").(string)
	userRole, _ := r.Context().Value("userRole").(string)
	isAdmin := userRole == "admin"

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	// 获取资源实例信息
	resourceInstance, err := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err != nil {
		h.respondError(w, "Resource instance not found", http.StatusInternalServerError)
		return
	}

	// 检查用户是否有权限查看真实密码
	canViewRealPassword := isAdmin
	if !canViewRealPassword && testbed.CurrentAllocUUID != nil {
		// 检查用户是否是当前分配者
		if allocation, err := h.allocationStorage.GetAllocationByUUID(*testbed.CurrentAllocUUID); err == nil && allocation != nil {
			canViewRealPassword = allocation.Requester == username
		}
	}

	// 获取类别信息
	category, _ := h.categoryStorage.GetCategoryByUUID(testbed.CategoryUUID)

	// 构建响应（不使用 ToResponse，手动构建以避免重复字段）
	dbPassword := "****"
	sshPassword := "****"
	if canViewRealPassword {
		dbPassword = testbed.MariaDBPasswd
		sshPassword = resourceInstance.Passwd
	}

	resp := map[string]interface{}{
		"id":            testbed.ID,
		"uuid":          testbed.UUID,
		"name":          testbed.Name,
		"category_uuid": testbed.CategoryUUID,
		"category_name": func() string {
			if category != nil {
				return category.Name
			}
			return ""
		}(),
		"resource_instance_uuid": testbed.ResourceInstanceUUID,
		"current_alloc_uuid":     testbed.CurrentAllocUUID,
		// MariaDB 连接信息
		"mariadb_port":   testbed.MariaDBPort,
		"mariadb_user":   testbed.MariaDBUser,
		"mariadb_passwd": dbPassword,
		// 兼容前端的字段名
		"db_port":     testbed.MariaDBPort,
		"db_user":     testbed.MariaDBUser,
		"db_password": dbPassword,
		// SSH 连接信息
		"host":                 resourceInstance.IPAddress,
		"ssh_port":             resourceInstance.Port,
		"ssh_user":             resourceInstance.SSHUser,
		"ssh_passwd":           sshPassword,
		"status":               testbed.Status,
		"last_health_check":    testbed.LastHealthCheck.Format(time.RFC3339),
		"last_health_check_at": testbed.LastHealthCheck.Format(time.RFC3339),
		"created_at":           testbed.CreatedAt.Format(time.RFC3339),
		"updated_at":           testbed.UpdatedAt.Format(time.RFC3339),
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"testbed": resp,
	})
}

// handleListMyResourceInstances 处理获取我的资源实例请求
// 返回：仅用户创建的所有实例
func (h *ExternalAPIHandler) handleListMyResourceInstances(w http.ResponseWriter, r *http.Request) {
	requester := r.Context().Value("username").(string)
	if requester == "" {
		h.respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 获取用户创建的资源实例
	myInstances, err := h.resourceStorage.ListResourceInstancesByCreatedBy(requester)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.ResourceInstanceResponse, len(myInstances))
	for i, inst := range myInstances {
		responses[i] = inst.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"instances": responses,
		"total":     len(responses),
	})
}

// handleListPublicResourceInstances 处理获取公开资源实例请求
// 只返回 is_public=true 的资源实例
// 支持 resource_type 参数过滤: virtual_machine, physical_machine
func (h *ExternalAPIHandler) handleListPublicResourceInstances(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("resource_type")

	var instances []*models.ResourceInstance
	var err error

	if resourceType != "" {
		var instanceType models.InstanceType
		switch resourceType {
		case "virtual_machine":
			instanceType = models.InstanceTypeVirtualMachine
		case "physical_machine":
			instanceType = models.InstanceTypeMachine
		default:
			h.respondError(w, "Invalid resource_type: "+resourceType+". Use virtual_machine or physical_machine", http.StatusBadRequest)
			return
		}

		instances, err = h.resourceStorage.ListPublicResourceInstancesByType(instanceType)
	} else {
		instances, err = h.resourceStorage.ListPublicResourceInstances()
	}

	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.ResourceInstanceResponse, len(instances))
	for i, inst := range instances {
		responses[i] = inst.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"instances": responses,
		"total":     len(responses),
	})
}

// AdminAPIHandler 管理 API 处理器
type AdminAPIHandler struct {
	service                 service.ResourcePoolService
	testbedStorage          storage.TestbedStorage
	allocationStorage       storage.AllocationStorage
	resourceStorage         storage.ResourceInstanceStorage
	categoryStorage         storage.CategoryStorage
	policyStorage           storage.QuotaPolicyStorage
	taskStorage             storage.ResourceInstanceTaskStorage
	userStorage             UserStorage
	deployer                deployer.DeployService
	configStorage           storage.ConfigStorage
	deploymentService       service.DeploymentService
	pipelineTemplateStorage storage.DeploymentPipelineTemplateStorage
}

// NewAdminAPIHandler 创建管理 API 处理器
func NewAdminAPIHandler(
	service service.ResourcePoolService,
	testbedStorage storage.TestbedStorage,
	allocationStorage storage.AllocationStorage,
	resourceStorage storage.ResourceInstanceStorage,
	categoryStorage storage.CategoryStorage,
	policyStorage storage.QuotaPolicyStorage,
) *AdminAPIHandler {
	return &AdminAPIHandler{
		service:           service,
		testbedStorage:    testbedStorage,
		allocationStorage: allocationStorage,
		resourceStorage:   resourceStorage,
		categoryStorage:   categoryStorage,
		policyStorage:     policyStorage,
	}
}

// SetTaskStorage 设置任务存储
func (h *AdminAPIHandler) SetTaskStorage(taskStorage storage.ResourceInstanceTaskStorage) {
	h.taskStorage = taskStorage
}

// SetUserStorage 设置用户存储（用于认证）
func (h *AdminAPIHandler) SetUserStorage(userStorage UserStorage) {
	h.userStorage = userStorage
}

// SetDeployer 设置部署服务
func (h *AdminAPIHandler) SetDeployer(deployer deployer.DeployService) {
	h.deployer = deployer
}

// SetConfigStorage 设置配置存储
func (h *AdminAPIHandler) SetConfigStorage(configStorage storage.ConfigStorage) {
	h.configStorage = configStorage
}

// SetDeploymentService 设置部署服务
func (h *AdminAPIHandler) SetDeploymentService(deploymentService service.DeploymentService) {
	h.deploymentService = deploymentService
}

// SetPipelineTemplateStorage 设置部署管道模板存储
func (h *AdminAPIHandler) SetPipelineTemplateStorage(pipelineTemplateStorage storage.DeploymentPipelineTemplateStorage) {
	h.pipelineTemplateStorage = pipelineTemplateStorage
}

// RegisterRoutes 注册路由
func (h *AdminAPIHandler) RegisterRoutes(router *mux.Router) {
	// Testbed 管理
	router.HandleFunc("/admin/testbeds", h.handleListTestbeds).Methods("GET")
	router.HandleFunc("/admin/testbeds", h.withAuth(h.handleCreateTestbed)).Methods("POST")
	router.HandleFunc("/admin/testbeds/{uuid}", h.handleGetTestbed).Methods("GET")
	router.HandleFunc("/admin/testbeds/{uuid}", h.handleUpdateTestbed).Methods("PUT")
	router.HandleFunc("/admin/testbeds/{uuid}", h.handleDeleteTestbed).Methods("DELETE")
	router.HandleFunc("/admin/testbeds/{uuid}/maintenance", h.handleSetTestbedMaintenance).Methods("PUT")

	// Azure Deployment 任务管理
	router.HandleFunc("/admin/deployment-tasks", h.withAuth(h.handleListDeploymentTasks)).Methods("GET")
	router.HandleFunc("/admin/deployment-tasks/{uuid}", h.withAuth(h.handleGetDeploymentTask)).Methods("GET")
	router.HandleFunc("/admin/deployment-tasks/{uuid}/logs", h.withAuth(h.handleGetDeploymentTaskLogs)).Methods("GET")
	router.HandleFunc("/admin/deployment-tasks/{uuid}/retry", h.withAuth(h.handleRetryDeploymentTask)).Methods("POST")

	// ResourceInstance 管理（需要认证）
	router.HandleFunc("/admin/resource-instances", h.withAuth(h.handleListResourceInstances)).Methods("GET")
	router.HandleFunc("/admin/resource-instances", h.withAuth(h.handleCreateResourceInstance)).Methods("POST")
	router.HandleFunc("/admin/resource-instances/{uuid}", h.withAuth(h.handleUpdateResourceInstance)).Methods("PUT")
	router.HandleFunc("/admin/resource-instances/{uuid}", h.withAuth(h.handleDeleteResourceInstance)).Methods("DELETE")
	router.HandleFunc("/admin/resource-instances/{uuid}/deploy", h.withAuth(h.handleDeployResourceInstance)).Methods("POST")
	router.HandleFunc("/admin/resource-instances/{uuid}/health", h.withAuth(h.handleCheckResourceInstanceHealth)).Methods("GET")
	router.HandleFunc("/admin/resource-instances/{uuid}/restore-snapshot", h.withAuth(h.handleRestoreSnapshot)).Methods("POST")
	router.HandleFunc("/admin/resource-instances/test-connection", h.withAuth(h.handleTestResourceInstanceConnection)).Methods("POST")

	// Category 管理
	router.HandleFunc("/admin/categories", h.handleListCategories).Methods("GET")
	router.HandleFunc("/admin/categories", h.handleCreateCategory).Methods("POST")
	router.HandleFunc("/admin/categories/{uuid}", h.handleUpdateCategory).Methods("PUT")
	router.HandleFunc("/admin/categories/{uuid}", h.handleDeleteCategory).Methods("DELETE")

	// Quota Policy 管理
	router.HandleFunc("/admin/quota-policies", h.handleListQuotaPolicies).Methods("GET")
	router.HandleFunc("/admin/quota-policies", h.handleCreateQuotaPolicy).Methods("POST")
	router.HandleFunc("/admin/quota-policies/by-category/{categoryUUID}", h.handleGetQuotaByCategory).Methods("GET")
	router.HandleFunc("/admin/quota-policies/{uuid}", h.handleUpdateQuotaPolicy).Methods("PUT")
	router.HandleFunc("/admin/quota-policies/{uuid}", h.handleDeleteQuotaPolicy).Methods("DELETE")

	// Allocation 管理
	router.HandleFunc("/admin/allocations", h.handleListAllAllocations).Methods("GET")
	router.HandleFunc("/admin/allocations/history", h.handleListAllocationHistory).Methods("GET")

	// Task 管理（注意：更具体的路由要放在 {uuid} 之前）
	router.HandleFunc("/admin/tasks", h.handleListTasks).Methods("GET")
	router.HandleFunc("/admin/tasks/statistics", h.handleGetTaskStatistics).Methods("GET")
	router.HandleFunc("/admin/tasks/failed", h.handleListFailedTasks).Methods("GET")
	router.HandleFunc("/admin/tasks/cleanup", h.handleCleanupOldTasks).Methods("POST")
	router.HandleFunc("/admin/tasks/{uuid}", h.handleGetTask).Methods("GET")

	// 按资源实例查询任务
	router.HandleFunc("/admin/resource-instances/{uuid}/tasks", h.handleListResourceInstanceTasks).Methods("GET")

	// Metrics
	router.HandleFunc("/admin/metrics", h.handleGetMetrics).Methods("GET")
	router.HandleFunc("/admin/metrics/usage", h.handleGetUsageStats).Methods("GET")

	// Replenish 触发
	router.HandleFunc("/admin/categories/{uuid}/replenish", h.handleTriggerReplenish).Methods("POST")

	// Cleanup 数据清理
	router.HandleFunc("/admin/cleanup/testbeds", h.handleCleanupTestbeds).Methods("POST")
	router.HandleFunc("/admin/cleanup/allocations", h.handleCleanupAllocations).Methods("POST")
	router.HandleFunc("/admin/cleanup/resource-instances", h.handleCleanupResourceInstances).Methods("POST")
	router.HandleFunc("/admin/cleanup/categories", h.handleCleanupCategories).Methods("POST")
	router.HandleFunc("/admin/cleanup/quota-policies", h.handleCleanupQuotaPolicies).Methods("POST")
	router.HandleFunc("/admin/cleanup/all", h.handleCleanupAll).Methods("POST")

	// 部署管道模板管理（需要认证）
	router.HandleFunc("/admin/pipeline-templates", h.withAuth(h.handleListPipelineTemplates)).Methods("GET")
	router.HandleFunc("/admin/pipeline-templates", h.withAuth(h.handleCreatePipelineTemplate)).Methods("POST")
	router.HandleFunc("/admin/pipeline-templates/{id}", h.withAuth(h.handleGetPipelineTemplate)).Methods("GET")
	router.HandleFunc("/admin/pipeline-templates/{id}", h.withAuth(h.handleUpdatePipelineTemplate)).Methods("PUT")
	router.HandleFunc("/admin/pipeline-templates/{id}", h.withAuth(h.handleDeletePipelineTemplate)).Methods("DELETE")
	router.HandleFunc("/admin/pipeline-templates/{id}/enable", h.withAuth(h.handleEnablePipelineTemplate)).Methods("POST")
	router.HandleFunc("/admin/pipeline-templates/{id}/disable", h.withAuth(h.handleDisablePipelineTemplate)).Methods("POST")
}

// withAuth 包装需要认证的处理器
func (h *AdminAPIHandler) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Debug logging
		log.Printf("[withAuth] Request: %s %s, Cookies: %v", r.Method, r.URL.Path, r.Cookies())

		// 从 cookie 获取 session
		cookie, err := r.Cookie("session_id")
		if err != nil {
			log.Printf("[withAuth] No session_id cookie found: %v", err)
			h.respondError(w, "Unauthorized: No session", http.StatusUnauthorized)
			return
		}

		log.Printf("[withAuth] Found session_id cookie: %s", cookie.Value)

		// 使用 UserStorage 验证 session
		if h.userStorage == nil {
			// 如果没有 userStorage，尝试从 query 参数获取 username（用于测试）
			username := r.URL.Query().Get("username")
			if username == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "username", username)
			handler(w, r.WithContext(ctx))
			return
		}

		session, err := h.userStorage.GetSessionWithUser(cookie.Value)
		if err != nil || session == nil {
			log.Printf("[withAuth] Session validation failed: err=%v, session=%v", err, session)
			h.respondError(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
			return
		}

		// 从 session 中获取用户信息
		userData, ok := session["user"].(map[string]interface{})
		if !ok {
			h.respondError(w, "Unauthorized: Invalid user data", http.StatusUnauthorized)
			return
		}

		username, ok := userData["username"].(string)
		if !ok {
			h.respondError(w, "Unauthorized: Invalid username", http.StatusUnauthorized)
			return
		}

		log.Printf("[withAuth] User authenticated: %s", username)
		// 将 username 添加到 context
		ctx := context.WithValue(r.Context(), "username", username)
		handler(w, r.WithContext(ctx))
	}
}

// getUsernameFromContext 从 context 获取用户名
func (h *AdminAPIHandler) getUsernameFromContext(r *http.Request) string {
	username, _ := r.Context().Value("username").(string)
	return username
}

// handleListTestbeds 处理列出 Testbed 请求
func (h *AdminAPIHandler) handleListTestbeds(w http.ResponseWriter, r *http.Request) {
	categoryUUID := r.URL.Query().Get("category")
	statusStr := r.URL.Query().Get("status")
	testbedUUID := r.URL.Query().Get("uuid")

	// 如果指定了 uuid，返回单个 testbed 详情（兼容前端）
	if testbedUUID != "" {
		testbed, err := h.service.GetTestbed(testbedUUID)
		if err != nil {
			h.respondError(w, "Testbed not found", http.StatusNotFound)
			return
		}

		resp := testbed.ToResponse()
		// 填充类别名称
		category, err := h.categoryStorage.GetCategoryByUUID(testbed.CategoryUUID)
		if err == nil {
			resp.CategoryName = &category.Name
		}

		// 填充资源实例信息
		resourceInstance, err := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
		if err == nil {
			resp.ResourceInstance = &models.ResourceInstanceInfo{
				UUID:         resourceInstance.UUID,
				Name:         resourceInstance.UUID,
				IPAddress:    resourceInstance.IPAddress,
				Port:         resourceInstance.Port,
				InstanceType: string(resourceInstance.InstanceType),
				SnapshotID: func() string {
					if resourceInstance.SnapshotID != nil {
						return *resourceInstance.SnapshotID
					}
					return ""
				}(),
				Status: string(resourceInstance.Status),
			}
			resp.Host = resourceInstance.IPAddress
			resp.SSHPort = resourceInstance.Port
		}

		h.respondJSON(w, map[string]interface{}{
			"success": true,
			"testbed": resp, // 单个 testbed，兼容前端
		})
		return
	}

	var status *models.TestbedStatus
	if statusStr != "" {
		s, err := models.ParseTestbedStatus(statusStr)
		if err != nil {
			h.respondError(w, fmt.Sprintf("invalid status: %s", statusStr), http.StatusBadRequest)
			return
		}
		status = &s
	} else {
		// 默认排除已删除的 testbed
		s := models.TestbedStatusAvailable
		status = &s
	}

	var categoryUUIDPtr *string
	if categoryUUID != "" {
		categoryUUIDPtr = &categoryUUID
	}

	// 支持分页参数
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}

	pageSize := 20
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}

	testbeds, total, err := h.service.ListTestbedsWithPagination(page, pageSize, status, categoryUUIDPtr)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取所有类别用于填充类别名称
	categories, _ := h.categoryStorage.ListCategories()
	categoryMap := make(map[string]string)
	for _, cat := range categories {
		categoryMap[cat.UUID] = cat.Name
	}

	// 获取所有策略用于计算过期时间
	policies, _ := h.policyStorage.ListQuotaPolicies()
	// 构建 policy key (categoryUUID + serviceTarget) -> policy 的映射
	policyMap := make(map[string]*models.QuotaPolicy)
	for _, p := range policies {
		key := p.CategoryUUID + "|" + string(p.ServiceTarget)
		policyMap[key] = p
	}

	// 获取所有资源实例用于填充资源实例信息（简化版）
	resources, _ := h.resourceStorage.ListResourceInstances()
	resourceMap := make(map[string]*models.ResourceInstance)
	for _, res := range resources {
		resourceMap[res.UUID] = res
	}

	responses := make([]models.TestbedResponse, 0, len(testbeds))
	for _, t := range testbeds {
		resp := t.ToResponse()
		// 填充类别名称
		if name, ok := categoryMap[t.CategoryUUID]; ok {
			resp.CategoryName = &name
		}
		// 填充资源实例基本信息
		if res, ok := resourceMap[t.ResourceInstanceUUID]; ok {
			resp.Host = res.IPAddress
			resp.SSHPort = res.Port
		}
		// 计算过期时间
		// 如果有当前分配，使用 allocation 的过期时间（已预先计算）
		if t.CurrentAllocUUID != nil {
			if alloc, err := h.allocationStorage.GetAllocationByUUID(*t.CurrentAllocUUID); err == nil && alloc.ExpiresAt != nil {
				expiresAtStr := alloc.ExpiresAt.Format(time.RFC3339)
				resp.ExpiresAt = &expiresAtStr
			}
		} else {
			// 如果没有分配，使用 testbed 的创建时间 + 策略的 MaxLifetimeSeconds
			policyKey := t.CategoryUUID + "|" + string(t.ServiceTarget)
			if policy, ok := policyMap[policyKey]; ok && policy.MaxLifetimeSeconds > 0 {
				expiresAt := t.CreatedAt.Add(time.Duration(policy.MaxLifetimeSeconds) * time.Second)
				expiresAtStr := expiresAt.Format(time.RFC3339)
				resp.ExpiresAt = &expiresAtStr
			}
		}
		responses = append(responses, resp)
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"testbeds":  responses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// handleGetTestbed 处理获取 Testbed 详情请求
func (h *AdminAPIHandler) handleGetTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	resp := testbed.ToResponse()
	// 填充类别名称
	category, err := h.categoryStorage.GetCategoryByUUID(testbed.CategoryUUID)
	if err == nil {
		resp.CategoryName = &category.Name
	}

	// 填充资源实例信息
	resourceInstance, err := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err == nil {
		resp.ResourceInstance = &models.ResourceInstanceInfo{
			UUID:         resourceInstance.UUID,
			Name:         resourceInstance.UUID, // ResourceInstance 没有 Name 字段，使用 UUID
			IPAddress:    resourceInstance.IPAddress,
			Port:         resourceInstance.Port,
			InstanceType: string(resourceInstance.InstanceType),
			SnapshotID: func() string {
				if resourceInstance.SnapshotID != nil {
					return *resourceInstance.SnapshotID
				}
				return ""
			}(),
			Status: string(resourceInstance.Status),
		}
		resp.Host = resourceInstance.IPAddress
		resp.SSHPort = resourceInstance.Port
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"testbed": resp,
	})
}

// handleSetTestbedMaintenance 设置 Testbed 维护模式（已废弃，改为删除 Testbed）
// Testbed 是一次性的，使用后应删除而非维护
func (h *AdminAPIHandler) handleSetTestbedMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	// 标记为已删除（一次性使用）
	testbed.MarkDeleted()
	err = h.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		h.respondError(w, "Failed to delete testbed", http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Testbed deleted",
	})
}

// handleListCategories 处理列出 Category 请求
func (h *AdminAPIHandler) handleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.ListCategories()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.CategoryResponse, len(categories))
	for i, c := range categories {
		response := c.ToResponse()

		// 获取 testbed 统计信息
		totalCount, err := h.testbedStorage.CountTestbedsByCategory(c.UUID)
		if err == nil {
			response.TotalCount = totalCount
		}

		allocatedCount, err := h.testbedStorage.CountAllocatedTestbedsByCategory(c.UUID)
		if err == nil {
			response.AllocatedCount = allocatedCount
		}

		response.AvailableCount = response.TotalCount - response.AllocatedCount

		responses[i] = response
	}

	h.respondJSON(w, map[string]interface{}{
		"success":    true,
		"categories": responses,
	})
}

// handleCreateCategory 处理创建 Category 请求
func (h *AdminAPIHandler) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.respondError(w, "name is required", http.StatusBadRequest)
		return
	}

	category := models.NewCategory(req.Name, req.Description)
	if req.Enabled != nil {
		category.Enabled = *req.Enabled
	} else {
		category.Disable()
	}

	err := h.service.CreateCategory(category)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusConflict)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"category": category.ToResponse(),
	})
}

// handleUpdateCategory 处理更新 Category 请求
func (h *AdminAPIHandler) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	category, err := h.service.GetCategory(uuid)
	if err != nil {
		h.respondError(w, "Category not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	if req.Enabled != nil {
		if *req.Enabled {
			category.Enable()
		} else {
			category.Disable()
		}
	}

	err = h.service.UpdateCategory(category)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"category": category.ToResponse(),
	})
}

// handleDeleteCategory 处理删除 Category 请求
func (h *AdminAPIHandler) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	err := h.service.DeleteCategory(uuid)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Category deleted successfully",
	})
}

// handleListQuotas 处理列出配额策略请求
func (h *AdminAPIHandler) handleListQuotas(w http.ResponseWriter, r *http.Request) {
	quotas, err := h.service.ListQuotaPolicies()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.QuotaPolicyResponse, len(quotas))
	for i, q := range quotas {
		responses[i] = q.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"quotas":  responses,
	})
}

// handleSetQuota 处理设置配额策略请求
func (h *AdminAPIHandler) handleSetQuota(w http.ResponseWriter, r *http.Request) {
	var req SetQuotaPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CategoryUUID == "" {
		h.respondError(w, "category_uuid is required", http.StatusBadRequest)
		return
	}

	if req.MinInstances < 0 || req.MaxInstances < req.MinInstances {
		h.respondError(w, "Invalid min_instances or max_instances", http.StatusBadRequest)
		return
	}

	policy := models.NewQuotaPolicy(req.CategoryUUID, req.MinInstances, req.MaxInstances, req.Priority, req.MaxLifetimeSeconds)
	if req.AutoReplenish != nil {
		policy.AutoReplenish = *req.AutoReplenish
	}
	if req.ReplenishThreshold > 0 {
		policy.ReplenishThreshold = req.ReplenishThreshold
	}

	err := h.service.SetQuotaPolicy(policy)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"policy":  policy.ToResponse(),
	})
}

// handleGetQuota 处理获取配额策略请求
func (h *AdminAPIHandler) handleGetQuota(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	policy, err := h.service.GetQuotaPolicy(uuid)
	if err != nil {
		h.respondError(w, "Quota policy not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"policy":  policy.ToResponse(),
	})
}

// handleTriggerReplenish 处理触发补充请求
func (h *AdminAPIHandler) handleTriggerReplenish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 调用服务层的补充方法
	err := h.service.ReplenishCategory(uuid)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Replenish triggered for category: " + uuid,
	})
}

// respondJSON 返回 JSON 响应
func (h *InternalAPIHandler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// respondError 返回错误响应
func (h *InternalAPIHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// respondJSON 返回 JSON 响应
func (h *ExternalAPIHandler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// respondError 返回错误响应
func (h *ExternalAPIHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// respondJSON 返回 JSON 响应
func (h *AdminAPIHandler) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// respondError 返回错误响应
func (h *AdminAPIHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}

// Request/Response DTOs

type AcquireTestbedRequest struct {
	CategoryUUID string `json:"category_uuid"`
	Requester    string `json:"requester"`
}

type ExtendAllocationRequest struct {
	AdditionalSeconds int `json:"additional_seconds"`
}

type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

type SetQuotaPolicyRequest struct {
	CategoryUUID       string `json:"category_uuid"`
	MinInstances       int    `json:"min_instances"`
	MaxInstances       int    `json:"max_instances"`
	Priority           int    `json:"priority"`
	AutoReplenish      *bool  `json:"auto_replenish"`
	ReplenishThreshold int    `json:"replenish_threshold"`
	MaxLifetimeSeconds int    `json:"max_lifetime_seconds"`
}

// handleUpdateTestbed 处理更新 Testbed 请求
func (h *AdminAPIHandler) handleUpdateTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	// 更新状态（如果提供）
	statusStr := r.URL.Query().Get("status")
	if statusStr != "" {
		status, err := models.ParseTestbedStatus(statusStr)
		if err != nil {
			h.respondError(w, "Invalid status", http.StatusBadRequest)
			return
		}
		testbed.Status = status
		testbed.UpdatedAt = testbed.CreatedAt // 简化处理
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"testbed": testbed.ToResponse(),
	})
}

// handleCreateTestbed 处理创建 Testbed 请求
// 这实际上是通过部署产品到 ResourceInstance 来创建 Testbed
func (h *AdminAPIHandler) handleCreateTestbed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceInstanceUUID string `json:"resource_instance_uuid"`
		CategoryUUID         string `json:"category_uuid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 验证 resource_instance_uuid
	if req.ResourceInstanceUUID == "" {
		h.respondError(w, "resource_instance_uuid is required", http.StatusBadRequest)
		return
	}

	// 验证资源实例存在
	resourceInstance, err := h.resourceStorage.GetResourceInstanceByUUID(req.ResourceInstanceUUID)
	if err != nil {
		h.respondError(w, "Resource instance not found", http.StatusNotFound)
		return
	}

	// 检查资源实例状态
	if resourceInstance.Status != models.ResourceInstanceStatusActive && resourceInstance.Status != models.ResourceInstanceStatusPending {
		h.respondError(w, fmt.Sprintf("Cannot deploy to resource instance in status: %s", resourceInstance.Status), http.StatusConflict)
		return
	}

	// 验证资源实例是虚拟机（只有虚拟机可以创建 testbed）
	if !resourceInstance.IsVirtualMachine() {
		h.respondError(w, "Testbed can only be created from VirtualMachine type resource instances", http.StatusBadRequest)
		return
	}

	// 如果提供了 category_uuid，验证资源实例未被其他类别占用
	if req.CategoryUUID != "" {
		category, err := h.categoryStorage.GetCategoryByUUID(req.CategoryUUID)
		if err != nil {
			h.respondError(w, "Category not found", http.StatusNotFound)
			return
		}
		_ = category // Category 验证通过
	}

	// 调用 ProvisionTestbed 创建 testbed
	testbed, err := h.service.ProvisionTestbed(context.Background(), req.ResourceInstanceUUID)
	if err != nil {
		h.respondError(w, "Failed to create testbed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 填充响应信息
	resp := testbed.ToResponse()
	if category, err := h.categoryStorage.GetCategoryByUUID(testbed.CategoryUUID); err == nil {
		resp.CategoryName = &category.Name
	}
	resp.Host = resourceInstance.IPAddress
	resp.SSHPort = resourceInstance.Port

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Testbed created successfully",
		"testbed": resp,
	})
}

// handleDeleteTestbed 处理删除 Testbed 请求
func (h *AdminAPIHandler) handleDeleteTestbed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	testbed, err := h.service.GetTestbed(uuid)
	if err != nil {
		h.respondError(w, "Testbed not found", http.StatusNotFound)
		return
	}

	// 检查当前状态
	if testbed.Status == models.TestbedStatusDeleted {
		h.respondJSON(w, map[string]interface{}{
			"success": true,
			"message": "Testbed already deleted",
		})
		return
	}

	// 如果 testbed 正在使用中，不允许删除
	if testbed.Status == models.TestbedStatusInUse || testbed.Status == models.TestbedStatusAllocated {
		h.respondError(w, "Cannot delete testbed in use. Release it first.", http.StatusConflict)
		return
	}

	// 标记为已删除
	testbed.MarkDeleted()
	err = h.testbedStorage.UpdateTestbed(testbed)
	if err != nil {
		h.respondError(w, "Failed to delete testbed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 如果有关联的 resource instance，异步触发快照回滚
	if testbed.ResourceInstanceUUID != "" {
		go func() {
			resourceInstance, err := h.resourceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
			if err == nil && resourceInstance.IsVirtualMachine() && resourceInstance.SnapshotID != nil && resourceInstance.SnapshotInstanceUUID != nil && h.deployer != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				_ = h.deployer.RestoreSnapshot(ctx, *resourceInstance.SnapshotInstanceUUID, *resourceInstance.SnapshotID)
			}
		}()
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Testbed deleted successfully",
	})
}

// handleListResourceInstances 处理列出资源实例请求
func (h *AdminAPIHandler) handleListResourceInstances(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	statusStr := r.URL.Query().Get("status")
	resourceType := r.URL.Query().Get("type")
	search := r.URL.Query().Get("search")

	// 获取所有实例
	instances, err := h.resourceStorage.ListResourceInstances()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 过滤
	var filtered []*models.ResourceInstance
	for _, inst := range instances {
		// 状态过滤
		if statusStr != "" {
			status, err := ParseResourceInstanceStatus(statusStr)
			if err == nil && inst.Status != status {
				continue
			}
		}

		// 类型过滤
		if resourceType != "" {
			// 前端使用 virtual_machine/physical_machine，后端使用 VirtualMachine/Machine
			var targetType models.InstanceType
			if resourceType == "virtual_machine" {
				targetType = models.InstanceTypeVirtualMachine
			} else if resourceType == "physical_machine" {
				targetType = models.InstanceTypeMachine
			} else {
				continue
			}
			if inst.InstanceType != targetType {
				continue
			}
		}

		// 搜索过滤
		if search != "" {
			// 简化处理：只检查 UUID 和 IP 地址
			if !contains(inst.UUID, search) && !contains(inst.IPAddress, search) {
				continue
			}
		}

		filtered = append(filtered, inst)
	}

	responses := make([]models.ResourceInstanceResponse, len(filtered))
	for i, inst := range filtered {
		responses[i] = inst.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"data":      responses,
		"instances": responses,
		"total":     len(responses),
	})
}

// handleCreateResourceInstance 处理创建资源实例请求
func (h *AdminAPIHandler) handleCreateResourceInstance(w http.ResponseWriter, r *http.Request) {
	var req CreateResourceInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" {
		h.respondError(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.ResourceType == "" {
		h.respondError(w, "resource_type is required", http.StatusBadRequest)
		return
	}
	if req.Host == "" {
		h.respondError(w, "host is required", http.StatusBadRequest)
		return
	}

	// 检查 IP 地址是否已存在
	existingInstance, err := h.resourceStorage.GetResourceInstanceByIPAddress(req.Host)
	if err == nil && existingInstance != nil {
		h.respondError(w, fmt.Sprintf("资源实例已存在，IP 地址 %s 已被使用", req.Host), http.StatusConflict)
		return
	}

	if req.SSHPort <= 0 {
		h.respondError(w, "ssh_port is required", http.StatusBadRequest)
		return
	}

	// 从认证上下文获取当前登录用户名
	createdBy := h.getUsernameFromContext(r)
	if createdBy == "" {
		createdBy = "unknown" // 如果获取失败，使用 unknown 而不是 admin
	}

	// 创建资源实例
	var instance *models.ResourceInstance

	// 转换前端 resource_type 到后端 InstanceType
	// 前端使用: virtual_machine, physical_machine
	// 后端使用: VirtualMachine, Machine
	var instanceType models.InstanceType
	switch req.ResourceType {
	case "virtual_machine":
		instanceType = models.InstanceTypeVirtualMachine
	case "physical_machine":
		instanceType = models.InstanceTypeMachine
	default:
		h.respondError(w, "Invalid resource_type: "+req.ResourceType+". Use virtual_machine or physical_machine", http.StatusBadRequest)
		return
	}

	now := time.Now()

	// 设置默认 SSH 用户
	sshUser := req.SSHUser
	if sshUser == "" {
		sshUser = "root" // 默认 SSH 用户
	}

	if instanceType == models.InstanceTypeVirtualMachine {
		// 虚拟机必须有 snapshot_id
		if req.SnapshotID == "" {
			h.respondError(w, "snapshot_id is required for virtual_machine", http.StatusBadRequest)
			return
		}

		// 处理 snapshot_instance_uuid
		var snapshotInstanceUUID *string
		if req.SnapshotInstanceUUID != "" {
			snapshotInstanceUUID = &req.SnapshotInstanceUUID
		}

		instance = &models.ResourceInstance{
			UUID:                 generateUUID(),
			InstanceType:         models.InstanceTypeVirtualMachine,
			SnapshotID:           &req.SnapshotID,
			SnapshotInstanceUUID: snapshotInstanceUUID,
			IPAddress:            req.Host,
			Port:                 req.SSHPort,
			SSHUser:              sshUser,
			Passwd:               req.Password,
			Description:          req.Description,
			IsPublic:             true, // 虚拟机强制公开
			CreatedBy:            createdBy,
			Status:               models.ResourceInstanceStatusPending, // 新创建的实例默认为 pending
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	} else {
		// 物理机不需要 snapshot_id
		// 使用前端传入的 is_public，默认为 true
		isPublic := true
		if req.IsPublic != nil {
			isPublic = *req.IsPublic
		}
		instance = &models.ResourceInstance{
			UUID:         generateUUID(),
			InstanceType: models.InstanceTypeMachine,
			SnapshotID:   nil,
			IPAddress:    req.Host,
			Port:         req.SSHPort,
			SSHUser:      sshUser,
			Passwd:       req.Password,
			Description:  req.Description,
			IsPublic:     isPublic,
			CreatedBy:    createdBy,
			Status:       models.ResourceInstanceStatusPending, // 新创建的实例默认为 pending
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	// 保存到数据库
	if err := h.resourceStorage.CreateResourceInstance(instance); err != nil {
		h.respondError(w, "Failed to create resource instance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 异步执行健康检查（不阻塞响应）
	if h.deployer != nil {
		go func(uuid string) {
			healthy, _ := h.deployer.CheckHealth(context.Background(), instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)
			var newStatus models.ResourceInstanceStatus
			if healthy {
				newStatus = models.ResourceInstanceStatusActive
			} else {
				newStatus = models.ResourceInstanceStatusUnreachable
			}
			if err := h.resourceStorage.UpdateResourceInstanceStatus(uuid, newStatus); err != nil {
				fmt.Printf("Warning: Failed to update resource instance status after creation: %v\n", err)
			}
		}(instance.UUID)

		// 虚拟机类型在创建后更新 /etc/resolv.conf
		if instance.InstanceType == models.InstanceTypeVirtualMachine {
			go func() {
				err := deployer.UpdateResolvConf(context.Background(), instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd, 30*time.Second)
				if err != nil {
					fmt.Printf("Warning: Failed to update resolv.conf for resource instance %s: %v\n", instance.UUID, err)
				}
			}()
		}
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Resource instance created successfully. Health check is running in background.",
		"data":    instance.ToResponse(),
	})
}

// handleUpdateResourceInstance 处理更新资源实例请求
func (h *AdminAPIHandler) handleUpdateResourceInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 获取现有实例
	instance, err := h.resourceStorage.GetResourceInstanceByUUID(uuid)
	if err != nil {
		h.respondError(w, "Resource instance not found", http.StatusNotFound)
		return
	}

	var req UpdateResourceInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 更新字段
	if req.Name != "" {
		// name 字段不存储在数据库中，忽略
	}
	if req.Host != "" {
		// 检查新 IP 地址是否被其他实例使用
		existingInstance, err := h.resourceStorage.GetResourceInstanceByIPAddress(req.Host)
		if err == nil && existingInstance != nil && existingInstance.UUID != uuid {
			h.respondError(w, fmt.Sprintf("IP 地址 %s 已被其他资源实例使用", req.Host), http.StatusConflict)
			return
		}
		instance.IPAddress = req.Host
	}
	if req.SSHPort > 0 {
		instance.Port = req.SSHPort
	}
	if req.SSHUser != "" {
		instance.SSHUser = req.SSHUser
	}
	if req.Password != "" {
		instance.Passwd = req.Password
	}
	if req.SnapshotID != nil {
		instance.SnapshotID = req.SnapshotID
	}
	if req.SnapshotInstanceUUID != nil {
		instance.SnapshotInstanceUUID = req.SnapshotInstanceUUID
	}
	if req.Description != nil {
		instance.Description = req.Description
	}
	if req.IsPublic != nil {
		instance.IsPublic = *req.IsPublic
	}
	// 更新状态
	if req.Status != "" {
		status, err := ParseResourceInstanceStatus(req.Status)
		if err == nil {
			instance.Status = status
		}
	}

	instance.UpdatedAt = time.Now()

	// 保存到数据库
	if err := h.resourceStorage.UpdateResourceInstance(instance); err != nil {
		h.respondError(w, "Failed to update resource instance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Resource instance updated successfully",
		"data":    instance.ToResponse(),
	})
}

// handleDeleteResourceInstance 处理删除资源实例请求
func (h *AdminAPIHandler) handleDeleteResourceInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 获取现有实例
	instance, err := h.resourceStorage.GetResourceInstanceByUUID(uuid)
	if err != nil {
		h.respondError(w, "Resource instance not found", http.StatusNotFound)
		return
	}

	// 检查是否有关联的 Testbed
	// TODO: 添加检查逻辑

	// 删除
	if err := h.resourceStorage.DeleteResourceInstance(instance.ID); err != nil {
		h.respondError(w, "Failed to delete resource instance: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Resource instance deleted successfully",
	})
}

// handleDeployResourceInstance 处理部署资源实例请求
func (h *AdminAPIHandler) handleDeployResourceInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var req struct {
		CategoryUUID string `json:"category_uuid"`
		DBPort       int    `json:"db_port"`
		DBUser       string `json:"db_user"`
		DBPassword   string `json:"db_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: 实现完整的部署逻辑
	testbed, err := h.service.ProvisionTestbed(context.Background(), uuid)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"testbed": testbed.ToResponse(),
	})
}

// handleCheckResourceInstanceHealth 处理检查资源实例健康状态请求
func (h *AdminAPIHandler) handleCheckResourceInstanceHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 获取资源实例
	instance, err := h.resourceStorage.GetResourceInstanceByUUID(uuid)
	if err != nil {
		h.respondError(w, "Resource instance not found", http.StatusNotFound)
		return
	}

	// 使用 deployer 检查健康状态
	if h.deployer == nil {
		h.respondError(w, "Deployer not available", http.StatusServiceUnavailable)
		return
	}

	// 执行健康检查，传入 SSH 凭据
	healthy, err := h.deployer.CheckHealth(context.Background(), instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)

	var status string
	var message string
	var newStatus models.ResourceInstanceStatus

	if healthy {
		status = "healthy"
		message = "资源实例健康"
		newStatus = models.ResourceInstanceStatusActive
	} else {
		status = "unhealthy"
		message = fmt.Sprintf("资源实例不可达: %s", err.Error())
		newStatus = models.ResourceInstanceStatusUnreachable
	}

	// 根据健康检查结果更新资源实例状态
	if err := h.resourceStorage.UpdateResourceInstanceStatus(uuid, newStatus); err != nil {
		// 更新状态失败，但仍返回健康检查结果
		fmt.Printf("Warning: Failed to update resource instance status: %v\n", err)
	}

	// 刷新实例数据以获取最新状态
	instance, _ = h.resourceStorage.GetResourceInstanceByUUID(uuid)

	h.respondJSON(w, map[string]interface{}{
		"success":    healthy,
		"healthy":    healthy,
		"status":     status,
		"message":    message,
		"ip_address": instance.IPAddress,
		"port":       instance.Port,
		// 返回更新后的资源实例状态
		"instance_status": string(instance.Status),
	})
}

type TestConnectionRequest struct {
	Host     string `json:"host"`
	SSHPort  int    `json:"ssh_port"`
	SSHUser  string `json:"ssh_user"`
	Password string `json:"password"`
}

// handleTestResourceInstanceConnection 测试资源实例连接（在创建前）
func (h *AdminAPIHandler) handleTestResourceInstanceConnection(w http.ResponseWriter, r *http.Request) {
	var req TestConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证必要字段
	if req.Host == "" || req.SSHPort == 0 || req.SSHUser == "" {
		h.respondError(w, "Missing required fields: host, ssh_port, ssh_user", http.StatusBadRequest)
		return
	}

	// 使用 deployer 测试连接
	if h.deployer == nil {
		h.respondError(w, "Deployer not available", http.StatusServiceUnavailable)
		return
	}

	// 执行连接测试，传入 SSH 凭据
	healthy, err := h.deployer.CheckHealth(context.Background(), req.Host, req.SSHPort, req.SSHUser, req.Password)

	if healthy {
		h.respondJSON(w, map[string]interface{}{
			"success": true,
			"healthy": true,
			"status":  "healthy",
			"message": "连接成功",
		})
	} else {
		h.respondJSON(w, map[string]interface{}{
			"success": false,
			"healthy": false,
			"status":  "unhealthy",
			"message": "连接失败: " + err.Error(),
		})
	}
}

// handleRestoreSnapshot 处理快照回滚请求
func (h *AdminAPIHandler) handleRestoreSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 获取资源实例
	instance, err := h.resourceStorage.GetResourceInstanceByUUID(uuid)
	if err != nil {
		h.respondError(w, "Resource instance not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// 检查是否有快照ID
	if instance.SnapshotID == nil || *instance.SnapshotID == "" {
		h.respondError(w, "Resource instance does not have a snapshot_id", http.StatusBadRequest)
		return
	}

	// 检查是否有快照实例UUID
	if instance.SnapshotInstanceUUID == nil || *instance.SnapshotInstanceUUID == "" {
		h.respondError(w, "Resource instance does not have a snapshot_instance_uuid", http.StatusBadRequest)
		return
	}

	// 从配置获取 CMP 凭据
	if h.configStorage == nil {
		h.respondError(w, "Config storage not available", http.StatusInternalServerError)
		return
	}

	accessKey, err := h.configStorage.GetConfig("cmp_access_key")
	if err != nil || accessKey == "" {
		h.respondError(w, "CMP access key not configured. Please configure it in system settings.", http.StatusBadRequest)
		return
	}

	secretKey, err := h.configStorage.GetConfig("cmp_secret_key")
	if err != nil || secretKey == "" {
		h.respondError(w, "CMP secret key not configured. Please configure it in system settings.", http.StatusBadRequest)
		return
	}

	apiURL, _ := h.configStorage.GetConfig("cmp_api_url")
	if apiURL == "" {
		apiURL = "http://devops-api.aishu.cn:8081"
	}

	// 创建 CMP 部署器
	cmpDeployer := deployer.NewCMPDeployService(apiURL, accessKey, secretKey)

	// 异步执行快照回滚（因为可能需要很长时间）
	go func() {
		fmt.Printf("[SnapshotRestore] Starting restore for instance %s, snapshot %s\n",
			*instance.SnapshotInstanceUUID, *instance.SnapshotID)

		// 更新资源实例状态为 pending（回滚中）
		h.resourceStorage.UpdateResourceInstanceStatus(uuid, models.ResourceInstanceStatusPending)

		// 执行快照回滚
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		err := cmpDeployer.RestoreSnapshot(ctx, *instance.SnapshotInstanceUUID, *instance.SnapshotID)
		if err != nil {
			fmt.Printf("[SnapshotRestore] Failed: %v\n", err)
			// 回滚失败，恢复状态
			h.resourceStorage.UpdateResourceInstanceStatus(uuid, models.ResourceInstanceStatusActive)
			return
		}

		// 通过健康检查轮询等待回滚完成（每30秒检查一次，最多等待10分钟）
		fmt.Printf("[SnapshotRestore] Snapshot initiated, polling health check every 30 seconds...\n")
		restoreCtx, restoreCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer restoreCancel()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		restoreCompleted := false
		for {
			select {
			case <-restoreCtx.Done():
				fmt.Printf("[SnapshotRestore] Timeout waiting for restore completion\n")
				h.resourceStorage.UpdateResourceInstanceStatus(uuid, models.ResourceInstanceStatusActive)
				return
			case <-ticker.C:
				// 执行健康检查
				healthy, err := cmpDeployer.CheckHealth(restoreCtx, instance.IPAddress, instance.Port, instance.SSHUser, instance.Passwd)
				if err != nil {
					fmt.Printf("[SnapshotRestore] Health check error: %v\n", err)
					continue
				}
				if healthy {
					fmt.Printf("[SnapshotRestore] Health check passed, restore completed\n")
					restoreCompleted = true
				} else {
					fmt.Printf("[SnapshotRestore] Health check not passed yet, waiting...\n")
				}

				if restoreCompleted {
					// 回滚完成，更新状态
					fmt.Printf("[SnapshotRestore] Completed for instance %s\n", uuid)
					h.resourceStorage.UpdateResourceInstanceStatus(uuid, models.ResourceInstanceStatusActive)
					return
				}
			}
		}
	}()

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Snapshot restore initiated for instance %s. This will take approximately 5 minutes to complete.", instance.UUID),
		"data": map[string]interface{}{
			"instance_uuid":             instance.UUID,
			"snapshot_id":               *instance.SnapshotID,
			"snapshot_instance_uuid":    *instance.SnapshotInstanceUUID,
			"estimated_completion_time": "5 minutes",
		},
	})
}

// handleListQuotaPolicies 处理列出配额策略请求
func (h *AdminAPIHandler) handleListQuotaPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.service.ListQuotaPolicies()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"policies": policies,
	})
}

// handleCreateQuotaPolicy 处理创建配额策略请求
func (h *AdminAPIHandler) handleCreateQuotaPolicy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CategoryUUID string `json:"category_uuid"`
		MinInstances *int   `json:"min_instances"`
		MaxInstances *int   `json:"max_instances"`
		// 旧字段名，保持向后兼容
		MaxAllocations     *int   `json:"max_allocations"`
		MaxLifetimeSeconds *int   `json:"max_lifetime_seconds"`
		AutoReplenish      *bool  `json:"auto_replenish"`
		ReplenishThreshold *int   `json:"replenish_threshold"`
		Priority           *int   `json:"priority"`
		ServiceTarget      string `json:"service_target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.CategoryUUID == "" {
		h.respondError(w, "category_uuid is required", http.StatusBadRequest)
		return
	}

	// 确定最大实例数（优先使用新字段名）
	maxInstances := 10 // 默认值
	if req.MaxInstances != nil {
		if *req.MaxInstances <= 0 {
			h.respondError(w, "max_instances must be greater than 0", http.StatusBadRequest)
			return
		}
		maxInstances = *req.MaxInstances
	} else if req.MaxAllocations != nil {
		if *req.MaxAllocations <= 0 {
			h.respondError(w, "max_allocations must be greater than 0", http.StatusBadRequest)
			return
		}
		maxInstances = *req.MaxAllocations
	}

	// 确定最小实例数
	minInstances := 0
	if req.MinInstances != nil {
		minInstances = *req.MinInstances
	}

	// 确定默认生命周期
	maxLifetimeSeconds := 86400 // 默认24小时
	if req.MaxLifetimeSeconds != nil {
		if *req.MaxLifetimeSeconds <= 0 {
			h.respondError(w, "max_lifetime_seconds must be greater than 0", http.StatusBadRequest)
			return
		}
		maxLifetimeSeconds = *req.MaxLifetimeSeconds
	}

	// 确定优先级
	priority := 100 // 默认值
	if req.Priority != nil {
		priority = *req.Priority
	}

	// 确定自动补充设置
	autoReplenish := false
	if req.AutoReplenish != nil {
		autoReplenish = *req.AutoReplenish
	}

	replenishThreshold := minInstances * 80 / 100
	if replenishThreshold < 1 {
		replenishThreshold = 1
	}
	if req.ReplenishThreshold != nil {
		replenishThreshold = *req.ReplenishThreshold
	}

	if autoReplenish && replenishThreshold <= 0 {
		h.respondError(w, "replenish_threshold is required when auto_replenish is true", http.StatusBadRequest)
		return
	}

	// 解析服务对象类型
	serviceTarget := models.ServiceTargetNormal
	if req.ServiceTarget != "" {
		parsed, err := models.ParseServiceTarget(req.ServiceTarget)
		if err != nil {
			h.respondError(w, "Invalid service_target: "+err.Error(), http.StatusBadRequest)
			return
		}
		serviceTarget = parsed
	}

	policy := &models.QuotaPolicy{
		UUID:               uuid.New().String(),
		CategoryUUID:       req.CategoryUUID,
		MinInstances:       minInstances,
		MaxInstances:       maxInstances,
		Priority:           priority,
		ServiceTarget:      serviceTarget,
		AutoReplenish:      autoReplenish,
		ReplenishThreshold: replenishThreshold,
		MaxLifetimeSeconds: maxLifetimeSeconds,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.service.SetQuotaPolicy(policy); err != nil {
		h.respondError(w, "Failed to create quota policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Quota policy created successfully",
		"policy":  policy.ToResponse(),
	})
}

// handleGetQuotaByCategory 处理按类别获取配额请求
func (h *AdminAPIHandler) handleGetQuotaByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryUUID := vars["categoryUUID"]

	policy, err := h.service.GetQuotaPolicy(categoryUUID)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"policy":  policy,
	})
}

// handleUpdateQuotaPolicy 处理更新配额策略请求
func (h *AdminAPIHandler) handleUpdateQuotaPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var req struct {
		MinInstances *int `json:"min_instances"`
		MaxInstances *int `json:"max_instances"`
		// 旧字段名，保持向后兼容
		MaxAllocations     *int    `json:"max_allocations"`
		MaxLifetimeSeconds *int    `json:"max_lifetime_seconds"`
		AutoReplenish      *bool   `json:"auto_replenish"`
		ReplenishThreshold *int    `json:"replenish_threshold"`
		Priority           *int    `json:"priority"`
		ServiceTarget      *string `json:"service_target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	existingPolicy, err := h.policyStorage.GetQuotaPolicyByUUID(uuid)
	if err != nil {
		h.respondError(w, "Quota policy not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// 处理最小实例数
	if req.MinInstances != nil {
		if *req.MinInstances < 0 {
			h.respondError(w, "min_instances must be greater than or equal to 0", http.StatusBadRequest)
			return
		}
		existingPolicy.MinInstances = *req.MinInstances
	}

	// 处理最大实例数（优先使用新字段名）
	if req.MaxInstances != nil {
		if *req.MaxInstances <= 0 {
			h.respondError(w, "max_instances must be greater than 0", http.StatusBadRequest)
			return
		}
		existingPolicy.MaxInstances = *req.MaxInstances
	} else if req.MaxAllocations != nil {
		if *req.MaxAllocations <= 0 {
			h.respondError(w, "max_allocations must be greater than 0", http.StatusBadRequest)
			return
		}
		existingPolicy.MaxInstances = *req.MaxAllocations
	}

	if req.MaxLifetimeSeconds != nil {
		if *req.MaxLifetimeSeconds <= 0 {
			h.respondError(w, "max_lifetime_seconds must be greater than 0", http.StatusBadRequest)
			return
		}
		existingPolicy.MaxLifetimeSeconds = *req.MaxLifetimeSeconds
	}

	if req.Priority != nil {
		existingPolicy.Priority = *req.Priority
	}

	if req.AutoReplenish != nil {
		existingPolicy.AutoReplenish = *req.AutoReplenish
	}

	if req.ReplenishThreshold != nil {
		if *req.ReplenishThreshold < 0 {
			h.respondError(w, "replenish_threshold must be greater than or equal to 0", http.StatusBadRequest)
			return
		}
		existingPolicy.ReplenishThreshold = *req.ReplenishThreshold
	}

	// 处理服务对象类型
	if req.ServiceTarget != nil {
		parsed, err := models.ParseServiceTarget(*req.ServiceTarget)
		if err != nil {
			h.respondError(w, "Invalid service_target: "+err.Error(), http.StatusBadRequest)
			return
		}
		existingPolicy.ServiceTarget = parsed
	}

	if existingPolicy.AutoReplenish && existingPolicy.ReplenishThreshold <= 0 {
		h.respondError(w, "replenish_threshold is required when auto_replenish is true", http.StatusBadRequest)
		return
	}

	existingPolicy.UpdatedAt = time.Now()

	if err := h.policyStorage.UpdateQuotaPolicy(existingPolicy); err != nil {
		h.respondError(w, "Failed to update quota policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Quota policy updated successfully",
		"policy":  existingPolicy.ToResponse(),
	})
}

// handleDeleteQuotaPolicy 处理删除配额策略请求
func (h *AdminAPIHandler) handleDeleteQuotaPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	if uuid == "" {
		h.respondError(w, "uuid is required", http.StatusBadRequest)
		return
	}

	// 通过 UUID 获取策略
	policy, err := h.policyStorage.GetQuotaPolicyByUUID(uuid)
	if err != nil {
		h.respondError(w, "Quota policy not found", http.StatusNotFound)
		return
	}

	// 检查是否有关联的 testbed
	count, err := h.testbedStorage.CountTestbedsByCategoryAndServiceTarget(policy.CategoryUUID, policy.ServiceTarget)
	if err != nil {
		h.respondError(w, "Failed to check associated testbeds: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		h.respondError(w, fmt.Sprintf("无法删除配额策略：存在 %d 个关联的 testbed 记录（类别: %s, 服务对象: %s）。请先删除这些 testbed。",
			count, policy.CategoryUUID, policy.ServiceTarget), http.StatusConflict)
		return
	}

	// 删除策略
	if err := h.policyStorage.DeleteQuotaPolicy(policy.ID); err != nil {
		h.respondError(w, "Failed to delete quota policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Quota policy deleted successfully",
	})
}

// handleListAllAllocations 处理列出所有分配请求
func (h *AdminAPIHandler) handleListAllAllocations(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现列表逻辑
	h.respondJSON(w, map[string]interface{}{
		"success":     true,
		"allocations": []interface{}{},
	})
}

// handleListAllocationHistory 处理列出分配历史请求
func (h *AdminAPIHandler) handleListAllocationHistory(w http.ResponseWriter, r *http.Request) {
	// TODO: 实现历史逻辑
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"history": []interface{}{},
	})
}

// handleGetMetrics 处理获取指标请求
func (h *AdminAPIHandler) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// 获取所有类别
	categories, err := h.categoryStorage.ListCategories()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 统计 Testbed 总数（不包含 deleted）
	totalTestbeds := 0
	availableTestbeds := 0
	allocatedTestbeds := 0
	for _, cat := range categories {
		total, err := h.testbedStorage.CountTestbedsByCategory(cat.UUID)
		if err == nil {
			totalTestbeds += total
		}
		allocated, err := h.testbedStorage.CountAllocatedTestbedsByCategory(cat.UUID)
		if err == nil {
			allocatedTestbeds += allocated
		}
	}

	// 统计所有可用 testbed
	availableTestbeds, _ = h.testbedStorage.CountAllAvailableTestbeds()

	// 统计活跃分配数
	activeAllocations, err := h.allocationStorage.ListActiveAllocations()
	if err != nil {
		activeAllocations = nil
	}
	activeAllocCount := 0
	if activeAllocations != nil {
		activeAllocCount = len(activeAllocations)
	}

	// 统计活跃用户数（有过分配的用户）
	activeUsers := make(map[string]bool)
	if activeAllocations != nil {
		for _, alloc := range activeAllocations {
			if alloc.Requester != "" {
				activeUsers[alloc.Requester] = true
			}
		}
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"metrics": map[string]interface{}{
			"total_testbeds":     totalTestbeds,
			"available_testbeds": availableTestbeds,
			"allocated_testbeds": allocatedTestbeds,
			"active_allocations": activeAllocCount,
			"total_users":        len(activeUsers),
		},
	})
}

// handleGetUsageStats 处理获取使用统计请求
func (h *AdminAPIHandler) handleGetUsageStats(w http.ResponseWriter, r *http.Request) {
	// 获取所有类别
	categories, err := h.categoryStorage.ListCategories()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建类别分布数据
	type CategoryDist struct {
		UUID      string `json:"uuid"`
		Name      string `json:"name"`
		Total     int    `json:"total"`
		Available int    `json:"available"`
		Allocated int    `json:"allocated"`
	}
	categoryDist := make([]CategoryDist, 0)
	for _, cat := range categories {
		total, _ := h.testbedStorage.CountTestbedsByCategory(cat.UUID)
		allocated, _ := h.testbedStorage.CountAllocatedTestbedsByCategory(cat.UUID)
		available := total - allocated
		if available < 0 {
			available = 0
		}
		categoryDist = append(categoryDist, CategoryDist{
			UUID:      cat.UUID,
			Name:      cat.Name,
			Total:     total,
			Available: available,
			Allocated: allocated,
		})
	}

	// 获取所有活跃分配
	activeAllocations, _ := h.allocationStorage.ListActiveAllocations()

	// 统计用户使用排行
	type UserStat struct {
		Username             string `json:"username"`
		CurrentAllocations   int    `json:"current_allocations"`
		TotalAllocations     int    `json:"total_allocations"`
		TotalDurationSeconds int64  `json:"total_duration_seconds"`
		LastUsedAt           string `json:"last_used_at"`
	}

	// 获取所有分配记录来统计用户使用情况
	allAllocations, _ := h.allocationStorage.ListAllocations()
	userStatsMap := make(map[string]*UserStat)

	for _, alloc := range allAllocations {
		username := alloc.Requester
		if username == "" {
			continue
		}

		if _, ok := userStatsMap[username]; !ok {
			userStatsMap[username] = &UserStat{
				Username: username,
			}
		}

		userStatsMap[username].TotalAllocations++

		// 计算使用时长
		if alloc.ExpiresAt != nil {
			duration := alloc.ExpiresAt.Sub(alloc.CreatedAt).Seconds()
			userStatsMap[username].TotalDurationSeconds += int64(duration)
		}

		// 更新最后使用时间
		if !alloc.UpdatedAt.IsZero() {
			userStatsMap[username].LastUsedAt = alloc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
		}
	}

	// 设置当前分配数
	if activeAllocations != nil {
		for _, alloc := range activeAllocations {
			username := alloc.Requester
			if username != "" && userStatsMap[username] != nil {
				userStatsMap[username].CurrentAllocations++
			}
		}
	}

	// 转换为 slice 并排序（按总分配数降序）
	userStats := make([]UserStat, 0, len(userStatsMap))
	for _, us := range userStatsMap {
		userStats = append(userStats, *us)
	}

	// 简单排序：按总分配数降序
	for i := 0; i < len(userStats)-1; i++ {
		for j := i + 1; j < len(userStats); j++ {
			if userStats[j].TotalAllocations > userStats[i].TotalAllocations {
				userStats[i], userStats[j] = userStats[j], userStats[i]
			}
		}
	}

	// 只取前10名
	if len(userStats) > 10 {
		userStats = userStats[:10]
	}

	// 最近活动（从分配记录中获取）
	type RecentActivity struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		User   string `json:"user"`
		Action string `json:"action"`
		Time   string `json:"time"`
	}

	recentActivity := make([]RecentActivity, 0)
	if allAllocations != nil {
		// 取最近10条分配记录
		limit := len(allAllocations)
		if limit > 10 {
			limit = 10
		}
		for i := 0; i < limit; i++ {
			alloc := allAllocations[len(allAllocations)-1-i]
			action := "acquired"
			title := fmt.Sprintf("申请了 Testbed")
			if alloc.Status == "released" {
				action = "released"
				title = fmt.Sprintf("释放了 Testbed")
			}
			recentActivity = append(recentActivity, RecentActivity{
				ID:     alloc.ID,
				Title:  title,
				User:   alloc.Requester,
				Action: action,
				Time:   alloc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"categories":      categoryDist,
			"users":           userStats,
			"recent_activity": recentActivity,
		},
	})
}

// handleCleanupTestbeds 处理清理所有 Testbed
func (h *AdminAPIHandler) handleCleanupTestbeds(w http.ResponseWriter, r *http.Request) {
	if err := h.testbedStorage.DeleteAllTestbeds(); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All testbeds cleaned up successfully",
	})
}

// handleCleanupAllocations 处理清理所有 Allocation
func (h *AdminAPIHandler) handleCleanupAllocations(w http.ResponseWriter, r *http.Request) {
	if err := h.allocationStorage.DeleteAllAllocations(); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All allocations cleaned up successfully",
	})
}

// handleCleanupResourceInstances 处理清理所有 ResourceInstance
func (h *AdminAPIHandler) handleCleanupResourceInstances(w http.ResponseWriter, r *http.Request) {
	if err := h.resourceStorage.DeleteAllResourceInstances(); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All resource instances cleaned up successfully",
	})
}

// handleCleanupCategories 处理清理所有 Category
func (h *AdminAPIHandler) handleCleanupCategories(w http.ResponseWriter, r *http.Request) {
	if err := h.categoryStorage.DeleteAllCategories(); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All categories cleaned up successfully",
	})
}

// handleCleanupQuotaPolicies 处理清理所有 QuotaPolicy
func (h *AdminAPIHandler) handleCleanupQuotaPolicies(w http.ResponseWriter, r *http.Request) {
	if err := h.policyStorage.DeleteAllQuotaPolicies(); err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All quota policies cleaned up successfully",
	})
}

// handleCleanupAll 处理清理所有 Resource Pool 数据
func (h *AdminAPIHandler) handleCleanupAll(w http.ResponseWriter, r *http.Request) {
	var errors []string

	if err := h.testbedStorage.DeleteAllTestbeds(); err != nil {
		errors = append(errors, "Failed to delete testbeds: "+err.Error())
	}
	if err := h.allocationStorage.DeleteAllAllocations(); err != nil {
		errors = append(errors, "Failed to delete allocations: "+err.Error())
	}
	if err := h.resourceStorage.DeleteAllResourceInstances(); err != nil {
		errors = append(errors, "Failed to delete resource instances: "+err.Error())
	}
	if err := h.categoryStorage.DeleteAllCategories(); err != nil {
		errors = append(errors, "Failed to delete categories: "+err.Error())
	}
	if err := h.policyStorage.DeleteAllQuotaPolicies(); err != nil {
		errors = append(errors, "Failed to delete quota policies: "+err.Error())
	}

	if len(errors) > 0 {
		h.respondJSON(w, map[string]interface{}{
			"success": false,
			"errors":  errors,
		})
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "All Resource Pool data cleaned up successfully",
	})
}

// Task Management Handlers

// handleListTasks 处理列出所有任务请求
func (h *AdminAPIHandler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}

	pageSize := 20
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}

	// 获取最近的任务
	tasks, err := h.service.ListRecentTasks(pageSize * page)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 分页
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize

	if startIdx >= len(tasks) {
		h.respondJSON(w, map[string]interface{}{
			"success": true,
			"tasks":   []interface{}{},
			"total":   len(tasks),
			"page":    page,
		})
		return
	}

	if endIdx > len(tasks) {
		endIdx = len(tasks)
	}

	pagedTasks := tasks[startIdx:endIdx]
	responses := make([]models.ResourceInstanceTaskResponse, len(pagedTasks))
	for i, t := range pagedTasks {
		responses[i] = t.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"tasks":   responses,
		"total":   len(tasks),
		"page":    page,
	})
}

// handleGetTask 处理获取任务详情请求
func (h *AdminAPIHandler) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	task, err := h.service.GetTask(uuid)
	if err != nil {
		h.respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"task":    task.ToResponse(),
	})
}

// handleGetTaskStatistics 处理获取任务统计请求
func (h *AdminAPIHandler) handleGetTaskStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetTaskStatistics()
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// handleListFailedTasks 处理列出失败任务请求
func (h *AdminAPIHandler) handleListFailedTasks(w http.ResponseWriter, r *http.Request) {
	// 默认列出最近7天的失败任务
	since := time.Now().Add(-7 * 24 * time.Hour)

	sinceStr := r.URL.Query().Get("since")
	if sinceStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, sinceStr)
		if err == nil {
			since = parsedTime
		}
	}

	if h.taskStorage == nil {
		h.respondError(w, "Task storage not available", http.StatusServiceUnavailable)
		return
	}

	tasks, err := h.taskStorage.ListFailedTasks(since)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.ResourceInstanceTaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = t.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"tasks":   responses,
		"total":   len(responses),
	})
}

// handleCleanupOldTasks 处理清理旧任务请求
func (h *AdminAPIHandler) handleCleanupOldTasks(w http.ResponseWriter, r *http.Request) {
	// 默认清理30天前的已完成/失败/取消任务
	days := 30
	daysStr := r.URL.Query().Get("days")
	if daysStr != "" {
		fmt.Sscanf(daysStr, "%d", &days)
	}

	if h.taskStorage == nil {
		h.respondError(w, "Task storage not available", http.StatusServiceUnavailable)
		return
	}

	olderThan := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	count, err := h.taskStorage.DeleteOldTasks(olderThan)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Deleted %d old tasks (older than %d days)", count, days),
		"deleted": count,
	})
}

// handleListResourceInstanceTasks 处理列出资源实例任务请求
func (h *AdminAPIHandler) handleListResourceInstanceTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	// 获取分页参数
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}

	pageSize := 20
	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}

	tasks, total, err := h.service.ListTasksByResourceInstance(uuid, page, pageSize)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]models.ResourceInstanceTaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = t.ToResponse()
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"tasks":   responses,
		"total":   total,
		"page":    page,
	})
}

// ResourceInstance Request/Response DTOs

type CreateResourceInstanceRequest struct {
	Name                 string  `json:"name"`
	ResourceType         string  `json:"resource_type"` // virtual_machine, physical_machine
	Host                 string  `json:"host"`
	SSHPort              int     `json:"ssh_port"`
	SSHUser              string  `json:"ssh_user"` // SSH 用户名，默认 root
	SnapshotID           string  `json:"snapshot_id"`
	SnapshotInstanceUUID string  `json:"snapshot_instance_uuid"`
	Password             string  `json:"passwd"`
	Description          *string `json:"description"`
	IsPublic             *bool   `json:"is_public"`
	CreatedBy            string  `json:"created_by"`
}

type UpdateResourceInstanceRequest struct {
	Name                 string  `json:"name"`
	Host                 string  `json:"host"`
	SSHPort              int     `json:"ssh_port"`
	SSHUser              string  `json:"ssh_user"`
	SnapshotID           *string `json:"snapshot_id"`
	SnapshotInstanceUUID *string `json:"snapshot_instance_uuid"`
	Password             string  `json:"passwd"`
	Description          *string `json:"description"`
	IsPublic             *bool   `json:"is_public"`
	Status               string  `json:"status"`
}

// ParseResourceInstanceStatus 解析资源实例状态字符串
func ParseResourceInstanceStatus(s string) (models.ResourceInstanceStatus, error) {
	switch s {
	case "available", "active":
		return models.ResourceInstanceStatusActive, nil
	case "unreachable", "terminating":
		return models.ResourceInstanceStatusUnreachable, nil
	case "pending":
		return models.ResourceInstanceStatusPending, nil
	default:
		return "", fmt.Errorf("invalid resource instance status: %s", s)
	}
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// generateUUID 生成新的 UUID
func generateUUID() string {
	return uuid.New().String()
}

// Azure Deployment Task Handlers

// handleListDeploymentTasks 列出部署任务
func (h *AdminAPIHandler) handleListDeploymentTasks(w http.ResponseWriter, r *http.Request) {
	if h.deploymentService == nil {
		h.respondError(w, "Deployment service not available", http.StatusServiceUnavailable)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	tasks, err := h.deploymentService.ListRecentTasks(limit)
	if err != nil {
		h.respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]DeploymentTaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = toDeploymentTaskResponse(t)
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"tasks":   responses,
		"total":   len(tasks),
	})
}

// handleGetDeploymentTask 获取部署任务详情
func (h *AdminAPIHandler) handleGetDeploymentTask(w http.ResponseWriter, r *http.Request) {
	if h.deploymentService == nil {
		h.respondError(w, "Deployment service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	taskUUID := vars["uuid"]

	if taskUUID == "" {
		h.respondError(w, "task_uuid is required", http.StatusBadRequest)
		return
	}

	task, err := h.deploymentService.GetTask(taskUUID)
	if err != nil {
		h.respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"task":    toDeploymentTaskResponse(task),
	})
}

// handleGetDeploymentTaskLogs 获取部署任务日志
func (h *AdminAPIHandler) handleGetDeploymentTaskLogs(w http.ResponseWriter, r *http.Request) {
	if h.deploymentService == nil {
		h.respondError(w, "Deployment service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	taskUUID := vars["uuid"]

	if taskUUID == "" {
		h.respondError(w, "task_uuid is required", http.StatusBadRequest)
		return
	}

	task, err := h.deploymentService.GetTask(taskUUID)
	if err != nil {
		h.respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	if task.LogDirectory == "" {
		h.respondError(w, "No logs available for this task", http.StatusNotFound)
		return
	}

	// TODO: 读取日志文件内容并返回
	// 这里简化实现，只返回日志目录路径
	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"logs":    fmt.Sprintf("Logs available at: %s", task.LogDirectory),
	})
}

// handleRetryDeploymentTask 重试部署任务
func (h *AdminAPIHandler) handleRetryDeploymentTask(w http.ResponseWriter, r *http.Request) {
	if h.deploymentService == nil {
		h.respondError(w, "Deployment service not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	taskUUID := vars["uuid"]

	if taskUUID == "" {
		h.respondError(w, "task_uuid is required", http.StatusBadRequest)
		return
	}

	// TODO: 实现重试逻辑
	h.respondJSON(w, map[string]interface{}{
		"success": false,
		"message": "Retry not implemented yet",
	})
}

// DeploymentTaskResponse 部署任务响应
type DeploymentTaskResponse struct {
	ID            int                    `json:"id"`
	TaskUUID      string                 `json:"task_uuid"`
	AllocationID  int                    `json:"allocation_id"`
	PipelineID    int                    `json:"pipeline_id"`
	BuildID       int                    `json:"build_id"`
	Status        string                 `json:"status"`
	Analyzing     bool                   `json:"analyzing"`
	LogDirectory  string                 `json:"log_directory"`
	ResultDetails map[string]interface{} `json:"result_details"`
	ErrorMessage  string                 `json:"error_message"`
	WebURL        string                 `json:"web_url"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// toDeploymentTaskResponse 转换为响应格式
func toDeploymentTaskResponse(task *storage.DeploymentTask) DeploymentTaskResponse {
	return DeploymentTaskResponse{
		ID:            task.ID,
		TaskUUID:      task.TaskUUID,
		AllocationID:  task.AllocationID,
		PipelineID:    task.PipelineID,
		BuildID:       task.BuildID,
		Status:        string(task.Status),
		Analyzing:     task.Analyzing,
		LogDirectory:  task.LogDirectory,
		ResultDetails: task.ResultDetails,
		ErrorMessage:  task.ErrorMessage,
		WebURL:        task.WebURL,
		CreatedAt:     task.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     task.UpdatedAt.Format(time.RFC3339),
	}
}

// ============ 部署管道模板管理 ============

// PipelineTemplateRequest 创建/更新部署管道模板请求
type PipelineTemplateRequest struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Organization       string                 `json:"organization"`
	Project            string                 `json:"project"`
	PipelineID         int                    `json:"pipeline_id"`
	PipelineParameters map[string]interface{} `json:"pipeline_parameters"`
	Enabled            bool                   `json:"enabled"`
}

// PipelineTemplateResponse 部署管道模板响应
type PipelineTemplateResponse struct {
	ID                 int                    `json:"id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Organization       string                 `json:"organization"`
	Project            string                 `json:"project"`
	PipelineID         int                    `json:"pipeline_id"`
	PipelineParameters map[string]interface{} `json:"pipeline_parameters"`
	Enabled            bool                   `json:"enabled"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
	CreatedBy          string                 `json:"created_by"`
}

// handleListPipelineTemplates 列出所有部署管道模板
func (h *AdminAPIHandler) handleListPipelineTemplates(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	templates, err := h.pipelineTemplateStorage.ListTemplates()
	if err != nil {
		h.respondError(w, fmt.Sprintf("Failed to list pipeline templates: %v", err), http.StatusInternalServerError)
		return
	}

	responses := make([]PipelineTemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = toPipelineTemplateResponse(t)
	}

	h.respondJSON(w, map[string]interface{}{
		"success":   true,
		"templates": responses,
	})
}

// handleCreatePipelineTemplate 创建部署管道模板
func (h *AdminAPIHandler) handleCreatePipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	var req PipelineTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Organization == "" || req.Project == "" || req.PipelineID == 0 {
		h.respondError(w, "name, organization, project and pipeline_id are required", http.StatusBadRequest)
		return
	}

	// 获取当前用户
	username := h.getUsernameFromContext(r)
	if username == "" {
		username = "admin"
	}

	template := &storage.DeploymentPipelineTemplate{
		Name:               req.Name,
		Description:        req.Description,
		Organization:       req.Organization,
		Project:            req.Project,
		PipelineID:         req.PipelineID,
		PipelineParameters: req.PipelineParameters,
		Enabled:            req.Enabled,
		CreatedBy:          username,
	}

	if err := h.pipelineTemplateStorage.CreateTemplate(template); err != nil {
		h.respondError(w, fmt.Sprintf("Failed to create pipeline template: %v", err), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"template": toPipelineTemplateResponse(template),
	})
}

// handleGetPipelineTemplate 获取部署管道模板详情
func (h *AdminAPIHandler) handleGetPipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	template, err := h.pipelineTemplateStorage.GetTemplate(id)
	if err != nil {
		h.respondError(w, "Template not found", http.StatusNotFound)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"template": toPipelineTemplateResponse(template),
	})
}

// handleUpdatePipelineTemplate 更新部署管道模板
func (h *AdminAPIHandler) handleUpdatePipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	var req PipelineTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.Name == "" || req.Organization == "" || req.Project == "" || req.PipelineID == 0 {
		h.respondError(w, "name, organization, project and pipeline_id are required", http.StatusBadRequest)
		return
	}

	// 获取现有模板
	template, err := h.pipelineTemplateStorage.GetTemplate(id)
	if err != nil {
		h.respondError(w, "Template not found", http.StatusNotFound)
		return
	}

	// 更新字段
	template.Name = req.Name
	template.Description = req.Description
	template.Organization = req.Organization
	template.Project = req.Project
	template.PipelineID = req.PipelineID
	template.PipelineParameters = req.PipelineParameters
	template.Enabled = req.Enabled

	if err := h.pipelineTemplateStorage.UpdateTemplate(template); err != nil {
		h.respondError(w, fmt.Sprintf("Failed to update pipeline template: %v", err), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success":  true,
		"template": toPipelineTemplateResponse(template),
	})
}

// handleDeletePipelineTemplate 删除部署管道模板
func (h *AdminAPIHandler) handleDeletePipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.pipelineTemplateStorage.DeleteTemplate(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.respondError(w, "Template not found", http.StatusNotFound)
			return
		}
		h.respondError(w, fmt.Sprintf("Failed to delete pipeline template: %v", err), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Pipeline template deleted successfully",
	})
}

// handleEnablePipelineTemplate 启用部署管道模板
func (h *AdminAPIHandler) handleEnablePipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.pipelineTemplateStorage.EnableTemplate(id); err != nil {
		h.respondError(w, fmt.Sprintf("Failed to enable pipeline template: %v", err), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Pipeline template enabled successfully",
	})
}

// handleDisablePipelineTemplate 禁用部署管道模板
func (h *AdminAPIHandler) handleDisablePipelineTemplate(w http.ResponseWriter, r *http.Request) {
	if h.pipelineTemplateStorage == nil {
		h.respondError(w, "Pipeline template storage not available", http.StatusServiceUnavailable)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.respondError(w, "Invalid template ID", http.StatusBadRequest)
		return
	}

	if err := h.pipelineTemplateStorage.DisableTemplate(id); err != nil {
		h.respondError(w, fmt.Sprintf("Failed to disable pipeline template: %v", err), http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Pipeline template disabled successfully",
	})
}

// toPipelineTemplateResponse 转换为响应格式
func toPipelineTemplateResponse(template *storage.DeploymentPipelineTemplate) PipelineTemplateResponse {
	return PipelineTemplateResponse{
		ID:                 template.ID,
		Name:               template.Name,
		Description:        template.Description,
		Organization:       template.Organization,
		Project:            template.Project,
		PipelineID:         template.PipelineID,
		PipelineParameters: template.PipelineParameters,
		Enabled:            template.Enabled,
		CreatedAt:          template.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          template.UpdatedAt.Format(time.RFC3339),
		CreatedBy:          template.CreatedBy,
	}
}
