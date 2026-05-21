## MODIFIED Requirements

### Requirement: `plugin.yaml` 保持精简并可声明菜单
系统 SHALL 保持 `plugin.yaml` 聚焦稳定插件元数据和治理信息，并允许声明菜单等需要进入治理投影的内容，但不得要求源码插件在 manifest 中重复声明后端路由清单。

#### Scenario: 源码插件后端路由不在 manifest 中重复维护
- **WHEN** 宿主解析源码插件的 `plugin.yaml`
- **THEN** manifest 不需要列出后端路由清单
- **AND** 后端路由注册代码与 DTO `g.Meta` 仍是源码插件路由的唯一真相源
- **AND** 宿主在注册时捕获路由归属和文档元数据，而不是从 `plugin.yaml` 读取第二套路由声明

### Requirement: 插件列表查询与启动同步必须保持无副作用读边界
系统 SHALL 将插件列表查询视为无副作用的读操作。列表查询可以读取源码插件清单、动态插件注册表、发布快照和治理投影，但不得创建、更新或删除插件治理表数据。插件扫描和治理同步仅能由显式同步操作或宿主启动同步触发；宿主启动同步也 MUST 保持差异驱动，在 registry、release、菜单、权限和资源引用均无差异时不得开启事务、不得写库、不得执行写后回读。

#### Scenario: 从管理页面查询插件列表
- **WHEN** 管理员打开插件管理并调用 `GET /api/v1/plugins`
- **THEN** 系统返回插件列表及当前治理状态
- **AND** 该请求不得写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`

#### Scenario: 显式同步插件时才允许写入治理投影
- **WHEN** 管理员通过 `POST /api/v1/plugins/sync` 触发插件同步
- **THEN** 系统扫描源码插件和动态插件产物
- **AND** 系统可以按 manifest 差异同步注册表、发布快照、资源索引、菜单和权限治理数据

#### Scenario: 启动同步无差异时不产生数据库副作用
- **WHEN** 宿主启动同步发现插件 manifest 与现有 registry、release、菜单、权限和资源引用投影完全一致
- **THEN** 系统不得写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`
- **AND** 系统不得开启空事务
- **AND** 系统不得为了刷新启动快照重复回读同一治理行
