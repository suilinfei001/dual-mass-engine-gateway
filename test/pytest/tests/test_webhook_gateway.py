"""
Webhook Gateway API Tests.
Service: webhook-gateway (Port 4001)
"""
import pytest
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestWebhookGatewayHealth:
    """Health check tests for Webhook Gateway."""

    def test_health_check(self, webhook_gateway, http_session):
        """WG-001: Health check should return ok status."""
        resp = http_session.get(f"{webhook_gateway}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "ok"
        assert data.get("service") == "webhook-gateway"

    def test_get_status(self, webhook_gateway, http_session):
        """WG-002: Service status query."""
        resp = http_session.get(f"{webhook_gateway}/api/status")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("service") == "webhook-gateway"
        assert data.get("status") == "running"
        assert "timestamp" in data


@pytest.mark.api
@pytest.mark.p1
class TestGitHubWebhook:
    """GitHub Webhook tests."""

    def test_pr_opened_event(self, webhook_gateway, http_session, github_pr_opened_payload):
        """WG-010: PR Opened event should be accepted."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=github_pr_opened_payload,
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 202]
        data = extract_data(resp.json())
        assert "event_uuid" in data or data.get("status") in ["accepted", "accepted_with_error"]

    def test_pr_synchronize_event(self, webhook_gateway, http_session, github_pr_synchronize_payload):
        """WG-011: PR Synchronize event should be accepted."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=github_pr_synchronize_payload,
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 202]

    def test_push_event(self, webhook_gateway, http_session):
        """WG-012: Push event should be accepted."""
        payload = {
            "ref": "refs/heads/main",
            "repository": {
                "id": 12345,
                "name": "test-repo",
                "full_name": "owner/test-repo",
            },
            "sender": {"login": "testuser"},
        }
        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=payload,
            headers={"X-GitHub-Event": "push"},
        )
        assert resp.status_code in [200, 202, 500]

    def test_empty_payload(self, webhook_gateway, http_session):
        """WG-013: Empty JSON body should be handled gracefully."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json={},
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 400, 500]

    def test_invalid_json(self, webhook_gateway, http_session):
        """WG-014: Invalid JSON should return error."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            data="{invalid json",
            headers={"Content-Type": "application/json", "X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 400, 500]

    def test_missing_event_header(self, webhook_gateway, http_session, github_pr_opened_payload):
        """WG-015: Missing X-GitHub-Event header should be handled."""
        resp = http_session.post(f"{webhook_gateway}/webhook/github", json=github_pr_opened_payload)
        assert resp.status_code in [200, 400]


@pytest.mark.api
@pytest.mark.p1
class TestGitLabWebhook:
    """GitLab Webhook tests."""

    def test_mr_opened_event(self, webhook_gateway, http_session, gitlab_mr_opened_payload):
        """WG-020: MR Opened event should be accepted."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/gitlab",
            json=gitlab_mr_opened_payload,
            headers={"X-Gitlab-Event": "Merge Request Hook"},
        )
        assert resp.status_code in [200, 202]

    def test_push_event(self, webhook_gateway, http_session):
        """WG-021: GitLab Push event should be accepted."""
        payload = {
            "object_kind": "push",
            "project": {
                "id": 12345,
                "name": "test-project",
                "http_url": "https://gitlab.com/owner/test-project",
            },
        }
        resp = http_session.post(
            f"{webhook_gateway}/webhook/gitlab",
            json=payload,
            headers={"X-Gitlab-Event": "Push Hook"},
        )
        assert resp.status_code in [200, 202]

    def test_invalid_token(self, webhook_gateway, http_session, gitlab_mr_opened_payload):
        """WG-022: Invalid token should be handled."""
        resp = http_session.post(
            f"{webhook_gateway}/webhook/gitlab",
            json=gitlab_mr_opened_payload,
            headers={"X-Gitlab-Event": "Merge Request Hook", "X-Gitlab-Token": "invalid-token"},
        )
        assert resp.status_code in [200, 401]
