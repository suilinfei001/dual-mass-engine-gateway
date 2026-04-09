package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/mux"
)

type ConsoleHandler struct {
	configStorage    *storage.MySQLConfigStorage
	userStorage      *storage.MySQLUserStorage
	taskStorage      Storage
	mysqlTaskStorage *storage.MySQLTaskStorage
	resourceStorage  storage.ResourceStorage
	aiMatcher        *ai.AIMatcher
	logAnalyzer      LogAnalyzerService
	server           *Server
	client           *Client
}

func NewConsoleHandler(
	configStorage *storage.MySQLConfigStorage,
	userStorage *storage.MySQLUserStorage,
	taskStorage Storage,
	resourceStorage storage.ResourceStorage,
	server *Server,
	client *Client,
	logAnalyzer LogAnalyzerService,
) *ConsoleHandler {
	return &ConsoleHandler{
		configStorage:    configStorage,
		userStorage:      userStorage,
		taskStorage:      taskStorage,
		resourceStorage:  resourceStorage,
		aiMatcher:        ai.NewAIMatcher(configStorage),
		logAnalyzer:      logAnalyzer,
		server:           server,
		client:           client,
	}
}

func (h *ConsoleHandler) SetMySQLTaskStorage(storage *storage.MySQLTaskStorage) {
	h.mysqlTaskStorage = storage
}

func (h *ConsoleHandler) RegisterRoutes(router *mux.Router, authMiddleware func(http.Handler) http.Handler, adminMiddleware func(http.Handler) http.Handler) {
	router.Handle("/api/admin/config", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleGetAllConfigs)))).Methods("GET")
	router.Handle("/api/admin/config/ai", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateAIConfig)))).Methods("PUT")
	router.Handle("/api/admin/config/ai/test", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleTestAIConfig)))).Methods("POST")
	router.Handle("/api/admin/config/azure-pat", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateAzurePAT)))).Methods("PUT")
	router.Handle("/api/admin/config/event-receiver", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateEventReceiverIP)))).Methods("PUT")
	router.Handle("/api/admin/config/event-receiver/test", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleTestEventReceiverConnection)))).Methods("POST")
	router.Handle("/api/admin/config/log-retention", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateLogRetention)))).Methods("PUT")
	router.Handle("/api/admin/config/ai-concurrency", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateAIConcurrency)))).Methods("PUT")
	router.Handle("/api/admin/config/ai-request-pool-size", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateAIRequestPoolSize)))).Methods("PUT")
	router.Handle("/api/admin/config/cmp", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateCMPConfig)))).Methods("PUT")
	router.Handle("/api/admin/cleanup", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupAll)))).Methods("POST")
	router.Handle("/api/admin/cleanup/tasks", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupTasks)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resources", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResources)))).Methods("POST")
	router.Handle("/api/admin/cleanup/users", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupUsers)))).Methods("POST")
	router.Handle("/api/admin/cleanup/task-results", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupTaskResults)))).Methods("POST")
	router.Handle("/api/admin/cleanup/sessions", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupSessions)))).Methods("POST")
	router.Handle("/api/admin/cleanup/all-sessions", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupAllSessions)))).Methods("POST")
	// Resource Pool 清理路由（代理到 resource-pool 服务）
	router.Handle("/api/admin/cleanup/resource-pool/testbeds", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolTestbeds)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resource-pool/allocations", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolAllocations)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resource-pool/resource-instances", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolInstances)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resource-pool/categories", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolCategories)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resource-pool/quota-policies", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolPolicies)))).Methods("POST")
	router.Handle("/api/admin/cleanup/resource-pool/all", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleCleanupResourcePoolAll)))).Methods("POST")
	router.Handle("/api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleSearchUsers)))).Methods("GET")
	router.Handle("/api/admin/users/paginated", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleGetUsersWithPagination)))).Methods("GET")
	router.Handle("/api/admin/users/{id}/password", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateUserPassword)))).Methods("PUT")
	// Retry AI matching for tasks with no_resource status
	router.Handle("/api/tasks/{id}/retry-ai", authMiddleware(http.HandlerFunc(h.handleRetryAIMatching))).Methods("POST")
	// Reanalyze task logs using AI
	router.Handle("/api/tasks/{id}/reanalyze", authMiddleware(http.HandlerFunc(h.handleReanalyzeLogs))).Methods("POST")
}

