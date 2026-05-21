## 1. CI Sharding and Basic Governance

- [x] 1.1 Fix browser E2E workflow PostgreSQL health check, explicitly using `pg_isready -U postgres -d linapro`.
- [x] 1.2 Change plugin-full E2E job to shard execution based on generic entry points, with shards covering `extension:plugin` and `plugins`.
- [x] 1.3 Generate unique artifact names for plugin-full shards, ensuring Playwright report, test-results, backend logs, and frontend logs do not overwrite each other.
- [x] 1.4 Verify shard failure causes the complete verification suite to fail and blocks dependent jobs.

## 2. Plugin-full Coverage Scope Convergence

- [x] 2.1 Review plugin-full plugin framework generic cases to retain, clarifying generic coverage files for menus, permissions, routes, i18n, tasks, and runtime resources.
- [x] 2.2 Converge root E2E manifest so plugin-full no longer selects root test file sets depending on specific official plugins.
- [x] 2.3 Confirm host-only still covers host full capability, plugin-full still covers official plugin own cases and plugin lifecycle.

## 3. Authentication Fixture and Host Slow Case Optimization

- [x] 3.1 Add lightweight authenticated page fixture in `hack/tests/fixtures/auth.ts` that does not auto-navigate to dashboard, preserving existing `adminPage` behavior.
- [x] 3.2 Prioritize migrating menu CRUD cases suitable for direct business route entry, reducing repeated dashboard loading.
- [x] 3.3 Prioritize migrating file management cases suitable for direct business route entry, reducing repeated dashboard loading.
- [x] 3.4 Evaluate and migrate role CRUD, parameter import, dictionary import, and similar pattern slow files.

## 4. Plugin Baseline and Ordinary Plugin Case Optimization

- [x] 4.1 Add idempotent baseline helper in plugin E2E fixture/support, supporting one-time synchronization, installation, enablement, available mock data loading, and plugin projection refresh.
- [x] 4.2 Migrate repeated `ensureSourcePluginEnabled` in ordinary plugin page tests to suite or shard-level baseline.
- [x] 4.3 Confirm plugin lifecycle tests still explicitly control installation, enablement, disablement, uninstallation, upload, synchronization, and cleanup state, not interfered with by ordinary baseline.

## 5. Lifecycle Heavy User Refactoring

- [x] 5.1 Refactor official source plugin lifecycle cases, retaining one representative official plugin's complete UI lifecycle, changing other official plugins to API/contract smoke plus page accessibility verification.
- [x] 5.2 Refactor dynamic runtime lifecycle cases, distinguishing runtime core UI lifecycle from dynamic demo API/functionality verification, merging repeated assembly and retaining key UI coverage.
- [x] 5.3 Review source plugin lifecycle cases, eliminating mergeable or API-convertible repeated lifecycle steps.

## 6. Verification and Acceptance Records

- [x] 6.1 Run `openspec validate` and fix all specification issues.
- [x] 6.2 Run affected module scope E2E smoke, covering at least `extension:plugin`, one official plugin functionality scope, and migrated host slow files.
- [x] 6.3 Record host-only optimization before/after wall clock, total test time, slowest file, and slowest case comparison.
- [x] 6.4 Record plugin-full optimization before/after wall clock, per-shard time, longest shard, and runner minutes change.
- [x] 6.5 Explicitly record that this change does not affect production API, database schema, runtime cache semantics, and i18n resources; if visible copy or script entry is added during implementation, supplement corresponding governance description.
- [x] 6.6 Complete tasks and execute `/lina-review`, reviewing CI sharding, fixtures, baselines, slow case refactoring, and verification records.

## Verification Records

- Passed `openspec validate --strict`.
- Passed `pnpm -C hack/tests exec tsc --noEmit`.
- Passed `pnpm -C hack/tests test:validate`.
- Passed `git diff --check`.
- Verified local service `http://127.0.0.1:5666` and `http://127.0.0.1:8080/api/v1/health` accessible.
- Host-only optimization baseline from provided logs: job approximately 36 minutes, Playwright report `197 passed (25.1m)`; migrated menu CRUD, file management, role CRUD, parameter import, dictionary import slow cases using `authenticatedPage` without pre-loaded dashboard.
- Plugin-full optimization baseline from provided logs: job approximately 2 hours, `pnpm test` approximately 112 minutes; changed to `extension:plugin` and `plugins` two generic shards.
- Complete host-only E2E passed: `244 passed, 1 skipped (14.6m)`.
- Complete plugin-full E2E passed: `516 passed, 8 skipped (42.8m)`.
- This change only adjusts CI workflow, E2E runner manifest, Playwright fixtures, and test code; does not modify production API, database schema, runtime cache semantics, or user-visible functionality; no new or modified frontend runtime copy, plugin manifest i18n, or apidoc i18n JSON, confirmed no i18n resource synchronization needed.

## Feedback

- [x] **FB-1**: Distinguish host module scope without `apps/lina-plugins` from plugin-full seam scope requiring official plugin workspace.
- [x] **FB-2**: Converge plugin-full scope, only retaining `plugins` and `plugin:<plugin-id>` as generic selection entries for source plugin own cases.
- [x] **FB-3**: Root `hack/tests` E2E code and configuration must not couple any specific official source plugin ID; plugin-related cases must be closed-loop in corresponding plugin directory.
- [x] **FB-4**: Root path E2E test files, configuration, test data, and baseline must not couple any specific plugin information; plugin-related test assets must be in corresponding plugin directory.
- [x] **FB-5**: E2E test file name prefixes no longer globally increment, changed to increment from `TC001` per current module directory.
- [x] **FB-6**: Fix generic plugin resource query layer inconsistency with host `sys_role.data_scope` enumeration.
- [x] **FB-7**: Fix host-only E2E still running some plugin-full or plugin-dependent cases.
- [x] **FB-8**: Fix cross-case state leakage in plugin-full E2E dynamic plugin example records and English layout regression cases.
- [x] **FB-9**: Fix role add/edit drawer async initialization race condition overwriting filled fields.
- [x] **FB-10**: Split `plugins` scope from single CI job into 5 Playwright shards for nightly plugin-full.
- [x] **FB-11**: Fix dynamic plugin disable sidebar menu hidden assertion race condition.
- [x] **FB-12**: Fix runtime cache invalidation and reconciler notification reason string hardcoding.
- [x] **FB-13**: Fix dynamic plugin multipart upload case assuming wasm artifact smaller than default upload limit.
- [x] **FB-14**: Fix source plugin management table column order assertion using outdated title and position contract.
- [x] **FB-15**: Fix multi-tenant plugin uninstall precondition dialog assertion using outdated localized reason text.
- [x] **FB-16**: Increase default file upload size limit from 20MB to 100MB, maintaining consistency across initialization, configuration template, backend fallback, and packaged assets.
