package scheduler

import (
	"fmt"
	"log"
	"strings"

	"github-hub/event-processor/internal/ai"
	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/executor"
	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"
)

const (
	MockBaseURL = "http://localhost:8090/mock"
)

var ErrNoResourceMatched = fmt.Errorf("no matching resource found for this task")

// ResourceInfo holds resource ID and request URL
type ResourceInfo struct {
	ResourceID int
	RequestURL string
	ChartURL   string // For deployment task, chart URL from basic_ci_all result
}

type AIMatcherInterface interface {
	MatchResource(req *ai.MatchRequest) (*ai.MatchResult, error)
	MatchResourceStream(req *ai.MatchRequest, callback func(string)) (*ai.MatchResult, error)
}

type TaskCreator struct {
	client          *api.Client
	storage         storage.TaskStorage
	resourceStorage storage.ResourceStorage
	aiMatcher       AIMatcherInterface
}

func NewTaskCreator(client *api.Client, store storage.TaskStorage, resourceStorage storage.ResourceStorage, aiMatcher AIMatcherInterface) *TaskCreator {
	return &TaskCreator{
		client:          client,
		storage:         store,
		resourceStorage: resourceStorage,
		aiMatcher:       aiMatcher,
	}
}

func (tc *TaskCreator) CreateFirstTask(event *api.Event) (*models.Task, error) {
	def := models.GetTaskDefinitionByOrder(1)
	if def == nil {
		return nil, fmt.Errorf("failed to get task definition for order 1")
	}

	resourceInfo, err := tc.getResourceURL(def.TaskName, event)
	if err != nil {
		task := models.NewBasicCITask(event.ID, def.ExecuteOrder, "")
		task.MarkNoResource(err.Error())
		return task, nil
	}
	task := models.NewBasicCITask(event.ID, def.ExecuteOrder, resourceInfo.RequestURL)
	task.ResourceID = resourceInfo.ResourceID
	return task, nil
}

func (tc *TaskCreator) CreateNextTask(eventID int, currentExecuteOrder int, event *api.Event, parentTestbedInfo ...*models.Task) (*models.Task, error) {
	nextOrder := models.GetNextExecuteOrder(currentExecuteOrder)
	if nextOrder == -1 {
		return nil, fmt.Errorf("no more tasks to create")
	}

	def := models.GetTaskDefinitionByOrder(nextOrder)
	if def == nil {
		return nil, fmt.Errorf("failed to get task definition for order %d", nextOrder)
	}

	log.Printf("[CreateNextTask] Creating task: eventID=%d, taskName=%s, currentOrder=%d, nextOrder=%d",
		eventID, def.TaskName, currentExecuteOrder, nextOrder)

	resourceInfo, err := tc.getResourceURL(def.TaskName, event)
	if err != nil {
		log.Printf("[CreateNextTask] Failed to get resource URL: %v", err)
		task := models.NewTask(
			eventID, def.TaskName, def.CheckType, def.Stage,
			def.StageOrder, def.CheckOrder, def.ExecuteOrder, "",
		)
		task.MarkNoResource(err.Error())
		return task, nil
	}

	log.Printf("[CreateNextTask] Got resource info: taskName=%s, resourceID=%d, chartURL=%s",
		def.TaskName, resourceInfo.ResourceID, resourceInfo.ChartURL)

	var task *models.Task
	if def.TaskName == "basic_ci_all" {
		task = models.NewBasicCITask(eventID, def.ExecuteOrder, resourceInfo.RequestURL)
		task.ResourceID = resourceInfo.ResourceID
		return task, nil
	}

	task = models.NewTask(
		eventID, def.TaskName, def.CheckType, def.Stage,
		def.StageOrder, def.CheckOrder, def.ExecuteOrder, resourceInfo.RequestURL,
	)
	task.ResourceID = resourceInfo.ResourceID
	// 对于 deployment 任务，设置 chart_url
	if def.TaskName == "deployment_deployment" {
		task.ChartURL = resourceInfo.ChartURL
		log.Printf("[CreateNextTask] Set task.ChartURL from resourceInfo: taskID=%d, chartURL=%s",
			task.ID, task.ChartURL)
	}

	// 如果有父任务传递了 testbed 信息，后续任务继承
	if len(parentTestbedInfo) > 0 && parentTestbedInfo[0] != nil {
		parentTask := parentTestbedInfo[0]
		task.TestbedUUID = parentTask.TestbedUUID
		task.TestbedIP = parentTask.TestbedIP
		task.SSHUser = parentTask.SSHUser
		task.SSHPassword = parentTask.SSHPassword
		task.AllocationUUID = parentTask.AllocationUUID
	}

	return task, nil
}

