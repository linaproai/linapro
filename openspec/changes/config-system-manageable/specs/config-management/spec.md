## ADDED Requirements

### Requirement: 系统参数管理面仅展示可系统维护的配置

系统 SHALL 为每条 `sys_config` 记录持久化 `system_manageable` 字段（`SMALLINT`，`1` 表示允许在系统参数设置管理面维护，`0` 表示不允许）。系统参数管理面的 List 与 Export MUST 仅返回 `system_manageable = 1` 的可见行。运行时配置读取 MUST 不受该字段影响。

#### Scenario: 列表不返回插件闭环配置

- **WHEN** 管理员打开系统参数设置列表
- **AND** 存在 `system_manageable = 0` 的行
- **THEN** 列表不包含这些行

### Requirement: 系统参数管理面不得变更不可系统维护的配置

对 `system_manageable = 0` 的行，管理面 Get/Update/Delete/Import 覆盖 MUST 拒绝或视为不存在。管理面 Create MUST 写入 `system_manageable = 1`。

#### Scenario: 管理面更新被拒绝

- **WHEN** 调用方对 `system_manageable = 0` 的配置请求管理面更新
- **THEN** 系统返回错误且 value 不变

### Requirement: 插件 SetValue 支持显式 SystemManageable

插件经 `HostConfig.SysConfig().SetValue(ctx, key, value, options)` 或 `BatchSetValue(ctx, items, options)` 写入时，`options` 可为 nil 或 `*SetSysConfigValueOptions`。当 `options` 为 nil 或 `options.SystemManageable` 为 nil 且首次插入时 MUST 写 `0`；更新时 MUST 保持原标记；当 `SystemManageable` 非 nil 时 MUST 写入对应标记。仅在插件入口维护的业务配置 MUST 传 `false`。

#### Scenario: 插件闭环写入

- **WHEN** 插件 `SetValue`/`BatchSetValue` 且 `options.SystemManageable = false` 或未指定（首次插入）
- **THEN** 行的 `system_manageable = 0` 且不出现在系统参数列表

#### Scenario: 插件显式进入系统参数页

- **WHEN** 插件 `SetValue` 且 `options.SystemManageable = true`
- **THEN** 行的 `system_manageable = 1`

### Requirement: 插件批量设置必须单事务单 revision

系统 SHALL 提供 `BatchSetValue`，在一次事务中写入全部 items，并在全部成功后仅推进一次 runtime-config revision。空 items MUST 成功且无副作用。多字段插件 settings 保存 MUST 使用 `BatchSetValue` 而非循环 `SetValue`。

#### Scenario: 批量写入多键

- **WHEN** 插件一次 `BatchSetValue` 写入多个 key
- **THEN** 所有 key 在同一事务中落库
- **AND** runtime-config revision 仅推进一次
