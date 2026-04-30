# Backend Hardcoded Chinese Audit

## Scope

The audit covered handwritten backend Go files, generated files, tests, plugin platform packages, source-plugin backends, export helpers, runtime projections, and API documentation sources.

## Categories

- Caller-visible business, authorization, validation, and user-facing errors.
- User-visible projections such as generated labels, status fallbacks, and runtime configuration reasons.
- User deliverables such as Excel headers, sheet names, and exported enum labels.
- Developer diagnostics that should be stable English text.
- Generated sources that must be fixed through SQL comments or generation inputs.
- Test fixtures and user-data examples that are counted separately.

## Decisions

Caller-visible errors must use `bizerr`. Plugin-owned messages remain in plugin resources. Generated DAO, DO, and Entity files are not manually edited. Developer diagnostics are English by default and wrapped structurally when they reach a caller boundary. Allowlist entries must describe category, reason, and applicability scope.

## Result

The archive records this audit as the baseline for ongoing scanner governance and future i18n reviews.
