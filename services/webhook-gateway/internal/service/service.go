// Package service provides business logic for webhook gateway service.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/webhook-gateway/internal/client"
)

// WebhookGatewayService Webhook 网关服务
type WebhookGatewayService struct {
	eventStoreClient *client.EventStoreClient
	config           *client.WebhookConfig
	logger           *logger.Logger
}

// NewWebhookGatewayService 创建 Webhook 网关服务
func NewWebhookGatewayService(
	ec *client.EventStoreClient,
	cfg *client.WebhookConfig,
	log *logger.Logger,
) *WebhookGatewayService {
	return &WebhookGatewayService{
		eventStoreClient: ec,
		config:          cfg,
		logger:          log,
	}
}

// HandleGitHubWebhook 处理 GitHub Webhook
func (s *WebhookGatewayService) HandleGitHubWebhook(ctx context.Context, headers map[string]string, payload []byte) (*client.WebhookDeliveryResult, error) {
	// 验证签名
	if s.config.Verification != nil && s.config.Verification.GitHubSecret != "" {
		if err := s.verifyGitHubSignature(headers, payload, s.config.Verification.GitHubSecret); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
	}

	// 解析 payload
	var ghPayload client.GitHubWebhookPayload
	if err := json.Unmarshal(payload, &ghPayload); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub payload: %w", err)
	}
	ghPayload.Headers = headers

	// 提取事件类型
	eventType := s.extractGitHubEventType(headers, &ghPayload)

	// 转换为通用事件
	forwardedEvent, err := s.convertGitHubEvent(eventType, &ghPayload, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event: %w", err)
	}

	s.logger.Info("Processing GitHub webhook",
		logger.String("event_uuid", forwardedEvent.UUID),
		logger.String("event_type", string(forwardedEvent.EventType)),
		logger.String("repo", forwardedEvent.RepoName),
		logger.Int("pr_number", forwardedEvent.PRNumber),
	)

	// 转发到 Event Store
	return s.eventStoreClient.CreateEvent(ctx, forwardedEvent)
}

// HandleGitLabWebhook 处理 GitLab Webhook
func (s *WebhookGatewayService) HandleGitLabWebhook(ctx context.Context, headers map[string]string, payload []byte) (*client.WebhookDeliveryResult, error) {
	// 验证签名
	if s.config.Verification != nil && s.config.Verification.GitLabSecret != "" {
		if err := s.verifyGitLabSignature(headers, payload, s.config.Verification.GitLabSecret); err != nil {
			return nil, fmt.Errorf("signature verification failed: %w", err)
		}
	}

	// 解析 payload
	var glPayload client.GitLabWebhookPayload
	if err := json.Unmarshal(payload, &glPayload); err != nil {
		return nil, fmt.Errorf("failed to parse GitLab payload: %w", err)
	}
	glPayload.Headers = headers

	// 提取事件类型
	eventType := s.extractGitLabEventType(headers, &glPayload)

	// 转换为通用事件
	forwardedEvent, err := s.convertGitLabEvent(eventType, &glPayload, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert event: %w", err)
	}

	s.logger.Info("Processing GitLab webhook",
		logger.String("event_uuid", forwardedEvent.UUID),
		logger.String("event_type", string(forwardedEvent.EventType)),
		logger.String("repo", forwardedEvent.RepoName),
		logger.Int("pr_number", forwardedEvent.PRNumber),
	)

	// 转发到 Event Store
	return s.eventStoreClient.CreateEvent(ctx, forwardedEvent)
}

// verifyGitHubSignature 验证 GitHub 签名
func (s *WebhookGatewayService) verifyGitHubSignature(headers map[string]string, payload []byte, secret string) error {
	signature := headers["X-Hub-Signature-256"]
	if signature == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}

	// 签名格式: sha256=<hex>
	if len(signature) < 8 || signature[:7] != "sha256=" {
		return fmt.Errorf("invalid signature format")
	}

	signatureBytes, err := hex.DecodeString(signature[7:])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// 计算 HMAC
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := h.Sum(nil)

	// 比较签名
	if !hmac.Equal(signatureBytes, expectedSignature) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

// verifyGitLabSignature 验证 GitLab 签名
func (s *WebhookGatewayService) verifyGitLabSignature(headers map[string]string, payload []byte, secret string) error {
	token := headers["X-Gitlab-Token"]
	if token == "" {
		return fmt.Errorf("missing X-Gitlab-Token header")
	}

	// GitLab 使用简单的 token 比较
	if token != secret {
		return fmt.Errorf("token mismatch")
	}

	return nil
}

