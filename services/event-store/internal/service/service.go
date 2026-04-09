// Package service provides business logic for event store service.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/quality-gateway/event-store/internal/storage"
	"github.com/quality-gateway/shared/pkg/logger"
)

// EventStoreService 事件存储服务
type EventStoreService struct {
	eventStorage       *storage.EventStorage
	qualityCheckStorage *storage.QualityCheckStorage
	logger             *logger.Logger
}

// NewEventStoreService 创建事件存储服务
func NewEventStoreService(
	es *storage.EventStorage,
	qs *storage.QualityCheckStorage,
	log *logger.Logger,
) *EventStoreService {
	return &EventStoreService{
		eventStorage:        es,
		qualityCheckStorage: qs,
		logger:             log,
	}
}

// CreateEvent 创建新事件
func (s *EventStoreService) CreateEvent(ctx context.Context, event *storage.Event) error {
	// 生成 UUID
	if event.UUID == "" {
		event.UUID = uuid.New().String()
	}

	// 设置默认状态
	if event.Status == "" {
		event.Status = storage.EventStatusPending
	}

	// 设置接收时间
	if event.ReceivedAt.IsZero() {
		event.ReceivedAt = time.Now()
	}

	s.logger.Info("Creating event",
		logger.String("uuid", event.UUID),
		logger.String("type", string(event.EventType)),
		logger.String("repo", event.RepoName),
	)

	return s.eventStorage.Create(ctx, event)
}

// GetEvent 获取单个事件
func (s *EventStoreService) GetEvent(ctx context.Context, eventUUID string) (*storage.Event, error) {
	return s.eventStorage.GetByUUID(ctx, eventUUID)
}

// ListEvents 列出事件
func (s *EventStoreService) ListEvents(ctx context.Context, filter *storage.EventFilter) ([]*storage.Event, error) {
	return s.eventStorage.List(ctx, filter)
}

// UpdateEventStatus 更新事件状态
func (s *EventStoreService) UpdateEventStatus(ctx context.Context, eventUUID string, status storage.EventStatus) error {
	s.logger.Info("Updating event status",
		logger.String("uuid", eventUUID),
		logger.String("status", string(status)),
	)
	return s.eventStorage.UpdateStatus(ctx, eventUUID, status)
}

// MarkEventProcessing 标记事件为处理中
func (s *EventStoreService) MarkEventProcessing(ctx context.Context, eventUUID string) error {
	return s.UpdateEventStatus(ctx, eventUUID, storage.EventStatusProcessing)
}

// MarkEventCompleted 标记事件为已完成
func (s *EventStoreService) MarkEventCompleted(ctx context.Context, eventUUID string) error {
	return s.eventStorage.UpdateProcessedAt(ctx, eventUUID, time.Now())
}

// MarkEventFailed 标记事件为失败
func (s *EventStoreService) MarkEventFailed(ctx context.Context, eventUUID string) error {
	return s.UpdateEventStatus(ctx, eventUUID, storage.EventStatusFailed)
}

// MarkEventCancelled 标记事件为已取消
func (s *EventStoreService) MarkEventCancelled(ctx context.Context, eventUUID string) error {
	return s.UpdateEventStatus(ctx, eventUUID, storage.EventStatusCancelled)
}

// CreateQualityCheck 创建质量检查
func (s *EventStoreService) CreateQualityCheck(ctx context.Context, check *storage.QualityCheck) error {
	// 设置默认状态
	if check.CheckStatus == "" {
		check.CheckStatus = string(storage.QualityCheckStatusPending)
	}

	s.logger.Info("Creating quality check",
		logger.String("event_uuid", check.EventUUID),
		logger.String("type", check.CheckType),
	)

	return s.qualityCheckStorage.Create(ctx, check)
}

// GetQualityChecks 获取事件的质量检查
func (s *EventStoreService) GetQualityChecks(ctx context.Context, eventUUID string) ([]*storage.QualityCheck, error) {
	return s.qualityCheckStorage.GetByEventUUID(ctx, eventUUID)
}

// UpdateQualityCheck 更新质量检查状态
func (s *EventStoreService) UpdateQualityCheck(ctx context.Context, id int64, status string, result string, score float64, details string) error {
	s.logger.Info("Updating quality check",
		logger.Int64("id", id),
		logger.String("status", status),
		logger.String("result", result),
		logger.Any("score", score),
	)
	return s.qualityCheckStorage.UpdateStatus(ctx, id, status, result, score, details)
}

// UpdateQualityCheckByType 根据类型更新质量检查
func (s *EventStoreService) UpdateQualityCheckByType(ctx context.Context, eventUUID string, checkType string, status string, result string, score float64, details string) error {
	s.logger.Info("Updating quality check by type",
		logger.String("event_uuid", eventUUID),
		logger.String("type", checkType),
		logger.String("status", status),
		logger.String("result", result),
		logger.Any("score", score),
	)
	return s.qualityCheckStorage.UpdateStatusByType(ctx, eventUUID, checkType, status, result, score, details)
}

// GetEventStatistics 获取事件统计
func (s *EventStoreService) GetEventStatistics(ctx context.Context) (*storage.EventStatistics, error) {
	return s.eventStorage.GetStatistics(ctx)
}

// GetQualityCheckStatistics 获取质量检查统计
func (s *EventStoreService) GetQualityCheckStatistics(ctx context.Context) (*storage.QualityCheckStatistics, error) {
	return s.qualityCheckStorage.GetStatistics(ctx)
}

// GetPendingEvents 获取待处理事件
func (s *EventStoreService) GetPendingEvents(ctx context.Context, limit int) ([]*storage.Event, error) {
	filter := &storage.EventFilter{
		Status: storage.EventStatusPending,
		Limit:  limit,
	}
	return s.eventStorage.List(ctx, filter)
}

// GetEventsByPR 获取 PR 相关事件
func (s *EventStoreService) GetEventsByPR(ctx context.Context, repoName string, prNumber int) ([]*storage.Event, error) {
	filter := &storage.EventFilter{
		RepoName: repoName,
		PRNumber: prNumber,
	}
	return s.eventStorage.List(ctx, filter)
}

// GetProcessingEvents 获取处理中的事件
func (s *EventStoreService) GetProcessingEvents(ctx context.Context) ([]*storage.Event, error) {
	filter := &storage.EventFilter{
		Status: storage.EventStatusProcessing,
	}
	return s.eventStorage.List(ctx, filter)
}
