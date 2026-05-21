## ADDED Requirements

### Requirement: Cluster mode must declare coordination backend
The system SHALL require `cluster.coordination` to be configured when `cluster.enabled=true`. The current version's only valid value MUST be `redis`. When `cluster.enabled=false`, the system MUST not require `cluster.coordination` to exist, nor must Redis configuration absence affect standalone startup.

#### Scenario: Cluster mode missing coordination
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination` is not declared
- **THEN** host startup fails
- **AND** error message clearly indicates cluster mode must configure `cluster.coordination=redis`

#### Scenario: Cluster mode configured with illegal coordination
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination=postgres`
- **THEN** host startup fails
- **AND** error message clearly indicates currently only `redis` is supported

#### Scenario: Standalone mode does not require Redis
- **WHEN** configuration declares `cluster.enabled=false`
- **AND** `cluster.coordination` is not declared
- **AND** `cluster.redis` is not declared
- **THEN** host starts successfully in standalone mode
- **AND** system must not attempt to connect to Redis

### Requirement: Redis configuration must use cluster namespace
The system SHALL read Redis connection configuration from `cluster.redis` when `cluster.coordination=redis`. Configuration MUST support `address`, `db`, `password`, `connectTimeout`, `readTimeout`, `writeTimeout`. All time durations MUST use unit-bearing duration strings and be parsed as `time.Duration`.

#### Scenario: Redis configuration parsed successfully
- **WHEN** configuration declares `cluster.coordination=redis`
- **AND** `cluster.redis.address="127.0.0.1:6379"`
- **AND** `cluster.redis.connectTimeout=3s`
- **AND** `cluster.redis.readTimeout=2s`
- **AND** `cluster.redis.writeTimeout=2s`
- **THEN** configuration service returns Redis configuration object
- **AND** timeout fields are all `time.Duration`

#### Scenario: Redis address missing
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** `cluster.redis.address` is empty
- **THEN** host startup fails
- **AND** error message contains missing field `cluster.redis.address`

#### Scenario: Redis timeout format illegal
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** `cluster.redis.readTimeout=2000`
- **THEN** host startup fails
- **AND** error message requires unit-bearing duration string

### Requirement: Cluster startup must complete Redis probe first
The system SHALL complete Redis coordination probe before HTTP service, scheduled tasks, plugin runtime, and business route startup. On probe failure, the system MUST refuse to start in cluster mode.

#### Scenario: Refuse startup when Redis unreachable
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis address is not connectable
- **THEN** host startup fails
- **AND** no HTTP business routes are registered
- **AND** leader election, cron, plugin runtime reconciler, or cache watcher not started

#### Scenario: Continue startup after Redis probe success
- **WHEN** configuration declares `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis ping succeeds
- **THEN** host continues initializing cluster, coordination, cron, and plugin runtime components
- **AND** health diagnostics show coordination backend as `redis`

### Requirement: SQLite dialect must prohibit cluster coordination
When database link is SQLite dialect, the system SHALL force `cluster.enabled=false` at memory layer. Even if configuration declares `cluster.coordination=redis`, the system MUST not connect to Redis or start cluster coordination.

#### Scenario: SQLite configured with Redis coordination
- **WHEN** `database.default.link` starts with `sqlite:`
- **AND** configuration declares `cluster.enabled=true`
- **AND** configuration declares `cluster.coordination=redis`
- **THEN** `IsClusterEnabled` returns `false`
- **AND** system outputs SQLite standalone mode warning
- **AND** system must not connect to Redis

### Requirement: Configuration template must show Redis cluster mode
The system SHALL provide Redis coordination configuration example in `manifest/config/config.template.yaml`, with clear comments explaining standalone mode does not need Redis, cluster mode must configure `cluster.coordination: redis`.

#### Scenario: Configuration template contains Redis coordination
- **WHEN** developer views `config.template.yaml`
- **THEN** file contains `cluster.coordination: redis` example
- **AND** file contains `cluster.redis.address`, `db`, `password`, `connectTimeout`, `readTimeout`, `writeTimeout` field descriptions
- **AND** comments explain `cluster.enabled=false` does not need Redis
