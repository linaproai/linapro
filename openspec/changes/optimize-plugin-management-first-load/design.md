## Context

插件管理页面当前首屏链路包含前端和后端两类同步成本：

- 前端 `apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue`在模块加载时静态导入详情、上传、授权、卸载、升级和生命周期前置条件弹窗；这些文件与共享视图约 `4215`行，首次进入页面时会进入同一个路由加载链路。
- 前端 grid 已传递 `pageNum`、`pageSize`，但后端 `apps/lina-core/api/plugin/v1/plugin_list.go`的 `ListReq`没有分页字段，服务层 `List`先构建完整 `managementList`，再在内存中过滤。
- 后端列表响应 `PluginItem`同时包含列表字段和弹窗字段：`dependencyCheck`、`requestedHostServices`、`authorizedHostServices`、`declaredRoutes`等都随列表返回。
- controller 列表路径会为所有列表项解析宿主服务表注释，并通过 `buildManagedCronJobMap`逐插件调用 `ListCronDeclarationsByPlugin`；该方法下游会再次 `ScanManifests()`，放大动态 `wasm` manifest 解析成本。
- 启动后已有异步 `PrewarmManagementList`，但它不阻塞 HTTP 可用；且现有缓存键受 locale 和 runtime bundle version 影响，预热竞态或语言不命中时，首个管理请求仍会同步构建完整治理读模型。

本设计属于 `apps/lina-core`宿主通用插件治理能力、管理工作台前端适配层与插件运行时派生缓存的组合变更。它不修改插件目录结构、不新增插件本地资源、不新增业务字典或数据库模型。

## Goals / Non-Goals

**Goals:**

- 让插件管理首次进入只触发一个有分页边界的摘要列表 API，并且首屏不依赖逐行详情请求。
- 将首屏列表后端成本限制在“发现摘要、注册表状态、当前页轻量投影”范围内，避免依赖检查、cron 声明、host service 资源表注释和动态路由审查进入列表路径。
- 保留详情、安装、升级、卸载、授权确认等治理动作需要的完整服务端判断，前端不自行推断依赖图或 host service 语义。
- 让摘要读模型和详情读模型具备明确缓存键、失效触发、集群同步和故障降级策略。
- 补充可证明的性能测试和 E2E：证明首屏列表没有后端 `N+1`和前端瀑布式补查。

**Non-Goals:**

- 不改变插件 manifest、动态插件 artifact 格式、host service 授权协议或依赖声明语义。
- 不新增 SQL 表、索引或持久化缓存模型；如实现过程中发现必须新增持久化查询辅助结构，需要另按数据库规则补充设计和任务。
- 不重做插件管理页面视觉布局，只调整数据加载边界和弹窗加载方式。
- 不让前端复制依赖检查、cron 声明或授权确认的服务端决策逻辑。

## Decisions

### 1. `GET /plugins`定义为摘要列表接口

`GET /plugins`继续保留 RESTful 只读语义和 `plugin:query`权限，但请求 DTO 增加 `pageNum`、`pageSize`，设置默认值和最大上限。响应列表项拆分为摘要 DTO，只包含首屏表格和行级操作需要的字段，例如：

- 插件身份与展示：`id`、`name`、`description`、`type`。
- 版本和运行时摘要：`version`、`effectiveVersion`、`discoveredVersion`、`runtimeState`、`upgradeAvailable`、`abnormalReason`、必要的 `lastUpgradeFailure`摘要。
- 治理状态：`installed`、`installedAt`、`enabled`、`autoEnableManaged`、`autoEnableForNewTenants`、`supportsMultiTenant`、`scopeNature`、`installMode`、`statusKey`、`updatedAt`。
- 授权摘要：`authorizationRequired`、`authorizationStatus`、`hasMockData`。

列表响应不得包含 `dependencyCheck`、`requestedHostServices`、`authorizedHostServices`、`declaredRoutes`或 cron 详情。完整字段由 `GET /plugins/{id}`和动作预览接口提供。

