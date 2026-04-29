## Context

The `framework-i18n-foundation` and `framework-i18n-improvements` iterations have already provided the host with request language parsing, runtime language pack aggregation, cache optimization, ETag negotiation, source plugin/dynamic plugin resource merging, and missing translation checks. The unified response middleware has also been localizing error text by request language through `i18n.LocalizeError(ctx, err)`.

The problems identified in this audit are not about "whether translation infrastructure exists" but about "whether runtime-visible messages have entered the infrastructure." Currently, a large amount of business logic still writes Chinese or mixed Chinese-English strings directly into `gerror.New`, import failure `Reason`, Excel headers, plugin bridging errors, plugin lifecycle `Message`, and frontend page `title/label`. Some of these strings are returned directly to the frontend, some are written to export files, and some are returned through plugin protocols to developers or plugin runtimes. See `runtime-message-i18n-audit.md` for details.

Key existing constraints:

- The project has no legacy compatibility burden, so it can directly adjust unified response error fields, plugin bridging error payloads, and internal error construction patterns.
- OpenAPI/API DTO source text must remain English; API document i18n resources and runtime UI i18n resources must not be mixed.
- Runtime translation content uses `manifest/i18n/<locale>/*.json` and plugin's own `manifest/i18n/<locale>/*.json` as the source of truth; no new database translation tables will be introduced.
- Backend hot paths must reuse the existing runtime translation cache; must not rebuild the full language pack in error handling, list projection, or batch export loops.
- User input, user-maintained business names, and raw errors returned from external systems are not framework-translatable messages and must be preserved as-is or embedded as parameters in localized templates.

## Goals / Non-Goals

**Goals:**

- Establish runtime message classification rules, clarifying which strings must be i18n, which must remain English logs, and which are user data or developer diagnostics.
- Upgrade backend business errors from "free text" to structured messages: stable error codes, translation keys, parameters, English source messages, and HTTP/gcode semantics.
- Make the unified response and frontend request interceptor prioritize consuming structured error fields, while preserving a localized `message` for simple callers to display.
- Clean up high-risk Chinese hardcoding in user, dictionary, system parameter, plugin, task, file, notification, organization, monitoring, and other modules.
- Make import templates, import failure reasons, export sheet names, headers, and enum values parse once by request language and be reusable.
- Give plugin bridging and host service errors stable machine codes; developer diagnostics default to English, and admin display is localized via i18n keys.
- Add `$t` / runtime language pack consumption rules for frontend pages, plugin frontends, and request error display.
- Add automated scanning and missing translation tests to prevent future hardcoded messages.

**Non-Goals:**

- No online translation editing backend, machine translation, or third-party translation platform.
- No localization of operations logs by user language; logs use stable English and structured fields.
- No translation of user input content, database user-maintained names, external system raw responses, or technical parameters such as SQL/protocol/file paths.
- No RTL layout or per-user language preference support in this iteration.
- No language registration by modifying old SQL seeds; languages continue to be driven by the `manifest/i18n/<locale>/` directory and default configuration metadata.

## Decisions

### Decision 1: Classify runtime messages by usage surface

**Choice**: Divide strings in code into six categories, each with a specified handling strategy:

- `UserMessage`: API errors, business validation, frontend toasts, admin result prompts — must use error codes or translation keys.
- `UserArtifact`: Excel headers, sheet names, import template examples, import failure reasons, export enum values — must render by request language.
- `UserProjection`: Menus, dictionaries, roles, tasks, audit/operation logs and other backend-owned display data — must use stable business keys for projection.
- `DeveloperDiagnostic`: Plugin protocol, WASM host call, manifest validation, CLI diagnostics — must have stable machine codes, default English source messages; localized when entering the admin console.
- `OpsLog`: Server logs and metrics — use English and structured fields, do not participate in runtime i18n.
- `UserData`: User input, external system content, database business values — preserved and returned as-is, not auto-translated.

**Rationale**: A single string cannot simultaneously serve log retrieval, frontend display, and protocol status determination. Classification first avoids cramming all strings into i18n and also avoids mistranslating user data.

**Alternative**: Scan all Chinese and replace everything with translation keys. Not adopted because it would damage comments, tests, user examples, protocol constants, and operations logs, with high maintenance cost and unclear semantics.

### Decision 2: New structured runtime message error model

**Choice**: Introduce unified business error construction capability in `apps/lina-core/pkg/bizerr`. `bizerr` only defines business semantics, runtime i18n metadata, and the closest GoFrame type error code; it no longer assigns custom integer response codes for each business error. The `code` field in unified responses returns to GoFrame `gcode.Code` type codes, such as `CodeInvalidParameter`, `CodeNotFound`, `CodeNotAuthorized`; specific business semantics are expressed through `errorCode`, `messageKey`, and `messageParams`.

