## ADDED Requirements

### Requirement: Cron Task Classification

The system SHALL support two types of scheduled tasks: Master-Only and All-Node.

#### Scenario: Define Master-Only task

- **WHEN** a scheduled task is registered with Master-Only type
- **THEN** the task executes only on the leader node

#### Scenario: Define All-Node task

- **WHEN** a scheduled task is registered with All-Node type
- **THEN** the task executes on all nodes

### Requirement: Master-Only Task Leader Check

The system SHALL check whether the current node is the leader before executing Master-Only tasks.

#### Scenario: Leader executes Master-Only task

- **WHEN** a Master-Only task triggers and the current node is the leader
- **THEN** the task executes normally

#### Scenario: Follower skips Master-Only task

- **WHEN** a Master-Only task triggers and the current node is a follower
- **THEN** the task returns immediately without executing any business logic

### Requirement: Existing Task Classification

The system SHALL classify existing scheduled tasks.

#### Scenario: Session Cleanup classified as Master-Only

- **WHEN** the Session Cleanup scheduled task is registered
- **THEN** the system marks it as Master-Only type

#### Scenario: Server Monitor Collector classified as All-Node

- **WHEN** the Server Monitor Collector scheduled task is registered
- **THEN** the system marks it as All-Node type

#### Scenario: Server Monitor Cleanup classified as Master-Only

- **WHEN** the Server Monitor Cleanup scheduled task is registered
- **THEN** the system marks it as Master-Only type