备选方案：

- 继续返回完整 `PluginItem`并依赖预热。拒绝原因：预热存在竞态和缓存键不命中，插件数量增长时仍会把 detail 成本带到首屏。
- 列表只返回 ID，再由前端逐行请求详情。拒绝原因：违反接口性能契约，会产生前端瀑布式调用和后端逐项详情装配。

### 2. 后端拆分摘要读模型和详情读模型

服务层新增或调整为明确的摘要列表路径，例如 `ListSummaries(ctx, input)`；详情路径继续通过 `Get(ctx, id)`或等价方法提供完整治理投影。

摘要列表路径的装配顺序为：

1. 校验 `plugin-runtime` freshness。
2. 获取已缓存或本次构建的轻量 manifest/registry snapshot。
3. 在轻量投影上完成过滤、排序、分页和总数计算。
4. 只对当前页构建摘要 DTO 需要的字段。

列表路径不得先为所有插件执行 `CheckPluginDependencies`、`ListCronDeclarationsByPlugin`、host service 表注释解析或 route review 转换后再裁剪字段。对于源码和动态插件 manifest 这种文件/产物发现源，无法完全按数据库侧分页；设计约束是先构建轻量发现索引，避免在分页前构建重型详情投影。注册表读取仍应使用集合化读取，不得按插件逐项查询。

详情路径保留完整依赖检查、host service 授权、声明路由和 cron 审查数据。详情仅针对一个插件 ID，允许执行该插件需要的详细装配；如后续出现批量详情场景，必须新增批量契约而不是循环调用单项详情。

### 3. cron 和依赖检查从列表路径移出

依赖检查结果仍由服务端产生，并在详情、安装、升级、卸载等决策路径展示。列表只提供阻断状态的必要摘要时，也必须来自服务端预计算的有界字段，不能通过前端解释依赖图。

cron 声明只在详情或授权审查需要展示 host service 资源时装配。若实现保留列表外的批量审查场景，应增加 `ListCronDeclarationsByPlugins(ctx, ids)`或等价批量契约，内部一次 manifest snapshot 扫描后按插件 ID 分组；不得在动态结果集循环中调用 `ListCronDeclarationsByPlugin`。

### 4. 前端首屏按需加载弹窗和详情数据

插件管理页面保留 `useVbenVxeGrid`和现有 Page 布局。弹窗组件通过 `defineAsyncComponent`、Vben 支持的异步 `connectedComponent`或局部懒加载包装按需加载；首屏路由 chunk 只包含表格、筛选、行操作和必要的轻量 helper。

前端 API 模型拆分为 `PluginListItem`和 `PluginDetail`或等价命名，避免把详情字段继续扩散到列表行类型。打开详情、安装授权、卸载确认等弹窗前，前端通过 `pluginDetail(id)`加载完整详情；升级弹窗继续优先使用既有 `pluginUpgradePreview(id)`，不从列表读取 preview 数据。

首屏不得为每行自动调用 `GET /plugins/{id}`。只有用户打开相关弹窗或执行治理动作时，才发起详情或预览请求。

### 5. 摘要缓存复用 `plugin-runtime`协调

摘要列表和详情读模型都是插件运行时派生缓存，权威源为源码插件嵌入 manifest、动态插件 active release/artifact manifest、`sys_plugin`注册表状态、宿主配置和运行时 i18n bundle。

缓存策略：

- 摘要列表缓存键至少包含 locale、runtime bundle version、plugin runtime revision、筛选条件归一化和分页参数；也可以缓存未分页轻量索引后按请求分页，但缓存内容不得包含 detail-only 字段。
- 详情缓存键至少包含 plugin ID、locale、runtime bundle version 和 plugin runtime revision。
- 写路径继续通过既有插件安装、启用、禁用、卸载、升级、active release 切换、源码同步和 i18n bundle 变化触发失效。
- `cluster.enabled=false`时使用进程内缓存和本地 revision；`cluster.enabled=true`时复用 `plugin-runtime` Redis revision/event，不新增只在当前节点可见的独立缓存域。
- 当无法确认 `plugin-runtime` freshness 且超过允许陈旧窗口时，不得返回过期摘要或详情；应按既有插件 runtime 故障策略 conservative-hide 或返回结构化错误。

