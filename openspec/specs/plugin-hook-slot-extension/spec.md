# plugin-hook-slot-extension Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: The host publishes a stable backend extension point contract
The system SHALL publishes a set of named, versioned, auditable backend extension points for plugins to register callbacks and extend behavior on host critical business events or host governance links.

#### Scenario: The host maintains a formal catalog of backend extension points
- **WHEN** The host publishes plugin backend expansion capabilities
- **THEN** The host MUST maintain a formal backend extension point directory
- **AND** At least `auth.login.succeeded`, `auth.logout.succeeded`, `system.started`, `plugin.installed`, `plugin.enabled`, `plugin.disabled`, `plugin.uninstalled` are made public in one issue
- **AND** Each extension point MUST describe the triggering time, context, execution sequence, execution mode and failure isolation strategy

#### Scenario: Login success event triggers Hook
- **WHEN** The user logs in successfully and the host publishes `auth.login.succeeded` Hook
- **THEN** The host distributes this event to enabled plugins according to the agreed context
- **AND** The context contains at least user ID, login time, client information, request context and current plugin running generation information

#### Scenario: Logout success event triggers Hook
- **WHEN** The user logs out successfully and the host publishes the `auth.logout.succeeded` Hook
- **THEN** The host dispatches events to enabled plugins that subscribe to this Hook
- **AND** Plugins can only read context fields exposed by the host

### Requirement: Hook execution failure MUST be isolated from the main process
The system SHALL implements isolation on timeouts, exceptions and return errors of plugin Hooks, and plugin extensions MUST not be allowed to destroy the host's main link.

#### Scenario: Plug-in Hook execution failed
- **WHEN** A plugin times out, crashes, or returns an error in the login success Hook
- **THEN** The main user login process still returns success
- **AND** The host records the execution failure information of the plugin
- **AND** Hooks of other plugins still continue to execute in order or are safely skipped according to the policy

#### Scenario: The currently active dynamic release Hook contract is continuously available
- **WHEN** A dynamic plugin has been switched to an active release, and then a staged upgrade, failed rollback, or active release reload occurs.
- **THEN** The host still distributes events according to the Hook contract declared inline in the current active release.
- **AND** The host will not lose the release's Hook statement because only the top-level manifest metadata is restored.

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
- **WHEN** The host exposes callback input objects such as Hook, Route, Cron, menu filtering or permission filtering to the plugin.
- **THEN** The host MUST give priority to exposing abstract interfaces rather than concrete structure pointers
- **AND** Plug-in callbacks only rely on the method contract exposed by the host
- **AND** When the host subsequently extends fields or capabilities, it should not require plugins to directly couple the internal structure implementation.

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

#### Scenario: The host exposes the optional main service middleware directory to the plugin
- **WHEN** The host opens HTTP route registration capabilities to source plugins.
- **THEN** The host MUST expose the published main service middleware directory to the plugin
- **AND** One phase includes at least `NeverDoneCtx`, `HandlerResponse`, `CORS`, `Ctx`, `Auth`, `OperLog`
- **AND** The plugin can select any subset according to its own route grouping needs and determine the combination order
- **AND** Plug-ins can also use host middleware in combination with plugin custom middleware

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

### Requirement: Host publishes frontend Slot extension point
The system SHALL publishes controlled Slot extension points for frontend pages and layouts, allowing plugins to insert UI content in locations exposed by the host.

#### Scenario: The host maintains a formal frontend slot directory
- **WHEN** The host can publish frontend slots to the outside world.
- **THEN** The host MUST maintain a formal frontend slot directory
- **AND** Expose at least `layout.user-dropdown.after` and `dashboard.workspace.after` in one phase
- **AND** Each slot MUST specify the hosting location, rendering container, recommended usage, ordering rules and failure degradation strategy

### Requirement: The host gives priority to publishing the universal frontend slot in the public interface layer.
The system SHALL gives priority to publishing universal frontend slots in the layout interface, login interface, workbench interface and CRUD general interface to avoid binding extension points to a single business module.

#### Scenario: The host releases the first batch of general public slots
- **WHEN** Host expands frontend slot capabilities
- **THEN** At least `layout.header.actions.before`, `layout.header.actions.after`, `auth.login.after`, `dashboard.workspace.before`, `crud.toolbar.after`, `crud.table.after` will be added in the first phase
- **AND** These Slots MUST be attached to the existing public interface layer rather than a private page of a business module
- **AND** The development document MUST simultaneously describe the host location and recommended content of each Slot.

