## Why

`framework-i18n-foundation` has delivered host, plugin, and frontend runtime i18n baseline capabilities, with full functionality verified under `zh-CN` and `en-US` bilingual environments. However, 78 feedback items gradually exposed three categories of systemic issues:

1. **Performance layer**: The `Translate` hot path clones the entire runtime message bundle (800+ keys) on every call (a single menu list request = ~100 large map allocations); the cache granularity is "all languages, all sectors cleared at once", so any plugin enable/disable event invalidates all languages; the runtime translation bundle API has no ETag/version stamp, causing full retransmission on every frontend language switch.
2. **Consistency layer**: Each of the 5 `*_i18n.go` adapter files independently decides "when to translate / when to skip / which Translate* to use", causing `dict` to `return` on default locale, `menu` to always `Translate`, and `sysconfig` to hardcode English/Chinese maps in Go (violating convention-based translation key decisions); the initial approach tried to centralize via `LocaleProjector`, but that would cause the i18n foundation service to reverse-couple with business entities and business protection rules, so the decision was changed to "business modules own projection rules, i18n only provides underlying capabilities"; `isSourceTextBackedRuntimeKey` leaks jobmgmt's `job.handler.*` naming prefix into the i18n package, creating a reverse dependency; the `Service` interface carries 5 responsibility categories with 18 methods, while business modules only need the first two but must stub the entire interface.
3. **Boundary layer**: `apidoc` and runtime bundle each maintain ~280 lines of structurally identical host/plugin/dynamic loaders and caches; frontend `loadMessages` uses `Promise.all` for three things with different failure semantics (runtime bundle / public config / third-party libraries); WASM custom section parsing is duplicated inside the i18n package, causing i18n to reverse-depend on the WASM file format.

Additionally, the current framework has only been verified under `zh-CN`/`en-US` bilingual; whether "convention-based translation keys + missing checks" truly requires zero business code changes when introducing a third language has never been stress-tested. This change introduces Traditional Chinese (`zh-TW`) as a litmus test, covering non-Latin characters and a more complete resource governance path; document direction is fixed to LTR per current host positioning, with RTL not supported in this iteration.

## What Changes

### P1 Performance Optimization
- Rewrite `Translate`/`TranslateSourceText`/`TranslateOrKey`/`TranslateWithDefaultLocale` hot paths to read directly from cache instead of cloning the entire message bundle; only methods like `BuildRuntimeMessages` that need to hand messages to the frontend retain clone semantics.
- Refactor `runtimeBundleCache` into a layered structure by `locale + sector (host/source-plugin/dynamic-plugin)`, so invalidation only clears relevant sectors or languages instead of clearing the entire cache.
- Runtime translation bundle API `/i18n/runtime/messages` outputs `ETag` (based on locale + bundleVersion) and supports `If-None-Match` 304 negotiation; backend maintains a `bundleVersion` atomic counter that auto-increments on any invalidation.
- Frontend `runtime-i18n.ts` switches from raw `fetch` to `requestClient`, integrating auth/error/degradation chain; adds `localStorage` persistence on top of in-memory cache, enabling zero-network language switching on subsequent page loads.

### P2 Consistency Convergence
- Converge projection rules in `menu_i18n.go`/`dict_i18n.go`/`sysconfig_i18n.go`/`jobmgmt_i18n.go`/`role.go`, but rules must remain within each business module's boundary; `internal/service/i18n` only exposes `ResolveLocale`, `Translate`, `TranslateSourceText` and other underlying capabilities, and must not provide business-entity-named projectors.
- Delete `englishLabels`/`chineseLabels` Go maps in `sysconfig_i18n.go`, replacing them with `config.field.<name>` keys through `manifest/i18n/<locale>/*.json`, validating that Decision 1 (convention-based translation keys) truly covers edge cases like export/import template headers.
- Introduce `RegisterSourceTextNamespace(prefix, reason)` explicit registry in the `i18n` package; remove the `isSourceTextBackedRuntimeKey` blacklist from `i18n_manage.go`; jobmgmt registers its own namespace in its own `init`.
- Split the `i18n.Service` large interface into `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer` four smaller interfaces; `serviceImpl` implements all four uniformly; business modules' `i18nSvc` fields only declare the minimum interfaces they actually depend on.

### P3 Boundary Cleanup
- Extract `pkg/i18nresource` common resource loader accepting `Subdir`, `LocaleSubdir`, `LayoutMode` and `PluginScopeNamespace` configuration; `apidoc_i18n_loader.go` and `i18n.go` resource loading shells share the same implementation, eliminating ~280 lines of duplicate code while preventing apidoc from reverse-depending on `internal/service/i18n`.
- Frontend `loadMessages` split: runtime bundle failure -> hit persistent cache or fallback; public config failure -> fire-and-forget without blocking; third-party library locale -> must await.
- WASM custom section parsing `parseWasmCustomSectionsForI18N` and `readWasmULEB128ForI18N` moved to `pkg/pluginbridge/pluginbridge_wasm_section.go`; i18n package only calls `pluginbridge.ReadCustomSection(content, name)`.