func (h *ConsoleHandler) handleGetAllConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.configStorage.GetAllConfigs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make(map[string]interface{})
	for _, c := range configs {
		result[c.ConfigKey] = map[string]interface{}{
			"value":       c.ConfigValue,
			"description": c.Description,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

func (h *ConsoleHandler) handleUpdateAIConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP    string `json:"ip"`
		Model string `json:"model"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.configStorage.SetAIConfig(&storage.AIConfig{
		IP:    req.IP,
		Model: req.Model,
		Token: req.Token,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "AI config updated successfully",
	})
}

func (h *ConsoleHandler) handleTestAIConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP    string `json:"ip"`
		Model string `json:"model"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.IP == "" || req.Model == "" || req.Token == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "AI IP, Model Name, and API Token are all required",
		})
		return
	}

	// Create HTTP client that skips TLS verification (for self-signed certificates)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	// Step 1: Get model list to find model_id
	listURL := fmt.Sprintf("https://%s/api/mf-model-manager/v1/llm/list?page=1&size=20&order=desc&rule=update_time&name=", req.IP)
	listReq, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to create list request: %v", err),
		})
		return
	}

	listReq.Header.Set("Authorization", "Bearer "+req.Token)

	listResp, err := client.Do(listReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to connect to AI server: %v", err),
		})
		return
	}
	defer listResp.Body.Close()

	if listResp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("AI server returned status %d", listResp.StatusCode),
		})
		return
	}

	var listRespData struct {
		Count int `json:"count"`
		Data  []struct {
			ModelID   string `json:"model_id"`
			ModelName string `json:"model_name"`
		} `json:"data"`
	}

	if err := json.NewDecoder(listResp.Body).Decode(&listRespData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to parse model list: %v", err),
		})
		return
	}

	// Find model_id by model_name
	var modelID string
	var foundModelName string
	for _, model := range listRespData.Data {
		if model.ModelName == req.Model {
			modelID = model.ModelID
			foundModelName = model.ModelName
			break
		}
	}

	if modelID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Model '%s' not found in AI server. Available models: %s", req.Model, getModelNames(listRespData.Data)),
		})
		return
	}

	// Step 2: Test the model
	testURL := fmt.Sprintf("https://%s/api/mf-model-manager/v1/llm/test", req.IP)
	testReqBody := map[string]interface{}{
		"model_id": modelID,
	}
	testBodyBytes, err := json.Marshal(testReqBody)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to marshal test request: %v", err),
		})
		return
	}

	testReq, err := http.NewRequest("POST", testURL, bytes.NewBuffer(testBodyBytes))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to create test request: %v", err),
		})
		return
	}

	testReq.Header.Set("Content-Type", "application/json")
	testReq.Header.Set("Authorization", "Bearer "+req.Token)

	testResp, err := client.Do(testReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to test model: %v", err),
		})
		return
	}
	defer testResp.Body.Close()

	if testResp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(testResp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("AI test failed with status %d: %s", testResp.StatusCode, string(bodyBytes)),
		})
		return
	}

	var testRespData struct {
		Status string `json:"status"`
		ID     string `json:"id"`
	}

	if err := json.NewDecoder(testResp.Body).Decode(&testRespData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to parse test response: %v", err),
		})
		return
	}

	if testRespData.Status != "ok" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("AI test failed with status: %s", testRespData.Status),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("AI connection successful! Model: %s (ID: %s)", foundModelName, modelID),
		"data": map[string]interface{}{
			"model_id":   modelID,
			"model_name": foundModelName,
			"test_id":    testRespData.ID,
		},
	})
}

// Helper function to get available model names
func getModelNames(models []struct {
	ModelID   string `json:"model_id"`
	ModelName string `json:"model_name"`
}) string {
	if len(models) == 0 {
		return "none"
	}
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.ModelName
	}
	return strings.Join(names, ", ")
}

