## Context

`framework-i18n-foundation` has established the "host resources + plugin resources + runtime aggregation" system in `apps/lina-core/internal/service/i18n/`, with full functionality under `zh-CN` / `en-US` bilingual. However, 78 feedback items gradually revealed systemic issues concentrated in three dimensions: **hot path performance, module consistency, and component boundaries**:

- The current `Translate` series methods clone the entire runtime message bundle via `cloneFlatMessageMap` on every call, while callers often only need 1 key -- known hot paths (menu list, dictionary list, common config) trigger dozens of full clones per request.
- `runtimeBundleCache` invalidation is "all languages, all sectors cleared at once", so any plugin enable/disable event invalidates the cache for all languages.
- Runtime i18n persistence tables (`sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content`) pull language registration, translation content, and business content variants from the development-time resource model back into a database dual-source model, requiring new languages to understand SQL, DAO, backend maintenance APIs, and cache invalidation paths -- conflicting with the goal of "adding a new language only requires adding JSON/YAML".
- The 5 `*_i18n.go` adapter files (`menu_i18n.go` / `dict_i18n.go` / `sysconfig_i18n.go` / `jobmgmt_i18n.go` / `role.go`) each independently decide "when to translate / when to skip / which Translate* to use", with `sysconfig_i18n.go` even directly hardcoding English/Chinese maps in Go (violating Convention 1). The initial `LocaleProjector` centralized approach could reduce duplication, but would cause the i18n foundation service to reverse-couple with business entities and business protection rules, so this design changed to business modules owning projection rules.
- `i18n_manage.go::isSourceTextBackedRuntimeKey` leaks `job.handler.*`, `job.group.default.*` and other jobmgmt-private naming prefixes into the i18n package, creating reverse dependency.
- The `Service` interface has 18 methods carrying multiple responsibility categories, requiring business module tests to stub the entire large interface.
- `apidoc/apidoc_i18n_loader.go` and `i18n/i18n.go` each maintain their own resource loading functions with host/source-plugin/dynamic-plugin traversal logic, duplicating ~280 lines.
- Frontend `runtime-i18n.ts` uses raw `fetch` bypassing `requestClient`, with incomplete failure semantics and persistent cache strategy; `loadMessages` uses `Promise.all` for three things (runtime bundle / public config / third-party libraries) with undifferentiated failure handling.
- The entire framework has only been verified under `zh-CN` / `en-US` bilingual; whether convention-based translation keys, missing checks, and runtime aggregation truly require zero business code changes when introducing a third language has never been stress-tested.

Since this project has no legacy compatibility burden, this design does not preserve old `Service` interface signatures or maintain old cache structures, but directly refactors.

## Goals / Non-Goals

**Goals:**
- Make the `Translate` hot path avoid large map cloning on cache hits, with single lookup approaching constant time.
- Give cache invalidation sector-level precision: plugin enable/disable only clears the relevant plugin sector, language switches do not affect each other.
- Make runtime translation bundles support ETag/304 negotiation at the HTTP layer, with zero-network language switching on subsequent frontend page loads.
- Make the 5 business modules' `*_i18n.go` converge projection rules within their own module boundaries, eliminating obvious drift, while avoiding the i18n foundation service reverse-knowing business entities.
- Split the large `Service` interface into multiple small interfaces with clear responsibilities, with business modules only declaring actual dependencies.
- Change the `source-text` namespace from an i18n package blacklist to explicit business module registration, eliminating reverse dependency.
- Make `apidoc` and runtime bundle share the same `ResourceLoader`, eliminating ~280 lines of duplicate implementation.
- Make frontend `loadMessages` handle three things independently by their failure semantics, with persistent cache fallback.
- Move WASM custom section parsing to `pluginbridge`, so i18n no longer reverse-depends on the WASM file format.
- Make Traditional Chinese `zh-TW` the stress test baseline as the third language, validating that "convention-based translation keys + missing checks" truly requires zero business code changes when adding new languages.
- Fix document direction to LTR per current host positioning, with Traditional Chinese pages showing "correct content, functional".
- Remove `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` three runtime i18n persistence tables, converging translation content to development-time JSON/YAML resources as single source of truth.

