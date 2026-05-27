## ADDED Requirements

### Requirement:插件公共数据能力不得暴露宿主数据库模型

系统 SHALL 确保插件公共数据能力和普通 capability 消费接口不暴露`*gdb.Model`、原始 SQL、DAO、DO、Entity 或宿主内部查询对象。需要数据库侧数据范围注入的宿主逻辑 MUST 通过宿主内部窄接口完成，动态插件数据访问 MUST 继续通过受治理 data service 和`hostServices`授权快照完成。

#### Scenario:普通插件消费组织或租户能力

- **WHEN** 普通源码插件或动态插件读取组织、租户或数据相关能力
- **THEN** 返回值使用 DTO、值对象、批量投影或结构化 data service 响应
- **AND** 不返回`*gdb.Model`、DAO、DO、Entity 或可拼接 SQL 的对象

#### Scenario:宿主内部执行数据范围过滤

- **WHEN** 宿主 service 需要在数据库查询阶段注入组织或租户数据范围
- **THEN** 该 service 使用宿主内部`ScopeService`等窄接口
- **AND** 过滤必须在数据库查询阶段完成，不能先查询全量数据再在内存过滤

#### Scenario:动态插件访问数据

- **WHEN** 动态插件需要查询或变更宿主授权的数据表
- **THEN** 它通过`pkg/plugin/capability/guest.Directory.Data()`获取受治理的`pkg/plugin/capability/data` facade 并发起结构化请求
- **AND** 宿主按当前 release 的`hostServices`授权快照、用户上下文和数据权限执行治理
- **AND** guest 不能获取宿主数据库连接、`gdb.Model`或 raw SQL 执行入口

### Requirement:数据能力接口必须优先支持批量投影和有界装配

系统 SHALL 为组织、租户和插件数据访问提供批量投影或结构化分页契约，避免普通插件或宿主列表路径通过循环调用单条详情、逐项 provider 查询或前端瀑布式调用完成装配。

#### Scenario:用户列表装配组织和租户投影

- **WHEN** 宿主或插件需要为用户列表装配部门、岗位或租户标签
- **THEN** 它使用批量投影接口一次性传入当前页用户 ID 集合
- **AND** provider 或宿主实现通过集合化查询、投影查询、缓存或快照完成装配
- **AND** 不得对每个用户循环调用单项详情方法

#### Scenario:动态插件列表查询

- **WHEN** 动态插件通过 data service 查询列表
- **THEN** 宿主在数据库侧完成授权、过滤、排序和分页
- **AND** 只返回当前接口需要的字段或稳定投影
- **AND** 不得先加载大集合到内存后分页或过滤
