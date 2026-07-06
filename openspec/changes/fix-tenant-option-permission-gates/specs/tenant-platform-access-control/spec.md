## ADDED Requirements

### Requirement: 前端租户候选加载必须遵守平台控制面权限门禁

系统 SHALL 在前端租户选项、头部租户切换器和宿主工作台适配层加载租户候选时先检查当前权限快照和租户上下文。缺少 `system:tenant:list` 时，前端 MUST NOT 请求平台租户列表接口；缺少 `system:tenant:auth:login-tenants` 时，前端 MUST NOT 请求登录租户候选接口。租户上下文中的页面 SHOULD 优先使用当前租户上下文完成渲染，不得因为候选租户列表为空或尚未加载而回退请求平台租户控制面。

#### Scenario: 租户管理员访问用户管理

- **WHEN** `linapro-tenant-core` 已启用
- **AND** 用户处于具体租户上下文
- **AND** 当前权限快照不包含 `system:tenant:list` 和 `system:tenant:auth:login-tenants`
- **AND** 用户打开用户管理页面
- **THEN** 前端不得请求平台租户列表接口
- **AND** 前端不得请求登录租户候选接口
- **AND** 页面使用当前租户上下文完成用户管理数据加载

#### Scenario: 租户管理员访问角色管理

- **WHEN** `linapro-tenant-core` 已启用
- **AND** 用户处于具体租户上下文
- **AND** 当前权限快照不包含 `system:tenant:list` 和 `system:tenant:auth:login-tenants`
- **AND** 用户打开角色管理页面
- **THEN** 头部租户切换器不得在布局加载时请求平台租户列表接口
- **AND** 头部租户切换器不得请求登录租户候选接口

#### Scenario: 平台管理员具备平台租户列表权限

- **WHEN** 用户处于平台上下文
- **AND** 当前权限快照包含 `system:tenant:list`
- **AND** 前端需要展示平台租户候选
- **THEN** 前端可以请求平台租户列表接口
- **AND** 请求结果仅用于平台上下文租户候选展示