**Non-Goals:**
- Not supporting RTL design language in this scope: icon mirroring, drawer/notification slide-in direction, table fixed column flipping, comprehensive CSS logical properties replacement, menu expand direction.
- Not introducing user-level language preference (`sys_user.preferred_locale`) in this scope.
- Not providing an internationalization visual admin management page in this scope; retaining export, missing checks, and diagnostic APIs as delivery auxiliary tools.
- Not introducing automatic machine translation, AI translation assistance, or external translation platform integration.
- Not providing online hotfix translation write capability; changing copy is done by modifying JSON/YAML resources and publishing.

## Decisions

### Decision 1: `Translate` hot path reads cache directly, abandoning cloning

**Choice**: Change `Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` to directly hold a read lock on `runtimeBundleCache` and look up values, no longer going through `buildRuntimeMessageCatalog -> cloneFlatMessageMap`. Only `BuildRuntimeMessages` (output to frontend) and `ExportMessages` retain clone semantics.

**Reason**:
- 99% of `Translate` call paths only read 1 key; cloning the entire 800+ key map is pure waste.
- Concurrent read-only access to `map[string]string` is safe as long as write paths (invalidation / cache rebuild) are protected by `sync.RWMutex`.
- Business modules do not modify the string after receiving it, so the "defensive" rationale for cloning does not hold.

**Constraints**:
- Cache rebuild must complete in a temporary map first, then atomically replace the cache entry; "clearing while writing" is prohibited.
- Any interface returning `map[string]string` to external modules must still clone once.

**Alternative**:
- Switch to `sync.Map`. Not adopted because the runtime cache does full rebuild on miss and full-table read-only on hit -- a typical read-heavy write-light scenario where `RWMutex + map` is more direct and easier to debug.

### Decision 2: `runtimeBundleCache` refactored to layered by locale + sector

**Choice**: Cache structure changed from

```
bundles map[locale]map[key]value
```

to

```
type localeCache struct {
    host       map[string]string                          // Immutable, loaded once at startup
    plugins    map[pluginID]map[string]string             // Source plugins, refreshed when source registry changes
    dynamic    map[pluginID]map[string]string             // Dynamic plugins, refreshed by plugin lifecycle hooks
    merged     map[string]string                          // Merged view by priority, invalidated when any sub-layer changes
    mergedAt   atomic.Uint64                              // Linked to global bundleVersion
}
type runtimeCache struct {
    sync.RWMutex
    locales map[string]*localeCache
    version atomic.Uint64                                 // bundleVersion, used for ETag
}
```

`Translate` reads `merged` first; on `merged` miss, merges in order dynamic > plugins > host and fills the cache.
Invalidation granularity:
- Dynamic plugin enable/disable -> only clears that plugin ID's dynamic sub-layer and merged across all locales
- Source plugin registry changes -> plugins and merged invalidated
- Resource metadata or test-triggered invalidation -> cleared by locale / sector

Each invalidation triggers `version.Add(1)`, driving frontend ETag negotiation.

**Reason**:
- The current "clear all" approach causes observable performance dips in multi-tenant / active management scenarios.
- The layered structure creates a 1:1 mapping of "who owns this layer" and "when to invalidate", making troubleshooting clear.
- The `merged` view ensures `Translate`'s hit path remains an O(1) map lookup.

**Alternative**:
- Do not introduce a `merged` view, but have `Translate` search three layers sequentially each time. Not adopted because the constant factor of multiple map lookups across languages far exceeds one.

### Decision 3: Runtime translation bundle API with ETag and persistent cache

**Choice**:
- Backend: `/i18n/runtime/messages` response headers add `ETag: "<locale>-<bundleVersion>"` and `Cache-Control: private, must-revalidate`; when request carries `If-None-Match` and it matches, return `304 Not Modified` (empty body).
- Frontend: `runtime-i18n.ts` switches to `requestClient`, persisting `{locale, etag, messages, savedAt}` to `localStorage`; on language switch, quickly renders with persistent cache first, then asynchronously negotiates with `If-None-Match` in the background; on 304 hit, keeps persistent data unchanged.
- Persistent cache TTL defaults to 7 days, after which a forced re-fetch is triggered (preventing version drift from users not restarting their browser for extended periods).

