## Why

The config-management component already has a set of unit tests, but `go test ./internal/service/config -cover` only reports `71.9%` coverage, which is still well below this repository's `80%` target. That component is also responsible for host-critical concerns such as static config reads, protected runtime parameters, public frontend settings, plugin path resolution, and multi-instance snapshot synchronization. Without stronger automated regression protection, later refactors or new config keys can easily introduce subtle regressions.

This iteration therefore needs to complete a maintainable unit-test suite for the config-management component, define the target clearly, fill the currently low-coverage branches, and treat `80%+` coverage as a delivery gate for the component.

## What Changes

- Add unit tests for `apps/lina-core/internal/service/config` to cover submodules with obvious gaps, including plugin dynamic storage paths, protected public-frontend config helpers, runtime-parameter snapshot caching, and the clustered revision controller.
- Add coverage for default values, fallback branches, invalid inputs, and cache behavior in config-read helpers and frequently used getters so error paths are no longer missed while only the happy path is exercised.
- Establish a coverage verification baseline for the config-management component and require `go test ./internal/service/config -cover` to reach at least `80%`, with the verification result recorded during implementation.
- Keep production behavior unchanged while making only minimal testability improvements when necessary, such as extracting injectible dependencies, isolating process-level cache state, and improving test fixtures so the tests remain stable and repeatable.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `config-management`: Add a unit-test coverage constraint for the config-management component and require core config services to pass complete automated regression coverage plus a package-level coverage verification of `80%` or higher before delivery.

## Impact

- **Backend code**: primarily affects test files under `apps/lina-core/internal/service/config/`, with only small implementation adjustments if needed to improve testability.
- **Test verification**: requires new or expanded package-level unit tests and execution of `go test ./internal/service/config -cover` or an equivalent coverage command.
- **Runtime behavior**: adds no external API, changes no database schema, and does not alter any frontend behavior; this is a quality-hardening iteration for a host configuration component.
