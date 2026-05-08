## 1. Dependencies and Scaffolding

- [x] 1.1 Add `github.com/gogf/gf/contrib/drivers/sqlite/v2` dependency to `apps/lina-core/go.mod`; run `go mod tidy` to confirm `go.sum` is synchronized
- [x] 1.2 Anonymous-import SQLite driver in `apps/lina-core/main.go` (or existing driver registration point), alongside existing MySQL driver registration
- [x] 1.3 Add `temp/sqlite/*.db` / `temp/sqlite/*.db-shm` / `temp/sqlite/*.db-wal` entries to repository root `.gitignore` to prevent SQLite data files from being committed (if `temp/` is already ignored, no changes needed; record investigation conclusion)
- [x] 1.4 Add SQLite link example (`sqlite::@file(./temp/sqlite/linapro.db)`) in comments of `apps/lina-core/manifest/config/config.template.yaml` and `config.yaml` `database.default.link`; default value remains MySQL link for backward compatibility
- [x] 1.5 Create `apps/lina-core/pkg/dialect/` directory skeleton: `dialect.go` (interface and `From()` dispatch function), `dialect_mysql.go`, `dialect_sqlite.go`, `dialect_sqlite_translate.go` (SQLite DDL translator implementation), with main file package comments and per-file file-level comments

## 2. Dialect Interface and MySQL Implementation

- [x] 2.1 Define public stable `Dialect` interface in `dialect.go`: `Name() string` / `TranslateDDL(ctx, sourceName, ddl) (string, error)` / `PrepareDatabase(ctx, link, rebuild) error` / `SupportsCluster() bool` / `OnStartup(ctx, runtime) error`, with interface comments for each method; public signatures must not reference host `internal` concrete service types
- [x] 2.2 Implement `From(link string) (Dialect, error)` in `dialect.go`: parse link protocol prefix (`mysql:` -> MySQL, `sqlite:` -> SQLite); unrecognized prefix returns clear `bizerr` error with list of supported prefixes
- [x] 2.3 Implement MySQL dialect: `TranslateDDL` returns input directly, `PrepareDatabase` migrates MySQL-specific link parsing, `DROP DATABASE` and `CREATE DATABASE` logic from existing `prepareInitDatabase` (including migrating existing `databaseNameFromMySQLLink` / `serverLinkFromMySQLLink` / `quoteMySQLIdentifier` utility functions), `SupportsCluster` returns `true`, `OnStartup` is a no-op
- [x] 2.4 Define `RuntimeConfig` and other narrow interfaces in `dialect.go`, adapted by host config service to implement `OverrideClusterEnabledForDialect(value bool)`; verify `pkg/dialect` can be imported by code outside `internal` without reverse dependency on host private packages
- [x] 2.5 Write MySQL dialect unit tests: cover `TranslateDDL(ctx, sourceName, ddl)` input-output byte equality, `sourceName` does not affect no-op, `PrepareDatabase` link parsing, `SupportsCluster` return value

## 3. SQLite Dialect and DDL Translator

- [x] 3.1 Implement `PrepareDatabase` in SQLite dialect: parse SQLite link for file path, automatically `mkdir -p` parent directory; `rebuild=true` deletes main db file and possible `.db-shm` / `.db-wal` auxiliary files; directory creation failure returns clear error including path
- [x] 3.2 Implement `SupportsCluster() = false`, `Name() = "sqlite"` in SQLite dialect
- [x] 3.3 Implement `OnStartup` in SQLite dialect: call `runtime.OverrideClusterEnabledForDialect(false)` (step 5.1 new method adapted by host config service), and output startup prompt log covering: current SQLite mode, cluster forced lock reason, do not use in production
- [x] 3.4 Implement SQLite DDL translator in `dialect_sqlite_translate.go`, covering these conversion rules:
  - Backtick identifier removal (with string-internal backtick recognition to avoid false positives)
  - Current SQL real-world `INT/BIGINT [UNSIGNED] PRIMARY KEY AUTO_INCREMENT`, `AUTO_INCREMENT PRIMARY KEY`, `NOT NULL AUTO_INCREMENT` + table-level `PRIMARY KEY(id)` and other syntax -> `INTEGER PRIMARY KEY AUTOINCREMENT`
  - `TINYINT / SMALLINT [UNSIGNED]` -> `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT / MEDIUMTEXT` -> `TEXT`
  - `DECIMAL(M,N)` -> `NUMERIC`
  - `INSERT IGNORE INTO` -> `INSERT OR IGNORE INTO`
  - Remove `ENGINE=InnoDB` / `ENGINE=MEMORY` / `DEFAULT CHARSET=utf8mb4` / `COLLATE=utf8mb4_general_ci` clauses
  - Remove column-level `COMMENT '...'` and table-level `COMMENT='...'` (both `COMMENT=` and `COMMENT '...'` syntax variants)
  - Remove only `ON UPDATE CURRENT_TIMESTAMP` part, preserve `DEFAULT CURRENT_TIMESTAMP`
  - Inline `KEY idx_xxx (col)` / `INDEX idx_xxx (col)` / `UNIQUE KEY uk_xxx (col)` / `UNIQUE INDEX uk_xxx (col)` / expression indexes currently used in SQL extracted as independent `CREATE INDEX` / `CREATE UNIQUE INDEX` statements after table creation
  - Delete entire `CREATE DATABASE IF NOT EXISTS xxx;` and `USE xxx;` statements
