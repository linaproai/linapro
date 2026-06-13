## Context

当前`apps/lina-core/internal/service/plugin/internal/catalog`同时承载插件清单事实源和插件治理执行路径。它既扫描、解析、校验`plugin.yaml`和插件资源，又读写`sys_plugin`、`sys_plugin_release`、授权快照和治理投影，还通过`SetBackendLoader`、`SetArtifactParser`、`SetDynamicManifestLoader`、`SetNodeStateSyncer`、`SetMenuSyncer`、`SetResourceRefSyncer`、`SetReleaseStateSyncer`、`SetHookDispatcher`接收`runtime`和`integration`回调。

这使`catalog`成为插件服务内部依赖环根：`plugin.New()`必须先构造`catalog`、`runtime`、`integration`，再通过 setter 回注互相需要的能力。该结构隐藏副作用调用时机，阻碍后续将插件服务装配改为构造函数显式注入，也让列表、同步、安装、启用、升级等路径难以判断哪些步骤只读、哪些步骤写库、哪些步骤触发菜单/资源/hook 副作用。

本变更只治理宿主插件服务内部边界，不修改插件对外协议、HTTP API、SQL schema、前端页面或插件清单语义。

## Goals / Non-Goals

**Goals:**

- 将插件基础类型和值对象迁入`plugintypes`叶子包，避免`catalog`继续作为所有类型的事实源。
- 将插件治理表读写、发布/授权/治理投影、node/release state 持久化迁入`store`，并通过稳定投影对外暴露。
- 将`catalog`收窄为插件清单事实源，只保留清单扫描、解析、校验和访问职责。
- 删除`catalog`的 setter wiring、反向回调字段和隐藏副作用调用。
- 明确`menuSyncer`、`hookDispatcher`、`resourceRefSyncer`等副作用的新调用 owner，避免重构后静默丢失治理同步。
- 增加静态边界验证，防止`plugintypes`、`catalog`和`store`重新产生反向依赖或内部模型泄漏。

**Non-Goals:**

- 不重构`plugin.New()`的全部 setter 装配；本变更只删除`catalog`相关 setter 环。
- 不将生命周期编排整体迁入`internal/lifecycle`。
- 不统一 source/dynamic upgrade 体系。
- 不迁移或重构`runtimecache`。
- 不合并`capabilityhost/internal/*cap`目录。
- 不继续改造`internal/wasm`分发层。
- 不以代码行数或目录数下降作为硬验收。

## Decisions

### D1：建立`plugintypes`作为叶子包

`plugintypes`承载插件 ID、类型、状态、安装状态、scope、generation、版本值对象和跨子组件共享的纯投影类型。该包不得依赖`catalog`、`store`、`runtime`、`integration`、`dao`、`do`或`entity`。

Rationale：类型和值对象当前散落在`catalog`，迫使不需要清单扫描能力的子组件也依赖`catalog`。叶子包让调用方表达“我只需要插件类型语义”，而不是隐式依赖清单服务。

Alternatives considered：继续让`catalog`导出类型。该做法短期改动少，但会让`catalog`继续保留跨子组件中心地位，无法拆除环根。

### D2：建立`store`作为插件治理存储 owner

`store`接管`registry.go`、`release.go`、`authorization.go`、`governance.go`以及当前通过`nodeStateSyncer`、`releaseStateSyncer`间接访问的节点/发布状态数据库读写。`store`对外返回自有投影，例如`PluginRecord`、`ReleaseRecord`、`AuthorizationSnapshot`，不得返回`DAO`、`DO`、`Entity`或 GoFrame 查询模型。

Rationale：插件治理表读写属于稳定存储职责，不属于清单扫描职责。集中到`store`后，`catalog`可以退回只读清单事实源，列表、生命周期、升级路径也可以直接选择读取清单或读取治理投影。

Alternatives considered：把 DAO 读写分散到`runtime`、`integration`和门面调用点。该做法会消除`catalog`上帝包，但会制造多个新的数据库 owner，增加事务边界和性能审查难度。

### D3：`catalog`只保留清单扫描、校验和访问

