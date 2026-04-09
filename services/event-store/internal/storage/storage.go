// Package storage provides database access for event store service.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/shared/pkg/storage"
)

// Config extends shared config with additional fields
type Config struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// DSN generates data source name
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Open creates a new database connection
func Open(cfg Config, log *logger.Logger) (*storage.DB, error) {
	sharedCfg := storage.Config{
		Driver:          cfg.Driver,
		DSN:             cfg.DSN(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}
	return storage.Open(sharedCfg, log)
}

// EventStorage 事件存储
type EventStorage struct {
	db     *storage.DB
	logger *logger.Logger
}

// NewEventStorage 创建事件存储
func NewEventStorage(db *storage.DB) *EventStorage {
	return &EventStorage{db: db}
}

// Create 创建事件
func (s *EventStorage) Create(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (uuid, event_type, status, source, repo_id, repo_name, repo_owner,
		                    pr_number, commit_sha, author, payload, received_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(ctx, query,
		event.UUID, event.EventType, event.Status, event.Source,
		event.RepoID, event.RepoName, event.RepoOwner, event.PRNumber,
		event.CommitSHA, event.Author, event.Payload, event.ReceivedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

// GetByUUID 根据 UUID 获取事件
func (s *EventStorage) GetByUUID(ctx context.Context, uuid string) (*Event, error) {
	query := `
		SELECT id, uuid, event_type, status, source, repo_id, repo_name, repo_owner,
		       pr_number, commit_sha, author, payload, received_at, processed_at, created_at, updated_at
		FROM events
		WHERE uuid = ?
	`
	var e Event
	err := s.db.QueryRow(ctx, query, uuid).Scan(
		&e.ID, &e.UUID, &e.EventType, &e.Status, &e.Source, &e.RepoID,
		&e.RepoName, &e.RepoOwner, &e.PRNumber, &e.CommitSHA, &e.Author,
		&e.Payload, &e.ReceivedAt, &e.ProcessedAt, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &e, err
}

// List 根据条件列出事件
func (s *EventStorage) List(ctx context.Context, filter *EventFilter) ([]*Event, error) {
	var whereClauses []string
	var args []interface{}

	if filter.EventType != "" {
		whereClauses = append(whereClauses, "event_type = ?")
		args = append(args, filter.EventType)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Source != "" {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.RepoName != "" {
		whereClauses = append(whereClauses, "repo_name = ?")
		args = append(args, filter.RepoName)
	}
	if filter.PRNumber > 0 {
		whereClauses = append(whereClauses, "pr_number = ?")
		args = append(args, filter.PRNumber)
	}
	if filter.Author != "" {
		whereClauses = append(whereClauses, "author = ?")
		args = append(args, filter.Author)
	}
	if filter.StartDate != nil {
		whereClauses = append(whereClauses, "received_at >= ?")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != nil {
		whereClauses = append(whereClauses, "received_at <= ?")
		args = append(args, filter.EndDate)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	limitClause := ""
	if filter.Limit > 0 {
		limitClause = " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		limitClause += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	query := `
		SELECT id, uuid, event_type, status, source, repo_id, repo_name, repo_owner,
		       pr_number, commit_sha, author, payload, received_at, processed_at, created_at, updated_at
		FROM events
	` + whereClause + `
		ORDER BY received_at DESC
	` + limitClause

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID, &e.UUID, &e.EventType, &e.Status, &e.Source, &e.RepoID,
			&e.RepoName, &e.RepoOwner, &e.PRNumber, &e.CommitSHA, &e.Author,
			&e.Payload, &e.ReceivedAt, &e.ProcessedAt, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, nil
}

// UpdateStatus 更新事件状态
func (s *EventStorage) UpdateStatus(ctx context.Context, uuid string, status EventStatus) error {
	query := `UPDATE events SET status = ?, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(ctx, query, status, uuid)
	return err
}

// UpdateProcessedAt 更新处理完成时间
func (s *EventStorage) UpdateProcessedAt(ctx context.Context, uuid string, processedAt time.Time) error {
	query := `UPDATE events SET processed_at = ?, status = ?, updated_at = NOW() WHERE uuid = ?`
	_, err := s.db.Exec(ctx, query, processedAt, EventStatusCompleted, uuid)
	return err
}

// GetStatistics 获取事件统计
func (s *EventStorage) GetStatistics(ctx context.Context) (*EventStatistics, error) {
	query := `
		SELECT
			COUNT(*) as total_events,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as pending_events,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as processing_events,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed_events,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as failed_events,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as cancelled_events
		FROM events
	`
	var stats EventStatistics
	err := s.db.QueryRow(ctx, query,
		EventStatusPending, EventStatusProcessing, EventStatusCompleted,
		EventStatusFailed, EventStatusCancelled,
	).Scan(
		&stats.TotalEvents, &stats.PendingEvents, &stats.ProcessingEvents,
		&stats.CompletedEvents, &stats.FailedEvents, &stats.CancelledEvents,
	)
	return &stats, err
}

// QualityCheckStorage 质量检查存储
type QualityCheckStorage struct {
	db *storage.DB
}

// NewQualityCheckStorage 创建质量检查存储
func NewQualityCheckStorage(db *storage.DB) *QualityCheckStorage {
	return &QualityCheckStorage{db: db}
}

// Create 创建质量检查
func (s *QualityCheckStorage) Create(ctx context.Context, check *QualityCheck) error {
	query := `
		INSERT INTO quality_checks (event_uuid, check_type, check_status, result, score, details)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.Exec(ctx, query,
		check.EventUUID, check.CheckType, check.CheckStatus,
		check.Result, check.Score, check.Details,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	check.ID = id
	return nil
}

// GetByEventUUID 根据事件 UUID 获取质量检查
func (s *QualityCheckStorage) GetByEventUUID(ctx context.Context, eventUUID string) ([]*QualityCheck, error) {
	query := `
		SELECT id, event_uuid, check_type, check_status, result, score, details,
		       started_at, completed_at, created_at, updated_at
		FROM quality_checks
		WHERE event_uuid = ?
		ORDER BY created_at
	`
	rows, err := s.db.Query(ctx, query, eventUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*QualityCheck
	for rows.Next() {
		var c QualityCheck
		err := rows.Scan(
			&c.ID, &c.EventUUID, &c.CheckType, &c.CheckStatus,
			&c.Result, &c.Score, &c.Details, &c.StartedAt,
			&c.CompletedAt, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		checks = append(checks, &c)
	}
	return checks, nil
}

// UpdateStatus 更新质量检查状态
func (s *QualityCheckStorage) UpdateStatus(ctx context.Context, id int64, status string, result string, score float64, details string) error {
	query := `
		UPDATE quality_checks
		SET check_status = ?, result = ?, score = COALESCE(?, score), details = COALESCE(?, details),
		    completed_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`
	_, err := s.db.Exec(ctx, query, status, result, score, details, id)
	return err
}

// UpdateStatusByType 根据检查类型更新状态
func (s *QualityCheckStorage) UpdateStatusByType(ctx context.Context, eventUUID string, checkType string, status string, result string, score float64, details string) error {
	query := `
		UPDATE quality_checks
		SET check_status = ?, result = ?, score = COALESCE(?, score), details = COALESCE(?, details),
		    completed_at = NOW(), updated_at = NOW()
		WHERE event_uuid = ? AND check_type = ?
	`
	execResult, err := s.db.Exec(ctx, query, status, result, score, details, eventUUID, checkType)
	if err != nil {
		return err
	}
	rows, _ := execResult.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetStatistics 获取质量检查统计
func (s *QualityCheckStorage) GetStatistics(ctx context.Context) (*QualityCheckStatistics, error) {
	query := `
		SELECT
			COUNT(*) as total_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as pending_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as running_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as passed_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as failed_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as skipped_checks,
			SUM(CASE WHEN check_status = ? THEN 1 ELSE 0 END) as cancelled_checks
		FROM quality_checks
	`
	var stats QualityCheckStatistics
	err := s.db.QueryRow(ctx, query,
		QualityCheckStatusPending, QualityCheckStatusRunning, QualityCheckStatusPassed,
		QualityCheckStatusFailed, QualityCheckStatusSkipped, QualityCheckStatusCancelled,
	).Scan(
		&stats.TotalChecks, &stats.PendingChecks, &stats.RunningChecks,
		&stats.PassedChecks, &stats.FailedChecks, &stats.SkippedChecks, &stats.CancelledChecks,
	)
	return &stats, err
}
