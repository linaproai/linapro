## ADDED Requirements

### Requirement: Platform Tenant Control Plane Must Require Platform Context

System SHALL treat platform tenant control plane interfaces as platform resource access boundaries. Callers in addition to having corresponding permission strings MUST be in platform context, and current request must not be acting-as-tenant context. Users in tenant context even if holding `system:tenant:*` permissions due to historical dirty data MUST be denied access to platform tenant control plane data.

#### Scenario: Platform context reads tenant list

- **WHEN** Platform admin in platform context with `system:tenant:list` permission
- **THEN** System allows calling platform tenant list interface
- **AND** Returns platform tenant governance view allowed data

#### Scenario: Tenant context holding abnormal platform permissions still denied

- **WHEN** Tenant user in tenant context
- **AND** Effective permission snapshot contains `system:tenant:list` due to historical dirty data
- **THEN** Calling platform tenant list interface MUST return structured permission error
- **AND** Response must not contain any other tenant data

#### Scenario: Platform admin acting-as-tenant cannot operate platform tenant control plane

- **WHEN** Platform admin enters some tenant's acting-as-tenant context
- **AND** Calls platform tenant create, update, delete, enable/disable or list interface
- **THEN** System MUST deny this platform control plane operation by tenant context
- **AND** Must not bypass current context boundary because operator original identity is platform admin

### Requirement: Platform Control Plane Errors Must Be Localizable and Auditable

Platform control plane context not meeting requirements when system SHALL return stable `bizerr` business error with machine code, message key, English fallback and necessary parameters. Rejection events SHALL be locatable by operation log or security log audit to calling user, current tenant context and target platform resource type.

#### Scenario: Platform context missing error localizable

- **WHEN** Tenant context calls platform tenant control plane interface
- **THEN** Response contains stable `errorCode`
- **AND** Response contains `messageKey` for runtime translation
- **AND** Chinese/English error resources and apidoc i18n resources both cover this error