`catalog`继续负责源码插件和动态插件清单的发现、解析、manifest 校验、资源声明访问和嵌入式清单读取。`Manifest`、`ResourceSpec`等结构按职责拆分：纯数据投影可迁入`plugintypes`，携带扫描路径、校验上下文或解析中间状态的结构留在`catalog`。

Rationale：清单是插件资源事实源，保留在`catalog`符合命名与职责。混入治理表写入和副作用才是复杂度来源。

Alternatives considered：把所有 manifest 结构一次性迁入`plugintypes`。该做法看似统一类型，但会把扫描和校验中间状态误提升为公共契约，扩大导出面。

### D4：副作用调用上提到编排入口

`catalog`不再持有`menuSyncer`、`hookDispatcher`、`resourceRefSyncer`。安装、同步、启用、禁用、卸载、升级等现有编排入口在写入`store`成功后显式调用`integration`或`runtime`提供的窄能力。首轮允许这些显式调用暂时留在插件门面或既有编排文件中，后续 lifecycle 下沉时再迁入`internal/lifecycle`。

Rationale：副作用调用时机必须跟业务动作绑定，而不是隐藏在清单或 registry helper 里。上提后可以直接审查事务、缓存失效和错误处理边界。

Alternatives considered：为`catalog`注入更窄的同步接口。虽然可减少字段数量，但仍保留“清单/存储 helper 隐式触发副作用”的设计问题。

### D5：扫描输入使用显式参数或下沉能力，不做实例字段

`backendLoader`、`dynamicManifestLoader`、`artifactParser`等现有字段分两类处理：可独立于`runtime`/`integration`的解析能力下沉到清单或资源读取层；必须由调用场景提供的信息改为扫描入口参数。它们不得继续作为`catalog.serviceImpl`构造后的可变字段。

Rationale：这些依赖并非`catalog`长期状态，而是特定扫描动作的输入或解析策略。显式参数能让调用方和测试清楚知道扫描依赖什么。

Alternatives considered：通过构造函数一次性注入这些接口。该做法能去掉 setter，但仍会让`catalog`长期持有`runtime`/`integration`能力，依赖方向不变。

### D6：边界测试作为验收门禁

本变更新增或扩展 import 边界治理测试，至少覆盖：

- `plugintypes`非测试代码不得导入`catalog`、`store`、`runtime`、`integration`、`dao`、`do`、`entity`。
- `catalog`非测试代码不得导入`runtime`、`integration`、`dao`、`do`、`entity`。
- `store`导出 API 不得泄漏`DAO`、`DO`、`Entity`。
- `catalog`包内不得保留`Set*` wiring 方法或`Syncer`/`Dispatcher`回调字段。

Rationale：本变更的主要收益是结构性边界，单纯功能测试很难防止后续回归。边界测试能把讨论结果变成持续约束。

Alternatives considered：只依赖代码审查。该做法会把环根治理变成审查记忆，不符合长期维护目标。

## Risks / Trade-offs

- [Risk] 删除`catalog`副作用回调后遗漏菜单、资源引用或 hook 同步。→ Mitigation：实现任务必须为每个被删除回调记录旧触发点和新触发点，并用现有生命周期/同步测试覆盖关键路径。
- [Risk] 首轮拆分会短期增加文件和 import churn。→ Mitigation：硬验收聚焦依赖方向和职责边界，不把代码行数下降作为成功条件。
- [Risk] `Manifest`、`ResourceSpec`归属误判会扩大公共类型或造成循环依赖。→ Mitigation：先迁纯值对象；对携带扫描/校验上下文的结构保持在`catalog`，必要时拆出稳定投影。
- [Risk] `store`成为新的上帝包。→ Mitigation：`store`只拥有插件治理表读写和稳定投影，不承载清单扫描、生命周期编排、菜单同步、hook 分发或运行时 reconciliation。
- [Risk] 缓存失效触发点随写入 owner 迁移而模糊。→ Mitigation：本变更不改变缓存语义；移动写库方法时必须保留原失效触发点，并在任务记录中说明无缓存语义变化或补充缓存一致性验证。
