## 1. 基线与边界确认

- [x] 1.1 静态梳理动态路由、详情、依赖检查、OpenAPI 投影、hook 分发和启停路径的`ScanManifests`、artifact 解析和`InvalidateAllCache`调用点。
  - 调用点记录：动态路由`runtime_route_match.go`和`runtime_route.go`重复调用`resolveActiveOrDesiredManifest`；OpenAPI 投影`openapi_projection.go`先`ScanManifests`再逐插件解析 active manifest；依赖检查`plugin_dependency.go`通过 projection 构建依赖快照；启停和集群回调集中在`plugin_runtime_cache.go`，其中 cluster revision 回调与动态单插件变更存在`InvalidateAllCache`；`catalog_manifest.go`和`store_release.go`是 artifact/release manifest 解析入口。
- [x] 1.2 建立 30 插件测试 fixture 或计数桩，记录改造前热路径解析次数、扫描次数、编译次数和关键 DB 访问次数。
  - 已在`catalog`同包测试建立 30 个动态 artifact fixture 和解析计数桩；稳态第二次扫描、索引化单插件读取、文件替换和按插件失效均有解析次数断言。数据库访问不在`catalog/store`缓存底层测试中发生，后续依赖与路由投影任务继续覆盖请求级扫描次数。
- [x] 1.3 记录影响分析：缓存一致性、数据权限、`i18n`、开发工具跨平台、测试策略和`apps/lina-core/pkg/plugin`README 是否需要同步。
  - 影响分析：缓存权威源为源码嵌入 manifest、动态 artifact 文件 stat 与 release 行；单机使用进程内缓存和显式失效，集群继续通过`plugin-runtime`修订号触发对账；缓存未命中回源解析。当前底层变更不新增 HTTP API、SQL、数据权限规则或用户可见文案；`i18n`仅影响后续 OpenAPI 投影缓存键；未修改开发工具脚本；`pkg/plugin`README 已存在 Access 命名同步改动，本读模型底层缓存暂不需要新增公共插件 README 说明。

## 2. 清单读模型缓存

- [x] 2.1 在`catalog`和`store`边界内实现源码 manifest、动态 desired manifest、release manifest 和 YAML 快照缓存。
- [x] 2.2 为动态 artifact 扫描实现目录枚举加文件`stat`守卫，并维护`pluginID`到 artifact 路径的索引。
- [x] 2.3 为单插件读取和 release 快照读取改用缓存索引路径，保留缓存未命中时的回源解析。
- [x] 2.4 在动态路由请求上下文中复用已解析 manifest，消除同请求重复 artifact 解析。
- [x] 2.5 补充缓存命中、未命中、文件替换、按插件失效、release LRU 和只读共享约束测试。
  - 已覆盖命中、文件替换、按插件失效、release manifest 缓存和只读克隆。release 缓存当前使用不可变 release 身份键和插件失效，未引入独立 LRU；原因是当前 release manifest 缓存只保存被读取的 active/archive manifest，键绑定 release ID、checksum 和 package path，先避免新增淘汰抽象，后续如出现内存压力再单独扩展。

## 3. Wasm 编译缓存与派生缓存失效

- [x] 3.1 将已知`pluginID`或 active release 的治理变更改为按 artifact 路径失效`WASM`编译缓存。
- [x] 3.2 将集群 peer 的`plugin-runtime`修订号回调改为 active artifact 差异对账，避免无条件全量失效。
  - Peer 首次观察 revision 时会建立 active artifact 快照并按当前 active path 做一次保守失效；后续 revision 只失效路径变化、新增或消失的 active artifact。无法读取 registry/release 时保守退回全量失效并记录日志。
- [x] 3.3 将`WASM`编译过程移出全局缓存写锁，实现 per-artifact single-flight 和失败重试。
- [x] 3.4 保持现有 lease 引用计数语义，并补充单插件失效、并发编译、编译失败重试和不同 artifact 互不阻塞测试。

## 4. 治理读投影收敛

