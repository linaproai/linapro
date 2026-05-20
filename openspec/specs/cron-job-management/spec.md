# 定时任务管理

## Purpose

定义定时任务管理行为，包括日志清理策略治理、内置任务投射、手动触发确认和 Shell 任务操作可用性。
## Requirements
### Requirement:日志清理策略

系统 SHALL 提供带任务级覆盖的全局默认清理策略，由内置系统定时任务定期执行。内置清理任务 SHALL 通过宿主源码注册并在服务启动时投射到 `sys_job`；交付 SQL 不得将 `host:cleanup-job-logs` 的初始化种子数据写入 `sys_job`。该内置任务的 `sys_job` 行是用于展示、日志和审计关联的治理投影，而非启动持久化加载使用的执行定义源。

#### Scenario:全局默认策略
- **当** 任务 `log_retention_override` 为空时
- **则** 任务日志按系统参数 `cron.log.retention` 清理

#### Scenario:任务级覆盖
- **当** 任务 `log_retention_override` 配置为 `{mode: days, value: 60}` 或 `{mode: count, value: 500}` 时
- **则** 系统按任务级策略清理该任务的日志，忽略全局默认

#### Scenario:无清理策略
- **当** 策略为 `mode=none` 时
- **则** 系统不清理该任务的日志

#### Scenario:内置清理任务
- **当** 宿主启动并同步内置任务时
- **则** 将 `host:cleanup-job-logs` 投射到 `sys_job`
- **且** 默认 `cron_expr` 每天午夜运行
- **且** 任务有 `is_builtin=1`
- **且** 交付 SQL 不将该任务种子到 `sys_job`
- **且** 清理任务运行时注册使用宿主代码定义而非后续的 `sys_job` 持久化扫描

### Requirement:手动任务触发必须要求确认

定时任务的"立即运行"操作 SHALL 在触发执行前显示确认弹窗，防止管理员误操作运行运维任务。内置定时任务在其当前代码拥有或插件拥有的处理器可用时 SHALL 保持可手动触发，但手动触发不得暗示管理员可以编辑存储在 `sys_job` 中的内置任务执行定义。

#### Scenario:触发操作要求确认
- **当** 管理员在定时任务列表中点击"立即运行"时
- **则** 前端显示确认弹窗
- **且** 管理员确认前不调用触发 API

#### Scenario:触发确认使用执行样式
- **当** "立即运行"的确认弹窗显示时
- **则** 复用现有确认组件模式
- **且** 确认操作使用执行特定的样式和措辞而非删除样式

#### Scenario:取消触发不做任何事
- **当** 管理员取消"立即运行"确认时
- **则** 前端不调用 `POST /job/{id}/trigger`
- **且** 任务列表状态保持不变

#### Scenario:Shell 编辑被阻断时触发仍可用
- **当** 因环境开关或缺少 Shell 权限无法创建或编辑 Shell 任务时
- **则** 可运行的 Shell 任务行仍显示可点击的"立即运行"操作
- **且** 点击后显示确认弹窗
- **且** 操作列仅显示一个编辑入口

#### Scenario:Shell 任务默认启用
- **当** 系统初始化 `cron.shell.enabled` 或因参数缺失回退时
- **则** 默认值为 `true`
- **且** 平台级不支持 Shell 的保护仍可安全禁用 Shell 任务

#### Scenario:允许手动触发内置任务
- **当** 管理员确认对处理器当前可用的 `is_builtin=1` 任务执行"立即运行"时
- **则** 后端 SHALL 通过当前宿主代码或插件处理器声明执行该任务
- **且** 后端 SHALL 创建关联到投射的 `sys_job.id` 的 `sys_job_log` 记录
- **且** 执行快照 SHALL 保留投射的任务元数据用于审计展示