**Reason**:
- The essential characteristic of runtime translation bundles is "the vast majority of requests get exactly the same content" -- an ideal scenario for ETag.
- Persistent cache solves the experience issue of "full retransmission of ~80KB on second login", enabling instant language switching.
- TTL provides a safety net against long-term divergence between cache and actual version.

**Alternative**:
- Use `Last-Modified`. Not adopted because bundle version changes are driven by plugin enable/disable, resource reload, or test-triggered invalidation, which do not fully equate to a single file timestamp; `ETag` is more accurate.
- Backend writes to Redis for centralized bundle version storage. Not adopted because bundle caching is inherently in-process `runtimeBundleCache`, and the version number should also be an attribute of the same object, not requiring external storage.

### Decision 4: Business modules own localization projection rules, i18n only provides underlying capabilities

**Choice**: The `apps/lina-core/internal/service/i18n` package does not provide business-entity-named centralized projectors like `LocaleProjector`. Menu, dictionary, config, scheduled tasks, roles, plugin runtime and other modules maintain translation key derivation, skip strategies, built-in record determination and fallback selection in their own `*_i18n.go` or equivalent files. Business modules depend on i18n's underlying capabilities through narrow interfaces:

```go
type menuI18nTranslator interface {
    Translate(ctx context.Context, key string, fallback string) string
}

type dictI18nTranslator interface {
    ResolveLocale(ctx context.Context, locale string) string
    Translate(ctx context.Context, key string, fallback string) string
}
```

**Reason**:
- i18n is a foundational component; it should only be responsible for language resolution, translation lookup, caching, resource loading, and missing checks, and cannot reverse-know business entities like `SysMenu`, `SysJob`, `admin` role, default task group, etc., or business protection rules.
- Business translation keys inherently belong to business module contracts, such as `dict.*`, `config.*`, `job.group.default.*`, `role.builtin.admin.name`, and should be maintained by the owning module.
- Narrow interfaces reduce the test stub surface without centralizing business decisions into the i18n package.

**Constraints**:
- `internal/service/i18n` must not import menu, dictionary, config, tasks, roles, plugin runtime or other business modules or business entities.
- The i18n package must not expose methods named after business entities like `ProjectMenu`, `ProjectDictType`, `ProjectBuiltinJob`, etc.
- Business modules can check `ResolveLocale(ctx, "") == i18n.DefaultLocale`, but that check must only serve projection rules for editable fields owned by that module.

**Rejected Approach**:
- `LocaleProjector` centralized projector. Rejected because while it reduces duplication, it centralizes business entities, business protection rules, and business translation key derivation into the i18n foundation service, violating core host boundary and module decoupling principles.

### Decision 5: Split `Service` large interface into four small interfaces

**Choice**:

```go
type LocaleResolver interface {
    ResolveRequestLocale(*ghttp.Request) string
    ResolveLocale(context.Context, string) string
    GetLocale(context.Context) string
}
type Translator interface {
    Translate(ctx context.Context, key, fallback string) string
    TranslateSourceText(ctx context.Context, key, sourceText string) string
    TranslateOrKey(ctx context.Context, key string) string
    TranslateWithDefaultLocale(ctx context.Context, key, fallback string) string
    LocalizeError(ctx context.Context, err error) string
}
type BundleProvider interface {
    BuildRuntimeMessages(ctx context.Context, locale string) map[string]any
    ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor
    BundleVersion(locale string) uint64
}
type Maintainer interface {
    ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput
    CheckMissingMessages(ctx context.Context, locale, prefix string) []MissingMessageItem
    DiagnoseMessages(ctx context.Context, locale, prefix string) []MessageDiagnosticItem
    InvalidateRuntimeBundleCache(scope InvalidateScope)
}
```

