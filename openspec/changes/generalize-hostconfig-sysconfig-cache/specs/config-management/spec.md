## MODIFIED Requirements

### Requirement: 系统配置运行时读取必须数据驱动并保持缓存一致

系统 SHALL 将`sys_config`作为宿主运行时系统配置的权威数据源。宿主配置服务 MUST 基于共享 revision 和本地快照缓存读取当前上下文可见的`sys_config`有效 key，而不是仅依赖 Go 代码中的硬编码 key 白名单。`sys_config`创建、更新、导入或删除导致有效配置变化后，系统 MUST 推进 runtime-config revision 并使后续读取重建快照。

#### Scenario: 读取自定义系统配置

- **WHEN** `sys_config`中存在 key 为`custom.feature.limit`的当前上下文可见记录
- **THEN** 宿主配置服务通过`GetRaw(ctx, "custom.feature.limit")`返回该记录的值
- **AND** 不需要为该 key 新增 Go 常量或修改硬编码白名单

#### Scenario: 静态配置 fallback

- **WHEN** 当前上下文可见的`sys_config`中不存在`workspace.basePath`
- **AND** 静态`config.yaml`中存在`workspace.basePath`
- **THEN** 宿主配置服务通过`GetRaw(ctx, "workspace.basePath")`返回静态配置值

#### Scenario: 租户覆盖优先于平台默认

- **WHEN** 当前上下文为租户`tenant-a`
- **AND** `sys_config`同时存在平台 key`custom.feature.limit=100`和租户 key`custom.feature.limit=50`
- **THEN** 宿主配置服务返回`50`

#### Scenario: 配置更新后刷新快照

- **WHEN** `sys_config`中`custom.feature.limit`从`100`更新为`200`
- **THEN** 系统推进 runtime-config revision
- **AND** 后续宿主配置读取返回`200`

#### Scenario: 配置删除后刷新快照

- **WHEN** `sys_config`中`custom.feature.limit`被删除
- **THEN** 系统推进 runtime-config revision
- **AND** 后续宿主配置读取不再返回被删除值
