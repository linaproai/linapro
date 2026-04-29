## Context

The current E2E suite already routes certain files into a serial queue via `hack/tests/config/execution-manifest.json` and splits full/smoke/module modes into parallel and serial phases via `run-suite.mjs`. Full regressions still expose a key gap: while test files can be assigned to serial or parallel execution, there is no auditable global-state mutation classification and no governance documentation for runtime caches, plugin lifecycle, system parameters, dictionaries, menu permissions, and other shared state.

This change targets test infrastructure and test governance; it does not alter business APIs or product runtime behavior. Existing E2E naming conventions, module directory boundaries, OpenSpec task tracking rules, and i18n continuous governance requirements must continue to be followed.

## Goals / Non-Goals

**Goals:**

- Define "shared global state" as a classifiable, verifiable, auditable E2E execution attribute.
- Enable runner scripts and validation scripts to detect when high-risk shared-state cases have not entered a serial boundary.
- Ensure plugin, mock data, language-pack cache, system parameter, dictionary, and menu-permission related tests have stable prerequisites and cleanup strategies.
- Refactor cache/ETag tests to validate protocol semantics, accepting that legitimate global version refreshes may return fresh resources during a full regression.
- Produce documented test conflict cases and governance rules for reuse when adding new test cases.

**Non-Goals:**

- Do not rewrite the existing Playwright framework or introduce a new test runner.
- Do not change real business module data models, API contracts, or runtime i18n behavior.
- Do not disable plugin lifecycle, cache invalidation, permission synchronization, or other real business mechanisms for testing convenience.
- Do not require all E2E tests to run serially; parallelism is preserved for files with no global-state contention.

## Decisions

### Decision 1: Introduce global-state classification in the execution manifest

Maintain machine-readable classifications for serial entries or test files in `execution-manifest.json`, such as `pluginLifecycle`, `runtimeI18nCache`, `systemConfig`, `dictionaryData`, `permissionMatrix`, `sharedDatabaseSeed`. The runner script continues to execute serial files as a set, while the validation script additionally checks classification completeness.

The alternative of relying solely on file paths or manual conventions to determine serial boundaries has low cost but cannot explain why a given file must be serial, nor automatically alert when new high-risk cases are added; therefore it is not adopted.

### Decision 2: Supplement manual classification with static heuristic detection

`validate-e2e.mjs` should scan for high-risk APIs, helpers, or keywords such as plugin `install/enable/disable/uninstall/sync`, system parameter writes, dictionary import/modify, menu permission modifications, runtime language-pack cache assertions, `localStorage` ETag caching, and more. When a high-risk pattern is detected but the file is not in the serial boundary or lacks a declared classification, validation should fail with actionable remediation guidance.

Heuristics do not pursue perfect semantic analysis; they serve only as an engineering guard against omissions. Complex false positives may be handled through explicit classification or allowlist entries, but each allowlist entry must document its justification.

### Decision 3: Consolidate plugin and mock-data prerequisites into fixtures

Tests that depend on source-plugin or dynamic-plugin state must prepare plugin state through unified fixtures/helpers. Fixtures must provide idempotent installation, enabling, necessary mock SQL loading, and frontend plugin projection refresh. Test files should not implicitly depend on another file having installed a plugin, nor assume that a specific plugin table or mock data already exists.

This adds a small amount of setup time but yields single-file runnability and full-regression stability.

### Decision 4: Cache tests validate protocol semantics rather than fixed response codes

Runtime i18n ETag, public config cache, plugin frontend resource generation, and similar tests should verify "whether the request carries the correct conditional header" and "whether the server response matches the current resource version." For example, when reloading with a stale ETag, `304` indicates the version has not changed, and `200 + new ETag + body` indicates the version has legitimately refreshed; both are correct.

Only when the test exclusively owns the resource version and there is no concurrent global-state mutation should a fixed `304` be asserted.

### Decision 5: Prefer stable fields for business-state assertions

When testing business counts, states, and permissions, prefer stable API fields, IDs, codes, labelKeys, and permission keys over inferring business state from localized UI text. UI copy should still be tested, but as separate presentation assertions, not as part of cross-language business-state computation.

This avoids language switching, traditional/simplified Chinese switching, or copy adjustments causing false business-test failures.

## Risks / Trade-offs

- [Risk] Overly broad serial classification slows down full regressions → Mitigation: classification covers only true global-state mutations; read-only audits, locally unique data, and files without shared state continue to run in parallel.
- [Risk] Heuristic detection produces false positives → Mitigation: allowlist entries with documented reasons are supported; reviewers must explain why parallel execution is safe.
- [Risk] Fixture auto-loading of mock SQL may mask missing prerequisites → Mitigation: fixtures only handle demo/mock data required by the test and keep SQL idempotent; business installation SQL still goes through plugin lifecycle verification.
- [Risk] Cache tests become less strict after accepting both `200` and `304` → Mitigation: the `200` branch must assert that the new ETag differs from the old ETag and that a response body is present, still verifying protocol correctness.
- [Risk] Adding classification and validation requires short-term cleanup of existing cases → Mitigation: migrate in batches, first covering known conflict files and high-risk modules, then progressively converging on other hits.
