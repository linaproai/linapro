## Context

This archive consolidates work from several closely related host-governance efforts completed during overlapping implementation windows:

1. A governed runtime configuration model for host-consumed settings. The host already depended on settings such as JWT expiry, session timeout, upload limits, and login IP blacklists, but the parameter-management layer did not yet provide a clear protected registry, runtime-safe validation, or multi-instance-friendly cache behavior.

2. A host-controlled API document generation system. GoFrame's built-in `/api.json` output could not distinguish source-plugin routes from host routes, could not project source-plugin routes by enablement state, and would have required a duplicate route declaration model if the project had tried to solve the problem through `plugin.yaml`.

3. A unified upload-size default. The repository still carried both a 10 MB and a 16 MB baseline across different default sources, causing inconsistent behavior in new environments.

4. A strengthened unit-test suite for the config-management component. Package-level coverage was at 71.9%, well below the 80% target, with significant gaps in plugin config, public frontend helpers, cache behavior, and cluster revision logic.

Several follow-up improvements were also folded in: plugin detail dialog and host-service presentation refinements, OpenSpec language-governance rules, structured logging and unified log sinks, configuration extension namespacing, and comment-conformance cleanup for affected backend code paths.

## Goals

- Ensure runtime configuration actually drives host behavior instead of existing only as editable key-value records.
- Keep hot-path runtime reads on process memory whenever possible, without breaking multi-instance convergence.
- Preserve source-plugin route flexibility while giving the host explicit route ownership metadata.
- Unify the host upload-size default at 20 MB across all initialization and fallback paths.
- Bring config-management unit-test coverage to 80% or above with meaningful regression protection for high-risk branches.
- Archive the iteration with English proposal, design, tasks, and delta specs.

## Non-Goals

- No route-prefix constraint is imposed on source plugins.
- No duplicate route declaration model is added to `plugin.yaml`.
- No dynamic-plugin middleware model is forced onto source plugins.
- No redesign of the file-upload module or addition of new runtime parameters.
- No removal of the administrator's ability to override `sys.upload.maxSize`.
- No repository-wide coverage governance beyond the `config` package.

## Themes

### 1. Runtime Configuration Model

#### Decision: Runtime configuration is modeled as a protected host-owned contract

Protected runtime parameters and protected public frontend settings are registered centrally in the host configuration service. The contract includes stable key ownership, default values, validation rules, runtime override lookup, and protection against rename and deletion. This keeps parameter governance in one place instead of scattering rules across auth, session, upload, UI bootstrap, and import flows.

#### Decision: Public frontend settings are exposed through a whitelist endpoint

Unauthenticated pages and bootstrap flows need a safe subset of host-managed settings. The design keeps a whitelist contract, structured typed response payloads, and no arbitrary public key lookup. This allows login pages and workspace bootstrap to consume branding and theme settings without exposing generic configuration access.

### 2. Upload Size Configuration

#### Decision: Normalize existing default sources directly instead of adding a compatibility migration layer

The existing host initialization SQL, config template, and backend static fallback are updated directly so 20 MB becomes the only default baseline. This keeps the implementation aligned with the project rule that this is a new project and existing SQL can be updated directly with re-initialization.

- Alternative: add a new SQL iteration file that only overrides the old default value.
- Why not: that would keep default-value ownership split across multiple places and would not fit the current no-compatibility-overhead expectation.

#### Decision: Treat `sys.upload.maxSize` as the single business source of truth and align every fallback with it

`sys.upload.maxSize` already represents host upload-size governance, so the 20 MB default must live not only in config-management seed data but also in the config template and the static fallback inside `config_upload.go`. That guarantees the database default, direct config-file reads, and upload-chain validation all see the same baseline.

- Alternative: change only the initialization SQL.
- Why not: if only SQL changes, any uninitialized or non-overridden path can still fall back to 10 MB or 16 MB and the split behavior remains.

#### Decision: Update user-facing error copy and derived artifacts together

Friendly messages shown when uploads exceed the limit, request-body protection assertions, and any embedded or packaged manifest artifact derived from the manifest all move to 20 MB together. That avoids a state where source code has been updated but build outputs or test baselines still expose the old default.

