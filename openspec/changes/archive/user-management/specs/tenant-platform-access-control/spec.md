# tenant-platform-access-control Specification

## Purpose
TBD - created by archiving change tenant-permission-boundary-hardening. Update Purpose after archive.
## Requirements
### Requirement: 平台租户控制面必须要求平台上下文

系统 SHALL 将平台租户控制面接口视为平台资源访问边界。调用方除具备对应权限字符串外，还 MUST 处于平台上下文，且当前请求不得是代管租户上下文。租户上下文中的用户即使因历史脏授权或异常角色关系持有 `system:tenant:*` 权限，也 MUST 被拒绝访问平台租户控制面数据。

#### Scenario: 平台上下文读取租户列表

- **WHEN** 平台管理员处于平台上下文并具备 `system:tenant:list` 权限
- **THEN** 系统允许调用平台租户列表接口
- **AND** 返回平台租户治理视图中允许查看的租户数据

#### Scenario: 租户上下文持有异常平台权限仍被拒绝

- **WHEN** 租户用户处于租户上下文
- **AND** 其有效权限快照因历史脏数据包含 `system:tenant:list`
- **THEN** 调用平台租户列表接口 MUST 返回结构化权限错误
- **AND** 响应不得包含任何其他租户的数据

#### Scenario: 平台管理员代管租户时不能操作平台租户控制面

- **WHEN** 平台管理员进入某租户的代管上下文
- **AND** 调用平台租户创建、更新、删除、启停或列表接口
- **THEN** 系统 MUST 按租户上下文拒绝该平台控制面操作
- **AND** 不得因操作者原始身份是平台管理员而绕过当前上下文边界

### Requirement: 平台控制面错误必须可本地化且可审计

平台控制面上下文不满足要求时，系统 SHALL 返回稳定 `bizerr` 业务错误，错误包含机器码、message key、英文 fallback 和必要参数。拒绝事件 SHALL 可被操作日志或安全日志审计定位到调用用户、当前租户上下文和目标平台资源类型。

#### Scenario: 平台上下文缺失错误可本地化

- **WHEN** 租户上下文调用平台租户控制面接口
- **THEN** 响应包含稳定 `errorCode`
- **AND** 响应包含用于运行时翻译的 `messageKey`
- **AND** 中英文错误资源和 apidoc i18n 资源均覆盖该错误

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
