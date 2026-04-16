## ADDED Requirements

### Requirement: 定时任务分类
系统 SHALL 支持两种类型的定时任务：Master-Only 和 All-Node。

#### Scenario: 定义 Master-Only 任务
- **WHEN** 注册定时任务时指定为 Master-Only 类型
- **THEN** 该任务仅在主节点执行

#### Scenario: 定义 All-Node 任务
- **WHEN** 注册定时任务时指定为 All-Node 类型
- **THEN** 该任务在所有节点执行

### Requirement: Master-Only 任务主节点检查
系统 SHALL 在 Master-Only 任务执行前检查当前节点是否为主节点。

#### Scenario: 主节点执行 Master-Only 任务
- **WHEN** Master-Only 任务触发且当前节点是主节点
- **THEN** 任务正常执行

#### Scenario: 从节点跳过 Master-Only 任务
- **WHEN** Master-Only 任务触发且当前节点是从节点
- **THEN** 任务立即返回，不执行任何业务逻辑

### Requirement: 现有定时任务分类
系统 SHALL 对现有定时任务进行分类。

#### Scenario: Session Cleanup 分类为 Master-Only
- **WHEN** Session Cleanup 定时任务注册时
- **THEN** 系统将其标记为 Master-Only 类型

#### Scenario: Server Monitor Collector 分类为 All-Node
- **WHEN** Server Monitor Collector 定时任务注册时
- **THEN** 系统将其标记为 All-Node 类型

#### Scenario: Server Monitor Cleanup 分类为 Master-Only
- **WHEN** Server Monitor Cleanup 定时任务注册时
- **THEN** 系统将其标记为 Master-Only 类型