#### Scenario:处理器不可用时阻断内置任务手动触发
- **当** 管理员确认对处理器不可用的 `is_builtin=1` 插件任务执行"立即运行"时
- **则** 后端 SHALL 以处理器不可用语义拒绝触发
- **且** 前端 SHALL 不为 `paused_by_plugin` 行显示可点击的"立即运行"操作

### Requirement:内置任务仅投射执行边界

系统 SHALL 将 `sys_job.is_builtin=1` 行视为框架拥有或插件拥有的定时任务的治理投影。这些行 SHALL 支持列表展示、详情展示、本地化投射、日志关联和手动触发查找，但管理员不得通过定时任务管理 API 或前端表单修改其执行定义。启动同步内置任务时，系统 MUST 使用源码声明生成的投影快照注册运行时调度任务；持久化启动扫描不得再次加载同一批 `is_builtin=1` 行作为执行定义来源。

#### Scenario:内置任务投射从声明同步
- **当** 宿主启动或插件生命周期同步运行时
- **则** 系统 SHALL 从宿主代码定义和插件 cron 声明 upsert 内置任务投射行
- **且** 每个投射行 SHALL 保持 `is_builtin=1`
- **且** 交付 SQL 不得作为内置任务种子行的来源
- **且** 运行时调度注册 SHALL 使用本次声明同步返回的投影快照

#### Scenario:拒绝内置任务定义变更
- **当** 调用方尝试通过定时任务管理 API 编辑、删除、启用、禁用、重置或以其他方式变更 `is_builtin=1` 任务的执行定义字段时
- **则** 后端 SHALL 以稳定业务错误拒绝请求
- **且** 前端 SHALL 对内置行隐藏或禁用这些变更操作

#### Scenario:内置任务投射保持可见
- **当** 管理员打开定时任务管理时
- **则** 内置行 SHALL 以本地化显示元数据保持可见
- **且** 其来源 SHALL 指明是宿主内置还是插件内置

#### Scenario:持久化启动扫描跳过内置任务
- **当** 持久化调度器在服务启动时加载任务
- **则** 查询和注册流程 MUST 排除 `is_builtin=1` 的任务
- **且** 内置任务不得因为持久化扫描而重复注册或重复读取

### Requirement:定时任务默认时区必须可配置

系统 SHALL 从配置键 `scheduler.defaultTimezone` 读取内置定时任务的默认时区，默认为 `UTC`。源码不得保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 等硬编码常量。

#### Scenario:缺失配置使用 UTC

- **当** 配置文件未声明 `scheduler.defaultTimezone`
- **且** 服务启动并注册内置任务时
- **则** 内置任务必须使用 `UTC` 作为默认时区

#### Scenario:自定义时区生效

- **当** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"`
- **且** 服务启动并注册内置任务时
- **则** 内置任务必须使用 `Asia/Shanghai` 作为默认时区

### Requirement:sys_job 表不得使用外键约束

系统 SHALL 从 `sys_job` 表移除 `fk_sys_job_group_id` 外键约束，在应用层维护 `group_id` 到 `sys_job_group` 的引用一致性，并在 `sys_job` 上保留 `KEY idx_group_id (group_id)` 用于基于分组的查询和清理路径。本仓库的其他关联表依赖应用层一致性，该表必须遵循该约定以避免高并发调度场景下的额外外键锁开销。

#### Scenario:sys_job 表不再包含外键约束

- **当** `make init` 完成数据库初始化时
- **则** `sys_job` 表不得包含 `fk_sys_job_group_id` 或任何指向 `sys_job_group` 的 `FOREIGN KEY` 约束

#### Scenario:sys_job 保留 group_id 索引

- **当** `make init` 完成数据库初始化时
- **则** `SHOW INDEX FROM sys_job` 必须包含列 `group_id` 上的 `idx_group_id`

#### Scenario:写路径验证 group_id 引用一致性

- **当** 上层调用方创建或更新带 `group_id` 的 `sys_job` 记录时
- **则** 服务层必须验证引用的分组存在
- **且** 验证失败必须返回 `bizerr` 业务错误而非依赖数据库外键拦截

