## MODIFIED Requirements

### Requirement: Demo read-only mode is controlled by the plugin's enabled state

The system SHALL treat the current installed-and-enabled state of the `demo-control` plugin as the runtime switch for demo read-only mode. `plugin.autoEnable` only controls whether the host automatically installs and enables the plugin during startup; it must not be the sole activation condition for demo protection.

#### Scenario: Default configuration does not auto-enable demo-control
- **WHEN** the host starts with the default delivered configuration and `plugin.autoEnable` does not contain `demo-control`
- **THEN** the host does not automatically install or enable `demo-control` during startup
- **AND** if administrators never enable the plugin manually, the system does not apply extra write interception

#### Scenario: Manual enablement activates demo read-only mode immediately
- **WHEN** `demo-control` is absent from `plugin.autoEnable`
- **AND** an administrator installs and enables the plugin through plugin governance
- **THEN** demo read-only mode takes effect for subsequent requests immediately
- **AND** `POST`, `PUT`, and `DELETE` requests are intercepted by `demo-control`

#### Scenario: `plugin.autoEnable` only controls startup auto-enablement
- **WHEN** the host configuration includes `demo-control` in `plugin.autoEnable`
- **THEN** the host automatically installs and enables the plugin during startup
- **AND** demo read-only mode becomes active because the plugin is enabled, not because the configuration key itself acts as an independent switch

### Requirement: Demo-control must block system write operations when enabled

The system MUST block system write requests across `/*` based on HTTP-method semantics whenever `demo-control` is enabled, while preserving query-style requests and the existing allowlist.

#### Scenario: Manual enablement blocks configuration writes
- **WHEN** an administrator has enabled `demo-control`
- **AND** a client sends `POST /api/v1/config`, `PUT /api/v1/config/{id}`, or `DELETE /api/v1/config/{id}`
- **THEN** `demo-control` rejects the request
- **AND** the response returns a clear demo read-only message

#### Scenario: Rejection messages explain the demo-mode reason
- **WHEN** `demo-control` blocks any system write request
- **THEN** the returned error message explicitly states that the request was rejected because demo mode is enabled
- **AND** the frontend does not present the case as a generic "no permission" error
