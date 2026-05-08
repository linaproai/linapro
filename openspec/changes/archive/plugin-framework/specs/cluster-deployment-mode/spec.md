## Requirements

### Requirement: Cluster Deployment Mode Configuration

The host SHALL provide `cluster.enabled` as the cluster mode switch, with `cluster.election.lease` and `cluster.election.renewInterval` as sub-config. Default is `false`.

#### Scenario: Default single-node mode
- **WHEN** `cluster.enabled` is not declared
- **THEN** the host starts in single-node mode
- **AND** the current node is treated as primary

#### Scenario: Explicit cluster mode
- **WHEN** `cluster.enabled=true`
- **THEN** the host starts in cluster mode
- **AND** election and primary-specific behavior are controlled by cluster mode

### Requirement: Single-Node Mode Primary Semantics

When `cluster.enabled=false`, the host treats the current node as primary and skips multi-node coordination.

#### Scenario: Skip election infrastructure
- **WHEN** host starts in single-node mode
- **THEN** no leader election loop or lease renewal starts

#### Scenario: Execute primary-only tasks directly
- **WHEN** single-node mode triggers primary-only logic
- **THEN** the current node executes directly without waiting

### Requirement: Plugin Runtime Topology Convergence

The host SHALL control dynamic plugin convergence by deployment mode.

#### Scenario: Single-node synchronous completion
- **WHEN** single-node mode executes plugin operations
- **THEN** the current node completes synchronously

#### Scenario: Cluster mode retains primary convergence
- **WHEN** cluster mode executes plugin operations
- **THEN** the system records target state and primary executes final switch

### Requirement: Node Projection Only in Cluster Mode

The host SHALL only maintain `sys_plugin_node_state` in cluster mode. Single-node mode does not require it as a prerequisite.
