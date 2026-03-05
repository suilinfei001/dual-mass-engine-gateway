package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"

	"github.com/gorilla/mux"
)

// LogAnalyzerService defines the interface for log analysis functionality
type LogAnalyzerService interface {
	FetchAndAnalyzeLogs(ctx context.Context, task *models.Task) ([]models.TaskResult, error)
}

type UserStorage interface {
	ValidatePassword(username, password string) (map[string]interface{}, error)
	CreateSession(userID int) (map[string]interface{}, error)
	GetSession(sessionID string) (map[string]interface{}, error)
	DeleteSession(sessionID string) error
}

type Server struct {
	Port             string
	storage          Storage
	userStorage      *storage.MySQLUserStorage
	resourceStorage  storage.ResourceStorage
	configStorage    *storage.MySQLConfigStorage
	client           *Client
	logAnalyzer      LogAnalyzerService
	mysqlTaskStorage *storage.MySQLTaskStorage
}

func (s *Server) UpdateClientURL(baseURL string) {
	if s.client != nil {
		s.client.BaseURL = baseURL
	}
}

type Storage interface {
	GetAllTasks() ([]models.TaskResponse, error)
	GetTaskByID(id int) (*models.TaskResponse, error)
	GetTasksByEventID(eventID int) ([]models.TaskResponse, error)
	DeleteAllTasks() error
	UpdateTaskURLsAndStatus(taskID int, requestURL, status string) error
}

type EventResponse struct {
	ID              int                    `json:"id"`
	EventID         string                 `json:"event_id"`
	EventType       string                 `json:"event_type"`
	EventStatus     string                 `json:"event_status"`
	Repository      string                 `json:"repository"`
	Branch          string                 `json:"branch"`
	TargetBranch    string                 `json:"target_branch,omitempty"`
	Author          string                 `json:"author,omitempty"`
	Pusher          string                 `json:"pusher,omitempty"`
	Action          string                 `json:"action,omitempty"`
	Payload         map[string]interface{} `json:"payload"`
	QualityChecks   []QualityCheckResponse `json:"quality_checks"`
	CurrentTaskName string                 `json:"current_task_name,omitempty"`
	LastTaskName    string                 `json:"last_task_name,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	ProcessedAt     string                 `json:"processed_at,omitempty"`
}

type QualityCheckResponse struct {
	ID              int      `json:"id"`
	CheckType       string   `json:"check_type"`
	CheckStatus     string   `json:"check_status"`
	Stage           string   `json:"stage"`
	StageOrder      int      `json:"stage_order"`
	CheckOrder      int      `json:"check_order"`
	StartedAt       *string  `json:"started_at,omitempty"`
	CompletedAt     *string  `json:"completed_at,omitempty"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	ErrorMessage    *string  `json:"error_message,omitempty"`
	Output          *string  `json:"output,omitempty"`
	RetryCount      int      `json:"retry_count"`
}

func NewServer(port string, storage Storage, userStorage *storage.MySQLUserStorage, client *Client) *Server {
	return &Server{
		Port:        port,
		storage:     storage,
		userStorage: userStorage,
		client:      client,
	}
}

func NewServerWithResources(
	port string,
	storage Storage,
	userStorage *storage.MySQLUserStorage,
	resourceStorage storage.ResourceStorage,
	configStorage *storage.MySQLConfigStorage,
	client *Client,
	logAnalyzer LogAnalyzerService,
	mysqlTaskStorage *storage.MySQLTaskStorage,
) *Server {
	return &Server{
		Port:             port,
		storage:          storage,
		userStorage:      userStorage,
		resourceStorage:  resourceStorage,
		configStorage:    configStorage,
		client:           client,
		logAnalyzer:      logAnalyzer,
		mysqlTaskStorage: mysqlTaskStorage,
	}
}