- [x] 3.5 Translator returns clear error when encountering uncovered MySQL features (`FULLTEXT INDEX` / `SPATIAL INDEX` / `GENERATED ALWAYS AS` / partition clauses / `ON DUPLICATE KEY UPDATE` / database-specific functions); error message includes `sourceName`, line number hint, and uncovered keyword
- [x] 3.6 Write SQLite DDL translator unit tests: automatically scan current host install SQL, plugin install SQL, host mock SQL, plugin mock SQL, and plugin uninstall SQL as fixtures; each fixture asserts both "translation succeeds without error" and "translation result executes successfully on temporary SQLite database"
- [x] 3.7 Write SQLite DDL translator negative tests: construct DDL input containing `FULLTEXT` / `GENERATED` / `ON DUPLICATE KEY UPDATE`, assert translator returns error with `sourceName`, line number hint, and uncovered keyword
- [x] 3.8 Write current real SQL form coverage tests: cover `AUTO_INCREMENT PRIMARY KEY`, table-level `PRIMARY KEY(id)`, `UNIQUE INDEX`, expression unique index `NULLIF(code, '')` and other existing syntax, ensuring no reliance on normalizing SQL files to bypass translation capability gaps

## 4. cmd init / mock Integration with Dialect Layer

- [x] 4.1 Refactor `apps/lina-core/internal/cmd/cmd_init_database.go`: `prepareInitDatabase` changed to `dialect.From(link).PrepareDatabase(ctx, link, rebuild)`; migrate all MySQL-specific logic into MySQL dialect implementation
- [x] 4.2 Modify shared SQL execution path in `apps/lina-core/internal/cmd/`: before calling `splitSQLStatements`, first call `dialect.From(link).TranslateDDL(ctx, sourceName, content)` to translate SQL file content; `sourceName` uses source file path; translation failure uses existing fail-fast path preserving failed filename, line number hint, and uncovered keyword
- [x] 4.3 Modify `apps/lina-core/internal/cmd/cmd_mock.go`: same translation integration approach as 4.2, covering `manifest/sql/mock-data/` assets; clarify `make mock` does not call `PrepareDatabase`, must depend on `make init` having completed initialization
- [x] 4.4 Write cmd init / mock integration tests: execute end-to-end init + mock flow with both MySQL and SQLite links, assert all current host and plugin SQL files execute successfully
- [x] 4.5 Verify `init --rebuild=true` correctly deletes database file and rebuilds in SQLite mode
- [x] 4.6 Write `make mock` dependency on initialization failure test: running mock with SQLite configuration without prior init or when database tables do not exist must fail fast, not create or rebuild database
- [x] 4.7 Write SQLite forward integration test: use `database.default.link` configured SQLite temporary file to execute embedded host `init --rebuild=true` -> `mock`, assert current host initialization SQL and mock SQL execute successfully and write `admin` / `user001`

## 5. Cluster Lock + Startup Warning

- [x] 5.1 Add `OverrideClusterEnabledForDialect(value bool)` method in `apps/lina-core/internal/service/config/config_cluster.go`: after dialect layer calls it, locks in-memory `cluster.enabled` to specified value; all subsequent `IsClusterEnabled` calls stably return that value, no longer reading configuration file
- [x] 5.2 Find the position before cluster service initialization in `apps/lina-core/internal/cmd/cmd_http_runtime.go` (or equivalent startup bootstrap entry), call `dialect.From(link).OnStartup(ctx, runtime)`; `runtime` adapted by host config service; this call must be earlier than `clusterSvc` initialization and election loop start
- [x] 5.3 Write unit test: simulate SQLite link startup scenario, assert `IsClusterEnabled` returns `false` and log output contains SQLite mode, cluster lock, and do-not-use-in-production necessary content
- [x] 5.4 Write unit test: simulate MySQL link + `cluster.enabled=true` startup scenario, assert `IsClusterEnabled` returns `true` and no SQLite-related warning logs

