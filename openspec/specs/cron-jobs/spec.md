# cron-jobs Specification

## Purpose
TBD - created by archiving change distributed-locker. Update Purpose after archive.
## Requirements
### Requirement: 定时任务分类
系统 SHALL 支持两种类型的定时任务：主节点专属任务和全节点任务。单节点模式下，主节点专属任务必须由当前节点直接执行；集群模式下，主节点专属任务仅在主节点执行。

#### Scenario: 定义主节点专属任务
- **WHEN** 注册定时任务时指定为主节点专属类型
- **THEN** 该任务在单节点模式下由当前节点执行
- **AND** 在集群模式下仅由主节点执行

#### Scenario: 定义全节点任务
- **WHEN** 注册定时任务时指定为全节点类型
- **THEN** 该任务在所有运行节点上执行

### Requirement: 主节点专属任务主节点检查
系统 SHALL 在执行主节点专属任务前根据部署模式判断当前节点是否应当执行。

#### Scenario: 集群模式主节点执行主节点专属任务
- **WHEN** `cluster.enabled=true` 且主节点专属任务触发时当前节点是主节点
- **THEN** 任务正常执行

#### Scenario: 集群模式从节点跳过主节点专属任务
- **WHEN** `cluster.enabled=true` 且主节点专属任务触发时当前节点是从节点
- **THEN** 任务立即返回
- **AND** 不执行任何业务逻辑

#### Scenario: 单节点模式执行主节点专属任务
- **WHEN** `cluster.enabled=false` 且主节点专属任务触发
- **THEN** 当前节点直接执行该任务

### Requirement: 现有定时任务分类
系统 SHALL 对现有定时任务按部署模式应用统一的调度规则。

#### Scenario: Session Cleanup 分类为主节点专属任务
- **WHEN** Session Cleanup 定时任务注册时
- **THEN** 系统将其标记为主节点专属任务
- **AND** 在单节点模式下由当前节点执行

#### Scenario: Server Monitor Collector 分类为全节点任务
- **WHEN** Server Monitor Collector 定时任务注册时
- **THEN** 系统将其标记为全节点任务
- **AND** 在所有节点执行

#### Scenario: Server Monitor Cleanup 分类为主节点专属任务
- **WHEN** Server Monitor Cleanup 定时任务注册时
- **THEN** 系统将其标记为主节点专属任务
- **AND** 在单节点模式下由当前节点执行

### Requirement: 用户可管理任务的调度范围一致性

系统 SHALL 将现有"主节点/全节点"调度语义统一应用到用户可管理的定时任务,不区分内置任务与用户任务。

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

### Requirement: 用户可管理任务的调度器注册

系统 SHALL 在启动和 CRUD 期间维护 gcron 注册表,使之与 `sys_job` 表中处于 `status=enabled` 的任务保持一致。

#### Scenario: 启动加载

- **WHEN** 宿主进程启动且 `service/cron` 启动
- **THEN** 系统 SHALL 扫描 `sys_job where status=enabled`
- **AND** 将每条任务按其 `scope / concurrency / timezone / cron_expr` 注册到 gcron

#### Scenario: CRUD 动态刷新

- **WHEN** 任务被创建、更新或删除
- **THEN** 系统 SHALL 原子地从 gcron 注销旧条目并重新注册新条目(如适用)
- **AND** 在刷新过程中加单任务互斥锁,避免与调度 tick 产生竞态

#### Scenario: 启用/禁用刷新

- **WHEN** 任务 `status` 从 `disabled` 变为 `enabled` 或反之
- **THEN** 系统 SHALL 在调度器中相应地注册或注销该任务

