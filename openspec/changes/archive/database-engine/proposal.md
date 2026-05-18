## Why

LinaPro currently supports only MySQL as its database engine. The configuration field `database.default.link` requires a `mysql:user@tcp(...)` connection string, all host and plugin install/mock/uninstall SQL assets are deeply coupled to MySQL dialect, and the `kvcache` module uses a MySQL-specific `LAST_INSERT_ID(value_int + delta)` atomic increment trick. This creates an unnecessary barrier for "demo / personal testing / zero-dependency local trial" scenarios -- users who want to quickly try the framework must first install and configure MySQL, or even `make init` will not work. This change introduces SQLite as an optional database engine for demo and testing scenarios, allowing users to run the entire framework with zero dependencies by changing a single line in `config.yaml`.

## What Changes

### New Dialect Abstraction Layer

- **Add** `apps/lina-core/pkg/dialect/` public stable package, defining a `Dialect` interface (`Name()` / `TranslateDDL(ctx, sourceName, ddl)` / `PrepareDatabase()` / `SupportsCluster()` / `OnStartup()`) as the unified boundary for database engine differences; the package exposes a stable narrow interface and must not bind host `internal` concrete service types in its public signatures; MySQL / SQLite concrete implementations are consolidated under `pkg/dialect/internal/mysql` and `pkg/dialect/internal/sqlite`
- **Add** MySQL internal dialect implementation: `TranslateDDL` is a no-op, `PrepareDatabase` reuses existing `DROP DATABASE` / `CREATE DATABASE` behavior, `SupportsCluster` returns `true`
- **Add** SQLite internal dialect implementation: `TranslateDDL` invokes the SQLite DDL translator, `PrepareDatabase` deletes the database file when `rebuild=true`, `SupportsCluster` returns `false`, `OnStartup` forces `cluster.enabled=false` and outputs startup prompt logs
- **Add** SQLite DDL translator covering the following MySQL-to-SQLite dialect mappings:
  - Backtick identifier removal
  - All `INT/BIGINT [UNSIGNED] AUTO_INCREMENT PRIMARY KEY` / `PRIMARY KEY AUTO_INCREMENT` / table-level `PRIMARY KEY(id)` combinations currently present in SQL files converted to SQLite-executable `INTEGER PRIMARY KEY AUTOINCREMENT` semantics
  - `TINYINT / SMALLINT [UNSIGNED]` to `INTEGER`
  - `VARCHAR(N) / CHAR(N) / LONGTEXT` to `TEXT`
  - `DECIMAL(M,N)` to `NUMERIC`
  - `INSERT IGNORE INTO` to `INSERT OR IGNORE INTO`
  - Removal of `ENGINE=` / `DEFAULT CHARSET=` / `COLLATE=` / column-level and table-level `COMMENT '...'`
  - Removal of only the `ON UPDATE` part from `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` (DAO layer already maintains `updated_at` automatically)
  - Inline `KEY` / `INDEX` / `UNIQUE KEY` / `UNIQUE INDEX` (including expression indexes currently present in SQL) extracted as standalone `CREATE INDEX` statements after table creation
  - `CONCAT(a, b, ...)` currently present in mock SQL translated to SQLite string concatenation `a || b || ...`
  - `CREATE DATABASE` / `USE database` statements deleted entirely

### Modified Host Bootstrap Commands

- **Modify** `apps/lina-core/internal/cmd/cmd_init_database.go`: migrate MySQL-specific link parsing and `DROP/CREATE DATABASE` logic into the MySQL dialect's `PrepareDatabase`; `prepareInitDatabase` dispatches to the corresponding dialect based on the `link` protocol prefix
- **Modify** `cmd_init.go` / `cmd_mock.go`: before executing SQL files, call the current dialect's `TranslateDDL(ctx, sourceName, ddl)` to convert single MySQL-source DDL into dialect-executable statements; `sourceName` uses the source file path or embedded asset path for error message file location
- **Clarify** `make mock` must depend on an already-initialized database and is not responsible for creating, rebuilding, or preparing the database; when `make init` has not been executed, it should fail fast and return a database error
- **Modify** SQLite mode: `rebuild=true` deletes the database file; on startup, if the database file parent directory does not exist, automatically `mkdir -p`

### Modified Cluster Mode Switch

- **Modify** `cluster-deployment-mode` specification: under SQLite links, the `cluster.enabled` configuration value is forced to `false` at the in-memory layer regardless of what the user writes in `config.yaml`; during startup, a clear prompt is output stating "currently in SQLite mode, only single-node deployment is supported, all features run in standalone mode, do not use in production"

### Modified Plugin Manifest Lifecycle

- **Modify** plugin install, uninstall, and mock data loading pipelines: before executing SQL resources under `manifest/sql/`, `manifest/sql/uninstall/`, and `manifest/sql/mock-data/`, uniformly pass through the current dialect's `TranslateDDL`; plugin source code requires no changes, single MySQL-dialect source SQL files execute correctly in SQLite mode

### Renamed kvcache Backend and Removed MySQL-Specific Syntax

