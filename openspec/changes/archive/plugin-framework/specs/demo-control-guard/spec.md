## Requirements

### Requirement: Demo Read-Only Mode Controlled by linapro-ops-demo-guard Enabled State

The system SHALL treat `linapro-ops-demo-guard`'s installed and enabled state as the runtime switch for demo protection. `plugin.autoEnable` only controls startup auto-enable; after startup it must not be treated as a separate runtime switch.

#### Scenario: Default config does not enable demo protection
- **WHEN** the host starts with default delivery config and `plugin.autoEnable` does not contain `linapro-ops-demo-guard`
- **THEN** the host does not install or enable `linapro-ops-demo-guard`
- **AND** deployments that never enabled the plugin do not block write operations by default

#### Scenario: Manual enablement activates demo protection
- **WHEN** an administrator installs and enables `linapro-ops-demo-guard`
- **THEN** the demo guard middleware takes effect on subsequent requests
- **AND** write requests are blocked by read-only demo rules

### Requirement: Host Must Ship linapro-ops-demo-guard Source Plugin with Source Tree

The system SHALL ship an official source plugin named `linapro-ops-demo-guard` so that deployments can enable this capability through startup configuration or plugin governance. The host must not continue exposing the old ID `demo-control` as the official demo read-only protection plugin ID.

#### Scenario: Host discovers linapro-ops-demo-guard source plugin
- **WHEN** the host scans source plugins and synchronizes registry data
- **THEN** it discovers `linapro-ops-demo-guard`
- **AND** operations staff can decide whether to enable it
- **AND** the plugin list must not show a `demo-control` official plugin entry

### Requirement: linapro-ops-demo-guard Plugin Must Block System Write Operations When Enabled

When enabled, `linapro-ops-demo-guard` SHALL block system write requests by HTTP method semantics while allowing read requests.

#### Scenario: No write interception when disabled
- **WHEN** `linapro-ops-demo-guard` is not enabled
- **THEN** `POST`, `PUT`, and `DELETE` requests are not rejected by the demo guard

#### Scenario: Query requests remain allowed
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request uses `GET`, `HEAD`, or `OPTIONS`
- **THEN** the demo guard allows the request to continue

#### Scenario: Write requests are rejected
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request uses `POST`, `PUT`, or `DELETE`
- **THEN** the demo guard rejects the request with a clear read-only demo message
- **AND** the request does not continue into business processing

### Requirement: linapro-ops-demo-guard Plugin Must Preserve Minimal Session Whitelist

The system SHALL preserve login, token refresh, tenant selection, tenant switching, and logout behavior when `linapro-ops-demo-guard` is enabled, keeping the demo environment usable.

#### Scenario: Login remains allowed
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `POST /api/v1/auth/login`
- **THEN** the demo guard allows the request to continue

#### Scenario: Logout remains allowed
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `POST /api/v1/auth/logout`
- **THEN** the demo guard allows the request to continue

#### Scenario: Multi-tenant session switch remains allowed
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `POST /api/v1/auth/select-tenant` or `POST /api/v1/auth/switch-tenant`
- **THEN** the demo guard allows the request to continue

### Requirement: linapro-ops-demo-guard Plugin Must Reject Plugin Governance Write Operations When Enabled

When `linapro-ops-demo-guard` is enabled, the system SHALL reject plugin governance write operations including plugin sync, dynamic package upload, install, uninstall, enable, and disable. Plugin management `GET`, `HEAD`, and `OPTIONS` requests remain allowed as read-only operations.

#### Scenario: Reject plugin install when linapro-ops-demo-guard is enabled
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `POST /api/v1/plugins/{id}/install`
- **THEN** the demo guard rejects the request with a read-only demo message

#### Scenario: Reject plugin enable and disable requests
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `PUT /api/v1/plugins/{id}/enable` or `PUT /api/v1/plugins/{id}/disable`
- **THEN** the demo guard rejects the request with a read-only demo message

#### Scenario: Reject plugin uninstall
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `DELETE /api/v1/plugins/{id}`
- **THEN** the demo guard rejects the request with a read-only demo message

#### Scenario: Reject plugin sync and upload write operations
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is `POST /api/v1/plugins/sync` or `POST /api/v1/plugins/dynamic/package`
- **THEN** the demo guard rejects the request with a read-only demo message

#### Scenario: Plugin management reads remain allowed
- **WHEN** `linapro-ops-demo-guard` is enabled
- **AND** the request is a plugin management query using `GET`, `HEAD`, or `OPTIONS`
- **THEN** the demo guard allows the request to continue

## REMOVED Requirements

### Requirement: Demo Read-Only Mode Controlled by demo-control Enabled State

**Reason**: This requirement binds to the old official plugin ID `demo-control`. This change renames the plugin to `linapro-ops-demo-guard` in a breaking manner, with no old ID compatibility.

**Migration**: Use the new requirement "Demo Read-Only Mode Controlled by linapro-ops-demo-guard Enabled State".

### Requirement: Host Must Ship demo-control Source Plugin with Source Tree

**Reason**: The official demo read-only protection plugin ID changed from `demo-control` to `linapro-ops-demo-guard`.

**Migration**: Use `apps/lina-plugins/linapro-ops-demo-guard` and manifest ID `linapro-ops-demo-guard`.

### Requirement: demo-control Plugin Must Block System Write Operations When Enabled

**Reason**: Guard behavior is preserved, but the old plugin ID `demo-control` is no longer a valid runtime identity.

**Migration**: Use the new requirement "linapro-ops-demo-guard Plugin Must Block System Write Operations When Enabled".

### Requirement: demo-control Plugin Must Preserve Minimal Session Whitelist

**Reason**: Session whitelist behavior is preserved, but the old plugin ID `demo-control` is no longer a valid runtime identity.

**Migration**: Use the new requirement "linapro-ops-demo-guard Plugin Must Preserve Minimal Session Whitelist".

### Requirement: demo-control Plugin Must Reject Plugin Governance Write Operations When Enabled

**Reason**: Plugin governance write operation rejection behavior is preserved, but the old plugin ID `demo-control` is no longer a valid runtime identity.

**Migration**: Use the new requirement "linapro-ops-demo-guard Plugin Must Reject Plugin Governance Write Operations When Enabled".
