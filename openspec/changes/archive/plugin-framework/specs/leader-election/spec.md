## Requirements

### Requirement: Leader Election Startup

The system SHALL only participate in leader election when `cluster.enabled=true`. Single-node mode does not start election.

#### Scenario: Single-node skips election
- **WHEN** `cluster.enabled=false`
- **THEN** no election loop starts and current node runs as primary

### Requirement: Lease Auto-Renewal

The system SHALL only renew leases in cluster mode when the current node is primary.

### Requirement: Primary Status Query

Single-node mode always returns `true` for primary status. Cluster mode returns the election result.
