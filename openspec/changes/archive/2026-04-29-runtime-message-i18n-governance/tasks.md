## 1. Foundation Governance and Scanning

- [x] 1.1 Finalize runtime message classification rules based on `runtime-message-i18n-audit.md`, defining the criteria for `UserMessage`, `UserArtifact`, `UserProjection`, `DeveloperDiagnostic`, `OpsLog`, `UserData` and the allowlist format
- [x] 1.2 Add backend hardcoded runtime message scanning script, covering `gerror.New*`, `gerror.Wrap*`, `Reason/Message/Fallback`, export headers, status text, and plugin bridging error construction
- [x] 1.3 Add frontend hardcoded runtime message scanning script or ESLint rules, covering `title`, `label`, `placeholder`, template text, `message.*`, `notification.*`, `Modal.confirm`
- [x] 1.4 Integrate scanning commands into local validation entry points and ensure comments, test fixtures, user example data, technical units, and English operations logs do not cause false-positive blocks
- [x] 1.5 Add missing translation checks for new keys in `zh-CN`, `en-US`, `zh-TW` runtime language packs, ensuring host and plugin resources are validated separately

## 2. Backend Structured Error Infrastructure

- [x] 2.1 Add runtime message error model and construction helper supporting `errorCode`, `messageKey`, `messageParams`, English fallback, and GoFrame `gcode` semantics
- [x] 2.2 Update the unified response middleware to recognize structured errors and output localized `message`, stable `errorCode`, `messageKey`, and `messageParams`
- [x] 2.3 Preserve existing `LocalizeError` as a fallback for legacy errors, but prohibit new business paths from continuing to use Chinese free-text errors
- [x] 2.4 Add unit tests for structured error rendering, covering `zh-CN`, `en-US`, `zh-TW`, missing key fallback, and parameter formatting
- [x] 2.5 Confirm that error localization reuses the existing runtime translation cache and does not build the full language pack on a single error hot path

## 3. Host Business Error and Import/Export Cleanup

- [x] 3.1 Clean up user module errors, user import failure reasons, user export headers, user import templates, and gender/status enum text
- [x] 3.2 Clean up dictionary type, dictionary data, combined export, and dictionary import business errors, sheet names, headers, status text, and failure reasons
- [x] 3.3 Clean up user-visible errors in system parameters, config import/export, file management, user messages, roles, menus, and notification modules
- [x] 3.4 Clean up user-visible errors related to scheduled tasks, task handlers, task logs, task metadata, cache, distributed locks, and runtime parameters
- [ ] 3.5 Clean up fixed Chinese messages in plugin lifecycle, source plugin upgrades, auto-enable, dynamic plugin runtime, frontend resource parsing, and plugin governance results
- [ ] 3.6 Add `zh-CN`, `en-US`, `zh-TW` resources for new runtime language keys related to users, dictionaries, config, plugins, and tasks
- [ ] 3.7 Implement request-level batch localization context for import/export, ensuring only cache lookup and parameter formatting inside row loops

## 4. Plugin Platform and Source Plugin Cleanup

- [ ] 4.1 Clean up mixed Chinese-English errors in `pkg/pluginbridge` bridge codec, host call codec, and host service codec, replacing with stable error codes and English developer diagnostics
- [ ] 4.2 Clean up user-visible error contracts in `pkg/pluginfs`, `pkg/plugindb`, plugin data host, WASM host service, and catalog validation
- [ ] 4.3 Update dynamic plugin guest JSON error convention to support `errorCode`, `messageKey`, `messageParams`, and localized `message`
- [ ] 4.4 Clean up backend business errors and export text in `demo-control`, `content-notice`, `org-center`, `monitor-loginlog`, `monitor-operlog`, `plugin-demo-source`, `plugin-demo-dynamic`
- [ ] 4.5 Write plugin new runtime messages into each plugin's own `manifest/i18n/<locale>/*.json`; prohibit centralizing plugin runtime keys into the host language pack
- [ ] 4.6 Update plugin-related unit tests to assert on error codes, translation keys, and localized display rather than asserting fixed Chinese error text

## 5. Frontend Console and Plugin Frontend Cleanup

