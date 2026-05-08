## Why

LinaPro is positioned as an AI-driven full-stack development framework. Internationalization must be framework infrastructure, not something every delivered project rebuilds from scratch. The frontend already had static locale bundle support through `vue-i18n`, but the host lacked project-level i18n capability: menus, dictionaries, system parameter metadata, plugin manifests, backend errors, import/export content, and many backend-returned labels were still stored and returned as single-language text. Delivery teams had to edit copy manually in many places, which could not reliably support multilingual project delivery.

Beyond the missing foundation, several systemic issues emerged across subsequent iterations:

1. **Performance**: The `Translate` hot path cloned the entire runtime message bundle on every call; cache invalidation cleared all languages and all sectors at once; the runtime translation bundle API had no ETag/version stamp, causing full retransmission on every frontend language switch.
2. **Consistency**: Business modules independently decided "when to translate / when to skip / which Translate* to use"; the `Service` interface carried 18 methods while business modules only needed a few; `sysconfig_i18n.go` hardcoded English/Chinese maps in Go; source-text namespace ownership leaked into the i18n package.
3. **Boundary**: `apidoc` and runtime bundle each maintained duplicate resource loaders; WASM custom section parsing was duplicated inside the i18n package; frontend `loadMessages` used `Promise.all` for three things with different failure semantics.
4. **Message governance**: Backend and plugin logic still contained large amounts of direct Chinese returns, mixed Chinese-English strings, or raw backend text being passed through -- in error messages, import failure reasons, Excel exports, plugin bridging errors, and frontend page labels.
5. **Hardcoded Chinese**: Handwritten non-test Go files contained Chinese string literals that could reach HTTP responses, plugin responses, export files, management UI projections, or runtime configuration display.
6. **UI gaps**: Manual regression found gaps in English localization, default seed display, dynamic plugin permission mounting, table and form layout, and built-in protection.
7. **Project positioning**: Repository documentation and system metadata still described LinaPro as a "backend management system" rather than an AI-driven full-stack development framework. The `lina-core` boundary between core host capabilities and workspace adaptation was not explicitly defined. README files lacked unified internationalization rules.

## What Changes

### I18n Infrastructure
- Establish a three-layer i18n model for static UI copy, dynamic metadata, and business content.
- Use a "file baseline plus database override" resource governance model (later simplified to file-only single source of truth).
- Add backend request-level locale resolution (`lang` query parameter > `Accept-Language` header > system default).
- Define stable translation-key conventions derived from business keys (`menu.<key>.title`, `dict.<type>.name`, etc.).
- Provide runtime message bundle and locale list APIs with aggregated message resources.
- Add i18n resource loading rules for plugins through `manifest/i18n/<locale>/` directories.
- Provide import/export, missing translation checks, and resource source diagnostics.

### Performance Optimization
- Rewrite `Translate`/`TranslateSourceText`/`TranslateOrKey`/`TranslateWithDefaultLocale` hot paths to read directly from cache instead of cloning.
- Refactor `runtimeBundleCache` into a layered structure by locale + sector (host/source-plugin/dynamic-plugin).
- Runtime translation bundle API outputs `ETag` and supports `If-None-Match` 304 negotiation.
- Frontend persists runtime translations to `localStorage` with 7-day TTL for zero-network language switching.
- Split frontend `loadMessages` by failure semantics: runtime bundle failure -> persistent fallback; public config failure -> fire-and-forget; third-party library locale -> must await.

### Interface and Boundary Improvements
- Split the `i18n.Service` large interface into `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer` four smaller interfaces.
- Extract `pkg/i18nresource` shared `ResourceLoader` used by both runtime bundle and apidoc loading.
- Move WASM custom section parsing to `pkg/pluginbridge`.
- Introduce `RegisterSourceTextNamespace` explicit registration for code-owned namespaces.
- Converge projection rules within business module boundaries; prohibit the i18n foundation service from reverse-perceiving business entities.
- Remove `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` runtime persistence tables; converge to JSON/YAML resources as single source of truth.
- Introduce Traditional Chinese (`zh-TW`) as a stress test third language; fix document direction to LTR.

### Message and Error Governance
- Establish a structured backend error model through `bizerr` with stable error codes, translation keys, English source messages, parameters, and GoFrame type codes.
- Classify runtime messages into six categories: `UserMessage`, `UserArtifact`, `UserProjection`, `DeveloperDiagnostic`, `OpsLog`, `UserData`.
- Define unified localization helpers for Excel exports, import templates, and import failure reasons.
- Define unified error return contracts for plugin bridging, host service calls, and plugin lifecycle results.
- Add automated scanning and test gates for hardcoded messages.

