## ADDED Requirements

### Requirement: Handler Registry

The system SHALL provide a unified scheduled job handler registry that manages available handlers on both host and plugin sides, serving as the single source of truth for task scheduling, UI dropdowns, and parameter validation.

#### Scenario: Register handler

- **WHEN** host or plugin registers a handler through `Register(ref, def)`
- **THEN** the system SHALL persist the handler definition in the in-memory registry
- **AND** `ref` SHALL be in the form `host:<name>` or `plugin:<pluginID>/<name>`, globally unique
- **AND** `def` SHALL include at least `DisplayName / Description / ParamsSchema / Source / Invoke`

#### Scenario: Duplicate registration conflict

- **WHEN** attempting to register an already-existing `ref`
- **THEN** the system SHALL return a clear error
- **AND** SHALL NOT overwrite the registered handler

#### Scenario: Unregister handler

- **WHEN** host or plugin calls `Unregister(ref)`
- **THEN** the system SHALL remove the handler from the in-memory registry
- **AND** trigger task state cascade

### Requirement: Handler Parameters JSON Schema

The system SHALL require handlers to declare a `JSON Schema draft-07` restricted scalar subset for parameters, used for parameter validation and UI dynamic rendering.

#### Scenario: Only accept restricted Schema subset

- **WHEN** host or plugin registers a handler's `ParamsSchema`
- **THEN** the root node SHALL be `type=object`
- **AND** field types SHALL only support `string / integer / number / boolean`
- **AND** keywords SHALL only support `properties / required / description / default / enum / format`
- **AND** the system SHALL reject `array`, nested `object`, `$ref`, `allOf`, `anyOf`, `oneOf`, `not`, `patternProperties` and other structures beyond this iteration's form mapping capability

#### Scenario: Validate parameters on task creation

- **WHEN** creating or modifying a task with `task_type=handler`
- **THEN** the system SHALL validate `params` against the handler's JSON Schema
- **AND** return specific field error messages on validation failure

#### Scenario: Validate parameters at execution time

- **WHEN** before handler execution
- **THEN** the system SHALL re-validate `params_snapshot` against the latest Schema
- **AND** on validation failure, log `status=failed` and abort the execution

#### Scenario: UI dropdown data

- **WHEN** frontend calls `GET /job/handler`
- **THEN** the system SHALL return all registered handler list
- **AND** the list SHALL include `ref / displayName / description / source / pluginId`

#### Scenario: UI detail fetch Schema

- **WHEN** frontend calls `GET /job/handler/{ref}`
- **THEN** the system SHALL return the handler's complete `ParamsSchema`
- **AND** the frontend can use it to dynamically render form controls

### Requirement: Plugin Handler Lifecycle

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

### Requirement: Dynamic Plugin Scheduled Job Declaration Contract

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

### Requirement: Handler Execution Contract

The system SHALL require all handlers to support context cancellation in their specification and exit promptly upon cancellation.

#### Scenario: Handler accepts context

- **WHEN** the system calls the handler's `Invoke(ctx, params)` method
- **THEN** the handler SHALL check `ctx.Done()` during long-blocking operations or pass `ctx` through to downstream
- **AND** upon receiving cancellation signal, promptly return `ctx.Err()`

#### Scenario: Handler returns structured result

- **WHEN** handler execution succeeds
- **THEN** the handler SHALL return a JSON-serializable result
- **AND** the system SHALL write the result to `sys_job_log.result_json`

#### Scenario: Handler throws error

- **WHEN** handler execution returns a non-nil error
- **THEN** the system SHALL record `status=failed` with `err_msg` as error.Error() summary
