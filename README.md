# Dual-Mass Engine Gateway

A comprehensive quality gateway system consisting of two independent modules:

1. **Event Receiver** - Deployed on external network (10.4.111.141), receives GitHub Webhook events
2. **Event Processor** - Deployed on internal network (10.4.174.125), processes quality check tasks with Azure DevOps integration and AI-powered analysis

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Sequence Diagrams](#sequence-diagrams)
- [Event Receiver](#event-receiver)
- [Event Processor](#event-processor)
- [Deployment](#deployment)
- [Development](#development)
- [Testing](#testing)

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        External Network                              │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  GitHub Webhook                                               │  │
│  │       │                                                       │  │
│  │       ▼                                                       │  │
│  │  ┌─────────────────┐                                         │  │
│  │  │ Event Receiver  │  (10.4.111.141:5001)                    │  │
│  │  │  - Receive events                                         │  │
│  │  │  - Store events                                           │  │
│  │  │  - Create quality checks                                  │  │
│  │  └─────────────────┘                                         │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ REST API
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Internal Network                              │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Event Processor (10.4.174.125)                              │  │
│  │  ┌────────────────────────────────────────────────────────┐   │  │
│  │  │ AI Resource Matching                                    │   │  │
│  │  │ - Match tasks to Azure resources                        │   │  │
│  │  │ - Support skip-mode resources                            │   │  │
│  │  └────────────────────────────────────────────────────────┘   │  │
│  │  ┌────────────────────────────────────────────────────────┐   │  │
│  │  │ Azure DevOps Integration                               │   │  │
│  │  │ - Execute pipelines on Azure                            │   │  │
│  │  │ - Fetch build logs                                      │   │  │
│  │  │ - Monitor build status                                 │   │  │
│  │  └────────────────────────────────────────────────────────┘   │  │
│  │  ┌────────────────────────────────────────────────────────┐   │  │
│  │  │ AI Log Analysis                                        │   │  │
│  │  │ - Concurrent log file analysis                         │   │  │
│  │  │ - Smart result merging                                 │   │  │
│  │  │ - Request pool management (global concurrency control)  │   │  │
│  │  └────────────────────────────────────────────────────────┘   │  │
│  │  ┌────────────────────────────────────────────────────────┐   │  │
│  │  │ Task Scheduler                                         │   │  │
│  │  │ - Sequential task execution                            │   │  │
│  │  │ - Auto-create next task on completion                 │   │  │
│  │  └────────────────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Azure DevOps                                             │  │
│  │  - Pipeline execution                                     │  │
│  │  - Build log storage                                      │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Sequence Diagrams

### 1. Event Processing with Azure DevOps

```
GitHub          Event Receiver      Event Processor      Azure DevOps      Database
  │                    │                    │                   │              │
  │──Webhook Event────>│                    │                   │              │
  │                    │──Create Event────────────────────────────────────────>│
  │                    │──Create Quality Checks────────────────────────────────>│
  │                    │                    │                   │              │
  │                    │                    │<──GET /api/events─│              │
  │                    │───Events List──────>│                   │              │
  │                    │                    │                   │              │
  │                    │                    │──AI Match Resources──────────────>│
  │                    │                    │<──Resource Info───│              │
  │                    │                    │                   │              │
  │                    │                    │──Create Task──────────────────>│
  │                    │                    │                   │              │
  │                    │                    │──Run Pipeline─────>│              │
  │                    │                    │                   │              │
  │                    │                    │<──Build ID/URL────│              │
  │                    │                    │                   │              │
  │                    │                    │──[Monitor Status]──>│              │
  │                    │                    │<──Status──────────│              │
  │                    │                    │                   │              │
  │                    │                    │──Fetch Logs────────>│              │
  │                    │                    │<──Log Content─────│              │
  │                    │                    │                   │              │
  │                    │                    │──AI Analyze Logs──│              │
  │                    │                    │                   │              │
  │                    │                    │──Save Results──────────────────>│
  │                    │                    │                   │              │
  │                    │<──PUT quality-checks/batch──────────────────────│
  │                    │──Update QC (completed)────────────────────────────>│
```

### 2. AI Resource Matching Flow

```
Event Processor           AI Service         Resource Storage     Database
     │                        │                      │                 │
     ├──GET All Resources────>│                      │                 │
     │                        │<──Resource List──────│                 │
     │                        │                      │                 │
     ├──Match Request───────>│                      │                 │
     │  (event + task)         │                      │                 │
     │                        │──Find Best Match─────>│                 │
     │                        │<──Matched Resource────│                 │
     │                        │                      │                 │
     │<──Resource ID/URL──────│                      │                 │
     │                        │                      │                 │
     ├──Update Task (resource_id, azure://url)────────────────────────────>│
     │                        │                      │                 │
```

### 3. AI Log Analysis Flow

```
Event Processor        AI Request Pool         AI Service        Azure DevOps
     │                        │                        │                  │
     ├──Get Build Logs──────>│                        │                  │
     │<──Log Files────────────│                        │                  │
     │                        │                        │                  │
     ├──Analyze N logs──────>│                        │                  │
     │  (concurrent)           │                        │                  │
     │                        │──Request N tokens─────>│                  │
     │                        │<──Tokens Granted──────│                  │
     │                        │                        │                  │
     │<──Analysis Results─────│                        │                  │
     │                        │                        │                  │
     ├──Release Tokens──────>│                        │                  │
     │                        │                        │                  │
     ├──Merge Results──────────────────────────────────────────────>│
     │                        │                        │                  │
     ├──Save Results──────────────────────────────────────────────────>│
```

### 4. Skip Execution Flow

```
Event Processor        Resource Storage        Database
     │                        │                      │
     ├──Get Resource────────>│                      │
     │<──Resource Info───────│                      │
     │                        │                      │
     ├──Check allow_skip─────>│                      │
     │                        │                      │
     │If allow_skip=true:      │                      │
     │──Mark Task Skipped────────────────────────────────────────>│
     │──Create Next Task────────────────────────────────────────────>│
     │                        │                      │
```

---

## Event Receiver

GitHub Webhook quality check service for monitoring and processing GitHub Pull Request and Push events with automated quality checks.

### Features

- **GitHub Webhook Integration**: Receives and processes GitHub PR and Push events
- **Smart Event Filtering**: Only processes PRs to main branch and main branch pushes
- **Multi-stage Quality Checks**: Supports Basic CI, Deployment, and Specialized Tests
- **MySQL Persistence**: Stores events and quality check data
- **RESTful API**: Complete API for querying and managing data
- **Docker Support**: Containerized deployment

### Quick Start

```bash
cd src/modules/event-receiver
./install_quality.sh
```

### Container Architecture

| Container | Port | Description |
|-----------|------|-------------|
| **quality-server** | 5001 | GitHub webhook quality engine |
| **quality-frontend** | 8081 | Web UI |
| **quality-mysql** | 3306 | MySQL database |

### Quality Check Stages

1. **Basic CI** (stage_order=1)
   - Compilation check
   - Code linting
   - Security scan
   - Unit tests

2. **Deployment** (stage_order=2)
   - Deployment check

3. **Specialized Tests** (stage_order=3)
   - API tests
   - Module E2E tests
   - Agent E2E tests
   - AI E2E tests

### API Reference

#### Event Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/events` | Get event list |
| `GET` | `/api/events/:id` | Get event details |
| `PUT` | `/api/events/:id/status` | Update event status |
| `DELETE` | `/api/events` | Delete all events |

#### Quality Checks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/events/:eventID/quality-checks` | Get quality check list |
| `PUT` | `/api/quality-checks/:id` | Update quality check status |
| `PUT` | `/api/events/:eventID/quality-checks/batch` | Batch update quality checks |

For detailed API documentation, see [Event Receiver README](src/modules/event-receiver/README.md).

---

## Event Processor

Event processor module responsible for processing quality check tasks from Event Receiver with Azure DevOps integration and AI-powered log analysis.

### Features

#### Core Features
- **Event Fetching**: Fetches all events from Event Receiver every 30 seconds
- **AI Resource Matching**: Automatically matches tasks to Azure DevOps resources
- **Azure DevOps Integration**: Executes pipelines on Azure DevOps
- **AI Log Analysis**: Concurrent analysis of build logs with smart result merging
- **Skip Execution**: Resources can be marked as skippable to bypass execution
- **Request Pool Management**: Global AI request pool prevents overload

#### AI Request Pool
- **Global Concurrency Control**: Manages total AI requests across all events
- **Configurable Pool Size**: Default 50, max 200
- **Per-Event Concurrency**: Default 20, max 50 concurrent log files per event
- **Automatic Waiting**: Events wait when pool is exhausted

#### Resource Management
- **Allow Skip**: Resources can be marked as skippable
- **Auto-Skip Tasks**: Tasks matched with skippable resources are marked as skipped
- **Azure Configuration**: Organization, Project, Pipeline ID per resource

### Quick Start

```bash
cd src/modules/event-processor
./deploy-event-processor.sh
```

### Container Architecture

| Container | Port | Description |
|-----------|------|-------------|
| **event-processor-server** | 5003 (external) | Backend API server (internal: 5002) |
| **event-processor-frontend** | 8082 | Frontend Web UI |
| **event-processor-mysql** | 3307 | MySQL database |

### Task Execution Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Event Processor                                  │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 1. Event Fetcher (every 30s)                             │      │
│  │    GET /api/events?status=pending                        │      │
│  └──────────────────────────────────────────────────────────┘      │
│                           │                                        │
│                           ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 2. AI Resource Matching                                   │      │
│  │    - Match task to Azure resources                        │      │
│  │    - Check if resource allows skip                        │      │
│  └──────────────────────────────────────────────────────────┘      │
│                           │                                        │
│                           ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 3. Task Executor - Azure DevOps or Skip                   │      │
│  │    - If allow_skip: mark as skipped                        │      │
│  │    - Else: Run Azure pipeline, fetch logs                  │      │
│  └──────────────────────────────────────────────────────────┘      │
│                           │                                        │
│                           ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 4. AI Log Analysis (concurrent)                           │      │
│  │    - Acquire tokens from global pool                      │      │
│  │    - Analyze each log file with AI                         │      │
│  │    - Smart merge results                                  │      │
│  │    - Release tokens                                        │      │
│  └──────────────────────────────────────────────────────────┘      │
│                           │                                        │
│                           ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 5. Result Writer                                          │      │
│  │    PUT /api/events/:id/quality-checks/batch              │      │
│  └──────────────────────────────────────────────────────────┘      │
│                           │                                        │
│                           ▼                                        │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ 6. Next Task Creator                                      │      │
│  │    - Create next task when current completes              │      │
│  └──────────────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
```

### Configuration

Admin Console (http://localhost:8082) provides configuration for:

| Setting | Description | Range/Default |
|---------|-------------|---------------|
| AI Server IP | AI model server address | - |
| AI Model | Model name for analysis | - |
| AI Token | API authentication token | - |
| Azure PAT | Azure DevOps Personal Access Token | - |
| AI Concurrency (Per Event) | Max concurrent log files per event | 1-50 (default: 20) |
| AI Request Pool Size (Global) | Total concurrent AI requests across all events | 1-200 (default: 50) |
| Log Retention Days | How long to keep log files | 1-30 (default: 7) |

**Important**: AI Request Pool Size must be greater than AI Concurrency. Recommended: not more than 2/3 of your AI model's actual concurrent capacity.

### Executable Resources

Resources define where and how quality checks are executed:

| Field | Description |
|-------|-------------|
| Resource Type | Type of check (basic_ci_all, deployment_deployment, etc.) |
| Allow Skip | If true, task is marked as skipped without execution |
| Organization | Azure DevOps organization |
| Project | Azure DevOps project |
| Pipeline ID | Azure pipeline ID |
| Pipeline Parameters | JSON parameters for pipeline execution |
| Repo Path | Git repository path |
| Creator ID | User who created the resource |

---

## Deployment

### Event Receiver Deployment

Event Receiver is deployed on external network server (10.4.111.141).

```bash
cd src/modules/event-receiver
./install_quality.sh
```

### Event Processor Deployment

Event Processor is deployed on internal network server (10.4.174.125).

```bash
cd src/modules/event-processor
./deploy-event-processor.sh
```

### Deployment Modes

| Mode | Command | Description |
|------|---------|-------------|
| Upgrade | `./deploy-event-processor.sh` | Update containers, preserve data (default) |
| Recovery | `./deploy-event-processor.sh -r` | Complete reinstall, delete old containers and data |

### Access URLs

| Service | URL |
|---------|-----|
| Event Receiver API | http://10.4.111.141:5001 |
| Event Receiver UI | http://10.4.111.141:8081 |
| Event Processor API | http://10.4.174.125:5003 |
| Event Processor UI | http://10.4.174.125:8082 |

---

## Development

### Project Structure

```
src/modules/
├── event-receiver/          # Event Receiver module
│   ├── cmd/server/          # Server entry point
│   ├── internal/            # Internal packages
│   ├── install/             # Installation scripts
│   └── README.md            # Module documentation
│
└── event-processor/         # Event Processor module
    ├── cmd/server/          # Server entry point
    ├── internal/
    │   ├── ai/              # AI matching, log analysis, request pool
    │   ├── api/             # REST API handlers
    │   ├── executor/        # Azure DevOps integration
    │   ├── models/          # Data models
    │   ├── monitor/         # Task monitoring
    │   ├── scheduler/       # Task scheduling
    │   ├── storage/         # Database layer
    │   └── mock/            # Mock server (testing)
    ├── frontend/            # Vue 3 frontend
    │   └── src/views/
    │       ├── Console.vue  # Admin console
    │       ├── Resources.vue # Resource management
    │       ├── Tasks.vue    # Task management
    │       └── Events.vue   # Event viewing
    ├── install/             # Installation scripts
    └── README.md            # Module documentation
```

### Local Development

**Event Receiver:**
```bash
cd src/modules/event-receiver/cmd/server
go run main.go
```

**Event Processor:**
```bash
cd src/modules/event-processor/cmd/server
go run main.go -db "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true"
```

### Running Tests

```bash
# Run all tests
cd src/modules/event-processor
go test ./...

# Run specific package tests
go test ./internal/ai/...
go test ./internal/scheduler/...

# Run tests with verbose output
go test -v ./internal/ai/...
```

### Test Coverage

- **AI Package**: Request pool management, log analysis, smart merge
- **Scheduler Package**: Task scheduling, PR cancellation, sequential execution
- **Storage Package**: Database operations, config management
- **API Package**: REST endpoints, authentication

---

## Documentation

- [Event Receiver README](src/modules/event-receiver/README.md) - Detailed Event Receiver documentation
- [Event Processor README](src/modules/event-processor/README.md) - Detailed Event Processor documentation
- [Project Structure](docs/PROJECT_STRUCTURE.md) - Project structure overview
- [CLAUDE.md](CLAUDE.md) - Development guidelines and project rules

---

## License

This project is licensed under the MIT License.

---

For questions or suggestions, please contact the project maintainers.
