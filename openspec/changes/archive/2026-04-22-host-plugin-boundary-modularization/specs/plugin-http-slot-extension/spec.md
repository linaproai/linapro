## MODIFIED Requirements

### Requirement: The host publishes common backend extension points through callback registration.
The system SHALL provides a minimized, callback-registered backend extension interface for source plugins, preventing plugin authors from maintaining complex declarative properties for common extension scenarios.

#### Scenario: Event Hook and registered interface belong to the same type of backend extension point
- **WHEN** The host publishes event-type Hooks and registered callback extension points at the same time
- **THEN** Both types of capabilities MUST be considered "callback registrations on the host's published backend extension points"
- **AND** Plug-in developers uniformly select extension points and register callback functions through Go type constants
- **AND** The host only allows the execution mode to determine whether the callback blocks the main process or executes asynchronously

#### Scenario: The host maintains a formal backend callback extension point directory
- **WHEN** The host publishes source plugin backend expansion capabilities
- **THEN** The host MUST provide a unified Go registration entrance and callback registration method
- **AND** Provide at least `http.route.register`, `cron.register`, `menu.filter`, `permission.filter` in one phase
- **AND** These extension points MUST be maintained in the same technical documentation as published Hooks

#### Scenario: The host manages all backend extension points in a unified Go type directory
- **WHEN** The host publishes event-type Hooks and registered callback extension points at the same time
- **THEN** The host MUST maintain these backend extension points using a unified Go `type` and constant directory
- **AND** Plug-in code, host scheduling code and technical documentation MUST all reference the same set of type constants
- **AND** Do not allow hardcoded backend extension point strings littered with host implementations or plugin examples

#### Scenario: The host declares the execution mode for the backend callback extension point
- **WHEN** The plugin registers a certain backend extension point callback with the host
- **THEN** The registration interface MUST explicitly declare the execution mode of the callback
- **AND** execution modes differentiate at least between `blocking` and `async`
- **AND** The host MUST verify whether the current extension point supports the declared execution mode
- **AND** Unsupported execution modes MUST be rejected during the registration phase

#### Scenario: The host exposes the callback input object as an interface type
- **WHEN** The host exposes callback input objects such as Hook, HTTP registration, Cron, menu filtering or permission filtering to the plugin.
- **THEN** The host MUST give priority to exposing abstract interfaces rather than concrete structure pointers
- **AND** Plug-in callbacks only rely on the method contract exposed by the host
- **AND** When the host subsequently extends fields or capabilities, it should not require plugins to directly couple the internal structure implementation.

#### Scenario: The plugin declares routing and middleware through a unified HTTP registration portal
- **WHEN** The host opens HTTP registration capabilities to source plugins.
- **THEN** The host MUST expose a unified HTTP registration entry object to the plugin
- **AND** This object provides both a route registrar and a global HTTP middleware registrar
- **AND** Plugins do not hold bare `*ghttp.Server` directly

#### Scenario: Plug-in registers HTTP routes governed by the host through callbacks
- **WHEN** A source plugin registers its own HTTP route
- **THEN** The host automatically assembles this route at startup
- **AND** When the plugin is disabled, these routing requests will be rejected or downgraded by the host
- **AND** Plug-in authors do not need to manually modify the host controller or routing skeleton code
- **AND** Plugin authors only need to maintain explicit import relationships of plugin backend packages in `apps/lina-plugins/lina-plugins.go`

#### Scenario: The host provides independent unprefixed route groupings for plugins
- **WHEN** The host opens HTTP route registration capabilities to source plugins.
- **THEN** The host MUST provide a plugin routing group independent of the main service `/api/v1` group
- **AND** The plugin routing group itself MUST not have any built-in fixed routing prefixes
- **AND** The plugin can choose whether to mount to `/api/v1`, other business prefixes or no prefix paths

#### Scenario: The host exposes the optional main service routing middleware directory to the plugin.
- **WHEN** The host opens HTTP routing group registration capabilities to source plugins.
- **THEN** The host MUST expose the published main service routing middleware directory to the plugin
- **AND** One phase includes at least `NeverDoneCtx`, `HandlerResponse`, `CORS`, `RequestBodyLimit`, `Ctx`, `Auth`, `Permission`
- **AND** The plugin can select any subset according to its own route grouping needs and determine the combination order
- **AND** Plug-ins can also combine host routing middleware with plugin custom grouping middleware

#### Scenario: The host exposes a governed global HTTP middleware registrar to plugins
- **WHEN** Source code plugins need to implement auditing or other request-level cross-cutting logic around the host dynamic request chain
- **THEN** The host MUST expose a global HTTP middleware register to the plugin that is independent of the route grouping
- **AND** Registrar uses GoFrame primitive pattern as scope
- **AND** The host uses `ghttp.Server` global middleware to assemble these processors
- **AND** After the plugin is deactivated, the host bypasses the corresponding middleware logic through the runtime switch without requiring rebuilding the routing tree.

#### Scenario: The plugin can split multiple routing groups with different governance strategies
- **WHEN** The same source plugin needs to expose both the authentication-free interface and the protected interface.
- **THEN** The plugin MUST be able to declare multiple independent route groups in one route registration callback
- **AND** The routing group registration method MUST be consistent with the host main service, and supports the `group.Group(prefix, func(group *ghttp.RouterGroup) { ... })` style
- **AND** Each routing group can choose to host any subset and combination of published middleware in any order
- **AND** Each routing group can continue to append its own sub-path prefix

#### Scenario: The plugin registers scheduled tasks that are controlled by the host's start and stop.
- **WHEN** A source plugin registers its own scheduled tasks
- **THEN** The host completes registration through the unified `cron` component
- **AND** After the plugin is disabled, the host will not execute the scheduled task callback of the plugin.
- **AND** Plug-in authors do not need to manually add scheduled task code at the host `cmd` layer

#### Scenario: Scheduled task register exposes master node identification capability
- **WHEN** The host exposes the scheduled task registration input object to the plugin
- **THEN** This object MUST provide an identification method for "whether the current node is the main node"
- **AND** The plugin can decide whether to execute timing logic that only takes effect on the master node based on this method

#### Scenario: Plug-ins participate in host management links by menu and permission filter
- **WHEN** The host generates a menu list or permissions list
- **THEN** The host publishes `menu.filter` and `permission.filter` callbacks to enabled plugins
- **AND** The plugin can only filter and judge the menu/permission descriptions exposed by the host.
- **AND** Failure to filter will not destroy the host's original menu and permission calculation process
