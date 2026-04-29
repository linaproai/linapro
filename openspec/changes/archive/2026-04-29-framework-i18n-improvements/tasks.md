## 1. P1 Performance Optimization: Translation Hot Path and Cache Layering

- [x] 1.1 Rewrite `Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` implementations in `apps/lina-core/internal/service/i18n/i18n.go` to "hold read lock and read cache directly", removing `cloneFlatMessageMap` calls from single-key lookup paths
- [x] 1.2 Preserve clone semantics for `BuildRuntimeMessages` and `ExportMessages`; add internal `lookupBundleKey(locale, key)` utility method to unify cache hit read management
- [x] 1.3 Refactor `runtimeBundleCache` structure to layer by `locale x sector (host / source-plugin / dynamic-plugin)`, adding `mergedView` and `bundleVersion` atomic counter
- [x] 1.4 Refactor `InvalidateRuntimeBundleCache` to accept `InvalidateScope` parameter, providing fine-grained invalidation by locale, sector, and plugin ID; migrate all call sites to explicit scope
- [x] 1.5 Remove runtime content translation cache design, confirming this iteration no longer provides `sys_i18n_content` business content translation table
- [x] 1.6 Add `Translate` single/batch call benchmark tests (`testing.B`), target: single call < 100ns on cache hit
- [x] 1.7 Add layered invalidation unit tests covering "host resource only clears host sector", "plugin enable/disable only clears dynamic sector", "source plugin registration only clears source plugin sector" scenarios
- [x] 1.8 Add review rule in `lina-review` skill: prohibit business modules from cloning runtime message bundles outside the i18n package; `InvalidateRuntimeBundleCache` must receive explicit scope

## 2. P1 Performance Optimization: Runtime Translation Bundle ETag Negotiation

- [x] 2.1 Add `BundleVersion()` method to `i18n.Service` returning the current runtime translation bundle version; any sector invalidation must increment this version
- [x] 2.2 Modify `apps/lina-core/internal/controller/i18n/i18n_v1_runtime_messages.go` to output `ETag: "<locale>-<version>"` and `Cache-Control: private, must-revalidate` in responses
- [x] 2.3 Implement `If-None-Match` negotiation in that controller, returning `304 Not Modified` with no message body when matched
- [x] 2.4 Add unit tests in `apps/lina-core/internal/controller/i18n/i18n_v1_runtime_test.go` covering ETag output, 304 response, and ETag must differ after version change
- [x] 2.5 Create E2E test case `TC0124-runtime-i18n-etag.ts` verifying backend ETag and 304 negotiation work correctly in multilingual switching flows (Traditional Chinese related verification deferred to Section 12 TC0127-TC0129)

## 3. P1 Performance Optimization: Frontend RequestClient and Persistent Cache

- [x] 3.1 Rewrite `apps/lina-vben/apps/web-antd/src/runtime/runtime-i18n.ts`: replace raw `fetch` with `requestClient`, add Bearer injection, error degradation, retry chain
- [x] 3.2 Add `localStorage` persistence layer in `runtime-i18n.ts`: `linapro:i18n:runtime:<locale>` key, value is `{etag, messages, savedAt}`, TTL 7 days
- [x] 3.3 Implement "persistent hit renders immediately, background If-None-Match negotiation, 304 does not update" fast path
- [x] 3.4 Refactor `loadMessages` in `apps/lina-vben/apps/web-antd/src/locales/index.ts` to split by failure semantics: runtime bundle failure -> persistent fallback + user notification; public config failure -> fire-and-forget; third-party library locale -> must await
- [x] 3.5 Add `runtime-i18n.test.ts` unit tests covering persistent hit, TTL expiration forced refresh, 304 path, and network error degradation scenarios
- [x] 3.6 Add `loadMessages` unit tests covering independent failure semantics for the three tasks

## 4. P2 Consistency Convergence: Business Module Localization Projection Boundaries