- [x] 5.1 Update `apps/lina-vben/apps/web-antd/src/api/request.ts` to prioritize `messageKey/messageParams` for request error display, with fallback to backend `message`
- [x] 5.2 Clean up server monitoring page static labels, tooltips, empty states, and time units, replacing all with `$t` or runtime language packs
- [x] 5.3 Clean up hardcoded Chinese in online users page query form and table columns
- [x] 5.4 Scan and clean up runtime-visible hardcoded Chinese in plugin frontend pages, preserving user data and test fixtures
- [x] 5.5 Add missing `zh-CN`, `en-US`, `zh-TW` translation keys in frontend static language packs and host runtime language packs
- [x] 5.6 Confirm that when `i18n.enabled=false`, the display still follows the default language, and the language switching hide logic is not affected by this error display refactoring

## 6. Testing and E2E

- [ ] 6.1 Add backend unit tests covering structured errors, runtime language resource missing, import failure reasons, and export header localization
- [ ] 6.2 Add plugin platform unit tests covering bridge/host service error codes, English developer diagnostics, and admin-side localization mapping
- [ ] 6.3 Add frontend unit tests covering request interceptor `messageKey` priority and fallback behavior
- [ ] 6.4 Create `hack/tests/e2e/i18n/TC0131-structured-error-localization.ts` to verify the same backend business error displays in different languages under `zh-CN`, `en-US`, `zh-TW` with stable error codes
- [ ] 6.5 Create `hack/tests/e2e/i18n/TC0132-localized-export-artifacts.ts` to verify user or dictionary export headers, statuses, and import failure reasons output in the current language
- [ ] 6.6 Create `hack/tests/e2e/i18n/TC0133-runtime-hardcoded-copy-regression.ts` to verify server monitoring page and online users page have no residual hardcoded Chinese after language switching
- [ ] 6.7 Run backend-related `go test`, frontend type check/build, hardcoded message scanning, and new E2E test cases

## 7. Documentation, Acceptance, and Review

- [x] 7.1 Update host and frontend i18n README to document runtime errors, import/export, plugin text, and hardcoded scanning governance rules, and maintain both English and Chinese README in sync
- [x] 7.2 Update `runtime-message-i18n-audit.md` to record remaining allowlist, cleaned modules, and follow-up observations after implementation
- [x] 7.3 Run `openspec status --change runtime-message-i18n-governance` to confirm proposal, design, spec, and task status are consistent
- [ ] 7.4 Call `/lina-review` to review implementation, spec compliance, i18n resource completeness, and test coverage

## Feedback

- [x] **FB-1**: Remove `bizerr` custom integer business error codes; restore interface `code` to GoFrame type error codes and govern business semantic codes by module namespace
- [x] **FB-2**: Add project specification and `lina-review` check requiring caller-visible interface errors to use `bizerr`, and fix clear violations in the current implementation
- [x] **FB-3**: Clean up hardcoded display text in built-in runtime parameter and public frontend parameter registries, and unify the default value source of truth
- [x] **FB-4**: Add structured metadata reading and error matching methods for `bizerr.Code`, and split `bizerr` implementation file responsibilities
- [x] **FB-5**: Clarify in the design document that i18n JSON is classified by locale directory, runtime semantic domain, and apidoc subdirectory, and plan the subsequent resource directory reorganization
- [x] **FB-6**: Implement host and plugin i18n JSON directory reorganization, update loader, dynamic plugin packaging, documentation, and tests
- [x] **FB-7**: Clean up residual Chinese hardcoded error messages in `apps/lina-core/internal/service/plugin/internal/runtime/artifact.go`
- [x] **FB-8**: Migrate runtime i18n checks from temporary Python scripts to a Go tool under `hack/tools/runtime-i18n`
- [x] **FB-9**: Add bilingual usage documentation for each tool directory under `hack/tools`
- [x] **FB-10**: Fix department code uniqueness, mock user association, and association table reverse query index issues in `org-center` plugin initialization SQL
- [ ] **FB-11**: Add optional database rebuild parameter to `make init`, change default database name to `linapro`, and add explicit idempotent database creation in initialization SQL
- [x] **FB-12**: Fix `plugin-demo-dynamic` standalone static page issue of built-in multilingual text, change to reuse plugin runtime i18n resources
- [x] **FB-13**: Clean up Chinese hardcoding in CLI and database initialization diagnostic errors in `apps/lina-core/internal/cmd`, unify to English developer diagnostics, and update unit test assertions
- [x] **FB-14**: Shorten English preference drawer tab display text to avoid `Appearance` and `Shortcut Keys` exceeding button background width
- [x] **FB-15**: Sort out host and source plugin seed/mock data boundaries, migrate or supplement demo test data to their respective `manifest/sql/mock-data` directories
