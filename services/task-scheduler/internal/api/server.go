// Package api provides HTTP handlers for task scheduler service.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/task-scheduler/internal/models"
	"github.com/quality-gateway/task-scheduler/internal/scheduler"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

// Server Task Scheduler API 服务器
type Server struct {
	config    sharedapi.Config
	scheduler *scheduler.Scheduler
	logger    *logger.Logger
	apiServer *sharedapi.Server
}

// NewServer 创建 API 服务器
func NewServer(cfg sharedapi.Config, sched *scheduler.Scheduler, log *logger.Logger) *Server {
	srv := sharedapi.New(cfg, log)

	return &Server{
		config:    cfg,
		scheduler: sched,
		logger:    log,
		apiServer: srv,
	}
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes() {
	router := s.apiServer.Router()

	// 健康检查
	router.GET("/health", s.healthCheck)

	// 任务路由
	router.GET("/api/tasks", s.listTasks)
	router.GET("/api/tasks/{id}", s.getTask)
	router.POST("/api/tasks/{id}/start", s.startTask)
	router.POST("/api/tasks/{id}/complete", s.completeTask)
	router.POST("/api/tasks/{id}/fail", s.failTask)
	router.POST("/api/tasks/{id}/cancel", s.cancelTask)

	// 事件路由
	router.POST("/api/events/{event-id}/cancel", s.cancelEventTasks)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting Task Scheduler API server",
		logger.String("address", s.config.Address()),
	)
	return s.apiServer.Start()
}

// Shutdown 关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	return s.apiServer.Shutdown()
}

// healthCheck 健康检查
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	sharedapi.OK(w, map[string]string{
		"status":  "healthy",
		"service": "task-scheduler",
	})
}

// listTasks 获取任务列表
func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	limit := 100
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	tasks, total, err := s.scheduler.ListTasks(limit, offset)
	if err != nil {
		s.logger.Error("Failed to list tasks", logger.Err(err))
		sharedapi.InternalError(w, "Failed to list tasks")
		return
	}

	data := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		data[i] = task.ToTaskResponse()
	}

	sharedapi.OK(w, map[string]interface{}{
		"tasks":  data,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// getTask 获取任务详情
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid task ID")
		return
	}

	task, err := s.scheduler.GetTask(id)
	if err != nil {
		s.logger.Error("Failed to get task",
			logger.Int("id", id),
			logger.Err(err),
		)
		sharedapi.NotFound(w, "Task not found")
		return
	}

	sharedapi.OK(w, task.ToTaskResponse())
}

// startTask 启动任务
func (s *Server) startTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid task ID")
		return
	}

	task, err := s.scheduler.StartTask(id)
	if err != nil {
		s.logger.Error("Failed to start task",
			logger.Int("id", id),
			logger.Err(err),
		)
		sharedapi.InternalError(w, err.Error())
		return
	}

	s.logger.Info("Task started",
		logger.Int("id", id),
		logger.String("task_name", task.TaskName),
	)

	sharedapi.OK(w, task.ToTaskResponse())
}

// completeTask 完成任务
func (s *Server) completeTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid task ID")
		return
	}

	var req struct {
		Results []models.TaskResult `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.scheduler.CompleteTask(id, req.Results); err != nil {
		s.logger.Error("Failed to complete task",
			logger.Int("id", id),
			logger.Err(err),
		)
		sharedapi.InternalError(w, "Failed to complete task")
		return
	}

	s.logger.Info("Task completed", logger.Int("id", id))

	sharedapi.OK(w, map[string]interface{}{
		"message": "Task completed successfully",
	})
}

// failTask 标记任务失败
func (s *Server) failTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid task ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.scheduler.FailTask(id, req.Reason); err != nil {
		s.logger.Error("Failed to mark task as failed",
			logger.Int("id", id),
			logger.Err(err),
		)
		sharedapi.InternalError(w, "Failed to mark task as failed")
		return
	}

	s.logger.Info("Task marked as failed",
		logger.Int("id", id),
		logger.String("reason", req.Reason),
	)

	sharedapi.OK(w, map[string]interface{}{
		"message": "Task marked as failed",
	})
}

// cancelTask 取消任务
func (s *Server) cancelTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid task ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if err := s.scheduler.CancelTask(id, req.Reason); err != nil {
		s.logger.Error("Failed to cancel task",
			logger.Int("id", id),
			logger.Err(err),
		)
		sharedapi.InternalError(w, "Failed to cancel task")
		return
	}

	s.logger.Info("Task cancelled",
		logger.Int("id", id),
		logger.String("reason", req.Reason),
	)

	sharedapi.OK(w, map[string]interface{}{
		"message": "Task cancelled successfully",
	})
}

// cancelEventTasks 取消事件的所有任务
func (s *Server) cancelEventTasks(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("event-id")

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		sharedapi.BadRequest(w, "Invalid event ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	count, err := s.scheduler.CancelEventTasks(eventID, req.Reason)
	if err != nil {
		s.logger.Error("Failed to cancel event tasks",
			logger.Int("event_id", eventID),
			logger.Err(err),
		)
		sharedapi.InternalError(w, "Failed to cancel event tasks")
		return
	}

	s.logger.Info("Event tasks cancelled",
		logger.Int("event_id", eventID),
		logger.Int("count", count),
	)

	sharedapi.OK(w, map[string]interface{}{
		"message": "Tasks cancelled successfully",
		"count":   count,
	})
}
