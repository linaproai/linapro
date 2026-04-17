# plugin-permission-governance Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: 插件菜单与权限复用 LinaPro 通用治理模块
系统 SHALL 要求插件菜单、按钮权限与角色授权复用 LinaPro 现有的菜单管理和角色管理体系。

#### Scenario: 插件注册菜单与权限
- **WHEN** 插件声明一组菜单、按钮权限和页面入口并完成运行时安装，或源码插件初始化 SQL 被宿主执行
- **THEN** 宿主将这些治理资源注册到 LinaPro 的菜单与角色授权体系中
- **AND** 插件权限标识使用插件命名空间前缀
- **AND** 插件菜单在授权前不会对无权限用户可见

#### Scenario: 插件菜单按 `menu_key` 维护授权关系
- **WHEN** 插件安装、升级或卸载一个菜单资源
- **THEN** 宿主通过 `menu_key` 定位对应 `sys_menu` 记录并维护 `sys_role_menu` 授权关系
- **AND** 插件不应通过硬编码整型 `menu_id` 维护角色授权

### Requirement: 插件角色授权在停用与重启用之间可恢复
系统 SHALL 在插件禁用时保留其角色授权关系，并在插件重新启用后恢复生效。

#### Scenario: 插件禁用后保留授权
- **WHEN** 一个已分配角色授权的插件被禁用
- **THEN** 宿主停止这些授权的实际生效
- **AND** 宿主不删除原有角色与插件菜单之间的授权关系

#### Scenario: 插件重新启用后恢复授权
- **WHEN** 一个之前被禁用的插件重新启用
- **THEN** 宿主恢复该插件菜单、按钮权限与角色授权的实际生效状态
- **AND** 管理员不需要重新分配角色菜单关系

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

### Requirement: 插件卸载时只清理治理资源，不默认删除业务数据
系统 SHALL 在插件卸载时移除宿主治理资源，但默认保留插件业务数据与历史记录。

#### Scenario: 卸载一个已启用动态插件
- **WHEN** 管理员卸载一个已启用动态插件
- **THEN** 宿主移除其菜单入口、资源引用、运行时产物和宿主挂载关系
- **AND** 宿主清理插件相关的角色菜单生效关系
- **AND** 宿主默认保留插件业务表和业务数据，除非管理员显式执行额外清理

