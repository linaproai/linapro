## ADDED Requirements

### Requirement: JWT Claims 携带租户身份
JWT Claims SHALL 增加 `TenantId int` 字段;`TenantId = 0` 表示平台级 token(平台管理员管理平台时使用),`TenantId > 0` 表示绑定到具体租户的 token。

#### Scenario: 普通租户用户登录后的 token
- **WHEN** 租户用户 U 登录到租户 A 成功
- **THEN** 返回 token Claims 包含 `UserId=U`、`TenantId=A`、`TokenId=...`
- **AND** 该 token 在所有非平台 API 上等价于 "U as A" 上下文

#### Scenario: 平台管理员的 token
- **WHEN** 平台管理员登录(`sys_user.tenant_id=0` + 持平台角色)
- **THEN** 返回 token Claims 中 `TenantId=0`
- **AND** 不需要再调 select-tenant

### Requirement: 登录流程的两阶段(认证 → 选租户)
认证成功 SHALL 与"选择租户"作为两个独立步骤:`POST /auth/login` 仅做凭证校验返回 `pre_token` + 用户可见租户列表;`POST /auth/select-tenant` 接受 `pre_token + tenant_id` 校验 membership 后签发正式 JWT。

#### Scenario: 多租户用户的登录流程
- **WHEN** 用户 U 提交用户名密码到 `/auth/login`
- **AND** U 在 multi 模式下有 2 个 active membership
- **THEN** 返回 `{ pre_token, tenants: [{id, code, name}, ...] }`,不返回正式 JWT
- **AND** 前端展示挑选器,用户选定后调 `/auth/select-tenant {pre_token, tenant_id}`
- **AND** 服务端签发正式 JWT 返回

#### Scenario: 单租户用户登录优化
- **WHEN** 用户 U 仅有 1 个 active membership
- **AND** `tenant.cardinality = single` 或 multi 但无歧义
- **THEN** `/auth/login` 直接返回正式 JWT(等同于服务端自动 select)
- **AND** 不需要前端挑选

### Requirement: pre_token 的安全约束
`pre_token` SHALL 是短期(60s)、单次使用、仅可用于 `/auth/select-tenant` 的特殊 token;不允许用于业务 API。

#### Scenario: pre_token 用于业务 API
- **WHEN** 客户端用 `pre_token` 调用任意业务 API
- **THEN** 返回 401 `bizerr.CodePreTokenNotAllowedForBusiness`
- **AND** 操作日志记录一条安全事件

#### Scenario: pre_token 重放
- **WHEN** 客户端用同一个 `pre_token` 第二次调 `/auth/select-tenant`
- **THEN** 返回 401 `bizerr.CodePreTokenAlreadyConsumed`
- **AND** 不签发 token

### Requirement: 切换租户接口
`POST /auth/switch-tenant` SHALL 在已认证状态下重新签发 token(详见 `tenant-membership` 能力);旧 token 立即作废并通过 cluster 广播。

#### Scenario: 切换租户后旧 token 立即失效
- **WHEN** 用户 U 持有 token T1(tenant=A),调用 `/auth/switch-tenant {tenant_id: B}` 返回 token T2(tenant=B)
- **AND** 客户端继续用 T1 请求业务 API
- **THEN** 服务端拒绝 T1(查 revoke 列表)
- **AND** 返回 `bizerr.CodeTokenRevoked`

### Requirement: 平台管理员 impersonation token
平台管理员 SHALL 通过 `POST /platform/tenants/{id}/impersonate` 临时切换为"以租户 T 视角操作";系统签发 impersonation token,标识 `acting_user_id = 平台管理员`,`tenant_id = T`,`is_impersonation = true`。

#### Scenario: 平台管理员开始 impersonation
- **WHEN** 平台管理员调用 `/platform/tenants/A/impersonate`
- **THEN** 返回 impersonation token
- **AND** 持该 token 的请求 `bizctx.TenantId = A`,`bizctx.ActingAsTenant = true`,`bizctx.ActingUserId = 平台管理员`

#### Scenario: impersonation 操作的审计
- **WHEN** 平台管理员以 impersonation token 修改租户 A 的某用户角色
- **THEN** 操作日志写入:`tenant_id=A`,`acting_user_id=平台管理员 user_id`,`on_behalf_of_tenant_id=A`,`action_kind='impersonation'`
- **AND** 该日志在租户管理员视图中标记为"平台管理员代为操作"

### Requirement: 登出与 token 撤销
`POST /auth/logout` SHALL 撤销当前 token,从 session store 删除会话,并广播 cluster 失效;若用户在多个租户有活跃 session,仅撤销当前 token 对应的 session。

#### Scenario: 登出仅撤销当前 token
- **WHEN** 用户 U 持有租户 A token T1 与租户 B token T2,在某客户端用 T1 登出
- **THEN** T1 立即失效,T2 不受影响
- **AND** session store 中仅删 (A, T1) 行
