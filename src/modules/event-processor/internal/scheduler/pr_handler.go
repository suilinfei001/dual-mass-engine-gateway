package scheduler

import (
	"log"
	"time"

	"github-hub/event-processor/internal/api"
	"github-hub/event-processor/internal/models"
)

type PRHandler struct {
	client    *api.Client
	creator   *TaskCreator
	scheduler *SchedulerWithStorage
}

func NewPRHandler(client *api.Client, creator *TaskCreator, scheduler *SchedulerWithStorage) *PRHandler {
	return &PRHandler{
		client:    client,
		creator:   creator,
		scheduler: scheduler,
	}
}

func (h *PRHandler) HandlePREvent(event *api.Event) error {
	prAction := h.getPRAction(event)
	if prAction == "" {
		log.Printf("Event %d is not a PR event or has no action", event.ID)
		return nil
	}

	switch prAction {
	case "opened":
		return h.handlePROpened(event)
	case "synchronize":
		return h.handlePRSynchronize(event)
	default:
		log.Printf("PR action %s not handled for event %d", prAction, event.ID)
		return nil
	}
}

func (h *PRHandler) getPRAction(event *api.Event) string {
	if event.EventType != "pull_request" {
		return ""
	}

	if action, ok := event.Payload["action"].(string); ok {
		return action
	}

	if action, ok := event.Payload["pr_action"].(string); ok {
		return action
	}

	return ""
}

func (h *PRHandler) getPRURL(event *api.Event) string {
	if prURL, ok := event.Payload["pr_url"].(string); ok {
		return prURL
	}
	return ""
}

func (h *PRHandler) handlePROpened(event *api.Event) error {
	log.Printf("Handling PR opened event %d", event.ID)

	existingTasks, err := h.scheduler.GetTasksByEventID(event.ID)
	if err == nil && len(existingTasks) > 0 {
		log.Printf("Tasks already exist for PR event %d, skipping", event.ID)
		return nil
	}

	task, err := h.creator.CreateFirstTask(event)
	if err != nil {
		log.Printf("Failed to create first task for event %d: %v", event.ID, err)
		if task != nil {
			if createErr := h.scheduler.storage.CreateTask(task); createErr != nil {
				log.Printf("Failed to save failed task: %v", createErr)
			}
		}
		return nil
	}

	return h.scheduler.storage.CreateTask(task)
}

func (h *PRHandler) handlePRSynchronize(event *api.Event) error {
	log.Printf("Handling PR synchronize event %d", event.ID)

	prURL := h.getPRURL(event)
	if prURL == "" {
		log.Printf("No pr_url found for PR event %d", event.ID)
		return nil
	}

	events, err := h.client.GetEvents()
	if err != nil {
		return err
	}

	var relatedEvents []*api.Event
	for i := range events {
		e := &events[i]
		if e.EventType == "pull_request" {
			ePRURL := h.getPRURL(e)
			if ePRURL == prURL {
				relatedEvents = append(relatedEvents, e)
			}
		}
	}

	targetEvent, eventsToCancel, eventsToComplete := h.filterEvents(relatedEvents, event)

	for _, e := range eventsToComplete {
		if err := h.completeEvent(e); err != nil {
			log.Printf("Failed to complete event %d: %v", e.ID, err)
		}
	}

	for _, e := range eventsToCancel {
		if err := h.cancelEvent(e); err != nil {
			log.Printf("Failed to cancel event %d: %v", e.ID, err)
		}
	}

	if targetEvent != nil && targetEvent.ID == event.ID {
		if event.EventStatus == "pending" {
			existingTasks, err := h.scheduler.GetTasksByEventID(event.ID)
			if err != nil || len(existingTasks) == 0 {
				task, err := h.creator.CreateFirstTask(event)
				if err != nil {
					return err
				}
				return h.scheduler.storage.CreateTask(task)
			}
		}
	}

	return nil
}