### 3. OpenAPI Governance

#### Decision: The host owns `/api.json`

The host no longer relies on GoFrame's default OpenAPI output as the source of truth. The host-managed OpenAPI builder scans real host routes for documentable static APIs, excludes internal and non-business routes, excludes plugin routes from the host-static route set, projects enabled source-plugin routes using captured route bindings, and projects enabled dynamic-plugin routes using runtime route contracts. This gives the host precise control over what appears in system API documentation.

#### Decision: Source-plugin route ownership is captured at registration time

Source plugins still define routes in code only. The host wraps route registration with an observable facade that records plugin ID, method, path, handler ownership, and DTO `g.Meta` documentation metadata when present. This avoids path-prefix heuristics and avoids duplicating route declarations in plugin manifests.

#### Decision: Source-plugin middleware remains plugin-owned

Source plugins keep their current flexibility for middleware registration and ordering. The host captures route ownership and performs real binding, but it does not try to reinterpret source-plugin middleware chains as dynamic-plugin-style declarative middleware descriptors.

### 4. Cache Strategy

#### Decision: Runtime reads use local snapshots plus shared revision convergence

Hot-path host behavior does not read `sys_config` for every request. Instead:

- each process keeps an immutable parsed snapshot in local cache
- writers bump a shared revision and clear their own local snapshot immediately
- other nodes converge through periodic revision synchronization
- single-node mode skips the shared coordination path and keeps only the local invalidation model

This design reduces hot-path overhead while preserving bounded cross-node convergence.

### 5. Plugin UI and Operational Follow-up

Several follow-up improvements were folded into the implementation window:

- Plugin detail dialog and host-service presentation refinements in the default admin workspace.
- Plugin resource grouping, labels, empty-state behavior, and layout consistency between detail and authorization dialogs.
- Pagination behavior for the dynamic plugin demo record list.
- Structured logging switch support and alignment of HTTP server logs with business log sinks.
- Host-specific server and logger extensions moved under explicit `extensions` namespaces.

### 6. Unit Test Coverage

#### Decision: Use package-level coverage as the acceptance baseline instead of per-file thresholds

Acceptance is based on `cd apps/lina-core && go test ./internal/service/config -cover` with the final result required to be >= 80%. The built-in Go coverage command is simple, local, and easy to wire into CI later. The gaps are currently spread across multiple files, so a package-level metric fits the goal of raising overall coverage in one pass.

#### Decision: Prioritize low-coverage high-risk branches instead of spreading effort evenly

This iteration prioritizes by coverage gap multiplied by risk:

1. `config_plugin.go`: plugin dynamic storage defaults, compatibility fallbacks, and override logic.
2. `config_public_frontend.go`: protected-key detection, shared validation entry point, whitelist metadata, and time-zone parsing.
3. `config_runtime_params_revision.go` / `config_runtime_params_cache.go`: cluster synchronization, cache rebuilds, and degraded fallbacks.
4. Remaining getters/helpers: default-value, empty-object, and defensive-branch checks.

#### Decision: Design test scenarios in groups of primary path, fallback path, and exception path

Each submodule covers at least two of three path types, and the critical modules cover all three:

- **Primary path**: normal config reads, cache hits, successful parsing.
- **Fallback path**: missing config, compatibility fallback, default values applied, cache reuse.
- **Exception path**: shared-KV failure, corrupt cached value, invalid input, empty object, or unparseable state.

## Risks and Trade-offs

- Source-plugin raw handlers remain registrable but are not automatically projected into OpenAPI without DTO metadata, which is intentional to avoid a second route-truth system.
- Derived package artifacts may still contain the old upload default. Mitigation: update or regenerate embedded resources as part of the implementation and check both source files and derived outputs.
- Local environments that already changed `sys.upload.maxSize` manually will not automatically become 20 MB. This change targets the default baseline; validation is done in a clean reinitialized environment.
- Global-state cross-contamination in config tests. Mitigation: use shared reset helpers and paired `t.Cleanup` calls to contain contamination.
- Coverage reaches the target but adds little value. Mitigation: the task breakdown explicitly prioritizes low-coverage hot spots and exception paths, not just easy happy-path tests.
