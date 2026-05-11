## ADDED Requirements

### Requirement: sys_user_role 增加租户维度
`sys_user_role` SHALL 增加 `tenant_id INT NOT NULL DEFAULT 0` 字段;UNIQUE 约束改为 `(user_id, role_id, tenant_id)`,表示"用户 U 在租户 T 内持有角色 R"。

#### Scenario: 用户在不同租户持有不同角色
- **WHEN** 用户 U 在租户 A 是 admin,在租户 B 是普通用户
- **THEN** `sys_user_role` 中存在两行:`(U, role_admin, A)` 与 `(U, role_user, B)`

### Requirement: 角色绑定按租户校验
绑定时 SHALL 校验 `sys_role.tenant_id` 与请求 `tenant_id` 匹配;租户管理员仅可绑定本租户角色;平台上下文角色仅可绑定给平台用户。

#### Scenario: 租户管理员绑定本租户角色
- **WHEN** 租户 A 管理员绑定角色 R(`sys_role.tenant_id=A`)给用户 U
- **THEN** 写入 `(U, R, A)` 一行
- **AND** 校验 U 在租户 A 存在 active membership

#### Scenario: 跨租户绑定被拒
- **WHEN** 租户 A 管理员尝试将租户 B 的角色绑定给某用户
- **THEN** 返回 `bizerr.CodeRoleTenantMismatch`
- **AND** 不写入

#### Scenario: 平台上下文角色绑定到普通用户被拒
- **WHEN** 任意用户尝试将 `tenant_id=0` 的角色绑给租户用户
- **THEN** 返回 `bizerr.CodePlatformRoleAssignmentForbidden`

### Requirement: 权限解析按当前租户过滤
权限/菜单解析 SHALL 仅取 `sys_user_role.tenant_id = bizctx.TenantId` 的关联;平台管理员上下文(`bizctx.TenantId=0`)取平台角色。

#### Scenario: 用户切换租户后权限刷新
- **WHEN** 用户 U 持租户 A token 时拥有角色 R_A
- **AND** 切换到租户 B 后持有角色 R_B
- **THEN** 切换后菜单/按钮权限按 R_B 解析
- **AND** R_A 的权限点不可见

### Requirement: 用户登出/被踢清理状态
当用户从某租户被移除(membership 删除)或租户被删除时,SHALL 级联删除 `sys_user_role.tenant_id = T` 相关行,但保留其他租户的关联。

#### Scenario: 移除 membership 触发清理
- **WHEN** 租户管理员从租户 A 移除用户 U
- **THEN** `sys_user_role(user_id=U, tenant_id=A)` 行被删除
- **AND** `sys_user_role(user_id=U, tenant_id=其他)` 行不受影响