func (tc *TaskCreator) CreateTaskForEvent(event *api.Event, executeOrder int) (*models.Task, error) {
	def := models.GetTaskDefinitionByOrder(executeOrder)
	if def == nil {
		return nil, fmt.Errorf("failed to get task definition for order %d", executeOrder)
	}

	resourceInfo, err := tc.getResourceURL(def.TaskName, event)
	if err != nil {
		task := models.NewTask(
			event.ID, def.TaskName, def.CheckType, def.Stage,
			def.StageOrder, def.CheckOrder, def.ExecuteOrder, "",
		)
		task.MarkNoResource(err.Error())
		return task, nil
	}

	var task *models.Task
	if def.TaskName == "basic_ci_all" {
		task = models.NewBasicCITask(event.ID, def.ExecuteOrder, resourceInfo.RequestURL)
	} else {
		task = models.NewTask(
			event.ID, def.TaskName, def.CheckType, def.Stage,
			def.StageOrder, def.CheckOrder, def.ExecuteOrder, resourceInfo.RequestURL,
		)
	}
	task.ResourceID = resourceInfo.ResourceID
	// 对于 deployment 任务，设置 chart_url
	if def.TaskName == "deployment_deployment" {
		task.ChartURL = resourceInfo.ChartURL
	}

	return task, nil
}

func (tc *TaskCreator) getResourceURL(taskName string, event *api.Event) (*ResourceInfo, error) {
	// deployment_deployment 任务特殊处理：不需要 AI 匹配，直接使用唯一的 deployment 资源
	if taskName == "deployment_deployment" {
		return tc.getDeploymentResource(event)
	}

	// 如果没有配置 resourceStorage 或 aiMatcher，使用 mock URL（用于测试）
	if tc.resourceStorage == nil || tc.aiMatcher == nil {
		mockURL, err := tc.getMockURLWithEventID(taskName, event.ID)
		if err != nil {
			return nil, err
		}
		return &ResourceInfo{RequestURL: mockURL}, nil
	}

	resources, err := tc.resourceStorage.GetAllResources()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get resources: %v", ErrNoResourceMatched, err)
	}

	// 没有可用资源时返回错误，任务将被标记为 no-resource
	if len(resources) == 0 {
		return nil, fmt.Errorf("%w: no resources available in storage", ErrNoResourceMatched)
	}

	eventDetail := map[string]interface{}{
		"event_type": event.EventType,
		"repository": event.Repository,
		"branch":     event.Branch,
		"task_name":  taskName,
		"commit_sha": event.CommitSHA,
		"pusher":     event.Pusher,
		"author":     event.Author,
		"payload":    event.Payload,
	}

	req := &ai.MatchRequest{
		TaskName:    taskName, // 直接使用 task_name，如 "basic_ci_all"
		EventDetail: eventDetail,
		Resources:   resources,
	}

	// Set the system prompt for AI matching
	if tc.aiMatcher != nil {
		if matcher, ok := tc.aiMatcher.(*ai.AIMatcher); ok {
			req.SystemPrompt = matcher.GetDefaultSystemPrompt()
		}
	}

	matchResult, err := tc.aiMatcher.MatchResource(req)
	if err != nil {
		return nil, fmt.Errorf("%w: AI matching failed for task %s: %v", ErrNoResourceMatched, taskName, err)
	}

	if matchResult == nil || matchResult.ResourceID == 0 {
		return nil, fmt.Errorf("%w: no resource matched for task %s", ErrNoResourceMatched, taskName)
	}

	matchedResource := tc.findResourceByID(resources, matchResult.ResourceID)
	if matchedResource == nil {
		return nil, fmt.Errorf("%w: matched resource not found", ErrNoResourceMatched)
	}

	// 强制校验仓库匹配（AI 可能返回错误结果）
	if !tc.isRepoMatched(matchedResource.RepoPath, event.Repository) {
		log.Printf("[TaskCreator] Resource %d repo_path '%s' does not match event repository '%s'",
			matchedResource.ID, matchedResource.RepoPath, event.Repository)
		return nil, fmt.Errorf("%w: resource repo_path '%s' does not match event repository '%s'",
			ErrNoResourceMatched, matchedResource.RepoPath, event.Repository)
	}

	requestURL := tc.buildRequestURL(matchedResource, event)
	return &ResourceInfo{
		ResourceID: matchedResource.ID,
		RequestURL: requestURL,
	}, nil
}

