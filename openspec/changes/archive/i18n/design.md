## Context

The current i18n foundation supports language discovery through `manifest/i18n/<locale>` directories, with default configuration maintaining enabled languages, sorting, and native names through `i18n.locales`. The host, default management workbench, shared frontend packages, and source plugin examples all provide `zh-TW` resources, and some tests use Traditional Chinese as a default delivery acceptance target.

This change does not remove the runtime i18n framework's multi-language extension capability, but adjusts LinaPro's default delivery content: only English and Simplified Chinese resources are maintained by default. Subsequent projects can still add other languages following the existing resource directory conventions.

## Goals / Non-Goals

**Goals:**

- Delete `zh-TW` default translation resources from host, source plugins, default management workbench, and shared frontend packages.
- Converge the built-in language list in default configuration to `en-US` and `zh-CN`.
- Clean up Traditional Chinese-specific tests, static checks, and description text, avoiding default CI or local validation continuing to require `zh-TW`.
- Maintain the "discover languages through resource directories and configuration" extension mechanism unchanged.

**Non-Goals:**

- Do not delete runtime i18n API, language discovery mechanism, cache mechanism, or ETag negotiation capability.
- Do not add database tables, SQL seed, Go language enumerations, or frontend hardcoded language lists.
- Do not migrate third-party language resources that user-customized projects may have added.
- Do not change the behavior of Chinese browser first visit defaulting to `zh-CN`.

## Decisions

### Direct deletion of default zh-TW resource directories

Delete the default `zh-TW` resource directories rather than retaining empty directories or placeholder JSON files.

Reason: The authoritative source for language registration is the resource directory and the `i18n.locales` allowlist; retaining placeholder directories creates ambiguity for language discovery and maintenance checks.

Alternative: Retain empty directories but disable from configuration. This approach still leaves a default resource skeleton that needs explanation and maintenance, which does not align with simplification goals.

### Frontend static language packs only retain en-US and zh-CN

Since the default delivery no longer supports Traditional Chinese, static packs should be deleted synchronously.

Reason: The default management workbench startup and offline fallback require static language packs; since the default no longer supports Traditional Chinese, static packs should also be deleted.

Alternative: Retain frontend `zh-TW` but delete backend resources. This approach creates inconsistency between the language switcher, runtime language list, and static pack availability.

### Traditional Chinese-specific E2E directly removed

General i18n tests are changed to cover `zh-CN` and `en-US` only.

Reason: Deleted languages should not continue as default project acceptance items; retaining tests would force subsequent changes to continue maintaining non-existent default resources.

Alternative: Convert Traditional Chinese tests to skipped. Skipped tests retain expired governance noise.

### Default configuration no longer lists zh-TW, but language discovery and addition flow unchanged

AGENTS constraints require new built-in languages to be added through resources and configuration metadata, prohibiting new Go enumerations, SQL seed, or frontend language lists; this removal follows the same boundary.

## Risks / Trade-offs

- Default delivery can no longer directly switch to Traditional Chinese -> Clarify through specifications and configuration that this is an intentional default scope convergence; projects needing Traditional Chinese can add it following resource directory conventions.
- Tests or scripts may contain implicit `zh-TW` references -> Verify default resource and test reference cleanup through static scanning of `zh-TW`, `Traditional Chinese` and related keywords.
- Directory deletion may affect glob loading order or static checks -> Run frontend `i18n:check`, typecheck, and related unit tests to confirm only `zh-CN` alignment with `en-US` is required.
- Cache consistency risk is low -> No new cache keys or runtime invalidation paths are added; runtime language list still uses configuration and resource directories as authoritative source.
