## ADDED Requirements

### Requirement: Cluster mode must use Redis coordination
The system SHALL use Redis coordination as the only supported distributed coordination implementation in PostgreSQL cluster mode. `cluster.enabled=true` MUST coexist with `cluster.coordination=redis` before allowing cluster startup flow.

#### Scenario: PostgreSQL cluster mode enables Redis coordination
- **WHEN** database link is PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis probe succeeds
- **THEN** host enters cluster mode
- **AND** leader election, cache coordination, session hot state, and kvcache all use coordination provider

#### Scenario: PostgreSQL cluster mode missing coordination
- **WHEN** database link is PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination` is missing
- **THEN** host startup fails
- **AND** must not fall back to PostgreSQL table coordination implementation

### Requirement: Standalone mode must not force Redis dependency
The system SHALL keep standalone implementation lean when `cluster.enabled=false`. Standalone mode MUST not start Redis coordination, not register Redis event subscriber, not use Redis lock for primary election.

#### Scenario: Standalone mode preserves process-internal coordination
- **WHEN** `cluster.enabled=false`
- **THEN** current node runs directly as primary node semantics
- **AND** cache revision uses process-internal state
- **AND** kvcache can continue using SQL table backend
- **AND** auth/session does not require Redis

### Requirement: Cluster mode must not use PostgreSQL as cross-node coordination primary implementation
The system SHALL prohibit cluster mode from depending on `sys_locker`, `sys_cache_revision`, or `sys_kv_cache` for cross-node consistency. These tables MAY be retained for standalone, testing, diagnostics, or future fallback implementation.

#### Scenario: Cluster mode cachecoord does not write sys_cache_revision
- **WHEN** `cluster.enabled=true` and `cluster.coordination=redis`
- **AND** business write path publishes cache revision
- **THEN** system uses Redis revision store
- **AND** does not depend on `sys_cache_revision` increment to notify other nodes

#### Scenario: Cluster mode leader election does not write sys_locker
- **WHEN** `cluster.enabled=true` and `cluster.coordination=redis`
- **AND** node participates in primary election
- **THEN** system uses Redis lock store
- **AND** does not depend on `sys_locker` to determine primary
