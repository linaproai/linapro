## Requirements

### Requirement: Cron Job Classification

The system SHALL support primary-only tasks and all-node tasks. Single-node mode executes primary-only tasks directly.

#### Scenario: Primary-only task in single-node mode
- **WHEN** `cluster.enabled=false` and a primary-only task triggers
- **THEN** the current node executes it

#### Scenario: Primary-only task in cluster mode
- **WHEN** `cluster.enabled=true` and current node is primary
- **THEN** the task executes
- **WHEN** current node is follower
- **THEN** the task is skipped

### Requirement: Existing Cron Job Classification
- Session Cleanup: primary-only
- Server Monitor Collector: all-node
- Server Monitor Cleanup: primary-only
