## Requirements

### Requirement: Host Publishes Stable Backend Extension Points

The system SHALL publish named, versioned, auditable backend extension points for plugins to register callbacks.

#### Scenario: Host maintains formal backend extension point directory
- **WHEN** the host publishes plugin backend extension capabilities
- **THEN** it maintains a formal directory including at least `auth.login.succeeded`, `auth.logout.succeeded`, `system.started`, `plugin.installed`, `plugin.enabled`, `plugin.disabled`, `plugin.uninstalled`
- **AND** each extension point documents trigger timing, context, execution order, execution mode, and failure isolation

#### Scenario: Hook execution failure is isolated from main flow
- **WHEN** a plugin hook times out, panics, or returns an error
- **THEN** the host main flow still succeeds
- **AND** the host records the failure for diagnostics
- **AND** other plugins' hooks continue executing

#### Scenario: Active release hook contract persists across upgrades
- **WHEN** a dynamic plugin has an active release and undergoes staged upgrade or rollback
- **THEN** the host still dispatches events based on the active release's embedded hook contract

### Requirement: Host Publishes Backend Callback Registration Extension Points

The system SHALL provide callback registration interfaces for source plugins including `http.route.register`, `http.request.after-auth`, `cron.register`, `menu.filter`, `permission.filter`.

#### Scenario: Plugins register HTTP routes via callback
- **WHEN** a source plugin registers HTTP routes
- **THEN** the host auto-mounts routes at startup
- **AND** disabled plugin routes are rejected or degraded
- **AND** the developer only maintains the import in `apps/lina-plugins/lina-plugins.go`

#### Scenario: Host provides independent prefix-free route groups for plugins
- **WHEN** the host opens HTTP route registration to plugins
- **THEN** it provides route groups independent of `/api/v1`
- **AND** plugins choose their own prefix and middleware combination

#### Scenario: Plugins can split multiple route groups with different governance
- **WHEN** a plugin needs both public and protected endpoints
- **THEN** it can declare multiple independent route groups in one registration callback
- **AND** each group selects its own middleware subset and order

### Requirement: Host Publishes Frontend Slot Extension Points

The system SHALL publish controlled frontend slots for layout and page content injection.

#### Scenario: Host maintains formal frontend slot directory
- **WHEN** the host publishes frontend slot capabilities
- **THEN** it maintains a directory including `layout.user-dropdown.after`, `dashboard.workspace.after`, `layout.header.actions.before/after`, `auth.login.after`, `crud.toolbar.after`, `crud.table.after`

#### Scenario: Plugin content renders within host containers
- **WHEN** an enabled plugin declares content for a slot
- **THEN** the host renders that content in the slot's container
- **AND** disabling the plugin immediately hides the content

### Requirement: Hook and Slot Execution Order is Predictable

The system SHALL define stable execution order for multiple plugins on the same hook or slot.

#### Scenario: Multiple plugins subscribe to the same hook
- **WHEN** multiple plugins subscribe to the same hook or slot
- **THEN** the host executes by manifest priority or default sort order
- **AND** execution order is consistent across nodes

### Requirement: Hook and Slot Identifiers Use Typed Constants

The system SHALL use Go types and TypeScript constants for extension point identifiers, prohibiting hardcoded strings.

#### Scenario: Unknown slot declaration
- **WHEN** a plugin declares an unpublished hook or slot identifier
- **THEN** the host rejects or skips the declaration