## 6. Plugin install / uninstall Pipeline Integration with Dialect Layer

- [x] 6.1 Find the install path executing plugin `manifest/sql/` in `apps/lina-core/internal/service/plugin/`, insert `dialect.From(link).TranslateDDL(ctx, sourceName, content)` translation before execution; `sourceName` includes plugin identifier, asset type, and failed filename; translation failure returns clear error with line number hint and uncovered keyword
- [x] 6.2 Apply same integration to uninstall path in same module, covering `manifest/sql/uninstall/` assets
- [x] 6.3 Apply same integration to mock-data loading path (consistent with step 4.3), covering plugin `manifest/sql/mock-data/` assets
- [x] 6.4 Write plugin lifecycle integration test: in SQLite mode, fully execute "install -> enable -> load mock -> disable -> uninstall" flow, assert each step plugin SQL assets successfully translated and executed
- [x] 6.5 Verify existing source-code plugins (`monitor-loginlog` / `monitor-operlog` / `monitor-server` / `org-center` / `content-notice` / `plugin-demo-source`) and dynamic plugins (`plugin-demo-dynamic`) install, enable, uninstall behavior is normal in SQLite mode

## 7. kvcache Backend Rename and Atomic Increment Rewrite

- [x] 7.1 Rename `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` directory to `sqltable/`; update all import paths and sub-file package declarations
- [x] 7.2 Rename constant `BackendMySQLMemory` to `BackendSQLTable` in `apps/lina-core/internal/service/kvcache/kvcache_backend.go`, string value from `"mysql-memory"` to `"sql-table"`; update all references
- [x] 7.3 Rewrite `sqltable/sqltable_ops.go` (originally `mysql_memory_ops.go`) `Incr` implementation: remove `gdb.Raw("LAST_INSERT_ID(value_int + ...)")` and `Raw("SELECT LAST_INSERT_ID()")` calls, replace with dialect-neutral CAS retry flow: read current integer snapshot, if missing use `INSERT IGNORE value_int=0` initialization then re-read, write new value through parameterized `UPDATE` with `value_int=<snapshot>` condition; on snapshot competition and database-suggested retryable lock conflicts, execute bounded backoff retry; ensure first `incr(delta)` returns `delta`, and do not modify original MySQL `sys_kv_cache ENGINE=MEMORY` delivery SQL
- [x] 7.4 Check other methods in `sqltable/` for remaining MySQL-specific SQL, if any similarly rewrite to dialect-neutral form
- [x] 7.5 Update `sqltable/` existing unit tests: tests originally using MySQL containers changed to parameterized execution (running both MySQL and SQLite), or add corresponding SQLite test suite; assert `Incr` behavior is consistent across both engines; cover concurrent first increment of missing key, first `delta` as initial value, incrementing non-integer value unchanged, concurrent increment without lost updates
- [x] 7.6 Update `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache.go`, `hostfn_service_cache_test.go`, `hostfn_service_lock_test.go` for possible `BackendMySQLMemory` literal references and `ENGINE=MEMORY` text in test fixtures

## 8. Startup Directory Creation and SQLite Path Governance

- [x] 8.1 Verify SQLite dialect `PrepareDatabase` auto-creates parent directory when it does not exist (step 3.1 implementation), write unit test covering path-not-exists scenario
- [x] 8.2 Verify SQLite link containing `~` (HOME expansion) or absolute path parsing behavior, clarify documentation convention (only relative working directory paths and absolute paths supported, `~` not expanded), write test coverage
- [x] 8.3 Verify that `~/.linapro/data/linapro.db` path written by user gets a clear error (does not parse `~`), error message hints user to use absolute path or working directory relative path

## 9. E2E Test Suite Parameterized Execution

