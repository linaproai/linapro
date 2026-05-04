## 1. Public Configuration Service Interface

- [x] 1.1 Update the public interface of `apps/lina-core/pkg/pluginservice/config` to provide `Get`, `Exists`, `Scan`, basic type reads, and `Duration` reads.
- [x] 1.2 Remove the `MonitorConfig` type alias and the `GetMonitor()` plugin-specific business method so the public component no longer references plugin business configuration structures.
- [x] 1.3 Add Go comments, error handling, and default-value semantics for generic configuration read methods in line with GoFrame v2 and project backend code standards.

## 2. Monitor Server Plugin Migration

- [x] 2.1 Add private configuration loading logic inside the `monitor-server` plugin to maintain the monitor configuration structure, defaults, duration parsing, and business validation.
- [x] 2.2 Migrate `monitor-server` scheduled collection registration and cleanup logic to the new generic configuration service read path.
- [x] 2.3 Search the repository and remove old public interface references such as `GetMonitor()` and `MonitorConfig`.

## 3. Tests and Verification

- [x] 3.1 Add unit tests for `pluginservice/config` covering arbitrary key reads, missing key defaults, struct scanning, basic type reads, successful duration parsing, and failed duration parsing.
- [x] 3.2 Add or update unit tests for `monitor-server` plugin configuration loading covering defaults, overrides, invalid duration values, and business validation.
- [x] 3.3 Run affected Go tests and fix any failures.

## 4. Governance Checks

- [x] 4.1 Confirm this change does not affect frontend pages, menus, routes, buttons, forms, prompts, runtime i18n, manifest i18n, or apidoc i18n resources, and record that conclusion in the implementation result.
- [x] 4.2 Confirm this change only reads static configuration files and does not add runtime mutable caches. If any cache is added, document the authoritative data source, consistency model, invalidation triggers, cross-instance synchronization, and failure fallback.
- [x] 4.3 Invoke `lina-review` after implementation for code and specification review.

## Feedback

- [x] **FB-1**: Reuse the monitor service instance in `monitor-server` scheduled tasks instead of constructing `monitorsvc.New()` on every cron execution.
- [x] **FB-2**: Allow dynamic plugins to read the complete static configuration through the `config` host service.
- [x] **FB-3**: Historical feedback: the dynamic plugin `config` host service methods had been narrowed to only `get`; this was superseded by FB-5.
- [x] **FB-4**: Historical feedback: preserve the richer guest config SDK methods and allow omitted `methods` on the `config` host service to default to `get`; this was superseded by FB-5.
- [x] **FB-5**: Allow dynamic plugins to call all read-only configuration methods through the `config` host service, with omitted `methods` defaulting to the full read-only method set.
