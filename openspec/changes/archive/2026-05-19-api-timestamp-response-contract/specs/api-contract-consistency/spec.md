## ADDED Requirements

### Requirement: Response instant fields use Unix timestamps in milliseconds
Public HTTP JSON response DTO fields that represent exact instants SHALL return Unix timestamps in milliseconds. The JSON type MUST be numeric, and the Go DTO type MUST be `int64` or `*int64`. Public response DTOs MUST NOT directly expose `time.Time`, `*time.Time`, `gtime.Time`, `*gtime.Time`, or formatted strings for instant fields. This rule only constrains the API response DTO boundary and does not constrain internal `dao`, `entity`, `do`, or `service` models.

#### Scenario: Defining an instant response field
- **WHEN** a developer adds or modifies `createdAt`, `updatedAt`, `deletedAt`, `loginAt`, `startedAt`, `endedAt`, `expiredAt`, or another public HTTP JSON response DTO field that represents an exact instant
- **THEN** the field type is `int64` or `*int64`
- **AND** the field value represents a Unix timestamp in milliseconds
- **AND** the field `dc` tag or API documentation includes `Unix timestamp in milliseconds`

#### Scenario: Defining a calendar-date response field
- **WHEN** a developer defines `birthday`, `businessDate`, `periodDate`, or another public HTTP JSON response DTO field that represents only a calendar date and not an exact instant
- **THEN** the field may use a `YYYY-MM-DD` string
- **AND** the field `dc` tag or API documentation clearly states `date-only` semantics
- **AND** callers do not need to infer an exact instant in the server time zone from the field

#### Scenario: Migrating an existing time-field contract
- **WHEN** a developer plans to migrate an existing public API instant field from a string or time-object response to a Unix timestamp in milliseconds
- **THEN** the migration is evaluated through the active OpenSpec change that owns the API contract change
- **AND** frontend parsing, page formatting, `apidoc i18n` resources, and related test impact are evaluated together
