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

## Related Entry Points

- `manifest/sql/014-scheduled-job-management.sql`: scheduled-job schema, seed data, menus, and dictionaries.
- `internal/cmd/cmd_http.go`: host wiring for job, job-group, job-log, and job-handler APIs.
- `internal/service/cron/cron.go`: host cron startup entrypoint.
