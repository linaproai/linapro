## Context

前两个变更已经完成底层拆解和装配直化：`catalog`只负责清单事实源，`store`接管治理持久化，`plugintypes`成为叶子类型包，`plugin.New()`不再依赖内部 wiring setter。当前剩余 C 阶段问题集中在职责归属和重复逻辑：

- `plugin_lifecycle.go`、`plugin_lifecycle_source.go`和`plugin_auto_enable.go`仍在根门面包承载长状态机、源码插件事务编排、自动启用和租户生命周期钩子。
- 旧`internal/lifecycle`仍主要是动态插件 SQL migration executor 和 runtime reconciler 转发，组件名与真实职责不一致。
- lifecycle veto 汇总、dynamic decision 汇总、卸载收尾、动态插件启用资格判断、decision/err 三段式处理存在多处同构复制。
- `plugin_list.go`、`plugin_dependency.go`和`internal/management`分散承担列表、摘要、详情和依赖快照装配，高频路径的扫描成本、缓存键和批量装配边界难以集中验证。
- 插件治理变化后的缓存失效入口仍散在多个方法中，后续 D 统一升级体系前需要先建立单一插件变化发布入口。

本变更是`localdocs/plugin-service-complexity-refactor-plan.md`中的变更 C。它必须保持 B 已建立的构造函数显式依赖和实例状态边界，并为后续 D 的统一升级组件提供稳定的生命周期编排层。

## Goals / Non-Goals

**Goals:**

- 将安装、卸载、启停、源码插件生命周期、启动自动启用和租户生命周期钩子编排下沉到`internal/lifecycle`。
- 将旧 SQL migration executor 迁入`internal/migration`，让`internal/lifecycle`成为真实 lifecycle orchestration owner。
- 根`plugin`门面只保留公共契约、平台治理守卫、输入轻量校验、缓存/锁入口协调和委托，不再直接访问`dao`或承载长状态机。
- 收敛 lifecycle veto/decision helper、卸载收尾、动态插件启用资格判断和重复 decision/err 处理逻辑。
- 收敛列表、摘要、详情和依赖快照投影构建入口，保留缓存命中路径和批量装配策略，明确扫描成本与`N+1`规避依据。
- 收敛插件治理变化后的缓存失效为单一`publishPluginChange`入口，继续使用`plugin-runtime`revision controller。
- 用静态边界测试阻断根门面重新触库、重增长生命周期长流程、列表路径重复扫描和缓存失效入口分散。

**Non-Goals:**

- 不统一 source/dynamic upgrade，也不删除`sourceupgrade`或`runtimeupgrade`包；该工作属于变更 D。
- 不拆分`runtime/route.go`鉴权和分发管线；该工作属于可选变更 E。
- 不改变 HTTP API、DTO、SQL schema、插件 manifest wire、动态插件 guest 协议或前端页面。
- 不把列表缓存深拷贝改为 JSON/gob 等运行时序列化往返；缓存命中路径必须保持低成本。
- 不引入`Options`、`Deps`或聚合结构体包装接口型运行期依赖。

## Decisions

### D1：重建`internal/lifecycle`为编排 owner

`internal/lifecycle`将接收 catalog、store、runtime、integration、migration、dependency、i18n、cache publisher、topology 和必要 host capability 的窄接口。它负责 Install、Uninstall、UpdateStatus、BootstrapAutoEnable、ReconcileAutoEnabledTenantPlugins、租户删除/禁用 precondition 与 notification 等生命周期编排。

Rationale：这些流程是插件生命周期领域内的状态机，不是 facade 公共契约本身。迁入 lifecycle 后，根门面可以稳定为治理守卫和委托层，后续 D 可以在 lifecycle 与 upgrade 之间建立明确边界。

Alternatives considered：只把大函数拆成根包多个文件。该方案可以降低单文件长度，但不能解决根门面直接触库和生命周期 owner 不清的问题。

### D2：SQL migration executor 独立为`internal/migration`

旧`internal/lifecycle`中执行动态插件 SQL、mock data、uninstall SQL 和 migration ledger 的能力迁入`internal/migration`。新 lifecycle 通过窄接口调用 migration executor，不直接落到 SQL 执行细节。

Rationale：migration 是生命周期中的一个执行阶段，但不是 lifecycle 编排本身。独立后可以让 lifecycle 函数按业务步骤表达，也避免后续 upgrade 统一时继续复用名实不符的包。

Alternatives considered：保留旧 lifecycle 包名并新增子文件。该方案会让同一组件同时承载迁移执行和完整编排，继续模糊职责。

### D3：保持平台治理守卫在根门面入口

