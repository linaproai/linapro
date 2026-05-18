## ADDED Requirements

### Requirement: Leader Election on Startup

The system SHALL automatically participate in leader election on service start, attempting to become the leader node.

#### Scenario: First startup becomes leader

- **WHEN** the service starts and no other leader exists
- **THEN** the system successfully acquires the leader lock, and the current node becomes the leader

#### Scenario: Leader already exists on startup

- **WHEN** the service starts while another node already holds the leader lock and it has not expired
- **THEN** the current node becomes a follower and periodically attempts to acquire leadership

#### Scenario: Failover after leader failure

- **WHEN** the original leader's lock expires
- **THEN** a follower successfully acquires the leader lock on its next attempt, becoming the new leader

### Requirement: Automatic Lease Renewal

The system SHALL provide automatic lease renewal for the leader node, ensuring continuous leadership.

#### Scenario: Periodic renewal succeeds

- **WHEN** the leader periodically executes lease renewal
- **THEN** the leader lock's expiration time is updated, and the leader continues holding leadership

#### Scenario: Renewal failure causes demotion

- **WHEN** the leader's renewal fails (e.g., database failure)
- **THEN** the current node demotes to a follower and stops executing Master-Only Jobs

### Requirement: Leader State Query

The system SHALL provide leader state query functionality to determine whether the current node is the leader.

#### Scenario: Query leader state

- **WHEN** the IsLeader method is called
- **THEN** the system returns a boolean indicating whether the current node is the leader

### Requirement: Task Classification Execution

The system SHALL determine whether to execute a task on the current node based on the task type.

#### Scenario: Master-Only task executes on leader

- **WHEN** a Master-Only scheduled task triggers and the current node is the leader
- **THEN** the task executes normally

#### Scenario: Master-Only task skips on follower

- **WHEN** a Master-Only scheduled task triggers and the current node is a follower
- **THEN** the task is skipped without producing any side effects

#### Scenario: All-Node task executes on all nodes

- **WHEN** an All-Node scheduled task triggers
- **THEN** the task executes normally on all nodes regardless of leader/follower status
