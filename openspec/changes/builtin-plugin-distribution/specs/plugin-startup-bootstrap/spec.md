## ADDED Requirements

### Requirement: 启动引导必须自动收敛 builtin 源码插件

系统 SHALL 在插件路由、cron、动态前端包预热和`plugin.autoEnable`托管普通插件之前，基于 manifest 中`distribution=builtin`的源码插件执行启动收敛。对每个`builtin`插件，系统 MUST 按依赖拓扑顺序自动安装未安装插件、安全升级待升级插件，并启用未启用插件。启动自动安装不得加载`manifest/sql/mock-data/`。

#### Scenario: 未安装 builtin 插件启动时自动安装并启用

- **WHEN** 源码插件声明`distribution: builtin`
- **AND** 注册表中该插件尚未安装
- **THEN** 启动引导执行现有安装生命周期、SQL 迁移、菜单权限同步和资源引用同步
- **AND** 启动引导随后启用该插件
- **AND** 启动自动安装不得执行`manifest/sql/mock-data/`中的 SQL

#### Scenario: 已安装未启用 builtin 插件启动时自动启用

- **WHEN** 源码插件声明`distribution: builtin`
- **AND** 注册表中该插件已安装但未启用
- **THEN** 启动引导执行现有启用生命周期
- **AND** 后续插件路由、cron 和前端包预热将该插件视为已启用

#### Scenario: builtin 插件发现更高版本时启动安全升级

- **WHEN** 已安装`builtin`源码插件的发现版本高于有效版本
- **THEN** 启动引导执行安全升级
- **AND** 成功后新发布成为有效发布
- **AND** 失败时宿主启动失败并保留升级失败诊断

#### Scenario: builtin 插件出现在 plugin.autoEnable 中不重复执行

- **WHEN** 同一插件既声明`distribution: builtin`又出现在`plugin.autoEnable`
- **THEN** 启动引导以`builtin`策略为准
- **AND** 后续`plugin.autoEnable`阶段不得重复安装、升级或启用该插件
- **AND** 系统记录可诊断 warning

### Requirement: builtin 启动收敛必须遵守集群主节点副作用边界

系统 SHALL 在集群模式下仅允许 primary 节点执行`builtin`插件安装、升级、启用、菜单写入、发布切换和共享状态推进。从节点 MUST 等待共享状态收敛并刷新本地 enabled snapshot 和 runtime 投影。

#### Scenario: primary 节点执行 builtin 生命周期写入

- **WHEN** `cluster.enabled=true`
- **AND** primary 节点启动时发现`builtin`插件需要安装、升级或启用
- **THEN** primary 节点执行共享生命周期副作用
- **AND** 成功后发布插件 runtime revision 或等价共享状态变化

#### Scenario: 非 primary 节点等待 builtin 收敛

- **WHEN** `cluster.enabled=true`
- **AND** 非 primary 节点启动时发现`builtin`插件尚未达到目标状态
- **THEN** 该节点等待 primary 写入共享稳定状态或等待窗口超时
- **AND** 该节点不得重复执行安装 SQL、升级 SQL、菜单写入或发布切换

### Requirement: builtin 生命周期变化必须刷新插件派生缓存

系统 SHALL 在`builtin`插件安装、升级或启用成功写入权威治理数据后，刷新插件管理读模型、enabled snapshot、runtime revision 和相关 i18n 运行时消息派生状态。缓存失效 MUST 与权威写入成功耦合，且在集群模式下通过共享修订号、事件或等价机制传播。

#### Scenario: builtin 安装成功后刷新 enabled snapshot

- **WHEN** 启动引导成功安装并启用`builtin`插件
- **THEN** 当前启动上下文中的 enabled snapshot 反映该插件已启用
- **AND** 后续路由接线不需要重新执行全量治理扫描才能看到该插件

#### Scenario: builtin 升级成功后发布 runtime 变化

- **WHEN** 启动引导成功升级`builtin`插件
- **THEN** 系统发布 runtime revision 或等价缓存变化
- **AND** 其他节点和当前会话后续读取使用新有效发布

#### Scenario: builtin 生命周期失败不发布成功缓存

- **WHEN** `builtin`插件安装、升级或启用失败
- **THEN** 系统不得发布表示目标发布已生效的 runtime revision
- **AND** 诊断投影可读取失败状态或失败账本
