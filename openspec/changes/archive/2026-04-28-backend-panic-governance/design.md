## Context

Production backend `panic` calls were concentrated in four areas: startup command tree and driver registration, source-plugin registration contracts, runtime configuration parsing, and Excel or dynamic-plugin helper functions. The first two categories are unrecoverable startup or registration failures and match the project rules. The latter two sit in ordinary request, file import/export, dynamic plugin loading, and runtime parameter reading paths, and should use `error` returns, unified business error responses, or controlled degradation instead.

The project is new and does not need compatibility with old APIs or historical behavior. Internal function signatures, call chains, and tests can therefore be adjusted directly instead of keeping helper shapes that would keep inviting misuse.

## Goals / Non-Goals

**Goals:**

- Define a production-code `panic` allowlist: startup, initialization, unrecoverable critical paths, `Must*` semantic constructors, and unknown panic rethrow.
- Convert recoverable errors in ordinary business paths into explicit `error` returns so controllers, services, and plugin loading flows can use the existing error channel.
- Use a "strict write validation, explicit read failure" strategy for invalid runtime configuration values, preventing bad data from being silently swallowed in high-traffic business APIs.
- Add static checks for new or retained `panic` calls, requiring new production-code panics to be present in the allowlist.
- Record the i18n decision: this change does not add UI copy, modify API DTO documentation source text, or change manifest/apidoc i18n resources.

**Non-Goals:**

- Do not remove startup or source-plugin registration fail-fast behavior.
- Do not change the database schema or add SQL initialization files.
- Do not add frontend page interactions or E2E tests; this change is covered by Go unit tests and static checks.
- Do not change panics that may be produced by GoFrame itself, except for controlled conversion in host middleware paths that are already identified.

## Decisions

1. Keep `Must*` and startup or registration panics, and convert runtime paths to `error`.

   A repository-wide `panic` ban was considered, but it would reduce diagnostics for startup configuration, DB driver registration, and source-plugin contract failures. The allowlist approach matches project rules and still lets unrecoverable errors fail fast.

2. Excel coordinate helpers no longer return only bare `string` values.

   The previous `cellName(col, row) string` helper could only panic when `CoordinatesToCellName` failed. Implementation should prefer direct calls such as `excelutil.SetCellValue(file, sheet, col, row, value)`. Call sites that truly need an A1 coordinate string use `cellName(...)(string, error)` and return the error through the call chain.

3. Dynamic plugin hostServices normalization is split into an error-returning version and a Must version.

   `NormalizeHostServiceSpecs` previously panicked on validation failure even though it was used with dynamic plugin artifacts, catalog releases, and authorization flows. The implementation adds `NormalizeHostServiceSpecsE` returning `([]*HostServiceSpec, error)` for dynamic input paths. `MustNormalizeHostServiceSpecs` may remain for genuine compile-time constant paths.

4. Runtime configuration reads return explicit errors.

   Protected parameters written to `sys_config` remain strictly validated by `ValidateProtectedConfigValue`. When manual SQL, external writes, or cache pollution produce invalid values in read snapshots, business getters no longer panic and no longer log a warning before returning defaults. They return `error` through the call chain instead, allowing controllers, middleware, and services to expose the configuration problem through the unified error channel.

5. Static checks use an allowlist file or script.

   Code review alone is not enough to prevent regressions. The new check scans production Go files under `apps/lina-core` and `apps/lina-plugins`, excludes `_test.go`, and requires every `panic(` call site to match an allowlist entry with a category and reason. This keeps the small set of necessary panics while blocking new runtime helper panics.

## Risks / Trade-offs

- Explicit runtime configuration errors will fail requests that depend on invalid configuration. This is the intended fail-visible behavior, while strict write validation still prevents normal management entries from saving invalid values.
- Helper signature changes touch multiple export flows. The implementation keeps the changes local and covers affected service/plugin packages with `go test`.
- An allowlist can degrade into "register and pass." Each allowlist entry must include its category and reason, and tests freeze the current allowed set.
- Dynamic plugin hostServices normalization returning errors requires better call-chain context. Artifact, catalog, and authorization paths wrap errors separately so users can identify the specific plugin or artifact.
