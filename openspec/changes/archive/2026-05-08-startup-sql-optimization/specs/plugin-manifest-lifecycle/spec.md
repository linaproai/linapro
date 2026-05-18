## MODIFIED Requirements

### Requirement: 插件列表查询无副作用

系统 SHALL 将插件列表查询视为无副作用的读操作。列表查询可读取发现的源码清单、动态插件注册表数据、发布快照和治理投影，但不得创建、更新或删除插件治理表数据。插件扫描和治理同步必须仅由显式同步操作或宿主启动同步操作触发。宿主启动同步也 SHALL 是差异驱动的：当插件 registry、release snapshot、菜单、权限和资源引用投影均无差异时，不得开启事务、不得写入数据库、不得执行写后回读。

#### Scenario: 从管理页面查询插件列表

- **当** 管理员打开插件管理并调用 `GET /api/v1/plugins` 时
- **则** 系统返回插件列表和当前治理状态
- **且** GET 请求不写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`

#### Scenario: 显式同步插件

- **当** 管理员通过 `POST /api/v1/plugins/sync` 触发插件同步时
- **则** 系统扫描源码插件和动态插件产物
- **且** 系统可从清单同步注册表、发布快照、资源索引、菜单和权限治理数据

#### Scenario: 启动同步无差异时不产生数据库副作用

- **当** 宿主启动同步发现插件清单与现有治理投影完全一致
- **则** 系统不得为该插件写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`
- **且** 系统不得开启空事务或为了刷新启动快照重复回读同一治理行
