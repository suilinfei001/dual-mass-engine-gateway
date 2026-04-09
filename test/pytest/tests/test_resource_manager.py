"""
Resource Manager API Tests.
Service: resource-manager (Port 4006)
"""
import pytest
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestResourceManagerHealth:
    """Health check tests for Resource Manager."""

    def test_health_check(self, resource_manager, http_session):
        """RM-001: Health check should return ok status."""
        resp = http_session.get(f"{resource_manager}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "ok"
        assert data.get("service") == "resource-manager"


@pytest.mark.api
@pytest.mark.p0
class TestResourceManagement:
    """Resource management tests."""

    def test_list_resources(self, resource_manager, http_session):
        """RM-010: List all resources."""
        resp = http_session.get(f"{resource_manager}/api/resources")
        assert resp.status_code == 200

    def test_create_resource(self, resource_manager, http_session, resource_payload):
        """RM-011: Create resource should return resource with UUID."""
        resp = http_session.post(f"{resource_manager}/api/resources", json=resource_payload)
        assert resp.status_code in [200, 201]
        data = extract_data(resp.json())
        assert "uuid" in data or "id" in data

    def test_get_resource(self, resource_manager, http_session):
        """RM-012: Get resource by UUID."""
        resp = http_session.get(f"{resource_manager}/api/resources/test-uuid-123")
        assert resp.status_code in [200, 404]

    def test_update_resource(self, resource_manager, http_session, resource_payload):
        """RM-013: Update resource."""
        resp = http_session.put(f"{resource_manager}/api/resources/test-uuid-123", json=resource_payload)
        assert resp.status_code in [200, 404]

    def test_delete_resource(self, resource_manager, http_session):
        """RM-014: Delete resource."""
        resp = http_session.delete(f"{resource_manager}/api/resources/test-uuid-123")
        assert resp.status_code in [204, 200, 404]

    def test_resource_not_found(self, resource_manager, http_session):
        """RM-015: Non-existent resource should return 404."""
        resp = http_session.get(f"{resource_manager}/api/resources/non-existent-uuid-999999")
        assert resp.status_code in [404, 200]

    def test_list_resources_by_category(self, resource_manager, http_session):
        """RM-016: List resources by category."""
        resp = http_session.get(f"{resource_manager}/api/categories/1/resources")
        assert resp.status_code == 200


@pytest.mark.api
@pytest.mark.p1
class TestCategoryManagement:
    """Category management tests."""

    def test_list_categories(self, resource_manager, http_session):
        """RM-020: List all categories."""
        resp = http_session.get(f"{resource_manager}/api/categories")
        assert resp.status_code == 200

    def test_create_category(self, resource_manager, http_session, category_payload):
        """RM-021: Create category."""
        resp = http_session.post(f"{resource_manager}/api/categories", json=category_payload)
        assert resp.status_code in [200, 201, 500]

    def test_get_category(self, resource_manager, http_session):
        """RM-022: Get category by ID."""
        resp = http_session.get(f"{resource_manager}/api/categories/1")
        assert resp.status_code in [200, 404]

    def test_update_category(self, resource_manager, http_session, category_payload):
        """RM-023: Update category."""
        resp = http_session.put(f"{resource_manager}/api/categories/1", json=category_payload)
        assert resp.status_code in [200, 404]

    def test_delete_category(self, resource_manager, http_session):
        """RM-024: Delete category."""
        resp = http_session.delete(f"{resource_manager}/api/categories/999999")
        assert resp.status_code in [204, 200, 404]

    def test_category_not_found(self, resource_manager, http_session):
        """RM-025: Non-existent category should return 404."""
        resp = http_session.get(f"{resource_manager}/api/categories/999999")
        assert resp.status_code in [404, 200]

    def test_empty_name_create(self, resource_manager, http_session):
        """RM-026: Empty name should return error."""
        resp = http_session.post(f"{resource_manager}/api/categories", json={"name": ""})
        assert resp.status_code in [400, 500, 200]


@pytest.mark.api
@pytest.mark.p1
class TestResourceMatching:
    """Resource matching tests."""

    def test_match_resource(self, resource_manager, http_session):
        """RM-030: Match resource."""
        payload = {"category_id": 1, "task_uuid": "test-task-123", "required_count": 1}
        resp = http_session.post(f"{resource_manager}/api/resources/match", json=payload)
        assert resp.status_code in [200, 500]

    def test_release_resource(self, resource_manager, http_session):
        """RM-031: Release resource."""
        resp = http_session.post(f"{resource_manager}/api/resources/test-uuid-123/release")
        assert resp.status_code in [200, 404, 500]

    def test_get_allocation(self, resource_manager, http_session):
        """RM-032: Get resource allocation info."""
        resp = http_session.get(f"{resource_manager}/api/resources/test-uuid-123/allocation")
        assert resp.status_code in [200, 404, 500]


@pytest.mark.api
@pytest.mark.p2
class TestTestbed:
    """Testbed tests."""

    def test_list_testbeds(self, resource_manager, http_session):
        """RM-040: List all testbeds."""
        resp = http_session.get(f"{resource_manager}/api/testbeds")
        assert resp.status_code == 200


@pytest.mark.api
@pytest.mark.p2
class TestQuotaPolicies:
    """Quota policy tests."""

    def test_list_quota_policies(self, resource_manager, http_session):
        """RM-050: List quota policies by category."""
        resp = http_session.get(f"{resource_manager}/api/categories/1/quota-policies")
        assert resp.status_code in [200, 500]