- [x] 4.1 Delete the originally planned `apps/lina-core/internal/service/i18n/i18n_projector.go` centralized projector approach, confirming it was introduced by this task group and would cause the i18n foundation service to reverse-couple with business entities and business protection rules
- [x] 4.2 Retain "when to translate / when to skip / which Translate*" projection decisions in each business module's own `*_i18n.go`; `i18n` package only provides `ResolveLocale` / `Translate` / `TranslateSourceText` and other underlying capabilities
- [x] 4.3 Refactor `apps/lina-core/internal/service/menu/menu_i18n.go`; menu translation key derivation owned by menu module
- [x] 4.4 Refactor `apps/lina-core/internal/service/dict/dict_i18n.go`; dictionary default language skip strategy and `dict.*` key conventions owned by dict module
- [x] 4.5 Refactor `apps/lina-core/internal/service/sysconfig/sysconfig_i18n.go`; config projection and field header translation keys owned by sysconfig module
- [x] 4.6 Refactor `apps/lina-core/internal/service/jobmgmt/jobmgmt_i18n.go`; built-in tasks and default task group protection rules owned by jobmgmt module
- [x] 4.7 Refactor `apps/lina-core/internal/service/role/role.go`; built-in admin role projection rules owned by role module
- [x] 4.8 Refactor `apps/lina-core/internal/service/plugin/internal/runtime/registry.go`; plugin metadata projection rules owned by plugin runtime module
- [x] 4.9 Add/retain business module localization projection tests covering default language skip, built-in protected record translation, and user records preserving original values

## 5. P2 Consistency Convergence: Delete sysconfig Hardcoded Label Maps

- [x] 5.1 Fill in `config.field.name` / `config.field.key` / `config.field.value` / `config.field.remark` / `config.field.createdAt` / `config.field.updatedAt` in `apps/lina-core/manifest/i18n/zh-CN/*.json` and `apps/lina-core/manifest/i18n/en-US/*.json`
- [x] 5.2 Refactor `sysconfig_i18n.go::buildLocalizedImportTemplateHeaders` and `buildLocalizedExportHeaders`; delete `englishLabels` / `chineseLabels` Go maps, replacing with `i18nSvc.Translate(ctx, "config.field."+name, fallback)` resolution
- [x] 5.3 Delete hardcoded `ResolveLocale == "en-US"` check inside `localizedConfigFieldLabel`
- [x] 5.4 Create E2E test case `hack/tests/e2e/settings/config/TC0125-config-export-headers-via-i18n-keys.ts` verifying export headers change with language switching without depending on Go maps
- [x] 5.5 Add review rule in `lina-review` skill: `apps/lina-core/internal/service/sysconfig/` and other business modules are prohibited from maintaining English/Chinese copy Go maps

## 6. P2 Consistency Convergence: Source-Text Namespace Explicit Registration

- [x] 6.1 Create `RegisterSourceTextNamespace(prefix, reason string)` and query functions in `apps/lina-core/internal/service/i18n/i18n_source_text_namespace.go`; data stored as package-level `sync.RWMutex` protected `map[string]string`
- [x] 6.2 Delete hardcoded entries in `apps/lina-core/internal/service/i18n/i18n_manage.go::isSourceTextBackedRuntimeKey`, replacing with namespace registry queries
- [x] 6.3 Add `init()` in `apps/lina-core/internal/service/jobmeta` or `jobmgmt` package to register `job.handler.` and `job.group.default.` namespaces
- [x] 6.4 Add unit tests covering "unregistered namespaces do not exempt missing checks" and "registered namespaces disappear from missing results" scenarios
- [x] 6.5 Add review rule in `lina-review` skill: prohibit hardcoded checks with business namespace prefixes like `job.handler.` / `job.group.` inside `apps/lina-core/internal/service/i18n/`

## 7. P2 Consistency Convergence: Service Interface Split

