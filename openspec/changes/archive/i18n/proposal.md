## Why

The default delivery currently maintains three sets of i18n resources for Simplified Chinese, Traditional Chinese, and English, increasing the synchronization cost for host, plugin, frontend runtime language packs, and API documentation resources. The project only needs to retain English and Simplified Chinese by default, so Traditional Chinese default resources should be removed to reduce the i18n maintenance complexity of built-in capabilities and plugin examples.

## What Changes

- **BREAKING**: Default delivery no longer provides `zh-TW` Traditional Chinese runtime language, plugin manifest language packs, or API documentation translation resources.
- Default configuration `i18n.locales` only retains `en-US` and `zh-CN`.
- Default management workbench and shared frontend language packs only retain `en-US` and `zh-CN` static resources.
- Remove or adjust E2E/unit test assertions targeting `zh-TW`, while retaining English and Simplified Chinese language governance checks.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `framework-i18n-foundation`: Default built-in languages converge from `zh-CN`, `en-US`, `zh-TW` to `zh-CN` and `en-US`, removing Traditional Chinese runtime language lists, page content, API documentation, and test acceptance requirements.
- `management-workbench-i18n`: Chinese browser language tags (including `zh-TW`) continue to fall back to `zh-CN` on first visit, but the default workbench no longer provides `zh-TW` static language packs.

## Impact

- Affects host `apps/lina-core/manifest/i18n` resource directory and default configuration templates.
- Affects source plugin `apps/lina-plugins/*/manifest/i18n` resource directories.
- Affects default management workbench and shared frontend language packs `apps/lina-vben/**/locales`.
- Affects Traditional Chinese-specific E2E, frontend unit tests, backend i18n-related tests, and i18n static check scripts.
- Does not add new REST APIs, database schema, SQL seed, permission boundaries, or runtime cache mechanisms.