- [x] 4.1 让插件详情、管理列表、hook 分发和依赖检查改用统一 manifest 快照读取路径。
  - 插件详情、管理列表和依赖检查继续通过`buildPluginProjection`、`management.WithManifestSnapshot`和`WithDependencySnapshotCache`共享单次 manifest/store 快照；hook 分发读取仍走`catalog`统一 manifest 缓存边界，动态 hook 的 active manifest 通过 release manifest 缓存回源，未新增副作用写入。
- [x] 4.2 在依赖快照构建时生成反向依赖索引，并用对照测试证明输出与旧遍历逻辑一致。
  - 已新增`ReverseDependencyIndex`，依赖快照缓存同时保存反向索引；`CheckReverse`可复用预构建索引，未传入时保持自动构建兼容路径。`dependency`包对照测试覆盖旧遍历输出一致性，并用 30 插件 fixture 验证只返回目标相关下游依赖。
- [x] 4.3 合并`DependencyCheck`端点内的多次扫描，单请求复用一个 request 级 manifest 快照。
  - `CheckPluginDependencies`现在自行启用 request cache，并用同一批 dependency snapshots 同时计算 install 与 reverse 结果；公开入口测试断言单次请求只构建一次 catalog/store dependency snapshot。
- [x] 4.4 为 OpenAPI 插件投影增加按`plugin-runtime`修订号、locale 和运行时翻译包版本的缓存；若实施时裁剪，需在任务记录说明原因。
  - 已实现动态路由 OpenAPI 路径子集缓存，缓存键包含`plugin-runtime`revision、locale 和 runtime bundle version；本地`publishPluginChange`和集群 peer revision 回调都会显式失效该投影。未缓存整个 OpenAPI 文档，避免影响请求 origin、静态路由和后续 apidoc 本地化流程。

## 5. 验证与审查

