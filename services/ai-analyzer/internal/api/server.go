// Package api provides HTTP handlers for AI analyzer service.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/quality-gateway/ai-analyzer/internal/analyzer"
	"github.com/quality-gateway/ai-analyzer/internal/pool"
	"github.com/quality-gateway/shared/pkg/logger"
	sharedapi "github.com/quality-gateway/shared/pkg/api"
)

// Server AI Analyzer API 服务器
type Server struct {
	config      sharedapi.Config
	logAnalyzer *analyzer.LogAnalyzer
	requestPool *pool.AIRequestPool
	logger      *logger.Logger
	apiServer   *sharedapi.Server
}

// NewServer 创建 API 服务器
func NewServer(cfg sharedapi.Config, logAnalyzer *analyzer.LogAnalyzer, requestPool *pool.AIRequestPool, log *logger.Logger) *Server {
	srv := sharedapi.New(cfg, log)

	return &Server{
		config:      cfg,
		logAnalyzer: logAnalyzer,
		requestPool: requestPool,
		logger:      log,
		apiServer:   srv,
	}
}

// RegisterRoutes 注册路由
func (s *Server) RegisterRoutes() {
	router := s.apiServer.Router()

	// 健康检查
	router.GET("/health", s.healthCheck)
	router.GET("/api/health", s.healthCheck)

	// 日志分析端点
	router.POST("/api/analyze", s.analyzeLog)
	router.POST("/api/analyze/batch", s.batchAnalyzeLogs)

	// 池管理端点
	router.POST("/api/config/pool-size", s.setPoolSize)
	router.GET("/api/config/pool-size", s.getPoolSize)
	router.GET("/api/pool/stats", s.getPoolStats)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.RegisterRoutes()
	s.logger.Info("Starting AI Analyzer API server",
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
	sharedapi.OK(w, map[string]interface{}{
		"status":  "healthy",
		"service": "ai-analyzer",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// analyzeLog 分析单个日志
func (s *Server) analyzeLog(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received analyze request",
		logger.String("remote_addr", r.RemoteAddr),
	)

	var req analyzer.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if req.LogContent == "" {
		sharedapi.BadRequest(w, "log_content is required")
		return
	}

	// 使用默认任务名称
	taskName := req.TaskName
	if taskName == "" {
		taskName = "basic_ci_all"
	}

	results, err := s.logAnalyzer.AnalyzeLog(req.LogContent, taskName)
	if err != nil {
		s.logger.Error("Analysis failed", logger.Err(err))
		sharedapi.InternalError(w, "Analysis failed: "+err.Error())
		return
	}

	sharedapi.OK(w, analyzer.LogAnalysisResult{Results: results})
}

// batchAnalyzeLogs 批量分析日志
func (s *Server) batchAnalyzeLogs(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received batch analyze request",
		logger.String("remote_addr", r.RemoteAddr),
	)

	var req analyzer.BatchAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.LogContents) == 0 {
		sharedapi.BadRequest(w, "log_contents is required")
		return
	}

	// 使用默认任务名称
	taskName := req.TaskName
	if taskName == "" {
		taskName = "basic_ci_all"
	}

	// 设置合理的超时时间
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	results, err := s.logAnalyzer.AnalyzeBatch(ctx, req.LogContents, taskName)
	if err != nil {
		s.logger.Error("Batch analysis failed", logger.Err(err))
		sharedapi.InternalError(w, "Batch analysis failed: "+err.Error())
		return
	}

	sharedapi.OK(w, analyzer.LogAnalysisResult{Results: results})
}

// setPoolSize 设置池大小
func (s *Server) setPoolSize(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received pool size update request",
		logger.String("remote_addr", r.RemoteAddr),
	)

	var req analyzer.PoolSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedapi.BadRequest(w, "Invalid request body")
		return
	}

	if req.Size <= 0 {
		sharedapi.BadRequest(w, "size must be greater than 0")
		return
	}

	if err := s.requestPool.Resize(req.Size); err != nil {
		s.logger.Error("Failed to resize pool", logger.Err(err))
		sharedapi.InternalError(w, "Failed to resize pool: "+err.Error())
		return
	}

	s.logger.Info("Pool size updated", logger.Int("size", req.Size))

	sharedapi.OK(w, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Pool size updated to %d", req.Size),
		"size":    req.Size,
	})
}

// getPoolSize 获取池大小
func (s *Server) getPoolSize(w http.ResponseWriter, r *http.Request) {
	stats := s.requestPool.GetStats()
	sharedapi.OK(w, map[string]interface{}{
		"total_size": stats.TotalSize,
	})
}

// getPoolStats 获取池统计
func (s *Server) getPoolStats(w http.ResponseWriter, r *http.Request) {
	stats := s.requestPool.GetStats()
	sharedapi.OK(w, map[string]interface{}{
		"total_size":     stats.TotalSize,
		"available":      stats.Available,
		"in_use":         stats.InUse,
		"usage_percent":  stats.UsagePercentage(),
	})
}
