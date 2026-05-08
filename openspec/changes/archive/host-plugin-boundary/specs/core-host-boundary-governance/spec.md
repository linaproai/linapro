## MODIFIED Requirements

### Requirement: The open source stage host only retains the framework core and management base.

The system SHALL converge `apps/lina-core` into the framework core and management base during the open source stage, and no longer have all management backend business modules built-in by default.

#### Scenario: Planning to add new backend module
- **WHEN** The team plans to add a new management backend module
- **THEN** First determine whether the capability belongs to the host base capabilities such as authentication, permissions, menus, plugin management, task scheduling, configuration, dictionary or files, etc.
- **AND** If it does not belong to the host base capability, priority will be given to the source plugin design rather than being directly incorporated into the host.

#### Scenario: Determine whether the ability should remain in the host
- **WHEN** A capability is reused by multiple modules and assumes unified governance responsibilities at the framework level.
- **THEN** The system keeps it on the host
- **AND** Do not continue to expand the host boundary to the business side due to the needs of an optional business module

### Requirement: The default backend first-level directory is stably provided by the host

The system SHALL provide the default backend first-level directory mount point by the host to ensure that developers do not need to repeatedly adjust the top-level navigation structure when expanding their business in the long term.

#### Scenario: The plugin provides a background function menu
- **WHEN** A source plugin needs to register the menu with the default backend
- **THEN** The menu of this plugin MUST be mounted to the stable first-level directory provided by the host
- **AND** Plugins MUST not bypass host management and create new first-level directory systems on their own

#### Scenario: Plugin not installed or enabled
- **WHEN** All submenus under a certain level of directory come from plugins that are not installed or enabled.
- **THEN** The host automatically hides the empty directory
- **AND** Do not keep empty shell parent directory in left navigation

### Requirement: The host stable directory MUST exist as a real governance record

The system SHALL maintain the nine first-level directories in the default backend as stable menu records owned by the host, instead of just temporarily assembling them in the frontend projection layer.

#### Scenario: Initialize the host stable directory
- **WHEN** The host initializes the default background menu skeleton
- **THEN** The host creates and maintains `dashboard`, `iam`, `org`, `setting`, `content`, `monitor`, `scheduler`, `extension`, `developer` 9 stable parents `menu_key`
- **AND** These directory records can be stably parsed by the plugin `parent_key`

#### Scenario: There is no visible submenu in a certain directory.
- **WHEN** The `Content Management`, `Organization Management` or `System Monitoring` directories currently do not have any visible submenus
- **THEN** They are hidden in the navigation projection
- **AND** The host does not delete the corresponding stable directory record

### Requirement: Authentication session kernel and unified event publishing capabilities remain on the host

The system SHALL retain the authentication session truth source and the publishing capabilities of unified login events and unified audit events on the host, rather than delegating them to optional source plugins.

#### Scenario: Planning online user plugin boundaries
- **WHEN** Team planning capability boundaries of `monitor-online`
- **THEN** plugin only carries online user query and forced offline management
- **AND** JWT verification, session touch refresh, timeout determination and cleanup tasks still remain on the host

#### Scenario: Planning log plugin boundaries
- **WHEN** Team planning capability boundaries of `monitor-loginlog` or `monitor-operlog`
- **THEN** The host publishes unified events on the authentication link and request link
- **AND** The host core link does not directly depend on the specific persistence implementation of these plugins.

### Requirement: Host and plugin MUST be decoupled through stable capability seams

The system SHALL complete the collaboration between the host and the plugin through stable joints such as capability interfaces, event Hooks, routing registers, and Cron registers, instead of scattering plugin-specific placeholder logic and a large number of `if pluginEnabled` branches in the host business code.

#### Scenario: Host invokes optional organizational capabilities
- **WHEN** User management, authentication or other host core modules require access to optional capabilities such as departments and positions.
- **THEN** The host accesses these capabilities through a unified organizational capability interface
- **AND** The plugin status judgment and function branches of `org-center` are not directly scattered in the host implementation.
- **AND** The host only holds the interface, DTO and empty implementation of this capability, and does not directly query or maintain the physical table of `org-center`

#### Scenario: The host expands plugin logging or monitoring capabilities
- **WHEN** Non-core capabilities are split into source plugins
- **THEN** The host only retains stable events, governance interfaces and registration entrances
- **AND** Do not keep a lot of functionality placeholder logic for individual plugins in hosting controllers, services or middleware

### Requirement: The host MUST not hold the source plugin's own business storage

The system SHALL treat the source plugin business table, corresponding ORM artifacts, and demo data as plugin private assets, and does not retain long-term copies during host default database initialization, Mock loading, or in the host source tree.

#### Scenario: Initialize the default database
- **WHEN** Administrator performs host default database initialization
- **THEN** The host only creates and initializes the host core tables and necessary Seed data
- **AND** Do not create any source plugin business table

#### Scenario: Migrate business modules to source plugins
- **WHEN** A certain business module has been migrated to an official source plugin
- **THEN** The `dao`, `do`, `entity` and direct table lookup logic corresponding to the business table of this module are no longer retained in the host source.
- **AND** The host only works with the plugin through the stable capability seam or plugin registration portal

#### Scenario: Load default demo data
- **WHEN** Administrator performs host default Mock data loading
- **THEN** The host does not write any source plugin business tables
- **AND** The plugin demo data is responsible for the plugin's own lifecycle resources
