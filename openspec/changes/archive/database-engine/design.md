## Context

LinaPro's database design is currently built entirely around MySQL, as evidenced by:

- **Link configuration**: `database.default.link` defaults to `mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true`, and `cmd_init_database.go` contains three utility functions (`databaseNameFromMySQLLink` / `serverLinkFromMySQLLink` / `quoteMySQLIdentifier`) that hardcode the assumption of MySQL protocol
- **DDL/DML dialect**: all host and plugin install, mock, and uninstall SQL assets use MySQL dialect, including `ENGINE=InnoDB` / `ENGINE=MEMORY` / `AUTO_INCREMENT` / `BIGINT UNSIGNED` / `TINYINT` / `LONGTEXT` / backtick identifiers / column and table-level `COMMENT '...'` / `DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci` / `INSERT IGNORE` / `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` / inline `KEY` and `INDEX` / `CREATE DATABASE IF NOT EXISTS` / `CONCAT(...)` and other features that SQLite does not support or has different syntax for
- **MEMORY engine semantics**: `sys_locker` (010), `sys_online_session` (006), and `sys_kv_cache` (013) use MySQL `MEMORY` engine, leveraging its "clear on restart" semantics to simplify distributed lock, online session, and volatile KV cache cleanup logic; this change must not modify original MySQL delivery SQL table structures or engine types for SQLite
- **Application-layer MySQL-specific code**: `apps/lina-core/internal/service/kvcache/internal/mysql-memory/mysql_memory_ops.go` lines 257 and 272 use `gdb.Raw("LAST_INSERT_ID(value_int + " + delta + ")")` and `Raw("SELECT LAST_INSERT_ID()")` for "atomic increment and retrieve new value", a MySQL session-level Last-Insert-ID trick that SQLite does not support
- **Cluster coordination dependency**: `sys_locker` / `sys_kv_cache` serve as both business storage and multi-node shared coordination carriers, relying on MySQL cross-connection shared visibility

GoFrame natively supports SQLite via `github.com/gogf/gf/contrib/drivers/sqlite/v2`, with file link format `sqlite::@file(path/to/file.db)`. The DAO/DO/Entity layer's `do.Xxx` pattern and GoFrame's built-in query builder automatically dialectize most business SQL. Combined with existing project rules prohibiting `FIND_IN_SET` / `GROUP_CONCAT` / `IF()` / `ON DUPLICATE KEY UPDATE` and other database-specific functions, the DAO layer's runtime queries are largely portable. What requires additional handling is limited to four categories: DDL dialect, initialization flow, MEMORY engine semantics, and kvcache's MySQL-specific SQL.

User constraints for this iteration have been clarified:

1. SQLite serves only demo/personal testing scenarios, not production
2. No SQLite-to-MySQL data migration needed; SQLite data may be lost
3. Business modules must be completely unaware of database engine differences; all adaptation converges in the data access layer
4. SQLite mode must force `cluster.enabled=false` and output prominent terminal log warnings
5. `TranslateDDL` interface must accept `sourceName` for source file or embedded asset location in translation errors
6. `make mock` must depend on an already-initialized database; it is not responsible for creating, rebuilding, or preparing the database
7. SQLite DDL translation coverage must fully align with current SQL file real-world syntax, not just example syntax
8. The sole runtime source for database dialect switching is the configuration file `database.default.link`; no command-line arguments or environment variables are introduced
9. `apps/lina-core/pkg/dialect/` is explicitly a public stable package providing a stable dialect capability boundary for the host, plugins, and toolchain

## Goals / Non-Goals

**Goals:**

- Users change `database.default.link` to `sqlite::@file(./temp/sqlite/linapro.db)` in `config.yaml` to run the entire framework with zero dependencies, no MySQL startup needed
- Business modules (`controller` / `service` / `model` / `dao`) require zero code changes, completely unaware of underlying database engine differences
- `pkg/dialect` as a public stable package exposes a narrow interface, not binding host `internal` concrete service types in public signatures
- Single MySQL-dialect SQL file source (maintaining the "single file per iteration" principle), no concurrent `xxx.mysql.sql` / `xxx.sqlite.sql` files
- Plugin SQL resources execute correctly in SQLite mode with no changes to plugin source code
- SQLite DDL translator must cover all real-world syntax forms in current host, plugin, and mock SQL files, including primary key ordering differences, table-level primary keys, `UNIQUE INDEX`, expression indexes, and `CONCAT(...)` in mock DML
- SQLite mode forces single-node mode lock at startup and prints clear startup prompt logs in the terminal to prevent misuse
- Default SQLite database file path `./temp/sqlite/linapro.db` is customizable via `link` configuration; parent directory is auto-created on startup if it does not exist
- Existing MySQL users' default configuration and runtime behavior are fully backward compatible

