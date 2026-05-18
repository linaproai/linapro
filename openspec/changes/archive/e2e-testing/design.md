## Context

### Current State

- `hack/tests/e2e/` already contains close to 100 `TC*.ts` files, but the directory tree still largely follows the legacy workbench grouping. The oversized `system/` directory still contains most host-owned test cases, even though it no longer matches the stable workbench menu boundaries.
- Frontend static routes have already stabilized around capability groups such as `IAM`, `System Settings`, `Task Scheduling`, `Extension Center`, `About`, and `Dashboard`, while organization, content, and monitoring capabilities have been heavily pluginized. If the E2E suite keeps following the old `system/monitor/plugin` grouping, locating and maintaining coverage will only become more expensive.
- `hack/tests/playwright.config.ts` still uses `workers: 1` and `fullyParallel: false`. In `hack/tests/fixtures/auth.ts`, the `adminPage` fixture performs a full UI login for each test.
- Existing page objects and some test files still contain a large number of `waitForTimeout(...)` calls. Fixed waits both increase total execution time and hide the real readiness signals of the product.
- `hack/tests/e2e/` also includes files that do not satisfy the test-case conventions, such as debug scripts and shared helpers. Duplicate TC IDs have already appeared, which shows that directory and numbering governance can no longer rely on manual discipline.
- Full regressions expose shared-state contention between tests: plugin lifecycle, runtime i18n caches, global configuration, dictionaries, menu permissions, and other cases mutate the same database or cache state when executed in parallel, causing protocol-correct tests to produce false failures.
- While test files can be assigned to serial or parallel execution via `hack/tests/config/execution-manifest.json`, there is no auditable global-state mutation classification and no governance documentation for runtime caches, plugin lifecycle, system parameters, dictionaries, menu permissions, and other shared state.

### Constraints

- The suite must continue to follow the `lina-e2e` conventions: `TC{NNNN}-{brief-name}.ts`, globally unique TC IDs, and each file being independently runnable.
- `make test` and `pnpm test` must keep their meaning as a full E2E regression entrypoint.
- Reorganizing the suite must not reduce valuable coverage, especially for plugin lifecycle, permission governance, task scheduling, and system configuration.
- This is a greenfield project, so there is no need to preserve long-term compatibility with the old test directory shape. The suite can converge directly to the target structure.
- This change targets test infrastructure and test governance; it does not alter business APIs or product runtime behavior. Existing E2E naming conventions, module directory boundaries, and i18n continuous governance requirements must continue to be followed.

## Goals / Non-Goals

**Goals:**

1. Align the E2E directory structure with the current stable workbench capability boundaries and plugin ownership.
2. Provide `smoke`, `module`, and `full` execution layers so developers can get the right level of feedback without always waiting for the entire suite.
3. Reuse authenticated state to remove the cost of repeated UI login in high-frequency fixtures.
4. Replace high-frequency fixed waits with state-based waits without sacrificing stability.
5. Introduce limited parallelism only where execution boundaries are clearly safe.
6. Add automated governance for directory ownership, TC uniqueness, helper placement, and execution manifests.
7. Define "shared global state" as a classifiable, verifiable, auditable E2E execution attribute.
8. Enable runner scripts and validation scripts to detect when high-risk shared-state cases have not entered a serial boundary.
9. Ensure plugin, mock data, language-pack cache, system parameter, dictionary, and menu-permission related tests have stable prerequisites and cleanup strategies.
10. Refactor cache/ETag tests to validate protocol semantics, accepting that legitimate global version refreshes may return fresh resources during a full regression.
11. Produce documented test conflict cases and governance rules for reuse when adding new test cases.

**Non-Goals:**

