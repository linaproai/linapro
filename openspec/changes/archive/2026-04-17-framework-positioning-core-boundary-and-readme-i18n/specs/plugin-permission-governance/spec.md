# plugin-permission-governance

## MODIFIED Requirements

### Requirement: 插件运行时可获取宿主权限上下文
系统 SHALL 为插件提供标准化的宿主权限上下文，使插件能够复用 LinaPro 的用户、角色、菜单权限与数据权限范围；插件自有资源接口不得额外要求插件管理后台的通用治理权限。

#### Scenario: 查询插件自有资源数据

- **WHEN** 一个已登录用户访问 `/plugins/{pluginId}/resources/{resource}` 这类插件自有资源接口
- **THEN** 宿主按该插件资源声明的权限或默认推导出的插件资源权限校验访问
- **AND** 若用户拥有对应的插件菜单/按钮权限，则允许继续执行插件资源查询
- **AND** 宿主不得额外要求用户拥有 `plugin:query` 这类插件管理后台查询权限

#### Scenario: 缺少插件资源权限时拒绝访问

- **WHEN** 一个用户请求插件自有资源接口但没有对应的插件资源权限
- **THEN** 宿主返回权限不足错误
- **AND** 不返回插件资源数据
