## Requirements

### Requirement: Cluster Topology Ownership

The host SHALL expose cluster mode, primary judgment, and node ID through `cluster.Service`. All consumers read from this abstraction, not independent package-level state.

#### Scenario: Plugin runtime reads cluster topology
- **WHEN** the plugin runtime needs cluster mode, primary status, or node ID
- **THEN** it reads from the injected topology abstraction
- **AND** `plugin` does not maintain independent cluster state

### Requirement: Election Encapsulation

The host SHALL treat leader election as an internal implementation detail of `cluster.Service`.

#### Scenario: Cluster mode starts topology services
- **WHEN** the host starts in cluster mode
- **THEN** `cluster.Service` internally manages election lifecycle
- **AND** callers do not depend on a standalone `election` service
