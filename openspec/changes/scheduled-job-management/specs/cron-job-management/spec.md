## ADDED Requirements

### Requirement: 定时任务管理数据模型

系统 SHALL 提供用户可管理的定时任务数据模型,至少包含 `sys_job_group`、`sys_job`、`sys_job_log` 三张表,并满足以下约束。

#### Scenario: 任务分组平铺

- **WHEN** 管理员新增、修改或查询分组
- **THEN** 分组为平铺结构(不支持 parent 层级)
- **AND** 系统 SHALL 保证有且仅有一个 `is_default=1` 的默认分组,不可删除

#### Scenario: 分组删除时任务迁移

- **WHEN** 管理员删除一个非默认分组
- **THEN** 系统 SHALL 在删除前将该分组下的所有任务迁移到默认分组
- **AND** 并在 UI 弹框提示影响范围与不可逆性

#### Scenario: 任务名组内唯一

- **WHEN** 在同一分组内创建或修改任务名
- **THEN** 系统 SHALL 拒绝与该组下已存在任务同名
- **AND** 不同分组下允许任务同名

#### Scenario: 任务类型区分

- **WHEN** 创建任务指定 `task_type=handler`
- **THEN** 系统 SHALL 要求 `handler_ref` 非空、`params` 按 handler 的 JSON Schema 校验通过、`timeout_seconds ∈ [1, 86400]`
- **AND** `shell_cmd / work_dir / env` 忽略写入

#### Scenario: Shell 任务字段

- **WHEN** 创建任务指定 `task_type=shell`
- **THEN** 系统 SHALL 要求 `shell_cmd` 非空、`timeout_seconds ∈ [1, 86400]`
- **AND** `handler_ref / params` 忽略写入
- **AND** `work_dir` 可空,非空时必须是宿主进程有权访问的存在目录

### Requirement: 任务执行超时

系统 SHALL 为所有任务类型保存公共 `timeout_seconds` 字段,并由调度执行器统一生效。

#### Scenario: 创建或修改任务时校验超时

- **WHEN** 管理员创建或修改任意 `handler` 或 `shell` 任务
- **THEN** 系统 SHALL 要求 `timeout_seconds ∈ [1, 86400]`
- **AND** 缺失或越界时拒绝保存

#### Scenario: Handler 任务超时

- **WHEN** `task_type=handler` 的任务执行时间超过 `timeout_seconds`
- **THEN** 系统 SHALL 取消该次执行的 `context`
- **AND** 日志最终状态为 `timeout`
- **AND** `err_msg` 记录超时时长

### Requirement: 定时任务 CRUD 接口

系统 SHALL 以 RESTful 方式提供定时任务的增删改查能力,符合项目 HTTP 方法语义规范。

#### Scenario: 列表查询

- **WHEN** 调用 `GET /job?groupId=&status=&taskType=&keyword=&page=&pageSize=`
- **THEN** 系统 SHALL 返回匹配条件的任务分页列表
- **AND** 返回字段包含任务基础信息、分组名称、handler 展示名(handler 类型)、最近一次执行状态

#### Scenario: 创建

- **WHEN** 调用 `POST /job` 提交符合 schema 的任务
- **THEN** 系统 SHALL 持久化记录
- **AND** 若 `status=enabled`,则立即在本节点调度器注册
- **AND** `executed_count=0`、`stop_reason=null`

#### Scenario: 更新

- **WHEN** 调用 `PUT /job/{id}` 修改字段
- **THEN** 系统 SHALL 从调度器注销旧记录并按新配置重新注册
- **AND** 保留 `executed_count` 不变(除非显式调用重置接口)

#### Scenario: 新增与编辑时严格校验公共调度字段

- **WHEN** 管理员在新增或编辑任务时提交 `定时表达式`、`任务状态`、`调度范围`、`并发策略`、`最大并发`、`最大执行次数`、`超时时间`、`任务时区` 等公共字段
- **THEN** 前后端 SHALL 对字段类型、可选值、数值范围与必填约束执行严格校验
- **AND** `定时表达式` SHALL 仅支持 5 段或 6 段 Cron 文本格式,不接受空值、超长值或其他段数
- **AND** 5 段表达式 SHALL 由系统在运行时自动补 `#` 秒占位,用户手工提交的 6 段表达式秒位 SHALL 使用真实秒值而不是 `#`
- **AND** `任务状态` 的写入值 SHALL 仅允许 `enabled` 与 `disabled`,`paused_by_plugin` SHALL 仅由系统在插件处理器失效时自动写入
- **AND** 校验失败时 SHALL 返回明确错误信息,指出是表达式格式、时区、枚举值还是数值范围不合法

