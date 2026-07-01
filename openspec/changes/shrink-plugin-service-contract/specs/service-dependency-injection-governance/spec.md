## ADDED Requirements

### Requirement: 插件服务消费者必须依赖最小稳定契约

系统 SHALL 要求宿主内部消费者通过构造函数接收启动期共享插件服务实例发布的最小稳定契约。普通消费者只需要启动、运行时、job、状态、provider env 或租户生命周期等部分能力时，MUST 依赖消费者本地私有窄接口；若消费者是插件管理控制器、HTTP 启动上下文、`RuntimeDelegate`或其他必须跨启动阶段传递统一插件服务入口的边界，MUST 只引用导出的`plugin.Service`。消费者 MUST NOT 引用`plugin`包导出的 facet 类型，`plugin`包拆分出来的 facet MUST 保持私有。

#### Scenario: 定时任务组件依赖插件 job 能力

- **WHEN** `cron`或`jobhandler`需要同步插件声明的定时任务
- **THEN** 它们通过构造函数接收插件 job 和生命周期 observer 所需的最小契约
- **AND** 测试替身只需要实现这些实际使用的方法
- **AND** 该最小契约在消费者包内私有声明，不依赖`plugin`包导出的 facet 类型

#### Scenario: API 文档组件依赖插件路由能力

- **WHEN** `apidoc`需要读取源码插件 route binding 或投影动态路由
- **THEN** 它依赖插件路由和状态读取所需的窄契约
- **AND** 不得依赖插件安装、卸载、上传或租户 lifecycle 等无关方法
- **AND** 该窄契约在消费者包内私有声明，不依赖`plugin`包导出的 facet 类型

#### Scenario: Runtime delegate 依赖插件状态和 hook 能力

- **WHEN** 启动期 cycle breaker 在插件服务完成构造前被传入认证、菜单、角色或 capability provider
- **THEN** delegate 通过`plugin.Service`绑定启动期共享插件服务实例
- **AND** delegate 绑定后只调用 hook、菜单过滤、状态读取、provider env 和租户生命周期等当前需要的方法
- **AND** 不得为该单一 cycle breaker 额外保留`runtimeDelegateService`等重复窄接口

#### Scenario: 共享实例语义保持不变

- **WHEN** 消费者从完整`plugin.Service`迁移到 facet 或窄接口
- **THEN** 注入对象仍然来自启动期构造的同一个插件服务实例
- **AND** 不得创建新的插件服务实例、缓存快照、route binding 状态或 runtime frontend bundle 缓存

#### Scenario: HTTP 启动上下文保存插件服务

- **WHEN** HTTP 启动装配需要在多个启动阶段、路由绑定 helper 或 controller 构造中复用插件服务
- **THEN** `httpRuntime`只保存单个`pluginSvc pluginsvc.Service`字段作为统一入口
- **AND** 不得在`httpRuntime`中为管理、启动、运行时 HTTP、集成或 job facet 增加多个插件服务字段
- **AND** 路由、frontend asset 和 hook helper 从`httpRuntime`接收插件服务时应继续使用该统一`pluginsvc.Service`入口，不得额外声明插件 runtime HTTP 或 integration 分类接口

#### Scenario: 插件管理控制器使用统一插件服务入口

- **WHEN** 插件管理控制器通过`NewV1`接收插件服务依赖
- **THEN** 构造函数和控制器字段应直接使用`pluginsvc.Service`统一入口
- **AND** 不得在 controller 包中额外声明重复的插件 management 分类接口
- **AND** 单元测试替身可以嵌入`pluginsvc.Service`并只覆盖测试路径实际调用的方法