func (tc *TaskCreator) findResourceByID(resources []*models.ExecutableResource, id int) *models.ExecutableResource {
	for _, r := range resources {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// isRepoMatched 检查资源的 repo_path 是否与事件仓库匹配
// repo_path 可以是：
// - "all" 或 "*" 表示匹配所有仓库
// - "org/repo" 精确匹配
// - "org/repo/subdir" 前缀匹配（事件的仓库应为 org/repo）
func (tc *TaskCreator) isRepoMatched(repoPath, eventRepository string) bool {
	if repoPath == "" {
		return false
	}

	// 特殊值：匹配所有仓库
	if repoPath == "all" || repoPath == "*" {
		return true
	}

	// 提取 repo_path 中的仓库部分（去掉子目录）
	// 例如：kweaver-ai/adp/context-loader/agent-retrieval -> kweaver-ai/adp
	parts := strings.Split(repoPath, "/")
	if len(parts) < 2 {
		// repo_path 格式不正确，无法匹配
		return false
	}

	// repo_path 的仓库部分（前两段）
	resourceRepo := parts[0] + "/" + parts[1]

	// 精确匹配
	return resourceRepo == eventRepository
}

func (tc *TaskCreator) buildRequestURL(resource *models.ExecutableResource, event *api.Event) string {
	// If the resource has Azure configuration, build an internal Azure URL
	if resource.Organization != "" && resource.Project != "" && resource.PipelineID > 0 {
		// Build internal Azure URL: azure://devops.aishu.cn/{org}/{project}/pipeline/{id}
		return executor.BuildAzureURL("devops.aishu.cn", resource.Organization, resource.Project, resource.PipelineID)
	}

	// Otherwise, use the microservice or pod name for backward compatibility
	baseURL := ""
	if resource.MicroserviceName != "" {
		baseURL = fmt.Sprintf("http://%s", resource.MicroserviceName)
	} else if resource.PodName != "" {
		baseURL = fmt.Sprintf("http://%s", resource.PodName)
	}

	if baseURL == "" {
		baseURL = MockBaseURL
	}

	return fmt.Sprintf("%s?event_id=%d&repo=%s&branch=%s", baseURL, event.ID, event.Repository, event.Branch)
}

func (tc *TaskCreator) getMockURL(taskName string, eventID int) string {
	url, _ := tc.getMockURLWithEventID(taskName, eventID)
	return url
}

func (tc *TaskCreator) getMockURLWithEventID(taskName string, eventID int) (string, error) {
	var requestPath string

	switch taskName {
	case "basic_ci_all":
		requestPath = fmt.Sprintf("%s/basic-ci?event_id=%d", MockBaseURL, eventID)
	case "deployment_deployment":
		requestPath = fmt.Sprintf("%s/deployment?event_id=%d", MockBaseURL, eventID)
	case "specialized_tests_api_test":
		requestPath = fmt.Sprintf("%s/api-test?event_id=%d", MockBaseURL, eventID)
	case "specialized_tests_module_e2e":
		requestPath = fmt.Sprintf("%s/module-e2e?event_id=%d", MockBaseURL, eventID)
	case "specialized_tests_agent_e2e":
		requestPath = fmt.Sprintf("%s/agent-e2e?event_id=%d", MockBaseURL, eventID)
	case "specialized_tests_ai_e2e":
		requestPath = fmt.Sprintf("%s/ai-e2e?event_id=%d", MockBaseURL, eventID)
	default:
		requestPath = fmt.Sprintf("%s/unknown?event_id=%d", MockBaseURL, eventID)
	}

	return requestPath, nil
}

// getDeploymentResource 获取 deployment 任务所需的资源信息
// 对于 deployment_deployment 任务，不需要 AI 匹配，直接使用唯一的 deployment 资源
// chart_url 从 basic_ci_all 任务结果中获取
func (tc *TaskCreator) getDeploymentResource(event *api.Event) (*ResourceInfo, error) {
	eventID := event.ID
	log.Printf("[getDeploymentResource] Starting for event %d", eventID)

	if tc.resourceStorage == nil {
		return nil, fmt.Errorf("%w: resourceStorage not initialized", ErrNoResourceMatched)
	}

	// 获取 deployment 类型的资源
	resources, err := tc.resourceStorage.GetAllResources()
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get resources: %v", ErrNoResourceMatched, err)
	}

	var deploymentResource *models.ExecutableResource
	for _, r := range resources {
		log.Printf("[getDeploymentResource] Checking resource: type=%s, name=%s", r.ResourceType, r.ResourceName)
		if r.ResourceType == models.ResourceTypeDeployment {
			deploymentResource = r
			break
		}
	}

	if deploymentResource == nil {
		return nil, fmt.Errorf("%w: no deployment resource available", ErrNoResourceMatched)
	}

	log.Printf("[getDeploymentResource] Found deployment resource: id=%d, name=%s", deploymentResource.ID, deploymentResource.ResourceName)

	// 从 basic_ci_all 任务结果中获取 chart_url
	chartURL, err := tc.getChartURLFromBasicCI(eventID)
	if err != nil {
		log.Printf("[getDeploymentResource] Failed to get chart URL from basic_ci_all: %v", err)
		log.Printf("[getDeploymentResource] Querying database to debug chart_url issue for event %d", eventID)

		// 调试：查询数据库中的任务和结果
		if tc.storage != nil {
			tasks, dbgErr := tc.storage.GetTasksByEventID(eventID)
			if dbgErr != nil {
				log.Printf("[getDeploymentResource] Debug: failed to get tasks: %v", dbgErr)
			} else {
				log.Printf("[getDeploymentResource] Debug: found %d tasks for event %d", len(tasks), eventID)
				for _, t := range tasks {
					log.Printf("[getDeploymentResource] Debug: task id=%d, name=%s, status=%s", t.ID, t.TaskName, t.Status)
					if t.TaskName == "basic_ci_all" {
						results, resErr := tc.storage.GetTaskResults(t.ID)
						if resErr != nil {
							log.Printf("[getDeploymentResource] Debug: failed to get results: %v", resErr)
						} else {
							log.Printf("[getDeploymentResource] Debug: found %d results for basic_ci_all", len(results))
							for _, r := range results {
								log.Printf("[getDeploymentResource] Debug: result check_type=%s, result=%s, extra=%v", r.CheckType, r.Result, r.Extra)
							}
						}
					}
				}
			}
		}

		chartURL = ""
	} else {
		log.Printf("[getDeploymentResource] Got chart URL: %s", chartURL)
	}

	// 构建 Azure URL
	requestURL := tc.buildRequestURL(deploymentResource, event)

	return &ResourceInfo{
		ResourceID: deploymentResource.ID,
		RequestURL: requestURL,
		ChartURL:   chartURL,
	}, nil
}

// getChartURLFromBasicCI 从 basic_ci_all 任务结果中获取 chart_url
func (tc *TaskCreator) getChartURLFromBasicCI(eventID int) (string, error) {
	// 获取该 event 的 basic_ci_all 任务
	tasks, err := tc.storage.GetTasksByEventID(eventID)
	if err != nil {
		return "", fmt.Errorf("failed to get tasks for event %d: %w", eventID, err)
	}

	// 查找 basic_ci_all 任务
	var basicCITask *models.Task
	for _, task := range tasks {
		if task.TaskName == "basic_ci_all" {
			basicCITask = task
			break
		}
	}

	if basicCITask == nil {
		return "", fmt.Errorf("basic_ci_all task not found for event %d", eventID)
	}

	log.Printf("[getChartURLFromBasicCI] Found basic_ci_all task (id=%d) for event %d", basicCITask.ID, eventID)

	// 获取任务结果
	results, err := tc.storage.GetTaskResults(basicCITask.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get results for basic_ci_all task: %w", err)
	}

	log.Printf("[getChartURLFromBasicCI] Got %d results for basic_ci_all task", len(results))

	// 查找 chart 值（通常在 compilation 检查结果的 extra 中）
	for _, result := range results {
		log.Printf("[getChartURLFromBasicCI] Checking result: check_type=%s, result=%s, extra=%v", result.CheckType, result.Result, result.Extra)
		if result.CheckType == "compilation" && result.Extra != nil {
			if chart, ok := result.Extra["chart"].(string); ok && chart != "" {
				log.Printf("[getChartURLFromBasicCI] Found chart URL: %s", chart)
				return chart, nil
			}
		}
	}

	return "", fmt.Errorf("chart URL not found in basic_ci_all results")
}

func (tc *TaskCreator) ShouldCreateNextTask(eventID int, currentTask *models.Task) bool {
	log.Printf("ShouldCreateNextTask called: eventID=%d, taskStatus=%s, IsSuccessful=%v, executeOrder=%d",
		eventID, currentTask.Status, currentTask.IsSuccessful(), currentTask.ExecuteOrder)

	if !currentTask.IsSuccessful() {
		return false
	}

	nextOrder := models.GetNextExecuteOrder(currentTask.ExecuteOrder)
	if nextOrder == -1 {
		return false
	}

	existingTasks, err := tc.storage.GetTasksByEventID(eventID)
	if err != nil {
		return false
	}

	for _, t := range existingTasks {
		if t.ExecuteOrder == nextOrder {
			return false
		}
	}

	log.Printf("ShouldCreateNextTask returning true: nextOrder=%d", nextOrder)
	return true
}

func (tc *TaskCreator) IsLastTask(executeOrder int) bool {
	result := executeOrder >= models.MaxExecuteOrder
	log.Printf("IsLastTask called: executeOrder=%d, MaxExecuteOrder=%d, result=%v", executeOrder, models.MaxExecuteOrder, result)
	return result
}