**Non-Goals:**

- No SQLite cluster deployment support (architecturally impossible; SQLite is an embedded single-instance database)
- No SQLite-to-MySQL bidirectional data migration tools
- No "startup production environment detection safety net" for SQLite mode (user explicitly deferred to a later iteration)
- No modification to business module code, no database-engine-related business branch logic
- No performance optimization for SQLite mode (demo/testing scenarios have extremely low concurrency requirements)
- No redesign or trimming of existing MEMORY tables for SQLite mode (`sys_locker` and `sys_online_session` degrade to plain tables in SQLite, which does not affect correctness)
- `make mock` must not replace `make init`; mock data loading must run on an already existing and initialized target database
- No command-line arguments, Make parameters, or environment variables for runtime database dialect selection; test environments requiring SQLite must generate or modify the test configuration file's `database.default.link` before startup

## Decisions

### D1: Single-Point Convergence with `pkg/dialect` Abstraction Layer

**Choice**: Add `apps/lina-core/pkg/dialect/` public stable package defining a unified `Dialect` interface as the sole boundary for database engine difference convergence. All dialect-related logic (DDL translation, database preparation, cluster capability query, startup hooks) is exposed to callers through this interface. The package targets host, plugin lifecycle, initialization commands, and toolchain reuse; public signatures must maintain a narrow interface and not directly reference host `internal` concrete service types; MySQL / SQLite concrete implementations converge under `pkg/dialect/internal/mysql` and `pkg/dialect/internal/sqlite`, with only factory functions, the `Dialect` interface, and necessary public facade capabilities retained in the public package.

```go
// Dialect abstracts all database engine difference points.
type Dialect interface {
    // Name returns the dialect name (e.g., "mysql" / "sqlite") for logging and diagnostics.
    Name() string

    // TranslateDDL translates single MySQL-dialect source DDL content into dialect-executable statements.
    // sourceName is the file path, embedded asset path, or caller-provided diagnostic name for error location.
    // MySQL implementation is a no-op; SQLite implementation performs token replacement + structure extraction.
    TranslateDDL(ctx context.Context, sourceName string, ddl string) (string, error)

    // PrepareDatabase prepares the database before executing DDL resources (create db / file, optional rebuild).
    // MySQL implementation executes DROP/CREATE DATABASE; SQLite implementation creates parent directory, optionally deletes file.
    PrepareDatabase(ctx context.Context, link string, rebuild bool) error

    // SupportsCluster indicates whether this dialect can serve as a multi-node shared state coordinator.
    // MySQL=true, SQLite=false.
    SupportsCluster() bool

    // OnStartup is called during the startup bootstrap phase for dialect-specific runtime initialization
    // (e.g., SQLite mode locks cluster.enabled and outputs startup prompt log).
    OnStartup(ctx context.Context, runtime RuntimeConfig) error
}

// RuntimeConfig is the stable narrow interface that the dialect startup phase needs, adapted by host config.Service.
type RuntimeConfig interface {
    OverrideClusterEnabledForDialect(value bool)
}

// From returns the corresponding dialect implementation based on the link's protocol prefix.
// Callers may only depend on the Dialect interface, not on concrete implementation types.
func From(link string) (Dialect, error)
```

**Rationale**:

- Business modules are unaware: all dialect differences do not appear in business paths, physical guarantee for constraint #3
- Single extension point: future PostgreSQL support only requires adding a `Dialect` implementation, no caller modifications needed
- Public package boundary stable: `pkg/dialect` does not expose `internal/service/config` or other host private types, nor exports `MySQLDialect` / `SQLiteDialect` concrete types, preventing plugins or toolchain from depending on host internal implementations
- Test isolation: dialect layer can be independently unit-tested (DDL translation, link parsing) without real database dependency
- Centralized startup hooks: `OnStartup` gives "SQLite lock cluster + startup prompt" and other startup behaviors a clear home, not scattered across locations

**Alternatives Rejected**:

- *Option A: `if isSQLite { ... } else { ... }` at each call site*. Rejected: violates constraint #3, engine judgment appears in business paths
- *Option B: fork a separate set of SQLite dialect SQL files*. Rejected: violates "single file per iteration" principle, and dual SQL files would drift over time
- *Option C: embed translation capability in GoFrame `gdb` driver layer*. Rejected: outside this project's scope, requires modifying the upstream framework; runtime queries are already auto-dialectized by `gdb`, what's needed is only DDL and initialization flow translation

