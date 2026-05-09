## ADDED Requirements

### Requirement: sys_role 增加租户与平台角色标识
`sys_role` SHALL 新增以下字段:
- `tenant_id INT NOT NULL DEFAULT 0`:平台角色 = 0,租户角色 = 所属租户。
- `is_platform_role BOOL NOT NULL DEFAULT FALSE`:与 `tenant_id` 一致性约束:`is_platform_role=true` ⇔ `tenant_id=0`。

#### Scenario: 平台角色与租户角色共存
- **WHEN** 系统初始化
- **THEN** seed 数据写入"平台超级管理员"`(is_platform_role=true, tenant_id=0)`
- **AND** 创建租户 T 时自动 provision "租户管理员"模板角色 `(is_platform_role=false, tenant_id=T)`

### Requirement: 角色查询按层级隔离
租户管理员调用 `GET /role/list` SHALL 仅返回 `tenant_id = current_tenant AND is_platform_role = false` 的角色;平台管理员可查全量。

#### Scenario: 租户管理员视图
- **WHEN** 租户 A 管理员查询角色列表
- **THEN** 返回租户 A 的角色,不含平台角色

#### Scenario: 平台管理员视图
- **WHEN** 平台管理员调用 `GET /platform/roles`
- **THEN** 返回所有租户的所有角色,带 `tenant_id` 列展示

### Requirement: 平台权限点前缀约定
权限点字符串 SHALL 使用前缀区分平台权限与租户权限:
- `platform:*`:平台级权限,只能挂在 `is_platform_role=true` 角色上。
- 其他前缀(如 `system:user:list`、`monitor:loginlog:list`):租户级权限,只能挂在 `is_platform_role=false` 角色上。

#### Scenario: 错误的权限挂载被拒
- **WHEN** 创建租户角色时挂上 `platform:tenant:create` 权限
- **THEN** 返回 `bizerr.CodePlatformPermissionOnTenantRole`
- **AND** 创建被拒

#### Scenario: 平台权限点的可见性
- **WHEN** 租户管理员查询可挂权限列表
- **THEN** 不返回 `platform:*` 前缀的权限

### Requirement: 角色名/编码作用域
角色 `name` 与 `code` SHALL 在 `(tenant_id, code)` 上唯一(平台角色全局唯一,租户角色租户内唯一);不同租户可重复使用同名角色。

#### Scenario: 同名角色跨租户允许
- **WHEN** 租户 A 创建角色 `code=manager`
- **AND** 租户 B 也创建角色 `code=manager`
- **THEN** 两次创建均成功
- **AND** 各自独立存在

### Requirement: 删除角色的租户校验
租户管理员 SHALL 仅可删除本租户角色;平台管理员 MAY 删除任意角色;删除前 MUST 校验该角色未绑定任何用户(否则需先解绑或显式 cascade)。

#### Scenario: 跨租户删除角色被拒
- **WHEN** 租户 A 管理员尝试删除租户 B 的角色
- **THEN** 返回 `bizerr.CodeRoleTenantForbidden`
