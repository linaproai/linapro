## REMOVED Requirements

### Requirement: The demo-control plugin must preserve a controlled plugin-governance whitelist

**REMOVED:** `demo-control` 启用后不再为插件管理保留写操作白名单。插件安装、卸载、启用和禁用都会改变系统状态，必须按演示只读模式写操作统一拦截。

## ADDED Requirements

### Requirement: The demo-control plugin must reject plugin-governance write operations when enabled

`demo-control` 启用后，系统 SHALL 拦截插件治理写操作，包括但不限于插件同步、动态插件包上传、安装、卸载、启用和禁用任意插件。插件管理的 `GET`、`HEAD` 和 `OPTIONS` 查询请求仍按只读请求放行。演示模式只保留登录和登出的最小会话白名单。

#### Scenario: Plugin installations are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `POST /api/v1/plugins/{id}/install`
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message

#### Scenario: Plugin enable and disable requests are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `PUT /api/v1/plugins/{id}/enable` or `PUT /api/v1/plugins/{id}/disable`
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message

#### Scenario: Plugin uninstalls are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `DELETE /api/v1/plugins/{id}`
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message

#### Scenario: Plugin sync and upload writes are rejected while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is `POST /api/v1/plugins/sync` or `POST /api/v1/plugins/dynamic/package`
- **THEN** the demo-control plugin rejects the request and returns a clear read-only demo message

#### Scenario: Plugin management reads stay allowed while demo-control is enabled
- **WHEN** `demo-control` is enabled by the host
- **AND** the request is a plugin management query using `GET`, `HEAD`, or `OPTIONS`
- **THEN** the demo-control plugin allows the request to continue as a read-only operation
