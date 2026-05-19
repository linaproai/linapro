## ADDED Requirements

### Requirement: Source-plugin HTTP registration must receive host-published dependency directory

The system SHALL expose the host-published dependency directory to source plugins in HTTP, global middleware, and Cron registration callbacks. Plugins obtain stable host capability adapters through this directory and use explicit dependency injection to construct plugin Controllers and Services.

#### Scenario: Plugin route registration constructs controllers

- **WHEN** a source plugin binds controllers in the `http.route.register` callback
- **THEN** the plugin obtains the host-published dependency directory from the registrar
- **AND** the plugin controller constructor receives the constructed plugin service or its explicit dependencies
- **AND** the plugin controller MUST NOT create host `pluginservice` adapter instances on its own

#### Scenario: Plugin global middleware reuses host dependencies

- **WHEN** a source plugin registers global HTTP middleware
- **THEN** the middleware service obtains plugin state, i18n, configuration, or other host capability dependencies through the registrar
- **AND** the middleware MUST NOT create new host service adapters on the request path

#### Scenario: Plugin Cron registration reuses host dependencies

- **WHEN** a source plugin registers scheduled tasks
- **THEN** the scheduled task service obtains host-published dependencies through the Cron registrar or equivalent context
- **AND** the scheduled task execution path MUST NOT temporarily create key host service graphs
