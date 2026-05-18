## 1. I18n Infrastructure and Data Model

- [x] 1.1 Design and implement core tables and seed data such as `sys_i18n_locale`, `sys_i18n_message`, and `sys_i18n_content` (later removed in favor of file-only resources)
- [x] 1.2 Add locale resolution, request-context locale injection, and a unified translation service in `apps/lina-core`
- [x] 1.3 Establish translation resource aggregation for host, project, plugin, and database override sources (later simplified to file-only single source of truth)
- [x] 1.4 Define stable translation-key conventions derived from business keys
- [x] 1.5 Provide runtime message bundle and locale list APIs, supporting aggregated message bundles by locale
- [x] 1.6 Provide i18n message export, missing translation checks, and override source diagnostics
- [x] 1.7 Remove `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` runtime persistence tables, converging to JSON/YAML resources as single source of truth

## 2. Performance Optimization: Translation Hot Path and Cache Layering

- [x] 2.1 Rewrite `Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` to hold read lock and read cache directly, removing `cloneFlatMessageMap` calls
- [x] 2.2 Preserve clone semantics for `BuildRuntimeMessages` and `ExportMessages`; add internal `lookupBundleKey` utility method
- [x] 2.3 Refactor `runtimeBundleCache` to layered structure by `locale x sector (host / source-plugin / dynamic-plugin)`, adding `mergedView` and `bundleVersion` atomic counter
- [x] 2.4 Refactor `InvalidateRuntimeBundleCache` to accept `InvalidateScope` parameter for fine-grained invalidation by locale, sector, and plugin ID
- [x] 2.5 Add `Translate` single/batch call benchmark tests; verify single call < 100ns on cache hit
- [x] 2.6 Add layered invalidation unit tests covering host resource, plugin enable/disable, and source plugin registration scenarios

## 3. Performance Optimization: Runtime Translation Bundle ETag Negotiation

- [x] 3.1 Add `BundleVersion()` method returning the current runtime translation bundle version
- [x] 3.2 Modify runtime messages controller to output `ETag` and `Cache-Control` headers
- [x] 3.3 Implement `If-None-Match` negotiation, returning 304 Not Modified when matched
- [x] 3.4 Add unit tests covering ETag output, 304 response, and ETag version change behavior

## 4. Performance Optimization: Frontend RequestClient and Persistent Cache

- [x] 4.1 Rewrite `runtime-i18n.ts`: replace raw `fetch` with `requestClient`, add Bearer injection, error degradation, retry chain
- [x] 4.2 Add `localStorage` persistence layer with `linapro:i18n:runtime:<locale>` key, TTL 7 days
- [x] 4.3 Implement "persistent hit renders immediately, background If-None-Match negotiation, 304 does not update" fast path
- [x] 4.4 Refactor `loadMessages` to split by failure semantics: runtime bundle -> persistent fallback; public config -> fire-and-forget; third-party locale -> must await
- [x] 4.5 Add unit tests covering persistent hit, TTL expiration, 304 path, and network error degradation

## 5. Service Interface Split and Module Boundaries

- [x] 5.1 Split `Service` interface into `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer` four small interfaces
- [x] 5.2 Converge `i18nSvc` field types in menu / dict / sysconfig / jobmgmt / role / usermsg / apidoc / plugin modules to minimum dependency interfaces
- [x] 5.3 Delete centralized `LocaleProjector` approach; retain projection decisions in each business module's own `*_i18n.go`
- [x] 5.4 Refactor `menu_i18n.go`, `dict_i18n.go`, `sysconfig_i18n.go`, `jobmgmt_i18n.go`, `role.go`, plugin runtime to own projection rules within module boundaries
- [x] 5.5 Delete `englishLabels` / `chineseLabels` Go maps in `sysconfig_i18n.go`, replacing with `config.field.<name>` translation keys
- [x] 5.6 Create `RegisterSourceTextNamespace` in `i18n` package; delete `isSourceTextBackedRuntimeKey` blacklist
- [x] 5.7 Add `init()` in `jobmgmt` package to register `job.handler.` and `job.group.default.` namespaces
- [x] 5.8 Add review rule: prohibit business modules from cloning runtime message bundles; `InvalidateRuntimeBundleCache` must receive explicit scope

## 6. Boundary Cleanup: ResourceLoader and WASM Parsing