### Requirement:用户创建定时任务必须遵循宿主数据权限

系统 SHALL 对 `sys_job.is_builtin=0` 的用户创建定时任务应用宿主数据权限治理。全部数据范围可访问所有用户创建任务；本部门数据范围仅访问创建人属于当前用户可见部门范围内的用户创建任务；仅本人数据范围仅访问 `created_by` 等于当前用户 ID 的用户创建任务。`sys_job.is_builtin=1` 的内置任务投影属于系统治理数据， SHALL 保持可见性、编辑保护和触发边界由现有内置任务规则控制，不按角色数据权限过滤。

#### Scenario:全部数据范围查询用户创建任务

- **当** 调用方拥有全部数据范围并查询定时任务列表时
- **则** 系统返回所有符合查询条件的用户创建任务
- **且** 内置任务投影按现有内置任务展示规则处理

#### Scenario:本部门范围查询用户创建任务

- **当** 普通用户角色数据范围为本部门数据
- **且** 查询定时任务列表时
- **则** 系统仅返回创建人属于当前用户可见部门范围内的用户创建任务
- **且** 不返回其他部门用户创建的任务

#### Scenario:仅本人范围查询用户创建任务

- **当** 普通用户角色数据范围为仅本人数据
- **且** 查询定时任务列表时
- **则** 系统仅返回 `created_by` 等于当前用户 ID 的用户创建任务

#### Scenario:内置任务不按数据权限过滤

- **当** 定时任务列表包含 `is_builtin=1` 的宿主或插件内置任务投影时
- **则** 系统不因当前用户角色数据范围为仅本人或本部门而过滤这些内置任务投影
- **且** 内置任务仍按功能权限、处理器可用性和内置任务保护规则控制可操作性

### Requirement:定时任务详情和变更必须校验数据权限

系统 SHALL 在读取、更新、删除、启用、禁用、重置或触发用户创建定时任务前校验目标任务是否位于当前调用方的数据权限范围内。目标任务不在范围内时，系统 SHALL 返回结构化数据不可见错误，且不得修改任务、调度器注册状态或创建执行日志。

#### Scenario:拒绝查看范围外任务详情

- **当** 普通用户角色数据范围为仅本人数据
- **且** 请求查看其他用户创建的任务详情
- **则** 系统返回结构化数据不可见错误
- **且** 不返回任务详情

#### Scenario:拒绝编辑范围外任务

- **当** 普通用户角色数据范围为本部门数据
- **且** 目标任务创建人不属于当前用户可见部门范围
- **则** 系统拒绝编辑、删除、启用或禁用该任务
- **且** 不刷新该任务的调度器注册状态

#### Scenario:拒绝触发范围外任务

- **当** 普通用户角色数据范围为仅本人数据
- **且** 请求手动触发其他用户创建的任务
- **则** 系统拒绝触发
- **且** 不创建新的 `sys_job_log` 执行记录

### Requirement:定时任务日志必须继承任务数据权限边界

系统 SHALL 对用户创建任务的执行日志应用与其所属任务一致的数据权限边界。调用方只能查询、查看、导出或终止其数据权限范围内任务产生的日志。内置任务日志继续按内置任务治理和功能权限控制。

#### Scenario:本部门范围查询任务日志

- **当** 普通用户角色数据范围为本部门数据
- **且** 查询定时任务日志列表
- **则** 系统仅返回其可见部门范围内用户创建任务产生的日志

#### Scenario:拒绝终止范围外任务日志

- **当** 普通用户角色数据范围为仅本人数据
- **且** 请求终止其他用户创建任务的 running 日志
- **则** 系统拒绝终止
- **且** 目标任务执行状态不因该请求改变