1. Do not replace Playwright or introduce a second testing framework.
2. Do not rewrite every page object in this change; prioritize the most expensive and most frequently reused paths.
3. Do not shrink business coverage or delete valuable tests as a shortcut for speed.
4. Do not treat backend startup, database initialization, or frontend build time as the core target of this iteration; this change focuses on the E2E suite itself.
5. Do not force all tests into parallel execution; high-risk shared-state cases may remain in a serial pool.
6. Do not change real business module data models, API contracts, or runtime i18n behavior.
7. Do not disable plugin lifecycle, cache invalidation, permission synchronization, or other real business mechanisms for testing convenience.

## Decisions

### D1. Reorganize the suite around stable capability boundaries instead of legacy URL buckets

**Decision**: Restructure `hack/tests/e2e/` around stable capability boundaries such as `iam/`, `settings/`, `scheduler/`, `extension/`, `monitor/`, `org/`, `content/`, `dashboard/`, and `about/`. Allow second-level directories for clear subdomains such as `scheduler/job/` and `monitor/operlog/`.

```text
hack/tests/
  e2e/
    auth/
    dashboard/
    about/
    iam/
      user/
      role/
      menu/
    settings/
      dict/
      config/
      file/
    org/
      dept/
      post/
      user-org/
    content/
      notice/
      message/
    monitor/
      operlog/
      loginlog/
      online/
      server/
    scheduler/
      job/
      job-group/
      job-log/
    extension/
      plugin/
```

**Rationale**: The primary job of the suite tree is to help developers quickly locate the regression surface of a capability. Capability boundaries have already stabilized in frontend routing, host menus, and plugin directories, so continuing to use an overloaded legacy bucket such as `system/` only adds cognitive overhead.

**Alternatives considered**:
- Keep `system/` and add documentation only: rejected because the directory itself still fails to express the new capability boundaries.
- Organize tests strictly by source ownership: rejected because regression navigation should start from user-facing capabilities, not physical source placement.

### D2. Move helpers, debug scripts, and governance scripts out of `e2e/`

**Decision**: Keep only real `TC*.ts` files under `hack/tests/e2e/`. Move shared API helpers, wait utilities, and data builders into `hack/tests/support/` or `hack/tests/fixtures/`. Move ad-hoc debug scripts into `hack/tests/debug/` or `hack/tests/scripts/`. Add automated validation for duplicate TC IDs, invalid files, and incorrect directory ownership.

**Rationale**: Mixing non-test files into `e2e/` weakens readability and makes review, scanning, and execution harder to reason about. Separating executable tests from support assets simplifies both governance and tooling.

**Alternative considered**:
- Allow a small number of colocated helpers next to tests: rejected because the rule would quickly erode and reintroduce the same drift.

### D3. Keep the full regression default and add manifest-driven smoke/module entrypoints

**Decision**:
- Keep `pnpm test` as the full regression entrypoint so it remains compatible with existing habits and `make test` expectations.
- Add `pnpm test:full` as an explicit full-suite alias.
- Add `pnpm test:smoke` and `pnpm test:module -- <scope>` as fast-feedback entrypoints driven by a suite manifest instead of ad-hoc globs.
- Maintain an execution manifest such as `config/execution-manifest.json` as the source of truth for smoke files, module scopes, and serial boundaries.

**Rationale**: Day-to-day development needs fast feedback, not a full regression on every iteration. Keeping `pnpm test` as a full run avoids surprise behavior changes, while manifest-driven fast entrypoints keep the workflow standardized and discoverable.

**Alternatives considered**:
- Change `pnpm test` to smoke: rejected because it would silently change an established workflow.
- Rely on README instructions for custom globs: rejected because discoverability and consistency would remain weak.

### D4. Reuse authenticated state via pre-generated `storageState`

**Decision**:
- Add a one-time login preparation step that generates an admin `storageState` before the suite runs.
- Update `adminPage` and similar fixtures to consume that prepared state directly.
- Keep real login flows available for authentication-focused tests such as login, logout, failed login, and unauthenticated redirect scenarios.

**Rationale**: Most back-office tests are validating post-login capability pages, not the login flow itself. Repeating UI login in every file is one of the biggest sources of avoidable runtime and instability.