func (h *PRHandler) filterEvents(relatedEvents []*api.Event, currentEvent *api.Event) (*api.Event, []*api.Event, []*api.Event) {
	var pendingEvents []*api.Event
	var processingEvents []*api.Event
	var eventsToComplete []*api.Event
	var eventsToCancel []*api.Event

	for _, e := range relatedEvents {
		switch e.EventStatus {
		case "completed", "failed", "cancelled":
			continue
		case "pending":
			pendingEvents = append(pendingEvents, e)
		case "processing":
			processingEvents = append(processingEvents, e)
		}
	}

	var targetEvent *api.Event

	if len(processingEvents) == 1 && len(pendingEvents) == 0 {
		targetEvent = processingEvents[0]
	} else if len(processingEvents) > 0 {
		latestProcessing := processingEvents[0]
		for _, e := range processingEvents {
			if e.CreatedAt > latestProcessing.CreatedAt {
				latestProcessing = e
			}
		}

		if len(pendingEvents) > 0 {
			latestPending := pendingEvents[0]
			for _, e := range pendingEvents {
				if e.CreatedAt > latestPending.CreatedAt {
					latestPending = e
				}
			}

			if latestPending.CreatedAt > latestProcessing.CreatedAt {
				targetEvent = latestPending
				for _, e := range processingEvents {
					eventsToCancel = append(eventsToCancel, e)
				}
				for _, e := range pendingEvents {
					if e.ID != targetEvent.ID {
						eventsToComplete = append(eventsToComplete, e)
					}
				}
			} else {
				targetEvent = latestProcessing
				for _, e := range pendingEvents {
					eventsToComplete = append(eventsToComplete, e)
				}
				for _, e := range processingEvents {
					if e.ID != targetEvent.ID {
						eventsToCancel = append(eventsToCancel, e)
					}
				}
			}
		} else {
			targetEvent = latestProcessing
			for _, e := range processingEvents {
				if e.ID != targetEvent.ID {
					eventsToCancel = append(eventsToCancel, e)
				}
			}
		}
	} else if len(pendingEvents) > 0 {
		latestPending := pendingEvents[0]
		for _, e := range pendingEvents {
			if e.CreatedAt > latestPending.CreatedAt {
				latestPending = e
			}
		}
		targetEvent = latestPending
		for _, e := range pendingEvents {
			if e.ID != targetEvent.ID {
				eventsToComplete = append(eventsToComplete, e)
			}
		}
	}

	return targetEvent, eventsToCancel, eventsToComplete
}

func (h *PRHandler) completeEvent(event *api.Event) error {
	tasks, err := h.scheduler.storage.GetTasksByEventID(event.ID)
	if err == nil {
		for _, task := range tasks {
			if task.Status == models.TaskStatusPending || task.Status == models.TaskStatusRunning {
				task.MarkCancelled("Superseded by newer PR commit")
				if err := h.scheduler.storage.UpdateTask(task); err != nil {
					log.Printf("Failed to cancel task %d for event %d: %v", task.ID, event.ID, err)
				} else {
					log.Printf("Cancelled task %d for event %d (superseded)", task.ID, event.ID)
				}
			}
		}
	}

	now := time.Now()
	return h.client.UpdateEventStatus(event.ID, "completed", now.Format(time.RFC3339))
}

func (h *PRHandler) cancelEvent(event *api.Event) error {
	tasks, err := h.scheduler.storage.GetTasksByEventID(event.ID)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Status == models.TaskStatusRunning {
			// Cancellation now handled through TaskExecutionService using BuildID
			if task.BuildID > 0 {
				log.Printf("Task %d has build_id=%d, will be cancelled by executor service", task.ID, task.BuildID)
			}

			if err := h.scheduler.CancelTask(task, "PR synchronized with newer commit"); err != nil {
				log.Printf("Failed to cancel running task %d: %v", task.ID, err)
			} else {
				log.Printf("Cancelled running task %d for event %d", task.ID, event.ID)
			}
		} else if task.Status == models.TaskStatusPending {
			if err := h.scheduler.CancelTask(task, "PR synchronized with newer commit"); err != nil {
				log.Printf("Failed to cancel pending task %d: %v", task.ID, err)
			} else {
				log.Printf("Cancelled pending task %d for event %d", task.ID, event.ID)
			}
		}
	}

	now := time.Now()
	if err := h.client.UpdateEventStatus(event.ID, "cancelled", now.Format(time.RFC3339)); err != nil {
		log.Printf("Failed to update event status to cancelled: %v", err)
	}

	return nil
}

func (h *PRHandler) IsPREvent(event *api.Event) bool {
	return event.EventType == "pull_request"
}
