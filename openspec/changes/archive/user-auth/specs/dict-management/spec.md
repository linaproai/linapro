## ADDED Requirements

### Requirement: 字典表加 tenant_id 与覆盖开关
`sys_dict_type` 与 `sys_dict_data` SHALL 增加 `tenant_id INT NOT NULL DEFAULT 0`;`sys_dict_type` 同时增加 `allow_tenant_override BOOL NOT NULL DEFAULT FALSE` 控制是否允许租户覆盖。

#### Scenario: 系统枚举字典禁止覆盖
- **WHEN** seed 数据 `dict_type='sys_user_status' allow_tenant_override=false`
- **AND** 租户管理员尝试修改其字典数据
- **THEN** 返回 `bizerr.CodeDictTypeNotOverridable`

#### Scenario: 业务字典允许覆盖
- **WHEN** `dict_type='business_priority' allow_tenant_override=true`
- **AND** 租户 A 添加字典项
- **THEN** 写入 `tenant_id=A` 行,fallback 时优先于 PLATFORM 默认

### Requirement: 字典读路径走 platform fallback
字典数据读取 SHALL 通过 `tenantcap.ReadWithPlatformFallback` 实现"租户覆盖优先 + PLATFORM fallback"语义。

#### Scenario: 租户未覆盖时 fallback
- **WHEN** 租户 B 查询 `business_priority` 但未设置覆盖
- **THEN** fallback 返回 `tenant_id=0` 的平台默认数据

### Requirement: 字典缓存 key 携带租户
字典缓存 key SHALL 形如 `dict:tenant=<id>:type=<type>`;租户写入触发本租户失效;平台默认写入触发全租户失效。

#### Scenario: 平台默认变更级联失效
- **WHEN** 平台管理员修改 `business_priority` 平台默认
- **THEN** 所有租户的对应字典缓存失效

### Requirement: 平台管理员管理平台默认
平台管理员 SHALL 通过 `/platform/dict/*` 接口管理 `tenant_id=0` 的平台默认字典;租户管理员 SHALL 通过 `/dict/*` 接口管理本租户覆盖。

#### Scenario: 租户管理员看不到平台 API
- **WHEN** 租户管理员尝试 `/platform/dict/type/list`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`
