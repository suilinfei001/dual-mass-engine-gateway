package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hugoh/go-designs/resource-pool/internal/ai"
	"github.com/hugoh/go-designs/resource-pool/internal/executor"
	"github.com/hugoh/go-designs/resource-pool/internal/storage"
)

// DeploymentService 部署服务接口
type DeploymentService interface {
	// DeployProductAsync 异步部署产品
	DeployProductAsync(ctx context.Context, allocationID int, pipelineID int, params map[string]interface{}) (*storage.DeploymentTask, error)

	// MonitorDeployments 监控部署任务（后台运行）
	MonitorDeployments(ctx context.Context, interval time.Duration)

	// GetTask 获取部署任务
	GetTask(taskUUID string) (*storage.DeploymentTask, error)

	// ListRecentTasks 列出最近的部署任务
	ListRecentTasks(limit int) ([]*storage.DeploymentTask, error)

	// SetTestbedStorage 设置 Testbed 存储
	SetTestbedStorage(testbedStorage storage.TestbedStorage)

	// SetResourceInstanceStorage 设置 ResourceInstance 存储
	SetResourceInstanceStorage(resourceInstanceStorage storage.ResourceInstanceStorage)
}

// DeploymentServiceImpl 部署服务实现
type DeploymentServiceImpl struct {
	deploymentStorage     storage.DeploymentTaskStorage
	configStorage         storage.ConfigStorage
	allocationStorage     storage.AllocationStorage
	testbedStorage        storage.TestbedStorage
	resourceInstanceStorage storage.ResourceInstanceStorage
	azureExecutor         *executor.AzureDeployExecutor
	logAnalyzer           *ai.DeploymentLogAnalyzer
}

// NewDeploymentService 创建部署服务
func NewDeploymentService(
	deploymentStorage storage.DeploymentTaskStorage,
	configStorage storage.ConfigStorage,
	allocationStorage storage.AllocationStorage,
) DeploymentService {
	return &DeploymentServiceImpl{
		deploymentStorage: deploymentStorage,
		configStorage:     configStorage,
		allocationStorage: allocationStorage,
	}
}

// SetTestbedStorage 设置 Testbed 存储
func (s *DeploymentServiceImpl) SetTestbedStorage(testbedStorage storage.TestbedStorage) {
	s.testbedStorage = testbedStorage
}

// SetResourceInstanceStorage 设置 ResourceInstance 存储
func (s *DeploymentServiceImpl) SetResourceInstanceStorage(resourceInstanceStorage storage.ResourceInstanceStorage) {
	s.resourceInstanceStorage = resourceInstanceStorage
}

// getAzureExecutor 获取 Azure 执行器（懒加载）
func (s *DeploymentServiceImpl) getAzureExecutor() (*executor.AzureDeployExecutor, error) {
	if s.azureExecutor == nil {
		azureConfig, err := s.configStorage.GetAzureConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get Azure config: %w", err)
		}
		s.azureExecutor = executor.NewAzureDeployExecutor(&executor.AzureConfig{
			Organization: azureConfig.Organization,
			Project:      azureConfig.Project,
			PAT:          azureConfig.PAT,
			BaseURL:      azureConfig.BaseURL,
		})
	}
	return s.azureExecutor, nil
}

// getLogAnalyzer 获取日志分析器（懒加载）
func (s *DeploymentServiceImpl) getLogAnalyzer() (*ai.DeploymentLogAnalyzer, error) {
	if s.logAnalyzer == nil {
		s.logAnalyzer = ai.NewDeploymentLogAnalyzer(s.configStorage)
	}
	return s.logAnalyzer, nil
}

