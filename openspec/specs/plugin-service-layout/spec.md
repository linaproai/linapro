# plugin-service-layout Specification

## Purpose
TBD - created by archiving change refactor-plugin-service-layout. Update Purpose after archive.
## Requirements
### Requirement: 插件服务根包必须保持 facade 职责清晰

系统 SHALL 将`apps/lina-core/internal/service/plugin`根包作为插件宿主服务的稳定 facade，根包主文件只保留公共契约、核心类型、构造、轻量校验和必要 wiring。复杂实现流程 SHOULD 放在同包职责明确的非主文件或既有`internal/<subcomponent>`中；当移动会造成循环依赖或只新增透传抽象时，系统 MAY 保持同包实现，但必须通过文件名和测试组织体现职责。

#### Scenario: 根包新增复杂实现

- **WHEN** 开发者为插件根服务新增生命周期、runtime upgrade、列表投影、host service 或缓存相关复杂流程
- **THEN** 实现文件名必须体现具体职责
- **AND** 不得把具体流程直接塞回`plugin.go`主文件

#### Scenario: 子组件移动会造成无意义转发

- **WHEN** 某段逻辑移动到`internal/<subcomponent>`只会产生新增接口和参数透传
- **THEN** 系统可以保持同包实现
- **AND** 通过窄函数、清晰命名和关联测试控制复杂度

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

### Requirement: 插件服务测试必须按被测职责组织

系统 SHALL 让插件服务单元测试文件与被测源码或明确主题关联。大测试文件 SHOULD 按 lifecycle、runtime upgrade、management list、startup auto-enable、tenant governance、host service 和测试 fixture 等职责拆分。共享 helper MUST 放在职责明确的测试支撑文件或`internal/testutil`，并由当前测试显式调用。

#### Scenario: 根包测试 helper 被多个测试文件复制

- **WHEN** 同一测试 helper 被多个根包测试文件重复实现
- **THEN** helper 应收敛到根包`*_test.go`支撑文件或`internal/testutil`
- **AND** 不得为了复用 helper 扩大生产代码导出面

#### Scenario: 大测试文件包含多个无关主题

- **WHEN** 一个`*_test.go`文件同时覆盖多个不同源码职责
- **THEN** 测试应按职责拆分到关联文件
- **AND** 每个测试仍必须自行构造依赖、数据和清理逻辑

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

### Requirement: 插件生命周期编排必须归属 lifecycle 子组件

系统 SHALL 将插件安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子编排归属到`internal/service/plugin/internal/lifecycle`。`internal/service/plugin`根门面 MUST 只保留公共契约、平台治理守卫、输入轻量校验、必要锁/缓存入口协调和委托，不得直接承载迁移后的生命周期长状态机。

#### Scenario: 根门面执行生命周期写操作

- **WHEN** 调用方通过插件根服务执行 Install、Uninstall、Enable、Disable 或 UpdateStatus
- **THEN** 根门面先执行平台上下文和入口治理守卫
- **AND** 根门面将业务编排委托给 lifecycle 子组件
- **AND** 根门面不得直接访问插件治理 DAO 或执行 SQL migration 细节

#### Scenario: lifecycle 编排执行状态机

- **WHEN** lifecycle 子组件处理安装、卸载或状态变更
- **THEN** 它通过构造函数注入的 catalog、store、runtime、integration、migration、dependency、i18n 和 cache publisher 窄契约完成编排
- **AND** 不通过 package-level service locator、构造后 setter 或反向持有插件根门面完成调用

### Requirement: 插件 SQL migration executor 必须独立于 lifecycle 编排

系统 SHALL 将插件生命周期 SQL 文件执行、mock data 执行、uninstall SQL 执行和 migration ledger 维护归属到独立`internal/service/plugin/internal/migration`组件。lifecycle 编排 MUST 只通过窄接口调用 migration executor，不得把 SQL 文件执行细节和生命周期状态机混在同一组件主流程中。

#### Scenario: 安装流程执行 SQL

- **WHEN** lifecycle 编排需要执行插件 install SQL 或 mock-data SQL
- **THEN** lifecycle 调用 migration 组件的窄接口
- **AND** migration 组件保持事务、方言转译和账本一致性语义
- **AND** lifecycle 编排函数只表达安装步骤和错误处理语义