- [x] 7.1 Split the interface in `apps/lina-core/internal/service/i18n/i18n.go` into `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer`; `Service` becomes a composition of these four small interfaces
- [x] 7.2 Converge `i18nSvc` field types in `menu` / `dict` / `sysconfig` / `jobmgmt` / `role` / `usermsg` modules to minimum dependency interfaces (in most cases a combination of `LocaleResolver + Translator`)
- [x] 7.3 Converge `apidoc` service's `i18nSvc` field to `LocaleResolver + Translator`
- [x] 7.4 Converge controller field types in `controller/i18n/`; management controllers only hold `Maintainer`, runtime controllers hold `BundleProvider + LocaleResolver`
- [x] 7.5 Update all related unit test mocks to only stub the small interfaces actually depended upon
- [x] 7.6 Add review rule in `lina-review` skill: business module field types should prefer minimum interface declarations, prohibiting default declarations of full `Service`

## 8. P3 Boundary Cleanup: Unified ResourceLoader

- [x] 8.1 Create stable public `ResourceLoader` component in `apps/lina-core/pkg/i18nresource/`, accepting `Subdir` / `LocaleSubdir` / `PluginScope` / `LayoutMode` configuration parameters, avoiding apidoc reverse-depending on `internal/service/i18n`
- [x] 8.2 Implement `LoadHostBundle(ctx, locale)` / `LoadSourcePluginBundles(ctx, locale)` / `LoadDynamicPluginBundles(ctx, locale, releases)` three clear-responsibility methods
- [x] 8.3 Refactor `apps/lina-core/internal/service/i18n/i18n.go` to use `i18nresource.ResourceLoader{Subdir: "manifest/i18n", LayoutMode: LocaleDirectory, PluginScope: Open}` replacing duplicate implementations
- [x] 8.4 Refactor `apps/lina-core/internal/service/apidoc/apidoc_i18n_loader.go` to use `i18nresource.ResourceLoader{Subdir: "manifest/i18n", LocaleSubdir: "apidoc", LayoutMode: LocaleSubdirectoryRecursive, PluginScope: RestrictedToPluginNamespace}` replacing duplicate implementations
- [x] 8.5 Delete duplicate directory traversal, ULEB128 decoding, and `wasm` section parsing code on both sides, achieving convergence of duplicate implementations
- [x] 8.6 Add `ResourceLoader` unit tests covering host, source plugin, and dynamic plugin sources as well as plugin namespace isolation

## 9. P3 Boundary Cleanup: WASM Parsing Moved to pluginbridge

- [x] 9.1 Create `ReadCustomSection(content []byte, name string)` and `ListCustomSections(content []byte)` public functions in `apps/lina-core/pkg/pluginbridge/pluginbridge_wasm_section.go`, migrating `wasm` file header validation, section traversal and ULEB128 decoding logic
- [x] 9.2 Centralize `pluginbridge.WasmSection*` section name constants in the `pluginbridge` package
- [x] 9.3 Delete `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` private functions in `apps/lina-core/internal/service/i18n/i18n_plugin_dynamic.go`, replacing with calls to `pluginbridge.ReadCustomSection`
- [x] 9.4 Adjust dynamic plugin apidoc resource loading flow in the `apidoc` package to uniformly use `pluginbridge.ReadCustomSection`
- [x] 9.5 Add `pluginbridge` WASM unit tests covering normal section read, file header errors, and ULEB128 out-of-bounds scenarios
- [x] 9.6 Verify plugin runtime `pkg/pluginhost` and i18n sharing the same parsing path has no regressions

## 10. Traditional Chinese Integration: Data and Resources

- [x] 10.1 In the default config file's `i18n` section, discover languages via JSON files and maintain default language, sorting and native names, no longer enabling `zh-TW` through SQL seed
- [x] 10.2 Delete this iteration's `zh-TW` SQL seed dependency, confirming adding built-in languages does not require modifying host `manifest/sql/` files
- [x] 10.3 Create `apps/lina-core/manifest/i18n/zh-TW/*.json` fully covering host runtime UI translation keys, consistent with `zh-CN/*.json` key set
- [x] 10.4 Create `apps/lina-core/manifest/i18n/zh-TW/apidoc/**/*.json` fully covering host API documentation translation keys
- [x] 10.5 Add `manifest/i18n/zh-TW/*.json` and corresponding `manifest/i18n/zh-TW/apidoc/**/*.json` in each source plugin directory (`org-center` / `monitor-online` / `monitor-loginlog` / `monitor-operlog` / `monitor-server` / `content-notice` / `plugin-demo-source` / `plugin-demo-dynamic` / `demo-control`)
- [x] 10.6 Create `apps/lina-vben/packages/locales/src/langs/zh-TW/{authentication.json,common.json,preferences.json,profile.json,ui.json}` five static language pack files
- [x] 10.7 Create `apps/lina-vben/apps/web-antd/src/locales/langs/zh-TW/{demos.json,page.json,pages.json}` three project-level language pack files
- [x] 10.8 In `apps/lina-vben/apps/web-antd/src/locales/index.ts`, change to deriving dayjs / antd locale by language code convention, avoiding continued addition of `case '<locale>'` branches when adding new languages
- [x] 10.9 Run `CheckMissingMessages(locale='zh-TW')` confirming it returns `total=0` (registered code-owned namespaces are exempt)