```go
type Meta struct {
    ErrorCode string     // Machine-readable business semantic code, e.g. USER_NOT_FOUND
    MessageKey string    // Runtime i18n key, e.g. error.user.not.found
    Fallback  string     // English source message
    TypeCode  gcode.Code // GoFrame type error code, e.g. gcode.CodeNotFound
}
```

Each component defines business errors centrally in a dedicated `*_code.go` file, for example:

```go
var CodeDictTypeExists = bizerr.MustDefine(
    "DICT_TYPE_EXISTS",
    "Dictionary type already exists",
    gcode.CodeInvalidParameter,
)
```

Business code only references component error definitions and does not directly write `errorCode`, `messageKey`, or raw numbers:

```go
return 0, bizerr.NewCode(CodeDictTypeExists)
```

All interface errors that may enter HTTP API, plugin calls, source plugin backend interfaces, WASM host service, or other caller response payloads must be created/wrapped through `bizerr.NewCode`, `bizerr.WrapCode`, or equivalent wrappers. `bizerr` internally uses GoFrame `gerror` to create actual error objects, so it retains stack traces, causes, and `gcode.Code` capabilities; direct `gerror.New*`, `errors.New`, or `fmt.Errorf` is only allowed as internal diagnostics not exposed to callers, or as low-level causes immediately wrapped by `bizerr.WrapCode`.

Business semantic identifiers must be governed by module namespace. Host modules use `<MODULE>_<CASE>` format, e.g. `USER_NOT_FOUND`, `DICT_TYPE_EXISTS`; plugin modules use `<PLUGIN>_<MODULE>_<CASE>` or equivalent prefix, e.g. `ORG_CENTER_DEPT_NOT_FOUND`. Business error definitions for the same module are maintained centrally in that module's `*_code.go`; business call sites are prohibited from writing strings directly or reusing business error definitions with inconsistent semantics across modules.

The unified response middleware identifies this metadata and outputs:

```json
{
  "code": 65,
  "message": "用户不存在",
  "errorCode": "USER_NOT_FOUND",
  "messageKey": "error.user.not.found",
  "messageParams": {"id": 12},
  "data": null
}
```

`code` is the GoFrame type error code expressing the error category; `message` is the display text resolved by the server according to request language; `errorCode` and `messageKey` are for frontend, tests, plugins, or third-party callers to make stable business semantic judgments.

**Rationale**: Currently `LocalizeError` can only treat `err.Error()` as a key or source text for translation; it cannot stably carry machine codes, parameters, and fallback, nor can it distinguish user-visible messages from developer diagnostics. The structured model preserves the existing middleware entry point while completing the contract.

**Alternative**: Continue with the convention of `gerror.New("error.user.notFound")`. Not adopted because pure string keys cannot enforce parameter completeness and cannot prevent continued Chinese free-text in business code.

### Decision 3: Language resources split by locale directory and semantic domain, no new database tables

**Choice**: All new runtime messages are written to the host or plugin's own `manifest/i18n/<locale>/*.json`. Language code is the first-level directory, and runtime resources are partitioned within the same language directory alongside API document resources:

```text
manifest/i18n/
  en-US/
    framework.json
    menu.json
    dict.json
    config.json
    error.json
    artifact.json
    public-frontend.json
    apidoc/
      common.json
      core-api-user.json
  zh-CN/
    framework.json
    menu.json
    dict.json
    config.json
    error.json
    artifact.json
    public-frontend.json
    apidoc/
      common.json
      core-api-user.json
```

Source plugins use the same layout, with resources still owned by the plugin directory:

```text
apps/lina-plugins/<plugin-id>/manifest/i18n/
  en-US/
    plugin.json
    menu.json
    error.json
    page.json
    artifact.json
    apidoc/
      plugin-api-main.json
  zh-CN/
    plugin.json
    menu.json
    error.json
    page.json
    artifact.json
    apidoc/
      plugin-api-main.json
```

Runtime JSON files are classified by semantic domain:

| File | Content Scope |
| ---- | ------------- |
| `framework.json` | Framework name, framework description, language name, host common display text |
| `menu.json` | Host or plugin menu, route titles, navigation-related text |
| `dict.json` | Dictionary type names, dictionary item labels, backend built-in enum display text |
| `config.json` | System config item names, descriptions, groups, public frontend config display text |
| `error.json` | `bizerr` business errors, validation errors, caller-visible failure reasons |
| `artifact.json` | Import/export sheet names, headers, template fields, failure reasons, artifact enum display |
| `public-frontend.json` | Login page, public pages, default console common text |
| `plugin.json` | Plugin name, plugin description, plugin metadata and plugin common text |
| `page.json` | Plugin frontend page titles, forms, tables, buttons, prompts |