### D2: SQLite DDL Translation Strategy Uses "Line Scan + Token Replacement + Structure Extraction"

**Choice**: The translator does not build a complete SQL AST. Instead, based on the existing `cmd_sql_split.go` (which already implements SQL multi-statement splitting with string/comment-aware scanning), for each CREATE TABLE / INSERT statement it performs:

- Lexical layer: pure token-level regex replacement (backtick removal, `AUTO_INCREMENT` / `UNSIGNED` / `LONGTEXT` keyword replacement, `INSERT IGNORE` rewriting)
- Structure layer: identify CREATE TABLE column definition blocks, extract inline `KEY` / `INDEX` / `UNIQUE KEY` lines as independent `CREATE INDEX` statements appended after table creation
- Line scan layer: identify and delete entire `CREATE DATABASE ... ;` and `USE ... ;` statements
- Coverage layer: use current repository real host install SQL, plugin install SQL, host mock SQL, plugin mock SQL, and plugin uninstall SQL as mandatory fixtures, covering real-world `INT/BIGINT [UNSIGNED] PRIMARY KEY AUTO_INCREMENT`, `AUTO_INCREMENT PRIMARY KEY`, `NOT NULL AUTO_INCREMENT` + table-level `PRIMARY KEY(id)`, `UNIQUE INDEX`, expression indexes (e.g., `NULLIF(code, '')`), `CONCAT(...)` and other syntax; if the implementation cannot losslessly translate a real-world syntax, the translator should be modified rather than asking SQL authors to temporarily avoid the syntax

**Rationale**:

- All current MySQL dialect DDL is from controlled sources (project-authored), so generated columns / partitioning / spatial indexes and other explicitly prohibited syntax will not appear; but existing files already contain multiple primary key and index arrangements, so the translator must use real file fixtures as the coverage benchmark for the complete project subset
- Unit tests automatically scan current real SQL assets as fixtures via glob, **end-to-end asserting install, mock, and uninstall translation results are executable on SQLite**, preventing translator implementation detail drift
- Implementation size approximately 200-300 LOC + tests, avoiding full SQL parser dependency

**Alternatives Rejected**:

- *Option A: introduce `vitess/sqlparser` or similar full parser*. Rejected: dependency too heavy, our DDL subset is extremely small, poor cost-benefit
- *Option B: hand-write schema-as-code DSL, generate DDL by dialect*. Rejected: requires rewriting all delivery SQL files, violates "demo purpose, minimal cost" iteration positioning

### D3: MySQL MEMORY Tables Remain Unchanged, Only Degrade to Plain Tables in SQLite Translation Results

**Choice**: Original MySQL delivery SQL maintains a single source; no table structure, index, or engine type changes for SQLite. `sys_locker` / `sys_online_session` / `sys_kv_cache` continue using existing `ENGINE=MEMORY` in MySQL SQL assets; SQLite mode only has the DDL translator remove `ENGINE=MEMORY` clauses before execution, resulting in plain SQLite tables.

**Rationale**:

- The three MEMORY tables' correctness **does not fundamentally depend on** "clear on restart" semantics:
  - `sys_locker` already has an `expire_time` field with application-layer TTL check as fallback; stale locks are cleaned on first TTL check after restart
  - `sys_online_session` already has `last_active_time` plus existing cron cleanup tasks; brief stale session residue after restart does not affect functionality
- `sys_kv_cache` remains a lossy cache; correctness depends on unique key constraints, application-layer TTL, background cleanup, and `incr`'s CAS retry, not on the table engine providing persistent reliable state; the cache must not serve as an authoritative data source
- "Clear on restart" semantics in MEMORY engine is a performance optimization (no need to scan expired entries); performance is not important in SQLite demo scenarios
- No business-layer "clear tables on SQLite startup" adaptation needed; constraint #3 is met
- Not modifying MySQL SQL assets ensures SQLite support is always an execution-time dialect adaptation, not degrading original delivery SQL to an SQLite-compatible subset

**Alternatives Rejected**:

- *Option A: use in-process `sync.Map` / `lru.Cache` for locker / kvcache / online_session in SQLite mode*. Rejected: violates constraint #3, business modules need to perceive backend differences; requires three new backend implementations and startup switching logic
- *Option B: clear these three tables on SQLite startup to simulate "clear on restart"*. Rejected: user has explicitly stated "business modules are unaffected by database engine differences"; simulation cost exceeds letting business rely on TTL fallback (which it already does)
- *Option C: change `013`'s `sys_kv_cache` to InnoDB or plain table*. Rejected: violates "do not modify original MySQL table structure or engine type" boundary; atomic increment should be achieved through dialect-neutral CAS, not by changing MySQL delivery SQL to obtain transactional row locks

