// Package service provides business logic for executor service.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/executor-service/internal/azure"
)

// ExecutorService 执行器服务
type ExecutorService struct {
	azureClient *azure.AzureClient
	logger      *logger.Logger

	// 执行记录存储
	executions sync.Map // executionID -> *ExecutionRecord
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ExecutionID string
	TaskUUID    string
	RunID       int64
	Status      azure.PipelineStatus
	Project     string
	Organization string
	CreatedAt   time.Time
}

// NewExecutorService 创建执行器服务
func NewExecutorService(ac *azure.AzureClient, log *logger.Logger) *ExecutorService {
	return &ExecutorService{
		azureClient: ac,
		logger:      log,
	}
}

// ExecuteTask 执行任务
func (s *ExecutorService) ExecuteTask(ctx context.Context, req *azure.TaskExecutionRequest) (*azure.TaskExecutionResponse, error) {
	// 生成执行 ID
	executionID := uuid.New().String()

	// 构建 Pipeline 请求
	pipelineReq := &azure.PipelineRunRequest{
		Organization: req.TaskType, // 使用任务类型作为项目标识
		Project:      req.TaskType,
		PipelineID:   1, // 默认 Pipeline ID，可配置
		SourceBranch: "main",
		Parameters:   make(map[string]string),
	}

	// 设置参数
	if req.ChartURL != "" {
		pipelineReq.Parameters["chart_url"] = req.ChartURL
	}
	if req.TestbedIP != "" {
		pipelineReq.Parameters["testbed_ip"] = req.TestbedIP
		pipelineReq.Parameters["testbed_ssh_port"] = fmt.Sprint(req.TestbedSSHPort)
		pipelineReq.Parameters["testbed_ssh_user"] = req.TestbedSSHUser
		if req.TestbedSSHPassword != "" {
			pipelineReq.Parameters["testbed_ssh_password"] = req.TestbedSSHPassword
		}
	}

	// 添加自定义参数
	for k, v := range req.Parameters {
		pipelineReq.Parameters[k] = v
	}

	s.logger.Info("Executing task",
		logger.String("execution_id", executionID),
		logger.String("task_uuid", req.TaskUUID),
		logger.String("task_type", req.TaskType),
		logger.String("chart_url", req.ChartURL),
	)

	// 调用 Azure 执行
	pipelineResp, err := s.azureClient.RunPipeline(ctx, pipelineReq)
	if err != nil {
		return nil, fmt.Errorf("failed to run pipeline: %w", err)
	}

	// 保存执行记录
	record := &ExecutionRecord{
		ExecutionID: executionID,
		TaskUUID:    req.TaskUUID,
		RunID:       pipelineResp.RunID,
		Status:      pipelineResp.Status,
		Project:     pipelineReq.Project,
		Organization: pipelineReq.Organization,
		CreatedAt:   time.Now(),
	}
	s.executions.Store(executionID, record)

	s.logger.Info("Task execution started",
		logger.String("execution_id", executionID),
		logger.Int64("run_id", pipelineResp.RunID),
		logger.String("run_url", pipelineResp.RunURL),
		logger.String("status", string(pipelineResp.Status)),
	)

	return &azure.TaskExecutionResponse{
		ExecutionID: executionID,
		RunID:       pipelineResp.RunID,
		RunURL:      pipelineResp.RunURL,
		Status:      pipelineResp.Status,
		CreatedAt:   time.Now(),
	}, nil
}

// GetExecutionStatus 获取执行状态
func (s *ExecutorService) GetExecutionStatus(ctx context.Context, executionID string) (*azure.TaskExecutionStatus, error) {
	// 查找执行记录
	recordI, ok := s.executions.Load(executionID)
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}
	record := recordI.(*ExecutionRecord)

	// 获取 Pipeline 状态
	pipelineStatus, err := s.azureClient.GetPipelineStatus(ctx, record.Organization, record.Project, record.RunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline status: %w", err)
	}

	// 更新记录中的状态
	record.Status = pipelineStatus.Status

	return &azure.TaskExecutionStatus{
		ExecutionID: executionID,
		RunID:       pipelineStatus.RunID,
		Status:      pipelineStatus.Status,
		Result:      pipelineStatus.Result,
		StartedAt:   pipelineStatus.StartedAt,
		CompletedAt: pipelineStatus.CompletedAt,
		Finished:    pipelineStatus.Finished,
	}, nil
}

// GetExecutionLogs 获取执行日志
func (s *ExecutorService) GetExecutionLogs(ctx context.Context, executionID string) ([]string, error) {
	// 查找执行记录
	recordI, ok := s.executions.Load(executionID)
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}
	record := recordI.(*ExecutionRecord)

	s.logger.Info("Getting execution logs",
		logger.String("execution_id", executionID),
		logger.Int64("run_id", record.RunID),
	)

	// 获取 Pipeline 日志
	logs, err := s.azureClient.GetPipelineLogs(ctx, record.Organization, record.Project, record.RunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline logs: %w", err)
	}

	return logs, nil
}

// CancelExecution 取消执行
func (s *ExecutorService) CancelExecution(ctx context.Context, executionID string) error {
	// 查找执行记录
	recordI, ok := s.executions.Load(executionID)
	if !ok {
		return fmt.Errorf("execution not found: %s", executionID)
	}
	record := recordI.(*ExecutionRecord)

	s.logger.Info("Canceling execution",
		logger.String("execution_id", executionID),
		logger.Int64("run_id", record.RunID),
	)

	// 取消 Pipeline
	if err := s.azureClient.CancelPipeline(ctx, record.Organization, record.Project, record.RunID); err != nil {
		return fmt.Errorf("failed to cancel pipeline: %w", err)
	}

	// 更新状态
	record.Status = azure.PipelineStatusCanceled

	return nil
}

// ListExecutions 列出执行记录
func (s *ExecutorService) ListExecutions() []*azure.TaskExecutionStatus {
	var executions []*azure.TaskExecutionStatus

	s.executions.Range(func(key, value interface{}) bool {
		record := value.(*ExecutionRecord)
		executions = append(executions, &azure.TaskExecutionStatus{
			ExecutionID: record.ExecutionID,
			RunID:       record.RunID,
			Status:      record.Status,
		})
		return true
	})

	return executions
}

// CleanupOldExecutions 清理旧的执行记录
func (s *ExecutorService) CleanupOldExecutions(olderThan time.Duration) int {
	count := 0
	cutoff := time.Now().Add(-olderThan)

	s.executions.Range(func(key, value interface{}) bool {
		record := value.(*ExecutionRecord)
		if record.CreatedAt.Before(cutoff) {
			// 只清理已完成的执行
			if record.Status == azure.PipelineStatusCompleted ||
			   record.Status == azure.PipelineStatusFailed ||
			   record.Status == azure.PipelineStatusCanceled {
				s.executions.Delete(key)
				count++
			}
		}
		return true
	})

	if count > 0 {
		s.logger.Info("Cleaned up old executions",
			logger.Int("count", count),
			logger.String("older_than", olderThan.String()),
		)
	}

	return count
}