HTTP 可达的生命周期写操作仍由根`plugin`门面执行平台上下文守卫、权限前置和显式锁/缓存入口协调，然后委托 lifecycle。内部启动路径和租户能力回调通过窄接口调用 lifecycle，但必须保留现有平台/租户上下文语义。

Rationale：平台治理守卫是 facade 面向外部调用的稳定边界；迁移编排不应放宽租户隔离或绕过既有错误语义。

Alternatives considered：把所有守卫一并下沉到 lifecycle。该方案会让内部启动调用与 HTTP 管理调用共享过宽入口，增加误用风险。

### D4：投影构建使用单一模式化入口而非多条流水线

列表、管理摘要、详情和依赖快照使用一个投影 builder，输入为稳定的 mode 与当前页/当前插件范围。builder 统一负责 manifest snapshot、store governance projection、runtime item、host services authorization、dependency summary、i18n 展示字段和 tenant provisioning projection 的批量装配。

Rationale：这些路径共享同一批清单、治理和 runtime 数据。统一入口可以集中控制扫描次数、缓存键、批量读取和字段完整性，避免后续为某个页面字段复制一条新流水线。

Alternatives considered：保留四条流水线，只抽公共 helper。该方案仍会让扫描顺序和缓存失效分散，难以验证`N+1`边界。

### D5：缓存失效统一为`publishPluginChange`

生命周期、同步、上传、启停、卸载、源码升级、动态升级和租户供应策略成功写入后统一调用一个插件变化发布入口。该入口负责更新`plugin-runtime`revision、失效管理读模型、触发 runtime/frontend/i18n/WASM 派生缓存刷新语义，并记录 reason。

Rationale：缓存一致性规则要求关键运行时数据失效幂等、可重试、可观测。单一入口让后续 D 合并升级路径时不会复制多套失效逻辑。

Alternatives considered：继续保留每个流程自己的失效函数。该方案短期改动较小，但容易在新增路径中遗漏管理读模型或 runtime 派生缓存。

### D6：业务控制流 context key 改为显式参数或 options

安装 mock data、启动自动启用、依赖快照缓存等普通业务控制流不再依赖隐式 context key。确实属于一次 startup orchestration 的大快照可以继续通过上下文传递，但只能作为启动性能优化输入，不能改变生命周期语义。

Rationale：显式参数让调用链和测试更直接，减少隐藏状态对编排迁移的影响。

Alternatives considered：保留 context key 并只迁移位置。该方案不会降低调用方理解成本。

## Risks / Trade-offs

- [Risk] 生命周期状态机迁移规模大，容易改变 install/uninstall/update status 细节。→ Mitigation：按 Install、Uninstall、UpdateStatus、源码生命周期、自动启用、租户钩子分批迁移，每批运行插件全包测试和相关精准测试。
- [Risk] 列表投影统一后可能遗漏某些管理弹窗需要的字段。→ Mitigation：保留现有`plugin_list_test.go`覆盖，并新增 summary/detail 字段完整性测试，任务记录中列出字段不变判断。
- [Risk] 缓存失效统一可能改变集群模式 freshness 语义。→ Mitigation：继续复用`revisionctrl.Controller`和`plugin-runtime`domain，补充单机/集群或静态审查记录，验证 reason、scope 和失效触发点不丢失。
- [Risk] 根门面不直接触库的边界可能与短期迁移冲突。→ Mitigation：先建立静态边界测试，迁移完成前允许测试以待完成任务形式失败；每完成一个迁移批次同步收紧 allowlist。

## Migration Plan

1. 建立 C 阶段基线和静态治理测试：根门面生产代码不得直接`dao.`访问；迁移后的 lifecycle/投影函数长度可控；缓存失效入口单一。
2. 将旧`internal/lifecycle`迁为`internal/migration`，保持现有动态 SQL 执行测试通过。
3. 新建真实`internal/lifecycle`编排 service，先迁 Install/Uninstall，再迁 UpdateStatus/Enable/Disable。
4. 迁移源码插件安装、卸载、回滚和启动 auto-enable，删除相关业务 context key。
5. 迁移租户生命周期 precondition/notification 和 tenant auto provisioning。
6. 收敛 lifecycle decision/veto、uninstall 收尾和动态启用资格 helper。
7. 收敛列表/摘要/详情/依赖快照投影 builder 和管理读模型缓存。
8. 收敛`publishPluginChange`缓存失效入口并补充验证。
9. 运行`go test ./internal/service/plugin/... -count=1`、`go test ./internal/cmd -count=1`、相关 i18n/cachecoord 测试、OpenSpec 严格校验和静态检索。

Rollback 使用普通 Git 回退。本变更不引入数据库迁移、HTTP API 契约或外部 wire 变化。

## Open Questions

- 无。
