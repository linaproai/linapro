## MODIFIED Requirements

### Requirement: Dynamic plugins access host distributed cache through authorized namespaces backed by MEMORY table

The system SHALL provide a governed cache service for dynamic plugins. Plugins can access the host generic KV cache foundation only through host-authorized named cache namespaces, and must not receive local cache implementations or other low-level cache clients directly. The generic cache module SHALL hide the underlying implementation through backend/provider abstraction. The current default backend is a MySQL `MEMORY` table, and future backends such as Redis can replace it. All backends SHALL be treated as lossy cache and MUST NOT be authoritative sources for permissions, configuration, plugin stable state, cache revisions, or any other reliable business state.

#### Scenario: Plugin accesses authorized cache namespace

- **WHEN** a plugin calls the cache service to execute `get`, `set`, `delete`, `incr`, or `expire`
- **THEN** the host only allows access to the current plugin's authorized `host-cache` resources
- **AND** the host executes the operation according to that cache namespace's naming rules and backend-agnostic TTL policy
- **AND** the default MySQL backend stores cache data in the shared database `MEMORY` cache table, not in host process-local cache

#### Scenario: Plugin cache is lost after database restart

- **WHEN** shared database restart clears the `MEMORY` cache table
- **THEN** plugin cache reads are handled as cache misses
- **AND** the system MUST NOT rely on `sys_kv_cache` to restore critical business state or cache revisions

#### Scenario: Plugin writes cache value exceeding field length limits

- **WHEN** a plugin calls the cache service to write data that exceeds namespace, cache key, or cache value length limits
- **THEN** the host returns an explicit error
- **AND** the host MUST NOT truncate the write
- **AND** the host MUST NOT write partial data

#### Scenario: Plugin attempts to access unauthorized cache namespace

- **WHEN** a plugin calls an unauthorized cache namespace
- **THEN** the host rejects the call
- **AND** the host does not expose underlying cache connection information to the guest

## ADDED Requirements

### Requirement: Plugin cache must not coordinate critical revisions

The system SHALL use an independent persistent revision mechanism to coordinate critical cache domains such as permissions, configuration, and plugin runtime. It MUST NOT store shared revisions for those domains in `sys_kv_cache`.

#### Scenario: Publish critical cache revision

- **WHEN** permission, runtime configuration, or plugin runtime critical cache domains publish a revision
- **THEN** the system writes to persistent revision storage
- **AND** the system MUST NOT write that critical cache-domain revision to `sys_kv_cache`

#### Scenario: Plugin cache clearing does not affect critical cache coordination

- **WHEN** `sys_kv_cache` becomes empty because of database restart or cache cleanup
- **THEN** committed critical cache revisions remain readable from persistent revision storage
- **AND** nodes can still determine whether local permission, configuration, and plugin runtime caches need refresh

### Requirement: Plugin cache increment must be atomic while the cache is alive

The system SHALL guarantee that `incr` for the same plugin cache key increments linearly while the shared database and cache table are alive. After database restart causes `MEMORY` cache loss, later increments may restart from the new cache value.

#### Scenario: Multiple nodes increment the same cache key concurrently

- **WHEN** multiple nodes concurrently execute `incr` on the same plugin cache key
- **THEN** every successful call returns a unique incremented integer value
- **AND** the final cache value equals the initial value plus the sum of all successful increments
- **AND** no node may lose increments through a read-modify-write race

#### Scenario: Increment a non-integer cache value

- **WHEN** a plugin executes `incr` on an existing string cache key
- **THEN** the host returns a structured error
- **AND** the original cache value remains unchanged

### Requirement: Plugin cache expiration cleanup must avoid hot-path full table scans

When reading plugin cache, the system SHALL execute read-only queries only. It MUST NOT delete data in the query request just because a cache entry is expired. Expiration cleanup must be handled by backend expiration filtering on read results and by background batch cleanup or write-path replacement.

#### Scenario: Read an expired cache key

- **WHEN** a plugin reads an expired cache key
- **THEN** the host returns a cache miss
- **AND** the host MUST NOT delete the cache row during that query request
- **AND** the host MUST NOT require this read to clean expired cache for any namespace

#### Scenario: Background batch cleanup removes expired cache

- **WHEN** the expired-cache batch cleanup task is triggered
- **THEN** the system deletes expired cache rows
- **AND** the default MySQL backend SHALL provide a built-in scheduled task that calls `CleanupExpired` once per hour
- **AND** in cluster mode, the task MUST NOT create uncontrolled duplicate pressure across multiple nodes
- **AND** backends that do not require external expiration cleanup, such as Redis, can implement `CleanupExpired` as a no-op and need not project the MySQL cleanup task