`serviceImpl` implements all four interfaces; `New()` return type remains `Service` (to provide complete capability to other packages), but `Service` changes to a composition of `interface { LocaleResolver; Translator; BundleProvider; Maintainer }`. Business module field types change to the minimum small interface (e.g., `menu.serviceImpl.translator Translator`).

**Reason**:
- Business module test stubs drop from 18 methods to 5 or fewer.
- When reading `menu` code, looking at `Translator` alone suffices to understand what capability it needs; `Maintainer` is irrelevant to it.

**Alternative**:
- Keep a single `Service` interface. Not adopted because the Go convention is "small interfaces, large implementations"; splitting does not affect implementation complexity, only optimizes the usage surface.

### Decision 6: Explicit registration for `source-text` namespaces

**Choice**: Add `RegisterSourceTextNamespace(prefix, reason string)` to the `i18n` package; `isSourceTextBackedRuntimeKey` changes to read from the registry; `jobmgmt` / future business modules register their own namespaces in their own `init()`.

```go
// jobmgmt/init.go
func init() {
    i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")
    i18n.RegisterSourceTextNamespace("job.group.default.", "code-owned default group")
}
```

**Reason**:
- Eliminates the reverse dependency of the i18n package on jobmgmt.
- When adding a new code-owned source text module, only one line in its own package `init` is needed, without modifying the i18n package.
- Missing checks uniformly respect this registry for all non-default target languages; `en-US`, `zh-TW` and other languages should not be falsely reported as missing just because code-owned namespaces lack duplicate JSON keys.

**Alternative**:
- Statically declare namespaces in manifests. Not adopted because this is a code contract, not a file contract, and is better bound to Go modules.

### Decision 7: `apidoc` and runtime bundle share common `ResourceLoader`

**Choice**: Add a new common resource loading component under `apps/lina-core/pkg/i18nresource/`:

```go
type ResourceLoader struct {
    HostFS        fs.FS
    SourcePlugins func() []SourcePlugin
    Subdir        string                              // "manifest/i18n"
    LocaleSubdir  string                              // e.g., "apidoc"
    PluginScope   PluginScope                         // Open | RestrictedToPluginNamespace
    LayoutMode    LayoutMode                          // LocaleDirectory | LocaleSubdirectoryRecursive
    ValueMode     ValueMode                           // StringifyScalars | StringOnly
    KeyFilter     KeyFilter
}

func (l ResourceLoader) LoadHostBundle(ctx context.Context, locale string) map[string]string
func (l ResourceLoader) LoadSourcePluginBundles(ctx context.Context, locale string) map[string]map[string]string
func (l ResourceLoader) LoadDynamicPluginBundles(ctx context.Context, locale string, releases []ReleaseRef) map[string]map[string]string
```

`apidoc_i18n_loader.go` and `i18n.go` resource loading functions become thin shells, working through different `i18nresource.ResourceLoader` instances.

**Reason**:
- Both sides have completely identical logic structures, differing only in `Subdir` and "whether to restrict plugin namespace".
- `ResourceLoader` has one implementation and one test suite, avoiding dual-track drift.
- `ResourceLoader` is placed in `pkg/i18nresource` rather than `internal/service/i18n`, avoiding apidoc depending on the runtime i18n service package just to reuse the loader.

**Alternative**:
- Merge the apidoc loader into the `i18n` package for unified management. Not adopted because apidoc translation resources are a document-specific domain, should not share lifecycle with runtime UI bundles, and should not cause apidoc to reverse-depend on `internal/service/i18n`.

### Decision 8: WASM custom section parsing moved to `pluginbridge`

**Choice**: Move `parseWasmCustomSections*` / `readWasmULEB128*` from `i18n_plugin_dynamic.go`, `apidoc_i18n_dynamic.go`, and plugin runtime to `apps/lina-core/pkg/pluginbridge/pluginbridge_wasm_section.go`:

```go
package pluginbridge

func ReadCustomSection(content []byte, name string) ([]byte, bool, error)
func ListCustomSections(content []byte) (map[string][]byte, error)
```

