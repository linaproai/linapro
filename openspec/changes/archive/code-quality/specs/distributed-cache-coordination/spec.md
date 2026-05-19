## ADDED Requirements

### Requirement: Cache-sensitive service instances must be shared by topology-aware construction boundaries

The system SHALL ensure service instances that depend on cache coordination, shared revision numbers, event subscriptions, distributed KV, distributed locks, session hot state, token state, or locally derived caches are created and shared by topology-aware construction boundaries. When `cluster.enabled=true`, related components MUST use the same coordination-backed backend or the same runtime service instance, and MUST NOT create isolated default instances visible only to the current node.

#### Scenario: Cluster-mode authentication middleware uses shared token and session state

- **WHEN** `cluster.enabled=true` and authentication middleware validates a request
- **THEN** the authentication middleware uses the startup-injected auth service and session store
- **AND** revoked token, pre-token, and session hot state reads use the same coordination-backed source of truth
- **AND** the authentication middleware MUST NOT construct an auth/session service graph visible only to the current node

#### Scenario: Plugin runtime cache uses shared plugin governance instance

- **WHEN** plugin management, dynamic routing, source-plugin route registrars, or plugin runtime cache read plugin enabled state
- **THEN** these paths reuse the same plugin governance instance or the same shared revision/event backend
- **AND** they MUST NOT expose disabled or uninstalled plugins because multiple plugin service instances hold different enabled snapshots

#### Scenario: Runtime configuration and i18n cache use shared invalidation paths

- **WHEN** runtime configuration or i18n resources change in cluster mode
- **THEN** the injected shared cachecoord/coordination dependency publishes scoped invalidation
- **AND** consumers reuse the same runtime configuration or i18n service instance to observe refreshes
- **AND** they MUST NOT bypass configured invalidation strategies through local new instances

### Requirement: Cache consistency reviews must cover instance sources

The system SHALL check cache-sensitive service instance sources, sharing boundaries, and failure strategies during cache consistency reviews. Reviews MUST identify whether newly added or modified `New()` calls would create independent cache state or independent subscription state.

#### Scenario: Reviewing new cache service construction

- **WHEN** a change adds or modifies service construction related to authentication, authorization, plugins, configuration, i18n, session, cachecoord, kvcache, lock, or notify
- **THEN** the review confirms the construction comes from startup or registrar explicitly passed shared dependencies
- **AND** the review flags implicit constructions that would cause local state splits

#### Scenario: Reviewing changes with no cache impact

- **WHEN** a change is confirmed to not involve caches, derived state, invalidation, or cross-instance synchronization
- **THEN** the review conclusion explicitly records no cache consistency impact
- **AND** the judgment MUST NOT be omitted