#### Scenario: 动态插件卸载执行 SQL

- **WHEN** lifecycle 编排需要执行动态插件 uninstall SQL
- **THEN** migration 组件负责 SQL asset 读取、方言转译、执行和 ledger 处理
- **AND** lifecycle 负责卸载状态机、资源清理和回滚/收尾调用顺序

### Requirement: 插件列表投影必须由单一投影构建入口维护

系统 SHALL 为插件管理列表、管理摘要、详情和依赖快照提供单一投影构建入口。该入口 MUST 明确输入 mode、当前页或目标插件范围，并统一处理 manifest snapshot、store governance projection、runtime item、host service authorization、dependency summary、租户供应策略和 i18n 展示字段的批量装配。

#### Scenario: 构建管理摘要列表

- **WHEN** 插件管理列表需要摘要投影
- **THEN** 系统通过统一投影构建入口选择 summary mode
- **AND** 复用同一批 manifest、store、runtime 和 dependency 快照
- **AND** 不为摘要列表复制一条独立 manifest 扫描和 runtime 合并流水线

#### Scenario: 构建插件详情

- **WHEN** 插件管理需要单个插件详情
- **THEN** 系统通过统一投影构建入口选择 detail mode 并限制目标插件范围
- **AND** 返回字段保持与列表和同步构建路径一致
- **AND** 不循环调用单项详情接口装配列表、摘要或依赖快照

### Requirement: 插件服务根门面不得直接访问治理 DAO

系统 SHALL 阻断迁移完成后的`internal/service/plugin`根门面生产代码直接访问`internal/dao`、`internal/model/do`或`internal/model/entity`完成插件治理读写。插件治理持久化 MUST 通过`store`，生命周期 SQL MUST 通过`migration`，列表投影 MUST 通过投影构建入口。

#### Scenario: 根门面新增治理读写

- **WHEN** 开发者在插件根包新增或修改生产 Go 代码
- **THEN** 静态治理测试检查根门面生产代码不得导入`lina-core/internal/dao`、`lina-core/internal/model/do`或`lina-core/internal/model/entity`
- **AND** 需要治理数据时必须改用 store、lifecycle、migration 或投影组件的窄契约

#### Scenario: 测试和启动装配例外

- **WHEN** 测试 fixture、治理扫描或启动装配需要接触多个组件
- **THEN** 该访问必须保留在测试、装配或验证边界
- **AND** 不得进入普通业务运行路径

### Requirement: 插件生命周期重复 helper 必须收敛并受长度治理

系统 SHALL 收敛 lifecycle veto 汇总、dynamic decision 汇总、卸载收尾、动态插件启用资格判断和 decision/err 处理等同构逻辑。本变更迁移或重写后的生命周期编排函数和投影函数 MUST 使用命名明确的窄函数表达步骤，单个迁移后业务函数 SHOULD 不超过 60 行；超过时必须在设计或任务记录中说明不可拆分原因。

#### Scenario: 汇总生命周期 veto

- **WHEN** source plugin lifecycle decision 或 dynamic plugin lifecycle decision 需要返回阻断原因
- **THEN** 系统使用同一套汇总 helper 或稳定等价抽象
- **AND** 不再维护两份逐行同构的计数和本地化逻辑

#### Scenario: 拆分 UpdateStatus 四象限

- **WHEN** lifecycle 编排处理 source/dynamic 与 enable/disable 组合
- **THEN** 每个组合分支进入命名明确的窄函数
- **AND** 主流程只表达分派、事务边界和错误语义

### Requirement: 插件升级编排必须归属 upgrade 子组件

系统 SHALL 将源码插件升级状态、动态插件升级 preview、升级 execute、失败诊断、release 提升和缓存发布编排归属到`internal/service/plugin/internal/upgrade`。`internal/service/plugin`根门面 MUST 只保留公开契约、平台治理守卫、必要锁入口协调、输入轻量校验和委托，不得继续承载迁移后的升级长流程。

#### Scenario: 根门面执行升级操作

- **WHEN** 调用方通过插件根服务执行 source 或 dynamic 插件升级
- **THEN** 根门面先执行平台上下文和治理守卫
- **AND** 根门面将升级业务编排委托给`upgrade`子组件
- **AND** 根门面不得直接导入`sourceupgrade`或`runtimeupgrade`包完成升级流程