`i18n`, `apidoc` and plugin runtime all access WASM sections through `pluginbridge.ReadCustomSection` / `pluginbridge.ListCustomSections`; business packages no longer directly maintain WASM file format parsing details.

**Reason**:
- The WASM file format is a natural responsibility of `pluginbridge`; i18n should only care about translation.
- Centralizing the parser means bug fixes and section type extensions only need one change.

**Alternative**:
- Move to `internal/util/wasm`. Not adopted because it violates the CLAUDE.md rule "prohibit adding internal/util catch-all directories".

### Decision 9: Frontend `loadMessages` split by failure semantics, with persistent fallback

**Choice**:

```ts
async function loadMessages(lang) {
  const persisted = readPersistedRuntime(lang);
  let runtime = persisted?.messages ?? {};
  let nextRuntimeVersion = persisted?.etag ?? '';

  try {
    const fresh = await loadRuntimeLocaleMessagesViaRequestClient(lang, persisted?.etag);
    if (fresh.notModified) {
      // 304: persisted data is valid
    } else {
      runtime = fresh.messages;
      nextRuntimeVersion = fresh.etag;
      writePersistedRuntime(lang, fresh);
    }
  } catch (err) {
    notifyDegraded('runtime-i18n', err);
    // runtime stays on persistent fallback
  }

  syncPublicFrontendSettings(lang).catch((err) => notifyDegraded('public-config', err));
  await loadThirdPartyMessage(lang);     // Must await
  return mergeMessages(appLocalesMap[lang] || {}, runtime);
}
```

**Reason**:
- The three things inherently have different failure semantics; unified `Promise.all` is over-coupling.
- Persistent fallback enables correct language switching even in weak network / offline scenarios.

**Alternative**:
- Do not introduce persistence; always wait for runtime bundle fetch before rendering. Not adopted due to poor experience and mismatch with ETag negotiation.

### Decision 10: Language discovery via resource conventions + default config metadata, Traditional Chinese as stress test baseline

**Choice**:
- Runtime built-in languages are auto-discovered from host `manifest/i18n/<locale>/*.json` files; adding new built-in languages no longer requires new Go constants, new SQL seeds, or frontend TS language list modifications.
- The `i18n` configuration section in the default config file maintains the small amount of metadata that cannot be derived from JSON filenames and that users may adjust, using a simplified structure:

```yaml
i18n:
  # Default locale used when the request locale is missing or unsupported.
  default: zh-CN
  # Enable i18n multi language.
  enabled: true

  # Locale display order and native names used by /i18n/runtime/locales.
  # Text direction is fixed to ltr by host convention and is not configurable.
  locales:
    - locale: en-US
      nativeName: English
    - locale: zh-CN
      nativeName: 简体中文
    - locale: zh-TW
      nativeName: 繁體中文
```

- Document direction is fixed to `ltr` per current host convention, with no `direction` field in configuration; the `locales` list is used for sorting, metadata overrides and enabled language whitelist, and languages not listed will not be exposed to the runtime language list. If `enabled=false`, the backend only accepts the default language, the frontend hides the language switch button and loads messages in the default language.
- Host and all source plugins' `manifest/i18n/zh-TW/*.json` and `manifest/i18n/zh-TW/apidoc/**/*.json` must be filled in, otherwise `CheckMissingMessages` will return non-empty results.
- Frontend `packages/locales/src/langs/<locale>/*.json` and `apps/web-antd/src/locales/langs/<locale>/*.json` are auto-discovered via directory convention; the language switch menu obtains data from `/i18n/runtime/locales`, without maintaining a static TS language list.
- `<html dir>` and antd `ConfigProvider.direction` are fixed to `ltr`; Traditional Chinese only validates translation resource completeness and page usability.
- dayjs / antd / vxe locales are loaded by deriving from the language code convention: first try the package name matching the full locale, then fall back to the language family, no longer requiring switch branch additions for each new language.
- Plural and number formatting: Traditional Chinese and Simplified Chinese both belong to CJK languages, so this iteration does not treat plural forms as an acceptance criterion for new languages; copy involving `count` continues using existing `vue-i18n` parameterization capabilities.