#### Scenario: The plugin inserts content in the public area of the login page
- **WHEN** An enabled plugin declares inserting frontend content into `auth.login.after`
- **THEN** The host renders this content after the login page form
- **AND** The content will be hidden immediately after the plugin is disabled

#### Scenario: Plug-in inserts content into the CRUD universal interface
- **WHEN** An enabled plugin declares inserting frontend content into `crud.toolbar.after` or `crud.table.after`
- **THEN** The host renders the content below the general Grid toolbar area or table area
- **AND** All pages that reuse this common interface can automatically obtain extension bits

#### Scenario: Plug-in inserts content into the host layout
- **WHEN** An enabled plugin declares inserting frontend content into `layout.user-dropdown.after`
- **THEN** The host attempts to load the frontend entry declared by the plugin at the corresponding location of the Slot.
- **AND** The Slot content of the source plugin MUST come from the real frontend source file, rather than relying solely on declarative JSON configuration
- **AND** These source files are placed in the `frontend/slots/` directory by default and are discovered and mounted by the host during build
- **AND** Plug-in content is only rendered within the scope of the container exposed by the host
- **AND** Plug-ins cannot have unauthorized access to undisclosed host internal implementations

#### Scenario: The plugin inserts the page entry into the menu bar in the upper right corner
- **WHEN** An enabled plugin declaration inserts the plugin menu entry into `layout.user-dropdown.after`
- **THEN** The host displays the entry copy in the menu bar in the upper right corner
- **AND** After clicking this entry, the host will open the plugin Tab page using internal page navigation.
- **AND** This process does not trigger a full page refresh

#### Scenario: Activate the entrance route in the upper right corner immediately after enabling the plugin online in the login state.
- **WHEN** The administrator enables a plugin that injects entries into `layout.user-dropdown.after` in the currently logged in session
- **THEN** The host can synchronously refresh the dynamic routing corresponding to the entry without logging in again.
- **AND** Users will not enter the 404 page after clicking this entry
- **AND** The host directly opens the plugin page in Tab mode.

#### Scenario: The current session regains focus after the plugin state changes
- **WHEN** Operations other than the currently logged in session change the plugin state which will inject `layout.user-dropdown.after` and the current tab regains focus
- **THEN** The host automatically synchronizes the visibility of the Slot and the corresponding dynamic routing
- **AND** The upper right corner entrance for enabled plugins reappears and can be opened normally
- **AND** The upper right corner entrance of disabled plugins is hidden in time

#### Scenario: Plugin Slot contract mismatch
- **WHEN** The frontend entry declared by the plugin is incompatible with the contract required by Slot
- **THEN** The host skips the rendering of the plugin content
- **AND** Host record contract mismatch error
- **AND** Other host contents of the current page are rendered normally.

### Requirement: Hook and Slot execution order is predictable
The system SHALL defines a stable execution order for multiple plugins on the same Hook or Slot.

#### Scenario: Multiple plugins subscribe to the same Hook
- **WHEN** Multiple plugins subscribe to the same backend Hook or frontend Slot at the same time
- **THEN** The host executes according to manifest explicit priority or unified default ordering rules
- **AND** The execution order under the same input remains consistent on each node

### Requirement: Hook and Slot identifiers MUST use special type definitions
The system SHALL uses specialized types to define legal plugin installation locations to avoid littering the host implementation and plugin examples with hard-coded strings.

#### Scenario: Backend extension points declared in Go
- **WHEN** The host implements the backend Hook slot or registered callback extension point
- **THEN** The host MUST declare legal backend extension point identifiers using Go `type` and constants
- **AND** The plugin backend example refers to the extension point through this type constant instead of writing the event name string directly.
- **AND** The host's internal service layer shall not maintain an additional set of semantically repeated alias constants.

#### Scenario: Front-end Slot slot declared in TypeScript
- **WHEN** The host implements the frontend Slot slot
- **THEN** The host MUST use TypeScript constants and type declarations to declare legal slot identifiers
- **AND** The host page, Slot loader and plugin frontend examples reference slots through a unified type instead of directly writing the Slot name string.

#### Scenario: Plugin declares unknown slot
- **WHEN** The plugin declares a Hook or Slot identifier that is not published by the host
- **THEN** The host rejects the declaration or skips loading
- **AND** The host logs the error message "The slot is not released or the contract is not supported"

