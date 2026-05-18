## ADDED Requirements

### Requirement: 平台字典与租户字典的可见性规则
租户管理员查询字典类型/数据 SHALL 仅返回:
- `tenant_id = bizctx.TenantId` 的本租户覆盖,以及
- `tenant_id = 0 AND allow_tenant_override` 仅作为 fallback 候选(实际可见性由读路径决定)。

#### Scenario: 租户管理员视图
- **WHEN** 租户 A 管理员查询字典类型列表
- **THEN** 返回 `tenant_id=A 的所有 ∪ tenant_id=0 且 allow_tenant_override=true 的标记为"平台默认"`
- **AND** `tenant_id=0 且 allow_tenant_override=false` 的字典类型隐式存在但不在管理列表中可改

### Requirement: 字典写入权限
租户管理员 SHALL 仅可对 `allow_tenant_override=true` 的字典类型创建/修改 `tenant_id=current` 的字典数据;不可改字典类型 metadata 自身;不可改 `tenant_id=0` 的平台默认。

#### Scenario: 租户管理员写覆盖
- **WHEN** 租户 A 管理员对 `business_priority` 添加字典项 "P0"
- **THEN** 写入 `(tenant_id=A, dict_type=business_priority, value="P0")`

#### Scenario: 修改字典类型 metadata 被拒
- **WHEN** 租户管理员尝试修改 `dict_type` 的 `name` 或 `allow_tenant_override`
- **THEN** 返回 `bizerr.CodePlatformPermissionRequired`

### Requirement: 字典数据 SQL 编排
所有字典 seed SQL SHALL 写入 `tenant_id=0`;mock 数据可选指定租户;现有 SQL 文件按"无历史负担"策略直接修改。

#### Scenario: seed 字典默认平台
- **WHEN** `make init` 执行 seed
- **THEN** 所有 seed 字典数据 `tenant_id=0`