### D4: `kvcache` Backend Renamed to `sqltable` and Rewritten with Dialect-Neutral SQL

**Choice**:

- Package path `apps/lina-core/internal/service/kvcache/internal/mysql-memory/` -> `apps/lina-core/internal/service/kvcache/internal/sqltable/`
- Constant `BackendMySQLMemory` -> `BackendSQLTable`, string value `"mysql-memory"` -> `"sql-table"`
- `Incr` operation implementation: `gdb.Raw("LAST_INSERT_ID(value_int + N)")` + `Raw("SELECT LAST_INSERT_ID()")` replaced with dialect-neutral CAS + bounded retry flow:
  1. Read current row's `value_kind`, `value_int`, `expire_at` snapshot
  2. If row does not exist, use `INSERT IGNORE value_int=0` for idempotent initialization of missing integer row, then re-read snapshot
  3. If existing row is not integer, return structured error without modifying original value
  4. Calculate `next=value_int+delta`, execute `UPDATE ... SET value_int=next ... WHERE value_kind=int AND value_int=<snapshot>`
  5. If `affected=0`, competing write has changed snapshot, retry with bounded backoff; for MySQL deadlock / lock wait timeout and SQLite busy/locked and other database-suggested retryable lock conflicts, apply the same bounded retry

**Rationale**:

- Package name `mysql-memory` is already misleading: under SQLite it still works but the name exposes MySQL implementation details, violating abstraction layer intent
- Rewritten implementation **single backend runs both MySQL and SQLite**, no need to add a separate SQLite backend implementation, conforming to constraint #3
- Project coding standards explicitly prohibit `FIND_IN_SET` and other MySQL-specific functions; `LAST_INSERT_ID(v + delta)` is substantively the same type of issue and should be eliminated together
- CAS approach does not depend on transactional row locks or InnoDB engine: MySQL MEMORY's single `UPDATE` executes atomically under table locks, SQLite writes serialize; `WHERE value_int=<snapshot>` detects competing writes, successful calls still return unique incrementing values
- Missing key initialization through `INSERT IGNORE value_int=0` and subsequent CAS update maintains consistency with existing semantics; first `incr(delta)` returns `delta`
- This approach preserves `013`'s original MySQL `ENGINE=MEMORY` asset while avoiding database-specific atomic tricks like `RETURNING` / `LAST_INSERT_ID`

**Alternatives Rejected**:

- *Option A: keep `mysql-memory` name, add parallel `sqlite` backend*. Rejected: two backend implementations have mostly duplicate code; startup dispatch logic adds unnecessary complexity
- *Option B: use `RETURNING` clause*. Rejected: introduces database version differences and proprietary syntax, less stable than CAS approach covering both MySQL and SQLite
- *Option C: change `sys_kv_cache` to InnoDB and use transaction locks*. Rejected: modifies existing MySQL delivery SQL table engine; SQLite support should be achieved through dialect translation and dialect-neutral syntax

### D5: SQLite Cluster Lock via `OnStartup` Hook in Startup Bootstrap

**Choice**:

- SQLite dialect instance's `OnStartup(ctx, runtime)` is called once during startup
- Implementation: calls `configSvc.OverrideClusterEnabledForDialect(false)` (new in-memory override method), and outputs clear startup prompt log
- `IsClusterEnabled` stably returns `false` after the override; no need for dialect awareness at each call site

**Startup Prompt Log Example**:

```
SQLite mode detected (database.default.link = sqlite::@file(./temp/sqlite/linapro.db))
SQLite mode only supports single-node deployment, cluster.enabled has been forced to false
Do not use in production
```

**Rationale**:

- Startup hook gives "lock cluster" behavior a single clear execution point, not scattered across `cluster.Service` / `config.Service` / various callers
- Override inside `IsClusterEnabled` rather than at each call site preserves reusability of all existing cluster-linked logic (frontend UI hiding, cron job degradation, cache coordination downgrade)
- Startup prompt log is clearly visible in terminal default log output

### D6: Default SQLite Database File Path `./temp/sqlite/linapro.db`, Customizable via `link`

**Choice**:

- Default value (given as example in `config.template.yaml` comments): `sqlite::@file(./temp/sqlite/linapro.db)`
- Users can freely change the path by modifying the `link` field, e.g., `sqlite::@file(/var/lib/myapp/db.sqlite)`
- SQLite dialect instance's `PrepareDatabase` automatically `mkdir -p` parent directory before execution; directory creation failure returns a clear error (including path and permission hints)
- Default path falls under the existing `temp/` ignore rule in the repository root `.gitignore`, no additional SQLite-specific ignore entries needed

