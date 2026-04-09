# Changelog

All notable changes to the Dual-Mass Engine Gateway project will be documented in this file.

## [1.0.0] - 2026-03-20

### New Features

#### Resource Pool Management Module
- **New Module**: Added complete resource pool management system (`src/modules/resource-pool/`)
  - Resource instance lifecycle management (create, allocate, release, recycle)
  - Testbed management with status tracking and health monitoring
  - Category management for resource organization
  - Quota policy management with flexible rules
  - Allocation history tracking and auditing
  - Deployment pipeline template management for Azure DevOps
  - Metrics dashboard for resource utilization monitoring

#### Resource Pool Scheduler
- **Core Scheduler Logic**: Implemented intelligent resource scheduling system
  - Automatic resource matching based on task requirements
  - Priority-based allocation with quota enforcement
  - Resource health monitoring and auto-recovery
  - Concurrent job processing with configurable limits

#### Deployment Integration
- **Deployment Service**: Full integration with Azure DevOps
  - Product deployment workflow
  - Snapshot creation and rollback capabilities
  - SSH-based remote deployment support
  - Deployment task tracking and status management

#### AI Request Pool
- **Global Concurrency Control**: Designed AI request pool to prevent LLM overload
  - Configurable pool size (default 50, max 200)
  - Per-event concurrency limits (default 20, max 50)
  - Automatic request queuing and waiting

### Improvements

#### Frontend UI/UX Optimization
- **Resource Pool Pages**: Complete UI/UX redesign using Data-Dense Dashboard pattern
  - Gradient headers with visual hierarchy
  - Responsive design for all screen sizes (375px - 1440px)
  - Enhanced status indicators and badges
  - Improved form layouts with validation feedback
  - Donut charts for resource utilization visualization
  - Consistent design language across all 11 resource pool pages

#### Event Receiver Module
- Refactored event-receiver module architecture
- Added webhook event ID return for better tracking
- Enhanced event status management with `cancelled` state support

### Bug Fixes

- Fixed deployment product and snapshot rollback issues
- Fixed resource pool scheduler core logic bugs
- Fixed quota policy creation API implementation
- Fixed concurrent AI analysis calls causing overload
- Fixed unit_test score always returning 0
- Fixed basic_ci_all status update issues
- Fixed AI analysis infinite loop
- Fixed quality checks auto-marking as pass incorrectly
- Fixed events author field being empty
- Fixed AI matching incorrect resources
- Fixed AI log analysis issues
- Fixed quality checks status update bugs
- Fixed recycle log bugs
- Fixed PR synchronize event special handling logic
- Fixed admin auto-logout on page refresh
- Fixed mock deployment runtime configuration

### Documentation

- Added deployment workflow documentation
- Added resource instance and testbed lifecycle documentation
- Added resource pool design documentation
- Added Azure DevOps integration guide
- Updated project structure documentation

### Testing

- Added comprehensive unit tests for resource pool module
- Added E2E tests for resource lifecycle
- Added performance benchmark tests
- Added API integration tests

---

## [0.9.0] - 2026-02-15

### New Features

#### Event Processor Module
- **Initial Implementation**: Basic framework for event processing
  - Event fetching from Event Receiver (30s interval)
  - Task scheduling and execution
  - Azure DevOps pipeline integration
  - AI-powered log analysis
  - Resource matching for tasks

#### Event Receiver Module
- **GitHub Webhook Integration**: Receive and process GitHub events
  - Pull Request event handling
  - Push event handling
  - Smart event filtering (main branch only)
  - Quality check creation and management

### Architecture

- Dual-engine architecture with external/internal network separation
- Event Receiver: External network (10.4.111.141:5001)
- Event Processor: Internal network (10.4.174.125:5002)

---

## [0.1.0] - 2026-01-01

### Added

- Initial project structure
- Basic README documentation
- Project initialization
