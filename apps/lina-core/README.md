# LinaPro Core Host

`apps/lina-core` is the GoFrame-based host service of `LinaPro`. It exposes reusable backend modules, shared governance capabilities, plugin runtime support, and the public/private APIs consumed by the default workspace.

## Responsibilities

- Provide the host-side RESTful APIs for system management, plugin governance, and shared platform capabilities.
- Own database migrations, seed data, runtime configuration, and generated `DAO` / `DO` / `Entity` artifacts.
- Load source plugins and dynamic `Wasm` plugins, and coordinate their lifecycle with host governance data.
- Run host-level cron tasks and the persisted scheduled-job subsystem used by the management workspace.

## Scheduled Job Management

The scheduled-job subsystem is part of the host service because it is a core governance capability.

- `internal/service/jobhandler`: handler registry, host handler bootstrap, plugin-linked handler registration, and schema validation.
- `internal/service/jobmgmt`: job groups, persisted jobs, execution logs, cron preview, retention cleanup, and lifecycle rules for built-in jobs.
- `internal/service/jobmgmt/scheduler`: persisted job loading, `gcron` registration, concurrency guards, execution dispatch, and cancellation.
- `internal/service/jobmgmt/shellexec`: guarded `Shell` execution, timeout handling, output truncation, and manual cancellation support.
- `internal/service/cron`: startup entrypoint that registers host cron jobs and loads persisted scheduled jobs after the host boots.

## Key Directories

```text
api/                API DTOs and route contracts
internal/cmd/       Service startup and route wiring
internal/controller/ HTTP controllers
internal/service/   Business services and cron orchestration
manifest/config/    Host runtime config
manifest/sql/       Host SQL migrations and seed data
manifest/sql/mock-data/ Optional host mock/demo SQL assets
pkg/                Stable shared packages for host and plugins
```

## Common Commands

```bash
go run main.go
make build
make dao
make ctrl
make init confirm=init
```

For release builds, use `make build platforms=linux/arm64` to cross-compile one target, or run multi-platform image publishing from the repository root:

```bash
make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

## Database Configuration

The host reads the active database dialect only from `database.default.link` in the runtime config. `PostgreSQL 14+` is the default production database. Prepare `PostgreSQL` before running `make init` or `make dev`; those commands do not start or manage the database.

For local development, start a matching container:

```bash
docker run --name linapro-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=linapro \
  -p 5432:5432 \
  --health-cmd pg_isready \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 5 \
  -d postgres:14-alpine
```

The default runtime link is:

```yaml
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
```

`make init` is an operations bootstrap command that uses the configured database account. That account must be able to connect to the system database, create and drop the target database, terminate target-database connections, create tables and indexes, write comments, and insert seed data. Permission failures are surfaced directly; the runtime does not provide a lower-privilege initialization fallback.

For external hosted `PostgreSQL`, such as `RDS` or `Aliyun PolarDB`, point `database.default.link` at the provider endpoint and run initialization with an account that has the permissions above.

For a single-node development demo, set the link to `SQLite`:

```yaml
database:
  default:
    link: "sqlite::@file(./temp/sqlite/linapro.db)"
```

`SQLite` mode is single-node only, automatically forces `cluster.enabled=false`, and is not supported for production deployments.

## Cluster Coordination

Single-node deployments keep the lightweight local mode: they do not require `Redis`, do not connect to `Redis`, and continue to use PostgreSQL plus process-local cache coordination where appropriate.

Clustered deployments must configure the Redis coordination backend explicitly:

```yaml
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: ""
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

The current backend only supports `redis`. The scalar `cluster.coordination` setting is intentionally stable so future backends can be added without exposing per-store coordination switches. Redis integration tests are opt-in; set `LINA_TEST_REDIS_ADDR`, for example `LINA_TEST_REDIS_ADDR=127.0.0.1:6379`, before running the Redis-specific Go test cases.

## Source Plugin Upgrade

Source plugins now follow an explicit development-time upgrade flow instead of
silently switching versions during startup.

- Host startup scans source plugins first, but if an installed source plugin still has a higher discovered `plugin.yaml` version than the effective `sys_plugin.version`, startup fails fast until the version drift has been resolved.
- Use the supported plugin workspace update flow to refresh source-plugin code before startup.
- Process every installed source plugin with a newer discovered version before starting the host.
- Dynamic plugins keep their existing runtime-managed `upload + install/reconcile` upgrade model.

## Related Entry Points

- `manifest/sql/011-scheduled-job-management.sql`: scheduled-job schema, seed data, menus, and dictionaries.
- `internal/cmd/cmd_http.go`: host wiring for job, job-group, job-log, and job-handler APIs.
- `internal/service/cron/cron.go`: host cron startup entrypoint.
