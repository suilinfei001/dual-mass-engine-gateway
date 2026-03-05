package scheduler

import (
	"fmt"
	"log"

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
	ResourceID  int
	RequestURL  string
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

func (tc *TaskCreator) CreateNextTask(eventID int, currentExecuteOrder int, event *api.Event) (*models.Task, error) {
	nextOrder := models.GetNextExecuteOrder(currentExecuteOrder)
	if nextOrder == -1 {
		return nil, fmt.Errorf("no more tasks to create")
	}

	def := models.GetTaskDefinitionByOrder(nextOrder)
	if def == nil {
		return nil, fmt.Errorf("failed to get task definition for order %d", nextOrder)
	}

	resourceInfo, err := tc.getResourceURL(def.TaskName, event)
	if err != nil {
		task := models.NewTask(
			eventID, def.TaskName, def.CheckType, def.Stage,
			def.StageOrder, def.CheckOrder, def.ExecuteOrder, "",
		)
		task.MarkNoResource(err.Error())
		return task, nil
	}

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

	return task, nil
}

func (tc *TaskCreator) getResourceURL(taskName string, event *api.Event) (*ResourceInfo, error) {
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
