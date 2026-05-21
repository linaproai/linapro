## ADDED Requirements

### Requirement: Cluster topology must be injected by unified coordination
The system SHALL create coordination provider through unified startup orchestration and inject it into cluster, locker, cachecoord, kvcache, auth, session, cron, and plugin runtime components needing cluster coordination. Business components MUST not parse Redis configuration themselves.

#### Scenario: Startup orchestration injects coordination
- **WHEN** host starts in cluster mode
- **THEN** startup orchestration first creates Redis coordination provider
- **AND** cluster service uses that provider for primary election
- **AND** other components receive provider or provider-backed service through constructor parameters or explicit setters

#### Scenario: Components prohibited from reading Redis configuration themselves
- **WHEN** `role` or `pluginruntimecache` needs to publish cross-node revision
- **THEN** they do so through `cachecoord` or coordination-backed controller
- **AND** do not read `cluster.redis.address`
- **AND** do not create Redis client

### Requirement: Node identity must be threaded through coordination events
The system SHALL carry stable node ID in coordination lock, revision event, plugin runtime event, and health diagnostics. Node ID MUST be uniformly provided by cluster/topology layer.

#### Scenario: Published event contains sourceNode
- **WHEN** node publishes cache invalidation event
- **THEN** event payload contains current node ID
- **AND** receiving node can ignore or diagnose duplicate events from itself

#### Scenario: Health diagnostics contain node ID
- **WHEN** querying system info or health status
- **THEN** response contains current node ID
- **AND** response contains whether current node is primary

### Requirement: Primary determination must be consistent with Redis lock state
The system SHALL use Redis leader lock holding status as authoritative source for `IsPrimary` in cluster mode. After renewal failure or lock loss, `IsPrimary` MUST immediately return false.

#### Scenario: Primary state change after renewal failure
- **WHEN** current primary node cannot renew leader lock
- **THEN** cluster service demotes this node to follower
- **AND** `IsPrimary` returns false
- **AND** master-only tasks stop executing
