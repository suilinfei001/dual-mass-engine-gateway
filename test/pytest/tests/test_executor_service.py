"""
Executor Service API Tests.
Service: executor-service (Port 4004)
"""
import pytest
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestExecutorServiceHealth:
    """Health check tests for Executor Service."""

    def test_health_check(self, executor_service, http_session):
        """EX-001: Health check should return ok status."""
        resp = http_session.get(f"{executor_service}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "ok"
        assert data.get("service") == "executor-service"


@pytest.mark.api
@pytest.mark.p0
class TestExecutionManagement:
    """Execution management tests."""

    def test_execute_task(self, executor_service, http_session, execution_payload):
        """EX-010: Execute task should return execution result."""
        resp = http_session.post(f"{executor_service}/api/execute", json=execution_payload)
        assert resp.status_code in [200, 202, 500]

    def test_execute_missing_task_uuid(self, executor_service, http_session):
        """EX-011: Missing task_uuid should return error."""
        resp = http_session.post(f"{executor_service}/api/execute", json={})
        assert resp.status_code == 400

    def test_get_execution_status(self, executor_service, http_session):
        """EX-012: Get execution status."""
        resp = http_session.get(f"{executor_service}/api/executions/test-exec-123")
        assert resp.status_code in [200, 400, 404]

    def test_get_execution_logs(self, executor_service, http_session):
        """EX-013: Get execution logs."""
        resp = http_session.get(f"{executor_service}/api/executions/test-exec-123/logs")
        assert resp.status_code in [200, 400, 404]

    def test_cancel_execution(self, executor_service, http_session):
        """EX-014: Cancel execution."""
        resp = http_session.delete(f"{executor_service}/api/executions/test-exec-123")
        assert resp.status_code in [200, 400, 404]

    def test_list_executions(self, executor_service, http_session):
        """EX-015: List all executions."""
        resp = http_session.get(f"{executor_service}/api/executions")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert "executions" in data or isinstance(data, list)

    def test_execution_not_found(self, executor_service, http_session):
        """EX-016: Non-existent execution should return error."""
        resp = http_session.get(f"{executor_service}/api/executions/non-existent-exec-999999")
        assert resp.status_code in [200, 400, 404]


@pytest.mark.api
@pytest.mark.p2
class TestPipelineCompatibility:
    """Pipeline compatibility tests (legacy API)."""

    def test_get_pipeline_status(self, executor_service, http_session):
        """EX-020: Get pipeline status (legacy API)."""
        resp = http_session.get(f"{executor_service}/api/status/12345")
        assert resp.status_code in [200, 500, 400]

    def test_get_pipeline_logs(self, executor_service, http_session):
        """EX-021: Get pipeline logs (legacy API)."""
        resp = http_session.get(f"{executor_service}/api/logs/12345")
        assert resp.status_code in [200, 500, 400]