#### Scenario: 批量删除

- **WHEN** 调用 `DELETE /job` 传入 id 列表
- **THEN** 系统 SHALL 从调度器注销这些任务
- **AND** 对 `is_builtin=1` 的任务拒绝删除并返回明确错误

### Requirement: 定时任务导航与可解释文案

系统 SHALL 在管理工作台中以更易理解的导航结构和文案呈现定时任务能力,降低首次使用成本。

#### Scenario: 系统管理菜单分组

- **WHEN** 管理员查看系统管理菜单
- **THEN** 前端 SHALL 在 `系统管理` 下提供一个名为 `定时任务` 的目录菜单
- **AND** `任务管理`、`分组管理`、`执行日志` SHALL 作为该目录下的子菜单展示

#### Scenario: 目录入口默认落到任务管理

- **WHEN** 用户访问 `定时任务` 目录入口路由
- **THEN** 前端 SHALL 默认落到 `任务管理` 页面
- **AND** 已存在的 `/system/job`、`/system/job-group`、`/system/job-log` 页面路由 SHALL 保持兼容可访问

#### Scenario: 插件不可用状态解释

- **WHEN** 任务列表返回 `status=paused_by_plugin`
- **THEN** 前端 SHALL 使用明确的中文状态文案提示这是“插件处理器不可用”
- **AND** SHALL 通过 tooltip 解释该状态表示任务依赖的插件处理器当前未注册、已被禁用或已被卸载

#### Scenario: Cron 表达式帮助文案

- **WHEN** 用户在新增或编辑任务时查看 `定时表达式` 字段
- **THEN** 前端 SHALL 明确说明同时支持 5 段与 6 段 Cron 表达式
- **AND** SHALL 说明 5 段表达式会自动按 `# 秒占位` 补齐为 6 段再交由调度器执行

#### Scenario: 调度范围与并发策略中文展示

- **WHEN** 任务列表渲染 `scope` 与 `concurrency` 字段
- **THEN** 前端 SHALL 使用易理解的中文标签展示调度范围与并发策略
- **AND** 不直接展示 `master_only / all_node / singleton / parallel` 等英文原始值
- **AND** `all_node` 的中文标签 SHALL 展示为 `所有节点执行`

#### Scenario: 公共调度字段帮助提示

- **WHEN** 用户在新增或编辑任务时查看 `调度范围`、`并发策略`、`日志保留` 字段
- **THEN** 前端 SHALL 在字段标签旁提供帮助提示入口
- **AND** 帮助文案 SHALL 解释各选项的含义、差异和适用场景
- **AND** 过长说明 SHALL 通过显式换行分段展示,避免提示浮层出现难以阅读的长行文本

#### Scenario: 跟随系统的日志保留说明

- **WHEN** 用户查看 `日志保留` 字段中的“跟随系统”选项
- **THEN** 前端 SHALL 展示当前系统级日志保留策略的解释
- **AND** SHALL 说明该任务未设置覆盖策略时会跟随系统默认策略执行清理

#### Scenario: 定时表达式字段样式

- **WHEN** 用户在新增或编辑任务时输入定时表达式
- **THEN** 前端 SHALL 使用更接近代码输入框的单行样式呈现该字段
- **AND** 至少提供等宽字体、代码框视觉样式与更适合表达式编辑的输入体验
- **AND** 不要求在新增或编辑页对表达式片段进行分段语法高亮

#### Scenario: 任务列表中的定时表达式代码化展示

- **WHEN** 任务管理列表渲染 `cron_expr` 列
- **THEN** 前端 SHALL 使用内联代码风格展示定时表达式
- **AND** SHALL 通过分段高亮提升 `* / # / 数字` 等表达式片段的可读性
- **AND** 不改变接口返回的原始表达式文本内容

