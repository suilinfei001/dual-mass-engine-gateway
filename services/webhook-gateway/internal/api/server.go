// Package api provides HTTP handlers for webhook gateway service.
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/webhook-gateway/internal/service"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

// Server Webhook 网关服务器
type Server struct {
	config    sharedapi.Config
	service   *service.WebhookGatewayService
	logger    *logger.Logger
	apiServer *sharedapi.Server
}

// NewServer 创建 Webhook 网关服务器
func NewServer(cfg sharedapi.Config, svc *service.WebhookGatewayService, log *logger.Logger) *Server {
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

	// Webhook 接收端点
	router.POST("/webhook/github", s.githubWebhook)
	router.POST("/webhook/gitlab", s.gitlabWebhook)

	// API 端点
	router.GET("/api/status", s.getStatus)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting Webhook Gateway API server",
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
		"service": "webhook-gateway",
	})
}

// githubWebhook 处理 GitHub Webhook
func (s *Server) githubWebhook(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 提取请求头
	headers := make(map[string]string)
	headers["X-GitHub-Event"] = r.Header.Get("X-GitHub-Event")
	headers["X-GitHub-Delivery"] = r.Header.Get("X-GitHub-Delivery")
	headers["X-Hub-Signature-256"] = r.Header.Get("X-Hub-Signature-256")
	headers["Content-Type"] = r.Header.Get("Content-Type")

	// 读取请求体
	payload, err := readRequestBody(r)
	if err != nil {
		s.logger.Error("Failed to read GitHub webhook body", logger.Err(err))
		sharedapi.InternalError(w, "Failed to read request body")
		return
	}

	s.logger.Info("Received GitHub webhook",
		logger.String("event", headers["X-GitHub-Event"]),
		logger.String("delivery", headers["X-GitHub-Delivery"]),
		logger.String("remote_addr", r.RemoteAddr),
	)

	// 处理 webhook
	result, err := s.service.HandleGitHubWebhook(r.Context(), headers, payload)
	if err != nil {
		s.logger.Error("Failed to process GitHub webhook", logger.Err(err))
		sharedapi.InternalError(w, "Failed to process webhook")
		return
	}

	duration := time.Since(startTime)
	s.logger.Info("GitHub webhook processed",
		logger.String("event_uuid", result.EventUUID),
		logger.Any("success", result.Success),
		logger.Int("status_code", result.StatusCode),
		logger.Any("duration_ms", duration.Milliseconds()),
	)

	if result.Success {
		sharedapi.OK(w, map[string]interface{}{
			"event_uuid": result.EventUUID,
			"status":     "accepted",
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		sharedapi.OK(w, map[string]interface{}{
			"event_uuid":    result.EventUUID,
			"status":        "accepted_with_error",
			"error_message": result.ErrorMessage,
		})
	}
}

// gitlabWebhook 处理 GitLab Webhook
func (s *Server) gitlabWebhook(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 提取请求头
	headers := make(map[string]string)
	headers["X-Gitlab-Event"] = r.Header.Get("X-Gitlab-Event")
	headers["X-Gitlab-Delivery"] = r.Header.Get("X-Gitlab-Delivery")
	headers["X-Gitlab-Token"] = r.Header.Get("X-Gitlab-Token")
	headers["Content-Type"] = r.Header.Get("Content-Type")

	// 读取请求体
	payload, err := readRequestBody(r)
	if err != nil {
		s.logger.Error("Failed to read GitLab webhook body", logger.Err(err))
		sharedapi.InternalError(w, "Failed to read request body")
		return
	}

	s.logger.Info("Received GitLab webhook",
		logger.String("event", headers["X-Gitlab-Event"]),
		logger.String("delivery", headers["X-Gitlab-Delivery"]),
		logger.String("remote_addr", r.RemoteAddr),
	)

	// 处理 webhook
	result, err := s.service.HandleGitLabWebhook(r.Context(), headers, payload)
	if err != nil {
		s.logger.Error("Failed to process GitLab webhook", logger.Err(err))
		sharedapi.InternalError(w, "Failed to process webhook")
		return
	}

	duration := time.Since(startTime)
	s.logger.Info("GitLab webhook processed",
		logger.String("event_uuid", result.EventUUID),
		logger.Any("success", result.Success),
		logger.Int("status_code", result.StatusCode),
		logger.Any("duration_ms", duration.Milliseconds()),
	)

	if result.Success {
		sharedapi.OK(w, map[string]interface{}{
			"event_uuid": result.EventUUID,
			"status":     "accepted",
			"duration_ms": duration.Milliseconds(),
		})
	} else {
		sharedapi.OK(w, map[string]interface{}{
			"event_uuid":    result.EventUUID,
			"status":        "accepted_with_error",
			"error_message": result.ErrorMessage,
		})
	}
}

// getStatus 获取服务状态
func (s *Server) getStatus(w http.ResponseWriter, r *http.Request) {
	sharedapi.OK(w, map[string]interface{}{
		"service":   "webhook-gateway",
		"status":    "running",
		"timestamp": time.Now().Unix(),
	})
}

// readRequestBody 读取请求体
func readRequestBody(r *http.Request) ([]byte, error) {
	// 限制请求体大小 (10MB)
	r.Body = http.MaxBytesReader(nil, r.Body, 10<<20)

	data := make([]byte, 0, 1024) // 预分配 1KB
	buf := make([]byte, 1024)
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "http: request body too large" {
				return nil, err
			}
			break
		}
	}
	return data, nil
}
