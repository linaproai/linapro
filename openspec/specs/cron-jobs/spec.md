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

