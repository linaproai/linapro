## ADDED Requirements

### Requirement: Protected runtime parameter cache must be bounded-consistent across nodes

The system SHALL synchronize protected runtime parameter cache through the unified cache coordination mechanism so that, in cluster mode, no node keeps using an old parameter snapshot indefinitely.

#### Scenario: Protected runtime parameter changed in cluster mode

- **WHEN** an administrator changes protected runtime parameters
- **THEN** the system commits the parameter change
- **AND** reliably publishes a runtime configuration cache revision
- **AND** other nodes refresh their local parameter snapshots within the staleness window allowed by the runtime configuration cache domain

#### Scenario: Runtime parameter revision publishing fails

- **WHEN** a parameter change requires runtime configuration cache refresh but revision publishing fails
- **THEN** the system returns a structured business error
- **AND** the caller MUST NOT receive a silent success result
- **AND** the system records a retryable failure reason

### Requirement: Runtime parameter reads must execute freshness checks

Before reading protected parameters that affect authentication, sessions, upload, scheduling, or other runtime behavior, the system SHALL verify that the local snapshot has not exceeded the allowed staleness window.

#### Scenario: Local parameter snapshot is already at the latest revision

- **WHEN** a node reads protected runtime parameters and its local revision has consumed the shared revision
- **THEN** the system returns parameters from the local cache snapshot
- **AND** does not requery the complete `sys_config` parameter set

#### Scenario: Local parameter snapshot lags behind shared revision

- **WHEN** a node reads protected runtime parameters and observes a newer shared revision
- **THEN** the system rebuilds the local parameter snapshot from `sys_config`
- **AND** subsequent reads use the snapshot for the new revision

#### Scenario: Freshness cannot be confirmed and the failure window is exceeded

- **WHEN** a node cannot read shared revisions and its local runtime parameter snapshot exceeds the failure window
- **THEN** the system returns a visible error or degrades according to the declared policy for that parameter domain
- **AND** the system MUST NOT silently use the old parameter snapshot indefinitely