#### Scenario: upgrade 子组件执行升级状态机

- **WHEN** `upgrade`子组件处理 source 或 dynamic 升级
- **THEN** 它通过构造函数注入的 catalog、store、lifecycle、runtime、integration、dependency、i18n、locker、cache publisher 和 topology 窄契约完成编排
- **AND** 不通过 package-level service locator、构造后 setter 或反向持有插件根门面完成调用

### Requirement: 插件服务不得保留 sourceupgrade 与 runtimeupgrade 平行包

系统 SHALL 在统一升级编排落地后删除`internal/sourceupgrade`和`internal/runtimeupgrade`平行包。源码插件升级和动态插件升级只允许通过`internal/upgrade`及其内部职责文件表达差异；静态治理测试 MUST 阻断生产代码重新导入旧平行包或重新创建旧目录。

#### Scenario: 生产代码重新导入旧升级包

- **WHEN** 插件服务生产 Go 文件导入`internal/sourceupgrade`或`internal/runtimeupgrade`
- **THEN** 静态治理测试失败
- **AND** 调用方必须改为依赖`internal/upgrade`的窄契约或同包内部 helper

#### Scenario: 旧升级目录重新出现

- **WHEN** `apps/lina-core/internal/service/plugin/internal/sourceupgrade`或`apps/lina-core/internal/service/plugin/internal/runtimeupgrade`目录重新出现
- **THEN** 静态治理测试失败
- **AND** 新增升级逻辑必须放入`internal/upgrade`

### Requirement: 插件升级测试必须覆盖统一编排边界

系统 SHALL 为统一升级编排提供职责明确的后端单元测试和静态边界测试。测试必须覆盖 source upgrade status、dynamic preview、source execute、dynamic execute、失败诊断、治理守卫单次执行、缓存发布和旧升级包清零。

#### Scenario: 验证 source 与 dynamic 共用失败诊断

- **WHEN** source 或 dynamic 插件升级在 SQL、callback、release switch 或缓存发布阶段失败
- **THEN** 测试验证失败诊断使用统一 phase 和 message key 语义
- **AND** 有效 release 与派生缓存权威来源保持正确

#### Scenario: 验证根门面不再承载升级长流程

- **WHEN** 开发者修改插件根门面升级文件
- **THEN** 静态边界测试检查根门面不导入旧升级包、不再通过公开 source upgrade 方法再入、不直接拼装 runtime upgrade preview 纯函数

### Requirement: 插件治理读投影必须共享清单快照和依赖索引

系统 SHALL 为插件管理列表、详情、依赖检查、OpenAPI 文档投影和 hook 分发复用同一批 manifest、store、runtime 和 dependency 快照。读取路径 MUST 避免为每个插件、每个依赖检查或每个文档请求重复全量扫描和重复 artifact 解析。

#### Scenario: 管理列表和详情复用统一投影入口

- **WHEN** 系统构建插件管理摘要列表或单插件详情
- **THEN** 系统通过统一投影构建入口读取清单和治理投影
- **AND** 单插件详情不得通过全量扫描全部插件来定位目标插件

#### Scenario: 反向依赖检查使用索引

- **WHEN** 系统为插件列表或卸载校验计算反向依赖
- **THEN** 系统基于当前快照集一次性构建`pluginID`到下游依赖方的索引
- **AND** 单次反向依赖查询不得遍历全部插件快照

#### Scenario: DependencyCheck 请求共享一次快照

- **WHEN** 调用方请求插件依赖检查
- **THEN** 依赖解析、安装快照投影和反向依赖投影共享同一个 request 级 manifest 快照
- **AND** 单次请求不得触发多次完整`ScanManifests`

### Requirement: OpenAPI 插件投影缓存必须绑定运行时和语言版本

系统 SHALL 将动态插件 OpenAPI 文档投影视为插件运行时展示派生缓存。缓存键 MUST 包含`plugin-runtime`修订号、当前 locale 和运行时翻译包版本。插件启停、升级、动态 artifact 刷新或翻译包变化后，后续 OpenAPI 请求 MUST 重建受影响投影。

#### Scenario: Runtime revision 变化后重建 OpenAPI 投影