When a semantic domain grows too large, it can be further split by business module, e.g. `user.json`, `job.json`, `plugin-lifecycle.json`, but file names must express business meaning and must not use numeric sequence prefixes like `00-`, `10-`. Load order is only used to ensure deterministic results, not to express business coverage relationships; when duplicate keys appear within the same locale, validation must fail.

The loading boundary between runtime and API document resources is as follows:

- Runtime i18n only loads `manifest/i18n/<locale>/*.json`, does not recurse into `apidoc/`.
- API document i18n only loads `manifest/i18n/<locale>/apidoc/**/*.json`.
- Host API document files use `common.json` and `core-api-<module>.json` naming.
- Plugin API document files use `plugin-api-<module>.json` or `<plugin-id>-api-<module>.json` naming.
- `en-US` API documents still use DTO English source text as the source of truth; `apidoc` files only contain necessary empty placeholders or non-source-text supplements, not re-maintained English mappings.

Runtime keys continue to be governed by existing namespaces:

- Host errors: `error.<domain>.<case>`
- Host import/export, templates and artifacts: `artifact.<module>.<section>.<field>`
- Host business projection and dictionary built-in text: maintained per existing runtime namespaces, e.g. `dict.<domain>...`, `user.<domain>...`
- Frontend pages: continue using `pages.<module>...`
- Plugin text: `plugin.<plugin-id>...`

English source messages are preserved in code fallbacks for developer understanding and resource-missing fallback; `zh-CN`, `en-US`, `zh-TW` directories must all contain corresponding runtime keys. `apidoc` resources continue to serve only API documentation and do not participate in runtime errors or UI display. Dynamic plugin source resources can be split for maintenance, but when packaged as Wasm runtime artifacts, they are still merged by locale into runtime i18n assets and apidoc i18n assets to avoid changing the runtime protocol format.

**Rationale**: Existing i18n improvements have clearly removed the runtime translation database dual-source model. Continuing to use file resources keeps deployment, caching, and missing checks simple; after splitting by locale directory, adding a new language only requires adding a new language directory, and maintainers can handle both runtime messages and API document messages in the same directory.

**Alternative**: Keep `manifest/i18n/<locale>.json` as a single file. Not adopted because as the number of runtime keys grows, the file becomes too large, with continuously increasing costs for lookup, review, and conflict identification.

**Alternative**: Write error messages to dictionary or configuration tables. Not adopted because error messages belong to code contracts and have a different lifecycle from runtime-editable business configurations.

### Decision 4: Import/export uses batch localization context

**Choice**: Build a request-level `MessageRenderer` or equivalent narrow interface for import/export. When entering the export/import flow, parse the current locale and all keys needed by the module in one shot; within the loop, only perform map lookups and parameter formatting. `excelutil` continues to only handle Excel file operations and does not directly depend on the i18n service; the business service is responsible for passing in localized sheet names, headers, enum text, and failure reasons.

**Rationale**: Exports may have 10,000 rows; if every cell calls the full translation chain, it amplifies lock contention and string processing costs. Batch pre-fetching makes localization cost proportional to field count, not row count.

**Alternative**: Auto-translate strings in `excelutil.SetCellValue`. Not adopted because the Excel utility layer cannot know whether a string is a translation key, user data, or technical value, and would mistranslate.

### Decision 5: Plugin bridging errors split into protocol layer and display layer

**Choice**: Plugin bridging, WASM host call, and host service protocol return stable status codes and error codes; default error source messages use English developer diagnostics. When these errors enter the admin UI or host unified response, they are rendered using `messageKey` and locale. Dynamic plugin guest-returned JSON errors should also support `errorCode/messageKey/messageParams/message`, and the host preserves structured fields when passing through.

**Rationale**: Plugin protocols are stable contracts between developers and runtime; they cannot depend on a natural language string to determine errors. But admin users still need localized prompts.

**Alternative**: Host call payload directly returns localized Chinese or English. Not adopted because guest plugins may not share the same language as the current admin user, and protocol-layer localization would cause instability in debugging and testing.

### Decision 6: Frontend request interceptor prioritizes consuming structured fields

**Choice**: Default console request error handling priority adjusted to:

1. If the backend returns `messageKey`, the frontend renders using `$t(messageKey, messageParams)`.
2. Otherwise, use the backend's `message` already localized by request language.
3. Otherwise, use the request library's default error text.

Page-level messages must use `$t` or runtime language packs; directly writing Chinese or mixed Chinese-English strings in user-visible locations such as `title`, `label`, `placeholder`, `Modal.confirm`, `message.*` is prohibited.

**Rationale**: The backend unified response provides localized `message` for simple callers, but the frontend has a more complete language runtime and dynamic switching capability. Prioritizing `messageKey` avoids the situation where the page language has been switched but old request messages still remain in the old language.

**Alternative**: Frontend fully trusts backend `message`. Not adopted because the frontend already has runtime language packs and `$t`; structured fields are more conducive to language switching and testing.