**Reason**:
- This `zh-TW` integration exposed design-level issues: if adding a language requires changes to SQL, backend Go constants, and frontend TS branches simultaneously, it means language registration is still multi-point enumeration rather than resource-driven.
- Default config only expresses information that cannot be stably derived or that users may adjust, such as default language, multi-language toggle, sorting and native names; other capabilities are derived from filenames, directories and third-party package naming conventions, following "convention over configuration".
- Traditional Chinese is used to validate new language resource discovery, missing checks, plugin resource coverage, and non-Latin character display, no longer carrying RTL acceptance goals.
- Fixed LTR aligns with the current project's primary positioning toward domestic Chinese developers, reducing configuration and maintenance complexity.
- RTL design language is a separate scope of work, excluded from this iteration.

**Alternative**:
- Choose `ja-JP`. Not adopted because this feedback explicitly requested Traditional Chinese, and Simplified/Traditional Chinese pairs are better suited to validate the low-risk maintenance path of "adding a new language only requires resource changes and optional YAML".
- Continue using `sys_i18n_locale` seed as the built-in language registry. Not adopted because it would require SQL modifications for new languages, conflicting with the goal of "only adding resource files + optional YAML metadata".
- Frontend continues maintaining `SUPPORT_LANGUAGES` / direction language registry / third-party locale switch. Not adopted because these TS enumerations would cause new languages to spread into application code.
- Support basic or full RTL. Not adopted because the current project positioning does not require RTL, and keeping it would increase configuration and testing maintenance complexity.

### Decision 11: Remove runtime i18n persistence tables, translation content maintained only by resource files

**Choice**:
- Delete `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` three tables and their DAO/DO/Entity, service interfaces, controller entries and tests.
- Runtime language list is only auto-discovered from `manifest/i18n/<locale>/*.json`, with the `i18n` section in the default config file supplementing default language, multi-language toggle, native names, sorting and enabled whitelist.
- Runtime messages are only aggregated from host JSON, source plugin JSON, and dynamic plugin WASM custom section snapshots; no database override layer exists.
- Retain export, missing checks and source diagnostics capabilities, but they are development/delivery-time auxiliary APIs: export results are used for offline proofreading and writing back to JSON, diagnostics sources only report host/source-plugin/dynamic-plugin.
- No longer provide a generic `sys_i18n_content` business content multilingual table. If a future business module needs record-level multilingual content, it should design its own storage and API within its own module boundary, without putting the business content model into the foundational i18n service.

**Reason**:
- The lowest-risk path for adding new languages should be "supplement resources + optional YAML metadata", not modifying SQL, backend Go, frontend TS and cache invalidation strategy.
- Database overrides create a dual-source truth of JSON and database, making auditing, missing checks, and delivery write-back more complex.
- No currently deployed business module consumes `sys_i18n_content`; putting it into the core prematurely encourages future modules to bypass their own boundaries and depend on foundational tables.
- The project has no legacy compatibility burden, so this over-designed path can be directly removed.

**Alternative**:
- Keep three tables but no longer seed new languages. Not adopted because as long as the tables and API exist, new languages and translation maintenance will still face dual-source semantics.
- Keep `sys_i18n_message` as an online hotfix channel. Not adopted because the current iteration goal is to reduce new language complexity; online hotfix translation can be redesigned as an optional plugin in the future.

## Boundary Reassessment

This feedback exposed a design-level misjudgment: `LocaleProjector` centralized business entities, business translation keys, and protected record determination into the i18n foundation service to reduce duplication. This design has been rejected and removed. After re-evaluating the current iteration, subsequent tasks proceed with the following boundaries:

