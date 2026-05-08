## ADDED Requirements

### Requirement: The Host Captures Source-Plugin Route Ownership at Registration Time
The system SHALL capture source-plugin route ownership, path, method, and documentation metadata during route registration instead of inferring ownership from prefixes or manifest entries.

#### Scenario: Source plugin registers an arbitrary legal route
- **WHEN** a source plugin binds a backend route through the host-provided source-plugin route facade
- **THEN** the plugin may use any legal route path
- **AND** the host records the owning `pluginID`
- **AND** no fixed route prefix is required

#### Scenario: Standard DTO handler metadata is captured automatically
- **WHEN** a source plugin binds a standard GoFrame DTO handler `func(ctx context.Context, req *Req) (res *Res, err error)`
- **THEN** the host captures path, method, tags, summary, description, and permission metadata from `Req.g.Meta`
- **AND** the plugin does not need to duplicate that route metadata elsewhere

#### Scenario: Raw handlers are still allowed but are not auto-documented
- **WHEN** a source plugin binds a raw handler such as `func(*ghttp.Request)`
- **THEN** the host still records route ownership
- **AND** the route is not automatically projected into OpenAPI without DTO metadata

### Requirement: Source-Plugin Middleware Composition Remains Plugin-Owned
The system SHALL continue to let source plugins maintain their own middleware composition and ordering.

#### Scenario: Source plugin defines its own middleware order
- **WHEN** a source plugin registers middleware through the source-plugin route facade
- **THEN** the host binds those middleware handlers in the declared order
- **AND** the host does not convert them into a dynamic-plugin-style controlled middleware descriptor model