func (s *Server) Start() error {
	router := mux.NewRouter()

	authMiddleware := CreateAuthMiddleware(s.userStorage)
	adminMiddleware := CreateAdminMiddleware()

	router.HandleFunc("/api/events", s.handleGetEvents).Methods("GET")
	router.HandleFunc("/api/events/{id}", s.handleGetEvent).Methods("GET")
	router.HandleFunc("/api/events/{id}/tasks", s.handleGetEventTasks).Methods("GET")
	router.HandleFunc("/api/tasks", s.handleGetTasks).Methods("GET")
	router.HandleFunc("/api/tasks/{id}", s.handleGetTask).Methods("GET")
	router.HandleFunc("/api/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/api/config/event-receiver", s.handleGetEventReceiverConfig).Methods("GET")

	userHandler := NewUserHandler(s.userStorage)
	userHandler.RegisterRoutes(router, authMiddleware, adminMiddleware)

	resourceHandler := NewResourceHandler(s.resourceStorage, s.userStorage)
	resourceHandler.RegisterRoutes(router, authMiddleware)

	if s.configStorage != nil {
		consoleHandler := NewConsoleHandler(s.configStorage, s.userStorage, s.storage, s.resourceStorage, s, s.client, s.logAnalyzer)
		consoleHandler.SetMySQLTaskStorage(s.mysqlTaskStorage)
		consoleHandler.RegisterRoutes(router, authMiddleware, adminMiddleware)
	}

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("/app/static")))

	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	handler := corsMiddleware(router)

	return http.ListenAndServe(":"+s.Port, handler)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleGetEventReceiverConfig(w http.ResponseWriter, r *http.Request) {
	if s.configStorage == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Config storage not available",
		})
		return
	}

	ip, err := s.configStorage.GetEventReceiverIP()
	if err != nil || ip == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"configured":      false,
			"message":         "Event Receiver API not configured. Please configure in admin console.",
			"eventReceiverIP": "",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"configured":      true,
		"eventReceiverIP": ip,
	})
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.client.GetEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]EventResponse, len(events))
	for i, event := range events {
		tasks, err := s.storage.GetTasksByEventID(event.ID)
		if err == nil && len(tasks) > 0 {
			var currentTaskName, lastTaskName string
			for _, task := range tasks {
				if task.Status != "passed" && task.Status != "failed" &&
					task.Status != "cancelled" && task.Status != "skipped" &&
					task.Status != "timeout" {
					currentTaskName = task.TaskName
					break
				}
			}
			for j := len(tasks) - 1; j >= 0; j-- {
				if tasks[j].Status == "passed" || tasks[j].Status == "failed" ||
					tasks[j].Status == "cancelled" || tasks[j].Status == "skipped" ||
					tasks[j].Status == "timeout" {
					lastTaskName = tasks[j].TaskName
					break
				}
			}
			response[i] = convertEventToResponseWithTasks(event, currentTaskName, lastTaskName)
		} else {
			response[i] = convertEventToResponse(event)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

func (s *Server) handleGetEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := s.client.GetEvent(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if event == nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	tasks, err := s.storage.GetTasksByEventID(event.ID)
	var currentTaskName, lastTaskName string
	if err == nil && len(tasks) > 0 {
		for _, task := range tasks {
			if task.Status != "passed" && task.Status != "failed" &&
				task.Status != "cancelled" && task.Status != "skipped" &&
				task.Status != "timeout" {
				currentTaskName = task.TaskName
				break
			}
		}
		for j := len(tasks) - 1; j >= 0; j-- {
			if tasks[j].Status == "passed" || tasks[j].Status == "failed" ||
				tasks[j].Status == "cancelled" || tasks[j].Status == "skipped" ||
				tasks[j].Status == "timeout" {
				lastTaskName = tasks[j].TaskName
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    convertEventToResponseWithTasks(*event, currentTaskName, lastTaskName),
	})
}

func (s *Server) handleGetEventTasks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	tasks, err := s.storage.GetTasksByEventID(eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    tasks,
	})
}

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	tasks, err := s.storage.GetAllTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status != "" {
		filtered := make([]models.TaskResponse, 0)
		for _, task := range tasks {
			if strings.EqualFold(task.Status, status) {
				filtered = append(filtered, task)
			}
		}
		tasks = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    tasks,
	})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := s.storage.GetTaskByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task == nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    task,
	})
}

func convertEventToResponse(event Event) EventResponse {
	qualityChecks := make([]QualityCheckResponse, len(event.QualityChecks))
	for i, qc := range event.QualityChecks {
		qualityChecks[i] = QualityCheckResponse{
			ID:              qc.ID,
			CheckType:       qc.CheckType,
			CheckStatus:     qc.CheckStatus,
			Stage:           qc.Stage,
			StageOrder:      qc.StageOrder,
			CheckOrder:      qc.CheckOrder,
			DurationSeconds: qc.DurationSeconds,
			ErrorMessage:    qc.ErrorMessage,
			Output:          qc.Output,
			RetryCount:      qc.RetryCount,
		}
		if qc.StartedAt != "" {
			qualityChecks[i].StartedAt = &qc.StartedAt
		}
		if qc.CompletedAt != "" {
			qualityChecks[i].CompletedAt = &qc.CompletedAt
		}
	}

	action := ""
	if a, ok := event.Payload["action"].(string); ok {
		action = a
	} else if event.EventType == "push" {
		action = "push"
	}

	author := event.Author
	if author == "" {
		if p, ok := event.Payload["pusher"].(string); ok {
			author = p
		}
	}

	createdAt := event.CreatedAt
	if createdAt == "" {
		createdAt = event.ProcessedAt
	}

	return EventResponse{
		ID:            event.ID,
		EventID:       event.EventID,
		EventType:     event.EventType,
		EventStatus:   event.EventStatus,
		Repository:    event.Repository,
		Branch:        event.Branch,
		TargetBranch:  event.TargetBranch,
		Author:        author,
		Pusher:        event.Pusher,
		Action:        action,
		Payload:       event.Payload,
		QualityChecks: qualityChecks,
		CreatedAt:     createdAt,
		UpdatedAt:     event.UpdatedAt,
		ProcessedAt:   event.ProcessedAt,
	}
}

func convertEventToResponseWithTasks(event Event, currentTaskName, lastTaskName string) EventResponse {
	resp := convertEventToResponse(event)
	resp.CurrentTaskName = currentTaskName
	resp.LastTaskName = lastTaskName
	return resp
}
