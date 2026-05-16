## ADDED Requirements

### Requirement: 插件启动扫描必须标记运行时升级状态

系统 SHALL 在宿主启动阶段扫描源码插件和动态插件的文件元数据，并比较数据库中的有效插件版本与文件中发现的目标版本。启动扫描不得自动执行升级、不得切换有效 release、不得运行升级 SQL，也不得因为发现可升级版本而阻断宿主启动。

#### Scenario: 发现文件版本高于数据库有效版本
- **WHEN** 宿主启动时数据库记录插件 `plugin-demo` 的有效版本为 `v0.1.0`
- **AND** 插件文件中的 manifest 版本为 `v0.2.0`
- **THEN** 系统将插件运行时状态标记为 `pending_upgrade`
- **AND** `sys_plugin.version` 和有效 `release_id` 仍保持 `v0.1.0`
- **AND** 宿主服务继续启动

#### Scenario: 发现文件版本低于数据库有效版本
- **WHEN** 宿主启动时数据库记录插件 `plugin-demo` 的有效版本为 `v0.2.0`
- **AND** 插件文件中的 manifest 版本为 `v0.1.0`
- **THEN** 系统将插件运行时状态标记为 `abnormal`
- **AND** 系统记录异常原因为文件版本低于数据库有效版本
- **AND** 宿主服务继续启动并要求管理员人工修复

#### Scenario: 发现版本与有效版本一致
- **WHEN** 宿主启动时数据库记录插件 `plugin-demo` 的有效版本为 `v0.2.0`
- **AND** 插件文件中的 manifest 版本为 `v0.2.0`
- **THEN** 系统将插件运行时状态标记为 `normal`
- **AND** 插件继续按已安装和启用状态参与运行时加载

### Requirement: 待升级插件的业务入口必须受控

系统 SHALL 在插件处于 `pending_upgrade`、`abnormal` 或 `upgrade_failed` 状态时保护插件业务入口，避免目标版本代码在旧数据库或异常元数据上直接运行。宿主基础插件管理接口和升级接口 MUST 保持可用，以便管理员完成升级或修复。

#### Scenario: 待升级源码插件存在业务路由
- **WHEN** 插件 `plugin-demo` 处于 `pending_upgrade`
- **AND** 用户访问该插件声明的业务 API 或页面路由
- **THEN** 系统阻止该插件业务入口正常执行
- **AND** 系统返回稳定的升级所需状态或隐藏对应菜单入口
- **AND** 插件管理页仍可展示并触发升级

#### Scenario: 异常插件存在 cron 任务
- **WHEN** 插件 `plugin-demo` 处于 `abnormal`
- **AND** 插件声明了 cron 任务
- **THEN** 系统不得调度该插件的 cron 任务
- **AND** 插件管理页展示异常原因和人工修复提示

### Requirement: 插件列表必须暴露有效版本和发现版本

系统 SHALL 在插件管理列表和详情响应中暴露插件运行时升级状态、数据库有效版本、文件发现版本、是否可升级、异常原因和最近升级失败信息。该状态 MUST 独立于插件安装状态和启用状态。

#### Scenario: 管理页查询待升级插件
- **WHEN** 管理页请求插件列表
- **AND** 插件 `plugin-demo` 的有效版本为 `v0.1.0` 且发现版本为 `v0.2.0`
- **THEN** 响应包含 `runtimeState=pending_upgrade`
- **AND** 响应包含 `effectiveVersion=v0.1.0`
- **AND** 响应包含 `discoveredVersion=v0.2.0`
- **AND** 响应包含 `upgradeAvailable=true`

#### Scenario: 管理页查询异常插件
- **WHEN** 管理页请求插件列表
- **AND** 插件 `plugin-demo` 的运行时状态为 `abnormal`
- **THEN** 响应包含稳定的异常原因代码
- **AND** 响应不得把异常状态混入 `installed` 或 `enabled` 字段

### Requirement: 插件管理页必须提供运行时升级操作

系统 SHALL 在插件管理页为 `pending_upgrade` 插件展示升级操作，并在执行升级前展示升级内容确认弹窗。升级完成后插件状态 SHALL 变为 `normal`；升级失败时 SHALL 展示失败状态和可诊断错误。

