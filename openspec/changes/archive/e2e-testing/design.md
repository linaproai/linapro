## Context

The current E2E execution chain already has a good governance foundation: `run-suite.mjs` supports full, host, smoke, and module modes; `execution-manifest.json` declares module scope, serial isolation files, and isolation categories; GitHub Actions already wraps browser E2E as `reusable-e2e-tests.yml`. Therefore, this optimization should prioritize reusing the existing runner and manifest rather than introducing a new test framework or rewriting Playwright organization.

From GitHub Actions logs, host-only mode is approximately 36 minutes, with Playwright tests approximately 25 minutes, mainly accumulated from a large number of 5-10 second UI cases. Plugin-full mode is approximately 2 hours, with Playwright tests approximately 112 minutes, mainly due to plugin lifecycle tests, plugin-specific functionality tests, and plugin-sensitive host regression tests executing serially in a single job.

## Goals / Non-Goals

**Goals:**

- Reduce plugin-full E2E wall clock to within 45 minutes as priority, while maintaining official plugin coverage.
- Reduce host-only E2E wall clock to within 30 minutes as priority, while preserving existing host functionality coverage.
- Reduce repeated precondition costs for ordinary plugin page tests.
- Preserve isolation semantics for plugin lifecycle, permission matrix, runtime i18n, shared database seed, and filesystem artifact high-risk tests.
- Allow optimization results to be reviewable through CI logs, Playwright reports, and test timing records.

**Non-Goals:**

- Do not change production API, database schema, frontend user functionality, or plugin runtime semantics.
- Do not remove critical E2E coverage for speed; when plugin-full scope needs narrowing, specific plugin behaviors must be recovered to their owning plugin directory for maintenance.
- Do not introduce new external test services or new CI platforms.
- Do not adjust product copy or i18n resources in this change.

## Decisions

### Plugin-full first uses generic module scope for CI sharding

Use existing runner capability to split plugin-full job, prioritizing shards:

- `extension:plugin`
- `plugins`

Each shard continues to use `make dev plugins=1` and official plugin submodule checkout, with artifact names containing shard names. `extension:plugin` only selects host plugin framework, dynamic test plugin, and generic plugin governance cases from the root directory that do not depend on specific official plugins; `plugins` serves as the generic entry for all source plugin owned tests. `plugin:<plugin-id>` still serves as the generic entry for selecting a single source plugin's own tests during local or temporary CI debugging, but routine CI no longer enumerates specific official plugin IDs. This minimizes runner changes, lets source plugin owned cases execute under a unified entry, and avoids manifest maintaining long-term aliases by official plugin business modules.

Alternative is to increase the Playwright worker count per job, but many plugin-full files in the current manifest are serialized due to shared state, and directly increasing workers has limited benefit for wall clock while being more prone to state pollution.

### Plugin-full and host-only maintain responsibility distinction

Host-only continues to cover host full capability; plugin-full focuses on root directory generic plugin framework tests, dynamic test plugin runtime, and each official source plugin directory's self-contained own E2E. Root `hack/tests` must not depend on any specific official plugin ID, path, menu, mock data, i18n key, or page locator; any case needing to verify specific official plugin behavior must move to `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` and run through `plugin:<plugin-id>`.

When wanting to run a host module in a main framework environment without `apps/lina-plugins`, use `pnpm test:host:module -- <scope>`. This entry reuses the host-only exclusion list, only executing host cases in the specified scope that do not depend on official plugin workspace.

Alternative is to let plugin-full continue executing `pnpm test` full volume with all modules split. This reduces wall clock but significantly increases runner minutes and continues confusing host-only and plugin-full verification responsibilities.

### New authenticated page fixture without auto-navigation

Preserve existing `adminPage` for backward compatibility while adding a lightweight fixture, such as `authenticatedPage`. This fixture only creates a page with admin storage state, without defaulting to `/dashboard/analytics`. When migrating high-time-cost files, test cases directly navigate to the target business route, avoiding repeated dashboard first-screen loading and subsequent business page navigation costs.

Alternative is to directly modify `adminPage` behavior, but this affects many existing tests with higher risk.

### Plugin baseline prepared idempotently at suite or shard level

Add shared helper capability allowing ordinary plugin page tests to declare needed plugin collections, executing once at suite or shard level:

- `syncPlugins`
- install/enable required plugins
- load plugin mock data when present
- refresh plugin projection

Lifecycle tests continue to self-control installation, enablement, disablement, and uninstallation, avoiding baseline interference with tested state.

Alternative is to continue `beforeEach ensureSourcePluginEnabled` in each test file, which is simple but costly, and repeatedly refreshing plugin projection will continue amplifying plugin-full runtime.

### Lifecycle heavy users use representative full chain plus batch contract smoke

Official plugin lifecycle tests should not run complete UI installation, enablement, disablement, uninstallation, route missing, and route recovery for every plugin. Retain one representative plugin for complete UI lifecycle, while other official plugins use API lifecycle, menu mounting, and page accessibility smoke verification.

Dynamic runtime tests split into core lifecycle and demo functionality verification. Assertions not needing real browser interaction such as upload size, API bridge, and data retention prioritize API/request layer verification; scenarios needing to verify host shell, iframe, new tab, and embedded runtime retain UI coverage.

## Risks / Trade-offs

- [Risk] CI sharding increases total runner minutes and artifact count. -> Start with limited shards, prioritizing plugin-related scope; artifact names include shard name for easy failure location.
- [Risk] Each shard independently initializes database and service, potentially exposing hidden cross-file dependencies. -> Maintain test file independence requirements; on failure prioritize fixing fixture preconditions rather than restoring cross-file dependencies.
- [Risk] New fixture coexisting with old `adminPage` may cause selection confusion. -> Only migrate high-time-cost files initially, with clear usage scenarios in fixture comments and task records.
- [Risk] Plugin baseline may mask lifecycle test expected initial states. -> Baseline only used for ordinary plugin functionality tests; lifecycle directories continue explicit cleanup and assembly.
- [Risk] Narrowing plugin-full scope may miss host and plugin framework regression. -> Root directory only retains generic plugin framework and dynamic test plugin coverage; specific official plugin behaviors must be closed-loop in corresponding plugin directories, selectable through `plugins` or `plugin:<plugin-id>`.
