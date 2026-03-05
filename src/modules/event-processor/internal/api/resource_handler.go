package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
	"github.com/gorilla/mux"
)

type ResourceHandler struct {
	resourceStorage storage.ResourceStorage
	userStorage     *storage.MySQLUserStorage
}

func NewResourceHandler(resourceStorage storage.ResourceStorage, userStorage *storage.MySQLUserStorage) *ResourceHandler {
	return &ResourceHandler{
		resourceStorage: resourceStorage,
		userStorage:     userStorage,
	}
}

func (h *ResourceHandler) RegisterRoutes(router *mux.Router, authMiddleware func(http.Handler) http.Handler) {
	router.HandleFunc("/api/resources", h.handleGetAllResources).Methods("GET")
	router.Handle("/api/resources/my", authMiddleware(http.HandlerFunc(h.handleGetMyResources))).Methods("GET")
	router.HandleFunc("/api/resources/{id}", h.handleGetResource).Methods("GET")
	router.Handle("/api/resources", authMiddleware(http.HandlerFunc(h.handleCreateResource))).Methods("POST")
	router.Handle("/api/resources/{id}", authMiddleware(http.HandlerFunc(h.handleUpdateResource))).Methods("PUT")
	router.Handle("/api/resources/{id}", authMiddleware(http.HandlerFunc(h.handleDeleteResource))).Methods("DELETE")
}

func (h *ResourceHandler) handleGetAllResources(w http.ResponseWriter, r *http.Request) {
	resources, err := h.resourceStorage.GetAllResources()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resources,
	})
}

func (h *ResourceHandler) handleGetMyResources(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resources, err := h.resourceStorage.GetResourcesByCreator(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resources,
	})
}

func (h *ResourceHandler) handleGetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	resource, err := h.resourceStorage.GetResource(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Resource not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resource,
	})
}

func (h *ResourceHandler) handleCreateResource(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userStorage.GetUser(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if user.IsAdmin() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Admin cannot create resources",
		})
		return
	}

	var req models.ExecutableResourceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !models.IsValidResourceType(req.ResourceType) {
		http.Error(w, "Invalid resource type", http.StatusBadRequest)
		return
	}

	// In skip mode, only validate the basic fields
	// In normal mode, validate Azure fields as well
	if !req.AllowSkip {
		if req.PipelineID <= 0 {
			http.Error(w, "Pipeline ID is required", http.StatusBadRequest)
			return
		}
	}

	if req.RepoPath == "" {
		http.Error(w, "Repo path is required", http.StatusBadRequest)
		return
	}

	// Auto-generate resource_name if not provided
	if req.ResourceName == "" {
		req.ResourceName = fmt.Sprintf("%s_%d", req.ResourceType, time.Now().Unix())
	}

	resource := models.NewExecutableResource(&req, userID, user.Username)

	if err := h.resourceStorage.CreateResource(resource); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resource,
	})
}

func (h *ResourceHandler) handleUpdateResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userStorage.GetUser(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	resource, err := h.resourceStorage.GetResource(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Resource not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin() && resource.CreatorID != userID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "You can only edit your own resources",
		})
		return
	}

	var req models.ExecutableResourceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields from request
	resource.AllowSkip = req.AllowSkip
	if req.Organization != "" {
		resource.Organization = req.Organization
	}
	if req.Project != "" {
		resource.Project = req.Project
	}
	if req.PipelineID > 0 {
		resource.PipelineID = req.PipelineID
	}
	if req.PipelineParams != nil {
		resource.PipelineParams = req.PipelineParams
	}
	if req.MicroserviceName != "" {
		resource.MicroserviceName = req.MicroserviceName
	}
	if req.PodName != "" {
		resource.PodName = req.PodName
	}
	if req.RepoPath != "" {
		resource.RepoPath = req.RepoPath
	}
	if req.Description != "" {
		resource.Description = req.Description
	}

	if err := h.resourceStorage.UpdateResource(resource); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    resource,
	})
}

func (h *ResourceHandler) handleDeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userStorage.GetUser(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	resource, err := h.resourceStorage.GetResource(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Resource not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !user.IsAdmin() && resource.CreatorID != userID {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "You can only delete your own resources",
		})
		return
	}

	if err := h.resourceStorage.DeleteResource(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Resource deleted successfully",
	})
}

func getUserIDFromRequest(r *http.Request) (int, error) {
	userIDStr := r.Context().Value(userIDKey)
	if userIDStr == nil {
		return 0, ErrUnauthorized
	}

	switch v := userIDStr.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			return 0, ErrUnauthorized
		}
		return id, nil
	default:
		return 0, ErrUnauthorized
	}
}

var ErrUnauthorized = &UnauthorizedError{}

type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string {
	return "unauthorized"
}