- [x] 9.1 Add SQLite mode execution channel to `hack/tests/`: test fixtures must write `database.default.link=sqlite::@file(./temp/sqlite/linapro.db)` into test configuration file before starting services and auto-prepare `temp/sqlite/linapro.db`; backend runtime must not read command-line arguments or environment variables as database dialect source
- [x] 9.2 Add test case `hack/tests/e2e/dialect/TC0164-sqlite-mode-startup.ts`: after starting SQLite mode, assert terminal log contains SQLite startup prompt, `/api/v1/cluster/status` (or equivalent endpoint) returns single-node status, login admin/admin123 succeeds
- [x] 9.3 Add test case `hack/tests/e2e/dialect/TC0165-sqlite-mode-business-zero-impact.ts`: in SQLite mode, fully run "user list -> create user -> modify user -> delete user" and "plugin list -> enable plugin -> disable plugin" two core business scenarios, assert behavior consistent with MySQL mode
- [x] 9.4 Add test case `hack/tests/e2e/dialect/TC0166-sqlite-mode-rebuild-and-reseed.ts`: execute `make init confirm=init rebuild=true` + `make mock confirm=mock` to verify SQLite mode rebuild database + load mock data complete flow
- [x] 9.5 Retain SQLite backend smoke channel in CI configuration, covering SQLite startup warning, single-node health, and admin login; full SQLite E2E channel remains as manual verification entry, avoiding full browser and plugin lifecycle regression on every main CI run

## 10. Documentation and README Sync

- [x] 10.1 Update `apps/lina-core/README.md` and `README.zh_CN.md`: add SQLite link example and "demo / testing purpose, not supported in production, not supported in cluster" explanation in "Database Configuration" section
- [x] 10.2 Update repository root `README.md` and `README.zh_CN.md`: add "optional SQLite mode without MySQL" explanation in "Quick Start / Environment Requirements" section
- [x] 10.3 Update `Makefile` top comments (without modifying commands): explain `make init` / `make mock` automatically dispatch to corresponding dialect based on `database.default.link` protocol prefix
- [x] 10.4 Check i18n impact: this change does not involve frontend UI text, menus, buttons, forms, tables, apidoc documentation or other translation resource adjustments; explicitly record "i18n resources do not need to be added, modified, or deleted" judgment

## 11. Self-Review and Acceptance

- [x] 11.1 Run `gofmt`, `go vet ./...` passes; run `go test ./...` backend unit tests all green
- [x] 11.2 Under MySQL link, fully execute `make init confirm=init` -> `make mock confirm=mock` -> `make dev` -> browser access management workbench -> login admin/admin123, confirm zero regression
- [x] 11.3 Under SQLite link, fully execute `make init confirm=init` -> `make mock confirm=mock` -> `make dev` -> browser access management workbench -> login admin/admin123, confirm terminal has clear SQLite standalone mode prompt and business functions are usable
- [x] 11.4 Under SQLite link + `cluster.enabled=true` configuration, confirm `IsClusterEnabled` actually returns `false` and warning log is prominent
- [x] 11.5 Under SQLite link, fully run plugin management "install -> enable -> uninstall" flow (covering at least `monitor-loginlog` source plugin and `plugin-demo-dynamic` dynamic plugin)
- [x] 11.6 Call `/lina-review` skill for final code and specification review
- [x] 11.7 Run `golangci-lint run` passes

## Feedback

