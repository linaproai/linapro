## ADDED Requirements

### Requirement: sys_online_session must include last_active_time index

The system SHALL maintain `KEY idx_last_active_time (last_active_time)` on the `sys_online_session` table to support inactive-session cleanup queries by `last_active_time` range and avoid full table scans.

#### Scenario: Index exists

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_online_session` MUST include `idx_last_active_time` on column `last_active_time`

#### Scenario: Inactive-session cleanup uses the index

- **WHEN** the service executes cleanup queries of the form `WHERE last_active_time < ?`
- **THEN** the database MUST select `idx_last_active_time` to avoid a full table scan
