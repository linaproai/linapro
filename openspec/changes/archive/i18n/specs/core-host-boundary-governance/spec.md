## ADDED Requirements

### Requirement: lina-core must maintain generic core host boundaries
The system SHALL treat `apps/lina-core` as the framework's core host service, prioritizing stability and reusability of general module interface capabilities, component capabilities, system governance capabilities, and plugin extension capabilities.

#### Scenario: Requirements only affect workspace display
- **WHEN** a requirement only changes table columns, filters, tree selectors, workspace aggregation, route assembly, or other page-specific display structures
- **THEN** the system prioritizes completing the change through workspace adaptation interfaces or frontend adaptation layers
- **AND** the requirement does not directly modify `lina-core`'s core domain contracts, general service semantics, or storage models

#### Scenario: Planning to modify core interfaces or models
- **WHEN** a developer plans to modify `lina-core`'s core interfaces, domain models, or persistence structures due to a frontend page requirement
- **THEN** the modification must be able to demonstrate it serves framework-level general capabilities rather than a single page form
- **AND** if it cannot be demonstrated, it should fall back to a workspace adaptation implementation

### Requirement: Workspace adaptation interfaces must be explicitly classified
The system SHALL explicitly classify outputs such as menu route projection, current-user workspace startup data, tree selectors, and dropdown options as workspace adaptation interfaces rather than general domain interfaces.

#### Scenario: Interface returns workspace assembly data
- **WHEN** an interface returns menu routes, host workspace startup data, tree selector nodes, or dropdown options
- **THEN** its interface description, DTO comments, and related specifications clearly mark this output as workspace adaptation semantics
- **AND** the output is not described as the general domain model itself

#### Scenario: General domain capabilities reused by multiple workspaces
- **WHEN** a capability needs to be consumed by different workspaces or different access methods simultaneously
- **THEN** the system prioritizes preserving stable general domain interfaces
- **AND** different workspaces' required menus, routes, display structures, or aggregation views are assembled through independent adaptation outputs
