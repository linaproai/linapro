## ADDED Requirements

### Requirement: 插件服务内部必须分离清单、存储和类型职责

系统 SHALL 在`apps/lina-core/internal/service/plugin/internal`内将插件基础类型、插件清单事实源和插件治理存储拆分为职责明确的内部子组件。`plugintypes` MUST 只承载插件 ID、状态、类型、scope、generation、版本和值对象等纯类型；`catalog` MUST 只承载插件清单扫描、解析、校验和访问；`store` MUST 作为插件治理表读写、发布/授权/治理投影和节点/发布状态持久化的 owner。`catalog`不得继续作为治理表读写、运行时状态同步或源码插件集成副作用的聚合点。

#### Scenario: 调用方只需要插件状态类型

- **WHEN** `runtime`、`integration`、`lifecycle`或测试代码只需要插件状态、类型、ID、scope、generation 或版本值对象
- **THEN** 调用方导入`plugintypes`
- **AND** 调用方不得为了这些纯类型依赖`catalog`

#### Scenario: 调用方需要治理表投影

- **WHEN** 插件列表、生命周期、升级或运行时路径需要读取`sys_plugin`、`sys_plugin_release`、插件授权快照、治理投影或节点状态
- **THEN** 调用方通过`store`获取稳定投影
- **AND** `store`不得向调用方返回`DAO`、`DO`、`Entity`、GoFrame 查询模型或私有缓存状态

#### Scenario: 调用方需要清单资源

- **WHEN** 调用方需要扫描、解析、校验或读取源码插件或动态插件 manifest 资源声明
- **THEN** 调用方通过`catalog`访问清单能力
- **AND** `catalog`不得在该读取过程中写入插件治理表或触发菜单、权限、资源引用、hook 分发副作用

### Requirement: 插件 catalog 不得持有 runtime 或 integration 反向回调

系统 SHALL 禁止`catalog`通过 setter、包级变量、全局 service locator 或构造后回注持有`runtime`、`integration`或插件门面能力。`catalog`不得声明或实现`Set*` wiring 方法，不得持有`menuSyncer`、`hookDispatcher`、`resourceRefSyncer`、`nodeStateSyncer`、`releaseStateSyncer`、`backendLoader`、`artifactParser`或`dynamicManifestLoader`等反向回调字段。需要特定扫描输入时，调用方 MUST 通过显式参数传入，或由职责明确的下层清单/资源读取能力直接处理。

#### Scenario: 构造插件服务

- **WHEN** `plugin.New()`构造`catalog`、`runtime`和`integration`
- **THEN** 构造过程不得调用`catalog.Set*`方法
- **AND** `catalog`实例完成构造后不得再通过 setter 接收`runtime`或`integration`能力

#### Scenario: catalog 执行 manifest 校验

- **WHEN** `catalog`校验动态插件 manifest 或 artifact 资源
- **THEN** 校验依赖必须来自`catalog`内部清单/资源读取能力或当前调用入口显式参数
- **AND** `catalog`不得通过长期持有的`runtime`回调读取 artifact、active release 或节点状态

#### Scenario: catalog 读取源码插件 backend 配置

- **WHEN** 源码插件清单扫描需要加载 backend hook、resource 或 route 声明
- **THEN** backend 装载能力必须由调用入口显式传入或下沉为清单扫描依赖
- **AND** `catalog`不得通过`SetBackendLoader`长期持有`integration`实例

### Requirement: 插件服务内部依赖方向必须由治理验证固化

系统 SHALL 为插件服务内部清单、存储和类型拆分提供静态边界验证。验证 MUST 随 Go 测试或等价治理命令执行，并在发现`plugintypes`依赖兄弟包、`catalog`依赖`runtime`/`integration`或`store`泄漏生成模型时失败。

#### Scenario: plugintypes 出现兄弟包依赖

- **WHEN** `plugintypes`非测试 Go 文件导入`catalog`、`store`、`runtime`、`integration`、`dao`、`do`或`entity`
- **THEN** 边界验证失败
- **AND** 错误信息指出违规文件和 import 路径

#### Scenario: catalog 重新依赖运行时或集成实现

- **WHEN** `catalog`非测试 Go 文件导入`runtime`或`integration`
- **THEN** 边界验证失败
- **AND** 调用方必须改用显式参数、`store`投影或编排层显式调用

#### Scenario: store 导出生成模型

- **WHEN** `store`公开 API、导出结构或接口方法暴露`dao`、`do`或`entity`类型
- **THEN** 边界验证或审查必须阻断
- **AND** `store`必须改为返回自有稳定投影
