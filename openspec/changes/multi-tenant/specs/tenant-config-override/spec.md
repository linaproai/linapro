## ADDED Requirements

### Requirement: 平台默认 + 租户覆盖读路径
对于支持租户覆盖的资源(`sys_dict_type` / `sys_dict_data` / `sys_config`),读取 SHALL 按以下顺序解析:
1. 先查 `tenant_id = bizctx.TenantId`,有结果直接返回。
2. 否则 fallback 到 `tenant_id = 0`(PLATFORM 默认),如有则返回。
3. 都无则返回空。

读取逻辑封装在 `tenantcap.ReadWithPlatformFallback` helper 中,业务 service 不重复实现。

#### Scenario: 租户覆盖优先
- **WHEN** 租户 A 设置了 `sys_config[key=upload.max_size, tenant_id=A] = 50MB`
- **AND** 平台默认 `sys_config[key=upload.max_size, tenant_id=0] = 10MB`
- **THEN** 租户 A 的请求读到 `50MB`
- **AND** 其他租户的请求 fallback 到 `10MB`

#### Scenario: 租户未覆盖时 fallback
- **WHEN** 租户 B 没有 `sys_config[key=upload.max_size, tenant_id=B]`
- **THEN** fallback 读 `tenant_id = 0` 的平台默认
- **AND** 返回 `10MB`

### Requirement: 写路径默认写当前租户
租户管理员 SHALL 仅可写自己租户的覆盖记录;平台管理员 SHALL 显式调用 `WritePlatformDefault(...)` 才能写 `tenant_id = 0` 的平台默认。

#### Scenario: 租户管理员写覆盖
- **WHEN** 租户 A 管理员调用 `PUT /tenant/config { key: upload.max_size, value: 50MB }`
- **THEN** 写入 `sys_config(tenant_id=A, key=upload.max_size, value=50MB)`(insert or update)
- **AND** 不影响 `tenant_id = 0` 的平台默认

#### Scenario: 平台管理员显式写平台默认
- **WHEN** 平台管理员调用 `PUT /platform/config/default { key: upload.max_size, value: 20MB }`
- **THEN** 写入 `sys_config(tenant_id=0, key=upload.max_size, value=20MB)`
- **AND** 操作日志以 `oper_type='other'` 记录平台默认配置变更,摘要或载荷包含 `platform_default_change`

#### Scenario: 租户管理员尝试写平台默认
- **WHEN** 租户管理员调用 `PUT /platform/config/default`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`

### Requirement: 字典类型的"是否允许租户覆盖"开关
`sys_dict_type` SHALL 增加 `allow_tenant_override BOOL DEFAULT false` 字段;`false` 时该字典类型仅平台管理员可改且所有租户共享(读路径不 fallback,直接读 `tenant_id = 0`)。

#### Scenario: 不允许覆盖的字典类型
- **WHEN** 字典类型 `sys_user_status`(系统枚举)`allow_tenant_override = false`
- **AND** 租户 A 管理员尝试修改其字典数据
- **THEN** 返回 `bizerr.CodeDictTypeNotOverridable`
- **AND** 修改被拒

#### Scenario: 允许覆盖的字典类型
- **WHEN** 字典类型 `business_priority`(业务优先级)`allow_tenant_override = true`
- **AND** 租户 A 添加自定义字典项
- **THEN** 写入 `sys_dict_data(tenant_id=A, dict_type='business_priority', ...)`
- **AND** 不影响其他租户

### Requirement: 缓存按租户覆盖维度失效
字典缓存 / 配置缓存 key SHALL 携带 `tenant_id`;租户写入覆盖时,触发该 `(tenant_id, dict_type|config_key)` 维度失效;平台写入默认时,触发所有租户的对应维度失效(因为 fallback 影响所有租户)。

#### Scenario: 租户覆盖触发本租户失效
- **WHEN** 租户 A 修改 `sys_config[key=upload.max_size]`
- **THEN** 仅 `(tenant_id=A, key=upload.max_size)` 缓存失效
- **AND** 其他租户的缓存不受影响

#### Scenario: 平台默认变更触发全租户失效
- **WHEN** 平台管理员修改 `sys_config[tenant_id=0, key=upload.max_size]`
- **THEN** 所有 `(tenant_id=*, key=upload.max_size)` 缓存失效
- **AND** 通过 cluster 广播一次性传播

### Requirement: 角色不走 fallback
`sys_role` 按 `tenant_id` 归属平台上下文或租户上下文;角色查询 SHALL 使用当前 `bizctx.TenantId` 作为权威边界,而非 fallback。租户管理员仅可查/改本租户角色,不可查/改平台上下文角色。

#### Scenario: 租户管理员查询角色列表
- **WHEN** 租户 A 管理员调用 `GET /role/list`
- **THEN** 返回 `tenant_id = A` 的角色
- **AND** 不返回平台角色,即使存在