#### Scenario: 超时时间与最大执行次数帮助提示

- **WHEN** 用户在新增或编辑任务时查看 `超时时间(秒)` 与 `最大执行次数` 字段
- **THEN** 前端 SHALL 在字段标签旁提供帮助提示入口
- **AND** 帮助文案 SHALL 解释超时时间的作用、任务超时后的影响以及最大执行次数的限制规则
- **AND** `最大执行次数=0` SHALL 被明确说明为“不限制执行次数”

#### Scenario: 时区字段支持常用选项与自定义输入

- **WHEN** 用户在新增或编辑任务时配置时区
- **THEN** 前端 SHALL 提供可搜索的常用时区下拉选项
- **AND** SHALL 默认选中宿主当前系统时区
- **AND** SHALL 允许用户输入自定义时区字符串,例如 `Asia/Shanghai` 或 `UTC`

#### Scenario: Shell 任务警告提示与表单间距

- **WHEN** 用户在新增或编辑任务时切换到 `Shell` 任务配置页签
- **THEN** 前端 SHALL 继续展示“Shell 任务会在宿主节点直接执行，请严格控制命令内容和环境变量。”警告提示
- **AND** 该提示块与下方表单区域之间 SHALL 保持不少于 `5px` 的垂直间距

### Requirement: 定时任务启用与禁用

系统 SHALL 允许管理员切换任务启停状态,并保证调度器注册表与数据库状态一致。

#### Scenario: 启用任务

- **WHEN** 调用 `PUT /job/{id}/status` 将 `status` 置为 `enabled`
- **THEN** 系统 SHALL 将任务注册到调度器
- **AND** 重置 `stop_reason` 为 `null`

#### Scenario: 禁用任务

- **WHEN** `status` 置为 `disabled`
- **THEN** 系统 SHALL 从调度器注销任务
- **AND** 运行中的实例不会被强制终止,但不再接受新的 tick

#### Scenario: 插件不可用阻止启用

- **WHEN** 尝试启用 `handler_ref` 来源插件已禁用/卸载的任务
- **THEN** 系统 SHALL 拒绝启用操作并返回明确错误

### Requirement: 执行次数策略

系统 SHALL 支持"执行 N 次后退出"策略,并在达到上限时自动禁用任务。

#### Scenario: 默认无限执行

- **WHEN** 任务 `max_executions=0`
- **THEN** 系统 SHALL 无执行次数上限

#### Scenario: 达到上限自动禁用

- **WHEN** 任务 `max_executions=N` 且 `executed_count` 达到 N
- **THEN** 系统 SHALL 立即将该任务 `status` 置为 `disabled`
- **AND** 写入 `stop_reason=max_executions_reached`
- **AND** 从调度器注销
- **AND** 在对应执行日志上保留该次完整记录

#### Scenario: 重置执行计数

- **WHEN** 调用 `POST /job/{id}/reset`
- **THEN** 系统 SHALL 将 `executed_count` 归零
- **AND** 清空 `stop_reason`

#### Scenario: 手动触发不计入次数

- **WHEN** 通过手动触发接口启动一次执行
- **THEN** 系统 SHALL 不将本次执行计入 `executed_count`
- **AND** 日志中 `trigger=manual`

### Requirement: 集群调度范围

系统 SHALL 支持每任务独立配置 `scope ∈ {master_only, all_node}`,并在调度时按范围语义执行。

#### Scenario: master-only 任务在非主节点跳过

- **WHEN** `scope=master_only` 且当前节点 `cluster.IsPrimary()=false`
- **THEN** 系统 SHALL 跳过本次执行
- **AND** 写入日志 `status=skipped_not_primary`

#### Scenario: all-node 任务每节点独立执行

- **WHEN** `scope=all_node`
- **THEN** 每个运行节点 SHALL 各自执行一份
- **AND** 日志中 `node_id` 字段记录执行节点

### Requirement: 并发策略

系统 SHALL 支持 `concurrency ∈ {singleton, parallel}`,并对 parallel 强制 `max_concurrency` 上限。

#### Scenario: 单例执行跳过

