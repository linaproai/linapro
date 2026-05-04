## ADDED Requirements

### Requirement: Scheduled job default timezone must be configurable

The system SHALL read the default timezone for built-in cron jobs from configuration key `scheduler.defaultTimezone`, defaulting to `UTC`. Source code MUST NOT keep hard-coded constants such as `defaultManagedJobTimezone = "Asia/Shanghai"`.

#### Scenario: Missing configuration uses UTC

- **WHEN** the configuration file does not declare `scheduler.defaultTimezone`
- **AND** the service starts and registers built-in jobs
- **THEN** built-in jobs MUST use `UTC` as the default timezone

#### Scenario: Custom timezone takes effect

- **WHEN** the configuration file sets `scheduler.defaultTimezone: "Asia/Shanghai"`
- **AND** the service starts and registers built-in jobs
- **THEN** built-in jobs MUST use `Asia/Shanghai` as the default timezone

### Requirement: sys_job table must not use foreign key constraints

The system SHALL remove the `fk_sys_job_group_id` foreign key constraint from the `sys_job` table, maintain `group_id` to `sys_job_group` reference consistency in the application layer, and keep `KEY idx_group_id (group_id)` on `sys_job` for group-based query and cleanup paths. Other association tables in this repository rely on application-level consistency, and this table MUST follow that convention to avoid extra foreign-key lock overhead in high-concurrency scheduler scenarios.

#### Scenario: sys_job table no longer contains foreign key constraints

- **WHEN** `make init` completes database initialization
- **THEN** the `sys_job` table MUST NOT contain `fk_sys_job_group_id` or any `FOREIGN KEY` constraint pointing to `sys_job_group`

#### Scenario: sys_job keeps group_id index

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_job` MUST include `idx_group_id` on column `group_id`

#### Scenario: Write path validates group_id reference consistency

- **WHEN** an upper-layer caller creates or updates a `sys_job` record with `group_id`
- **THEN** the service layer MUST validate that the referenced group exists
- **AND** validation failure MUST return a `bizerr` business error instead of relying on database foreign-key interception
