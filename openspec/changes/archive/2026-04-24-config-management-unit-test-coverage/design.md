## Context

`apps/lina-core/internal/service/config` is the host-level configuration service component. It is responsible for several kinds of behavior:

- reading `config.yaml` static configuration and applying default fallbacks;
- reading, validating, caching, and snapshot-synchronizing protected runtime parameters from `sys_config`;
- assembling the public-frontend whitelist and validating protected public settings;
- reading host configuration for plugin dynamic storage paths, metadata, and OpenAPI behavior;
- synchronizing runtime-parameter revision state in single-node and clustered modes.

The package already includes some tests, but `cd apps/lina-core && go test ./internal/service/config -cover` currently reports `71.9%`, which still misses the `80%` target. Coverage details show the main gaps in the following areas:

- `config_plugin.go`: plugin dynamic storage defaults, compatibility fallbacks, and override logic are barely covered;
- `config_public_frontend.go`: helper branches around `PublicFrontendSettingSpecs`, `IsProtectedConfigParam`, `ValidateProtectedConfigValue`, and time-zone parsing are under-tested;
- `config_runtime_params_cache.go`: uncommon branches such as cache hits, cache invalidation, fallback-on-error, and invalid cached value cleanup are not covered well;
- `config_runtime_params_revision.go`: clustered revision reads, sync, increments, and shared-KV error branches do not yet have a full test set;
- a small set of getters/helpers, such as `GetJwtSecret`, `GetSessionTimeout`, `GetUploadPath`, and `mustScanMetadataConfig`, still lack default-value and empty-object branch checks.

This change adds no new feature. It strengthens automated protection across those host-critical paths and turns `80%+` package-level coverage into an explicit delivery bar for the config-management component.

## Goals / Non-Goals

**Goals:**
- Make package-level unit-test coverage for `apps/lina-core/internal/service/config` reach and remain at `80%` or above.
- Prioritize default-value, fallback, exception, cache, and cluster-sync tests for the currently low-coverage high-risk submodules.
- Make the new tests repeatable and isolated, so caches, config adapters, and global overrides do not pollute one another.
- Keep external behavior unchanged while allowing small, testability-oriented cleanup if needed so future coverage work becomes easier.

**Non-Goals:**
- Do not add new config-management pages, APIs, database schema, or runtime-parameter features.
- Do not expand this work into repository-wide coverage governance for all of `apps/lina-core`; the goal is limited to the `config` package.
- Do not add meaningless test stub code or break current abstractions just to inflate coverage.
- Do not replace unit tests with E2E coverage; this iteration still focuses on service-layer and pure-function regression paths.

## Decisions

### 1. Use package-level coverage as the acceptance baseline instead of per-file thresholds

Acceptance for this change is based on:

- `cd apps/lina-core && go test ./internal/service/config -cover`

with the final result required to be `>= 80%`.

Reasons:

- the built-in Go coverage command is simple, local, and easy to wire into CI later;
- the gaps are currently spread across multiple files, so a package-level metric fits the goal of raising overall coverage in one pass;
- per-file thresholds are more detailed but add higher maintenance cost, which is not justified for this iteration.

An alternative was to assign a minimum percentage per file, but the benefit is too small right now and risks over-coupling implementation to test metrics.

### 2. Prioritize low-coverage high-risk branches instead of spreading effort evenly across all files

This iteration prioritizes by coverage gap multiplied by risk:

1. `config_plugin.go`: governs plugin dynamic storage paths, currently has almost no coverage, and contains compatibility fallback plus override logic.
2. `config_public_frontend.go`: covers protected-key detection, the shared validation entry point, whitelist metadata, and time-zone parsing.
3. `config_runtime_params_revision.go` / `config_runtime_params_cache.go`: govern cluster synchronization, cache rebuilds, and degraded fallbacks, which are the easiest places for subtle regressions during refactors.
4. Remaining getters/helpers: add default-value, empty-object, and defensive-branch checks.

That approach raises both coverage and risk protection quickly with the smallest amount of test code. The alternative of giving every file one or two average tests would still leave the most important exception branches exposed.

### 3. Reuse the existing fake services and state-reset patterns, and only make minimal testability refactors when required

`config_runtime_params_test.go` already provides a `fakeRuntimeParamKVCacheService`, runtime-parameter fixtures, and patterns for handling local config state during tests. This change continues that approach:

- control `kvcache` behavior through fake implementations or replaceable dependencies;
- explicitly reset static caches, runtime snapshots, revision state, and overrides in each test;
- use paired cleanup helpers for process-level state such as plugin storage overrides and runtime snapshot caches.

If current code is not isolated enough for stable tests, only minimal cleanup is allowed, such as extracting reset helpers, isolating global-variable access, or exposing necessary package-private helpers. Structural rewrites that change production semantics are out of scope.

### 4. Design test scenarios in groups of primary path, fallback path, and exception path

To avoid coverage numbers rising while regression protection stays shallow, each submodule should cover at least two of the following three path types, and the critical modules should cover all three:

- **primary path**: normal config reads, cache hits, successful parsing;
- **fallback path**: missing config, compatibility fallback, default values applied, cache reuse;
- **exception path**: shared-KV failure, corrupt cached value, invalid input, empty object, or unparseable state.

For example:

- plugin-config tests must cover the default directory, `runtime.storagePath` fallback, and override cleanup;
- public-frontend tests must cover protected-key detection, validation dispatch, and invalid enum/boolean failures;
- revision-controller tests must cover clustered `GetInt`/`Incr` success plus error propagation;
- snapshot-cache tests must cover local cache hits, rebuilds after revision changes, invalid entry removal, and fallback behavior.

### 5. Record coverage as an implementation verification result and avoid introducing another test framework

The repository already uses Go's standard `testing` package, `go test -cover`, and package-local fake implementations to cover the required scenarios. This change therefore does not add assertion libraries or monkey-patch dependencies. That keeps the test style consistent and avoids expanding the dependency surface just to support the coverage work.

## Risks / Trade-offs

- **[Global-state cross-contamination]** -> Config adapters, static caches, runtime snapshots, and overrides are all process-level state. Use shared reset helpers and paired `t.Cleanup` calls to contain contamination.
- **[Over-designing production code for tests]** -> Only minimal testability cleanup is allowed; do not add abstractions that are detached from actual business semantics.
- **[Coverage reaches the target but adds little value]** -> The task breakdown explicitly prioritizes low-coverage hot spots and exception paths, not just easy happy-path tests.
- **[Cluster/cache tests become brittle]** -> Prefer deterministic tests built on fake `kvcache.Service` implementations and package-local state control instead of any real external environment.

## Migration Plan

1. Start by improving fixtures and state-reset helpers for the low-coverage `config` submodules.
2. Add unit tests incrementally for plugin config, public frontend config, runtime revision, runtime snapshot cache, and getter helpers.
3. After each batch, run `go test ./internal/service/config -cover` and watch the package-level percentage.
4. Record the final coverage result once all tasks are complete and only move to review after the result reaches `80%` or above.

## Open Questions

- Should a later iteration wire the `config` package coverage check into a shared CI gate? For this iteration, command-level verification during implementation is enough.
- If some branches can only be tested stably after a heavier refactor, should small dependency-boundary extraction be allowed? The default for this change is yes, but only for the smallest necessary refactor rather than an architectural rewrite.