func (h *ConsoleHandler) handleUpdateAzurePAT(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PAT string `json:"pat"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.configStorage.SetAzurePAT(req.PAT); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Azure PAT updated successfully",
	})
}

func (h *ConsoleHandler) handleUpdateEventReceiverIP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.configStorage.SetEventReceiverIP(req.IP); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.server != nil {
		h.server.UpdateClientURL(req.IP)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Event receiver IP updated successfully",
	})
}

func (h *ConsoleHandler) handleTestEventReceiverConnection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IP string `json:"ip"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate IP format
	if req.IP == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Event Receiver IP is required",
		})
		return
	}

	// Ensure URL has protocol
	testURL := req.IP
	if !strings.HasPrefix(testURL, "http://") && !strings.HasPrefix(testURL, "https://") {
		testURL = "http://" + testURL
	}

	// Test connection by calling the events API
	// Add a timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	eventsURL := fmt.Sprintf("%s/api/events", testURL)
	resp, err := client.Get(eventsURL)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Connection failed: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("API returned HTTP %d: %s", resp.StatusCode, string(body)),
		})
		return
	}

	// Parse response to get event count
	var result struct {
		Data []interface{} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		// Still return success if we got a 200 response, even if parsing failed
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Connection successful! Unable to parse event count.",
			"data": map[string]interface{}{
				"event_count": 0,
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Connection successful! Found %d event(s).", len(result.Data)),
		"data": map[string]interface{}{
			"event_count": len(result.Data),
		},
	})
}

func (h *ConsoleHandler) handleUpdateLogRetention(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Days int `json:"days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate range
	if req.Days < 1 || req.Days > 30 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Log retention days must be between 1 and 30",
		})
		return
	}

	if err := h.configStorage.SetLogRetentionDays(req.Days); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Log retention days updated successfully",
		"data": map[string]interface{}{
			"days": req.Days,
		},
	})
}

func (h *ConsoleHandler) handleUpdateAIConcurrency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Concurrency int `json:"concurrency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate range
	if req.Concurrency < 1 || req.Concurrency > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "AI concurrency must be between 1 and 50",
		})
		return
	}

	if err := h.configStorage.SetAIConcurrency(req.Concurrency); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("AI concurrency set to %d", req.Concurrency),
		"data": map[string]interface{}{
			"concurrency": req.Concurrency,
		},
	})
}