- [x] 5.1 运行覆盖变更包的`go test <changed-package> -count=1`，涉及路由绑定或构造变更时运行宿主启动绑定包测试。
  - 已验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/store ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/dependency ./internal/service/plugin/internal/lifecycle ./internal/service/plugin -count=1`通过；`go test ./internal/cmd -count=1`通过；`git diff --check`通过。针对修复回归还单独运行了`go test ./internal/service/plugin -run 'TestInstallPersistsExplicitDynamicInstallMode|TestUninstallForceClearsDynamicOrphanWhenArtifactsMissing' -count=1 -v`、`go test ./internal/service/plugin -run TestInstallSameVersionDynamicPluginRefreshesArchivedReleaseArtifact -count=1 -v`和`go test ./internal/service/plugin/internal/store -run 'TestLoadReleaseManifestUsesReleaseCache|TestLoadReleaseManifestCacheRejectsMissingArchive|TestParseManifestSnapshotUsesContentCache' -count=1 -v`。
- [x] 5.2 使用计数桩或单元测试断言 30 插件 fixture 下热路径 artifact 解析、单插件详情扫描、依赖检查扫描和单插件编译缓存失效成本有界。
  - 覆盖证据：`catalog`30 artifact fixture 断言稳态扫描、索引化单插件读取、文件替换和按插件失效的 parse 次数有界；`dependency`30 插件反向依赖索引测试断言只返回目标相关下游依赖；公开`CheckPluginDependencies`测试断言单请求只构建一次 catalog/store dependency snapshot；`wasm`测试覆盖单插件失效、per-artifact single-flight、编译失败重试和不同 artifact 互不阻塞。
  - 回归修复记录：动态安装现在把已验证并应用显式`install_mode`的 desired manifest 传入 runtime reconcile，避免 runtime 再次从 catalog 读取原始 manifest；release manifest cache 增加文件`stat`和`os.SameFile`守卫，active archive 删除或替换后不会返回过期 manifest；直接覆写 artifact 的测试 fixture 显式调用`InvalidateManifestCache`，与生产上传路径保持一致。
  - DI 来源检查：`openapi.New(catalogSvc, storeSvc, openapiRevision, i18nSvc)`由根插件 service 启动装配传入；`DeferredRevisionReader`在 runtime cache revision controller 创建后绑定同一共享实例；`i18nSvc`复用根 service 注入实例。未新增外部运行期依赖或独立服务图。
  - 影响判断：缓存七要素沿用设计记录，权威源为源码嵌入 manifest、动态 artifact、`sys_plugin_release`快照；一致性模型为本地显式失效加`plugin-runtime`修订号；失效点覆盖上传、安装、启停、刷新、卸载和 peer revision；集群复用 revision/event；展示缓存允许修订号预算内短暂陈旧，执行路径仍校验 freshness；失败时回源解析或保守拒绝；恢复依赖修订号检查、事件监听、safety sweep 和进程重启重建。`i18n`无新增用户可见文案，OpenAPI 投影缓存键包含 locale 和 runtime bundle version；数据权限无 HTTP/SQL/数据访问契约变化，无新增数据暴露；未修改开发工具脚本，无跨平台脚本影响；`apps/lina-core/pkg/plugin`README 的现有改动属于其他工作区差异，本次内部读模型缓存不需要同步公共插件 README。
- [x] 5.3 运行`openspec validate plugin-runtime-read-model-performance --strict`。
  - 已验证：`openspec validate plugin-runtime-read-model-performance --strict`通过。
- [x] 5.4 执行`lina-review`，审查缓存一致性七要素、DI 来源检查、`N+1`规避、`i18n`影响判断和 README 同步判断。
  - Lina 审查范围：`plugin-runtime-read-model-performance`当前变更文件，包含`apps/lina-core/internal/service/plugin`下的目标实现和测试、未跟踪新增缓存文件，以及`openspec/changes/plugin-runtime-read-model-performance`下 OpenSpec 文档。工作区中`apps/lina-core/pkg/plugin/README.md`、`apps/lina-core/pkg/plugin/README.zh-CN.md`、`apps/lina-core/pkg/plugin/pluginhost/*`、`openspec/specs/plugin-host-domain-capabilities/spec.md`和`openspec/changes/plugin-runtime-auth-snapshot-guardrails/`属于并行未纳入本次目标范围的既有差异，未修改、未回退。
  - 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/backend-go.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`；后端审查同步使用`goframe-v2`技能。
  - 审查结论：未发现阻塞问题。缓存一致性七要素已覆盖，权威源为源码嵌入 manifest、动态 artifact 文件和`sys_plugin_release`快照；一致性模型为本地显式失效加`plugin-runtime`修订号；失效点覆盖上传、安装、启停、刷新、卸载和 peer revision；集群同步复用 revision/event；展示投影允许修订号预算内短暂陈旧，执行路径仍校验 freshness；失败降级回源解析或保守拒绝；恢复依赖修订号检查、事件监听、safety sweep 和进程重启重建。
  - DI 来源检查：`openapi.New(catalogSvc, storeSvc, openapiRevision, i18nSvc)`由根插件 service 启动装配创建；`DeferredRevisionReader`在`runtimeCacheRevisionCtrl`创建后绑定根插件 service；`i18nSvc`复用根共享实例。未发现业务构造函数或请求路径临时创建关键服务图，未新增外部运行期依赖。
  - 性能与`N+1`判断：manifest 缓存、release/YAML 快照缓存、请求级 dependency snapshot、`ReverseDependencyIndex`、OpenAPI 投影缓存和 per-artifact WASM single-flight 均将重复解析、扫描、编译或反向依赖遍历收敛到有界路径；未发现新的循环 DB 查询或逐插件逐项详情补查。
  - 影响判断：本变更不新增 HTTP API、DTO、路由、权限标签、SQL、前端页面、用户可见文案或数据访问授权边界；`i18n`无新增文案，OpenAPI 投影缓存键包含 locale 和 runtime bundle version；数据权限无新增暴露；开发工具和脚本未修改，无跨平台脚本影响；`apps/lina-core/pkg/plugin`README 的并行差异不属于本内部读模型缓存变更，本次不需要同步公共插件 README；无用户可观察前端行为变化，未触发 E2E 新增要求。
  - 当前验证证据：`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/store ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/dependency ./internal/service/plugin/internal/lifecycle ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`git diff --check`通过；`openspec validate plugin-runtime-read-model-performance --strict`通过。