为避免首个请求并发击穿，摘要构建应使用已有单飞机制或新增局部 singleflight，但不得新增全局服务定位或独立缓存敏感服务实例。

### 6. 权限、数据权限、`i18n`和数据库边界

插件管理属于平台治理控制面，不承载租户/组织业务数据列表；本变更继续依赖 `plugin:query`和各治理动作权限，不新增常规数据权限过滤。安全收益是首屏列表减少依赖、host service、声明路由和 cron 详情暴露，敏感治理细节进入受控详情和动作路径。

`i18n`影响预计为类型和接口文档边界变化，不新增用户可见文案。若实现新增或调整前端文案、API `dc`源文本或错误消息，必须同步宿主运行时语言包和 `apidoc`资源，并执行 `make i18n.check`或等价验证。

数据库预计无变更。若实现中为了支撑分页或排序引入新的表字段、索引或 SQL，需要新增当前迭代幂等 SQL，并运行数据库初始化、DAO 生成和相关 Go 编译门禁。

## Risks / Trade-offs

- [Risk] 详情弹窗首次打开会比过去略多一次详情请求 → Mitigation：只在用户表达意图后加载，且详情路径只装配单插件；弹窗显示局部 loading，避免阻塞首屏。
- [Risk] 列表和详情 DTO 拆分会触碰前端类型与多个弹窗入参 → Mitigation：先建立清晰的 `PluginListItem`/`PluginDetail`类型边界，再逐个弹窗改为接收 detail 数据。
- [Risk] 摘要读模型若仍复用完整 `runtime.PluginItem`，实际性能不会改善 → Mitigation：任务和测试必须断言列表路径不调用依赖检查、逐插件 cron 声明和 detail-only 投影函数。
- [Risk] 预热覆盖语言不足导致首次请求仍冷启动 → Mitigation：预热轻量摘要，且通过 singleflight 控制冷启动并发；实现记录预热 locale 策略，不能把正确性依赖于预热完成。
- [Risk] 缓存失效遗漏会让管理页显示过期插件状态 → Mitigation：缓存键绑定 `plugin-runtime` revision，写路径复用既有 revision/event，测试覆盖安装/启用/禁用/升级后的摘要失效。
- [Risk] E2E 只验证页面出现而无法证明接口链路改善 → Mitigation：E2E 必须监听网络请求，断言首屏只有列表请求且无逐行详情请求，打开弹窗后才出现对应详情或预览请求。

## Migration Plan

1. 先扩展后端 DTO 和服务接口，保留 `GET /plugins/{id}`作为详情兼容路径。
2. 切换前端列表类型和详情加载流程，确保首屏不依赖旧列表中的 detail-only 字段。
3. 调整缓存预热和失效测试，确保单机和集群策略有审查记录。
4. 删除列表路径中的依赖检查、cron 扫描和 host service/detail 投影装配。
5. 运行后端、前端、E2E 和 OpenSpec 验证。

回滚策略为恢复前端从列表行读取完整字段、后端 `GET /plugins`返回旧 `PluginItem`投影，并保留详情接口不变；由于不涉及数据库迁移，回滚不需要数据修复。

## Open Questions

- `pageSize`最大值建议为 `100`，默认值建议为 `20`。实现时可按现有 Vben grid 默认分页习惯微调，但必须有服务端上限。
- `lastUpgradeFailure`在列表中保留完整还是摘要字段，需要以当前表格和状态提示的实际使用为准；若只用于弹窗诊断，应移入详情。