- [x] 6.1 Create `ResourceLoader` in `pkg/i18nresource/` accepting `Subdir` / `LocaleSubdir` / `PluginScope` / `LayoutMode` configuration
- [x] 6.2 Implement `LoadHostBundle`, `LoadSourcePluginBundles`, `LoadDynamicPluginBundles` methods
- [x] 6.3 Refactor i18n service to use `i18nresource.ResourceLoader` replacing duplicate implementations
- [x] 6.4 Refactor apidoc loader to use `i18nresource.ResourceLoader` with `RestrictedToPluginNamespace` configuration
- [x] 6.5 Create `ReadCustomSection` and `ListCustomSections` in `pkg/pluginbridge/pluginbridge_wasm_section.go`
- [x] 6.6 Delete `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` from i18n package, replace with `pluginbridge.ReadCustomSection`
- [x] 6.7 Adjust dynamic plugin apidoc resource loading to use `pluginbridge.ReadCustomSection`

## 7. Traditional Chinese Integration and Fixed LTR

- [x] 7.1 Auto-discover built-in languages from `manifest/i18n/<locale>/*.json`; maintain default language, sorting, native names in default config file `i18n` section
- [x] 7.2 Create `manifest/i18n/zh-TW/*.json` for host and all source plugins
- [x] 7.3 Create `manifest/i18n/zh-TW/apidoc/**/*.json` for host and all source plugins
- [x] 7.4 Create frontend static language packs for `zh-TW`
- [x] 7.5 Derive dayjs / antd / vxe locale by language code convention instead of switch branches
- [x] 7.6 Fix `<html dir>` and `ConfigProvider.direction` to `ltr`
- [x] 7.7 Run `CheckMissingMessages(locale='zh-TW')` confirming `total=0`
- [x] 7.8 Add E2E test cases covering Traditional Chinese language switching, `<html dir>` assertion, and key page text completeness

## 8. Dynamic Metadata I18n

- [x] 8.1 Update menu capability to return localized menu titles from stable `menu_key` values
- [x] 8.2 Update dictionary capability to return localized dictionary type names and labels
- [x] 8.3 Update config capability to return localized config names, remarks, and public frontend config
- [x] 8.4 Update system information capability to return localized project introduction and component descriptions
- [x] 8.5 Localize built-in protected role display names in role management APIs
- [x] 8.6 Localize built-in cron job, job group, and execution log display metadata

## 9. Default Workspace I18n Flow

- [x] 9.1 Extend `vue-i18n` loading flow to merge local static bundles, host runtime bundles, and plugin bundles
- [x] 9.2 Refresh public frontend config, dynamic menus, routes, and pages when the language changes
- [x] 9.3 Update frontend request interceptor to prioritize `messageKey/messageParams` for error display
- [x] 9.4 Clean up server monitoring page, online users page, and plugin frontend pages for `$t` usage
- [x] 9.5 Add missing `zh-CN`, `en-US`, `zh-TW` translation keys in frontend static and runtime language packs

## 10. Plugin I18n Integration Contract

- [x] 10.1 Define plugin `manifest/i18n/<locale>/` locale directory convention and host loading/removal rules
- [x] 10.2 Update plugin lifecycle flows so installation, upgrade, enablement, disablement, and uninstallation maintain plugin translation resource snapshots
- [x] 10.3 Update plugin page integration so host-embedded plugin pages participate in locale context and runtime message refresh
- [x] 10.4 Update plugin manifest and lifecycle to automatically cover new languages without modifying host or plugin code

## 11. Structured Error Infrastructure

- [x] 11.1 Add `bizerr` runtime message error model and construction helper with `errorCode`, `messageKey`, `messageParams`, English fallback, and `gcode` semantics
- [x] 11.2 Update unified response middleware to recognize structured errors and output localized `message`, stable `errorCode`, `messageKey`, and `messageParams`
- [x] 11.3 Define business error codes in module-specific `*_code.go` files
- [x] 11.4 Preserve `LocalizeError` as fallback for legacy unstructured errors
- [x] 11.5 Add unit tests for structured error rendering across `zh-CN`, `en-US`, `zh-TW`

## 12. Backend Hardcoded Chinese Cleanup

- [x] 12.1 Audit backend Chinese string findings and classify by category (caller-visible, user-visible, deliverable, developer diagnostic, generated, test, user-data)
- [x] 12.2 Replace caller-visible Chinese backend and plugin errors with module-owned structured `bizerr` codes
- [x] 12.3 Localize backend-owned projections and exported deliverables through runtime i18n or structured fields
- [x] 12.4 Convert plugin-platform developer diagnostics to stable English text and wrap boundary errors structurally
- [x] 12.5 Govern generated schema descriptions through SQL comments or generation inputs
- [x] 12.6 Clean up Chinese hardcoding in CLI and database initialization diagnostic errors

## 13. Import/Export and Plugin Platform Cleanup

