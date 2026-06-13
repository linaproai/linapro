## ADDED Requirements

### Requirement: 插件服务内部构造必须无 setter 回注

系统 SHALL 在插件服务生产构造路径中通过构造函数逐项显式传入运行期依赖。`plugin.New()`、`runtime.New()`、`integration.New()`、`lifecycle.New()`、`sourceupgrade.New()`和其他插件内部 service 构造完成后 MUST 处于可用状态，不得依赖构造后`Set*`方法、运行期 service locator、包级默认实例或`ValidateRequiredDependencies`兜底校验完成生产接线。

#### Scenario: 构造插件根服务

- **WHEN** 宿主启动路径调用`plugin.New()`
- **THEN** `plugin.New()`按依赖拓扑构造内部 service
- **AND** 构造过程不调用`runtimeSvc.Set*`、`integrationSvc.Set*`、`lifecycleSvc.Set*`或等价 wiring setter
- **AND** 构造完成后不需要调用`ValidateRequiredDependencies`

#### Scenario: 新增运行期依赖

- **WHEN** `runtime`、`integration`、`lifecycle`或`sourceupgrade`新增运行期依赖
- **THEN** 该依赖作为构造函数中的独立接口参数出现
- **AND** 不得通过`Dependencies`、`Deps`、`Options`或包级变量聚合传入
- **AND** 编译错误必须暴露所有未同步的生产和测试构造点

### Requirement: 插件服务内部组件不得互持宽 service 形成循环

系统 SHALL 使用窄契约表达插件服务内部组件之间的运行期调用。`runtime`不得为了菜单同步、资源引用同步、hook 分发、权限菜单过滤或依赖校验持有完整`integration.Service`或插件根`Service`；`integration`不得为了动态 job 执行持有完整`runtime.Service`；`sourceupgrade`不得持有插件根门面或要求门面实现其内部接口。

#### Scenario: runtime 触发集成副作用

- **WHEN** runtime reconciliation 或动态路由路径需要菜单、资源引用、hook 或权限菜单过滤能力
- **THEN** runtime 只依赖对应窄接口
- **AND** 这些窄接口由组合根显式注入
- **AND** runtime 不保存`integration.Service`完整实例

#### Scenario: source upgrade 执行升级

- **WHEN** source upgrade 需要读取清单、治理投影、生命周期动作、runtime 状态或缓存失效
- **THEN** `sourceupgrade`构造函数只接收这些能力的窄接口
- **AND** 不接收插件根`Service`
- **AND** 插件根包不再声明实现`sourceupgrade`内部`Service`接口

### Requirement: 插件服务包级可变运行期状态必须清零

系统 SHALL 禁止插件服务生产代码通过包级变量保存运行期可变状态。生命周期 observer、integration route binding/enablement snapshot、runtime cache observed revision、runtime reconciler once/mutex、WASM host service runtime 和类似状态 MUST 归属于 service 实例、启动期共享对象或缓存协调后端。

#### Scenario: integration 多实例共享状态

- **WHEN** 同一进程中创建多个插件 service 实例且 route guard 需要读取插件启用状态
- **THEN** 组合根显式创建共享状态对象并注入`integration`
- **AND** `integration`包内不存在`defaultSharedState`或等价包级可变状态
- **AND** 测试 fixture 使用独立共享状态避免测试间串扰

#### Scenario: runtime reconciler 控制并发

- **WHEN** runtime 启动 background reconciler 或执行一次 reconciliation
- **THEN** once/mutex 状态保存在 runtime service 实例或显式共享对象上
- **AND** `runtime`包生产代码不得声明包级`sync.Once`或`sync.Mutex`来控制运行期流程

#### Scenario: lifecycle observer 注册

- **WHEN** 测试或运行期需要注册插件生命周期 observer
- **THEN** observer 表归属于插件 service 实例或显式测试对象
- **AND** 插件根包不得用包级 map 保存 observer

### Requirement: capabilityhost 领域适配器必须保留在同一组件边界内

系统 SHALL 将`capabilityhost`的领域适配器实现保留在`capabilityhost`组件边界内。单文件`capabilityhost/internal/*cap`微包不得继续作为长期结构；迁移后每个领域适配器可以使用`capabilityhost_<domain>.go`等职责文件组织，但不得引入新的无稳定契约微包。

#### Scenario: 构造 capabilityhost

- **WHEN** 宿主创建 capability host service directory
- **THEN** 构造函数继续逐项接收接口型运行期依赖
- **AND** 不使用`Deps`、`Options`或聚合结构体隐藏接口依赖
- **AND** 领域 adapter 不需要通过 import 别名跨多个内部微包拼装

### Requirement: 插件服务接线治理必须由静态测试固化

系统 SHALL 提供静态治理测试，阻断生产代码重新引入插件内部 wiring setter、包级可变运行期状态、旧插件 runtimecache import、WASM Configure 配置入口或 testutil 对生产接线流程的复刻。

#### Scenario: 生产代码重新引入 setter

- **WHEN** 插件服务生产 Go 文件声明`Set*` wiring 方法或`plugin.New()`调用内部 service setter
- **THEN** 静态治理测试失败
- **AND** 错误信息指出违规文件和方法名

#### Scenario: 生产代码重新引入包级可变状态

- **WHEN** 插件服务生产 Go 文件声明包级`sync.Once`、`sync.Mutex`、`atomic.Pointer`、map 或 slice 用于运行期状态
- **THEN** 静态治理测试失败
- **AND** 调用方必须改为 service 实例字段、构造期共享对象或缓存协调后端

#### Scenario: testutil 复刻生产接线

- **WHEN** `internal/testutil`需要创建插件服务相关 fixture
- **THEN** 它复用生产构造函数或测试专用组合根
- **AND** 不得重新维护一份与`plugin.New()`同构的 setter 接线流程
