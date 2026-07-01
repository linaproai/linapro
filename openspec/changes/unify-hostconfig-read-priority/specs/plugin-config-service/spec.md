## MODIFIED Requirements

### Requirement:宿主公开配置必须通过独立服务读取

系统 SHALL 通过 `HostServices.HostConfig()` 向源码插件暴露宿主配置只读读取能力。源码插件通过该服务读取宿主配置时不得受公开 key 白名单限制；空 key 或 `.` MUST 按宿主配置组件语义返回完整静态配置快照。非 root 配置键的读取顺序 MUST 为当前上下文可见的`sys_config`有效快照、GoFrame 当前静态配置源中的`config.yaml`值、系统已有默认值、`nil`。该服务不得提供写入、保存、热重载或运行时修改宿主配置的方法。动态插件通过 `hostconfig.get` 读取宿主配置时，仍 MUST 先通过 `hostServices` 授权快照校验对应 key，动态插件不得绕过 manifest 授权读取宿主配置。

#### Scenario:源码插件读取任意宿主配置键

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取宿主配置键 `database.default.link`
- **THEN** 系统按宿主当前配置源返回该键的配置值
- **AND** 该读取不要求 key 预先登记到公开白名单

#### Scenario:源码插件优先读取系统配置快照

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `custom.feature.limit`
- **AND** 当前上下文可见的`sys_config`中存在`custom.feature.limit`
- **AND** 静态`config.yaml`中也存在`custom.feature.limit`
- **THEN** 系统返回`sys_config`中的有效值

#### Scenario:源码插件读取静态配置 fallback

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `workspace.basePath`
- **AND** 当前上下文可见的`sys_config`中不存在`workspace.basePath`
- **AND** 静态`config.yaml`中存在`workspace.basePath`
- **THEN** 系统返回静态配置值

#### Scenario:源码插件读取系统默认值 fallback

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取 `sys.jwt.expire`
- **AND** 当前上下文可见的`sys_config`中不存在`sys.jwt.expire`
- **AND** 静态`config.yaml`中不存在`sys.jwt.expire`
- **AND** 系统默认值元数据存在`sys.jwt.expire`
- **THEN** 系统返回系统默认值
- **AND** 该读取不要求 key 预先登记到公开白名单

#### Scenario:源码插件读取缺失宿主配置键

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取不存在的宿主配置键
- **AND** `sys_config`、静态`config.yaml`和系统默认值元数据都没有该 key
- **THEN** 系统返回未找到语义
- **AND** 不因 key 未登记到白名单而返回权限错误

#### Scenario:源码插件读取完整宿主配置快照

- **WHEN** 源码插件通过 `HostServices.HostConfig()` 读取空 key 或 `.`
- **THEN** 系统返回宿主当前配置源中的完整配置快照
- **AND** 该读取不要求逐个 key 预先登记到公开白名单

#### Scenario:动态插件宿主配置读取仍受授权快照限制

- **WHEN** 动态插件通过 `hostconfig.get` 读取宿主配置键
- **THEN** 宿主先按当前 release 的 `hostServices` 授权快照校验该 key
- **AND** 未授权 key 的读取必须被拒绝

#### Scenario:动态插件授权后使用统一读取优先级

- **WHEN** 动态插件通过 `hostconfig.get` 读取已授权的宿主配置键
- **AND** 该 key 同时存在当前上下文可见的`sys_config`值和静态`config.yaml`值
- **THEN** 系统返回`sys_config`中的有效值
- **AND** 授权通过后的读取优先级与源码插件 HostConfig 保持一致
