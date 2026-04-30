# Backend Hardcoded Chinese I18n Governance

## Why

Backend review found handwritten non-test Go files with Chinese string literals that could reach HTTP responses, plugin responses, export files, management UI projections, or runtime configuration display. Existing runtime i18n work established the foundation, but the remaining backend cleanup needed a dedicated governance change.

## What Changes

- Establish a backend hardcoded-Chinese audit list and classify each finding by caller-visible error, user-visible projection, deliverable text, developer diagnostic, generated source, test fixture, or user-data example.
- Replace caller-visible Chinese errors with module-owned `bizerr` codes and runtime i18n resources.
- Localize backend-owned projections and deliverables such as department tree fallback labels, post exports, login or operation log exports, demo-mode rejection text, plugin summaries, cron shell reasons, and system information durations.
- Convert plugin-platform developer diagnostics to stable English text and wrap boundary errors structurally when needed.
- Govern generated schema text at SQL comments or generation inputs rather than hand-editing generated files.
- Add scanner gates and allowlist documentation to prevent regressions.

## Impact

The change affects lina-core backend services, plugin platform packages, source-plugin backends, host and plugin runtime i18n resources, generated schema sources, backend tests, plugin tests, and scanner tooling. It does not change the product boundary; it applies existing i18n and error-governance rules to concrete backend remnants.
