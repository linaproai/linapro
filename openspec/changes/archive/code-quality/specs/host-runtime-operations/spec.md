## ADDED Requirements

### Requirement: Host must provide an anonymous health probe endpoint

The host SHALL provide `GET /api/v1/health` under the public route group, accessible without login state. The endpoint MUST be exposed through standard API DTO and controller flow, return service self-check results, and include one lightweight database probe. When the database is unavailable or the probe times out, it MUST return HTTP `503` with a stable redacted unavailable reason. When healthy, it MUST return HTTP `200` with `{"status":"ok","mode":"<single|master|slave>"}`. Probe timeout SHALL be controlled by configuration key `health.timeout`, default `5s`, and parsed as `time.Duration`.

#### Scenario: Health probe returns 200 when database is healthy

- **WHEN** a caller anonymously accesses `GET /api/v1/health`
- **AND** the database is reachable and the probe completes within the timeout
- **THEN** the endpoint returns `200` and the response body includes `status="ok"` plus current deployment `mode`
- **AND** the endpoint does not require an `Authorization` header

#### Scenario: Health probe returns 503 when database is unavailable

- **WHEN** a caller anonymously accesses `GET /api/v1/health`
- **AND** the database probe does not return within `health.timeout`
- **THEN** the endpoint returns `503` and the response body includes `status="unavailable"` plus a stable redacted reason
- **AND** raw database, network, or schema errors MUST only be logged and MUST NOT be returned to anonymous callers

#### Scenario: Health probe is not mounted under protected routes

- **WHEN** the service starts and route registration completes
- **THEN** `/api/v1/health` MUST be registered under the public route group
- **AND** Auth and Permission middleware MUST NOT intercept that route

### Requirement: Host must reuse GoFrame HTTP graceful shutdown

The host SHALL use the built-in process signal handling and HTTP graceful shutdown behavior of GoFrame `Server.Run()` for `SIGTERM`, `SIGINT`, and similar shutdown signals. `internal/cmd/cmd_http.go` MUST NOT register `os/signal` again or reimplement an HTTP server shutdown loop. After GoFrame `Server.Run()` returns, the host MUST clean up owned runtime resources in this order: stop the cron scheduler, stop the cluster service, and close the database connection pool. Host-owned cleanup MUST be bounded by `shutdown.timeout`, default `30s`; `shutdown.timeout` MUST use a unit-bearing string configuration value parsed as `time.Duration`.

#### Scenario: SIGTERM reuses GoFrame HTTP shutdown

- **WHEN** the host process receives `SIGTERM`
- **THEN** HTTP server shutdown MUST be triggered by GoFrame `Server.Run()` built-in signal handling
- **AND** `cmd_http.go` MUST NOT also register an extra `signal.NotifyContext` or equivalent `os/signal` listener
- **AND** the Cron scheduler MUST stop accepting new triggers and wait for in-flight jobs after `Server.Run()` returns
- **AND** the cluster service MUST stop after the Cron scheduler shuts down
- **AND** the database connection pool MUST close after Cron shutdown
- **AND** host-owned runtime cleanup MUST complete within `shutdown.timeout`

#### Scenario: Owned runtime cleanup times out

- **WHEN** shutdown exceeds `shutdown.timeout`
- **THEN** the host MUST log a timeout warning and return an error
- **AND** the process MUST NOT hang forever

### Requirement: Upload file access endpoint must belong to the file module and use host unified authorization

The host SHALL declare `GET /api/v1/uploads/*` through file API DTOs and the file controller, and register it with the file controller under the protected route group. Unified Auth and Permission middleware MUST handle authentication and permission checks. Anonymous callers MUST NOT directly access uploaded files. The endpoint permission tag SHALL align with the file module menu/button permission. The implementation MUST query file metadata from the relative storage path in the URL and read file streams through the file service storage backend. It MUST NOT concatenate local upload directories or directly access local filesystem paths in `internal/cmd/cmd_http.go`.

#### Scenario: Unauthenticated access is rejected

- **WHEN** an anonymous caller requests `GET /api/v1/uploads/<path>`
- **THEN** the host MUST return a standard unauthenticated response, such as 401 or equivalent business code
- **AND** file content MUST NOT appear in the response body

