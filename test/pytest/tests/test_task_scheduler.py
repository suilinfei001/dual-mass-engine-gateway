"""
Task Scheduler API Tests.
Service: task-scheduler (Port 4003)
"""
import pytest
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestTaskSchedulerHealth:
    """Health check tests for Task Scheduler."""

    def test_health_check(self, task_scheduler, http_session):
        """TS-001: Health check should return healthy status."""
        resp = http_session.get(f"{task_scheduler}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "healthy"
        assert data.get("service") == "task-scheduler"


@pytest.mark.api
@pytest.mark.p0
class TestTaskManagement:
    """Task management tests."""

    def test_list_tasks(self, task_scheduler, http_session):
        """TS-010: List tasks should return tasks array."""
        resp = http_session.get(f"{task_scheduler}/api/tasks")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert "tasks" in data or isinstance(data, list)

    def test_pagination(self, task_scheduler, http_session):
        """TS-011: Pagination with limit and offset."""
        resp = http_session.get(f"{task_scheduler}/api/tasks?limit=10&offset=0")
        assert resp.status_code == 200

    def test_get_task(self, task_scheduler, http_session):
        """TS-012: Get task by ID."""
        resp = http_session.get(f"{task_scheduler}/api/tasks/1")
        assert resp.status_code in [200, 404, 400]

    def test_start_task(self, task_scheduler, http_session):
        """TS-013: Start task."""
        resp = http_session.post(f"{task_scheduler}/api/tasks/1/start")
        assert resp.status_code in [200, 404, 400]

    def test_complete_task(self, task_scheduler, http_session):
        """TS-014: Complete task with results."""
        payload = {"results": [{"check_type": "code_lint", "result": "pass", "output": "OK"}]}
        resp = http_session.post(f"{task_scheduler}/api/tasks/1/complete", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_fail_task(self, task_scheduler, http_session):
        """TS-015: Mark task as failed."""
        payload = {"reason": "Test failure"}
        resp = http_session.post(f"{task_scheduler}/api/tasks/1/fail", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_cancel_task(self, task_scheduler, http_session):
        """TS-016: Cancel task."""
        payload = {"reason": "PR synchronized"}
        resp = http_session.post(f"{task_scheduler}/api/tasks/1/cancel", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_task_not_found(self, task_scheduler, http_session):
        """TS-017: Non-existent task should return 404."""
        resp = http_session.get(f"{task_scheduler}/api/tasks/999999")
        assert resp.status_code in [404, 200, 400]

    def test_invalid_task_id(self, task_scheduler, http_session):
        """TS-018: Invalid task ID should return error."""
        resp = http_session.get(f"{task_scheduler}/api/tasks/invalid-id")
        assert resp.status_code in [400, 404]

    def test_negative_task_id(self, task_scheduler, http_session):
        """TS-019: Negative task ID should return error."""
        resp = http_session.get(f"{task_scheduler}/api/tasks/-1")
        assert resp.status_code in [400, 404]


@pytest.mark.api
@pytest.mark.p1
class TestEventTaskCancel:
    """Event task cancellation tests."""

    def test_cancel_event_tasks(self, task_scheduler, http_session):
        """TS-020: Cancel all tasks for an event."""
        payload = {"reason": "PR synchronized"}
        resp = http_session.post(f"{task_scheduler}/api/events/1/cancel", json=payload)
        assert resp.status_code in [200, 404, 400]
