## 1. I18n Infrastructure and Data Model

- [x] 1.1 Design and implement core tables and seed data such as `sys_i18n_locale`, `sys_i18n_message`, and `sys_i18n_content`.
- [x] 1.2 Add locale resolution, request-context locale injection, and a unified translation service in `apps/lina-core`.
- [x] 1.3 Establish translation resource aggregation for host, project, plugin, and database override sources, and define cache/invalidation behavior.

## 2. Host Runtime APIs and Maintenance Capabilities

- [x] 2.1 Provide runtime message bundle and locale list APIs, supporting aggregated message bundles by locale.
- [x] 2.2 Provide i18n message import/export, missing translation checks, and override source diagnostics.
- [x] 2.3 Require new user-visible error messages, validation messages, and system prompts to prefer translation keys, and add baseline tests.

## 3. Dynamic Metadata I18n

- [x] 3.1 Update menu capability to return localized menu titles, parent names, and route titles from stable `menu_key` values.
- [x] 3.2 Update dictionary capability to return dictionary type names, dictionary labels, and related descriptions according to the current language.
- [x] 3.3 Update config capability to return localized config names, remarks, and brand/login-page copy from public frontend config.
- [x] 3.4 Update system information capability to return localized project introduction, component descriptions, and related display copy.

## 4. Default Workspace I18n Flow

- [x] 4.1 Extend the `apps/lina-vben` `vue-i18n` loading flow to merge local static bundles, host runtime bundles, and plugin bundles.
- [x] 4.2 Refresh public frontend config, dynamic menus, routes, and pages depending on dynamic metadata when the language changes.
- [x] 4.3 Add frontend regression coverage for key pages such as login, system information, menu rendering, and dictionary display.

## 5. Plugin I18n Integration Contract

- [x] 5.1 Define the plugin `manifest/i18n/<locale>/` locale directory convention and host loading/removal rules.
- [x] 5.2 Update plugin lifecycle flows so installation, upgrade, enablement, disablement, and uninstallation maintain plugin translation resource snapshots.
- [x] 5.3 Update plugin page integration so host-embedded plugin pages participate in locale context and runtime message refresh.

## 6. Multilingual Business Content and Delivery Governance

- [x] 6.1 Define the common business-module integration contract for `sys_i18n_content`, including anchor fields, fallback strategy, and cache rules.
- [x] 6.2 Provide delivered-project multilingual resource templates, key naming rules, and missing translation validation rules.
- [x] 6.3 Add backend unit tests, frontend tests, and necessary E2E coverage for language switching, plugin lifecycle changes, and dynamic metadata localization.

## Feedback