func (h *ConsoleHandler) handleUpdateAIRequestPoolSize(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PoolSize int `json:"poolSize"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate range
	if req.PoolSize < 1 || req.PoolSize > 200 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "AI request pool size must be between 1 and 200",
		})
		return
	}

	// Validate that pool size is greater than AI concurrency
	concurrency, _ := h.configStorage.GetAIConcurrency()
	if req.PoolSize <= concurrency {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("AI request pool size (%d) must be greater than AI concurrency (%d)", req.PoolSize, concurrency),
		})
		return
	}

	if err := h.configStorage.SetAIRequestPoolSize(req.PoolSize); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload the request pool config asynchronously to avoid blocking
	go func() {
		ai.ReloadGlobalRequestPool(h.configStorage)
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("AI request pool size set to %d", req.PoolSize),
		"data": map[string]interface{}{
			"poolSize": req.PoolSize,
		},
	})
}

func (h *ConsoleHandler) handleUpdateCMPConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Save CMP_ACCESS_KEY
	if req.AccessKey != "" {
		if err := h.configStorage.SetConfig(models.ConfigKeyCMPAccessKey, req.AccessKey); err != nil {
			http.Error(w, "Failed to save CMP access key: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Save CMP_SECRET_KEY
	if req.SecretKey != "" {
		if err := h.configStorage.SetConfig(models.ConfigKeyCMPSecretKey, req.SecretKey); err != nil {
			http.Error(w, "Failed to save CMP secret key: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "CMP config updated successfully",
	})
}

func (h *ConsoleHandler) handleCleanupAll(w http.ResponseWriter, r *http.Request) {
	var errors []string

	if err := h.taskStorage.DeleteAllTasks(); err != nil {
		errors = append(errors, "Failed to delete tasks: "+err.Error())
	}

	if err := h.resourceStorage.DeleteAllResources(); err != nil {
		errors = append(errors, "Failed to delete resources: "+err.Error())
	}

	if err := h.userStorage.DeleteAllUsers(); err != nil {
		errors = append(errors, "Failed to delete users: "+err.Error())
	}

	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"errors":  errors,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All data cleaned up successfully",
	})
}

func (h *ConsoleHandler) handleCleanupTasks(w http.ResponseWriter, r *http.Request) {
	if err := h.taskStorage.DeleteAllTasks(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All tasks cleaned up successfully",
	})
}

func (h *ConsoleHandler) handleCleanupResources(w http.ResponseWriter, r *http.Request) {
	if err := h.resourceStorage.DeleteAllResources(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All resources cleaned up successfully",
	})
}

func (h *ConsoleHandler) handleCleanupUsers(w http.ResponseWriter, r *http.Request) {
	if err := h.userStorage.DeleteAllUsers(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All non-admin users cleaned up successfully",
	})
}

func (h *ConsoleHandler) handleCleanupTaskResults(w http.ResponseWriter, r *http.Request) {
	if err := h.taskStorage.DeleteAllTasks(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All task results cleaned up successfully",
	})
}

func (h *ConsoleHandler) handleCleanupSessions(w http.ResponseWriter, r *http.Request) {
	count, err := h.userStorage.DeleteExpiredSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Cleaned up %d expired sessions", count),
		"data": map[string]interface{}{
			"count": count,
		},
	})
}

func (h *ConsoleHandler) handleCleanupAllSessions(w http.ResponseWriter, r *http.Request) {
	count, err := h.userStorage.DeleteAllSessions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Cleaned up all %d sessions", count),
		"data": map[string]interface{}{
			"count": count,
		},
	})
}

func (h *ConsoleHandler) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	users, total, err := h.userStorage.GetUsersWithPaginationAndKeyword(page, pageSize, keyword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    users,
		"total":   total,
	})
}

func (h *ConsoleHandler) handleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.userStorage.UpdatePassword(userID, string(hashedPassword)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

func (h *ConsoleHandler) handleGetUsersWithPagination(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	// 设置默认值
	page := 1
	pageSize := 10

	// 解析参数
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 获取用户列表
	users, total, err := h.userStorage.GetUsersWithPagination(page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"users":      users,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (h *ConsoleHandler) handleRetryAIMatching(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get the task from storage
	storageTask, err := h.taskStorage.GetTaskByID(taskID)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if storageTask == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Check if task is in no_resource status
	if storageTask.Status != string(models.TaskStatusNoResource) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Task is not in no_resource status",
		})
		return
	}

	// Get the event
	event, err := h.client.GetEvent(storageTask.EventID)
	if err != nil {
		http.Error(w, "Failed to get event", http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Get all resources
	resources, err := h.resourceStorage.GetAllResources()
	if err != nil {
		http.Error(w, "Failed to get resources", http.StatusInternalServerError)
		return
	}

	if len(resources) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No resources available for matching",
		})
		return
	}

	// Prepare AI match request - use task_name directly, not converted taskType
	eventDetail := map[string]interface{}{
		"event_type": event.EventType,
		"repository": event.Repository,
		"branch":     event.Branch,
		"task_name":  storageTask.TaskName,
		"commit_sha": event.CommitSHA,
		"pusher":     event.Pusher,
		"author":     event.Author,
		"payload":    event.Payload,
	}

	req := &ai.MatchRequest{
		TaskName:    storageTask.TaskName, // 直接使用 task_name，如 "basic_ci_all"
		EventDetail: eventDetail,
		Resources:   resources,
		SystemPrompt: h.aiMatcher.GetDefaultSystemPrompt(),
	}

	// Call AI matching
	matchResult, err := h.aiMatcher.MatchResource(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, ai.ErrAINotConfigured) {
			w.WriteHeader(http.StatusServiceUnavailable) // 503
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    "AI_NOT_CONFIGURED",
				"message": "AI service is not configured. Please contact the administrator to configure AI settings in the admin console.",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "AI matching failed: " + err.Error(),
			})
		}
		return
	}

	if matchResult == nil || matchResult.ResourceID == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "No resource matched for this task",
		})
		return
	}

	// Find the matched resource
	var matchedResource *models.ExecutableResource
	for _, res := range resources {
		if res.ID == matchResult.ResourceID {
			matchedResource = res
			break
		}
	}

	if matchedResource == nil {
		http.Error(w, "Matched resource not found", http.StatusInternalServerError)
		return
	}

	// Build request URL
	requestURL := h.buildRequestURL(matchedResource, event)

	// Update the task with new URL and status using Storage interface
	if err := h.taskStorage.UpdateTaskURLsAndStatus(storageTask.ID, requestURL, "pending"); err != nil {
		http.Error(w, "Failed to update task: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Task updated successfully with matched resource",
		"data": map[string]interface{}{
			"task_id":       taskID,
			"resource_id":   matchResult.ResourceID,
			"resource_name": matchedResource.ResourceName,
			"request_url":   requestURL,
		},
	})
}

func (h *ConsoleHandler) handleReanalyzeLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	log.Printf("[ConsoleHandler] Reanalyze request received for task_id: %d", taskID)

	// Check if mysqlTaskStorage is available
	if h.mysqlTaskStorage == nil {
		log.Printf("[ConsoleHandler] MySQLTaskStorage not available")
		http.Error(w, "Storage not available", http.StatusInternalServerError)
		return
	}

	// Get task from storage (returns *models.Task)
	task, err := h.mysqlTaskStorage.GetTask(taskID)
	if err != nil {
		log.Printf("[ConsoleHandler] Failed to get task: %v", err)
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Verify it's an Azure task with build_id
	if task.BuildID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Task is not an Azure task or has no build ID",
		})
		return
	}

	log.Printf("[ConsoleHandler] Task found: task_name=%s, build_id=%d", task.TaskName, task.BuildID)

	// Check if log analyzer is available
	if h.logAnalyzer == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Log analyzer service is not available",
		})
		return
	}

	// Set analyzing status to true before starting
	// Try to start AI analysis using CAS operation
	started, err := h.mysqlTaskStorage.TryStartAnalysis(taskID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start analysis: %v", err), http.StatusInternalServerError)
		return
	}
	if !started {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "AI analysis is already in progress for this task.",
		})
		return
	}

	log.Printf("[ConsoleHandler] Acquired analysis lock for task_id: %d", taskID)

	// Run AI analysis asynchronously to avoid timeout (processing 44 log files takes too long)
	go func() {
		defer func() {
			// Reset analyzing flag when done (to allow manual retry if needed)
			h.mysqlTaskStorage.UpdateTaskAnalyzing(taskID, false)
			log.Printf("[ConsoleHandler] Released analysis lock for task_id: %d", taskID)
		}()

		log.Printf("[ConsoleHandler] Starting AI analysis for task_id: %d", taskID)
		results, err := h.logAnalyzer.FetchAndAnalyzeLogs(context.Background(), task)
		if err != nil {
			// Check if it's "already in progress" error (shouldn't happen since we used CAS, but just in case)
			if err.Error() == "analysis already in progress" {
				log.Printf("[ConsoleHandler] AI analysis already in progress for task_id: %d", taskID)
				return
			}
			log.Printf("[ConsoleHandler] AI analysis failed: %v", err)
			return
		}
		log.Printf("[ConsoleHandler] AI analysis completed successfully, got %d results", len(results))
		// Save results to database
		if err := h.mysqlTaskStorage.SaveTaskResults(taskID, results); err != nil {
			log.Printf("[ConsoleHandler] Failed to save task results: %v", err)
		} else {
			log.Printf("[ConsoleHandler] Saved %d results for task_id: %d", len(results), taskID)
		}
	}()

	// Return immediately - analysis runs in background
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "AI analysis started in background. Refresh the page to see results.",
		"data": map[string]interface{}{
			"task_id":  taskID,
			"build_id": task.BuildID,
		},
	})
}

func (h *ConsoleHandler) buildRequestURL(resource *models.ExecutableResource, event *Event) string {
	requestURL := resource.RequestURL()

	// Azure URL 格式不需要追加参数，参数由 executor 服务处理
	if strings.HasPrefix(requestURL, "azure://") {
		return requestURL
	}

	// 非 Azure URL，追加参数（向后兼容）
	if resource.PipelineParams == nil || len(resource.PipelineParams) == 0 {
		return fmt.Sprintf("%s?event_id=%d", requestURL, event.ID)
	}
	params := make([]string, 0)
	params = append(params, fmt.Sprintf("event_id=%d", event.ID))
	for k, v := range resource.PipelineParams {
		params = append(params, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s?%s", requestURL, strings.Join(params, "&"))
}

func (h *ConsoleHandler) buildCancelURL(resource *models.ExecutableResource, event *Event) string {
	return fmt.Sprintf("%s/cancel?event_id=%d", resource.RequestURL(), event.ID)
}

func CreateAuthMiddleware(userStorage *storage.MySQLUserStorage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := userStorage.GetSessionWithUser(cookie.Value)
			if err != nil || session == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userData := session["user"].(map[string]interface{})
			ctx := r.Context()
			ctx = context.WithValue(ctx, userIDKey, userData["id"])
			ctx = context.WithValue(ctx, userRoleKey, userData["role"])
			ctx = context.WithValue(ctx, usernameKey, userData["username"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CreateAdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(userRoleKey)
			if role == nil || role != "admin" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Admin access required",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
	usernameKey contextKey = "username"
)

func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID := r.Context().Value(contextKey("user_id"))
	if userID == nil {
		return 0, false
	}

	switch v := userID.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func GetUserRoleFromContext(r *http.Request) (string, bool) {
	role := r.Context().Value(contextKey("user_role"))
	if role == nil {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}

func GetUsernameFromContext(r *http.Request) (string, bool) {
	username := r.Context().Value(contextKey("username"))
	if username == nil {
		return "", false
	}

	usernameStr, ok := username.(string)
	return usernameStr, ok
}

// Resource Pool 清理处理器（代理到 resource-pool 服务）

// getResourcePoolBaseURL 获取 Resource Pool 服务的基础 URL
func (h *ConsoleHandler) getResourcePoolBaseURL() string {
	// 从配置中获取 resource-pool 服务地址
	// 默认使用 Docker 网络内的服务地址
	url, _ := h.configStorage.GetEventReceiverIP()
	// 如果配置了 event-receiver IP，尝试使用相同模式的 resource-pool 地址
	// 否则使用默认的 Docker 网络地址
	if url != "" && strings.Contains(url, "10.4.174.125") {
		return "http://resource-pool-server:5003"
	}
	return "http://resource-pool-server:5003"
}

// handleCleanupResourcePoolTestbeds 清理所有 Testbed
func (h *ConsoleHandler) handleCleanupResourcePoolTestbeds(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/testbeds", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCleanupResourcePoolAllocations 清理所有 Allocation
func (h *ConsoleHandler) handleCleanupResourcePoolAllocations(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/allocations", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCleanupResourcePoolInstances 清理所有 ResourceInstance
func (h *ConsoleHandler) handleCleanupResourcePoolInstances(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/resource-instances", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCleanupResourcePoolCategories 清理所有 Category
func (h *ConsoleHandler) handleCleanupResourcePoolCategories(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/categories", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCleanupResourcePoolPolicies 清理所有 QuotaPolicy
func (h *ConsoleHandler) handleCleanupResourcePoolPolicies(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/quota-policies", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCleanupResourcePoolAll 清理所有 Resource Pool 数据
func (h *ConsoleHandler) handleCleanupResourcePoolAll(w http.ResponseWriter, r *http.Request) {
	baseURL := h.getResourcePoolBaseURL()
	resp, err := http.Post(baseURL+"/admin/cleanup/all", "application/json", nil)
	if err != nil {
		http.Error(w, "Failed to connect to resource-pool service: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
