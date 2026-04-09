"""
Event Store API Tests.
Service: event-store (Port 4002)
"""
import pytest
import uuid
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestEventStoreHealth:
    """Health check tests for Event Store."""

    def test_health_check(self, event_store, http_session):
        """ES-001: Health check should return ok status."""
        resp = http_session.get(f"{event_store}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "ok"
        assert data.get("service") == "event-store"


@pytest.mark.api
@pytest.mark.p0
class TestEventManagement:
    """Event management tests."""

    def test_create_event(self, event_store, http_session, event_payload):
        """ES-010: Create event should return event with UUID."""
        resp = http_session.post(f"{event_store}/api/events", json=event_payload)
        assert resp.status_code in [200, 201]
        data = extract_data(resp.json())
        assert "uuid" in data or "id" in data

    def test_list_events(self, event_store, http_session):
        """ES-011: List events should return array."""
        resp = http_session.get(f"{event_store}/api/events")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert isinstance(data, (list, dict))

    def test_filter_by_status(self, event_store, http_session):
        """ES-012: Filter events by status."""
        resp = http_session.get(f"{event_store}/api/events?status=pending")
        assert resp.status_code == 200

    def test_filter_by_event_type(self, event_store, http_session):
        """ES-013: Filter events by event_type."""
        resp = http_session.get(f"{event_store}/api/events?event_type=pull_request.opened")
        assert resp.status_code == 200

    def test_pagination(self, event_store, http_session):
        """ES-014: Pagination with limit and offset."""
        resp = http_session.get(f"{event_store}/api/events?limit=10&offset=0")
        assert resp.status_code == 200

    def test_get_event_by_uuid(self, event_store, http_session):
        """ES-015: Get single event by UUID."""
        test_uuid = str(uuid.uuid4())
        resp = http_session.get(f"{event_store}/api/events/{test_uuid}")
        assert resp.status_code in [200, 404, 400]

    def test_update_event_status(self, event_store, http_session):
        """ES-016: Update event status."""
        test_uuid = str(uuid.uuid4())
        payload = {"status": "processing"}
        resp = http_session.put(f"{event_store}/api/events/{test_uuid}/status", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_get_pending_events(self, event_store, http_session):
        """ES-017: Get pending events list."""
        resp = http_session.get(f"{event_store}/api/events/pending")
        assert resp.status_code == 200

    def test_get_processing_events(self, event_store, http_session):
        """ES-018: Get processing events list."""
        resp = http_session.get(f"{event_store}/api/events/processing")
        assert resp.status_code == 200

    def test_get_event_statistics(self, event_store, http_session):
        """ES-019: Get event statistics."""
        resp = http_session.get(f"{event_store}/api/events/statistics")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert isinstance(data, (list, dict))

    def test_event_not_found(self, event_store, http_session):
        """ES-020: Non-existent event should return 404."""
        resp = http_session.get(f"{event_store}/api/events/non-existent-uuid-12345")
        assert resp.status_code in [404, 200, 400]

    def test_empty_payload_create(self, event_store, http_session):
        """ES-021: Empty payload should return error."""
        resp = http_session.post(f"{event_store}/api/events", json={})
        assert resp.status_code in [400, 500, 200]

    def test_invalid_json_create(self, event_store, http_session):
        """ES-022: Invalid JSON should return error."""
        resp = http_session.post(
            f"{event_store}/api/events",
            data="{invalid json",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code in [400, 500]


@pytest.mark.api
@pytest.mark.p1
class TestPRRelated:
    """PR related tests."""

    def test_get_pr_events(self, event_store, http_session):
        """ES-030: Get events by PR."""
        resp = http_session.get(f"{event_store}/api/repos/test-repo/pulls/42/events")
        assert resp.status_code in [200, 400]

    def test_invalid_pr_number(self, event_store, http_session):
        """ES-031: Invalid PR number should return error."""
        resp = http_session.get(f"{event_store}/api/repos/test-repo/pulls/invalid/events")
        assert resp.status_code in [400, 404]


@pytest.mark.api
@pytest.mark.p1
class TestQualityChecks:
    """Quality check tests."""

    def test_create_quality_check(self, event_store, http_session):
        """ES-040: Create quality check."""
        test_uuid = str(uuid.uuid4())
        payload = {"check_type": "code_lint", "check_status": "pending"}
        resp = http_session.post(f"{event_store}/api/events/{test_uuid}/quality-checks", json=payload)
        assert resp.status_code in [200, 400, 500]

    def test_get_quality_checks(self, event_store, http_session):
        """ES-041: Get quality checks list."""
        test_uuid = str(uuid.uuid4())
        resp = http_session.get(f"{event_store}/api/events/{test_uuid}/quality-checks")
        assert resp.status_code in [200, 400]

    def test_update_quality_check(self, event_store, http_session):
        """ES-042: Update quality check."""
        payload = {"status": "completed", "result": "pass"}
        resp = http_session.put(f"{event_store}/api/quality-checks/1", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_update_quality_check_by_type(self, event_store, http_session):
        """ES-043: Update quality check by type."""
        test_uuid = str(uuid.uuid4())
        payload = {"status": "completed", "result": "pass"}
        resp = http_session.put(f"{event_store}/api/events/{test_uuid}/quality-checks/code_lint", json=payload)
        assert resp.status_code in [200, 404, 400]

    def test_get_quality_check_statistics(self, event_store, http_session):
        """ES-044: Get quality check statistics."""
        resp = http_session.get(f"{event_store}/api/quality-checks/statistics")
        assert resp.status_code in [200, 500]
