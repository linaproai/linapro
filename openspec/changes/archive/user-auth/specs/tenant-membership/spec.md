## ADDED Requirements

### Requirement: 1:N 用户-租户成员关系
`multi-tenant` 插件 SHALL 维护 `plugin_multi_tenant_user_membership(user_id, tenant_id, status, joined_at)` 表,UNIQUE 约束 `(user_id, tenant_id)`,允许同一用户拥有多条 active membership。

#### Scenario: 用户加入第二个租户
- **WHEN** 用户 U 已是租户 A 成员,加入租户 B
- **THEN** `user_membership` 表新增 `(U, B, active, ...)` 一行
- **AND** U 登录后可选择进入 A 或 B

### Requirement: 1:1 兼容策略预留
系统 SHALL 以代码默认值 `multi` 运行,允许同一用户拥有多条 active membership;`single` 作为后续受控管理设置保留。启用 `single` 策略后,系统 SHALL 拒绝任何会让用户拥有多条 active membership 的写入操作。

#### Scenario: single 模式下添加第二条 membership
- **WHEN** `single` 策略启用且用户 U 已有 active membership 在租户 A
- **AND** 平台管理员尝试将 U 加入租户 B
- **THEN** 返回 `bizerr.CodeMembershipExceedsCardinality`
- **AND** 写入被拒

### Requirement: 平台管理员是不带 membership 的特殊用户
`tenant_id=0` 的平台上下文角色 SHALL 仅可分配给 `sys_user.tenant_id = 0` 的用户;平台管理员用户 MUST NOT 拥有 `plugin_multi_tenant_user_membership` 行。

#### Scenario: 平台管理员尝试加入租户
- **WHEN** 平台管理员 U(tenant_id=0)被尝试加入租户 A
- **THEN** 返回 `bizerr.CodePlatformUserCannotJoinTenant`
- **AND** 写入被拒;若需让 U 在租户内操作,应使用 impersonation

### Requirement: 租户内管理能力
租户内管理能力 SHALL 由当前租户上下文、角色数据权限与 `system:*` 功能权限共同推导,不在 membership 表中维护额外管理员布尔字段。

#### Scenario: 租户管理员可见的菜单
- **WHEN** 租户管理员登录到该租户
- **THEN** 菜单通过"用户管理"承载本租户成员关系管理,并可继续使用"角色管理"等租户级管理项
- **AND** 不包含独立"租户工作台"目录、`/tenant/members`、`/tenant/plugins` 或"租户管理"、"系统插件安装"等平台级菜单

### Requirement: 用户被踢出租户
当一条 membership 被删除或 status 置为 `removed` 时,系统 SHALL 立即作废该用户在该租户的所有 token、session 与权限缓存,但保留全局 `sys_user` 与其他租户的 membership。

#### Scenario: 移除 membership 触发会话失效
- **WHEN** 租户管理员从租户 A 中移除用户 U
- **THEN** U 在租户 A 的会话 / token 立即失效
- **AND** U 在其他租户的会话不受影响
- **AND** 操作日志记录 `tenant_id = A`,`acting_user_id = 租户管理员`

### Requirement: 用户可见租户列表
`GET /auth/login-tenants` 与 `GET /tenant/membership/me` SHALL 返回当前用户所有 `status=active` 的 membership 与对应租户基础信息(id、code、name)。

#### Scenario: 1:N 用户登录后选租户
- **WHEN** 用户 U 登录认证成功
- **AND** U 有 2 条 active membership
- **THEN** `/auth/login-tenants` 返回长度为 2 的列表
- **AND** 前端展示挑选器,选定后调 `/auth/select-tenant`

### Requirement: 切换租户的 token 重签
`POST /auth/switch-tenant` SHALL 接受 `target_tenant_id`,校验当前用户在目标租户存在 active membership,然后:
1. 加入旧 token 到 revoke 列表(短期 + cluster 广播)。
2. 重新签发携带新 `TenantId` 的 JWT。
3. 删除旧 session,创建新 session。
4. 返回新 token 与新菜单/权限。

#### Scenario: 1:N 用户切换租户成功
- **WHEN** 用户 U 持有租户 A 的 token,调用 `/auth/switch-tenant {target_tenant_id: B}`
- **AND** U 在 B 有 active membership
- **THEN** 旧 token 立即失效
- **AND** 返回新 token,Claims 中 `TenantId = B`

#### Scenario: 切换到无 membership 的租户
- **WHEN** 用户 U 调用 `/auth/switch-tenant {target_tenant_id: C}`
- **AND** U 在 C 没有 active membership
- **THEN** 返回 `bizerr.CodeTenantMembershipMissing`
- **AND** 旧 token 不被撤销
