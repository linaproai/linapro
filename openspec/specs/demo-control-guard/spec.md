# Demo Control Guard

## Purpose

Define how the official `demo-control` source plugin enables read-only demo protection while preserving required session access and blocking plugin governance writes when demo mode is enabled.

## Requirements

### Requirement: Demo read-only mode is controlled by the plugin's enabled state

The system MUST treat the installed-and-enabled state of `demo-control` as the runtime switch for demo protection. `plugin.autoEnable` only controls startup installation and enablement; it MUST NOT be treated as a separate runtime switch after startup.

#### Scenario: Demo protection stays disabled in the default configuration
- **WHEN** the host starts with default delivered configuration and `plugin.autoEnable` does not contain `demo-control`
- **THEN** the host does not install or enable `demo-control`
- **AND** deployments that never enable the plugin do not block writes by default

#### Scenario: Manual enablement activates demo protection
- **WHEN** an administrator installs and enables `demo-control`
- **THEN** demo-control middleware becomes active for later requests
- **AND** write requests are blocked by read-only demo rules

### Requirement: The host must deliver a demo-control source plugin with the source tree

The system MUST deliver an official source plugin named `demo-control` so deployments can enable the capability through startup config or plugin governance.

#### Scenario: The host discovers the demo-control source plugin
- **WHEN** the host scans source plugins and synchronizes registry data
- **THEN** it discovers `demo-control`
- **AND** operators can decide whether to enable it

### Requirement: The demo-control plugin must block system write operations when enabled

When enabled, demo-control MUST block write requests across the system by HTTP method semantics while allowing read-style requests.

#### Scenario: No write interception while disabled
- **WHEN** `demo-control` is not enabled
- **THEN** `POST`, `PUT`, and `DELETE` requests are not rejected by demo-control

#### Scenario: Query-style requests remain allowed
- **WHEN** `demo-control` is enabled
- **AND** a request uses `GET`, `HEAD`, or `OPTIONS`
- **THEN** demo-control allows the request to continue

#### Scenario: Write requests are rejected
- **WHEN** `demo-control` is enabled
- **AND** a request uses `POST`, `PUT`, or `DELETE`
- **THEN** demo-control rejects the request with a clear read-only demo message
- **AND** the request does not continue into business processing

### Requirement: The demo-control plugin must preserve a minimal session whitelist

The system MUST preserve login and logout behavior while demo-control is enabled so the demo environment remains usable.

#### Scenario: Login stays allowed
- **WHEN** `demo-control` is enabled
- **AND** the request is `POST /api/v1/auth/login`
- **THEN** demo-control allows the request to continue

#### Scenario: Logout stays allowed
- **WHEN** `demo-control` is enabled
- **AND** the request is `POST /api/v1/auth/logout`
- **THEN** demo-control allows the request to continue

### Requirement: The demo-control plugin must reject plugin-governance write operations when enabled

When `demo-control` is enabled, the system SHALL reject plugin governance writes, including plugin synchronization, dynamic package upload, installation, uninstallation, enablement, and disablement. Plugin management `GET`, `HEAD`, and `OPTIONS` requests remain allowed as read-only operations.

#### Scenario: Plugin installations are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled
- **AND** the request is `POST /api/v1/plugins/{id}/install`
- **THEN** demo-control rejects the request with a read-only demo message

#### Scenario: Plugin enable and disable requests are rejected
- **WHEN** `demo-control` is enabled
- **AND** the request is `PUT /api/v1/plugins/{id}/enable` or `PUT /api/v1/plugins/{id}/disable`
- **THEN** demo-control rejects the request with a read-only demo message

#### Scenario: Plugin uninstalls are rejected
- **WHEN** `demo-control` is enabled
- **AND** the request is `DELETE /api/v1/plugins/{id}`
- **THEN** demo-control rejects the request with a read-only demo message

#### Scenario: Plugin sync and upload writes are rejected
- **WHEN** `demo-control` is enabled
- **AND** the request is `POST /api/v1/plugins/sync` or `POST /api/v1/plugins/dynamic/package`
- **THEN** demo-control rejects the request with a read-only demo message

#### Scenario: Plugin management reads stay allowed
- **WHEN** `demo-control` is enabled
- **AND** the request is a plugin management query using `GET`, `HEAD`, or `OPTIONS`
- **THEN** demo-control allows the request to continue

