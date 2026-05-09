## ADDED Requirements

### Requirement: 登录流程支持租户感知
登录接口 `POST /auth/login` SHALL 在多租户能力启用时按"先认证、后选租户"两阶段执行;具体流程参见 `tenant-aware-authentication` 能力规范。

#### Scenario: 多租户启用时的两阶段登录
- **WHEN** `multi-tenant` 插件启用且用户有多个 active membership
- **THEN** `/auth/login` 仅返回 `pre_token` + 可见租户列表
- **AND** 客户端必须再调 `/auth/select-tenant` 才能拿到正式 JWT

#### Scenario: 多租户未启用时保持原行为
- **WHEN** `multi-tenant` 插件未启用
- **THEN** `/auth/login` 直接返回正式 JWT,Claims 中 `TenantId=0`
- **AND** 不返回 pre_token,前端流程与单租户时代等价

### Requirement: JWT Claims 必须携带 TenantId
JWT 签发逻辑 SHALL 在 Claims 中包含 `TenantId`;签发时强制写入 `bizctx.TenantId` 当前值,禁止留空或默认 0(除非确实是平台管理员管理平台模式)。

#### Scenario: 租户用户 token 携带租户
- **WHEN** 租户 A 用户登录到租户 A 成功
- **THEN** JWT Claims `TenantId = A`
- **AND** 反序列化后写入 `bizctx.TenantId`

#### Scenario: token 篡改 TenantId 被拒
- **WHEN** 客户端伪造 token,Claims `TenantId` 与 session store 不一致
- **THEN** 鉴权中间件拒绝(返回 401)
- **AND** 日志记录潜在攻击事件

### Requirement: 切换租户接口
系统 SHALL 提供 `POST /auth/switch-tenant {target_tenant_id}` 接口,语义参见 `tenant-aware-authentication` 与 `tenant-membership` 规范;切换成功时 MUST 立即作废旧 token 并签发新 token。

#### Scenario: 切换成功
- **WHEN** 用户调用 `/auth/switch-tenant`
- **THEN** 返回新 token,旧 token 加入 revoke 列表

### Requirement: 平台管理员 impersonation 接口
系统 SHALL 提供 `POST /platform/tenants/{id}/impersonate` 接口,仅平台管理员可调用并签发 impersonation token;非平台管理员调用 MUST 返回 403。

#### Scenario: 平台管理员开始 impersonation
- **WHEN** 平台管理员调用 impersonate 接口
- **THEN** 返回 impersonation token,Claims 标记 `is_impersonation=true`
