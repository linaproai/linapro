## MODIFIED Requirements

### Requirement: 用户可管理任务的调度范围一致性

系统 SHALL 将现有"主节点/全节点"调度语义统一应用到用户可管理的定时任务；用户可管理任务指 `sys_job.is_builtin=0` 的任务。内置任务也 MUST 遵循相同的 scope 执行规则，但其执行定义来源于宿主代码或插件声明，而不是由 `sys_job` 记录驱动。

#### Scenario: 用户创建 master-only 任务

- **WHEN** 用户创建 `scope=master_only` 的定时任务
- **THEN** 该任务 SHALL 在集群模式下仅由主节点执行
- **AND** 在单节点模式下由当前节点直接执行

#### Scenario: 用户创建 all-node 任务

- **WHEN** 用户创建 `scope=all_node` 的定时任务
- **THEN** 该任务 SHALL 在所有运行节点上各自执行一份
- **AND** 每次执行的 `sys_job_log.node_id` 记录触发节点

#### Scenario: 非主节点 master-only 跳过记录

- **WHEN** `cluster.enabled=true` 且 `scope=master_only` 任务在非主节点触发
- **THEN** 系统 SHALL 立即返回且不执行业务逻辑
- **AND** 写入一条 `sys_job_log` 记录,`status=skipped_not_primary`

#### Scenario: 内置任务遵循统一 scope 规则

- **WHEN** 宿主代码或插件声明的内置任务被注册到调度器
- **THEN** 系统 SHALL 按声明中的 `scope` 应用 master-only 或 all-node 执行规则
- **AND** `sys_job.is_builtin=1` 投影记录不得成为该内置任务的执行定义来源

### Requirement: 用户可管理任务的调度器注册

系统 SHALL 在启动和 CRUD 期间维护 gcron 注册表,使之与 `sys_job` 表中处于 `status=enabled` 且 `is_builtin=0` 的用户自定义任务保持一致。`sys_job.is_builtin=1` 的内置任务 MUST NOT 由持久化调度器启动扫描注册；内置任务 SHALL 由宿主代码或插件声明同步路径注册。

#### Scenario: 启动加载

- **WHEN** 宿主进程启动且 `service/cron` 启动
- **THEN** 系统 SHALL 扫描 `sys_job where status=enabled and is_builtin=0`
- **AND** 将每条用户自定义任务按其 `scope / concurrency / timezone / cron_expr` 注册到 gcron
- **AND** 不得通过该持久化扫描注册 `is_builtin=1` 的内置任务

#### Scenario: CRUD 动态刷新

- **WHEN** 用户自定义任务被创建、更新或删除
- **THEN** 系统 SHALL 原子地从 gcron 注销旧条目并重新注册新条目(如适用)
- **AND** 在刷新过程中加单任务互斥锁,避免与调度 tick 产生竞态

#### Scenario: 启用/禁用刷新

- **WHEN** 用户自定义任务 `status` 从 `disabled` 变为 `enabled` 或反之
- **THEN** 系统 SHALL 在调度器中相应地注册或注销该任务

#### Scenario: 内置任务不参与持久化加载

- **WHEN** `sys_job` 中存在 `is_builtin=1 and status=enabled` 的内置任务投影
- **AND** 宿主进程启动并执行持久化调度器加载
- **THEN** 持久化调度器 MUST NOT 将该记录作为注册来源
- **AND** 该内置任务只能由对应的宿主代码定义或插件 cron 声明注册