- [x] 13.1 Clean up user module import/export headers, templates, failure reasons, and enum text
- [x] 13.2 Clean up dictionary type, data, combined export import/export content
- [x] 13.3 Clean up system parameters, config import/export, file management errors and export content
- [x] 13.4 Clean up scheduled tasks, task handlers, task logs, and task metadata errors
- [x] 13.5 Clean up plugin lifecycle, source plugin upgrades, dynamic plugin runtime errors
- [x] 13.6 Implement request-level batch localization context for import/export
- [x] 13.7 Clean up mixed Chinese-English errors in plugin bridge, filesystem, database, WASM host service packages

## 14. UI and Content Localization

- [x] 14.1 Close residual Chinese text in English mode for dashboard, shared shell surfaces, access management, organization center, system settings, content management, system monitoring, scheduler center, and extension center
- [x] 14.2 Fix label wrapping in high-frequency forms and drawers for English
- [x] 14.3 Fix English table header wrapping and compressed fixed action columns
- [x] 14.4 Mount dynamic plugin route permission buttons under owning plugin menu
- [x] 14.5 Improve menu tree expand interactions with clickable pointer affordance
- [x] 14.6 Localize generated Unassigned department nodes and built-in config display
- [x] 14.7 Add confirmation for scheduled-job Run Now action
- [x] 14.8 Align role display between user management and role management
- [x] 14.9 Protect built-in dictionaries and system parameters from deletion
- [x] 14.10 Adjust pagination page-size selector width for English `items/page`

## 15. Workbench and Demo Control

- [x] 15.1 Converge default workbench to real navigation entries and operational semantics
- [x] 15.2 Optimize analysis page metric grouping, time range switching, and chart title semantics
- [x] 15.3 Add dark-mode logo subtle cyan edge glow
- [x] 15.4 Preserve user-explicit theme preference when synchronizing public frontend config
- [x] 15.5 Block plugin governance writes when demo-control is enabled
- [x] 15.6 Add E2E coverage for workbench English i18n, role/user seed display, dynamic menu permissions

## 16. Project Governance and Documentation

- [x] 16.1 Update `CLAUDE.md` and active specs to unify project positioning as "AI-driven full-stack development framework"
- [x] 16.2 Establish lina-core core host boundary as top-level requirement and classify workspace adaptation interfaces
- [x] 16.3 Update OpenAPI titles, descriptions, system info page, login page, and script banners
- [x] 16.4 Create repository root English `README.md` and Chinese mirror `README.zh-CN.md`
- [x] 16.5 Inventory existing subdirectory READMEs and progressively add bilingual mirrors
- [x] 16.6 Update `openspec/config.yaml` and active specs for new project positioning constraints

## 17. Database Bootstrap Safety

- [x] 17.1 Update root `Makefile` init/mock targets to require explicit `confirm` variable
- [x] 17.2 Update `apps/lina-core/Makefile` init/mock targets with same confirmation logic
- [x] 17.3 Add `confirm` parameter and validation to backend init/mock commands
- [x] 17.4 Adjust common SQL execution to first-error-stops and return failure status
- [x] 17.5 Update command help text and error messages with correct usage examples
- [x] 17.6 Add unit tests covering confirmation guard scenarios

## 18. Plugin Installation Governance

- [x] 18.1 Unify source and dynamic plugin install review into a single detail dialog
- [x] 18.2 Dynamic plugin authorization snapshot formed at install time, reused on subsequent enables
- [x] 18.3 Plugin resource interfaces checked by plugin resource permissions
- [x] 18.4 Source plugin example pages with install/uninstall SQL and lifecycle verification
- [x] 18.5 Dynamic plugin example pages with install/uninstall SQL and lifecycle verification

## 19. Automated Governance and Testing

- [x] 19.1 Add backend hardcoded runtime message scanning script
- [x] 19.2 Add frontend hardcoded runtime message scanning script or ESLint rules
- [x] 19.3 Integrate scanning commands into local validation entry points
- [x] 19.4 Add missing translation checks for all enabled built-in languages
- [x] 19.5 Add review rules prohibiting hardcoded Go label maps, business entity projectors in i18n package, and full `Service` interface declarations
- [x] 19.6 Run full E2E test suite including Traditional Chinese, ETag negotiation, structured error localization, and English layout regression
- [x] 19.7 Update host and frontend i18n README documentation for new language process, ETag, and scanning governance

## 20. Documentation and Review

- [x] 20.1 Update `apps/lina-core/manifest/i18n/README.md` and Chinese mirror for ETag negotiation, new language process, and source namespace registration
- [x] 20.2 Update `apps/lina-vben/apps/web-antd/src/locales/README.md` and Chinese mirror for fixed LTR direction and persistent cache strategy
- [x] 20.3 Update root `CLAUDE.md` "i18n continuous governance requirements" with new rules
- [x] 20.4 Call `lina-review` to complete code and specification review across all change groups
- [x] 20.5 Run `make test` full E2E passes confirming all regression coverage
