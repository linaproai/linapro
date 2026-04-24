## Context

The host development-only tool configuration is currently concentrated in `apps/lina-core/hack/config.yaml`. That file is shared by the local `init` / `mock` SQL execution path, `gf gen dao` code generation, and repository-level upgrade tooling that reads development metadata. Today the same database connection settings are declared twice in one file: once for `database.default.link` and again for `gfcli.gen.dao[].link`. As soon as those two values diverge, the development toolchain starts to drift.

Local `init` / `mock` execution currently passes the full SQL file content directly to `g.DB().Exec`. Multi-statement SQL files work only because the development DSN explicitly enables `multiStatements=true`. That constraint not only makes the DSN inconsistent with the one used by `gf gen dao`, it also binds development-only command behavior to a MySQL driver feature and makes it harder to unify development database configuration later.

This change only covers development-only tooling config and local SQL execution. It does not touch runtime config files such as `manifest/config/config.yaml`, and it does not change delivered SQL content or runtime default resource sources.

## Goals / Non-Goals

**Goals:**
- Remove duplicated database connection settings inside `apps/lina-core/hack/config.yaml` so development tooling shares one base connection definition.
- Remove `multiStatements=true` from development-only DSNs and stop local SQL bootstrap commands from depending on driver-level multi-statement support.
- Preserve the existing `init` / `mock` semantics for ordered execution, fail-fast behavior, and error localization.
- Add repeatable unit tests that cover SQL splitting and execution boundaries.

**Non-Goals:**
- Do not change runtime config structures or the delivery strategy under `manifest/config/`.
- Do not extend a unified rendering mechanism to plugin `backend/hack/config.yaml` files.
- Do not rewrite delivered SQL semantics or add a new database driver or external SQL parser dependency.

## Decisions

### 1. Reuse one development database connection in host `hack/config.yaml` through YAML anchors
- Define one shared database connection anchor in the file and let both `database.default.link` and `gfcli.gen.dao[].link` reference it.
- Remove `multiStatements=true` from the shared DSN so local SQL execution and `gf gen dao` use the same base connection settings.
- This fixes single-file configuration drift directly without adding extra scripts or a template rendering step.

**Alternatives considered:**
- Render `hack/config.yaml` dynamically through scripts or Makefile logic: more flexible, but too heavy for the narrow goal of removing duplication inside one file.
- Keep two DSNs and rely on comments to keep them aligned: does not structurally prevent future drift.

### 2. Split SQL statements explicitly in the command layer instead of relying on driver multi-statement execution
- Add SQL splitting helpers under `apps/lina-core/internal/cmd/` and turn each SQL file into an ordered statement list that is executed one statement at a time.
- The splitter must ignore blank fragments and correctly handle common comments and semicolons inside string literals so statements are not cut incorrectly.
- `executeSQLAssetsWithExecutor` keeps its fail-fast behavior, but the execution granularity changes from "run the whole file at once" to "run statements inside the file in order and stop the full bootstrap as soon as any statement fails."

**Alternatives considered:**
- Keep relying on `multiStatements=true`: conflicts with the goal of unifying development DSNs and is unfriendly to future multi-database adaptation.
- Introduce an external SQL parser: more powerful, but adds dependency and maintenance cost that is not justified by the current delivered SQL complexity.

### 3. Cover the behavior change with command-layer unit tests
- `apps/lina-core/internal/cmd/cmd_test.go` already covers confirmation tokens, resource discovery, and fail-fast behavior; this change extends that layer with statement splitting and sequential execution tests.
- The new tests focus on four cases: multi-statement files executing in order, blank/comment fragments being skipped, semicolons inside string literals not causing a split, and immediate termination after a statement failure.
- If `hack/config.yaml` structure needs verification, use lightweight config-loading tests or golden-style assertions for anchor expansion instead of depending on a real database.

## Risks / Trade-offs

- [Risk] A custom SQL splitter may miss edge cases and split a delivered SQL file incorrectly. -> Mitigation: write targeted tests against the current delivered SQL style and prioritize common boundaries such as strings, line comments, and block comments.
- [Risk] Once failure granularity changes from file-level to statement-level, debugging becomes harder if error logs lose context. -> Mitigation: keep the failing file name in error output and add statement index context in logs.
- [Risk] YAML anchors solve only single-file duplication; future cross-file config sharing would still need another approach. -> Mitigation: make the design explicit that this change only solves duplication inside the host's single development config file.

## Migration Plan

1. Update `apps/lina-core/hack/config.yaml` first to converge database connection settings through YAML anchors and remove `multiStatements=true`.
2. Implement SQL splitting and sequential execution in the command layer so local `init` / `mock` continues to work with the shared DSN.
3. Add and run unit tests under `apps/lina-core/internal/cmd` to validate splitting, execution order, and fail-fast semantics.
4. If a local database environment is available, run one additional development bootstrap verification; otherwise unit tests are the required proof for this change.

## Open Questions

- Do current delivered SQL files contain any MySQL-specific edge case, such as stored procedure bodies with custom delimiters, that the splitter must support explicitly? If not, the first implementation can intentionally target conventional DDL/DML files only.