- Task 7 `Service` interface split: Retained. It reduces business modules' dependency surface on the complete i18n service; the direction is correct; during implementation, business module fields must declare the minimum interface and must not elevate business rules back into the i18n package.
- Task 8 `ResourceLoader`: Adjusted. The shared loader cannot be placed in `internal/service/i18n`, otherwise apidoc would reverse-depend on the runtime i18n service package just to reuse the loader; changed to `pkg/i18nresource` public component.
- Task 9 WASM parsing migration to `pluginbridge`: Retained. The WASM file format belongs to plugin bridge / plugin runtime infrastructure; moving it out of the i18n package reduces coupling.
- Tasks 10-12 Traditional Chinese and fixed LTR: Retained after adjustment. New language resources, fixed direction, and missing checks should not require business Go code, SQL seed, or frontend TS language enumeration changes; if adding a language requires modifying these enumerations, it should be treated as a design issue.
- FB-4 Runtime i18n persistence tables: Removed. The foundational i18n service only provides resource aggregation, distribution and validation, not database overrides or generic business content multilingual models.
- Review rules: Added i18n foundational component boundary checks, prohibiting `internal/service/i18n` from introducing business entity projectors, business key determination constants, or business-entity-named methods.

## Risks / Trade-offs

- [Risk] After Decision 1 removes cloning, if any business code assumes the returned map is writable and modifies it, the cache will be corrupted. -> Mitigation: Not a concern for codegen/lint phase; rely on grep `[]string\|map[string]string` write operations during refactoring and add unit tests; after refactoring, `Translate` series only returns `string`, with no map exposed to business callers.

- [Risk] Decision 2's sector cache refactoring involves multiple invalidation call sites (plugin runtime, i18n manage, apidoc loader), making it easy to miss during migration. -> Mitigation: Replace bare calls with `Maintainer.InvalidateRuntimeBundleCache(scope InvalidateScope)`, with scope explicitly passed by callers; after refactoring, use `lina-review` to validate all invalidation entry points carry scope parameters.

- [Risk] Decision 3's ETag negotiation may show stale translations to users when frontend persistence and backend bundleVersion are inconsistent (e.g., an invalidation did not trigger version increment). -> Mitigation: Every invalidation path must `version.Add(1)`, covered by review rules; persistent 7-day TTL provides fallback; `reloadActiveLocaleMessages(force=true)` still available for forced refresh.

- [Risk] Decision 5's interface splitting may cause downstream modules to change type signatures on a large scale, with high migration cost. -> Mitigation: `Service` still exists as a composite type; business modules changing field types to smaller interfaces is an optional optimization done per-module, not blocking feature delivery; subsequent review rules encourage but do not mandate it.

- [Risk] Decision 10's introduction of Traditional Chinese may cause `CheckMissingMessages` and E2E inspections to be permanently red due to missing translation resources. -> Mitigation: Traditional Chinese manifest completion must be an independent task in tasks.md, and any module copy change review rule requires three-language synchronization (`zh-CN` + `en-US` + `zh-TW`); CI `CheckMissingMessages` threshold for `zh-TW` matches `en-US`, with missing items blocking.

- [Risk] Automatic simplified-to-traditional conversion may produce individual terms that do not fully conform to Traditional Chinese usage conventions. -> Mitigation: This iteration first ensures resource key completeness, runtime discovery, and page usability; subsequent terminology proofreading only requires modifying `zh-TW` JSON resources.

- [Trade-off] Decision 7's `ResourceLoader` abstraction adds an extra layer of indirect calls to apidoc and runtime bundle loading flows. -> Accepted, because eliminating ~280 lines of duplicate implementation far outweighs the reading cost of one abstraction layer; the abstraction must be in stable public components like `pkg/i18nresource`, not in the runtime i18n service package.

- [Trade-off] Decision 10's fixed LTR sacrifices future configuration flexibility for automatic direction switching by language. -> Accepted, because the current host positioning prioritizes reducing configuration complexity and maintenance cost.

- [Trade-off] Decision 11's removal of database overrides means online translation hotfixing via API is no longer possible. -> Accepted, because the current iteration prioritizes reducing new language complexity; translation changes are done by modifying resource files and publishing; future hotfix capability should be designed as an optional plugin.

## Migration Plan

