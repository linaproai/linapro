## ADDED Requirements

### Requirement: 动态插件通过版本化宿主服务协议获取全部宿主能力

系统 SHALL 在保留`lina_env.host_call`入口的前提下，为动态插件提供版本化的宿主服务调用协议，并要求全部宿主能力通过统一宿主服务通道发布，而不是继续线性增加新的专用 opcode。

#### Scenario: Guest 调用结构化宿主服务

- **WHEN** guest SDK 发起一次宿主服务调用
- **THEN** 宿主通过统一请求 envelope 解析`service`、`method`、资源标识（如`storage.resources.paths`、URL 模式或`table`）和请求载荷
- **AND** 宿主服务注册表定位对应的服务处理器
- **AND** 宿主以统一响应 envelope 返回业务结果或结构化错误

#### Scenario: 未知 service 或 method 被拒绝

- **WHEN** 插件调用一个宿主不支持的`service`或`method`
- **THEN** 宿主返回显式的“不支持”错误
- **AND** 宿主不得进入任何实际宿主服务逻辑

#### Scenario: 已有最小 Host Call 被统一重构

- **WHEN** 宿主收敛当前动态插件能力模型
- **THEN** 已实现的日志、状态和数据访问能力也通过统一宿主服务协议对 guest 暴露
- **AND** 宿主不得继续维护面向插件的平行公开协议

### Requirement: 宿主服务访问同时受宿主服务声明推导的能力分类和资源授权约束

系统 SHALL 对每一次宿主服务调用同时执行粗粒度 capability 校验和细粒度资源授权校验，但 capability 集合必须由`hostServices`声明自动推导，而不是要求插件作者在`plugin.yaml`中重复维护第二份`capabilities`列表；任一层不满足都必须拒绝调用。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者在动态插件清单中声明`hostServices`
- **THEN** 构建器校验 service、method、资源声明（如低优先级服务的逻辑`resourceRef`、`storage.resources.paths`、URL 模式或`resources.tables`）和策略参数是否合法
- **AND** 宿主根据这些 methods 自动推导内部 capability 分类快照
- **AND** 将归一化后的宿主服务策略写入运行时产物
- **AND** 宿主装载产物后恢复为当前 release 的服务授权快照

#### Scenario: 缺少授权的宿主服务调用被拒绝

- **WHEN** 插件调用未声明的 service、method 或未授权的资源标识
- **THEN** 宿主返回显式拒绝错误
- **AND** 宿主不执行任何目标宿主服务逻辑

### Requirement: 资源型宿主服务声明属于权限申请而非自动授权

系统 SHALL 将所有资源型 hostServices 声明解释为权限申请清单，而不是插件在运行时自动拥有的资源访问权；其中`storage`当前使用`resources.paths`，`network`使用 URL 模式，`data`当前使用`resources.tables`，其余低优先级服务（`cache`、`lock`、`notify`）继续沿用逻辑`resourceRef`规划，并分别表示缓存命名空间、逻辑锁名和通知通道标识。

#### Scenario: 清单声明宿主服务资源申请

- **WHEN** 开发者在动态插件清单中声明`storage.resources.paths`、`network`的 URL 模式、`data.resources.tables`，或声明`cache`、`lock`、`notify`等低优先级服务的逻辑`resourceRef`
- **THEN** 这些声明表示插件申请对应宿主资源权限
- **AND** 声明本身不得直接视为运行时已授权结果

#### Scenario: 运行时只认宿主确认后的授权快照

- **WHEN** 动态插件在运行时调用一个资源型宿主服务
- **THEN** 宿主仅按安装或启用阶段确认后的授权快照执行校验
- **AND** 未被宿主确认的声明必须被拒绝
- **AND** 插件不能通过“先声明后直接调用”的方式跳过宿主授权确认

### Requirement: 宿主服务调用携带执行上下文并支持审计

系统 SHALL 为每一次宿主服务调用附带统一的执行上下文，并记录最小充分的调用审计信息。

#### Scenario: 请求型上下文调用宿主服务

- **WHEN** 动态插件在处理路由请求期间调用宿主服务
- **THEN** 宿主向服务处理器传入`pluginId`、执行来源、当前路由标识、当前用户身份快照和数据范围快照
- **AND** 需要用户上下文的服务方法可以复用这些治理信息做校验

#### Scenario: 系统型上下文调用宿主服务

- **WHEN** 动态插件在 Hook、定时任务或生命周期流程中调用宿主服务
- **THEN** 宿主向服务处理器传入无用户身份的系统型执行上下文
- **AND** 要求用户上下文的服务方法必须拒绝该调用

#### Scenario: 宿主记录宿主服务调用摘要

- **WHEN** 一次宿主服务调用结束
- **THEN** 宿主记录`pluginId`、service、method、资源标识摘要（如`path`、URL 模式、`resourceRef`或`table`）、结果状态和耗时
- **AND** 失败调用保留错误摘要用于诊断
- **AND** 宿主默认不记录敏感请求体和敏感响应体原文