// DeployProductAsync 异步部署产品
func (s *DeploymentServiceImpl) DeployProductAsync(ctx context.Context, allocationID int, pipelineID int, params map[string]interface{}) (*storage.DeploymentTask, error) {
	log.Printf("[DeploymentService] Starting async deployment for allocation_id=%d, pipeline_id=%d", allocationID, pipelineID)

	// 1. 验证 allocation 存在
	allocation, err := s.allocationStorage.GetAllocation(allocationID)
	if err != nil {
		return nil, fmt.Errorf("allocation not found: %w", err)
	}
	if allocation == nil {
		return nil, fmt.Errorf("allocation with ID %d not found", allocationID)
	}

	// 2. 替换参数中的占位符
	finalParams, err := s.replacePlaceholders(allocation.TestbedUUID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	log.Printf("[DeploymentService] Parameters after replacement: %+v", finalParams)

	// 3. 获取 Azure 执行器
	azureExec, err := s.getAzureExecutor()
	if err != nil {
		return nil, err
	}

	// 4. 运行 Azure Pipeline
	runResult, err := azureExec.RunPipeline(ctx, pipelineID, finalParams, "refs/heads/test")
	if err != nil {
		return nil, fmt.Errorf("failed to run pipeline: %w", err)
	}

	log.Printf("[DeploymentService] Pipeline started: build_id=%d, build_number=%s", runResult.BuildID, runResult.BuildNumber)

	// 5. 创建部署任务记录
	task := &storage.DeploymentTask{
		TaskUUID:     storage.GenerateUUID(),
		AllocationID: allocationID,
		PipelineID:   pipelineID,
		BuildID:      runResult.BuildID,
		Status:       storage.DeploymentTaskStatusRunning,
		Analyzing:    false,
		LogDirectory: "",
		ResultDetails: make(map[string]interface{}),
		ErrorMessage: "",
		WebURL:       runResult.WebURL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.deploymentStorage.CreateTask(task); err != nil {
		return nil, fmt.Errorf("failed to create deployment task: %w", err)
	}

	log.Printf("[DeploymentService] Deployment task created: task_uuid=%s", task.TaskUUID)

	return task, nil
}

// replacePlaceholders 替换参数中的占位符为实际值
func (s *DeploymentServiceImpl) replacePlaceholders(testbedUUID string, params map[string]interface{}) (map[string]interface{}, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	// 复制参数以避免修改原始参数
	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = v
	}

	// 如果 storage 不可用，直接返回原始参数
	if s.testbedStorage == nil || s.resourceInstanceStorage == nil {
		log.Printf("[DeploymentService] Testbed or ResourceInstance storage not available, skipping placeholder replacement")
		return result, nil
	}

	// 获取 Testbed
	testbed, err := s.testbedStorage.GetTestbedByUUID(testbedUUID)
	if err != nil {
		log.Printf("[DeploymentService] Failed to get testbed: %v", err)
		return result, nil
	}

	// 获取 ResourceInstance
	resourceInstance, err := s.resourceInstanceStorage.GetResourceInstanceByUUID(testbed.ResourceInstanceUUID)
	if err != nil {
		log.Printf("[DeploymentService] Failed to get resource instance: %v", err)
		return result, nil
	}

	log.Printf("[DeploymentService] Replacing placeholders for resource instance: ip=%s, ssh_user=%s",
		resourceInstance.IPAddress, resourceInstance.SSHUser)

	// 替换占位符
	// 字符串类型的值
	for key, value := range result {
		if strValue, ok := value.(string); ok {
			switch strValue {
			case "to_be_replace_host":
				result[key] = resourceInstance.IPAddress
				log.Printf("[DeploymentService] Replaced %s: %s -> %s", key, strValue, resourceInstance.IPAddress)
			case "to_be_replace_ssh_user":
				result[key] = resourceInstance.SSHUser
				log.Printf("[DeploymentService] Replaced %s: %s -> %s", key, strValue, resourceInstance.SSHUser)
			case "to_be_replace_ssh_password":
				result[key] = resourceInstance.Passwd
				log.Printf("[DeploymentService] Replaced %s: %s -> ***", key, strValue)
			}
		}
	}

	return result, nil
}

// MonitorDeployments 监控部署任务（后台运行）
func (s *DeploymentServiceImpl) MonitorDeployments(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 60 * time.Second // 默认 60 秒
	}

	log.Printf("[DeploymentService] Starting deployment monitor with interval: %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[DeploymentService] Deployment monitor stopped")
			return
		case <-ticker.C:
			s.checkAndProcessTasks(ctx)
		}
	}
}

// checkAndProcessTasks 检查并处理待处理的任务
func (s *DeploymentServiceImpl) checkAndProcessTasks(ctx context.Context) {
	log.Printf("[DeploymentService] Checking for pending/running deployment tasks...")

	// 获取运行中的任务
	tasks, err := s.deploymentStorage.GetRunningTasks()
	if err != nil {
		log.Printf("[DeploymentService] Failed to get running tasks: %v", err)
		return
	}

	log.Printf("[DeploymentService] Found %d running tasks", len(tasks))

	for _, task := range tasks {
		if err := s.processTask(ctx, task); err != nil {
			log.Printf("[DeploymentService] Failed to process task %s: %v", task.TaskUUID, err)
		}
	}
}