**Alternative considered**:
- Use API login to inject token or cookie state: not chosen for now because a UI-generated `storageState` keeps the coupling to login internals lower.

### D5. Replace fixed waits with reusable state-based waits

**Decision**: Govern waits at the page-object and fixture boundaries by extracting shared readiness helpers for:
- table readiness
- drawer and modal readiness
- toast and feedback readiness
- route readiness
- dropdown visibility and confirmation overlays

Prioritize the highest-frequency page objects such as menus, roles, dictionaries, users, and configuration pages.

**Rationale**: Fixed waits are both a linear performance tax and a common source of flakes. Centralizing wait behavior inside shared helpers and page objects yields broad benefits with a contained change surface.

**Alternative considered**:
- Replace `waitForTimeout` calls one by one inside individual tests: rejected because the payoff is fragmented and hard to sustain.

### D6. Use file-level pool splitting for parallel safety

**Decision**:
- Keep `fullyParallel: false` so each file still runs in order.
- Raise file-level throughput with a limited worker count controlled by configuration and environment variables.
- Place obviously shared-state scenarios such as plugin lifecycle, permission governance, runtime config, import/export, and scheduling into an explicit serial pool.
- Run the full suite in two phases: a parallel pool for isolated files and a serial pool for shared-state files.

**Rationale**: Current tests already have meaningful file-level isolation, so `workers: 1` artificially limits throughput. At the same time, turning on broad parallelism would amplify shared-state conflicts. A pool-splitting strategy gives a practical balance between stability and speed.

**Alternatives considered**:
- Enable broad multi-worker execution for everything: rejected because the shared-state surface is still too large.
- Keep the suite permanently single-worker: rejected because it cannot satisfy the efficiency goal.

### D7. Enforce suite governance with automated validation

**Decision**: Add a validation script that checks at minimum:
- global TC ID uniqueness
- non-`TC` files under `hack/tests/e2e/`
- allowed directory ownership for test files
- valid smoke, serial, and module references in the execution manifest

The validator serves both as the migration acceptance tool and as a standing guardrail for future changes.

**Rationale**: Duplicate TC IDs and invalid files have already shown that manual discipline is insufficient. Automated validation is inexpensive and provides durable governance.

### D8. Introduce global-state classification in the execution manifest

**Decision**: Maintain machine-readable classifications for serial entries or test files in `execution-manifest.json`, such as `pluginLifecycle`, `runtimeI18nCache`, `systemConfig`, `dictionaryData`, `permissionMatrix`, `sharedDatabaseSeed`. The runner script continues to execute serial files as a set, while the validation script additionally checks classification completeness.

**Rationale**: Relying solely on file paths or manual conventions to determine serial boundaries has low cost but cannot explain why a given file must be serial, nor automatically alert when new high-risk cases are added.

**Alternative considered**:
- Rely on file paths or manual conventions only: rejected because it cannot explain serial rationale or auto-detect omissions.

### D9. Supplement manual classification with static heuristic detection

**Decision**: `validate-e2e.mjs` should scan for high-risk APIs, helpers, or keywords such as plugin `install/enable/disable/uninstall/sync`, system parameter writes, dictionary import/modify, menu permission modifications, runtime language-pack cache assertions, `localStorage` ETag caching, and more. When a high-risk pattern is detected but the file is not in the serial boundary or lacks a declared classification, validation should fail with actionable remediation guidance.

**Rationale**: Heuristics do not pursue perfect semantic analysis; they serve only as an engineering guard against omissions. Complex false positives may be handled through explicit classification or allowlist entries, but each allowlist entry must document its justification.

### D10. Consolidate plugin and mock-data prerequisites into fixtures

**Decision**: Tests that depend on source-plugin or dynamic-plugin state must prepare plugin state through unified fixtures/helpers. Fixtures must provide idempotent installation, enabling, necessary mock SQL loading, and frontend plugin projection refresh. Test files should not implicitly depend on another file having installed a plugin, nor assume that a specific plugin table or mock data already exists.

