## Requirements

### Requirement: 插件服务内部必须分离清单、存储和类型职责

系统 SHALL 在 `apps/lina-core/internal/service/plugin/internal` 内将插件基础类型、插件清单事实源和插件治理存储拆分为职责明确的内部子组件。`plugintypes` MUST 只承载纯类型和值对象；`catalog` MUST 只承载清单扫描、解析、校验和访问；`store` MUST 作为治理表读写和投影的 owner。`catalog` 不得继续作为治理表读写或副作用聚合点。

### Requirement: 插件 catalog 不得持有 runtime 或 integration 反向回调

系统 SHALL 禁止 `catalog` 通过 setter、包级变量或构造后回注持有 `runtime`、`integration` 或插件门面能力。`catalog` 不得声明 `Set*` wiring 方法，不得持有 `menuSyncer`、`hookDispatcher`、`resourceRefSyncer` 等反向回调字段。

### Requirement: 插件服务内部构造必须无 setter 回注

系统 SHALL 在插件服务生产构造路径中通过构造函数逐项显式传入运行期依赖。`plugin.New()`、`runtime.New()`、`integration.New()`、`lifecycle.New()`、`sourceupgrade.New()` 和 `upgrade.New()` 构造完成后 MUST 处于可用状态，不得依赖构造后 `Set*` 方法或 `ValidateRequiredDependencies` 兜底。

### Requirement: 插件服务内部组件不得互持宽 service 形成循环

系统 SHALL 使用窄契约表达插件服务内部组件之间的运行期调用。`runtime` 不得持有完整 `integration.Service`；`integration` 不得持有完整 `runtime.Service`；`sourceupgrade`/`upgrade` 不得持有插件根门面。

### Requirement: 插件服务包级可变运行期状态必须清零

系统 SHALL 禁止插件服务生产代码通过包级变量保存运行期可变状态。生命周期 observer、integration shared state、runtime cache observed revision、runtime reconciler once/mutex、WASM host service runtime 等 MUST 归属于 service 实例、启动期共享对象或缓存协调后端。

### Requirement: capabilityhost 领域适配器必须保留在同一组件边界内

系统 SHALL 将 `capabilityhost` 的领域适配器实现保留在 `capabilityhost` 组件边界内。单文件 `capabilityhost/internal/*cap` 微包不得继续作为长期结构。

### Requirement: 插件服务接线治理必须由静态测试固化

系统 SHALL 提供静态治理测试，阻断生产代码重新引入 wiring setter、包级可变运行期状态、旧 runtimecache import、WASM Configure 配置入口或 testutil 对生产接线流程的复刻。

### Requirement: 插件生命周期编排必须归属 lifecycle 子组件

系统 SHALL 将插件安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子编排归属到 `internal/lifecycle`。根门面 MUST 只保留公共契约、平台治理守卫和委托。

### Requirement: 插件 SQL migration executor 必须独立于 lifecycle 编排

系统 SHALL 将插件生命周期 SQL 文件执行归属到独立 `internal/migration` 组件。lifecycle 编排 MUST 只通过窄接口调用 migration executor。

### Requirement: 插件列表投影必须由单一投影构建入口维护

系统 SHALL 为插件管理列表、管理摘要、详情和依赖快照提供单一投影构建入口，统一处理 manifest snapshot、store governance projection、runtime item、host service authorization、dependency summary 和 i18n 展示字段的批量装配。

### Requirement: 插件服务根门面不得直接访问治理 DAO

系统 SHALL 阻断迁移完成后的根门面生产代码直接访问 `internal/dao`、`internal/model/do` 或 `internal/model/entity`。治理持久化 MUST 通过 `store`，SQL MUST 通过 `migration`。

### Requirement: 插件生命周期重复 helper 必须收敛并受长度治理

系统 SHALL 收敛 lifecycle veto 汇总、dynamic decision 汇总、卸载收尾等同构逻辑。迁移后业务函数 SHOULD 不超过 60 行。

### Requirement: 插件升级编排必须归属 upgrade 子组件

系统 SHALL 将源码插件升级和动态插件升级的 preview、execute、失败记账、release 提升和缓存发布归属到 `internal/upgrade`。根门面 MUST 只保留平台治理守卫和委托。

### Requirement: 插件服务不得保留 sourceupgrade 与 runtimeupgrade 平行包

系统 SHALL 在统一升级编排落地后删除 `internal/sourceupgrade` 和 `internal/runtimeupgrade` 平行包。静态治理测试 MUST 阻断生产代码重新导入旧平行包。

### Requirement: 插件升级测试必须覆盖统一编排边界

系统 SHALL 为统一升级编排提供职责明确的后端单元测试和静态边界测试，覆盖 source upgrade status、dynamic preview、source execute、dynamic execute、失败诊断、治理守卫单次执行、缓存发布和旧升级包清零。

### Requirement: 插件治理读投影必须共享清单快照和依赖索引

系统 SHALL 让插件详情、依赖检查、OpenAPI 投影和 hook 分发复用同一批清单快照。反向依赖检查改为索引直查，不得遍历全部插件快照。

### Requirement: OpenAPI 插件投影缓存必须绑定运行时和语言版本

系统 SHALL 让 OpenAPI 插件投影缓存键包含运行时 bundle 版本和 locale，避免旧投影跨版本或跨语言命中。

### Requirement: 动态插件 route 分发必须按职责拆分

系统 SHALL 将动态插件路由分发代码按职责拆分为路由匹配、鉴权与权限查询、请求封装和响应写回。入口文件 SHOULD 不超过 400 行。公共 host call helper MUST 位于 wasm 公共层，不得定义于具体领域文件。

### Requirement: 插件服务内部依赖方向必须由治理验证固化

系统 SHALL 为插件服务内部清单、存储和类型拆分提供静态边界验证，覆盖 `plugintypes` 零兄弟依赖、`catalog` 不依赖 `runtime`/`integration`/`dao`、`store` 不泄漏生成模型。

### Requirement: 插件运行时组合 delegate 不得静默伪装成功

系统 SHALL 将 delegate 限定为最小组合接缝。delegate MUST 提供可诊断的绑定状态；未绑定运行期调用 MUST 返回明确错误。

### Requirement: 插件内部 cache 和升级 adapter 必须暴露缺失依赖

系统 SHALL 要求 cache notifier、dependency validator、upgrade cache publisher 等窄 adapter 在依赖缺失时返回明确错误。
