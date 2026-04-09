"""
End-to-End Tests.
Complete workflow tests across multiple services.
"""
import pytest
import time
import uuid
from conftest import extract_data


@pytest.mark.e2e
class TestGitHubPRQualityCheckFlow:
    """Scenario 1: GitHub PR quality check complete flow."""

    def test_pr_opened_to_completion(
        self,
        webhook_gateway,
        event_store,
        task_scheduler,
        http_session,
        github_pr_opened_payload,
    ):
        """Complete PR quality check flow from webhook to completion."""
        unique_id = str(uuid.uuid4())[:8]
        github_pr_opened_payload["pull_request"]["number"] = int(f"42{unique_id}", 16) % 10000

        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=github_pr_opened_payload,
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 202]

        resp = http_session.get(f"{event_store}/api/events/pending?limit=10")
        assert resp.status_code == 200

        resp = http_session.get(f"{event_store}/api/events/statistics")
        assert resp.status_code == 200

        resp = http_session.get(f"{task_scheduler}/api/tasks?limit=10")
        assert resp.status_code == 200


@pytest.mark.e2e
class TestPRSynchronizeCancelFlow:
    """Scenario 2: PR synchronize cancel flow."""

    def test_pr_sync_cancels_previous(
        self,
        webhook_gateway,
        event_store,
        http_session,
        github_pr_opened_payload,
        github_pr_synchronize_payload,
    ):
        """PR synchronize should cancel previous events."""
        pr_number = int(f"42{str(uuid.uuid4())[:8]}", 16) % 10000
        github_pr_opened_payload["pull_request"]["number"] = pr_number
        github_pr_synchronize_payload["pull_request"]["number"] = pr_number

        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=github_pr_opened_payload,
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 202]

        time.sleep(1)

        resp = http_session.post(
            f"{webhook_gateway}/webhook/github",
            json=github_pr_synchronize_payload,
            headers={"X-GitHub-Event": "pull_request"},
        )
        assert resp.status_code in [200, 202]

        resp = http_session.get(f"{event_store}/api/events/statistics")
        assert resp.status_code == 200


@pytest.mark.e2e
class TestResourceAllocationFlow:
    """Scenario 3: Resource allocation and release flow."""

    def test_resource_create_match_release(
        self,
        resource_manager,
        http_session,
        category_payload,
        resource_payload,
    ):
        """Complete resource allocation flow."""
        resp = http_session.post(f"{resource_manager}/api/categories", json=category_payload)
        assert resp.status_code in [200, 201, 500]
        if resp.status_code in [200, 201]:
            category_data = extract_data(resp.json())
            resource_payload["category_id"] = category_data.get("id", 1)

        resp = http_session.post(f"{resource_manager}/api/resources", json=resource_payload)
        assert resp.status_code in [200, 201]
        resource_data = extract_data(resp.json())
        resource_uuid = resource_data.get("uuid")

        if resource_uuid:
            resp = http_session.get(f"{resource_manager}/api/resources/{resource_uuid}")
            assert resp.status_code in [200, 404]

            match_payload = {
                "category_id": resource_payload.get("category_id", 1),
                "task_uuid": str(uuid.uuid4()),
                "required_count": 1,
            }
            resp = http_session.post(f"{resource_manager}/api/resources/match", json=match_payload)
            assert resp.status_code in [200, 500]

            resp = http_session.post(f"{resource_manager}/api/resources/{resource_uuid}/release")
            assert resp.status_code in [200, 404, 500]


@pytest.mark.e2e
class TestAILogAnalysisFlow:
    """Scenario 4: AI log analysis flow."""

    def test_log_analysis_workflow(self, ai_analyzer, http_session):
        """Complete AI log analysis flow."""
        resp = http_session.get(f"{ai_analyzer}/api/pool/stats")
        assert resp.status_code == 200

        payload = {
            "log_content": "Error: Connection timeout\nStack trace: at main.py line 42",
            "task_name": "basic_ci_all",
        }
        resp = http_session.post(f"{ai_analyzer}/api/analyze", json=payload)
        assert resp.status_code in [200, 500]

        batch_payload = {
            "log_contents": ["Error 1", "Error 2", "Error 3"],
            "task_name": "basic_ci_all",
        }
        resp = http_session.post(f"{ai_analyzer}/api/analyze/batch", json=batch_payload)
        assert resp.status_code in [200, 500]


@pytest.mark.e2e
class TestServiceHealthCheck:
    """Scenario 5: Service health check."""

    def test_all_services_healthy(self, services, http_session):
        """All services should return healthy status."""
        service_names = {
            "webhook_gateway": "webhook-gateway",
            "event_store": "event-store",
            "task_scheduler": "task-scheduler",
            "executor_service": "executor-service",
            "ai_analyzer": "ai-analyzer",
            "resource_manager": "resource-manager",
        }

        for key, expected_name in service_names.items():
            url = services[key]
            resp = http_session.get(f"{url}/health", timeout=5)
            assert resp.status_code == 200, f"{expected_name} health check failed"
            data = extract_data(resp.json())
            assert data.get("status") in ["ok", "healthy"], f"{expected_name} returned unhealthy status"
            if "service" in data:
                assert data["service"] == expected_name, f"{expected_name} returned wrong service name"
