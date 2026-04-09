package service

import (
	"testing"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/webhook-gateway/internal/client"
)

func TestWebhookGatewayService_NewService(t *testing.T) {
	log := logger.New(logger.Config{Service: "test"})
	service := NewWebhookGatewayService(nil, nil, log)
	if service == nil {
		t.Error("Expected service to be created")
	}
}

func TestWebhookSource_Constants(t *testing.T) {
	tests := []struct {
		source   client.WebhookSource
		expected string
	}{
		{client.WebhookSourceGitHub, "github"},
		{client.WebhookSourceGitLab, "gitlab"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.source))
			}
		})
	}
}

func TestWebhookEventType_Constants(t *testing.T) {
	tests := []struct {
		eventType client.WebhookEventType
		expected  string
	}{
		{client.GitHubPullRequestOpened, "pull_request.opened"},
		{client.GitHubPullRequestSynchronized, "pull_request.synchronize"},
		{client.GitHubPullRequestClosed, "pull_request.closed"},
		{client.GitHubPush, "push"},
		{client.GitHubRelease, "release"},
		{client.GitLabMergeRequestOpened, "merge_request.opened"},
		{client.GitLabMergeRequestUpdated, "merge_request.updated"},
		{client.GitLabMergeRequestClosed, "merge_request.closed"},
		{client.GitLabPush, "push"},
		{client.GitLabRelease, "release"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.eventType))
			}
		})
	}
}

func TestForwardedEvent_Fields(t *testing.T) {
	event := &client.ForwardedEvent{
		UUID:      "test-uuid",
		EventType: client.GitHubPullRequestOpened,
		Status:    "pending",
		Source:    client.WebhookSourceGitHub,
		RepoID:    12345,
		RepoName:  "test/repo",
		RepoOwner: "test",
		PRNumber:  42,
		Author:    "testuser",
	}

	if event.UUID != "test-uuid" {
		t.Errorf("Expected UUID=test-uuid, got %s", event.UUID)
	}
	if event.EventType != client.GitHubPullRequestOpened {
		t.Errorf("Expected EventType=pull_request.opened, got %s", event.EventType)
	}
	if event.Source != client.WebhookSourceGitHub {
		t.Errorf("Expected Source=github, got %s", event.Source)
	}
	if event.RepoName != "test/repo" {
		t.Errorf("Expected RepoName=test/repo, got %s", event.RepoName)
	}
	if event.PRNumber != 42 {
		t.Errorf("Expected PRNumber=42, got %d", event.PRNumber)
	}
}

func TestWebhookDeliveryResult_Fields(t *testing.T) {
	result := &client.WebhookDeliveryResult{
		Success:      true,
		EventUUID:    "event-uuid",
		StatusCode:   200,
		ErrorMessage: "",
	}

	if !result.Success {
		t.Error("Expected Success=true")
	}
	if result.EventUUID != "event-uuid" {
		t.Errorf("Expected EventUUID=event-uuid, got %s", result.EventUUID)
	}
	if result.StatusCode != 200 {
		t.Errorf("Expected StatusCode=200, got %d", result.StatusCode)
	}
}

func TestWebhookConfig_Default(t *testing.T) {
	config := client.DefaultWebhookConfig()

	if config.Port != 4001 {
		t.Errorf("Expected Port=4001, got %d", config.Port)
	}
	if config.EventStoreURL != "http://localhost:4002" {
		t.Errorf("Expected EventStoreURL=http://localhost:4002, got %s", config.EventStoreURL)
	}
	if config.Verification == nil {
		t.Error("Expected Verification to be set")
	}
}

func TestGitHubWebhookPayload_Fields(t *testing.T) {
	payload := &client.GitHubWebhookPayload{
		Action: "opened",
		Sender: &client.GitActor{
			ID:    123,
			Login: "testuser",
		},
		Repository: &client.GitRepository{
			ID:       456,
			Name:     "test-repo",
			FullName: "owner/test-repo",
		},
		PullRequest: &client.GitPullRequest{
			Number: 42,
			Title:  "Test PR",
			State:  "open",
		},
	}

	if payload.Action != "opened" {
		t.Errorf("Expected Action=opened, got %s", payload.Action)
	}
	if payload.Sender.Login != "testuser" {
		t.Errorf("Expected Sender.Login=testuser, got %s", payload.Sender.Login)
	}
	if payload.Repository.FullName != "owner/test-repo" {
		t.Errorf("Expected Repository.FullName=owner/test-repo, got %s", payload.Repository.FullName)
	}
	if payload.PullRequest.Number != 42 {
		t.Errorf("Expected PullRequest.Number=42, got %d", payload.PullRequest.Number)
	}
}

func TestGitLabWebhookPayload_Fields(t *testing.T) {
	payload := &client.GitLabWebhookPayload{
		ObjectKind: "merge_request",
		EventType:  "merge_request",
		User: &client.GitActor{
			ID:    123,
			Login: "testuser",
		},
		Project: &client.GitLabProject{
			ID:      456,
			Name:    "test-project",
			HTTPURL: "https://gitlab.com/owner/test-project",
		},
		ObjectAttributes: &client.GitLabMergeRequestAttrs{
			IID:    42,
			Title:  "Test MR",
			State:  "opened",
			Action: "open",
		},
	}

	if payload.ObjectKind != "merge_request" {
		t.Errorf("Expected ObjectKind=merge_request, got %s", payload.ObjectKind)
	}
	if payload.User.Login != "testuser" {
		t.Errorf("Expected User.Login=testuser, got %s", payload.User.Login)
	}
	if payload.Project.Name != "test-project" {
		t.Errorf("Expected Project.Name=test-project, got %s", payload.Project.Name)
	}
	if payload.ObjectAttributes.IID != 42 {
		t.Errorf("Expected ObjectAttributes.IID=42, got %d", payload.ObjectAttributes.IID)
	}
}