- **WHEN** `concurrency=singleton` 且本节点已有实例在 running
- **THEN** 系统 SHALL 跳过本 tick
- **AND** 写入日志 `status=skipped_singleton`

#### Scenario: 并行超限跳过

- **WHEN** `concurrency=parallel` 且本节点 running 实例数 ≥ `max_concurrency`
- **THEN** 系统 SHALL 跳过本 tick
- **AND** 写入日志 `status=skipped_max_concurrency`

#### Scenario: 并行配额内触发

- **WHEN** `concurrency=parallel` 且 running 实例数 < `max_concurrency`
- **THEN** 系统 SHALL 启动新的执行实例,与已有实例并行

### Requirement: 时区处理

系统 SHALL 为每个任务保存独立 `timezone` 字段,调度器按该时区解析 `cron_expr`。

#### Scenario: 任务按声明时区触发

- **WHEN** 任务 `timezone=Asia/Shanghai` 且 `cron_expr` 表达每天 8 点
- **THEN** 系统 SHALL 在 UTC+8 的每日 8 点触发该任务
- **AND** 与服务器本地时区无关

#### Scenario: 时区默认值

- **WHEN** 创建任务时未指定 `timezone`
- **THEN** 系统 SHALL 默认填入服务器进程时区
- **AND** UI 表单默认选中服务器时区

### Requirement: 手动触发与手动终止

系统 SHALL 提供立即触发一次执行与终止正在执行实例的能力。

#### Scenario: 立即触发

- **WHEN** 调用 `POST /job/{id}/trigger`
- **THEN** 系统 SHALL 跳过调度时钟直接启动一次执行
- **AND** 日志 `trigger=manual`
- **AND** 该次执行不计入 `executed_count`
- **AND** 本次执行仍受 `concurrency / max_concurrency / scope / timeout` 约束

#### Scenario: 停用任务也可手动触发

- **WHEN** 任务当前 `status=disabled`
- **THEN** 管理员仍 SHALL 能够通过 `立即执行` 手动触发该任务
- **AND** 该操作不自动将任务状态切换为 `enabled`
- **AND** 若任务处于 `paused_by_plugin` 或运行前校验失败,系统 SHALL 继续拒绝触发并返回明确原因

#### Scenario: 终止运行中实例

- **WHEN** 调用 `POST /job/log/{logId}/cancel` 且目标日志 `status=running`
- **THEN** 系统 SHALL 向本次执行的 context 发送取消信号
- **AND** 对 shell 任务 kill 其进程组
- **AND** 日志最终状态置为 `cancelled`
- **AND** `end_at` 写入取消时间

#### Scenario: 终止已结束实例拒绝

- **WHEN** 目标日志 `status` 已不是 `running`
- **THEN** 系统 SHALL 返回明确错误

### Requirement: 执行日志

系统 SHALL 为每次触发(包括被跳过的触发)记录一条执行日志,记录执行快照与结果。

#### Scenario: 清空全部日志

- **WHEN** 管理员在执行日志页执行“清空全部日志”
- **THEN** 后端 SHALL 在无 `jobId` 条件时安全删除全部执行日志
- **AND** 不因 ORM 缺少 `WHERE` 条件保护而报错

#### Scenario: 批量删除选中日志

- **WHEN** 管理员在执行日志列表中勾选多条日志并执行批量删除
- **THEN** 前端 SHALL 提供与监控管理-操作日志一致的批量删除入口
- **AND** 后端 SHALL 仅删除选中的执行日志记录
- **AND** 成功后列表与勾选状态 SHALL 同步刷新

#### Scenario: 日志字段

- **WHEN** 一次执行产生日志
- **THEN** 日志 SHALL 包含 `job_id / job_snapshot / node_id / trigger / params_snapshot / start_at / end_at / duration_ms / status / err_msg / result_json`

#### Scenario: Shell 任务输出

- **WHEN** 任务 `task_type=shell` 执行结束
- **THEN** 日志 `result_json` SHALL 包含 `stdout / stderr / exit_code`
- **AND** stdout 与 stderr 各自截留前 64KB,超出追加 `...[truncated]` 标记

#### Scenario: Handler 任务返回