### Requirement: Master-Only 定时任务必须基于 Redis primary 状态
系统 SHALL 在集群模式下通过 Redis leader election 的 primary 状态决定 Master-Only 定时任务是否执行。

#### Scenario: primary 执行 Master-Only 任务
- **WHEN** Master-Only 定时任务触发
- **AND** 当前节点持有 Redis leader lock
- **THEN** 当前节点执行任务

#### Scenario: follower 跳过 Master-Only 任务
- **WHEN** Master-Only 定时任务触发
- **AND** 当前节点未持有 Redis leader lock
- **THEN** 当前节点跳过任务
- **AND** 不产生业务副作用

### Requirement: coordination KV kvcache backend 不得注册 SQL 过期清理任务
系统 SHALL 根据 kvcache backend 判断是否注册 KV 过期清理任务。coordination KV backend MUST 不注册 SQL table cleanup job。

#### Scenario: 集群 coordination KV backend 启动
- **WHEN** 宿主以集群模式启动并使用 coordination KV kvcache backend
- **THEN** 系统不投射 `host:kvcache-cleanup-expired` 任务
- **AND** Redis TTL 负责缓存过期

### Requirement: 集群 watcher 任务必须作为 revision 兜底
系统 SHALL 保留权限拓扑、运行时配置和插件运行时同步任务作为 Redis event 的兜底机制。watcher MUST 读取 Redis revision 而不是 PostgreSQL revision 表。

#### Scenario: watcher 发现 revision 前进
- **WHEN** 节点错过 Redis event
- **AND** watcher 读取到 Redis revision 高于本地 observed revision
- **THEN** 节点刷新本地派生缓存
- **AND** 更新 observed revision

### Requirement: 任务分组必须按租户作用域隔离

系统 SHALL 将 `sys_job_group` 作为租户作用域资源处理。任务分组列表、详情、创建、更新、删除、分组下任务计数和删除迁移 MUST 使用当前租户上下文限制到对应 `tenant_id`。租户上下文不得读取、更新、删除或迁移其他租户或平台上下文的任务分组。

#### Scenario: 租户查询只返回本租户任务分组

- **WHEN** 租户 A 用户查询任务分组列表
- **THEN** 系统只返回 `tenant_id = A` 的任务分组
- **AND** 不返回平台 `tenant_id = 0` 或租户 B 的任务分组

#### Scenario: 租户创建任务分组写入当前租户

- **WHEN** 租户 A 用户创建任务分组
- **THEN** 新记录的 `tenant_id` MUST 等于租户 A
- **AND** 分组编码唯一性按 `(tenant_id, code)` 校验
- **AND** 不得依赖数据库默认值写入 `tenant_id = 0`

#### Scenario: 租户更新范围外任务分组被拒绝

- **WHEN** 租户 A 用户请求更新租户 B 或平台的任务分组 ID
- **THEN** 系统 MUST 返回结构化不可见或权限错误
- **AND** 不修改目标任务分组

#### Scenario: 删除任务分组只迁移当前租户任务

- **WHEN** 租户 A 用户删除本租户任务分组
- **THEN** 系统仅将租户 A 中属于该分组的任务迁移到租户 A 的默认分组
- **AND** 不修改租户 B 或平台任务的 `group_id`
- **AND** 删除和迁移在同一租户边界内原子完成

### Requirement: 任务分组数据权限和缓存边界必须明确

任务分组读写 SHALL 与定时任务管理的数据权限治理兼容。读取类接口在数据库查询阶段注入租户过滤；详情和写操作在操作前校验目标分组可见性。任务分组本身不应引入独立跨节点缓存；若后续添加缓存，必须按 `tenant_id` 分桶并使用显式作用域失效。

#### Scenario: 任务分组计数不泄露其他租户任务

- **WHEN** 租户 A 查询任务分组列表并返回每个分组的任务数量
- **THEN** 每个计数只统计租户 A 可见任务
- **AND** 不因其他租户存在同名或同 ID 分组而改变计数

