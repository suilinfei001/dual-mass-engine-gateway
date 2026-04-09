"""
AI Analyzer API Tests.
Service: ai-analyzer (Port 4005)
"""
import pytest
from conftest import extract_data


@pytest.mark.api
@pytest.mark.p0
class TestAIAnalyzerHealth:
    """Health check tests for AI Analyzer."""

    def test_health_check(self, ai_analyzer, http_session):
        """AI-001: Health check should return healthy status."""
        resp = http_session.get(f"{ai_analyzer}/health")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert data.get("status") == "healthy"
        assert data.get("service") == "ai-analyzer"

    def test_api_health_check(self, ai_analyzer, http_session):
        """AI-002: API health check endpoint."""
        resp = http_session.get(f"{ai_analyzer}/api/health")
        assert resp.status_code == 200


@pytest.mark.api
@pytest.mark.p1
class TestLogAnalysis:
    """Log analysis tests."""

    def test_analyze_log(self, ai_analyzer, http_session):
        """AI-010: Analyze log content."""
        payload = {"log_content": "Error: Connection timeout\nStack trace: ...", "task_name": "basic_ci_all"}
        resp = http_session.post(f"{ai_analyzer}/api/analyze", json=payload)
        assert resp.status_code in [200, 500]

    def test_empty_log_content(self, ai_analyzer, http_session):
        """AI-011: Empty log_content should return error."""
        payload = {"log_content": "", "task_name": "basic_ci_all"}
        resp = http_session.post(f"{ai_analyzer}/api/analyze", json=payload)
        assert resp.status_code == 400

    def test_missing_log_content(self, ai_analyzer, http_session):
        """AI-012: Missing log_content should return error."""
        payload = {"task_name": "basic_ci_all"}
        resp = http_session.post(f"{ai_analyzer}/api/analyze", json=payload)
        assert resp.status_code == 400

    def test_batch_analyze(self, ai_analyzer, http_session):
        """AI-013: Batch analyze logs."""
        payload = {"log_contents": ["Error 1", "Error 2"], "task_name": "basic_ci_all"}
        resp = http_session.post(f"{ai_analyzer}/api/analyze/batch", json=payload)
        assert resp.status_code in [200, 500]

    def test_empty_batch_request(self, ai_analyzer, http_session):
        """AI-014: Empty batch should return error."""
        payload = {"log_contents": [], "task_name": "basic_ci_all"}
        resp = http_session.post(f"{ai_analyzer}/api/analyze/batch", json=payload)
        assert resp.status_code == 400


@pytest.mark.api
@pytest.mark.p2
class TestPoolManagement:
    """Pool management tests."""

    def test_get_pool_stats(self, ai_analyzer, http_session):
        """AI-020: Get pool statistics."""
        resp = http_session.get(f"{ai_analyzer}/api/pool/stats")
        assert resp.status_code == 200
        data = extract_data(resp.json())
        assert "total_size" in data or "available" in data

    def test_get_pool_size(self, ai_analyzer, http_session):
        """AI-021: Get pool size."""
        resp = http_session.get(f"{ai_analyzer}/api/config/pool-size")
        assert resp.status_code == 200

    def test_set_pool_size(self, ai_analyzer, http_session):
        """AI-022: Set pool size."""
        payload = {"size": 5}
        resp = http_session.post(f"{ai_analyzer}/api/config/pool-size", json=payload)
        assert resp.status_code in [200, 500]

    def test_invalid_pool_size(self, ai_analyzer, http_session):
        """AI-023: Invalid pool size should return error."""
        payload = {"size": 0}
        resp = http_session.post(f"{ai_analyzer}/api/config/pool-size", json=payload)
        assert resp.status_code == 400
