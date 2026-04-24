# demo-control-guard Specification

## Purpose
TBD - created by archiving change demo-control-plugin. Update Purpose after archive.
## Requirements
### Requirement: The host must use `plugin.autoEnable` to control the demo-mode switch

The system MUST treat whether `plugin.autoEnable` includes `demo-control` as the only switch that turns demo protection on or off.

#### Scenario: Demo protection stays disabled in the default config
- **WHEN** the host starts with the default delivered configuration
- **THEN** the default config does not force `demo-control` into `plugin.autoEnable`
- **AND** instances that do not enable `demo-control` do not block write requests by default

#### Scenario: Demo protection is turned on through the auto-enable list
- **WHEN** the deployment config adds `demo-control` to `plugin.autoEnable` in the host main config file
- **THEN** the host automatically installs and enables that plugin during startup
- **AND** the demo-control middleware becomes active across the `/*` request chain

### Requirement: The host must deliver a demo-control source plugin with the source tree

The system MUST deliver an official source plugin named `demo-control` with the source tree so deployments can enable the capability through `plugin.autoEnable`.

#### Scenario: The host discovers the demo-control source plugin
- **WHEN** the host scans the source-plugin directory and synchronizes the plugin registry
- **THEN** the host discovers the `demo-control` source plugin
- **AND** operators can decide through `plugin.autoEnable` whether it should be enabled during startup

### Requirement: The demo-control plugin must block system write operations when enabled

The system MUST block write requests across `/*` based on request `HTTP Method` semantics when `demo-control` is enabled, while still preserving query-style requests.

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

