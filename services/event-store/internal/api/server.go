// Package api provides HTTP handlers for event store service.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/quality-gateway/event-store/internal/service"
	"github.com/quality-gateway/event-store/internal/storage"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
	"github.com/quality-gateway/shared/pkg/logger"
)

// Server 事件存储服务器
type Server struct {
	config    sharedapi.Config
	service   *service.EventStoreService
	logger    *logger.Logger
	apiServer *sharedapi.Server
}

// NewServer 创建事件存储服务器
func NewServer(cfg sharedapi.Config, svc *service.EventStoreService, log *logger.Logger) *Server {
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
	router := s.apiServer.Router()

	// 健康检查
	router.GET("/health", s.healthCheck)

	// 事件相关
	router.POST("/api/events", s.createEvent)
	router.GET("/api/events", s.listEvents)
	router.GET("/api/events/{uuid}", s.getEvent)
	router.PUT("/api/events/{uuid}/status", s.updateEventStatus)
	router.GET("/api/events/statistics", s.getEventStatistics)
	router.GET("/api/events/pending", s.getPendingEvents)
	router.GET("/api/events/processing", s.getProcessingEvents)

	// PR 相关
	router.GET("/api/repos/{repo_name}/pulls/{pr_number}/events", s.getEventsByPR)

	// 质量检查相关
	router.GET("/api/events/{uuid}/quality-checks", s.getQualityChecks)
	router.POST("/api/events/{uuid}/quality-checks", s.createQualityCheck)
	router.PUT("/api/quality-checks/{id}", s.updateQualityCheck)
	router.PUT("/api/events/{uuid}/quality-checks/{type}", s.updateQualityCheckByType)
	router.GET("/api/quality-checks/statistics", s.getQualityCheckStatistics)

	// Webhook 接收
	router.POST("/api/webhook/github", s.githubWebhook)
	router.POST("/api/webhook/gitlab", s.gitlabWebhook)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting Event Store API server",
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
		"status":  "ok",
		"service": "event-store",
	})
}

// createEvent 创建事件
func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	var event storage.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.service.CreateEvent(r.Context(), &event); err != nil {
		s.logger.Error("Failed to create event", logger.Err(err))
		sharedapi.InternalError(w, "Failed to create event")
		return
	}

	sharedapi.OK(w, event)
}

// listEvents 列出事件
func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	filter := &storage.EventFilter{}

	// 解析查询参数
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		filter.EventType = storage.EventType(eventType)
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = storage.EventStatus(status)
	}
	if source := r.URL.Query().Get("source"); source != "" {
		filter.Source = source
	}
	if repoName := r.URL.Query().Get("repo_name"); repoName != "" {
		filter.RepoName = repoName
	}
	if prNumber := r.URL.Query().Get("pr_number"); prNumber != "" {
		if n, err := strconv.Atoi(prNumber); err == nil {
			filter.PRNumber = n
		}
	}
	if author := r.URL.Query().Get("author"); author != "" {
		filter.Author = author
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = &t
		}
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = &t
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if n, err := strconv.Atoi(offset); err == nil && n > 0 {
			filter.Offset = n
		}
	}

	events, err := s.service.ListEvents(r.Context(), filter)
	if err != nil {
		s.logger.Error("Failed to list events", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list events")
		return
	}

	sharedapi.OK(w, events)
}

// getEvent 获取单个事件
func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	event, err := s.service.GetEvent(r.Context(), uuid)
	if err != nil {
		s.logger.Error("Failed to get event", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get event")
		return
	}
	if event == nil {
		sharedapi.NotFound(w, "Event not found")
		return
	}

	sharedapi.OK(w, event)
}

