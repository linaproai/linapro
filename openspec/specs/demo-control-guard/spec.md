# demo-control-guard Specification

## Purpose
Define how the official `demo-control` source plugin enables read-only demo protection while preserving the minimum governance and session allowlists required for administration.

## Requirements
### Requirement: Demo read-only mode is controlled by the plugin's enabled state

The system MUST treat the current installed-and-enabled state of `demo-control` as the runtime switch for demo protection. `plugin.autoEnable` only controls whether the host automatically installs and enables the plugin during startup; it must not act as a separate runtime switch once the host is running.

#### Scenario: Demo protection stays disabled in the default configuration
- **WHEN** the host starts with the default delivered configuration and `plugin.autoEnable` does not contain `demo-control`
- **THEN** the host does not automatically install or enable `demo-control`
- **AND** deployments that never enable the plugin do not block write requests by default

#### Scenario: Manual plugin enablement activates demo protection immediately
- **WHEN** `demo-control` is absent from `plugin.autoEnable`
- **AND** an administrator installs and enables the plugin through plugin governance
- **THEN** the demo-control middleware becomes active for subsequent requests immediately
- **AND** write requests start being blocked according to the demo read-only rules

#### Scenario: Auto-enable still handles startup activation
- **WHEN** the deployment config adds `demo-control` to `plugin.autoEnable`
- **THEN** the host automatically installs and enables that plugin during startup
- **AND** demo protection becomes active because the plugin is enabled, not because the config key bypasses plugin state

### Requirement: The host must deliver a demo-control source plugin with the source tree

The system MUST deliver an official source plugin named `demo-control` with the source tree so deployments can enable the capability through `plugin.autoEnable` or through plugin governance.

#### Scenario: The host discovers the demo-control source plugin
- **WHEN** the host scans the source-plugin directory and synchronizes the plugin registry
- **THEN** the host discovers the `demo-control` source plugin
- **AND** operators can decide through `plugin.autoEnable` or explicit governance actions whether it should be enabled

### Requirement: The demo-control plugin must block system write operations when enabled

The system MUST block write requests across `/*` based on request HTTP-method semantics when `demo-control` is enabled, while still preserving query-style requests.

#### Scenario: No write interception while demo-control is disabled
- **WHEN** `demo-control` is not enabled by the host
- **THEN** `POST`, `PUT`, and `DELETE` requests are not rejected by extra demo-control logic

#### Scenario: Query-style requests remain allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request enters the system request chain with method `GET`, `HEAD`, or `OPTIONS`
- **THEN** the demo-control plugin allows the request to continue

#### Scenario: Write requests are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request enters the system request chain with method `POST`, `PUT`, or `DELETE`
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message
- **AND** the request does not continue into later business processing

#### Scenario: Non-API-prefix write requests are still covered
- **WHEN** `demo-control` is enabled by the host
- **AND** the request path is not under `/api/v1`
- **AND** the request method is `POST`, `PUT`, or `DELETE`
- **THEN** the demo-control plugin rejects that request as well

#### Scenario: Rejection messages explain the demo-mode reason
- **WHEN** `demo-control` rejects any system write request
- **THEN** the returned error message explicitly states that the request was rejected because demo mode is enabled
- **AND** the frontend does not present the case as a generic no-permission error

### Requirement: The demo-control plugin must preserve a controlled plugin-governance whitelist

The system MUST preserve a minimal plugin-governance whitelist while `demo-control` is enabled: plugins other than `demo-control` itself may still be installed, uninstalled, enabled, and disabled.

#### Scenario: Other plugin installations stay allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `POST /api/v1/plugins/{id}/install`
- **AND** `{id}` is not `demo-control`
- **THEN** the demo-control plugin allows the request to continue into plugin installation processing

#### Scenario: Other plugin enable and disable requests stay allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `PUT /api/v1/plugins/{id}/enable` or `PUT /api/v1/plugins/{id}/disable`
- **AND** `{id}` is not `demo-control`
- **THEN** the demo-control plugin allows the request to continue into plugin state updates

#### Scenario: Other plugin uninstalls stay allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `DELETE /api/v1/plugins/{id}`
- **AND** `{id}` is not `demo-control`
- **THEN** the demo-control plugin allows the request to continue into plugin uninstall processing

#### Scenario: Demo-control cannot change its own governance state while enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** a request attempts to install, uninstall, enable, or disable `demo-control` itself
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message

### Requirement: The demo-control plugin must preserve a minimal session whitelist

The system MUST preserve login and logout behavior while `demo-control` is enabled so the demo environment does not lose baseline usability.

#### Scenario: Login stays allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `POST /api/v1/auth/login`
- **THEN** the demo-control plugin allows the request to continue into the authentication chain

#### Scenario: Logout stays allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `POST /api/v1/auth/logout`
- **THEN** the demo-control plugin allows the request to continue into later processing