### Backend Hardcoded Chinese Cleanup
- Classify all Chinese string findings in backend Go source by category.
- Replace caller-visible Chinese errors with module-owned `bizerr` codes and runtime i18n resources.
- Localize backend-owned projections and deliverables through runtime i18n or structured fields.
- Convert plugin-platform developer diagnostics to stable English text.
- Govern generated schema text at SQL comments or generation inputs.

### UI and Content Localization
- Remove remaining Chinese copy from framework-delivered English pages and seed displays.
- Improve English layout for tables, forms, and search areas with long English labels.
- Mount dynamic plugin route permission buttons under the owning plugin menu.
- Add confirmation for scheduled-job Run Now.
- Converge default workbench to real navigation entries and operational semantics.
- Block plugin governance writes when demo-control is enabled.
- Protect built-in dictionaries and system parameters from deletion while keeping them editable.
- Align role display between user management and role management.

### Project Governance and Documentation
- Unify LinaPro project positioning as "AI-driven full-stack development framework".
- Establish lina-core as the core host service, separate from the default management workspace.
- Establish full-repository README bilingual governance: English `README.md` + Chinese `README.zh_CN.md`.
- Add explicit confirmation guards for database `init`/`mock` commands.
- Standardize plugin installation review with a unified detail dialog and authorization snapshot.

## Capabilities

### New Capabilities
- `i18n-infrastructure`: Locale resolution, translation resource aggregation, runtime message bundle distribution, three-layer model, performance optimization (zero-copy hot path, layered cache, ETag/304), Service interface split, ResourceLoader, source-text namespace registration, Traditional Chinese support, fixed LTR direction, file-only single source of truth.
- `project-positioning-governance`: Unified project positioning as AI-driven full-stack development framework, system metadata and user-visible copy alignment.
- `readme-localization-governance`: Full-repository README bilingual mirror governance.
- `core-host-boundary-governance`: Core host boundary and workspace adaptation interface classification.
- `database-bootstrap-commands`: Explicit confirmation guards and first-error-stops semantics for database init/mock commands.
- `message-governance`: Runtime message classification, structured error model through `bizerr`, import/export localization, plugin error contracts, automated scanning, backend hardcoded Chinese cleanup.
- `demo-control-guard`: Demo-control plugin blocks plugin governance writes when enabled.

### Modified Capabilities
- `menu-management`: Localized menu titles from stable `menu_key`, button permissions as short action words, dynamic plugin button mounting under owning plugin menu, clickable menu tree rows.
- `dict-management`: Localized dictionary names and labels, tag style dropdown with readable options, built-in dictionary types editable but not deletable.
- `config-management`: Localized config metadata, import/export headers via translation keys, public frontend config i18n, built-in system parameters editable but not deletable.
- `login-page-presentation`: Localized login page title/description/subtitle from host public config, language-switch refresh.
- `system-info`: Localized project description and component descriptions, unified project positioning across languages.
- `system-api-docs`: English source copy in DTOs, independent apidoc i18n resources, OpenAPI metadata aligned with project positioning.
- `plugin-manifest-lifecycle`: Plugin i18n resource declaration and lifecycle management, unified install review dialog, authorization snapshot reuse.
- `plugin-ui-integration`: Plugin pages in host locale context, multiple host integration modes, dynamic routing, hot-upgrade flows.
- `plugin-runtime-loading`: WASM custom section parsing in pluginbridge, shared ResourceLoader.
- `plugin-permission-governance`: Plugin resource interfaces checked by plugin resource permissions.
- `role-management`: Built-in protected role localization, consistent role display across pages.
- `cron-job-management`: Localized built-in job metadata, trigger confirmation.
- `dashboard-workbench`: Runtime i18n for workbench copy, real navigation entries, metric semantics, theme preference, dark-mode logo.
- `base-layout`: Default management workspace semantics instead of "backend management system".

## Impact

- **Backend capabilities**: Affects request context, shared middleware, `bizerr` error model, config/menu/dictionary/plugin/system information/cron/role services, the i18n resource model, runtime message bundle APIs, import/export localization, plugin platform error contracts, and automated scanning in `apps/lina-core`.
- **Database model**: Initially added locale/translation/content tables (later removed in favor of file-only resources); affects seed/mock data localization for all built-in modules.
- **Frontend capabilities**: Affects `vue-i18n` initialization, runtime message loading with ETag/persistent cache, language switch refresh, public frontend config sync, dynamic menu refresh, request interceptor error handling, English layout adaptation, and plugin page integration in `apps/lina-vben`.
- **Plugin ecosystem**: Affects resource organization around `plugin.yaml` in `apps/lina-plugins`, plugin translation resource directories, plugin lifecycle management, install review dialogs, and plugin error contracts.
- **Documentation and governance**: Affects `CLAUDE.md`, repository README files, OpenAPI metadata, system information page, and scanning tooling under `hack/tools/`.