1. Performance Optimization (P1):
   - Rewrite `Translate*` to no longer clone; add benchmark unit tests; add cache hit rate assertions.
   - Refactor `runtimeBundleCache` to layered structure; migrate all `InvalidateRuntimeBundleCache()` call sites to `InvalidateScope`.
   - Backend implements `bundleVersion` and `ETag`; modify `runtime_messages.go` controller to read `If-None-Match`, write `ETag`, return 304.
   - Frontend `runtime-i18n.ts` switches to `requestClient`; implement `localStorage` persistence and 304 negotiation; add unit tests.

2. Consistency Convergence (P2):
   - Delete centralized `LocaleProjector` approach; refactor 5 `*_i18n.go` to maintain projection rules within each business module; delete `englishLabels`/`chineseLabels` in `sysconfig_i18n.go` and add corresponding `config.field.*` translation keys.
   - Implement `RegisterSourceTextNamespace`; `jobmgmt` registers in its own `init()`; delete `i18n_manage.go::isSourceTextBackedRuntimeKey` blacklist.
   - Split `Service` interface into four small interfaces; business module field types converge per module (menu/dict/sysconfig/jobmgmt/role/usermsg/apidoc/plugin).

3. Boundary Cleanup (P3):
   - Extract `ResourceLoader` in `pkg/i18nresource`; `apidoc_i18n_loader.go` and `i18n.go` resource loading functions share it; delete duplicate implementations.
   - Frontend `loadMessages` splits failure semantics; add unit tests covering weak network / timeout degradation.
   - WASM custom section parsing moved to `pluginbridge`; i18n and plugin runtime switch call paths; delete WASM utility functions in i18n.

4. Traditional Chinese + Fixed LTR:
   - Maintain default language, multi-language toggle, sorting and native names in the default config file's `i18n` section; delete `017-framework-i18n-improvements.sql` seed for `zh-TW`.
   - Fill in host and all source plugins' `manifest/i18n/zh-TW/*.json` and `manifest/i18n/zh-TW/apidoc/**/*.json`.
   - Fill in frontend `packages/locales/src/langs/zh-TW/*.json` and `apps/web-antd/src/locales/langs/zh-TW/*.json`.
   - `setI18nLanguage(locale)` fixed to set `<html dir="ltr">`; antd `ConfigProvider` fixed to inject `direction="ltr"`.
   - dayjs / antd / vxe derive third-party locale by language code convention.
   - `CheckMissingMessages` threshold for `zh-TW` matches `en-US`.
   - E2E adds `TC0124` covering Traditional Chinese language switching, `<html dir>` assertion, and key page text completeness.

5. FB-4 Single Source of Truth for Resources:
   - Delete `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` table creation SQL, re-initialize database and regenerate DAO/DO/Entity.
   - Delete i18n database overrides, business content table, import API and corresponding tests; retain export, missing checks, and diagnostic APIs.
   - Converge runtime cache sectors to host/source-plugin/dynamic-plugin, and update ETag and cache invalidation tests.
   - Update documentation and review rules, clarifying that adding new languages must not modify SQL, backend language enumerations, or frontend TS language lists.

6. Verification and Review:
   - `make test` full E2E passes (including new `TC0124`).
   - `lina-review` reviews P1/P2/P3/Traditional Chinese four groups of changes separately.
   - Benchmark report: `Translate` single call < 100ns (target); hot path 100-call total latency decrease >= 80%.

## Open Questions

- Is Decision 3's persistent TTL default of 7 days appropriate? Does it need to be configurable (exposed in `sys_config`)? Current conclusion: fixed at 7 days for the initial phase, with configuration opened if needed in the future.
- Should Decision 10's Traditional Chinese manifest be manually translated key by key, or temporarily use placeholders with subsequent proofreading? Current conclusion: the initial phase generates complete `zh-TW` resources via simplified-to-traditional conversion with manual correction of key metadata, but E2E must guarantee "no Simplified Chinese / English residuals displayed under Traditional Chinese" -- enforced through `CheckMissingMessages`.
- Should the plural API introduce a custom `$tn` hook, or directly use vue-i18n's built-in? Current tendency is to directly use vue-i18n's `t` plural syntax (`{ count, plural, zero {...} one {...} other {...} }`), initially only placing one or two examples on batch operation prompt copy, with subsequent expansion as needed.
