# Cron Handler Registry Specification

## Purpose

Define the scheduled-job handler registry, parameter schema contract, plugin built-in job handler lifecycle, dynamic plugin cron declaration contract, and handler execution behavior.

## Requirements

### Requirement: Handler registry

The system SHALL provide a unified scheduled-job handler registry that manages available host-side and plugin-side handlers and acts as the single source of truth for scheduling, UI options, and parameter validation.

#### Scenario: Register handler

- **WHEN** the host or a plugin registers a handler through `Register(ref, def)`
- **THEN** the system SHALL store the handler definition in the in-memory registry
- **AND** `ref` uses a globally unique form such as `host:<name>` or `plugin:<pluginID>/cron:<name>`
- **AND** `def` includes at least `DisplayName / Description / ParamsSchema / Source / Invoke`

#### Scenario: Duplicate registration conflict

- **WHEN** a caller attempts to register an existing `ref`
- **THEN** the system SHALL return an explicit error
- **AND** SHALL NOT overwrite the registered handler

#### Scenario: Unregister handler

- **WHEN** the host or a plugin calls `Unregister(ref)`
- **THEN** the system SHALL remove that handler from the in-memory registry
- **AND** trigger job status cascade handling

### Requirement: Handler parameter JSON Schema

The system SHALL require handlers to declare parameters with the supported scalar subset of `JSON Schema draft-07` so parameters can be validated and UI forms can be rendered dynamically.

#### Scenario: Only supported schema subset is accepted

- **WHEN** the host or a plugin registers a handler `ParamsSchema`
- **THEN** the root node SHALL be `type=object`
- **AND** field types SHALL only include `string / integer / number / boolean`
- **AND** supported keywords SHALL only include `properties / required / description / default / enum / format`
- **AND** the system SHALL reject unsupported structures such as `array`, nested `object`, `$ref`, `allOf`, `anyOf`, `oneOf`, `not`, and `patternProperties`

#### Scenario: Validate parameters when creating a job

- **WHEN** a caller creates or updates a job with `task_type=handler`
- **THEN** the system SHALL validate `params` against the selected handler JSON Schema
- **AND** return field-specific errors when validation fails

#### Scenario: Validate parameters during execution

- **WHEN** a handler is about to execute
- **THEN** the system SHALL validate `params_snapshot` against the latest schema again
- **AND** when validation fails, write a `status=failed` log and stop that execution

#### Scenario: UI option data

- **WHEN** the frontend calls `GET /job/handler`
- **THEN** the system SHALL return all registered handlers
- **AND** the list includes `ref / displayName / description / source / pluginId`

#### Scenario: UI fetches schema details

- **WHEN** the frontend calls `GET /job/handler/{ref}`
- **THEN** the system SHALL return the complete `ParamsSchema` for that handler
- **AND** the frontend can render form controls from that schema

### Requirement: Plugin built-in job handler lifecycle

The system SHALL subscribe to plugin install, enable, disable, and uninstall events and synchronize the handler registry, plugin built-in scheduler entries, and related job projection state. The execution source for plugin built-in jobs SHALL be plugin declarations and the handler registry. `sys_job.is_builtin=1` projection rows SHALL be used for display, log linkage, and status governance, and MUST NOT be used as the registration source for startup persistent scanning.

#### Scenario: Lifecycle callbacks and response boundary

- **WHEN** a plugin install, enable, disable, or uninstall request succeeds
- **THEN** the system SHALL synchronize the handler registry, built-in job projection, and scheduler entries through explicit lifecycle callbacks in the same request chain
- **AND** the system SHALL NOT depend on a separate best-effort asynchronous event bus for later compensation

#### Scenario: Plugin install creates built-in job projections

- **WHEN** a plugin is installed successfully but is not yet enabled
- **THEN** the system SHALL first synchronize declared built-in scheduled jobs for that plugin into `sys_job`
- **AND** if the corresponding handler is not currently available, those jobs SHALL enter `paused_by_plugin` directly
- **AND** later plugin enablement SHALL restore executability through the handler registry
- **AND** the install phase MUST NOT register executable scheduler entries solely from the `sys_job` projection

#### Scenario: Plugin enable restores built-in job handlers

- **WHEN** a plugin is enabled
- **THEN** the system SHALL read that plugin's declared built-in scheduled jobs
- **AND** register corresponding handlers only in the `plugin:<pluginID>/cron:<name>` form
- **AND** restore jobs with `stop_reason=plugin_unavailable` to `status=enabled`
- **AND** register scheduler entries from the plugin declaration and current projected `sys_job.id`

