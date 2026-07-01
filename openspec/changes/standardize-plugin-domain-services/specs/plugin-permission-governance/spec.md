## MODIFIED Requirements

### Requirement: 插件运行时可获取宿主权限上下文

系统 SHALL 为插件提供标准化的宿主权限上下文，使插件能够复用 LinaPro 的用户、角色、菜单权限与数据权限范围；插件自有资源接口不得额外要求插件管理后台的通用治理权限。源码插件和动态插件进入宿主领域方法时，领域方法 MUST 仅从标准业务`ctx`读取当前用户、租户、权限和数据权限上下文，并执行目标数据边界、租户、组织范围和状态机治理。动态插件的`hostServices`授权摘要 MUST 保留在 dispatcher 内部校验，不得作为额外领域上下文传入 owner。

#### Scenario: 查询插件自有资源数据

- **WHEN** 一个已登录用户访问`/plugins/{pluginId}/resources/{resource}`这类插件自有资源接口
- **THEN** 宿主按该插件资源声明的权限或默认推导出的插件资源权限校验访问
- **AND** 若用户拥有对应的插件菜单/按钮权限，则允许继续执行插件资源查询
- **AND** 宿主不得额外要求用户拥有`plugin:query`这类插件管理后台查询权限

#### Scenario: 缺少插件资源权限时拒绝访问

- **WHEN** 一个用户请求插件自有资源接口但没有对应的插件资源权限
- **THEN** 宿主返回权限不足错误
- **AND** 不返回插件资源数据

#### Scenario: 插件调用宿主领域方法

- **WHEN** 源码插件或动态插件调用宿主领域方法
- **THEN** 领域方法从标准业务`ctx`读取当前用户、租户、权限和数据权限上下文
- **AND** 动态插件 dispatcher 在进入领域方法前完成`service + method + resource`授权校验
- **AND** 领域方法不得因为调用方是插件而跳过目标数据边界、租户边界和状态机治理

### Requirement: 动态插件领域方法授权与用户菜单权限分离

系统 SHALL 将动态插件领域方法授权与工作台菜单/按钮权限分离。动态插件领域方法调用由动态`host service registry`注册事实和安装或启用阶段确认的`hostServices`授权快照控制；运行时不再额外要求当前用户拥有某个管理工作台菜单或按钮权限。领域方法仍 MUST 执行数据权限、租户边界、目标资源约束、状态机和数量上限治理。

#### Scenario: 已授权动态插件调用管理方法

- **WHEN** 动态插件已在安装阶段获得某个已注册领域管理方法授权
- **THEN** 运行时调用该方法不再校验当前用户是否拥有对应工作台菜单权限
- **AND** 领域方法继续校验目标数据边界和状态机

#### Scenario: 未授权动态插件调用管理方法

- **WHEN** 动态插件未在运行时授权快照中拥有对应领域管理方法
- **THEN** 宿主返回能力拒绝错误
- **AND** 不进入领域方法实现

#### Scenario: 未注册动态插件方法被拒绝

- **WHEN** 动态插件声明或调用未注册到动态`host service registry`的领域方法
- **THEN** 宿主返回能力不可用或能力拒绝错误
- **AND** 不得因为当前用户拥有某个菜单或按钮权限而放行

## ADDED Requirements

### Requirement: 源码插件统一 Service 不使用 hostServices 授权声明

系统 SHALL 允许源码插件通过`pluginhost.Services`获取完整类型化统一`Service`目录。源码插件不需要通过插件清单、`hostServices`或`plugin.Admin().Require(...)`等字符串式声明授权管理能力；统一`Service`方法的安全边界 MUST 由领域 owner 从标准业务`ctx`读取当前用户、租户、数据权限和系统调用标识后执行租户、数据权限、状态机和数量上限治理。

#### Scenario: 源码插件获取统一服务目录

- **WHEN** 源码插件注册路由、hook、cron、生命周期或 provider 时需要宿主领域能力
- **THEN** 它通过`pluginhost.Services.<Domain>()`获取类型化统一`Service`
- **AND** 不需要维护独立字符串式管理能力声明
- **AND** 不得通过`pluginhost.Services.Admin()`获取管理目录

#### Scenario: 源码插件管理方法执行领域治理

- **WHEN** 源码插件调用统一`Service`中的管理动作
- **THEN** 领域方法使用标准业务`ctx`中的当前用户、租户、数据权限和系统调用标识执行校验
- **AND** 不得因为调用方是源码插件而跳过目标数据边界和状态机校验

## REMOVED Requirements

### Requirement: 源码插件 AdminService 不使用字符串式管理能力声明

**Reason**: `AdminService`和`pluginhost.Services.Admin()`被废除，源码插件通过统一领域`Service`调用读取、执行、写入和管理方法。

**Migration**: 源码插件调用点迁移到`pluginhost.Services.<Domain>()`和最窄统一`*cap.Service`注入；删除`AdminService`类型、`Services.Admin()`目录和相关测试替身。
