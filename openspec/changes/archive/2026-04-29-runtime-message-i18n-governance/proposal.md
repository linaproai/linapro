## Why

The project already has runtime i18n resource loading, language negotiation, and frontend language pack aggregation in place, but both backend and backend implementation logic still contain large amounts of direct Chinese returns, mixed Chinese-English strings, or raw backend text being passed through. Error messages, import failure reasons, Excel export content, plugin bridging errors, plugin lifecycle results, some frontend page labels, and server log/audit display text — if they continue to use hardcoded languages, adding new languages in the future will only be possible through manual spot-checking of every location, and the same request will display inconsistently across different language environments.

This iteration needs to build on the existing i18n foundation by adding "runtime visible message governance" capabilities — clearly delineating boundaries between user-visible messages, machine-readable error codes, development/operations logs, and user input content, forming scannable, testable, and maintainable multilingual implementation constraints.

## What Changes

- Add runtime message i18n governance specifications, covering backend API errors, business validation failures, import/export files, plugin bridging/host service errors, plugin lifecycle results, frontend static page text, and request error pass-through.
- Establish a structured backend error model: business errors must contain a stable error code, translation key, English source message, parameters, and default fallback message; HTTP responses are localized by request language while preserving machine-readable error codes.
- Clarify the boundary between logs and user-visible messages: operations logs use stable English and structured fields, must not depend on localized text; audit/operation logs, import failure reports, and export files must project user-facing names, statuses, and results by request language.
- Define unified localization helper capabilities for Excel exports, import templates, and import failure reasons; prohibit directly hardcoding headers, statuses, genders, success/failure display text in `user_excel.go`, `dict_*`, plugin exports, and other business files.
- Define a unified error return contract for the plugin platform and plugin examples: plugin bridging protocol and host service calls return stable error codes and English developer source messages, and admin-visible results are localized via i18n keys, avoiding mixed Chinese-English strings like "network request URL 非法" or "解析 data list request tag 失败".
- The default management console frontend continues to prioritize `$t` and runtime language packs; for pages not yet integrated — such as monitoring pages, online users pages, plugin pages — add translation keys, and standardize the request interceptor's handling priority for backend `errorCode/messageKey/messageParams/message`.
- Add automated scanning and test gates, distinguishing comments, test fixtures, user data examples, technical protocol constants, and actual runtime-visible messages; when adding or modifying runtime messages, the `zh-CN`, `en-US`, `zh-TW` runtime language packs and plugin language packs must be updated in sync.
- This iteration only establishes the plan and specifications; it does not directly clean up all hardcoded implementations. Subsequent `/opsx:apply` will implement in batches per the task list.

## Capabilities

### New Capabilities

- `runtime-message-i18n-governance`: Identification of runtime-visible messages, error code and translation key modeling, localized responses, import/export localization, plugin error contracts, frontend pass-through handling, and automated governance.

### Modified Capabilities

- None. This iteration constrains all runtime-visible messages through a new cross-module governance capability, avoiding duplication of the same i18n rules across multiple business capability specifications.

## Impact

- **Backend host**: Affects `apps/lina-core/internal/service/**`, `apps/lina-core/internal/controller/**`, `apps/lina-core/pkg/pluginbridge/**`, `apps/lina-core/pkg/pluginfs/**`, `apps/lina-core/pkg/excelutil/**` and other error return, import/export, and plugin platform capabilities.
- **Source plugins**: Affects `apps/lina-plugins/*/backend/internal/service/**` for business errors, export headers, operation status mappings, and plugin example error responses; plugins must place runtime user-visible messages in their own `manifest/i18n/<locale>/*.json`.
- **Frontend console**: Affects `apps/lina-vben/apps/web-antd/src/api/request.ts`, monitoring/online-users pages, upload/export/plugin management pages for error display and hardcoded label handling.
- **Language resources**: Requires maintaining host and plugin runtime language packs, adding error messages, import/export headers, enum display, and frontend page text keys; does not reuse `apidoc` resources.
- **Testing and governance**: Requires adding hardcoded message scanning scripts, missing translation checks, backend error localization unit tests, import/export language tests, and key frontend page i18n tests.
- **Performance**: Localization lookups must reuse the existing runtime translation cache; must not repeatedly build the full language pack in error hot paths or batch export loops.