#### Scenario: Authenticated caller without permission is rejected

- **WHEN** an authenticated caller without file read permission requests `GET /api/v1/uploads/<path>`
- **THEN** the host MUST return a standard forbidden response, such as 403 or equivalent business code

#### Scenario: Authenticated caller with permission gets the file

- **WHEN** an authenticated caller with file read permission requests `GET /api/v1/uploads/<path>`
- **AND** the file exists
- **THEN** the host returns `200` with file content
- **AND** file content is read through the file service and storage backend, not through a local-path handler in `cmd_http.go`

### Requirement: Host must remove empty audit placeholder packages

Host source MUST NOT keep zero-file placeholder directories such as `apps/lina-core/pkg/auditi18n/` and `apps/lina-core/pkg/audittype/`. Real audit-log capability MAY be introduced by a separate iteration, but the main codebase MUST NOT retain placeholders that imply audit capability already exists.

#### Scenario: Repository does not contain empty audit placeholder directories

- **WHEN** `apps/lina-core/pkg/` is inspected
- **THEN** `auditi18n` and `audittype` directories MUST NOT exist
- **OR** if they exist, they MUST contain at least one effective `.go` file

### Requirement: HTTP entry code must be split by responsibility

The host SHALL keep `apps/lina-core/internal/cmd/cmd_http.go` focused on HTTP command entry orchestration. HTTP runtime service construction, API route binding, frontend static resource serving, host OpenAPI binding, and post-start lifecycle hooks MUST be maintained in separately named source files in the same package, so one HTTP entry file does not carry multiple infrastructure implementation details.

#### Scenario: HTTP entry file remains lightweight orchestration

- **WHEN** a maintainer opens `apps/lina-core/internal/cmd/cmd_http.go`
- **THEN** the file SHOULD contain only `HttpInput`, `HttpOutput`, and `Main.Http` startup orchestration
- **AND** concrete route binding, runtime construction, static resource serving, and OpenAPI handlers MUST live in independent `cmd_http_*.go` files

### Requirement: Configuration service interface must compose categories

The top-level host configuration `Service` interface SHALL compose narrower responsibility-based interfaces through embedding. Configuration capabilities for auth, login, cluster, frontend, i18n, cron, host runtime, delivery metadata, plugin, upload, and runtime parameter synchronization MUST be maintained in clearly named category interfaces, avoiding direct accumulation of every method on the top-level `Service`.

#### Scenario: Service interface composes category interfaces

- **WHEN** a maintainer reviews `apps/lina-core/internal/service/config/config.go`
- **THEN** the `Service` interface MUST embed multiple category interfaces
- **AND** methods in each category interface MUST keep responsibility comments next to their declarations
- **AND** `serviceImpl` MUST continue implementing the complete `Service` contract

### Requirement: Scheduler default timezone must be configurable

The host SHALL replace the hard-coded default timezone in `cron_managed_jobs.go` with configuration key `scheduler.defaultTimezone`, defaulting to `UTC`. Source code MUST NOT keep hard-coded constants such as `defaultManagedJobTimezone = "Asia/Shanghai"`.

#### Scenario: Missing configuration uses UTC

- **WHEN** the configuration file does not declare `scheduler.defaultTimezone`
- **THEN** the host uses `UTC` as the default timezone when registering built-in jobs

#### Scenario: Custom configured timezone takes effect

- **WHEN** the configuration file sets `scheduler.defaultTimezone: "Asia/Shanghai"`
- **THEN** the host uses `Asia/Shanghai` when registering built-in jobs

### Requirement: Middleware service interface must split by responsibility

The host SHALL split the middleware `Service` interface into two narrower interfaces: `HTTPMiddleware` for methods and factories that can be directly installed into GoFrame route groups, and `RuntimeSupport` for non-middleware support methods such as `SessionStore()` and `PublishedRouteMiddlewares()`. The top-level `Service` MUST embed both interfaces for backward compatibility.

#### Scenario: Middleware service splits HTTP and runtime support

- **WHEN** a maintainer reviews `apps/lina-core/internal/service/middleware/`
- **THEN** the `Service` interface MUST embed `HTTPMiddleware` and `RuntimeSupport`
- **AND** `serviceImpl` MUST continue implementing the complete `Service` contract
