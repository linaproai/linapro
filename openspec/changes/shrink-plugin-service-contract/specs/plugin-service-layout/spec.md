## ADDED Requirements

### Requirement: 插件根服务契约必须按真实消费边界收敛

系统 SHALL 要求`apps/lina-core/internal/service/plugin`根`Service`契约只组合由真实职责边界支撑的私有 facet 接口。每个 facet MUST 对应明确调用场景和职责边界，MUST 不为了视觉分类创建没有真实职责差异的接口分组。插件 service 包对外可访问的根接口 MUST 只保留`Service`，拆分出来的 facet 接口 MUST 使用私有定义，避免外部包直接依赖多个插件服务入口。

#### Scenario: 控制器依赖插件管理能力

- **WHEN** 插件管理控制器构造时需要访问插件列表、详情、安装、卸载、状态变更、升级、依赖检查、动态包上传或资源查询能力
- **THEN** 控制器依赖统一`Service`入口提供的插件管理方法
- **AND** 控制器不得重复声明插件 management 分类接口
- **AND** 插件管理 facet 不得作为导出类型被控制器引用

#### Scenario: 启动编排依赖插件启动和运行时能力

- **WHEN** HTTP 启动编排需要执行插件启动治理、动态路由绑定、frontend bundle 预热、OpenAPI 投影或 runtime reconciler 启动
- **THEN** 启动编排依赖统一`Service`入口提供的对应启动或运行时方法
- **AND** 启动上下文不得为管理、启动、运行时 HTTP、集成或 job facet 拆出多个插件服务字段
- **AND** 插件启动和运行时 facet 不得作为导出类型被启动编排引用

#### Scenario: 外部包引用插件服务类型

- **WHEN** `apps/lina-core/internal/service/plugin`包外代码需要保存或接收插件服务依赖
- **THEN** 它只能引用导出的`plugin.Service`或在消费者包内声明私有窄接口
- **AND** 不得引用`plugin`包导出的 facet 接口类型

#### Scenario: 无真实消费者的分类接口

- **WHEN** 某个插件服务接口分组只表达视觉分类而没有真实职责差异
- **THEN** 系统不得保留该分组
- **AND** 方法应留在已有真实 facet 中或保持根组合契约直接可见

### Requirement: 插件根服务重复方法必须合并或删除

系统 SHALL 删除插件根服务上仅包装统一入口、无生产入口或语义重复的方法。合并后的方法 MUST 使用清晰的命名类型或选项结构表达调用意图，不得引入反射式分发、字符串式通用动作或无边界的动态参数。

#### Scenario: 插件 job 查询统一入口

- **WHEN** 调用方需要读取插件声明的定时任务
- **THEN** 系统通过单一 job 查询契约表达可执行过滤、已安装过滤、插件 ID 过滤和 handler 返回需求
- **AND** 不再为可执行、单插件声明和已安装声明保留多个根服务方法

#### Scenario: 插件状态变更统一入口

- **WHEN** 调用方需要启用或禁用插件
- **THEN** 系统通过`UpdateStatus`状态变更契约传入目标状态和可选动态插件授权确认
- **AND** 根服务不得暴露`Enable`、`Disable`或`SetStatus`等价状态入口
- **AND** 公共 capability 接口可以继续使用`SetStatus`表达插件领域能力动作

#### Scenario: Hook 包装方法删除

- **WHEN** 调用方需要分发插件 hook 事件
- **THEN** 调用方使用统一 hook 事件分发契约并传入具体 extension point
- **AND** 根服务不得为每个 auth hook 事件保留额外包装方法
