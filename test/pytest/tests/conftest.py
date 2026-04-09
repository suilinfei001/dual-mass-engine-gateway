"""
Pytest configuration and fixtures for API testing.
"""
import pytest
import requests
import os
import uuid


SERVICES = {
    "webhook_gateway": os.getenv("WEBHOOK_GATEWAY_URL", "http://localhost:4001"),
    "event_store": os.getenv("EVENT_STORE_URL", "http://localhost:4002"),
    "task_scheduler": os.getenv("TASK_SCHEDULER_URL", "http://localhost:4003"),
    "executor_service": os.getenv("EXECUTOR_SERVICE_URL", "http://localhost:4004"),
    "ai_analyzer": os.getenv("AI_ANALYZER_URL", "http://localhost:4005"),
    "resource_manager": os.getenv("RESOURCE_MANAGER_URL", "http://localhost:4006"),
}


def extract_data(response_json):
    """Extract data from API response that wraps in {"success": true, "data": ...}."""
    if isinstance(response_json, dict):
        if "data" in response_json:
            return response_json["data"]
        if "success" in response_json and response_json["success"]:
            return response_json.get("data", response_json)
    return response_json


@pytest.fixture(scope="session")
def services():
    """Return service URLs."""
    return SERVICES


@pytest.fixture(scope="session")
def webhook_gateway(services):
    """Webhook Gateway service URL."""
    return services["webhook_gateway"]


@pytest.fixture(scope="session")
def event_store(services):
    """Event Store service URL."""
    return services["event_store"]


@pytest.fixture(scope="session")
def task_scheduler(services):
    """Task Scheduler service URL."""
    return services["task_scheduler"]


@pytest.fixture(scope="session")
def executor_service(services):
    """Executor Service URL."""
    return services["executor_service"]


@pytest.fixture(scope="session")
def ai_analyzer(services):
    """AI Analyzer service URL."""
    return services["ai_analyzer"]


@pytest.fixture(scope="session")
def resource_manager(services):
    """Resource Manager service URL."""
    return services["resource_manager"]


@pytest.fixture(scope="session")
def http_session():
    """Create a requests session for connection pooling."""
    session = requests.Session()
    session.headers.update({"Content-Type": "application/json"})
    yield session
    session.close()


@pytest.fixture
def github_pr_opened_payload():
    """Sample GitHub PR opened webhook payload."""
    return {
        "action": "opened",
        "repository": {
            "id": 12345,
            "name": "test-repo",
            "full_name": "owner/test-repo",
            "html_url": "https://github.com/owner/test-repo",
            "owner": {"login": "owner"},
        },
        "pull_request": {
            "number": 42,
            "title": "Test PR",
            "head": {"ref": "feature-branch", "sha": "abc123def456"},
        },
        "sender": {"login": "testuser"},
    }


@pytest.fixture
def github_pr_synchronize_payload():
    """Sample GitHub PR synchronize webhook payload."""
    return {
        "action": "synchronize",
        "repository": {
            "id": 12345,
            "name": "test-repo",
            "full_name": "owner/test-repo",
            "owner": {"login": "owner"},
        },
        "pull_request": {
            "number": 42,
            "title": "Test PR",
            "head": {"ref": "feature-branch", "sha": "newsha456"},
        },
        "sender": {"login": "testuser"},
    }


@pytest.fixture
def gitlab_mr_opened_payload():
    """Sample GitLab MR opened webhook payload."""
    return {
        "object_kind": "merge_request",
        "event_type": "merge_request",
        "project": {
            "id": 12345,
            "name": "test-project",
            "http_url": "https://gitlab.com/owner/test-project",
        },
        "object_attributes": {
            "iid": 42,
            "title": "Test MR",
            "action": "open",
            "source_branch": "feature-branch",
        },
        "user": {"login": "testuser"},
    }


@pytest.fixture
def event_payload():
    """Sample event creation payload."""
    return {
        "event_type": "pull_request.opened",
        "source": "github",
        "repo_name": "test/repo",
        "repo_owner": "testowner",
        "pr_number": 123,
        "author": "testuser",
        "commit_sha": "abc123def456",
        "status": "pending",
        "payload": "{}",
    }


@pytest.fixture
def resource_payload():
    """Sample resource creation payload."""
    return {
        "name": f"test-resource-pytest-{uuid.uuid4().hex[:8]}",
        "ip_address": "192.168.1.100",
        "ssh_port": 22,
        "ssh_user": "root",
        "ssh_password": "password",
        "category_id": 1,
        "is_public": True,
    }


@pytest.fixture
def category_payload():
    """Sample category creation payload."""
    return {
        "name": f"test-category-pytest-{uuid.uuid4().hex[:8]}",
        "description": "Test category for pytest",
    }


@pytest.fixture
def execution_payload():
    """Sample task execution payload."""
    return {
        "task_uuid": f"test-uuid-pytest-{uuid.uuid4().hex[:8]}",
        "task_type": "basic_ci_all",
        "chart_url": "http://example.com/chart.tgz",
        "testbed_ip": "192.168.1.100",
        "testbed_ssh_port": 22,
        "testbed_ssh_user": "root",
        "testbed_ssh_password": "password",
    }


@pytest.fixture
def analyze_payload():
    """Sample log analysis payload."""
    return {
        "log_content": "Error: Connection timeout\nStack trace: ...",
        "task_name": "basic_ci_all",
    }


def pytest_configure(config):
    """Configure custom markers."""
    config.addinivalue_line("markers", "p0: Core functionality tests (must pass)")
    config.addinivalue_line("markers", "p1: Important functionality tests")
    config.addinivalue_line("markers", "p2: Auxiliary functionality tests")
    config.addinivalue_line("markers", "p3: Boundary condition tests")
    config.addinivalue_line("markers", "e2e: End-to-end tests")
    config.addinivalue_line("markers", "api: API interface tests")
