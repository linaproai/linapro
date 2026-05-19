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

