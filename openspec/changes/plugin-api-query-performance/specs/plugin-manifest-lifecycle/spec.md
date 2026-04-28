## ADDED Requirements

### Requirement: Plugin list query is side-effect free

系统 SHALL 将插件列表查询作为无副作用读操作处理。插件列表查询可以读取已发现的源码插件清单、动态插件注册表、插件 release 快照和治理投影数据，但 MUST NOT 在查询过程中创建、更新或删除插件治理表数据。插件扫描和治理表同步 MUST 由显式同步动作触发。

#### Scenario: Query plugin list from management page

- **WHEN** 管理员访问插件管理页并触发 `GET /api/v1/plugins`
- **THEN** 系统返回插件列表和当前治理状态
- **AND** 系统不得因为本次 GET 请求写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`

#### Scenario: Synchronize plugins explicitly

- **WHEN** 管理员触发插件同步动作并调用 `POST /api/v1/plugins/sync`
- **THEN** 系统扫描源码插件和动态插件产物
- **AND** 系统可以按清单内容同步插件注册表、release 快照、资源索引、菜单和权限治理数据

### Requirement: Plugin host-service metadata lookup does not emit schema probing errors

系统 SHALL 在插件列表投影 hostServices 数据表资源时，以只读方式查询宿主数据库元数据。该查询 MUST NOT 对 `information_schema.TABLES` 触发错误的业务库表结构探测；当数据库不支持该元数据查询或查询失败时，插件列表接口 SHALL 降级为返回原始表名。

#### Scenario: Resolve data table comments for dynamic plugin permissions

- **WHEN** 插件列表包含声明 `data.resources.tables` 的动态插件
- **THEN** 系统尝试读取这些表的注释并用于权限审查展示
- **AND** 查询过程不得产生 `SHOW FULL COLUMNS FROM TABLES` 错误日志

#### Scenario: Metadata lookup unavailable

- **WHEN** 当前数据库方言不支持宿主表注释查询或元数据查询失败
- **THEN** 插件列表接口仍然成功返回
- **AND** hostServices 权限展示使用原始表名作为降级信息
