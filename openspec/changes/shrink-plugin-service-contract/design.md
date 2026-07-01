## Context

`plugin.Service`当前同时暴露插件管理、启动编排、动态路由、前端资源、源码插件集成、定时任务、状态查询、provider env、租户生命周期和 lifecycle observer 等能力。实际调用方通常只使用其中少数方法，例如插件控制器只需要管理方法，`jobhandler`只需要生命周期 observer、启用插件 ID 和单插件可执行 job，`httpstartup`只需要启动和运行时路由方法。

项目已有`move-narrow-interfaces-to-consumers`和`standardize-plugin-domain-services`变更，说明接口治理方向已经从“为了分类而分类”转向“真实 owner 和真实消费者边界”。本次变更沿用该方向，只在插件根包内部保留服务职责私有 facet；单一消费者的特殊组合只有在完整`Service`会显著扩大不稳定依赖面时才放在消费者包内。

## Goals / Non-Goals

**Goals:**

- 缩小`plugin.Service`审查面，让调用方依赖插件服务的具体消费边界。
- 删除无生产入口或仅包装统一入口的方法，减少重复契约。
- 合并插件 job 查询和状态变更的重复窄方法，同时保持参数语义清晰。
- 保持启动期显式依赖注入和共享插件服务实例，不引入 service locator、聚合依赖结构体或新的运行期服务图。

**Non-Goals:**

- 不改变 HTTP API、DTO、路由、权限标签、数据库结构或插件 manifest。
- 不改变动态`hostServices`wire method、`pluginbridge`guest API 或`pkg/plugin/capability`领域契约。
- 不迁移插件内部 lifecycle、runtime、integration、frontend、upgrade 的业务实现。
- 不新增 E2E；本次是内部 Go 契约治理，无用户可观察行为变化。

## Decisions

### 使用真实消费者 facet，而不是纯注释分类

在`plugin`根包新增生产者侧私有 facet 接口，例如管理、启动、运行时 HTTP、集成、job、状态、provider env 和租户生命周期。`Service`作为唯一导出的根组合接口保留；插件管理控制器、HTTP 启动上下文和`RuntimeDelegate`等统一入口边界继续使用`Service`，定时任务和`apidoc`等消费者按需要使用本地私有窄接口或包内私有 facet。

选择该方案是因为这些 facet 对应稳定职责边界，同时不会把多个插件服务入口暴露给包外调用方。单纯在`Service`里加分组注释不能降低审查成本；多级`Plugins().Runtime().List...`accessor 会增加调用层级和初始化顺序理解成本。`RuntimeDelegate`只绑定启动期同一个插件根服务实例，额外保留单用途窄接口会增加接口数量和维护成本，因此该边界复用`Service`。

### 只合并语义重复且参数可清晰表达的方法

插件 job 查询合并为`ListManagedJobs(ctx, ManagedJobQuery)`，用查询参数表达可执行、已安装、插件 ID 和 handler 是否需要返回。这样可以替代当前四个`List*Job*`方法，又避免调用方循环调用单项详情接口。

插件状态变更合并为`UpdateStatus(ctx, pluginID, UpdateStatusOptions)`，保留目标状态和动态插件授权确认输入。`Enable`、`Disable`和`SetStatus`不再作为根服务方法暴露；公共 capability 接口仍按插件能力命名规则保留`SetStatus`。

Auth hook 的三个包装方法删除，调用方使用已有`DispatchHookEvent`表达具体事件。该方法已经是统一 hook 分发入口，保留包装方法只会扩大接口。

### 保留清晰状态查询和租户生命周期方法

`IsInstalled`、`IsEnabled`、`IsProviderEnabled`、`IsEnabledAuthoritative`以及租户生命周期前置校验/通知方法不合并成通用`Check`或`Notify`。这些方法的错误语义、降级语义和调用方含义不同，强行泛化会降低可读性。

### 不新增运行期依赖或缓存语义

所有调用方继续复用启动期同一个插件服务实例或其 facet 接口。新增 facet 只是 Go 类型边界变化，不新增缓存、锁、数据库查询、状态快照或跨实例协调机制。

## Risks / Trade-offs

- **调用点较多，容易遗漏测试 fake** → 使用静态检索旧方法名和编译门禁发现遗漏。
- **facet 过多可能重新变成分类噪音** → 仅保留有真实职责边界支撑的包内私有 facet；没有职责差异的分组不创建。
- **合并 job 查询可能隐藏 handler 返回边界** → `ManagedJobQuery`显式包含`IncludeHandlers`，管理投影默认不发布 handler，运行时注册路径才请求 handler。
- **状态变更合并可能弱化 enable/disable 可读性** → `UpdateStatusOptions`保留目标状态字段和授权字段，控制器仍按启用/禁用 API 显式构造请求。

## Migration Plan

1. 创建`ManagedJobQuery`和`UpdateStatusOptions`，在根门面中实现新委托方法。
2. 调整 lifecycle、integration、runtime delegate、控制器、启动编排、`cron`、`jobhandler`和`apidoc`调用点。
3. 删除旧根服务方法和无生产入口的`SyncSourcePlugins`，同步测试替身。
4. 使用`rg`确认旧方法名不再被生产代码引用。
5. 运行 OpenSpec 严格校验、相关 Go 测试和 Lina 审查。

## Open Questions

无。用户已确认按计划实施，项目无历史兼容负担，内部 Go 契约可一次性破坏性收敛。
