## Requirements

### Requirement: Plugin Menu and Permission Reuse Lina Governance Modules

The system SHALL require plugin menus, button permissions, and role authorizations to reuse Lina's existing menu and role management systems.

#### Scenario: Plugin registers menus and permissions
- **WHEN** a plugin declares menus, button permissions, and page entries and completes installation
- **THEN** the host registers these into Lina's menu and role authorization system
- **AND** permission identifiers use plugin namespace prefix

#### Scenario: Plugin menus maintain authorization by menu_key
- **WHEN** a plugin installs, upgrades, or uninstalls a menu resource
- **THEN** the host locates `sys_menu` by `menu_key` and maintains `sys_role_menu` authorization

### Requirement: Plugin Role Authorization Persists Across Disable/Enable Cycles

The system SHALL preserve role authorization relationships when a plugin is disabled and restore them when re-enabled.

#### Scenario: Plugin disabled preserves authorization
- **WHEN** a plugin with role authorizations is disabled
- **THEN** the host stops the authorizations from taking effect
- **AND** does not delete the authorization relationships

#### Scenario: Plugin re-enabled restores authorization
- **WHEN** a previously disabled plugin is re-enabled
- **THEN** the host restores menu, button permission, and role authorization to active state

### Requirement: Plugin Runtime Can Access Host Permission Context

The system SHALL provide standardized host permission context for plugins to reuse Lina's user, role, department, and data permission scope.

#### Scenario: Plugin backend processes a request
- **WHEN** a plugin API, hook, or task needs the current user permission context
- **THEN** the host provides user ID, role IDs, menu permission codes, and data permission scope

### Requirement: Plugin Uninstall Only Cleans Governance Resources

The system SHALL remove host governance resources on plugin uninstall but preserve plugin business data by default.

#### Scenario: Uninstall an enabled dynamic plugin
- **WHEN** an administrator uninstalls an enabled dynamic plugin
- **THEN** the host removes menus, resource references, runtime artifacts, and mount info
- **AND** cleans up role-menu relationships
- **AND** preserves plugin business tables and data

### Requirement: 源码插件 AdminService 不使用字符串式管理能力声明

系统 SHALL 允许源码插件通过`pluginhost.Services.Admin()`获取完整类型化管理服务目录。源码插件管理能力不需要通过插件清单或字符串式声明授权；管理方法的安全边界 MUST 由领域 owner 的`AdminService`执行租户、数据权限、状态机、系统 actor 和审计治理。

#### Scenario: 源码插件获取管理目录

- **WHEN** 源码插件注册路由、hook、cron、生命周期或 provider 时需要管理能力
- **THEN** 它通过`pluginhost.Services.Admin().<Domain>()`获取类型化`AdminService`
- **AND** 不需要维护独立字符串式管理能力声明

### Requirement: 动态插件领域方法授权与用户菜单权限分离

系统 SHALL 将动态插件领域方法授权与工作台菜单/按钮权限分离。动态插件领域方法调用由安装或启用阶段确认的`hostServices`授权快照控制；运行时不再额外要求当前用户拥有某个管理工作台菜单或按钮权限。领域方法仍 MUST 执行数据权限、租户边界、目标资源约束、状态机、数量上限和审计治理。

### Requirement: 插件权限治理不得替代数据权限治理

系统 SHALL 明确插件安装授权、插件菜单权限和插件能力授权只决定插件是否可以进入某个领域方法。领域方法对目标记录、租户、组织范围、数据范围、状态机和执行动作的校验 MUST 独立执行，且不得被插件安装授权或菜单授权替代。