### Third Language: Traditional Chinese as Stress Test
- Runtime auto-discovers built-in languages from `manifest/i18n/<locale>/*.json`; the `i18n` section in the default config file maintains default language, multi-language toggle, sorting, native names and enabled whitelist; runtime language list, language switching, missing translation checks, resource source diagnostics must all automatically cover `zh-TW` without modifying business module code, SQL seeds, or frontend TS language lists.
- Host `manifest/i18n/zh-TW/*.json` and all source plugin `apps/lina-plugins/<plugin-id>/manifest/i18n/zh-TW/*.json` must be filled in; frontend `apps/lina-vben/apps/web-antd/src/locales/langs/zh-TW/*.json` and `packages/locales/src/langs/zh-TW/*.json` must have corresponding static translations.
- `apidoc i18n` also adds `zh-TW`; plugin apidoc translation resources must be filled in simultaneously.
- Plural/number formatting follows `ICU MessageFormat` style API convention on the frontend; initial scope is limited to `count`-type copy (list statistics, batch delete prompts), without requiring all pages to adopt immediately.

### Fixed LTR Direction (Included in This Change)
- `<html dir>` and `Ant Design Vue`'s `ConfigProvider.direction` are fixed to `ltr` per current host convention; default config does not provide a `direction` field.
- Acceptance criteria: under Traditional Chinese, pages show "correct content, functional", without providing RTL mirrored layout.

### Single Source of Truth for Resources (Included in This Change)
- Remove `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` three runtime i18n persistence tables; no longer provide database translation overrides, language registration seeds, or generic business content multilingual tables.
- Retain export, missing checks and source diagnostics APIs as development/delivery-time auxiliary capabilities; export results are used for offline proofreading and writing back to JSON resources, not for re-importing to the database.
- The only required change for adding a new language is corresponding JSON resources in host/plugin/frontend, with optional changes being `i18n` metadata in the default config file.

### Out of Scope (Separate Future Changes)
- RTL design language: icon mirroring, drawer/notification slide-in direction, table fixed column flipping, comprehensive CSS logical properties replacement, menu expand direction.
- User-level language preference persisted to `sys_user`.
- Internationalization visual admin management page.
- Online hotfix translation override capability; if needed in the future, should be designed as an optional plugin.
- Automatic machine translation / AI translation assistance.

## Capabilities

### New Capabilities
- `framework-i18n-runtime-performance`: Zero-copy hot path for runtime translation lookup, sector-based layered cache invalidation, and frontend ETag/304 negotiation with persistent cache capabilities.
- `framework-i18n-module-projection`: Localization projection rules within business module boundaries, encapsulating translation key derivation, skip strategies, and fallback selection, while prohibiting the i18n foundation service from reverse-perceiving business entities.
- `framework-i18n-source-text-registry`: Explicit registration mechanism for code-owned source text namespaces, enabling missing checks, diagnostics, and export to identify "translation keys owned by code sources" without the i18n package reverse-perceiving specific business modules.

### Modified Capabilities
- `framework-i18n-foundation`: Translation service interface split into `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer` multiple smaller interfaces; runtime translation bundle API adds ETag/304 negotiation semantics; adds `zh-TW` as a built-in enabled language; runtime direction fixed to LTR; translation resource loading shares a unified resource loader across host, source plugins, and dynamic plugins; runtime i18n persistence tables removed, translation content converged to JSON/YAML resources as single source of truth.
- `system-api-docs`: API documentation translation resource loading switched to unified resource loader, with `zh-TW` coverage synchronized.
- `plugin-runtime-loading`: WASM custom section reading capability elevated to `pluginbridge` public capability; plugin runtime and i18n share the same parser.
- `config-management`: Config import/export headers switched to translation key resolution, deleting backend hardcoded `englishLabels`/`chineseLabels` Go mappings.
- `plugin-manifest-lifecycle`: Plugin manifest and lifecycle must automatically cover new languages without modifying host code or plugin code.

## Impact

- **Backend capabilities**: Rewrite `Translate*` hot paths and cache layer in `apps/lina-core/internal/service/i18n/`; split `Service` interface; adjust 5 `*_i18n.go` adapter files and delete centralized projector; extract `pkg/i18nresource`, `pkg/pluginbridge/pluginbridge_wasm_section.go`; `apidoc_i18n_loader.go` and `i18n.go` resource loading shells share loader; `config-management` deletes hardcoded Go label maps.
- **Database**: Remove runtime i18n persistence tables; built-in language enablement driven by `manifest/i18n/<locale>/*.json` and the `i18n` section in the default config file.
- **Frontend capabilities**: `runtime-i18n.ts` switched to `requestClient` with `localStorage` integration; `loadMessages` split failure semantics; language menu driven by runtime language metadata, document direction fixed to LTR; new `zh-TW` static language packs.
- **Resource files**: Host `manifest/i18n/zh-TW/*.json`, `manifest/i18n/zh-TW/apidoc/**/*.json`; each source plugin's `manifest/i18n/zh-TW/*.json` and `manifest/i18n/zh-TW/apidoc/**/*.json`; frontend `packages/locales/src/langs/zh-TW/*.json` and `apps/web-antd/src/locales/langs/zh-TW/*.json`.
- **Testing**: Backend supplement `Translate` hot path benchmark tests and layered cache invalidation unit tests; frontend supplement runtime ETag/persistent cache unit tests; new E2E cases covering `zh-TW` language switching, key page text completeness, fixed `<html dir="ltr">` assertions; run `lina-review` for frontend-backend consistency validation.
- **Delivery and maintenance**: `README.md` and `manifest/i18n/README.md` documentation synchronized to describe ETag, new language process, and fixed LTR boundaries; `OpenSpec specs` synchronized to update corresponding clauses in the `framework-i18n-foundation` main spec.
