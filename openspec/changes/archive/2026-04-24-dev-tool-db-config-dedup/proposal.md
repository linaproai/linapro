## Why

`apps/lina-core/hack/config.yaml` currently duplicates database connection settings across the development-only toolchain. `database.default.link` and `gfcli.gen.dao[].link` can drift apart, and that duplication also makes it harder to converge development-time connection parameters for future multi-database scenarios. At the same time, the local `init` / `mock` SQL execution path currently depends on `multiStatements=true` in the MySQL DSN. That ties command behavior to a driver-specific capability, which works against unified development-time configuration and later database adapter expansion.

## What Changes

- Converge the host development-only tool config by rewriting the duplicated database connection settings in `apps/lina-core/hack/config.yaml` to use shared YAML anchors.
- Remove `multiStatements=true` from the host development-only tool config and align the connection behavior used by `database.default.link` and `gfcli.gen.dao[].link`.
- Adjust local `init` / `mock` SQL execution so delivered SQL still runs in file order with fail-fast behavior without relying on driver-level multi-statement execution.
- Add unit tests for SQL splitting, blank/comment handling, and failure interruption semantics so the development tooling change remains stable.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `database-bootstrap-commands`: Update the local SQL bootstrap command behavior so development-only SQL sources no longer depend on `multiStatements` in the DSN while still preserving ordered execution and fail-fast semantics.

## Impact

- Affected code: `apps/lina-core/hack/config.yaml`, `apps/lina-core/internal/cmd/`, and related command unit tests.
- Affected systems: the host development-only bootstrap/tooling flow (`make init`, `make mock`, `gf gen dao`).
- Dependencies: GoFrame command/database access paths and local SQL parsing/execution helper logic.
