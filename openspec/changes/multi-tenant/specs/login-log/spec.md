## ADDED Requirements

### Requirement: 登录日志加租户与 impersonation 字段
`monitor-loginlog` 表 SHALL 增加:
- `tenant_id INT NOT NULL DEFAULT 0`:登录到的租户。
- `acting_user_id INT`:实际操作人(impersonation 时 = 平台管理员)。
- `on_behalf_of_tenant_id INT`:impersonation 时 = 目标租户。
- `is_impersonation BOOL`。

#### Scenario: 普通登录
- **WHEN** 用户 U 登录到租户 A
- **THEN** loginlog 记录 `(tenant_id=A, user_id=U, acting_user_id=U, is_impersonation=false)`

#### Scenario: impersonation 登录
- **WHEN** 平台管理员 P 通过 impersonate 进入租户 A
- **THEN** loginlog 记录 `(tenant_id=A, user_id=P, acting_user_id=P, on_behalf_of_tenant_id=A, is_impersonation=true)`

### Requirement: 登录日志查询按租户隔离
登录日志查询接口 SHALL 经 `tenantcap.Apply` 过滤;租户管理员仅可见本租户的登录日志;平台管理员可见全量并支持按租户筛选。

#### Scenario: 跨租户日志不可见
- **WHEN** 租户 A 管理员查询登录日志
- **THEN** 仅返回 `tenant_id=A` 的记录

### Requirement: 日志列表展示 impersonation 标记
租户管理员视图中,impersonation 记录 SHALL 显示"平台管理员代为操作"角标,便于审计;平台管理员视图中显示 acting_user 与 on_behalf_of_tenant 详情。

#### Scenario: 角标渲染
- **WHEN** 租户管理员查询登录日志
- **AND** 列表中存在 `is_impersonation=true` 的行
- **THEN** UI 列表该行显示"代为操作"角标 + 平台管理员 username
