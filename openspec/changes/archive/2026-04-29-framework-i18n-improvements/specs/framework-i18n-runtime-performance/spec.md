## ADDED Requirements

### Requirement: Translation lookup hot path must avoid cloning the entire runtime message bundle
The host system SHALL let `Translate`, `TranslateSourceText`, `TranslateOrKey`, `TranslateWithDefaultLocale` and other single-value-returning translation lookup methods directly hold a read lock on the internal message bundle and look up values when the cache hits, and MUST NOT clone or copy the entire runtime message bundle. Only when a method semantically needs to return the message set to the caller (such as the runtime translation bundle API or message export API), the system MAY clone once before returning.

#### Scenario: Single key translation lookup does not clone entire message bundle on cache hit
- **WHEN** a business module calls `Translate(ctx, key, fallback)` and the current language runtime message cache already exists
- **THEN** the system only holds a read lock on the internal message bundle and looks up the value, directly returning the found string
- **AND** no `cloneFlatMessageMap` or equivalent full `map[string]string` copy is performed
- **AND** the caller still receives semantically consistent results as before

#### Scenario: Translation bundle retains clone semantics when returned externally
- **WHEN** a controller calls `BuildRuntimeMessages` or `ExportMessages` to return the message set to the frontend or for export
- **THEN** the system clones once before handing over the message set, ensuring the caller can safely hold it independently
- **AND** this clone does not corrupt or overwrite the internal cache

### Requirement: Runtime translation cache must support layered invalidation by language and sector
The host system SHALL organize the runtime translation message cache into a "language x sector (host / source-plugin / dynamic-plugin)" layered structure, and provide fine-grained invalidation capabilities by sector dimension. Any business-event-triggered invalidation MUST only clear the affected language or sector, and MUST NOT "clear all" caches for all languages and all sectors. The host core i18n must not introduce a database override sector or runtime business content cache; translation content uses development-time JSON/YAML resources as the single source of truth.

#### Scenario: Host resource invalidation only clears target language host sector
- **WHEN** a maintenance tool or test flow triggers `en-US` host resource cache invalidation
- **THEN** the system only clears the `en-US` language's host sector cache and merged view
- **AND** `zh-CN` and other enabled languages' caches retain their original values
- **AND** source plugin and dynamic plugin sectors in `en-US` do not need to be reloaded

#### Scenario: Dynamic plugin enable/disable only clears that plugin's relevant sectors
- **WHEN** a dynamic plugin is enabled, disabled, or upgraded
- **THEN** the system only clears the dynamic plugin sector cache and merged view related to that plugin
- **AND** host and unaffected plugins' translation data continues to hit cache
- **AND** during re-merge, only that plugin ID's resources are loaded or removed

### Requirement: Runtime translation bundle API must support ETag negotiation
The host system SHALL output an `ETag` header in the `/i18n/runtime/messages` API response, with a value derived from the current language and runtime translation bundle version (`bundleVersion`) that must differ when the version changes. The system MUST receive the `If-None-Match` header from requests and return `304 Not Modified` without carrying a message body when the value matches the current response ETag. Every sector cache invalidation MUST trigger `bundleVersion` auto-increment, ensuring different bundle contents for the same language have different ETags.

#### Scenario: Same bundle returns 304 on second request
- **WHEN** the frontend first requests the runtime translation bundle with `Accept-Language: en-US` and saves the returned `ETag`
- **AND** no cache invalidation has occurred on the backend between the two requests
- **AND** the frontend carries `If-None-Match` equal to the previous `ETag` in the second request
- **THEN** the backend returns `304 Not Modified` without carrying a message body

#### Scenario: ETag must change after translation resource changes
- **WHEN** any sector (host / source-plugin / dynamic-plugin) experiences cache invalidation
- **THEN** `bundleVersion` auto-increments
- **AND** the `ETag` returned on the next request for the same language differs from before
- **AND** requests carrying the old `If-None-Match` return `200` with the latest message body

### Requirement: Default management workbench must persist runtime translations via ETag and integrate auth chain
The default management workbench SHALL call the runtime translation bundle API through the unified `requestClient`, enabling it to participate in the authentication, error handling, and degradation chain. The frontend SHALL persist each successful response's `{locale, etag, messages, savedAt}` to `localStorage`, and on subsequent page loads or language switches, prioritize using persistent data for fast rendering, then negotiate with `If-None-Match` in the background. Persistent data MUST have a TTL of no more than 7 days, forcing a re-fetch when the TTL is exceeded.

#### Scenario: Zero-network language switching on subsequent page loads
- **WHEN** a user has successfully loaded the runtime translation bundle in a language and written it to persistent cache
- **AND** the user reopens the page or switches to that language within 7 days
- **THEN** the frontend directly uses persistent data to complete `vue-i18n` injection
- **AND** asynchronously negotiates with `If-None-Match` in the background; on `304` hit, does not update in-memory or persistent data

#### Scenario: Persistent data forces refresh when TTL expires
- **WHEN** the persistent entry's `savedAt` is more than 7 days from the current time
- **THEN** the frontend ignores persistent data and sends a request with empty `If-None-Match` or without that header
- **AND** after successful fetch, updates in-memory and persistent data, refreshing `savedAt`

#### Scenario: Degrades to persistent fallback when runtime translation fails
- **WHEN** the runtime translation bundle API fails due to network error, timeout, or server 5xx
- **AND** a valid entry for that language exists in persistent cache
- **THEN** the frontend uses the persistent entry to complete rendering, without blocking the page
- **AND** the frontend notifies the user through the unified degradation notification mechanism that translations may have version discrepancies