## 11. Fixed LTR Direction Integration

- [x] 11.1 In `apps/lina-vben/packages/locales/src/i18n.ts`, fix runtime language direction to `ltr`, without maintaining a static direction language registry
- [x] 11.2 In `setI18nLanguage(locale)`, fixed to set `document.documentElement.dir = 'ltr'`, and retain reactive `direction` state for component use
- [x] 11.3 In `apps/lina-vben/apps/web-antd/src/bootstrap.ts` and `App.vue` (or corresponding root component), inject fixed `ltr` direction into `Ant Design Vue`'s `ConfigProvider`
- [x] 11.4 Verify `<html dir>` and `ConfigProvider direction` always remain `ltr` during language switching
- [x] 11.5 Create E2E test case `TC0126-traditional-chinese-ltr-direction-switch.ts` verifying `<html dir>` and antd component direction remain `ltr` during multilingual switching
- [x] 11.6 Add scope description in `lina-review` skill: this change does not support RTL design language

## 12. Traditional Chinese Regression and Stress Testing

- [x] 12.1 Create E2E test case `TC0127-traditional-chinese-page-content-audit.ts`; under Traditional Chinese, visit each page of the framework's default delivery menu routes, confirming no Simplified Chinese / English residuals
- [x] 12.2 Create E2E test case `TC0128-traditional-chinese-plugin-pages.ts` covering source plugin pages and dynamic plugin example pages under Traditional Chinese display
- [x] 12.3 Create E2E test case `TC0129-traditional-chinese-apidoc.ts` verifying `/api.json?lang=zh-TW` returns API documentation groups, summaries, and parameter descriptions displayed in Traditional Chinese
- [x] 12.4 In CI / local `make test` pipeline, ensure `CheckMissingMessages(locale='zh-TW')` threshold matches `en-US` (both `total=0`), with missing items blocking
- [x] 12.5 Write "adding new language process" documentation draft, placed in `apps/lina-core/manifest/i18n/README.md` and Chinese mirror

## 13. Performance Verification and Benchmarking

- [x] 13.1 Write `Translate` hot path benchmark test `apps/lina-core/internal/service/i18n/i18n_bench_test.go` covering single key lookup, batch 100 lookups, and cache miss rebuild scenarios
- [x] 13.2 Verify benchmark results: single `Translate` < 100ns on cache hit; batch 100 calls decrease >= 80% compared to before refactoring
- [x] 13.3 Verify runtime translation bundle API returns empty body and Content-Length 0 on ETag 304 hit
- [x] 13.4 Verify frontend shows `304` instead of `200` in Network panel when switching languages on subsequent page loads
- [x] 13.5 Attach before/after benchmark data in PR description or review report

## 14. Documentation and Review

- [x] 14.1 Update `apps/lina-core/manifest/i18n/README.md` and Chinese mirror `README.zh_CN.md`, documenting ETag negotiation, new language process, and source namespace registration
- [x] 14.2 Update `apps/lina-vben/apps/web-antd/src/locales/README.md` and Chinese mirror, documenting fixed LTR direction and persistent cache strategy
- [x] 14.3 Update root `CLAUDE.md` "i18n continuous governance requirements" paragraph, adding "prohibit modifying Go code when adding new languages" and "runtime cache invalidation must have explicit scope" rules
- [x] 14.4 Call `lina-review` skill to complete code and specification review for P1 / P2 / P3 / Traditional Chinese + Fixed LTR four groups of changes separately
- [x] 14.5 Run `make test` full E2E passes (including new `TC0124` ~ `TC0129`)
- [x] 14.6 After fixing all blocking issues exposed in review and E2E, confirm `openspec validate framework-i18n-improvements` passes