**Rationale**: This adds a small amount of setup time but yields single-file runnability and full-regression stability.

### D11. Cache tests validate protocol semantics rather than fixed response codes

**Decision**: Runtime i18n ETag, public config cache, plugin frontend resource generation, and similar tests should verify "whether the request carries the correct conditional header" and "whether the server response matches the current resource version." For example, when reloading with a stale ETag, `304` indicates the version has not changed, and `200 + new ETag + body` indicates the version has legitimately refreshed; both are correct.

Only when the test exclusively owns the resource version and there is no concurrent global-state mutation should a fixed `304` be asserted.

**Rationale**: Global resource versions may legitimately refresh during a full regression due to plugin lifecycle, language-pack operations, or other parallel tests. Protocol-semantic assertions avoid false failures while still verifying cache behavior correctness.

### D12. Prefer stable fields for business-state assertions

**Decision**: When testing business counts, states, and permissions, prefer stable API fields, IDs, codes, labelKeys, and permission keys over inferring business state from localized UI text. UI copy should still be tested, but as separate presentation assertions, not as part of cross-language business-state computation.

**Rationale**: This avoids language switching, traditional/simplified Chinese switching, or copy adjustments causing false business-test failures.

## Risks / Trade-offs

- **Directory migration causes many import path updates**: migrate module by module, update support-file destinations first, and run targeted regressions after each batch.
- **Authenticated-state reuse could hide login issues**: keep `auth/` tests on real login flows.
- **Incorrect serial/parallel classification could introduce flakes**: start conservatively, allow worker fallback, and keep obviously shared-state files in the serial pool.
- **Partial wait cleanup could reduce the speed gain**: prioritize the highest-frequency page objects and shared flows first.
- **A poorly curated smoke pack could drift away from real risk**: seed it with login, workspace navigation, core management CRUD, and plugin-governance paths, then evolve it with the product.
- **Overly broad serial classification slows down full regressions**: classification covers only true global-state mutations; read-only audits, locally unique data, and files without shared state continue to run in parallel.
- **Heuristic detection produces false positives**: allowlist entries with documented reasons are supported; reviewers must explain why parallel execution is safe.
- **Fixture auto-loading of mock SQL may mask missing prerequisites**: fixtures only handle demo/mock data required by the test and keep SQL idempotent; business installation SQL still goes through plugin lifecycle verification.
- **Cache tests become less strict after accepting both `200` and `304`**: the `200` branch must assert that the new ETag differs from the old ETag and that a response body is present, still verifying protocol correctness.
- **Adding classification and validation requires short-term cleanup of existing cases**: migrate in batches, first covering known conflict files and high-risk modules, then progressively converging on other hits.

## Migration Plan

1. Inventory the current E2E tree, support files, duplicate TC IDs, fixed-wait hotspots, and shared-state risks.
2. Create the target directories and support folders, move helpers/debug files, and repair imports.
3. Move `TC*.ts` files module by module and fix duplicate TC IDs and naming conflicts.
4. Add the execution manifest plus `smoke`, `module`, and `full` entrypoints while preserving the existing full-regression default.
5. Introduce `storageState` generation and the new authenticated fixtures, then validate high-frequency modules outside `auth/`.
6. Apply the first round of state-based wait cleanup, define serial/parallel pools, and run targeted plus full regressions.
7. Extend the execution manifest with machine-readable isolation categories and update the runner to report serial/parallel boundaries.
8. Add high-risk heuristic detection to the validation script and remediate known conflict cases.
9. Consolidate plugin and mock-data prerequisites into idempotent fixtures; adjust cache/ETag assertions to protocol semantics.
10. Document conflict governance rules, record timing and stability baselines, and verify with full regression.

## Open Questions

- No blocking open questions remain. `pnpm test` continues to mean a full regression run, while the fast-feedback entrypoints are additive.