// processTask 处理单个任务
func (s *DeploymentServiceImpl) processTask(ctx context.Context, task *storage.DeploymentTask) error {
	log.Printf("[DeploymentService] Processing task: uuid=%s, build_id=%d", task.TaskUUID, task.BuildID)

	// 获取 Azure 执行器
	azureExec, err := s.getAzureExecutor()
	if err != nil {
		return err
	}

	// 获取 Pipeline 状态
	status, err := azureExec.GetStatus(ctx, task.BuildID)
	if err != nil {
		log.Printf("[DeploymentService] Failed to get status for build_id=%d: %v", task.BuildID, err)
		return err // 暂时不更新状态，下次重试
	}

	log.Printf("[DeploymentService] Task %s: status=%s, result=%s", task.TaskUUID, status.Status, status.Result)

	// 检查是否完成
	if !azureExec.IsCompleted(status.Status) {
		// 还在运行中，不更新状态
		return nil
	}

	// Pipeline 已完成，触发日志分析
	log.Printf("[DeploymentService] Task %s completed, triggering log analysis", task.TaskUUID)

	// 使用 CAS 操作启动分析，防止重复
	started, err := s.deploymentStorage.TryStartAnalysis(task.ID)
	if err != nil {
		return fmt.Errorf("failed to try start analysis: %w", err)
	}
	if !started {
		log.Printf("[DeploymentService] Analysis already in progress for task %s", task.TaskUUID)
		return nil
	}

	// 分析完成后更新任务状态
	defer func() {
		if err := s.deploymentStorage.ResetAnalyzingFlag(task.ID); err != nil {
			log.Printf("[DeploymentService] Failed to reset analyzing flag: %v", err)
		}
	}()

	// 获取并分析日志
	result, err := s.fetchAndAnalyzeLogs(ctx, task)
	if err != nil {
		log.Printf("[DeploymentService] Failed to analyze logs for task %s: %v", task.TaskUUID, err)
		// 分析失败，标记为失败但不中断流程
		s.deploymentStorage.UpdateTaskStatus(task.ID, storage.DeploymentTaskStatusFailed)
		s.deploymentStorage.UpdateBuildInfo(task.ID, task.BuildID, task.WebURL)
		return nil
	}

	// 更新任务状态
	finalStatus := storage.DeploymentTaskStatusCompleted
	if !result.Success {
		finalStatus = storage.DeploymentTaskStatusFailed
	}

	task.Status = finalStatus
	task.ResultDetails = map[string]interface{}{
		"success":       result.Success,
		"error_message": result.ErrorMessage,
		"mariadb_port":  result.MariaDBPort,
		"mariadb_user":  result.MariaDBUser,
		"app_port":      result.AppPort,
		"health_status": result.HealthStatus,
		"deployment_id": result.DeploymentID,
		"summary":       result.Summary,
	}
	task.UpdatedAt = time.Now()

	if result.ErrorMessage != "" {
		task.ErrorMessage = result.ErrorMessage
	}

	if err := s.deploymentStorage.UpdateTask(task); err != nil {
		log.Printf("[DeploymentService] Failed to update task: %v", err)
	}

	log.Printf("[DeploymentService] Task %s completed with status=%s", task.TaskUUID, finalStatus)
	return nil
}

// fetchAndAnalyzeLogs 获取并分析日志
func (s *DeploymentServiceImpl) fetchAndAnalyzeLogs(ctx context.Context, task *storage.DeploymentTask) (*ai.DeploymentAnalysisResult, error) {
	log.Printf("[DeploymentService] Fetching and analyzing logs for task %s", task.TaskUUID)

	// 获取 Azure 执行器
	azureExec, err := s.getAzureExecutor()
	if err != nil {
		return nil, err
	}

	// 创建日志目录
	logDir := filepath.Join("/tmp/resource_pool_logs", fmt.Sprintf("build_%d", task.BuildID))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 获取日志列表
	logEntries, err := azureExec.GetLogList(ctx, task.BuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get log list: %w", err)
	}

	log.Printf("[DeploymentService] Found %d log entries for build_id=%d", len(logEntries), task.BuildID)

	// 获取每个日志的内容
	for _, entry := range logEntries {
		content, err := azureExec.GetLogContent(ctx, task.BuildID, entry.LogID)
		if err != nil {
			log.Printf("[DeploymentService] Failed to get log content for log_id=%d: %v", entry.LogID, err)
			continue
		}

		// 保存到文件
		logFile := filepath.Join(logDir, fmt.Sprintf("log_%d.txt", entry.LogID))
		if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
			log.Printf("[DeploymentService] Failed to write log file: %v", err)
			continue
		}

		log.Printf("[DeploymentService] Saved log: log_id=%d, size=%d bytes", entry.LogID, len(content))
	}

	// 更新任务的日志目录
	task.LogDirectory = logDir
	if err := s.deploymentStorage.UpdateTask(task); err != nil {
		log.Printf("[DeploymentService] Failed to update task with log directory: %v", err)
	}

	// 使用 AI 分析日志
	analyzer, err := s.getLogAnalyzer()
	if err != nil {
		return nil, fmt.Errorf("failed to get log analyzer: %w", err)
	}

	result, err := analyzer.AnalyzeLogDirectory(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze logs: %w", err)
	}

	log.Printf("[DeploymentService] Log analysis completed: success=%v", result.Success)
	return result, nil
}

// GetTask 获取部署任务
func (s *DeploymentServiceImpl) GetTask(taskUUID string) (*storage.DeploymentTask, error) {
	return s.deploymentStorage.GetTaskByUUID(taskUUID)
}

// ListRecentTasks 列出最近的部署任务
func (s *DeploymentServiceImpl) ListRecentTasks(limit int) ([]*storage.DeploymentTask, error) {
	return s.deploymentStorage.ListRecentTasks(limit)
}

// Helper function for result details JSON
func (d *DeploymentServiceImpl) marshalResultDetails(details map[string]interface{}) (json.RawMessage, error) {
	if details == nil {
		return nil, nil
	}
	return json.Marshal(details)
}