- [x] **FB-1**: Must not modify original MySQL delivery SQL `013-dynamic-plugin-host-service-extension.sql` `sys_kv_cache` table structure or `ENGINE=MEMORY` engine type for SQLite support; SQLite adaptation must converge in `pkg/dialect` execution-time translation and `kvcache incr` dialect-neutral CAS implementation
- [x] **FB-2**: Frontend internal package `unbuild --stub` generated `dist` stubs must not contain build-machine local absolute paths; source distribution should be able to resolve source entry from current repository location
- [x] **FB-3**: Role management add/edit drawer menu permission selector selected count lacks left spacing, and modifying permission checkboxes then close icon, cancel button, and mask close have no feedback and cannot close
- [x] **FB-4**: `kvcache incr` database retryable write conflict detection must not hardcode MySQL/SQLite error text in `sqltable` business implementation; should be delegated to `pkg/dialect` driver error classification capability
- [x] **FB-5**: `host:state` `state.get` and `state.delete` must not bypass generated `dao.SysPluginState` and directly use `g.DB().Model` to access `sys_plugin_state`
- [x] **FB-6**: Check plugin generic resources and `plugin-demo-source` fixed table database access patterns; dynamic manifest table access may retain dynamic models, fixed source plugin table access should reuse plugin's own DAO and generated column names
- [x] **FB-7**: MySQL dialect database initialization must not hardcode `linapro` database name in implementation logic; should fully use database name configured in `database.default.link`
- [x] **FB-8**: MySQL dialect database initialization must not parse GoFrame database link on its own; should reuse `gdb`-parsed `ConfigNode` to obtain target database name and construct server-level connection configuration
- [x] **FB-9**: `pkg/dialect` public package must not directly expose `MySQLDialect` / `SQLiteDialect` concrete implementations; MySQL and SQLite dialect implementations should be split into `pkg/dialect/internal/mysql` and `pkg/dialect/internal/sqlite`, providing stable boundary through public `Dialect` interface, factory functions, and necessary public facade capabilities
- [x] **FB-10**: SQL keyword and identifier matching in SQLite DDL translator must avoid case-sensitivity gaps, especially when table-level primary key column names and column definitions have inconsistent casing
- [x] **FB-11**: Reduce main CI SQLite verification time; CI only runs SQLite backend smoke, full SQLite E2E remains as manual verification entry
- [x] **FB-12**: SQLite backend smoke script only serves testing and CI verification, should be placed in `hack/tests/scripts/` test scripts directory, not general `hack/scripts/` directory
- [x] **FB-13**: English locale plugin management list uninstalled dynamic plugin names and descriptions still display Chinese; should display plugin manifest localized text according to current language
- [x] **FB-14**: Role management add/edit drawer menu permission selected count and left mode selection lack stable spacing; checking permissions then close icon, cancel button, and mask close should properly trigger unsaved confirmation and close
- [x] **FB-15**: Management workbench i18n first startup language should auto-detect browser language; Chinese browser defaults to Chinese, other browsers default to English; must not overwrite user's saved language preference
- [x] **FB-16**: Role management add/edit drawer menu permission tree should maintain directory-contains-menu, menu-contains-button display structure; dynamically added permissions should be grouped under a dynamic permissions group after system menus, not displayed at the top root level
- [x] **FB-17**: Remove LinaPro installation script that only wraps Git clone, along with related installer documentation, test entry, and OpenSpec capability description
- [x] **FB-18**: `make build` and `make image` need to support cross-platform builds; `make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1` must build and push multi-architecture Docker images
- [x] **FB-19**: Normal user authorized only for org management and content management directory permissions should not get permission error when accessing corresponding plugin pages due to page reusing dictionary options interface lacking `system:dict:query`
- [x] **FB-20**: `hack/config.yaml` `build.os` / `build.arch` / `build.platform` should converge to `build.platforms` array; array items support `<goos>/<goarch>` and `auto`, command-line override uses `platforms=linux/amd64,linux/arm64`
- [x] **FB-21**: Add GitHub Actions nightly build workflow, automatically build and publish `linux/amd64` and `linux/arm64` multi-architecture Docker images to `ghcr.io` every night
- [x] **FB-22**: `make build` and `make image` support `config=<path>` build configuration parameter for specifying config file path instead of default `hack/config.yaml`
- [x] **FB-23**: Add GitHub Actions tag release image workflow, only build and publish corresponding tag and `latest` `linux/amd64`, `linux/arm64` multi-architecture Docker images when new tag is created
- [x] **FB-24**: Add GitHub Actions nightly test workflow, automatically run full Go/frontend unit tests and full Playwright E2E suite every day at `Asia/Shanghai 00:00`, continuously verifying project health
- [x] **FB-25**: Remove Nightly Test Go unit test directory hardcoding, change to auto-discover workspace modules from `go.work` and execute full `go test ./...`
- [x] **FB-26**: Main CI SQLite smoke and SQLite E2E support code should not continue asserting deleted "Switch database.default.link back to a MySQL link" startup log
- [x] **FB-27**: Locate and fix `Go unit tests / Go unit tests` job failure in main CI
- [x] **FB-28**: SQLite runtime enabling `monitor-server` plugin should not error on first service monitoring data collection due to implicit `Save()` lacking conflict columns
- [x] **FB-29**: SQLite mode operation log management list interface should not get `Database Operation Error` due to `EXISTS ((SELECT ...))` subquery syntax
- [x] **FB-30**: Go unit test entry and GitHub Actions backend unit test flow should uniformly use `go test -race` to detect potential race conditions
- [x] **FB-31**: SQLite mode host system info and `monitor-server` service monitoring must not continue using MySQL-specific `SELECT VERSION()` for database version; page should display non-empty SQLite version info
- [x] **FB-32**: Locate and fix GitHub Actions run `25544849415` `Main CI / Go unit tests` failure