**Rationale**:

- `temp/` directory already has "temporary artifacts" semantics in the project (CLAUDE.md specifies e2e screenshots go here); SQLite demo/test database semantics match
- Relative path `./temp/sqlite/` resolves from `gfile.Pwd()`, consistent with `make init` / `make dev` working directory
- Automatic `mkdir -p` means first startup requires no manual directory creation, zero-dependency experience

### D7: Dialect Switching Only via `config.yaml`, No Command-Line Arguments

**Choice**:

- `make init` / `make mock` / `make dev` and other commands remain unchanged; no `dialect=sqlite` / `--dialect=...` parameters added, no `LINAPRO_*` environment variables read as runtime dialect source
- The only switching method is modifying `database.default.link` in the configuration file
- `cmd init` / `cmd mock` internally read `link` then call `dialect.From(link)` for automatic dispatch
- E2E tests requiring SQLite mode must have test fixtures write `database.default.link` into the test configuration file before starting services; runtime dialect source must not come from environment variables or command-line arguments, even in test channels the configuration file remains the sole runtime source; main CI lightweight SQLite smoke similarly switches dialect by writing to configuration file

**Rationale**:

- Single source of truth avoids confusion from command-line parameter and configuration file state inconsistency
- Dialect switching is a low-frequency action (demo scenario is typically set once); configuration file approach is sufficiently intuitive
- Implementation path has one fewer layer (no need to pass dialect parameters through Makefile and cmd flag parsing)

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **DDL/DML translator misses some MySQL syntax** -> translation result fails on SQLite | Unit tests use current real SQL assets as fixtures, end-to-end asserting install, mock, uninstall translation results are executable; new SQL files forced through same test path |
| **New plugin SQL introduces uncovered MySQL syntax** (discovered after archive) | Plugin install pipeline returns clear error on translation failure, locating specific SQL file and line number; spec `plugin-manifest-lifecycle` clarifies "plugin SQL must be processable by default dialect translator" |
| **No MEMORY table "clear on restart" semantics in SQLite mode** -> brief stale locks/sessions after restart | Both data types already have TTL / scheduled cleanup; application layer auto-cleans on first TTL check; demo scenarios are insensitive to brief residue |
| **`Incr` changed from `LAST_INSERT_ID(v+d)` to CAS + lock conflict retry** -> competing calls may retry due to snapshot changes under high concurrency | Only used in `kvcache` one place; MySQL MEMORY and SQLite can both detect competing writes through single conditional `UPDATE`, successful calls still linearly increment; exceeding bounded retry limit returns clear error to caller |
| **Users mistakenly deploy SQLite mode in multi-instance setup** | Startup prompt log + cluster forced lock + documentation explicit warning -- triple defense |
| **`temp/sqlite/` directory cleaned by CI causing test failures** | `temp/` is a temporary artifacts directory and already ignored; CI smoke does not need to retain db file before `make init`, initialization auto-rebuilds |
| **Existing plugin mock SQL contains `INSERT INTO ... ON DUPLICATE KEY UPDATE`** | Project rules already prohibit this syntax; this iteration only does a fallback check; if violations found, treat as bug and fix in this change |

## Migration Plan

This change is **fully transparent** to existing MySQL users, with no breaking migration:

- Existing `config.yaml` default is still MySQL link, behavior identical to before the change
- Existing SQL files, business modules, DAO layer all remain unchanged
- Only observable change: `make init` internally adds one dialect dispatch (MySQL path `TranslateDDL` is a no-op, zero overhead)

For users wanting to switch to SQLite, the migration path is:

1. Change `database.default.link` to `sqlite::@file(./temp/sqlite/linapro.db)`
2. Run `make init` -> automatically creates `temp/sqlite/` directory and database file, loads all DDL
3. (Optional) Run `make mock` to load demo data; this command requires database to have been initialized by `make init`, will not create or rebuild the database itself
4. Start `make dev`, note the terminal will output SQLite standalone mode prompt

To switch back to MySQL: restore the original `link` value. The two modes' data is completely independent with no mutual impact.

## Open Questions

None. The prior exploration phase and this round of feedback have clarified all key decisions (SQLite user persona, no data migration needed, business module zero adaptation, cluster lock + startup prompt, file path, package renaming, `TranslateDDL` accepts `sourceName`, `make mock` depends on `make init`, DDL translation covers current real SQL syntax, `kvcache incr` dialect-neutral atomic flow, configuration-file-only, no safety net, `pkg/dialect` as public stable package).
