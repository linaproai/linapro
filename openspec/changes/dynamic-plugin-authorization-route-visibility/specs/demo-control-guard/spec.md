## MODIFIED Requirements

### Requirement: demo-control 的演示只读模式由插件启用状态控制

系统 SHALL 以 `demo-control` 插件当前是否处于已安装且已启用状态，作为演示只读模式是否生效的运行时开关；`plugin.autoEnable` 仅用于控制宿主启动时是否自动安装并启用该插件，不得再作为演示只读模式的唯一激活条件。

#### Scenario: 默认配置不自动开启 demo-control
- **WHEN** 宿主使用默认交付配置启动，且 `plugin.autoEnable` 未包含 `demo-control`
- **THEN** 宿主不会在启动阶段自动安装并启用 `demo-control`
- **AND** 若管理员随后未手动启用该插件，则系统不会额外拦截写请求

#### Scenario: 通过插件治理手动启用后立即生效
- **WHEN** `demo-control` 未出现在 `plugin.autoEnable` 中
- **AND** 管理员已通过插件治理流程将 `demo-control` 安装并启用
- **THEN** 演示只读模式立即对后续请求生效
- **AND** `POST`、`PUT`、`DELETE` 等写请求会被 `demo-control` 拦截

#### Scenario: plugin.autoEnable 仅承担启动期自动启用职责
- **WHEN** 宿主主配置将 `demo-control` 加入 `plugin.autoEnable`
- **THEN** 宿主在启动时自动安装并启用该插件
- **AND** 演示只读模式之所以生效，是因为该插件已处于启用状态，而不是因为命中了配置项本身

### Requirement: demo-control 在启用时必须拦截系统写操作

系统 MUST 在 `demo-control` 插件已启用时，基于请求 `HTTP Method` 语义拦截 `/*` 范围内的系统写请求，同时继续保留查询类请求和既有白名单路径。

#### Scenario: 手动启用后拦截参数设置写请求
- **WHEN** 管理员手动启用了 `demo-control`
- **AND** 前端或 API 客户端发起 `POST /api/v1/config`、`PUT /api/v1/config/{id}` 或 `DELETE /api/v1/config/{id}` 请求
- **THEN** `demo-control` 拒绝该请求
- **AND** 响应返回明确的演示只读提示

#### Scenario: 拦截错误消息明确说明演示模式原因
- **WHEN** `demo-control` 拦截任一系统写请求
- **THEN** 返回的错误消息必须明确说明请求因“演示模式已开启”而被拒绝
- **AND** 前端不得将该场景展示为通用的“没有权限访问此资源”
