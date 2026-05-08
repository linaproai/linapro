## ADDED Requirements

### Requirement: Host must provide topology-aware cache coordination

The system SHALL provide unified cache coordination for publishing explicit scoped revisions for any legal cache domain, synchronizing in-process derived caches, and differentiating single-node and cluster strategies according to `cluster.enabled`.

#### Scenario: Single-node mode uses local coordination

- **WHEN** `cluster.enabled=false` and a business write path publishes a cache change
- **THEN** the system only updates the local revision in the current process
- **AND** the corresponding cache domain in the current process is invalidated or refreshed immediately
- **AND** the system MUST NOT depend on a shared revision table or distributed coordination component

#### Scenario: Cluster mode uses shared revisions

- **WHEN** `cluster.enabled=true` and a business write path publishes a cache change
- **THEN** the system persistently increments the shared revision for the corresponding cache domain and scope
- **AND** all nodes refresh local cache after observing the new revision on a request path or watcher path
- **AND** revision publishing must be idempotent, retryable, and observable

### Requirement: Shared cache revisions must be persistent and atomically incremented

The system SHALL store critical cache-domain revisions in persistent shared storage and ensure concurrent increments for the same cache domain and scope are not lost.

#### Scenario: Concurrent revision publishing for the same scope

- **WHEN** multiple nodes concurrently publish changes for the same cache domain and scope
- **THEN** the system generates a monotonically increasing revision for every successful publish
- **AND** the final shared revision increases by at least the number of successful publishes
- **AND** no node may overwrite another node's increment through a read-modify-write race

#### Scenario: Revisions remain available after database restart

- **WHEN** the shared database restarts and recovers
- **THEN** committed cache revisions still exist
- **AND** newly started nodes can use persistent revisions to determine whether local cache must be refreshed

#### Scenario: Lossy cache must not carry critical revisions

- **WHEN** the system publishes revisions for critical cache domains such as permissions, runtime configuration, or plugin runtime
- **THEN** the system writes to the persistent revision table
- **AND** MUST NOT store critical revisions in `sys_kv_cache` or any other lossy cache

### Requirement: Cache-domain policy configuration must not gate usage

The system SHALL allow callers to publish and read revisions for any legal cache-domain string directly. It MUST NOT require prior cache-domain registration before a domain participates in coordination. Critical cache domains SHALL declare authoritative data source, consistency model, invalidation trigger, maximum tolerated staleness, cross-instance synchronization mechanism, and failure fallback in their owning implementation code. Unconfigured domains SHALL use the component default policy.

#### Scenario: Use an unconfigured policy domain

- **WHEN** host module or plugin logic publishes a revision for a new legal cache-domain string
- **THEN** the system accepts the domain and uses default consistency and failure policy
- **AND** the caller does not need to modify `cachecoord` component source or delivery manifest to add that domain

#### Scenario: Configure critical cache-domain policy

- **WHEN** a host critical cache domain needs a staleness window or fallback behavior different from the default
- **THEN** that domain's implementation code configures authoritative source and maximum tolerated staleness
- **AND** that domain's implementation code configures refresh-failure fallback behavior
- **AND** review can use that configuration to determine whether the domain satisfies cluster consistency requirements

#### Scenario: Critical cache exceeds staleness window

- **WHEN** a node in cluster mode cannot read the shared revision and local cache exceeds the domain's maximum staleness window
- **THEN** the system handles the request according to that domain's failure policy
- **AND** permission caches MUST NOT silently allow requests after the failure window is exceeded

### Requirement: Critical write paths must reliably publish invalidation

Critical write paths for permissions, configuration, plugin runtime stable state, and equivalent domains MUST reliably publish the corresponding cache-domain revision after business data changes succeed. If publishing fails, callers MUST NOT receive silent success.

#### Scenario: Publish cache revision inside the transaction

- **WHEN** the business data change and cache revision publishing can use the same database transaction
- **THEN** the system commits the business data and revision increment in the same transaction
- **AND** there is no state where business data commits successfully but the revision is missing

#### Scenario: Publishing failure returns an error

- **WHEN** a critical write path completes business data change but cache revision publishing fails
- **THEN** the system returns a structured business error
- **AND** the system records observable logs
- **AND** the caller can retry the operation or trigger a repair flow

### Requirement: Cache coordination state must be observable

The system SHALL expose cache coordination state with at least cache domain, scope, local revision, shared revision, last sync time, latest error, and stale seconds.

#### Scenario: Query cache coordination state

- **WHEN** operations tooling or health checks query cache coordination state
- **THEN** the system returns synchronization state for configured or touched cache domains
- **AND** cluster mode can identify whether a node lags behind the shared revision

#### Scenario: Cache synchronization failure is diagnosable

- **WHEN** a node fails to refresh a cache domain
- **THEN** the system records the latest failure reason and time
- **AND** subsequent state queries can show that domain as abnormal or stale

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

### Requirement: Permission topology cache must reliably invalidate across nodes

After role, menu, user-role, plugin permission menu, or permission resource relationships change, the system SHALL reliably invalidate token permission snapshots on all nodes through the unified cache coordination mechanism.

#### Scenario: Role menu permission changes

- **WHEN** an administrator updates role menu or button permissions
- **THEN** the system commits role permission relationship changes
- **AND** reliably publishes a permission topology cache revision
- **AND** all nodes discard old token permission snapshots within the staleness window allowed by the permission cache domain

#### Scenario: Menu permission identifier changes

- **WHEN** an administrator creates, updates, deletes, or disables menu permissions
- **THEN** the system publishes a permission topology cache revision
- **AND** later protected API permission checks MUST NOT keep using old menu permission topology indefinitely

#### Scenario: Plugin permission topology changes

- **WHEN** plugin install, enable, disable, uninstall, or synchronization changes plugin menus or button permissions
- **THEN** the system publishes a permission topology cache revision
- **AND** affected permissions participate in authorization decisions on all nodes according to the latest plugin state

### Requirement: Permission topology invalidation publish failure must not silently succeed

Critical permission topology write paths MUST return structured errors or roll back business changes when they cannot publish a permission cache revision. This prevents cluster nodes from continuing to use old authorization snapshots.

#### Scenario: Permission revision publishing fails

- **WHEN** a role, menu, or user-role write path needs to publish a permission topology revision but publishing fails
- **THEN** the system returns a structured business error
- **AND** callers MUST NOT receive a success response claiming the permission change is fully effective
- **AND** the system records the failure reason for retry or repair

#### Scenario: Protected API sees stale permission cache

- **WHEN** a protected API validates permissions, cannot confirm local permission snapshot freshness, and exceeds the failure window
- **THEN** the system rejects the request according to fail-closed policy
- **AND** the system MUST NOT continue allowing uncertain permissions because of an old local permission snapshot