- **WHEN** 插件`P`升级或禁用并发布新的`plugin-runtime`修订号
- **THEN** 后续 OpenAPI 文档请求不得继续返回旧修订号下的插件`P`路由投影
- **AND** 系统重建或失效对应缓存条目

#### Scenario: Locale 隔离 OpenAPI 投影

- **WHEN** 两个请求使用不同 locale 访问 OpenAPI 文档
- **THEN** 系统使用不同缓存键读取或构建文档投影
- **AND** 不得把旧语言版本的插件文档投影返回给当前请求

### Requirement: 插件运行时组合 delegate 不得静默伪装成功

系统 SHALL 将插件服务内部用于打破启动构造循环的 delegate 限定为最小组合接缝。delegate MUST 提供可诊断的绑定状态；当运行期写入、副作用发布、缓存刷新、依赖校验或认证事件回调在未绑定状态下被调用时，系统 MUST 返回明确错误，不得静默返回成功。只读投影方法在接口无法返回错误时 MAY fail-closed 或返回输入投影，但不得让调用方误以为副作用已经执行。

#### Scenario: 未绑定 runtime delegate 处理认证事件

- **WHEN** `RuntimeDelegate` 尚未绑定插件根服务
- **AND** 调用方触发登录成功、登录失败或登出回调
- **THEN** delegate 返回明确错误
- **AND** 不报告认证事件副作用已经成功执行

#### Scenario: 插件根服务构造后绑定 delegate

- **WHEN** 宿主启动构造插件根服务
- **THEN** 启动装配在插件根服务创建完成后显式绑定 runtime delegate
- **AND** 测试或审查可以确认运行期使用前 delegate 已处于绑定状态

### Requirement: 插件内部 cache 和升级 adapter 必须暴露缺失依赖

系统 SHALL 要求插件内部 cache notifier、dependency validator、source upgrade cache publisher 和 cache freshener 等窄 adapter 在依赖缺失时返回明确错误。生产构造路径 MUST 传入根插件服务或对应窄接口实例；adapter 不得因为 service 为 nil 而返回 nil。

#### Scenario: upgrade cache publisher 缺少根服务

- **WHEN** source upgrade 流程调用未绑定根服务的 cache publisher
- **THEN** publisher 返回明确错误
- **AND** 不发布插件运行时缓存已失效的假成功结果

#### Scenario: dependency validator 缺少根服务

- **WHEN** 生命周期或升级流程调用未绑定根服务的 dependency validator
- **THEN** validator 返回明确错误
- **AND** 不把依赖校验视为通过

### Requirement: 动态插件 route 分发必须按职责拆分

系统 SHALL 将动态插件 route 分发实现按入口编排、路由匹配、鉴权权限、请求 envelope 和响应写回拆分到职责明确的`route*.go`文件。`apps/lina-core/internal/service/plugin/internal/runtime/route.go`MUST 保持为入口和核心 dispatcher 编排文件，行数 MUST 不超过`400`行。拆分不得改变动态插件公开 API 路径、内部 route contract、访问级别、权限查询、数据权限边界、缓存 freshness 检查或响应 envelope 语义。

#### Scenario: route 入口文件保持瘦身

- **WHEN** 静态测试读取`apps/lina-core/internal/service/plugin/internal/runtime/route.go`
- **THEN** 文件行数不得超过`400`行
- **AND** 文件不得重新承载 JWT 解析、角色菜单查询、path pattern 编译、请求 envelope 构造和响应写回的完整实现

#### Scenario: 动态路由公开契约不变

- **WHEN** 动态插件 route contract 声明`path: /api/v1/reports/{id}`、`method: GET`和受保护访问级别
- **THEN** 宿主仍按既有`/x/{plugin-id}/api/v1/...`公开路径匹配和分发
- **AND** 鉴权、权限菜单查询和 session touch 语义与拆分前保持一致
- **AND** 请求传递给动态插件的 route snapshot、路径参数、header、cookie、query、body 和响应写回语义保持一致

#### Scenario: route 拆分不新增运行期依赖

- **WHEN** 系统构造插件 runtime service
- **THEN** route 分发仍使用既有构造函数传入的 executor、registry、session、auth、权限过滤和缓存 freshness 能力
- **AND** 拆分文件不得通过包级变量、service locator 或临时`New()`创建运行期依赖