#### Scenario: 待升级插件显示升级按钮
- **WHEN** 管理员打开插件管理页
- **AND** 插件 `plugin-demo` 已安装且运行时状态为 `pending_upgrade`
- **THEN** 原安装动作位置展示为升级动作
- **AND** 管理员点击升级时系统打开升级确认弹窗

#### Scenario: 升级确认弹窗展示升级内容
- **WHEN** 管理员准备升级插件 `plugin-demo`
- **THEN** 弹窗展示升级前版本、目标版本、manifest 变化摘要、依赖检查结果、hostServices 授权变化和升级风险提示
- **AND** 管理员确认前系统不得执行升级副作用

#### Scenario: 升级成功后状态恢复正常
- **WHEN** 管理员确认升级插件 `plugin-demo`
- **AND** 升级流程全部成功
- **THEN** 插件运行时状态变为 `normal`
- **AND** 插件有效版本变为目标 manifest 版本
- **AND** 管理页刷新后不再展示升级按钮

#### Scenario: 异常插件提示人工修复
- **WHEN** 管理员打开插件管理页
- **AND** 插件 `plugin-demo` 的运行时状态为 `abnormal`
- **THEN** 管理页展示异常状态和修复说明
- **AND** 管理页不得展示普通升级确认动作

### Requirement: 插件运行时升级必须通过 REST API 显式执行

系统 SHALL 提供只读升级预览 API 和有副作用的升级执行 API。升级预览 MUST 使用 `GET`，升级执行 MUST 使用 `POST /plugins/{id}/upgrade`，且执行请求必须经过权限校验、确认校验和服务端状态重读。

#### Scenario: 获取插件升级预览
- **WHEN** 管理页请求 `GET /plugins/{id}/upgrade/preview`
- **AND** 插件处于 `pending_upgrade`
- **THEN** 系统返回升级前后版本、manifest 差异摘要、依赖检查、升级 SQL 摘要、授权变化和风险提示
- **AND** 系统不修改数据库状态

#### Scenario: 执行插件升级
- **WHEN** 管理员请求 `POST /plugins/{id}/upgrade`
- **AND** 请求通过插件管理权限校验
- **AND** 插件当前状态仍为 `pending_upgrade`
- **THEN** 系统执行运行时升级编排
- **AND** 升级成功后更新有效版本和运行时状态

#### Scenario: 拒绝非待升级插件升级请求
- **WHEN** 管理员请求 `POST /plugins/{id}/upgrade`
- **AND** 插件当前状态不是 `pending_upgrade`
- **THEN** 系统拒绝该请求
- **AND** 响应包含稳定业务错误码和可本地化消息键

### Requirement: 运行时升级必须执行插件自定义升级回调

系统 SHALL 为源码插件提供可选升级回调接口。宿主触发插件升级时传入升级前 manifest 快照和目标 manifest 快照，使插件可以执行自定义数据迁移、状态清理和兼容性处理。

#### Scenario: 目标版本插件实现升级回调
- **WHEN** 插件 `plugin-demo` 从 `v0.1.0` 升级到 `v0.2.0`
- **AND** 目标版本源码插件实现升级回调
- **THEN** 宿主调用该升级回调
- **AND** 回调请求包含 `v0.1.0` 的 manifest 快照
- **AND** 回调请求包含 `v0.2.0` 的 manifest 快照

#### Scenario: 插件未实现升级回调
- **WHEN** 插件 `plugin-demo` 处于 `pending_upgrade`
- **AND** 目标版本插件未实现升级回调
- **THEN** 宿主跳过插件自定义升级步骤
- **AND** 宿主继续执行标准升级 SQL 和治理资源同步

#### Scenario: 升级回调失败
- **WHEN** 插件升级回调返回错误
- **THEN** 宿主停止后续升级步骤
- **AND** 插件运行时状态变为 `upgrade_failed`
- **AND** 系统记录失败阶段和错误详情

### Requirement: 生命周期前置回调必须替代旧 Guard/Can* 契约

系统 SHALL 提供统一生命周期回调模型，使源码插件和动态插件均可在安装、升级、禁用、卸载、租户禁用、租户删除和安装模式切换等操作执行前返回允许或阻断决定。相同生命周期能力在源码插件和动态插件中 MUST 使用同一组 `Before*` 操作名称，例如 `BeforeInstall`、`BeforeUpgrade`、`BeforeDisable`、`BeforeUninstall`、`BeforeTenantDisable`、`BeforeTenantDelete` 和 `BeforeInstallModeChange`；不得为同一能力额外引入 `Can*`、guard、pre-* 等并行命名。系统 MUST 删除旧 Lifecycle Guard 与 `Can*` 插件契约，不得同时保留旧 Guard 注册、执行或兼容适配入口。