### Requirement: The host provides slot technical documentation for plugin developers.
The system SHALL precipitates the frontend and backend slot directories, type definitions and example usage into technical documents that plugin developers can directly consult.

#### Scenario: Publish plugin development guide
- **WHEN** The host adds, adjusts or officially releases Hook/Slot slots
- **THEN** Host synchronization update `apps/lina-plugins/README.md`
- **AND** Clear distinction between "released slots" and "subsequent planned slots" in the document
- **AND** The recommended citation method for Go and frontend source plugins is given in the document

### Requirement: Dynamic plugin routing management metadata is concentrated in `g.Meta`

The system SHALL requires the dynamic plugin to centrally define the management metadata of backend dynamic routing in `g.Meta` of the `api` layer request structure to avoid introducing a second set of scattered routing management configuration sources.

#### Scenario: Dynamic plugin declares minimum governance fields

- **WHEN** Developer defines a dynamic plugin backend interface
- **THEN** This interface can declare `access`, `permission`, `operLog` in `g.Meta`
- **AND** `access` only supports `public` and `login`
- **AND** If `access` is not declared, it will be processed as `login`

#### Scenario: Public routing governance boundaries are limited

- **WHEN** Developer declares a `public` dynamic route
- **THEN** This route MUST not declare `permission`
- **AND** This route MUST not rely on host login state injection
- **AND** Illegal configurations will be rejected during the host loading phase

### Requirement: Dynamic plugin permission declaration reuses the host’s existing permission system

The system SHALL automatically connects the `permission` statement in dynamic routing to the host's existing `sys_menu.perms` permission system, instead of introducing an independent dynamic permission storage model.

#### Scenario: Dynamic routing permissions are materialized as hidden menu items

- **WHEN** A dynamic route declares a valid `permission`
- **THEN** The host automatically generates corresponding hidden permission menu items during the plugin menu synchronization phase.
- **AND** These permission menu items are mounted in the plugin's exclusive dynamic routing permission directory.
- **AND** The permission value directly reuses the `permission` declared by the dynamic routing

#### Scenario: Dynamic routing permissions are synchronized with the plugin life cycle

- **WHEN** Dynamic plugins are enabled, disabled, uninstalled or switched to active versions
- **THEN** The host synchronously adds, updates or removes the corresponding hidden permission menu items
- **AND** The default administrator role continues to have these permissions automatically

### Requirement: Dynamic plugins do not directly combine host management middleware

The system SHALL maintains dynamic plugins as a restricted business extension model and does not open the free assembly capabilities of host `Auth`, `Ctx`, `OperLog` and other management middleware to dynamic plugins.

#### Scenario: Dynamic governance is uniformly executed by the host

- **WHEN** A dynamic plugin route is hit by an external request
- **THEN** Login verification, permission verification and business context injection are uniformly executed by the host based on the routing contract
- **AND** The dynamic plugin only declares governance requirements and does not directly call the host governance middleware.

### Requirement: Dynamic plugins reuse public bridge components to reduce writing complexity

The system SHALL abstracts the dynamic plugin bridge envelope, binary codec, guest-side processor adaptation, error response assistance, and typed guest controller adaptation into `apps/lina-core/pkg` public components, preventing plugin authors from repeatedly writing the underlying ABI, codec templates, and manual conversion logic from envelope to API DTO in each dynamic plugin.

#### Scenario: Dynamic plugin controller directly reuses API request and response DTO

- **WHEN** Developer implements guest controller for a dynamic plugin route that has declared DTO in `backend/api/.../v1`
- **THEN** The controller can declare methods using the form `func(ctx context.Context, req *v1.XxxReq) (res *v1.XxxRes, err error)`
- **AND** guest route dispatcher matches runtime `RequestType` based on request DTO type name
- **AND** The dynamic routing contract built by the host continues to reuse the same API DTO metadata

#### Scenario: typed guest controller can still access the bridge context and write custom responses

- **WHEN** typed guest controller needs to read `pluginId`, `requestId`, identity snapshot, routing metadata, or return download stream / additional response header / custom status code
- **THEN** `pkg/pluginbridge` MUST provide helper methods for reading the bridge envelope from `context.Context`
- **AND** The component MUST provide helper methods for writing response headers, raw response bodies, or custom status codes
- **AND** Plugin authors do not need to fall back to directly declaring `func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)` to complete these scenarios

