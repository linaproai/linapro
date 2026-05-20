## Why

API responses can currently expose time fields as formatted strings, `*gtime.Time`, or `*time.Time`, which makes frontend display inconsistent across languages, time zones, and page-specific formatting needs. The project needs to formalize the rule that API responses return raw timestamps for instants, and the frontend formats them according to the page and user locale.

## What Changes

- Add an API response time-field contract to the project API design rules: public HTTP JSON response DTO fields that represent instants must return Unix timestamps in milliseconds.
- Prohibit public response DTOs from directly exposing `time.Time`, `*time.Time`, `gtime.Time`, `*gtime.Time`, or formatted strings for instant fields.
- Clarify that the rule applies only at the API response DTO boundary and does not constrain internal `dao`, `entity`, or `service` models.
- Clarify that calendar-date fields such as `birthday`, `businessDate`, and `periodDate` may use `YYYY-MM-DD` strings, but their documentation must state the `date-only` semantics.
- Add this rule to the manual RESTful/API DTO review requirements in `lina-review`.
- Migrate existing host and source-plugin public response DTO instant fields to Unix millisecond timestamps, and update frontend display formatting and `apidoc i18n` descriptions accordingly.
- Do not add a static scanning tool.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `api-contract-consistency`: Extend the API response DTO contract so instant fields return Unix timestamps in milliseconds and become part of the review standard.

## Impact

- Affects the API design rules and backend API DTO constraints in `AGENTS.md`.
- Affects the RESTful/API DTO review checklist in `.agents/skills/lina-review/SKILL.md`.
- Affects existing host and source-plugin API response DTO field types for instant fields, plus response projection code that maps internal time values to Unix millisecond timestamps.
- Affects frontend API types and page-level time formatting where those fields are displayed.
- Affects host and source-plugin `zh-CN` `apidoc i18n JSON` descriptions for migrated time fields.
- Does not change database schema, runtime UI language packs, cache consistency behavior, data permissions, or development scripts.
