// Package client provides data models for webhook gateway service.
package client

import "time"

// WebhookSource Webhook 来源
type WebhookSource string

const (
	WebhookSourceGitHub WebhookSource = "github"
	WebhookSourceGitLab WebhookSource = "gitlab"
)

// WebhookEventType Webhook 事件类型
type WebhookEventType string

const (
	// GitHub Events
	GitHubPullRequestOpened     WebhookEventType = "pull_request.opened"
	GitHubPullRequestSynchronized WebhookEventType = "pull_request.synchronize"
	GitHubPullRequestClosed     WebhookEventType = "pull_request.closed"
	GitHubPush                  WebhookEventType = "push"
	GitHubRelease               WebhookEventType = "release"

	// GitLab Events
	GitLabMergeRequestOpened WebhookEventType = "merge_request.opened"
	GitLabMergeRequestUpdated WebhookEventType = "merge_request.updated"
	GitLabMergeRequestClosed WebhookEventType = "merge_request.closed"
	GitLabPush               WebhookEventType = "push"
	GitLabRelease            WebhookEventType = "release"
)

// GitHubWebhookPayload GitHub Webhook Payload 结构
type GitHubWebhookPayload struct {
	// 事件头信息
	Headers map[string]string `json:"-"`

	// 基础信息
	Action      string `json:"action,omitempty"`
	Sender      *GitActor `json:"sender,omitempty"`
	Repository  *GitRepository `json:"repository,omitempty"`
	Organization *GitOrganization `json:"organization,omitempty"`

	// Pull Request 相关
	PullRequest *GitPullRequest `json:"pull_request,omitempty"`

	// Push 相关
	Ref        string `json:"ref,omitempty"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Commits    []GitCommit `json:"commits,omitempty"`

	// Release 相关
	Release *GitRelease `json:"release,omitempty"`
}

// GitActor GitHub/GitLab Actor
type GitActor struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// GitRepository 仓库信息
type GitRepository struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Owner       *GitActor `json:"owner,omitempty"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// GitOrganization 组织信息
type GitOrganization struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

// GitPullRequest Pull Request 信息
type GitPullRequest struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	State     string `json:"state"`
	Title     string `json:"title"`
	Body      string `json:"body,omitempty"`
	User      *GitActor `json:"user,omitempty"`
	Head      *GitBranchRef `json:"head,omitempty"`
	Base      *GitBranchRef `json:"base,omitempty"`
	Mergeable *bool  `json:"mergeable,omitempty"`
	Merged    bool   `json:"merged"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
}

// GitBranchRef 分支引用
type GitBranchRef struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	SHA   string `json:"sha"`
	Repo  *GitRepository `json:"repo,omitempty"`
}

// GitCommit 提交信息
type GitCommit struct {
	ID       string `json:"id"`
	Message  string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Author   *GitActor `json:"author,omitempty"`
}

// GitRelease 发布信息
type GitRelease struct {
	ID          int64  `json:"id"`
	TagName     string `json:"tag_name"`
	Name        string `json:"name,omitempty"`
	Body        string `json:"body,omitempty"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Author      *GitActor `json:"author,omitempty"`
}

// GitLabWebhookPayload GitLab Webhook Payload 结构
type GitLabWebhookPayload struct {
	// 事件头信息
	Headers map[string]string `json:"-"`

	// 基础信息
	ObjectKind string `json:"object_kind,omitempty"`
	EventType  string `json:"event_type,omitempty"`

	// 用户信息
	User *GitActor `json:"user,omitempty"`

	// 项目信息
	Project *GitLabProject `json:"project,omitempty"`

	// Merge Request 相关
	ObjectAttributes *GitLabMergeRequestAttrs `json:"object_attributes,omitempty"`

	// Push 相关
	Ref        string `json:"ref,omitempty"`
	Before     string `json:"before,omitempty"`
	After      string `json:"after,omitempty"`
	Commits    []GitLabCommit `json:"commits,omitempty"`
	TotalCommits int `json:"total_commits_count,omitempty"`
}

// GitLabProject GitLab 项目信息
type GitLabProject struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"name_with_namespace"`
	HTTPURL           string `json:"http_url"`
	SSHURL            string `json:"ssh_url"`
	DefaultBranch     string `json:"default_branch"`
}

// GitLabMergeRequestAttrs Merge Request 属性
type GitLabMergeRequestAttrs struct {
	ID          int64     `json:"id"`
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	State       string    `json:"state"`
	Action      string    `json:"action"`
	SourceBranch string   `json:"source_branch"`
	TargetBranch string   `json:"target_branch"`
	Source      *GitLabProjectRef `json:"source,omitempty"`
	Target      *GitLabProjectRef `json:"target,omitempty"`
	Author      *GitActor `json:"author,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MergedAt    *time.Time `json:"merged_at,omitempty"`
}

// GitLabProjectRef GitLab 项目引用
type GitLabProjectRef struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	HTTPURL   string `json:"http_url"`
}

// GitLabCommit GitLab 提交信息
type GitLabCommit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Author    *GitActor `json:"author,omitempty"`
}

// ForwardedEvent 转发到 Event Store 的事件
type ForwardedEvent struct {
	UUID        string                 `json:"uuid"`
	EventType   WebhookEventType       `json:"event_type"`
	Status      string                 `json:"status"`
	Source      WebhookSource          `json:"source"`
	RepoID      int64                  `json:"repo_id"`
	RepoName    string                 `json:"repo_name"`
	RepoOwner   string                 `json:"repo_owner"`
	RepoURL     string                 `json:"repo_url"`
	PRNumber    int                    `json:"pr_number,omitempty"`
	PRTitle     string                 `json:"pr_title,omitempty"`
	CommitSHA   string                 `json:"commit_sha,omitempty"`
	Ref         string                 `json:"ref,omitempty"`
	Author      string                 `json:"author"`
	AuthorEmail string                 `json:"author_email,omitempty"`
	Payload     string                 `json:"payload"`
	ReceivedAt  time.Time              `json:"received_at"`
}

// WebhookDeliveryResult Webhook 投递结果
type WebhookDeliveryResult struct {
	Success      bool      `json:"success"`
	EventUUID    string    `json:"event_uuid"`
	StatusCode   int       `json:"status_code"`
	ErrorMessage string    `json:"error_message,omitempty"`
	DeliveredAt  time.Time `json:"delivered_at"`
}

// WebhookVerificationConfig Webhook 验证配置
type WebhookVerificationConfig struct {
	GitHubSecret string `json:"github_secret"`
	GitLabSecret string `json:"gitlab_secret"`
}

// WebhookConfig Webhook Gateway 配置
type WebhookConfig struct {
	Port            int                       `json:"port"`
	EventStoreURL   string                    `json:"event_store_url"`
	APIToken        string                    `json:"api_token"`
	Verification    *WebhookVerificationConfig `json:"verification"`
}

// DefaultWebhookConfig 默认配置
func DefaultWebhookConfig() *WebhookConfig {
	return &WebhookConfig{
		Port:          4001,
		EventStoreURL: "http://localhost:4002",
		APIToken:      "",
		Verification: &WebhookVerificationConfig{
			GitHubSecret: "",
			GitLabSecret: "",
		},
	}
}
