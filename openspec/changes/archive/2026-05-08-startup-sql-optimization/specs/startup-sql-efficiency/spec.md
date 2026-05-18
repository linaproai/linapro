## ADDED Requirements

### Requirement: 默认启动不得输出 SQL 明细日志

系统 SHALL 将 SQL 明细日志视为显式诊断能力，而不是普通启动的默认输出。交付配置中的 `database.default.debug` 默认值 MUST 为 `false`。当该配置为 `false` 时，服务启动日志不得输出 ORM 逐条 SQL 明细；当调用方显式设置为 `true` 时，系统可以按 GoFrame 行为输出 SQL 明细。

#### Scenario: 默认配置启动不输出 SQL 明细

- **WHEN** 使用交付默认配置启动后端服务
- **THEN** `database.default.debug` 为 `false`
- **AND** 启动日志不得包含 `SHOW FULL COLUMNS`、`SELECT ... FROM`、`INSERT INTO` 等 ORM SQL 明细行

#### Scenario: 显式开启 SQL debug

- **WHEN** 管理员将 `database.default.debug` 设置为 `true`
- **THEN** 后端服务允许输出 ORM SQL 明细
- **AND** 该行为仅作为诊断模式，不作为默认开发或生产配置

### Requirement: 启动链路必须复用插件治理快照

系统 SHALL 在一次 HTTP 启动编排内复用同一组插件治理启动快照。`BootstrapAutoEnable`、源码插件 HTTP 路由注册、运行时前端包预热、插件只读投影和 cron 内置任务同步等启动阶段不得重复构造等价的 `sys_plugin`、`sys_plugin_release`、`sys_menu`、`sys_plugin_resource_ref` 全表快照。

#### Scenario: 同一次启动只构造一次插件 catalog 快照

- **WHEN** 后端执行一次 HTTP 启动编排
- **THEN** 插件 catalog 启动快照最多构造一次
- **AND** 后续插件启动阶段复用该快照读取 `sys_plugin` 与 `sys_plugin_release`

#### Scenario: 同一次启动只构造一次插件 integration 快照

- **WHEN** 后端执行一次 HTTP 启动编排
- **THEN** 插件 integration 启动快照最多构造一次
- **AND** 后续菜单和资源引用同步复用该快照读取 `sys_menu` 与 `sys_plugin_resource_ref`

#### Scenario: 启动写入同步更新快照

- **WHEN** 启动同步阶段创建或更新插件 registry、release、menu 或 resource ref 投影
- **THEN** 系统必须同步更新当前启动快照
- **AND** 后续启动阶段不得为了读取刚写入的投影再次全表扫描

### Requirement: 插件清单无差异同步不得产生数据库副作用

系统 SHALL 将插件清单同步实现为差异驱动。对于 registry、release snapshot、manifest menu、dynamic route permission menu 和 resource ref 均无差异的插件，同步过程不得开启事务、不得写入数据库、不得执行写后回读。

#### Scenario: 无差异源码插件同步不写库

- **WHEN** 源码插件 manifest 与数据库中 registry、release、菜单、权限和资源引用投影完全一致
- **THEN** 启动同步该插件时不得执行 `INSERT`、`UPDATE`、`DELETE`
- **AND** 不得产生空的 `BEGIN` / `COMMIT` 事务

#### Scenario: 菜单声明变更时才进入菜单事务

- **WHEN** 插件 manifest 的菜单或动态 route permission 发生变化
- **THEN** 系统可以开启菜单同步事务并写入必要变更
- **AND** 事务提交后必须更新启动快照中的菜单投影

#### Scenario: 发布快照变更时才同步 release metadata

- **WHEN** 插件 manifest 生成的 release snapshot 与当前 `sys_plugin_release` 行一致
- **THEN** 系统不得更新 release metadata
- **AND** 不得为了 refresh release snapshot 再次查询同一 release 行

### Requirement: 内置定时任务启动注册必须避免重复持久化扫描

系统 SHALL 将源码声明作为内置定时任务执行定义的权威来源。启动同步内置任务后，调度注册 MUST 使用声明派生的 `sys_job` projection snapshot；持久化调度器启动扫描不得再次加载同一批 `is_builtin=1` 任务作为执行定义。

#### Scenario: 内置任务按声明派生快照注册

- **WHEN** 后端启动并同步内置任务投影
- **THEN** 系统使用同步返回的 projection snapshot 注册内置任务
- **AND** 注册过程不需要按 ID 重新读取同一内置任务行

#### Scenario: 持久化扫描跳过内置任务

- **WHEN** 持久化调度器执行启动加载
- **THEN** 查询条件必须排除 `is_builtin=1` 的任务
- **AND** 只加载用户创建的启用任务或非内置插件任务

### Requirement: 启动必须输出阶段摘要而不是依赖 SQL 明细

系统 SHALL 在启动完成后输出启动阶段摘要日志。摘要至少包含插件扫描数量、插件同步变更数量、no-op 插件数量、启动快照构造次数、内置任务投影数量、启动阶段耗时。摘要不得包含完整 SQL 文本。

#### Scenario: 启动摘要包含插件同步统计

- **WHEN** 后端完成插件启动同步
- **THEN** 日志输出插件扫描数量、发生变更的插件数量和 no-op 插件数量
- **AND** 日志不得包含完整 SQL 语句文本

#### Scenario: 启动摘要包含快照构造次数

- **WHEN** 后端完成 HTTP 启动编排
- **THEN** 日志输出 catalog、integration 和 job 启动快照构造次数
- **AND** 同一启动阶段的重复快照构造会被测试或审查识别为回归

### Requirement: 启动 SQL 效率必须有自动化回归覆盖

系统 SHALL 提供自动化测试或 smoke 脚本覆盖启动 SQL 效率关键边界。测试不得依赖 GoFrame 元数据探测的精确 SQL 条数，但必须约束项目可控行为，包括默认不输出 SQL 明细、插件 no-op 同步无写入、无空事务、共享启动快照不重复构造。

#### Scenario: 默认启动日志 smoke

- **WHEN** 测试使用默认数据库 debug 配置启动后端服务
- **THEN** 测试断言启动日志不包含 ORM SQL 明细
- **AND** 测试断言启动摘要日志存在

#### Scenario: 插件 no-op 同步回归测试

- **WHEN** 测试准备一个已同步且无 manifest 差异的插件
- **THEN** 再次执行启动同步不得产生写入 SQL
- **AND** 不得产生空事务

#### Scenario: 启动快照复用回归测试

- **WHEN** 测试执行一次 HTTP 启动编排或等价启动编排单元
- **THEN** catalog、integration 和 job 启动快照构造次数必须分别保持在预算内
- **AND** 超出预算时测试失败
