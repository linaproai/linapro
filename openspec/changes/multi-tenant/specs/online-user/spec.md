## ADDED Requirements

### Requirement: sys_online_session 增加租户字段
`sys_online_session` SHALL 加 `tenant_id` 列;主键改为 `(tenant_id, token_id)`;index `(user_id, tenant_id)`。

#### Scenario: 同用户多租户多会话
- **WHEN** 用户 U 在租户 A 与租户 B 各登录一次
- **THEN** session store 包含两行 `(A, T1, U)` 与 `(B, T2, U)`
- **AND** 两行互不干扰

### Requirement: 在线会话查询按租户过滤
`GET /online/list` SHALL 经 `tenantcap.Apply` 过滤;租户管理员仅可见本租户在线会话。平台管理员仅通过 `/platform/online/list` 管理平台接口查看全量会话并标记 `tenant_id`;impersonation 模式下仍仅可见目标租户会话。

#### Scenario: 租户管理员视图
- **WHEN** 租户 A 管理员查询在线列表
- **THEN** 仅返回 `tenant_id=A` 的会话

### Requirement: 踢人接口按租户校验
`POST /online/{token_id}/kick` SHALL 校验目标 session 的 `tenant_id` 与请求 `tenant_id` 匹配;不匹配返回 403。平台管理员踢除跨租户会话必须使用显式 `/platform/online/{tenant_id}/{token_id}/kick` 接口并记录审计;impersonation 模式不绕过目标租户校验。

#### Scenario: 跨租户踢人被拒
- **WHEN** 租户 A 管理员尝试踢租户 B 的会话
- **THEN** 返回 403 `bizerr.CodeOnlineSessionTenantForbidden`

### Requirement: session 清理任务按租户感知
session 过期清理定时任务 SHALL 按 `(tenant_id, expires_at)` 联合索引扫描;清理粒度按租户独立计费,但任务本身为平台级(`tenant_id=0`)。

#### Scenario: 租户暂停时会话立即失效
- **WHEN** 租户 T 被暂停
- **THEN** `session` store 中 `tenant_id=T` 的所有行立即被删除
- **AND** 该租户用户后续请求被拒