// updateEventStatus 更新事件状态
func (s *Server) updateEventStatus(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	var req struct {
		Status storage.EventStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.service.UpdateEventStatus(r.Context(), uuid, req.Status); err != nil {
		s.logger.Error("Failed to update event status", logger.Err(err))
		sharedapi.InternalError(w, "Failed to update event status")
		return
	}

	sharedapi.OK(w, map[string]string{"status": "updated"})
}

// getEventStatistics 获取事件统计
func (s *Server) getEventStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := s.service.GetEventStatistics(r.Context())
	if err != nil {
		s.logger.Error("Failed to get event statistics", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get event statistics")
		return
	}
	sharedapi.OK(w, stats)
}

// getPendingEvents 获取待处理事件
func (s *Server) getPendingEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	events, err := s.service.GetPendingEvents(r.Context(), limit)
	if err != nil {
		s.logger.Error("Failed to get pending events", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get pending events")
		return
	}
	sharedapi.OK(w, events)
}

// getProcessingEvents 获取处理中的事件
func (s *Server) getProcessingEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.service.GetProcessingEvents(r.Context())
	if err != nil {
		s.logger.Error("Failed to get processing events", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get processing events")
		return
	}
	sharedapi.OK(w, events)
}

// getEventsByPR 获取 PR 相关事件
func (s *Server) getEventsByPR(w http.ResponseWriter, r *http.Request) {
	repoName := r.PathValue("repo_name")
	if repoName == "" {
		sharedapi.BadRequest(w, "repo_name is required")
		return
	}

	prNumberStr := r.PathValue("pr_number")
	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid pr_number")
		return
	}

	events, err := s.service.GetEventsByPR(r.Context(), repoName, prNumber)
	if err != nil {
		s.logger.Error("Failed to get events by PR", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get events by PR")
		return
	}
	sharedapi.OK(w, events)
}

// getQualityChecks 获取质量检查
func (s *Server) getQualityChecks(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	checks, err := s.service.GetQualityChecks(r.Context(), uuid)
	if err != nil {
		s.logger.Error("Failed to get quality checks", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get quality checks")
		return
	}
	sharedapi.OK(w, checks)
}

// createQualityCheck 创建质量检查
func (s *Server) createQualityCheck(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if uuid == "" {
		sharedapi.BadRequest(w, "UUID is required")
		return
	}

	var check storage.QualityCheck
	if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}
	check.EventUUID = uuid

	if err := s.service.CreateQualityCheck(r.Context(), &check); err != nil {
		s.logger.Error("Failed to create quality check", logger.Err(err))
		sharedapi.InternalError(w, "Failed to create quality check")
		return
	}

	sharedapi.OK(w, check)
}

// updateQualityCheck 更新质量检查
func (s *Server) updateQualityCheck(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid check ID")
		return
	}

	var req struct {
		Status  string  `json:"status"`
		Result  string  `json:"result"`
		Score   float64 `json:"score"`
		Details string  `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.service.UpdateQualityCheck(r.Context(), id, req.Status, req.Result, req.Score, req.Details); err != nil {
		s.logger.Error("Failed to update quality check", logger.Err(err))
		sharedapi.InternalError(w, "Failed to update quality check")
		return
	}

	sharedapi.OK(w, map[string]string{"status": "updated"})
}

// updateQualityCheckByType 根据类型更新质量检查
func (s *Server) updateQualityCheckByType(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	checkType := r.PathValue("type")
	if uuid == "" || checkType == "" {
		sharedapi.BadRequest(w, "UUID and type are required")
		return
	}

	var req struct {
		Status  string  `json:"status"`
		Result  string  `json:"result"`
		Score   float64 `json:"score"`
		Details string  `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.service.UpdateQualityCheckByType(r.Context(), uuid, checkType, req.Status, req.Result, req.Score, req.Details); err != nil {
		s.logger.Error("Failed to update quality check by type", logger.Err(err))
		sharedapi.InternalError(w, "Failed to update quality check")
		return
	}

	sharedapi.OK(w, map[string]string{"status": "updated"})
}

// getQualityCheckStatistics 获取质量检查统计
func (s *Server) getQualityCheckStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := s.service.GetQualityCheckStatistics(r.Context())
	if err != nil {
		s.logger.Error("Failed to get quality check statistics", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get quality check statistics")
		return
	}
	sharedapi.OK(w, stats)
}

// githubWebhook 处理 GitHub Webhook
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	// 解析 webhook payload
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("Failed to decode GitHub webhook", logger.Err(err))
		sharedapi.BadRequest(w, "Invalid payload")
		return
	}

	// 提取事件类型
	headers := r.Header
	eventType := headers.Get("X-GitHub-Event")

	s.logger.Info("Received GitHub webhook",
		logger.String("event_type", eventType),
	)

	// TODO: 根据 eventType 处理不同类型的 GitHub 事件
	// 这里需要根据实际需求实现事件转换逻辑

	sharedapi.OK(w, map[string]string{"status": "received"})
}

// gitlabWebhook 处理 GitLab Webhook
func (s *Server) gitlabWebhook(w http.ResponseWriter, r *http.Request) {
	// 解析 webhook payload
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error("Failed to decode GitLab webhook", logger.Err(err))
		sharedapi.BadRequest(w, "Invalid payload")
		return
	}

	// 提取事件类型
	headers := r.Header
	eventType := headers.Get("X-GitLab-Event")

	s.logger.Info("Received GitLab webhook",
		logger.String("event_type", eventType),
	)

	// TODO: 根据 eventType 处理不同类型的 GitLab 事件
	// 这里需要根据实际需求实现事件转换逻辑

	sharedapi.OK(w, map[string]string{"status": "received"})
}