// extractGitHubEventType 提取 GitHub 事件类型
func (s *WebhookGatewayService) extractGitHubEventType(headers map[string]string, payload *client.GitHubWebhookPayload) client.WebhookEventType {
	eventHeader := headers["X-GitHub-Event"]

	switch eventHeader {
	case "pull_request":
		switch payload.Action {
		case "opened", "reopened":
			return client.GitHubPullRequestOpened
		case "synchronize":
			return client.GitHubPullRequestSynchronized
		case "closed":
			return client.GitHubPullRequestClosed
		}
	case "push":
		return client.GitHubPush
	case "release":
		return client.GitHubRelease
	}

	return client.WebhookEventType(eventHeader)
}

// extractGitLabEventType 提取 GitLab 事件类型
func (s *WebhookGatewayService) extractGitLabEventType(headers map[string]string, payload *client.GitLabWebhookPayload) client.WebhookEventType {
	eventHeader := headers["X-Gitlab-Event"]

	switch eventHeader {
	case "Merge Request Hook":
		switch payload.ObjectAttributes.Action {
		case "open", "reopen":
			return client.GitLabMergeRequestOpened
		case "update":
			return client.GitLabMergeRequestUpdated
		case "close", "merge":
			return client.GitLabMergeRequestClosed
		}
	case "Push Hook":
		return client.GitLabPush
	case "Release Hook":
		return client.GitLabRelease
	}

	return client.WebhookEventType(eventHeader)
}

// convertGitHubEvent 将 GitHub 事件转换为通用事件
func (s *WebhookGatewayService) convertGitHubEvent(eventType client.WebhookEventType, payload *client.GitHubWebhookPayload, rawPayload []byte) (*client.ForwardedEvent, error) {
	event := &client.ForwardedEvent{
		UUID:       uuid.New().String(),
		EventType:  eventType,
		Status:     "pending",
		Source:     client.WebhookSourceGitHub,
		Payload:    string(rawPayload),
		ReceivedAt: time.Now(), // 使用当前时间
	}

	// 提取仓库信息
	if payload.Repository != nil {
		event.RepoID = payload.Repository.ID
		event.RepoName = payload.Repository.Name
		event.RepoOwner = payload.Repository.Owner.Login
		event.RepoURL = payload.Repository.HTMLURL
	}

	// 提取 PR 信息
	if payload.PullRequest != nil {
		event.PRNumber = payload.PullRequest.Number
		event.PRTitle = payload.PullRequest.Title
		event.Ref = payload.PullRequest.Head.Ref
		event.CommitSHA = payload.PullRequest.Head.SHA
	}

	// 提取作者信息
	if payload.Sender != nil {
		event.Author = payload.Sender.Login
	}

	// 提取 Push 信息
	if payload.Ref != "" {
		event.Ref = payload.Ref
		if len(payload.Commits) > 0 {
			event.CommitSHA = payload.Commits[0].ID
		}
	}

	return event, nil
}

// convertGitLabEvent 将 GitLab 事件转换为通用事件
func (s *WebhookGatewayService) convertGitLabEvent(eventType client.WebhookEventType, payload *client.GitLabWebhookPayload, rawPayload []byte) (*client.ForwardedEvent, error) {
	event := &client.ForwardedEvent{
		UUID:       uuid.New().String(),
		EventType:  eventType,
		Status:     "pending",
		Source:     client.WebhookSourceGitLab,
		Payload:    string(rawPayload),
		ReceivedAt: time.Now(), // 使用当前时间
	}

	// 提取项目信息
	if payload.Project != nil {
		event.RepoID = payload.Project.ID
		event.RepoName = payload.Project.Name
		event.RepoURL = payload.Project.HTTPURL
	}

	// 提取 MR 信息
	if payload.ObjectAttributes != nil {
		event.PRNumber = payload.ObjectAttributes.IID
		event.PRTitle = payload.ObjectAttributes.Title
		event.Ref = payload.ObjectAttributes.SourceBranch
	}

	// 提取作者信息
	if payload.User != nil {
		event.Author = payload.User.Login
	}

	// 提取 Push 信息
	if payload.Ref != "" {
		event.Ref = payload.Ref
		if len(payload.Commits) > 0 {
			event.CommitSHA = payload.Commits[0].ID
		}
	}

	return event, nil
}
