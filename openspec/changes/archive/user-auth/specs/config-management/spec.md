## ADDED Requirements

### Requirement: sys_config 加 tenant_id 与覆盖语义
`sys_config` SHALL 新增 `tenant_id INT NOT NULL DEFAULT 0` 字段;UNIQUE 约束改为 `(tenant_id, config_key)`;读路径 platform fallback,写路径默认写当前租户。

#### Scenario: 租户覆盖配置
- **WHEN** 租户 A 设置 `upload.max_size=50MB`
- **THEN** 写入 `(tenant_id=A, key=upload.max_size, value=50MB)`
- **AND** 平台默认 `(tenant_id=0)` 不变

#### Scenario: 租户未覆盖时 fallback
- **WHEN** 租户 B 读取 `upload.max_size` 但未设置覆盖
- **THEN** 返回 `tenant_id=0` 的平台默认值

### Requirement: 平台默认与租户覆盖的接口分层
配置管理接口 SHALL 按权限分层:平台默认管理通过 `/platform/config/*` 仅平台管理员可调用,租户覆盖管理通过 `/tenant/config/*` 由租户管理员调用。

#### Scenario: 租户管理员越权
- **WHEN** 租户管理员尝试 `/platform/config/*`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`

### Requirement: 配置缓存按租户失效
配置缓存 key SHALL 携带 `tenant_id`;租户写入 MUST 触发本租户对应配置缓存失效;平台默认写入 MUST 触发所有租户对应配置缓存级联失效。

#### Scenario: 级联失效
- **WHEN** 平台管理员修改 `tenant_id=0` 的 `upload.max_size`
- **THEN** 所有节点的所有租户对应缓存失效