#### Scenario: 插件阻断升级
- **WHEN** 插件注册了 `BeforeUpgrade` 前置回调
- **AND** 管理员请求升级该插件
- **AND** 回调返回阻断决定和原因键
- **THEN** 宿主拒绝升级
- **AND** 响应包含该阻断原因
- **AND** 插件有效版本不得变化

#### Scenario: 插件阻断卸载
- **WHEN** 插件注册了 `BeforeUninstall` 前置回调
- **AND** 管理员请求卸载该插件
- **AND** 回调返回阻断决定和原因键
- **THEN** 宿主拒绝卸载
- **AND** 响应包含 `PLUGIN_LIFECYCLE_PRECONDITION_VETOED`
- **AND** 插件安装状态不得变化

#### Scenario: 动态插件阻断安装
- **WHEN** 动态插件 artifact 声明了 `BeforeInstall` 生命周期前置处理器
- **AND** 管理员请求安装该动态插件
- **AND** 动态插件处理器返回阻断决定和原因键
- **THEN** 宿主拒绝安装
- **AND** 响应包含 `PLUGIN_LIFECYCLE_PRECONDITION_VETOED`
- **AND** 宿主不得执行该插件 install SQL、治理资源同步或安装状态写入

#### Scenario: 动态插件阻断升级
- **WHEN** 动态插件 artifact 声明了 `BeforeUpgrade` 生命周期前置处理器
- **AND** 管理员请求升级该动态插件
- **AND** 动态插件处理器返回阻断决定和原因键
- **THEN** 宿主拒绝升级
- **AND** 响应包含该阻断原因
- **AND** 插件有效版本不得变化

#### Scenario: 事件 Hook 不承担前置阻断语义
- **WHEN** 动态插件监听 `plugin.installed`、`plugin.enabled`、`plugin.disabled`、`plugin.uninstalled` 或 `plugin.upgraded` 事件
- **THEN** 这些事件 Hook 仅用于生命周期完成后的事件通知或后续动作
- **AND** 系统不得把这些事件 Hook 当作 `Before*` 前置阻断机制使用

### Requirement: 插件升级必须保证缓存和集群一致性

系统 SHALL 在插件运行时升级成功或失败后按插件 ID 和资源作用域失效运行时缓存，并在集群模式下通过宿主统一的集群协调机制通知其他节点。`cluster.enabled=true` 时，不得仅依赖当前节点内存状态判断升级完成。

#### Scenario: 单机模式升级成功
- **WHEN** `cluster.enabled=false`
- **AND** 插件升级成功
- **THEN** 系统按插件 ID 失效本地插件状态、菜单、权限、路由、cron、i18n 和 apidoc 相关缓存
- **AND** 系统不强制依赖分布式协调组件

#### Scenario: 集群模式升级成功
- **WHEN** `cluster.enabled=true`
- **AND** 插件升级成功
- **THEN** 系统写入共享修订状态或发布集群事件
- **AND** 其他节点收到通知后刷新对应插件的运行时状态和缓存
- **AND** 同一插件不得在多个节点并发执行升级

### Requirement: 插件升级失败必须可诊断和可重试

系统 SHALL 在升级失败时保留失败状态、失败阶段、错误码、错误消息键、升级前 manifest 快照和目标 manifest 快照。管理员 MUST 能在插件管理页查看失败原因，并在修复后重试升级或执行人工修复。

#### Scenario: 升级 SQL 失败
- **WHEN** 插件升级执行 upgrade SQL 时失败
- **THEN** 系统将插件运行时状态标记为 `upgrade_failed`
- **AND** 系统记录失败阶段为 SQL 执行阶段
- **AND** 系统保留有效版本不切换到目标版本

#### Scenario: 管理员重试失败升级
- **WHEN** 插件状态为 `upgrade_failed`
- **AND** 管理员修复失败原因后重新发起升级
- **THEN** 系统重新执行升级预检
- **AND** 系统基于幂等迁移记录跳过已成功完成的迁移步骤或安全重试
