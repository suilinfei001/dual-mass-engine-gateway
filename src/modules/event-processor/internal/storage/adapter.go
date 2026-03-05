package storage

import (
	"time"

	"github-hub/event-processor/internal/models"
)

type TaskStorageAdapter struct {
	*MySQLTaskStorage
}

func NewTaskStorageAdapter(storage *MySQLTaskStorage) *TaskStorageAdapter {
	return &TaskStorageAdapter{MySQLTaskStorage: storage}
}

func (a *TaskStorageAdapter) GetAllTasks() ([]models.TaskResponse, error) {
	tasks, err := a.ListTasks()
	if err != nil {
		return nil, err
	}

	var responses []models.TaskResponse
	for _, task := range tasks {
		// Fetch results from database
		results, err := a.MySQLTaskStorage.GetTaskResults(task.ID)
		if err == nil {
			task.Results = results
		}
		responses = append(responses, convertTaskToResponse(task))
	}
	return responses, nil
}

func (a *TaskStorageAdapter) GetTaskByID(id int) (*models.TaskResponse, error) {
	task, err := a.MySQLTaskStorage.GetTask(id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}

	// Fetch results from database
	results, err := a.MySQLTaskStorage.GetTaskResults(id)
	if err == nil {
		task.Results = results
	}

	resp := convertTaskToResponse(task)
	return &resp, nil
}

func (a *TaskStorageAdapter) GetTasksByEventID(eventID int) ([]models.TaskResponse, error) {
	tasks, err := a.MySQLTaskStorage.GetTasksByEventID(eventID)
	if err != nil {
		return nil, err
	}

	var responses []models.TaskResponse
	for _, task := range tasks {
		// Fetch results from database
		results, err := a.MySQLTaskStorage.GetTaskResults(task.ID)
		if err == nil {
			task.Results = results
		}
		responses = append(responses, convertTaskToResponse(task))
	}
	return responses, nil
}

func convertTaskToResponse(task *models.Task) models.TaskResponse {
	var startTime, endTime, createdAt, updatedAt string
	if task.StartTime != nil && !task.StartTime.Time.IsZero() {
		startTime = task.StartTime.Time.Format(time.RFC3339)
	}
	if task.EndTime != nil && !task.EndTime.Time.IsZero() {
		endTime = task.EndTime.Time.Format(time.RFC3339)
	}
	if !task.CreatedAt.IsZero() {
		createdAt = task.CreatedAt.Format(time.RFC3339)
	}
	if !task.UpdatedAt.IsZero() {
		updatedAt = task.UpdatedAt.Format(time.RFC3339)
	}

	var results []models.TaskResultResponse
	for _, r := range task.Results {
		results = append(results, models.TaskResultResponse{
			CheckType: r.CheckType,
			Result:    r.Result,
			Output:    r.Output,
			Extra:     r.Extra,
		})
	}

	errorMsg := ""
	if task.ErrorMessage != nil {
		errorMsg = *task.ErrorMessage
	}

	return models.TaskResponse{
		ID:           task.ID,
		TaskID:       task.TaskID,
		TaskName:     task.TaskName,
		EventID:      task.EventID,
		Stage:        task.Stage,
		StageOrder:   task.StageOrder,
		ExecuteOrder: task.ExecuteOrder,
		ResourceID:   task.ResourceID,
		RequestURL:   task.RequestURL,
		BuildID:      task.BuildID,
		LogFilePath:  task.LogFilePath,
		Status:       string(task.Status),
		StartTime:    &startTime,
		EndTime:      &endTime,
		ErrorMessage: errorMsg,
		Results:      results,
		Analyzing:    task.Analyzing,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func (a *TaskStorageAdapter) DeleteAllTasks() error {
	return a.MySQLTaskStorage.DeleteAllTasks()
}

func (a *TaskStorageAdapter) UpdateTaskURLsAndStatus(taskID int, requestURL, status string) error {
	task := &models.Task{
		ID:         taskID,
		RequestURL: requestURL,
		Status:     models.TaskStatus(status),
	}
	return a.MySQLTaskStorage.UpdateTaskURLsAndStatus(task)
}
