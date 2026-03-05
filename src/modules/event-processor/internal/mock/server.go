package mock

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MockServer struct {
	port       int
	server     *http.Server
	tasks      map[string]*MockTask
	tasksMu    sync.RWMutex
	taskTimers map[string]*time.Timer
	timersMu   sync.Mutex
}

type MockTask struct {
	TaskID       string       `json:"task_id"`
	TaskName     string       `json:"task_name"`
	RequestURL   string       `json:"request_url"`
	StartTime    string       `json:"start_time"`
	EndTime      string       `json:"end_time"`
	Status       string       `json:"status"`
	ExecuteOrder int          `json:"execute_order"`
	ErrorMessage string       `json:"error_message"`
	Results      []MockResult `json:"results"`
	EventID      int          `json:"event_id"`
}

type MockResult struct {
	CheckType string                 `json:"check_type"`
	Result    string                 `json:"result"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type TaskRequest struct {
	EventID int `json:"event_id"`
}

func NewMockServer(port int) *MockServer {
	return &MockServer{
		port:       port,
		tasks:      make(map[string]*MockTask),
		taskTimers: make(map[string]*time.Timer),
	}
}

func (s *MockServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/mock/basic-ci", s.handleBasicCI)
	mux.HandleFunc("/mock/basic-ci/cancel", s.handleCancel)
	mux.HandleFunc("/mock/deployment", s.handleDeployment)
	mux.HandleFunc("/mock/deployment/cancel", s.handleCancel)
	mux.HandleFunc("/mock/api-test", s.handleAPITest)
	mux.HandleFunc("/mock/api-test/cancel", s.handleCancel)
	mux.HandleFunc("/mock/module-e2e", s.handleModuleE2E)
	mux.HandleFunc("/mock/module-e2e/cancel", s.handleCancel)
	mux.HandleFunc("/mock/agent-e2e", s.handleAgentE2E)
	mux.HandleFunc("/mock/agent-e2e/cancel", s.handleCancel)
	mux.HandleFunc("/mock/ai-e2e", s.handleAIE2E)
	mux.HandleFunc("/mock/ai-e2e/cancel", s.handleCancel)
	mux.HandleFunc("/mock/status/", s.handleStatus)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Printf("Mock server starting on port %d", s.port)
	return s.server.ListenAndServe()
}

func (s *MockServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *MockServer) handleBasicCI(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "basic_ci_all",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/basic-ci?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 1,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 3*time.Second)

	s.respondJSON(w, task)
}

func (s *MockServer) handleDeployment(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "deployment_deployment",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/deployment?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 2,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 1*time.Minute)

	s.respondJSON(w, task)
}

func (s *MockServer) handleAPITest(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "specialized_tests_api_test",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/api-test?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 3,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 2*time.Second)

	s.respondJSON(w, task)
}

func (s *MockServer) handleModuleE2E(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "specialized_tests_module_e2e",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/module-e2e?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 4,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 2*time.Second)

	s.respondJSON(w, task)
}

func (s *MockServer) handleAgentE2E(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "specialized_tests_agent_e2e",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/agent-e2e?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 5,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 2*time.Second)

	s.respondJSON(w, task)
}

func (s *MockServer) handleAIE2E(w http.ResponseWriter, r *http.Request) {
	eventID := s.getEventID(r)
	taskID := uuid.New().String()

	task := &MockTask{
		TaskID:       taskID,
		TaskName:     "specialized_tests_ai_e2e",
		RequestURL:   fmt.Sprintf("http://localhost:%d/mock/ai-e2e?event_id=%d", s.port, eventID),
		StartTime:    time.Now().Format(time.RFC3339),
		Status:       "running",
		ExecuteOrder: 6,
		EventID:      eventID,
	}

	s.tasksMu.Lock()
	s.tasks[taskID] = task
	s.tasksMu.Unlock()

	s.scheduleTaskCompletion(taskID, 2*time.Second)

	s.respondJSON(w, task)
}

func (s *MockServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Path[len("/mock/status/"):]

	s.tasksMu.RLock()
	task, exists := s.tasks[taskID]
	s.tasksMu.RUnlock()

	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	s.respondJSON(w, task)
}

func (s *MockServer) handleCancel(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "task_id required", http.StatusBadRequest)
		return
	}

	s.tasksMu.Lock()
	task, exists := s.tasks[taskID]
	if exists && task.Status == "running" {
		task.Status = "cancelled"
		task.EndTime = time.Now().Format(time.RFC3339)
		task.ErrorMessage = "Task cancelled by request"
	}
	s.tasksMu.Unlock()

	s.timersMu.Lock()
	if timer, exists := s.taskTimers[taskID]; exists {
		timer.Stop()
		delete(s.taskTimers, taskID)
	}
	s.timersMu.Unlock()

	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	s.respondJSON(w, map[string]string{
		"status":  "cancelled",
		"task_id": taskID,
	})
}

func (s *MockServer) scheduleTaskCompletion(taskID string, delay time.Duration) {
	timer := time.AfterFunc(delay, func() {
		s.tasksMu.Lock()
		task, exists := s.tasks[taskID]
		if !exists || task.Status != "running" {
			s.tasksMu.Unlock()
			return
		}

		task.Status = "pass"
		task.EndTime = time.Now().Format(time.RFC3339)
		task.Results = s.generateResults(task.TaskName)
		s.tasksMu.Unlock()

		log.Printf("Mock task %s completed with status: %s", taskID, task.Status)
	})

	s.timersMu.Lock()
	s.taskTimers[taskID] = timer
	s.timersMu.Unlock()
}

func (s *MockServer) generateResults(taskName string) []MockResult {
	switch taskName {
	case "basic_ci_all":
		return []MockResult{
			{CheckType: "compilation", Result: "pass"},
			{CheckType: "code_lint", Result: "pass"},
			{CheckType: "security_scan", Result: "pass"},
			{CheckType: "unit_test", Result: "pass", Extra: map[string]interface{}{"score": 95}},
		}
	case "deployment_deployment":
		return []MockResult{
			{CheckType: "deployment", Result: "pass", Extra: map[string]interface{}{
				"node_ip":   "192.168.1.100",
				"node_port": "22",
				"node_user": "deployer",
			}},
		}
	case "specialized_tests_api_test":
		return []MockResult{
			{CheckType: "api_test", Result: "pass"},
		}
	case "specialized_tests_module_e2e":
		return []MockResult{
			{CheckType: "module_e2e", Result: "pass"},
		}
	case "specialized_tests_agent_e2e":
		return []MockResult{
			{CheckType: "agent_e2e", Result: "pass"},
		}
	case "specialized_tests_ai_e2e":
		return []MockResult{
			{CheckType: "ai_e2e", Result: "pass"},
		}
	default:
		return []MockResult{}
	}
}

func (s *MockServer) getEventID(r *http.Request) int {
	var eventID int
	fmt.Sscanf(r.URL.Query().Get("event_id"), "%d", &eventID)
	return eventID
}

func (s *MockServer) respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *MockServer) GetTaskStatus(taskID string) (*MockTask, bool) {
	s.tasksMu.RLock()
	defer s.tasksMu.RUnlock()
	task, exists := s.tasks[taskID]
	return task, exists
}
