## Context

LinaPro already defines API contract rules for REST semantics, path parameters, and API documentation tags, but response DTO time fields do not yet have a consistent boundary. If the backend returns formatted strings, `*gtime.Time`, or `*time.Time` directly, the server fixes the time format, language, and time-zone interpretation before the frontend can apply page-specific or locale-specific formatting.

This change defines the governance rule for public HTTP JSON response DTOs and migrates the existing host and source-plugin response DTO instant fields that are part of the current LinaPro API surface. Internal database models, `entity`, `dao`, and `service` code can continue using Go time types. Only values crossing the HTTP response boundary need to be projected into stable JSON numeric timestamps.

## Goals / Non-Goals

**Goals:**

- Formalize that instant fields in API responses return Unix timestamps in milliseconds.
- Add response time-field checks to the manual RESTful/API DTO review checklist in `lina-review`.
- Distinguish instant fields from calendar-date fields so fields such as `birthday` are not incorrectly modeled as timestamps.
- Migrate existing public response DTO instant fields across the host and source plugins, including frontend type/display adjustments and `apidoc i18n` updates.

**Non-Goals:**

- Do not add a static scanning tool or automatic check command.
- Do not change database fields, internal service models, or GoFrame-generated `dao/entity` structures.
- Do not add runtime configuration, cache, permission, or data-permission logic.

## Decisions

### Decision 1: Instant fields use Unix timestamps in milliseconds

Public HTTP JSON response DTO fields that represent instants use `int64` or `*int64` and carry Unix timestamps in milliseconds. Milliseconds work naturally with frontend runtimes and avoid ambiguous mixtures of seconds, milliseconds, and microseconds. Field documentation must include `Unix timestamp in milliseconds` so callers can understand the unit from the API contract.

The main alternative was continuing to return RFC3339 strings. That format is readable, but it still fixes serialization at the server boundary and does not fit the frontend requirement to localize formatting per page and user language.

### Decision 2: The rule applies only at the response DTO boundary

The rule does not require internal domain models to stop using `time.Time` or `gtime.Time`. Database reads, business calculations, comparisons, and GoFrame-generated structures can keep using time types. Controllers or services convert these values when projecting into public response DTOs.

This preserves internal time semantics while keeping the external JSON contract stable.

### Decision 3: Calendar dates are an explicit exception

Fields such as `birthday`, `businessDate`, and `periodDate` express business calendar dates, not exact instants. These fields may remain `YYYY-MM-DD` strings, but their documentation must state `date-only` semantics. Callers must not need to infer an instant at midnight in a server time zone.

### Decision 4: Manual review gate, no static scanner

Per the confirmed scope, this change does not introduce a static scanning tool. `lina-review` checks API DTO changes for time-field type, unit documentation, calendar-date exception documentation, and frontend, `apidoc i18n`, and test impact when existing field types are migrated.

### Decision 5: Existing API migration uses a shared projection helper

Existing host and source-plugin controllers project internal `time.Time`, `*gtime.Time`, or legacy textual time values through `lina-core/pkg/apitime`. This package is placed under `pkg` because the same response boundary projection is needed by host code and source plugins. It returns `*int64` values in Unix milliseconds while preserving nil/zero-value semantics for optional response fields.

## Risks / Trade-offs

- [Risk] Manual review can miss individual new fields. -> Mitigation: Document the rule in `AGENTS.md` and the `lina-review` checklist so API DTO changes must explicitly be checked.
- [Risk] Changing existing fields from strings to numbers can break frontend callers. -> Mitigation: This change updates frontend API types and displayed timestamp formatting for the migrated fields, and records review searches and compile/type validation.
- [Risk] The boundary between date fields and instant fields can be unclear. -> Mitigation: Date-only fields must document `date-only` semantics and should use names that make the date-only meaning clear.
- [Risk] Frontend display formats can still diverge. -> Mitigation: The backend provides one raw instant representation, and frontend code remains responsible for page-specific and locale-specific formatting.