- **BREAKING (internal)** `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` renamed to `sqltable/`; constant `BackendMySQLMemory` renamed to `BackendSQLTable`, string value changed from `"mysql-memory"` to `"sql-table"`
- **Modify** `Incr` operation: replace `LAST_INSERT_ID(value_int + delta)` + `SELECT LAST_INSERT_ID()` MySQL-specific pattern with dialect-neutral CAS retry: read current integer snapshot, if missing use `INSERT IGNORE value_int=0` for idempotent initialization, then execute parameterized `UPDATE` with `value_int=<snapshot>` condition to write new value; on `affected=0` due to competing writes, retry with bounded backoff, ensuring successful calls yield linear increments
- **Modify** `plugin-cache-service` specification: default MySQL delivery SQL retains original `sys_kv_cache ENGINE=MEMORY` table structure and engine type; SQLite mode only removes engine clauses via DDL translator to produce a plain SQLite table. Both modes rely on application-layer TTL and scheduled cleanup, and continue treating the cache as lossy

### Modified Configuration Defaults and Dependencies

- **Modify** `apps/lina-core/manifest/config/config.template.yaml` and `config.yaml` to add SQLite link examples in comments (default value remains MySQL link for zero-change experience for MySQL users)
- **Add** `go.mod` dependency: `github.com/gogf/gf/contrib/drivers/sqlite/v2`
- **Confirm** SQLite default path resides under `temp/` directory which is already ignored by the repository root `.gitignore`, no additional SQLite-specific ignore entries needed

### New E2E and Unit Tests

- **Add** SQLite DDL translator unit tests that automatically scan host install SQL, plugin install SQL, host mock SQL, plugin mock SQL, and plugin uninstall SQL as fixtures, asserting translation results execute successfully on SQLite
- **Add** SQLite mode E2E test cases: through test fixtures that write test configuration files before startup to switch `database.default.link`, without introducing command-line arguments or environment variables as runtime dialect sources, verify business modules behave identically under SQLite engine; main CI runs lightweight backend SQLite smoke only, full SQLite E2E remains as manual verification entry
- **Add** unit tests for startup cluster locking and startup prompt log output

## Capabilities

### New Capabilities

- `database-dialect-abstraction`: defines the database dialect abstraction layer, `Dialect` interface, and link-prefix-based dialect dispatch mechanism as the single convergence point for MySQL, SQLite, and other database engine differences; defines the SQLite DDL translator's coverage scope and executable result guarantees for MySQL-dialect DDL

### Modified Capabilities

- `database-bootstrap-commands`: adds dialect-dispatched init/rebuild semantics; adds mandatory dialect translation before SQL resource execution; adds SQLite database file path and parent directory auto-creation semantics
- `cluster-deployment-mode`: adds SQLite link `cluster.enabled` forced to `false` requirement; adds startup prompt log visibility requirement
- `plugin-cache-service`: clarifies MySQL delivery SQL `sys_kv_cache` retains existing `ENGINE=MEMORY` table structure and engine type; SQLite mode only removes engine clauses at execution time via DDL translator, and continues treating cache as lossy
- `plugin-manifest-lifecycle`: adds requirement that plugin SQL resources (`manifest/sql/` / `uninstall/` / `mock-data/`) must pass through current dialect translation before execution

## Impact

### Affected Code

- `apps/lina-core/pkg/dialect/` (new)
- `apps/lina-core/internal/cmd/cmd_init_database.go`, `cmd_init.go`, `cmd_mock.go` (refactored)
- `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` -> `sqltable/` (renamed + `Incr` rewritten)
- `apps/lina-core/internal/service/kvcache/kvcache_backend.go` (constant name adjustment)
- `apps/lina-core/internal/service/plugin/` (plugin install/uninstall pipeline integrates dialect translation)
- `apps/lina-core/internal/cmd/cmd_http_runtime.go` or equivalent startup bootstrap entry (SQLite startup cluster lock + startup prompt log)
- `apps/lina-core/manifest/config/config.template.yaml`, `config.yaml` (comments add SQLite link example)

### Affected Tests

- New dialect translator unit tests (covering current host/plugin install, mock, uninstall SQL assets)
- New startup cluster locking and log output unit tests
- Reuse existing E2E suite, add SQLite mode parameterized execution channel; main CI uses backend SQLite smoke covering startup prompt, single-node health, and admin login flow

### Affected Dependencies

- `apps/lina-core/go.mod` adds `github.com/gogf/gf/contrib/drivers/sqlite/v2`
- `.gitignore` does not need new SQLite-specific entries; repository root already has `temp/` ignore rule covering default SQLite database file path

### Unaffected

- Business modules (`controller` / `service` / `model` / `dao`): zero code changes, zero awareness of database engine differences (constraint #3)
- Existing MySQL users: default configuration remains MySQL, behavior fully backward compatible
- Plugin source code: single MySQL-dialect SQL files continue working, SQLite handled transparently by dialect layer
- `gf gen dao` / `gf gen ctrl` workflows: unchanged
- `apidoc` / i18n resources: unaffected by dialect switch