#### Scenario: Plugin disable unregisters handlers and cascades jobs

- **WHEN** a plugin is disabled
- **THEN** the system SHALL unregister all handlers for that plugin
- **AND** unregister all built-in scheduler entries for that plugin
- **AND** set jobs with `handler_ref=plugin:<pluginID>/cron:*` and `status=enabled` to `paused_by_plugin`
- **AND** write `stop_reason=plugin_unavailable` on the corresponding jobs

#### Scenario: Plugin uninstall preserves job data

- **WHEN** a plugin is uninstalled
- **THEN** the system SHALL perform the same cascade as disable, marking jobs `paused_by_plugin` and unregistering scheduler entries
- **AND** it SHALL NOT delete existing jobs or historical logs
- **AND** the UI SHALL clearly highlight the job and indicate that the handler is unavailable

#### Scenario: UI visibility

- **WHEN** the job list returns a `paused_by_plugin` job
- **THEN** the frontend SHALL explicitly indicate in the status column that the plugin handler is unavailable
- **AND** disable Run Now and Enable buttons

### Requirement: Dynamic plugin scheduled job declaration contract

The system SHALL provide an independent scheduled-job declaration contract for dynamic plugins and keep it separate from the runtime host service boundary. Scheduled jobs declared by dynamic plugins SHALL be the execution source for plugin built-in jobs. The host SHALL synchronize them as `sys_job.is_builtin=1` projections and register or unregister scheduler entries according to plugin lifecycle.

#### Scenario: Dynamic plugin registers scheduled jobs through cron host service code

- **WHEN** a dynamic plugin needs to provide built-in scheduled jobs
- **THEN** the plugin SHALL use an independent `cron` host service in guest code and call `pluginbridge.Cron().Register(...)` to submit job metadata and guest-handler binding information
- **AND** the host SHALL run a controlled discovery through the reserved guest registration entry point to collect these declarations
- **AND** collected declarations SHALL enter the unified job projection path
- **AND** collected declarations SHALL be the runtime registration source for plugin built-in jobs

#### Scenario: Runtime host service remains focused

- **WHEN** the host exposes the guest-side `runtime` host service SDK
- **THEN** its responsibility SHALL remain limited to runtime logs, status, and lightweight information queries
- **AND** it SHALL NOT directly own the registration governance entry point for dynamic-plugin scheduled jobs
- **AND** if future plugin-side job governance is needed, it should be added through an independent `cron` or `job` host service

#### Scenario: Dynamic plugin authorization page displays cron registration details

- **WHEN** the dynamic plugin authorization review page is opened before install or enable, and the current release has declared the `cron` host service and successfully discovered scheduled-job contracts
- **THEN** the frontend SHALL display that service name as Task Service
- **AND** show summary information under that card, including registered job name, expression, scheduling scope, and concurrency policy
- **AND** governance target summary labels for data, storage, network, and similar services SHALL use concise resource-type labels such as Data Table, Storage Path, and Access URL without request or authorization prefixes
- **AND** the `runtime` service card SHALL be listed at the bottom of host services, after Task Service
- **AND** the Task Service card SHALL directly show the task panel without rendering an additional scheduled-job summary tag row
- **AND** each task-panel property title, such as expression, scheduling scope, and concurrency policy, SHALL be emphasized in bold
- **AND** background colors for the authorization-page checklist and resource-type labels SHALL match the corresponding labels on the details page
- **AND** the plugin details page label Current Effective Scope SHALL be normalized to Effective Scope

### Requirement: Handler execution contract

The system SHALL require every handler to support context cancellation and exit as soon as practical after cancellation.

#### Scenario: Handler accepts context

- **WHEN** the system calls handler method `Invoke(ctx, params)`
- **THEN** the handler SHALL check `ctx.Done()` during long blocking operations or pass `ctx` downstream
- **AND** return `ctx.Err()` as soon as practical after receiving cancellation

#### Scenario: Handler returns structured result

- **WHEN** a handler executes successfully
- **THEN** the handler SHALL return a JSON-serializable result
- **AND** the system SHALL write the result into `sys_job_log.result_json`

#### Scenario: Handler returns error

- **WHEN** handler execution returns a non-nil error
- **THEN** the system SHALL write `status=failed` and set `err_msg` to a summary of `error.Error()`
