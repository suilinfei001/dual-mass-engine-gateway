// Package api provides HTTP handlers for executor service.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/executor-service/internal/azure"
	"github.com/quality-gateway/executor-service/internal/service"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

// Server 执行器服务器
type Server struct {
	config sharedapi.Config
	service *service.ExecutorService
	logger  *logger.Logger
	apiServer *sharedapi.Server
}

// NewServer 创建执行器服务器
func NewServer(cfg sharedapi.Config, svc *service.ExecutorService, log *logger.Logger) *Server {
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

	// 执行相关
	router.POST("/api/execute", s.executeTask)
	router.GET("/api/executions/{id}", s.getExecutionStatus)
	router.GET("/api/executions/{id}/logs", s.getExecutionLogs)
	router.DELETE("/api/executions/{id}", s.cancelExecution)
	router.GET("/api/executions", s.listExecutions)

	// Pipeline 相关 (兼容旧接口)
	router.GET("/api/status/{buildId}", s.getPipelineStatus)
	router.GET("/api/logs/{buildId}", s.getPipelineLogs)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting Executor Service API server",
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
		"status":  "ok",
		"service": "executor-service",
	})
}

// executeTask 执行任务
func (s *Server) executeTask(w http.ResponseWriter, r *http.Request) {
	var req azure.TaskExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	// 设置默认值
	if req.TaskUUID == "" {
		sharedapi.BadRequest(w, "task_uuid is required")
		return
	}

	resp, err := s.service.ExecuteTask(r.Context(), &req)
	if err != nil {
		s.logger.Error("Failed to execute task", logger.Err(err))
		sharedapi.InternalError(w, "Failed to execute task")
		return
	}

	sharedapi.OK(w, resp)
}

// getExecutionStatus 获取执行状态
func (s *Server) getExecutionStatus(w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")
	if executionID == "" {
		sharedapi.BadRequest(w, "execution ID is required")
		return
	}

	status, err := s.service.GetExecutionStatus(r.Context(), executionID)
	if err != nil {
		s.logger.Error("Failed to get execution status", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get execution status")
		return
	}

	sharedapi.OK(w, status)
}

// getExecutionLogs 获取执行日志
func (s *Server) getExecutionLogs(w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")
	if executionID == "" {
		sharedapi.BadRequest(w, "execution ID is required")
		return
	}

	logs, err := s.service.GetExecutionLogs(r.Context(), executionID)
	if err != nil {
		s.logger.Error("Failed to get execution logs", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get execution logs")
		return
	}

	sharedapi.OK(w, map[string]interface{}{
		"execution_id": executionID,
		"logs":         logs,
		"count":        len(logs),
	})
}

// cancelExecution 取消执行
func (s *Server) cancelExecution(w http.ResponseWriter, r *http.Request) {
	executionID := r.PathValue("id")
	if executionID == "" {
		sharedapi.BadRequest(w, "execution ID is required")
		return
	}

	if err := s.service.CancelExecution(r.Context(), executionID); err != nil {
		s.logger.Error("Failed to cancel execution", logger.Err(err))
		sharedapi.InternalError(w, "Failed to cancel execution")
		return
	}

	sharedapi.OK(w, map[string]string{
		"status":       "canceled",
		"execution_id": executionID,
	})
}

// listExecutions 列出执行记录
func (s *Server) listExecutions(w http.ResponseWriter, r *http.Request) {
	executions := s.service.ListExecutions()

	sharedapi.OK(w, map[string]interface{}{
		"executions": executions,
		"count":      len(executions),
	})
}

// getPipelineStatus 获取 Pipeline 状态 (兼容旧接口)
func (s *Server) getPipelineStatus(w http.ResponseWriter, r *http.Request) {
	buildId := r.PathValue("buildId")
	if buildId == "" {
		sharedapi.BadRequest(w, "build ID is required")
		return
	}

	// 尝试将 buildId 作为 executionID 查找
	status, err := s.service.GetExecutionStatus(r.Context(), buildId)
	if err != nil {
		s.logger.Error("Failed to get pipeline status", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get pipeline status")
		return
	}

	// 转换为旧格式响应
	response := map[string]interface{}{
		"build_id":    status.RunID,
		"status":      string(status.Status),
		"result":      string(status.Result),
		"finished":    status.Finished,
	}

	if status.StartedAt != nil {
		response["started_at"] = status.StartedAt.Format(time.RFC3339)
	}
	if status.CompletedAt != nil {
		response["completed_at"] = status.CompletedAt.Format(time.RFC3339)
	}

	sharedapi.OK(w, response)
}

// getPipelineLogs 获取 Pipeline 日志 (兼容旧接口)
func (s *Server) getPipelineLogs(w http.ResponseWriter, r *http.Request) {
	buildId := r.PathValue("buildId")
	if buildId == "" {
		sharedapi.BadRequest(w, "build ID is required")
		return
	}

	// 尝试将 buildId 作为 executionID 查找
	logs, err := s.service.GetExecutionLogs(r.Context(), buildId)
	if err != nil {
		s.logger.Error("Failed to get pipeline logs", logger.Err(err))
		sharedapi.InternalError(w, "Failed to get pipeline logs")
		return
	}

	// 合并所有日志
	combinedLogs := ""
	for _, log := range logs {
		combinedLogs += log + "\n"
	}

	sharedapi.OK(w, map[string]interface{}{
		"build_id": buildId,
		"logs":     combinedLogs,
	})
}