- **WHEN** 任务 `task_type=handler` 执行成功并返回结构化结果
- **THEN** 日志 `result_json` SHALL 序列化 handler 返回值

#### Scenario: 执行失败

- **WHEN** 任务执行抛出错误
- **THEN** 日志 `status=failed`
- **AND** `err_msg` 记录错误摘要

### Requirement: 日志清理策略

系统 SHALL 提供全局默认清理策略与任务级覆盖,由系统内置定时任务定期执行清理。

#### Scenario: 全局默认策略

- **WHEN** 任务 `log_retention_override` 为空
- **THEN** 该任务日志 SHALL 按系统参数 `cron.log.retention` 的策略清理

#### Scenario: 任务级覆盖

- **WHEN** 任务 `log_retention_override` 配置为 `{mode: days, value: 60}` 或 `{mode: count, value: 500}`
- **THEN** 系统 SHALL 按任务级策略清理该任务日志,忽略全局默认

#### Scenario: 不清理策略

- **WHEN** 策略 `mode=none`
- **THEN** 系统 SHALL 不清理该任务对应日志

#### Scenario: 内置清理任务

- **WHEN** 系统初始化 seed
- **THEN** 系统 SHALL 内置一个 `host:cleanup-job-logs` handler 任务
- **AND** 默认 `cron_expr` 为每日凌晨触发
- **AND** 该任务 `is_builtin=1`

### Requirement: 系统内置任务的部分只读

系统 SHALL 保护 `is_builtin=1` 的任务关键字段不被修改,但开放运维调整运行参数。

#### Scenario: 可修改字段

- **WHEN** 管理员修改 `is_builtin=1` 任务的 `cron_expr / timezone / status / timeout_seconds / max_executions / log_retention_override`
- **THEN** 系统 SHALL 接受修改并应用

#### Scenario: 锁定字段

- **WHEN** 请求修改 `task_type / handler_ref / params / scope / concurrency / group_id / name` 任一字段
- **THEN** 系统 SHALL 拒绝修改并返回明确错误

#### Scenario: 编辑页锁定提示排版

- **WHEN** 管理员打开系统内置任务的编辑弹窗
- **THEN** 前端 SHALL 为“公共调度字段锁定说明”和“处理器引用与参数锁定说明”提示块保留清晰的上下留白
- **AND** 提示块与相邻表单区域之间的上下间隔 SHALL 不小于 `5px`

#### Scenario: 禁止删除

- **WHEN** 请求删除 `is_builtin=1` 任务
- **THEN** 系统 SHALL 拒绝操作

#### Scenario: 升级不覆盖用户修改

- **WHEN** 新版本 seed SQL 再次执行且任务已存在
- **THEN** 系统 SHALL 仅更新锁定字段(当 `seed_version` 较新时)
- **AND** 不覆盖用户已修改的开放字段

### Requirement: 权限与审计

系统 SHALL 通过菜单权限与按钮权限约束任务管理操作,并对敏感操作记录审计日志。

#### Scenario: 菜单与按钮权限

- **WHEN** 管理员访问任务管理相关页面或操作
- **THEN** 系统 SHALL 校验以下权限:菜单 `system:job:list / system:jobgroup:list / system:joblog:list`
- **AND** 任务按钮权限为 `system:job:add / edit / remove / status / trigger / reset`
- **AND** 分组按钮权限为 `system:jobgroup:add / edit / remove`
- **AND** 日志按钮权限为 `system:joblog:remove / cancel`

#### Scenario: Shell 组合权限

- **WHEN** 用户创建、修改或手动触发 `task_type=shell` 的任务
- **THEN** 系统 SHALL 同时校验对应的基础任务权限(`system:job:add / edit / trigger`)与附加权限 `system:job:shell`

#### Scenario: Shell 终止组合权限

- **WHEN** 用户终止一个属于 `task_type=shell` 的 running 实例
- **THEN** 系统 SHALL 同时校验 `system:joblog:cancel` 与附加权限 `system:job:shell`

#### Scenario: 操作审计

- **WHEN** 执行任务创建/修改/删除/启停/手动触发/手动终止
- **THEN** 系统 SHALL 写入 `oper_log`,记录操作人、操作类型、任务 ID 与关键字段快照