- [x] **FB-1**: Use `page-i18n-audit.md` to close residual Chinese text in English mode for dashboard and shared shell surfaces, including statistic cards, workspace demo copy, default current-user profile display, route titles, and tab-title refresh.
- [x] **FB-2**: Close residual Chinese text in English mode for the access management module, covering user, role, and menu page search fields, table headers, action buttons, drawers/authorization modals, and host default role/menu governance data.
- [x] **FB-3**: Close residual Chinese text in English mode for the organization center, covering department/post pages, tree selectors, drawer forms, plugin menus, and default organization/post seed data.
- [x] **FB-4**: Close residual Chinese text in English mode for system settings, covering dictionary, parameter, and file pages plus import/export/upload modals, shared components, and default config/file-scene data.
- [x] **FB-5**: Close residual Chinese text in English mode for content management, covering notice lists, edit/preview modals, notice type/status, and default notice demo content.
- [x] **FB-6**: Close residual Chinese text in English mode for system monitoring, covering online users, service monitoring, operation logs, login logs, built-in plugin pages, detail modals, and default log display content.
- [x] **FB-7**: Close residual Chinese text in English mode for the scheduler center, covering side menus, tab titles, job/group/execution-log pages, job forms, and default job/group/log seed data.
- [x] **FB-8**: Close residual Chinese text in English mode for extension center, development center, and dynamic plugin example pages, covering plugin management, shared system information title flows, dynamic plugin example pages, and plugin manifest/install SQL default content.
- [x] **FB-9**: Close residual Chinese text in English mode for host shared components and non-menu entry pages, covering export confirmation, tree select, upload/cropper, profile center, security settings, authentication pages, and other reused components.
- [x] **FB-10**: Establish a unified localization completion strategy for host and built-in plugin delivered `seed/mock/demo` data, filling missing `en-US` resources, default content translations, and regression cases so English sweeps no longer show framework-provided Chinese content.
- [x] **FB-11**: Stop applying frontend English projection to user-editable business fields such as `sys_user`; keep nicknames, names, creators, and similar database values unchanged, and add `TC0109` regression coverage.
- [x] **FB-12**: Remove i18n projection from editable business data such as departments, posts, roles, menus, dictionaries, parameters, notices, and scheduler data, keeping database values in management pages and selectors, and add `TC0110` regression coverage.
- [x] **FB-13**: Improve dynamic route permission synthetic menu display text on the role creation page so raw permission strings are not exposed, and add dynamic plugin permission readability regression coverage.
- [x] **FB-14**: Fix source-plugin pages for departments, posts, notices, online users, login logs, and operation logs that directly displayed untranslated i18n keys, and add `TC0111` regression coverage.
- [x] **FB-15**: Use Playwright to sweep the default workspace in English, record display issues for search labels, table headers, form labels, and shared navigation with long English copy, and save screenshots plus follow-up suggestions.
- [x] **FB-16**: Establish a shared English layout adaptation baseline for side navigation, tabs, and CRUD search areas so long English copy avoids menu truncation and frequent search-label wrapping.
- [x] **FB-17**: Fix label wrapping in high-frequency forms and drawers in English for profile center, users, departments, posts, parameters, and plugin upload, and standardize English label widths and shorter copy where needed.
- [x] **FB-18**: Fix English table header wrapping and compressed fixed action columns in user, role, menu, dictionary, file, monitoring log, scheduler center, and plugin management lists, adding header overflow handling and short English headers where needed.
- [x] **FB-19**: Add English layout regression case `TC0112`, covering shared navigation, profile center, user/dictionary/scheduler pages, and single-line constraints for search labels, form labels, and table headers.
- [x] **FB-20**: Adjust the profile center password-change form input width in English so long placeholder copy is fully visible at the default desktop viewport, and add matching layout regression assertions.
- [x] **FB-21**: Narrow the profile center password-change input width while still fully showing English placeholders, and add placeholder/width sweep assertions for other profile center English inputs.
- [x] **FB-22**: Unify form container width and input width between profile basic information and password-change tabs so English form lengths and placeholder display stay consistent.
- [x] **FB-23**: Fix system parameter list, import template, and export flows so config metadata i18n is fully applied, while `key`, `value`, and edit backfill keep original governance semantics.
- [x] **FB-24**: Clarify and complete multilingual projection boundaries for public frontend config copy, ensuring login/brand APIs return localized copy in non-default languages while parameter detail/edit views continue to show raw config values with regression coverage.
- [x] **FB-25**: Add an English display exception for the built-in protected super administrator role on the role management page, localizing only the protected `admin` role in English lists while other editable roles keep database values, with dedicated regression coverage.
- [x] **FB-26**: Fix menu management list and related tree selectors that still displayed Chinese menu names in English, ensuring menu trees, parent names, and role menu trees return localized titles with English regression coverage.
- [x] **FB-27**: Improve default admin menu names in English with shorter, natural titles instead of literal translations, and sync frontend static routes, host/plugin runtime locale bundles, and menu regression assertions.
- [x] **FB-28**: Update i18n unit tests that still expected old English menu titles so runtime message tests match the current `Dashboard` baseline.
- [x] **FB-29**: Add permission-audit exemption for the runtime locale list API so the public language-switch startup API satisfies static API permission audit rules.
- [x] **FB-30**: Sync the preferences default side navigation width snapshot so frontend unit tests match the current English layout baseline.
- [x] **FB-31**: Stabilize source-plugin enable/disable flow in system API docs E2E to avoid global plugin state and menu refresh effects on parallel cases.
- [x] **FB-32**: Fix plugin page English static copy sweep assertions that still expected old long labels so E2E matches the current short English label baseline.
- [x] **FB-33**: Fix the default analysis page E2E that still depended on the old duplicated chart-title count so assertions match current key chart cards.
- [x] **FB-34**: Stabilize the language-switch E2E helper local-storage polling to avoid occasional execution-context destruction during navigation.
- [x] **FB-35**: Fill the version field type in the shared plugin E2E fixture so TypeScript validation covers source-plugin upgrade cases.
- [x] **FB-36**: Extend the plugin E2E page object sidebar menu locator range so top-level dynamic plugin menus can be clicked reliably in English.
- [x] **FB-37**: Move English runtime page sweeps that enable/disable dynamic plugins into serial E2E and avoid sidebar locator false matches against top navigation.
- [x] **FB-38**: Stabilize plugin lifecycle E2E sidebar menu, plugin header, and dynamic upload modal locators to avoid language baseline and page-content false matches in serial cases.
- [x] **FB-39**: Fix scheduler job list enum fields that exposed raw i18n keys, ensuring scope and concurrency policy display in the current language.
- [x] **FB-40**: Align monitoring log and service monitoring E2E with the current copy baseline, fixing clear confirmation, detail fields, and refresh information assertions.
- [x] **FB-41**: Fix dictionary global-effect regression where label updates did not reliably propagate to department and user pages, and add cleanup-path verification.
- [x] **FB-42**: Fix empty button menu names under "Plugin Management" in menu management so built-in plugin button permissions display readable names and return localized titles.
- [x] **FB-43**: Close residual Chinese button menu titles and repeated resource names in English for menu management, unifying host and source-plugin button titles to short action words with regression coverage.
- [x] **FB-44**: Fix dictionary type and data lists that still displayed Chinese built-in governance data in dictionary management under English.
- [x] **FB-45**: Fix built-in log types, statuses, and default log summaries that still displayed Chinese in audit log and login log lists under English.
- [x] **FB-46**: Fix job, group, and execution log lists that still displayed Chinese built-in scheduler data in cron job management under English.
- [x] **FB-47**: Fix authentication-management and user-logout audit summaries that still displayed Chinese in audit logs under English.
- [x] **FB-48**: Fully localize system API documentation in English for host, source plugin, and dynamic plugin route groups, endpoint descriptions, request/response parameter descriptions, and validation metadata.
- [x] **FB-49**: Fix frontend type-check blockers exposed by English API documentation convergence, covering permission display formatting and unused workspace imports.
- [x] **FB-50**: Migrate system API documentation i18n to English API metadata source copy plus independent `apidoc i18n JSON` translation resources, and add corresponding specification and review rules.
- [x] **FB-51**: Migrate `apidoc i18n JSON` from English source-copy keys to stable structured keys, preventing English wording changes from invalidating translation mappings, and explicitly exclude `eg/example` sample values from translation.
- [x] **FB-52**: Empty `en-US` apidoc translation resources into placeholder files so English API docs are driven directly by English API source copy, while preserving non-English translation coverage validation.
- [x] **FB-53**: Split plugin apidoc i18n resources into plugin-owned directories, with lina-core apidoc discovery, merge, and rendering across those resources.
- [x] **FB-54**: Remove temporary Chinese-to-English conversion for generated code and database metadata from the apidoc service layer, ensuring API docs display generated metadata as supplied by its source.
- [x] **FB-55**: Move host and plugin apidoc i18n resources under `manifest/i18n/<locale>/apidoc/`, keeping API documentation translations under the same i18n root as runtime UI resources but isolated in a separate subdomain.
- [x] **FB-56**: Adjust business runtime i18n JSON delivery rules to support hierarchical maintenance while preserving the internal flat governance model.
- [x] **FB-57**: Adjust apidoc i18n JSON to support hierarchical maintenance, multi-file splitting, and common fallback deduplication, with unit tests and E2E verification.
- [x] **FB-58**: Adjust bottom pagination page-size selector width under English so `items/page` displays completely, and add English layout regression assertions.
- [x] **FB-59**: Update the operation log plugin audit route metadata storage and display flow to reuse apidoc i18n resources for module names and operation summaries by current language, covering list, detail, and export scenarios.
- [x] **FB-60**: Based on `backend-owned-data-i18n-audit.md`, close frontend/backend data translation mappings by moving localization of built-in roles, cron jobs, job groups, execution logs, and login log display data to the backend, and remove historical `display-l10n` mappings.
- [x] **FB-61**: Decouple the host plugin service from operation log plugin persistence structures by removing host-side audit record payload encoding and letting the operation log plugin middleware write directly through its own service.
- [x] **FB-62**: Replace operation-log-specific metadata in the host dynamic route runtime with generic dynamic route metadata, leaving operation log field projection to the operation log plugin middleware.
- [x] **FB-63**: Add key implementation comments to the operation log plugin audit middleware explaining metadata sources, override order, response fallback, and persistence strategy.
- [x] **FB-64**: Clarify runtime translation method semantics for current locale, source copy, translation-key placeholders, and default-locale fallback, preventing missing English translations from implicitly displaying Chinese default-locale content.
- [x] **FB-65**: Remove source-registration metadata duplicates from scheduler English runtime i18n JSON so host, source plugin, and dynamic plugin English display comes consistently from registered source copy.
- [x] **FB-66**: Replace string literals for scheduler job list sort fields and default sort direction with named constants that express the external sorting contract.
- [x] **FB-67**: Remove login log plugin hard-coded reverse lookup from Chinese/English source text to translation keys, and instead pass stable message keys from authentication events for backend current-language translation.
- [x] **FB-68**: Move login log i18n keys out of the shared authentication hook contract and back into the login log plugin, leaving the shared component to publish only stable authentication reason codes.
- [x] **FB-69**: Further narrow the shared authentication hook payload contract by removing authentication log display fallback copy constants from the shared component.
- [x] **FB-70**: Improve the cron job detail modal layout in English so long field labels display on one line, with layout regression verification.
- [x] **FB-71**: Fix hard-coded notice type options in the `content-notice` plugin `data.ts` and `notice-modal.vue` search/create/edit forms, load them from the `sys_notice_type` dictionary, and add `TC0121` regression coverage.
- [x] **FB-72**: Remove the main application's obsolete login log directory `apps/lina-vben/apps/web-antd/src/views/monitor/loginlog/` and its route/menu registration, with the feature fully handled by the `monitor-loginlog` plugin.
- [x] **FB-73**: Add `lina-review` rules that prohibit frontend `Select` options from becoming a second data source beside table-column `DictTag` for the same dictionary, prohibit shared `pages.*` locale namespaces from duplicating `sys_*` dictionary labels, and require host consumers of plugin dictionary semantics to use backend-returned localized fields instead of frontend enum maps.
- [x] **FB-76**: Remove the main application's obsolete notice directory `apps/lina-vben/apps/web-antd/src/views/system/notice/`; move the preview modal still reused by the message center to `views/system/message/notice-preview-modal.vue`, with the feature fully handled by the `content-notice` plugin.
- [x] **FB-77**: Refactor the host message center so it no longer depends on `pages.status.notice` / `pages.status.announcement`: list/detail APIs add localized `typeLabel`, frontend badge/list/detail preview consumers use API labels directly, frontend locale keys are removed, and `TC0122` API contract regression coverage is added.
- [x] **FB-74**: Improve the three-state button group layout in cron job details, avoiding plugin-unavailable state wrapping against edges and adding English layout regression coverage.
- [x] **FB-75**: Simplify plugin-unavailable status display copy to `Unavailable` and its localized equivalent, restore the status field to a single-column detail form display, and keep stop reason and prompts for detailed explanations and regression assertions.
- [x] **FB-78**: Refactor the host message center to be fully category-agnostic: remove notice/announcement category and legacy type constants, replace numeric `Type` with `CategoryCode string` plus `TypeLabel`/`TypeColor`, resolve host categories through `notify.category.{code}.{label,color}` i18n conventions, move notice/announcement category ownership into the `content-notice` plugin resources, update frontend models to `categoryCode`/`typeColor`, remove numeric type helpers, and add `TC0123` to verify dynamic plugin category extensibility.