### Decision 7: Logs remain English, audit display projected by language

**Choice**: `logger.*` operations logs uniformly use English fixed templates and structured fields; when business errors enter logs, they record `errorCode/messageKey` and key parameters, not the localized user display text as the sole information. Operation logs, login logs, task logs, plugin upgrade results, and other user-facing display data store stable codes and parameters, and are projected by request language for lists, details, and exports.

**Rationale**: Logs are machine-retrieval and cross-team troubleshooting assets; localizing by user language reduces stability. Audit display is a user interface and should be localized.

**Alternative**: Logs also output by request language. Not adopted because async tasks, startup flows, and multi-user shared logs have no reliable single request language.

### Decision 8: Scanning gates focus on "runtime-visible positions," no blanket raw-character rules

**Choice**: Add a Go tool under `hack/tools/runtime-i18n` to maintain runtime message scanning and language pack coverage checks. The tool uses auditable rule patterns to identify high-risk positions in phases, and can later evolve within the same tool to Go AST and frontend AST/ESLint rules:

- Go: `gerror.New*`, `gerror.Wrap*`, `panic(gerror...)`, `Reason/Message/Fallback` fields, export header arrays, status text mappings, plugin bridging error construction.
- Vue/TypeScript: `title/label/placeholder`, template text nodes, `message.*`, `notification.*`, `Modal.confirm`, table column definitions.
- Allowlist: comments, test fixtures, user example data, technical units, protocol constants, English operations logs.

Scan results enter CI or local `make` checks; new exceptions must state the reason and responsible module in `hack/tools/runtime-i18n/allowlist.json`.

**Rationale**: A simple `rg "\p{Han}"` produces too many false positives and cannot be sustained long-term. Semantic-position scanning can serve as a maintainable gate.

**Alternative**: Rely only on code review. Not adopted because current hardcoding is widespread and manual review alone is prone to regression.

## Risks / Trade-offs

- **Risk: Error response field changes affect frontend and plugin callers** -> The project has no compatibility burden, so the unified response can be upgraded directly; the `message` field is preserved to reduce caller cost.
- **Risk: Translation key count grows rapidly** -> Use namespace conventions and missing checks; split resource files by locale directory and semantic domain; plugin messages go into each plugin's own manifest.
- **Risk: Batch export localization affects performance** -> Use request-level batch pre-fetching and map lookups; prohibit building bundles inside row loops.
- **Risk: Scan rule false positives or misses** -> First phase outputs warn/report to generate lists; after stabilization, switch to blocking; allowlist must carry classification and reason.
- **Risk: Developer diagnostics and user prompts are confused** -> Plugin bridging protocol defaults to English developer messages; admin display goes through `messageKey` separately.
- **Risk: Underlying database/driver errors are not translatable** -> Only translate outer business templates; keep underlying errors as technical parameters or hide them in user prompts.

## Migration Plan

1. Add structured error and message rendering base components, integrate into the unified response middleware, and keep existing `LocalizeError` as a fallback for unstructured errors.
2. Update the frontend request interceptor to support `messageKey/messageParams/errorCode`, and add page-level `$t` rules.
3. Clean up host core API errors by priority: user, dictionary, system parameter, user message, file, task, plugin lifecycle.
4. Clean up import/export: user, dictionary, system parameter, post, operation log headers, sheets, enum values, and failure reasons.
5. Clean up plugin platform: pluginbridge, pluginfs, plugindb, catalog, runtime, wasm host service error codes and English developer diagnostics.
6. Clean up source plugin business errors and plugin frontend text; write resources into each plugin's own `manifest/i18n/<locale>/*.json`.
7. Reorganize host and plugin i18n resource directories: migrate runtime resources to `manifest/i18n/<locale>/*.json` and API document resources to `manifest/i18n/<locale>/apidoc/**/*.json`.
8. Establish scanning scripts and test gates; add missing translation, error localization, Excel language, and frontend page i18n tests.
9. Run full backend tests, frontend build, and relevant E2E tests; confirm that key errors and export content are consistently controlled under `zh-CN`, `en-US`, `zh-TW`.

Rollback strategy: This iteration is mainly code and resource governance and does not involve database migration. If a batch of cleanups introduces issues, the corresponding error construction and language resources can be rolled back per module; new unified response fields can be preserved without affecting callers that only read `message`.

## Open Questions

- Whether the plugin bridging protobuf structure should directly extend `errorCode/messageKey/messageParams` fields, or first carry JSON error objects in the payload, needs to be confirmed during implementation based on current ABI compatibility strategy.
- The locale source for command-line output — whether to use environment variables, configuration default language, or explicit parameters — needs to be confirmed when handling `cmd_*` files.
- How much detail of underlying database errors to show in user prompts needs to be balanced by security and diagnosability in specific modules.