## Feedback

<!-- Feedback on issues exposed during implementation, appended via lina-feedback skill -->

- [x] **FB-1**: New files in `apps/lina-core/internal/service/i18n/` missing `i18n_` prefix, and `lina-review` not explicitly covering service component file naming checks
- [x] **FB-2**: `LocaleProjector` causes the i18n foundation service to reverse-couple with menu, dictionary, config, tasks, roles, plugin metadata and other business rules; needs immediate removal and re-evaluation of subsequent design boundaries
- [x] **FB-3**: Adding built-in languages should not require modifying backend Go enumerations, SQL seeds, or frontend TS language lists; should be driven by `manifest/i18n/<locale>/*.json` and simplified YAML metadata
- [x] **FB-4**: Remove `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` three runtime i18n persistence tables, converging translation content to development-time JSON/YAML resources as single source of truth
- [x] **FB-5**: Move language metadata configuration to default config file `i18n` section, remove `direction` configuration and fix to LTR, while eliminating multilingual hardcoding in backend fallback and frontend third-party locale fallback
- [x] **FB-6**: Add configuration item comments to default config file `i18n` section, remove Arabic resources and change to Traditional Chinese `zh-TW` support
- [x] **FB-7**: Version SQL contains `DELETE` / `DROP` / `UPDATE` statements that clean database data; needs removal and confirmation of similar script risks
- [x] **FB-8**: Merge runtime language metadata from independent language config file into default config file `i18n` section, add `enabled` toggle, ensure backend language parsing and frontend language switch button work per config when multi-language is disabled or `locales` items are removed, supplement `TC0130` regression coverage
- [x] **FB-9**: Remove default language hardcoding in `config_i18n.go`, `i18n_locale.go` and fallback paths, ensuring default language and language list only come from default config file `i18n` section
- [x] **FB-10**: Remove `config_i18n.go` dependency on default config template path and embedded file reading, changing to only read the system's loaded `i18n` config section
- [x] **FB-11**: After `config.Service` adds `GetI18n`, `role` unit test mocks not synchronously adapted, causing backend full unit test compilation failure
- [x] **FB-12**: `i18n.enabled=false` and `zh-TW` export header override still lean E2E; need to supplement backend/frontend unit protection and expand parameter export E2E assertions
- [x] **FB-13**: `zh-TW` apidoc and packed manifest resources lack unit-level integrity protection; plugin apidoc resource missing may be silently skipped by tests
- [x] **FB-14**: Runtime bundle cache full/by-language invalidation version increment lacks unit tests; `i18n.enabled=false` lacks controller-layer default language return shape coverage
- [x] **FB-15**: New i18n E2E coverage for plugin status, plugin apidoc, frontend ETag flow, Traditional Chinese login page and raw key leakage detection still incomplete
- [x] **FB-16**: Full E2E exposes default brand logo static resource missing; preferences drawer and user drawer page object locators too broad causing unstable verification
- [x] **FB-17**: Full E2E serial exposes dynamic plugin installation failure with cron authorization metadata not localized from un-enabled wasm resources, and install + enable shortcut authorization chain payload coverage for cron host service incomplete
- [x] **FB-18**: Default brand logo already unified to `/logo.png` in system config data; frontend default preferences and E2E assertions should not continue using `/linapro-mark.png`
- [x] **FB-19**: `permission-display.ts` dynamic route permission display template and fragment vocabulary still maintained in TypeScript; needs convergence to frontend i18n JSON resources with unknown permission fragment fallback display retained
- [x] **FB-20**: i18n unit tests and packed/apidoc integrity tests still hardcode `zh-TW` and other target languages; adding `ja-JP` would require test governance to modify code again
