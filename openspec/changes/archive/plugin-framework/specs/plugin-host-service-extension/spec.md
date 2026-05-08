## Requirements

### Requirement: Dynamic Plugins Access All Capabilities via Versioned Host Service Protocol

The system SHALL provide a versioned host-service invoke protocol on top of `lina_env.host_call`, requiring all capabilities through the unified channel.

#### Scenario: Guest calls structured host service
- **WHEN** the guest SDK invokes a host service
- **THEN** the host parses `service`, `method`, resource identifier, and payload
- **AND** dispatches to the registered service handler
- **AND** returns a unified response envelope

#### Scenario: Unknown service or method is rejected
- **WHEN** the plugin calls an unsupported `service` or `method`
- **THEN** the host returns an explicit unsupported error

### Requirement: Host Service Access Governed by Capability Derivation and Resource Authorization

The system SHALL enforce both coarse-grained capability checks and fine-grained resource authorization, with capabilities auto-derived from `hostServices`.

#### Scenario: Plugin declares host service policy
- **WHEN** the developer declares `hostServices` in the manifest
- **THEN** the builder validates service, method, resource declarations, and policy parameters
- **AND** the host auto-derives capability classification from methods
- **AND** writes the normalized policy to the artifact

#### Scenario: Unauthorized host service call is rejected
- **WHEN** the plugin calls an undeclared service, method, or unauthorized resource
- **THEN** the host returns an explicit rejection error

### Requirement: Resource Host Service Declarations Are Permission Requests

All resource-type `hostServices` declarations represent permission requests, not automatic runtime grants. `storage` uses `resources.paths`, `network` uses URL patterns, `data` uses `resources.tables`, and low-priority services use logical `resourceRef`.

#### Scenario: Manifest declares resource requests
- **WHEN** the developer declares resource-type host services
- **THEN** the declarations represent permission applications
- **AND** are not automatically authorized at runtime

#### Scenario: Runtime only recognizes confirmed authorization snapshots
- **WHEN** the plugin calls a resource-type host service at runtime
- **THEN** the host only validates against the install/enable-time confirmed snapshot

### Requirement: Host Service Calls Carry Execution Context and Support Audit

Every host service call SHALL carry a unified execution context and record minimal audit information.

#### Scenario: Request-bound context
- **WHEN** a plugin calls a host service during route processing
- **THEN** the host passes plugin ID, execution source, route ID, user identity, and data scope

#### Scenario: System-bound context
- **WHEN** a plugin calls a host service from a hook, cron, or lifecycle flow
- **THEN** the host passes a user-less system context
- **AND** methods requiring user context reject the call

#### Scenario: Audit summary recorded
- **WHEN** a host service call completes
- **THEN** the host records plugin ID, service, method, resource summary, result status, and duration
