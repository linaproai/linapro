## 1. 基线盘点与迁移清单

- [x] 1.1 盘点`catalog.Service`全部方法、`serviceImpl`字段和`catalog_wiring.go`setter，按清单读取、治理存储、纯类型、依赖语义、副作用回调五类建立迁移清单。

  记录：

  | 类别 | 当前成员 | 迁移方案 |
  |------|----------|----------|
  | 清单读取 | `ManifestReader`、`SQLAssetCatalog`、`FrontendAssetCatalog`、`ScanEmbeddedSourceManifests`、`ScanManifests`、`ScanManifestsByID`、`ReadSourcePluginManifestContent`、`ReadSourcePluginAssetContent`、`LoadManifestFromYAML`、`LoadManifestFromArtifactPath`、`GetDesiredManifest`、`GetActiveManifest`、`ValidateManifest`、`ValidateUploadedRuntimeManifest`、SQL/Frontend 资源路径发现 | 保留在`catalog`；`LoadReleaseManifest`依赖发布行和授权快照，拆分为`store`读取发布快照后由`catalog`按显式 artifact/package 路径解析 |
  | 治理存储 | `Registry`、`ReleaseStore`、`Governance`中读写`sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_plugin_migration`、授权快照、启动快照刷新、`SetRegistryRuntimeState`、`SetAutoEnableForNewTenants`、`SyncRegistryReleaseReference`、`SyncReleaseMetadata` | 迁入`store`，对外返回`store`自有投影；生产接口不返回`DAO`、`DO`、`Entity`或 GoFrame 查询模型 |
  | 纯类型 | `status.go`、`type.go`、`scope.go`、`generation.go`、`metadata.go`中的状态/类型/scope/generation/release/runtime/migration/resource 枚举与纯派生函数，版本比较和 dependency spec 的纯值对象 | 迁入`plugintypes`；不保留`catalog`别名或 wrapper |
  | 依赖语义 | `catalog/dependency.go`与`internal/dependency`共同承担 dependency spec 克隆、语法、求值和 registry snapshot 应用 | dependency 声明结构和纯克隆迁入`plugintypes`；依赖求值保留在`internal/dependency`；`catalog`不再作为依赖文法事实源 |
  | 副作用回调 | `Wiring`、`catalog_wiring.go`和`serviceImpl`字段`backendLoader`、`artifactParser`、`dynamicManifestLoader`、`nodeStateSyncer`、`menuSyncer`、`resourceRefSyncer`、`releaseStateSyncer`、`hookDispatcher` | 删除`catalog`setter 和长期字段；扫描输入改为显式参数或下沉清单资源读取；菜单、资源引用、hook、节点/发布状态同步上提到编排入口 |

  DI 来源检查：任务 1 仅记录基线，不新增运行期依赖；后续新增`store`时在对应任务记录 owner、创建位置和传递路径。
- [x] 1.2 为`menuSyncer`、`hookDispatcher`、`resourceRefSyncer`、`nodeStateSyncer`、`releaseStateSyncer`记录旧触发点、新调用 owner、错误处理语义和对应回归测试。

  记录：

  | 回调 | 旧触发点 | 新 owner | 错误处理语义 | 回归验证 |
  |------|----------|----------|--------------|----------|
  | `menuSyncer` | `SyncManifest`中 source 插件已安装且版本匹配时同步菜单 | `plugin`门面显式同步/source install/source enable 编排入口，首轮不下沉 lifecycle | 保持原语义：同步失败返回错误，阻断当前同步或生命周期动作 | 复用/补充插件同步、安装和启用测试，断言菜单同步仍发生且失败可见 |
  | `hookDispatcher` | `SetPluginStatus`写入状态后分发 enable/disable hook | `plugin.UpdateStatus`或等价启用/禁用编排入口 | 保持原顺序：治理状态写入成功后分发；hook 失败返回错误 | 复用/补充 lifecycle observer、启用/禁用测试 |
  | `resourceRefSyncer` | `syncMetadataDependents`在已安装插件的 metadata 同步后写资源引用 | 安装、同步、升级和 release 切换编排入口 | 保持原语义：资源引用同步失败返回错误，阻断当前治理动作 | 复用/补充安装、动态升级、资源引用测试 |
  | `nodeStateSyncer` | `SyncManifest`、`SyncMetadata`、`SetPluginStatus`间接同步节点状态，`BuildGovernanceSnapshot`读取当前节点状态 | 写入迁入`store`并由编排入口显式调用；治理读取由`store`按当前 node id 查询投影 | 保持原语义：节点状态同步失败返回错误；治理读取失败返回错误 | 复用/补充 runtime reconciler、startup consistency、治理快照测试 |
  | `releaseStateSyncer` | `SetPluginStatus`写入 registry 后同步 active release runtime state | 启用/禁用编排入口显式调用 runtime/store 发布状态同步 | 保持原语义：发布状态同步失败返回错误 | 复用/补充启用/禁用和 active release 状态测试 |

  缓存一致性：任务 1 不改变触发点；后续实现必须证明管理读模型、runtime cache revision 和启动快照刷新未因 owner 迁移遗漏。
- [x] 1.3 判定`Manifest`、`ResourceSpec`、依赖声明、版本值对象等结构的归属：纯投影迁入`plugintypes`，携带扫描/校验上下文的结构保留在`catalog`。

  记录：

  | 结构/函数 | 归属 | 理由 |
  |-----------|------|------|
  | `Manifest`、`ArtifactManifest`、`ArtifactSpec`、`ArtifactSQLAsset`、`ArtifactFrontendAsset`、`PublicAssetSpec`、`MenuSpec`、`HookSpec`、`ResourceSpec`、`RouteSpec`及清单校验函数 | `catalog` | 这些结构直接承载`plugin.yaml`/artifact 解析、资源路径、扫描上下文或校验语义，仍属于清单事实源 |
  | `ManifestSnapshot`、`ResourceRefDescriptor`、`HostServiceAuthorizationInput`、`HostServiceAuthorizationDecision` | `store`投影或`plugintypes`纯投影按使用点拆分 | 它们是发布/授权/治理持久化投影，不应继续挂在`catalog`；其中无 DAO 依赖的输入/快照可成为`plugintypes`值对象，写库方法归`store` |
  | `PluginType`、安装/启用状态、`ScopeNature`、`InstallMode`、generation 状态、`ReleaseStatus`、`HostState`、`NodeState`、`RuntimeUpgradeState`、`MigrationDirection`、`MigrationExecutionStatus`、resource kind/owner/filter/order/operation/access mode 枚举及纯派生/normalize/string 函数 | `plugintypes` | 纯类型和值对象，调用方不应为了使用状态语义导入`catalog` |
  | `DependencySpec`、`PluginDependencySpec`、`FrameworkDependencySpec`和 clone/normalize helper | `plugintypes` | 依赖声明是清单中的纯值对象，同时被`dependency`求值、manifest snapshot 和测试复用；求值器仍归`internal/dependency` |
  | 语义版本比较、版本范围值对象 | `plugintypes` | 只依赖字符串/语义版本规则，不需要清单扫描或数据库 |
  | `RuntimeUpgradeProjection`、`RuntimeUpgradeFailure` | `plugintypes`投影，读取失败迁移/发布行的方法归`store` | 投影本身是纯读模型；从`sys_plugin_release`和`sys_plugin_migration`装配投影属于`store`治理读取 |

## 2. 建立`plugintypes`叶子包

- [x] 2.1 新建`apps/lina-core/internal/service/plugin/internal/plugintypes`，迁移插件 ID、状态、安装状态、类型、scope、generation、版本等纯类型和值对象。

  记录：已创建`plugintypes`叶子包，并迁移插件类型、安装/启用状态、scope/install mode、generation、语义版本校验/比较、依赖声明值对象、resource 枚举、runtime/governance 纯状态枚举、`ResourceRefDescriptor`和`RuntimeUpgradeProjection`等纯投影。`Manifest`、`MenuSpec`、`ResourceSpec`和菜单类型校验仍保留在`catalog`，因为它们承载`plugin.yaml`解析和 manifest 校验上下文。`HostServiceAuthorizationInput`/`Decision`仍随授权写库方法暂留`catalog`，在`store`迁移时一起转为存储投影或纯值对象。

  DI 来源检查：本任务仅新增纯值对象包，无运行期依赖、无构造函数依赖和无共享实例变化。
- [x] 2.2 更新`catalog`、`runtime`、`integration`、`lifecycle`、`dependency`、门面和测试中对这些纯类型的引用，不保留旧别名或 wrapper。

  记录：已将调用方更新为直接导入`plugintypes`，并删除`catalog`中已迁移的`status.go`、`type.go`、`scope.go`、`dependency.go`、`catalog_code.go`等旧纯类型文件；未保留`catalog`别名或 wrapper。当前仍从`catalog`引用的菜单类型属于 manifest 校验结构，授权确认类型属于待迁移的存储投影，不计入本阶段纯类型残留。
- [x] 2.3 验证`plugintypes`非测试代码不依赖`catalog`、`store`、`runtime`、`integration`、`dao`、`do`或`entity`。

  验证：

  - `rg '^import|"apps/lina-core/internal/service/plugin/internal/(catalog|store|runtime|integration|dao|do|entity)|"apps/lina-core/internal/(dao|model/(do|entity))' apps/lina-core/internal/service/plugin/internal/plugintypes -n`仅显示标准库导入。
  - `cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过。

## 3. 建立`store`治理存储子组件

- [x] 3.1 新建`apps/lina-core/internal/service/plugin/internal/store`主文件、`Service`接口、`serviceImpl`和稳定投影类型，接口不暴露`DAO`、`DO`、`Entity`或 GoFrame 查询模型。

  记录：已新增`internal/store`，以`Service`接口、`serviceImpl`、`PluginRecord`、`ReleaseRecord`、`MigrationRecord`、`NodeStateRecord`、`ManifestSnapshot`、`RuntimeStatePatch`和`GovernanceSnapshot`作为稳定投影。`store`生产实现内部使用`DAO/DO`完成治理表读写，但导出接口和投影不返回`DAO`、`DO`、`Entity`或 GoFrame 查询模型。

  DI 来源检查：`store`由`plugin.New()`和`testutil.NewServices()`在组合根创建；依赖显式来自`catalog.Service`的只读清单/资源 helper 和当前节点`Topology`适配器；同一`store`实例传递给`lifecycle`、`runtime`、`integration`、`frontend`、`openapi`、`sourceupgrade`和根门面复用。
- [x] 3.2 迁移`catalog/registry.go`中的插件注册表读写、安装状态、启用状态、自动启用策略和 runtime state 写入能力到`store`。

  记录：`SyncManifest`、`GetRegistry`、`ListAllRegistries`、`SetPluginStatus`、`SetPluginInstalled`、`SetRegistryRuntimeState`、`SetAutoEnableForNewTenants`、`SyncRegistryReleaseReference`和启动期 registry snapshot 刷新均已迁入`store`。调用方改为通过`storeSvc`读写治理状态，`catalog`只继续提供`GetDesiredManifest`、扫描和资源访问。
- [x] 3.3 迁移`catalog/release.go`、`catalog/authorization.go`、`catalog/governance.go`中的发布、授权快照和治理投影读写能力到`store`。

  记录：发布行读写、manifest snapshot 构建/解析、hostServices 授权快照、卸载清理策略、治理快照和 runtime-upgrade 投影均已迁入`store`。host service 授权 helper 测试已放入`internal/store`，store 侧继续校验插件自有表边界。
- [x] 3.4 迁移当前由`runtime`通过`nodeStateSyncer`、`releaseStateSyncer`提供给`catalog`的节点状态和发布状态 DAO 读写职责到`store`，保留运行时 reconciliation 的业务调用语义。

  记录：节点状态和发布状态写库由`runtime`/reconciler 显式调用`storeSvc`完成，`catalog`不再持有`nodeStateSyncer`或`releaseStateSyncer`回调。reconciler 安装、卸载、升级和回滚路径仍在治理状态写入后刷新 runtime/frontend/openapi/cache 派生状态。
- [x] 3.5 更新调用方，让治理表读写走`store`稳定投影，清单扫描访问继续走`catalog`。

  记录：根门面、`runtime`、`integration`、`lifecycle`、`frontend`、`openapi`、`dependency`、`sourceupgrade`、测试 fixture 均已改为组合`catalog`清单能力与`store`治理投影。`catalog`包生产代码静态检索无`DAO/DO/Entity`、`runtime`或`integration`导入。

## 4. 收窄`catalog`清单职责

- [x] 4.1 收窄`catalog.Service`，移除治理存储方法、setter wiring 方法和副作用回调接口。

  记录：`catalog.Service`现在仅组合`ManifestReader`、`SQLAssetCatalog`、`FrontendAssetCatalog`和`ManifestMetadata`。治理存储、发布、授权、启动快照和 runtime-upgrade 方法已从`catalog`接口移除。
- [x] 4.2 删除`catalog_wiring.go`和`serviceImpl`中的`backendLoader`、`artifactParser`、`dynamicManifestLoader`、`nodeStateSyncer`、`menuSyncer`、`resourceRefSyncer`、`releaseStateSyncer`、`hookDispatcher`字段。

  记录：`catalog_wiring.go`已删除，`serviceImpl`仅保留`configSvc`。静态边界测试`TestCatalogSetterWiringRemoved`覆盖`catalog`包内无`Set*`方法、无旧回调字段，且`plugin.New()`不调用`catalogSvc.Set*`。
- [x] 4.3 将 artifact 解析、active dynamic manifest 加载和 source backend config 装载改为清单/资源读取能力或扫描入口显式参数，避免`catalog`长期持有`runtime`或`integration`。

  记录：artifact manifest 解析和源码 backend hook/resource YAML 装载已下沉为`catalog`清单解析能力；active dynamic manifest 加载由`runtime`、`frontend`、`openapi`和`integration`通过`store`发布投影显式解析，不再通过`catalog`长期持有的 runtime 回调。
- [x] 4.4 合并或重命名`manifest_validate.go`与`manifest_validation.go`，确保结构校验、格式校验和资源校验职责可读。

  记录：旧`manifest_validation.go`已删除，manifest 结构校验和资源校验收敛到`manifest_validate.go`及清单解析相关文件，避免两个相近文件继续分散职责。
- [x] 4.5 收敛`catalog/dependency.go`与`internal/dependency`的边界，使依赖文法与依赖求值不再跨包互相引用。

  记录：依赖声明纯值对象和 clone/normalize helper 归入`plugintypes`，依赖求值和快照构建保留在`internal/dependency`。`catalog/dependency.go`已删除，调用方直接使用`plugintypes`声明和`dependency`求值器。

## 5. 上提副作用调用点

- [x] 5.1 在显式同步、安装、卸载、启用、禁用、升级等现有编排入口中，用显式调用替代`catalog`内部菜单同步、资源引用同步和 hook 分发。

  记录：显式同步、源码安装/卸载/启用/禁用、动态安装/卸载/启用/禁用、上传、runtime upgrade 和 source upgrade 路径均改为在门面、`runtime`或`sourceupgrade`编排中显式调用`store`、`integration`、`runtime`和缓存刷新能力。`catalog`扫描和读取路径不触发菜单、资源引用、hook 或缓存副作用。
- [x] 5.2 保留既有缓存失效、管理读模型失效、运行时状态同步和错误处理语义，并在任务记录中说明缓存一致性是否无语义变化。

  记录：治理写库 owner 从`catalog`迁移到`store`，未新增缓存种类或改变集群同步模型。既有`markRuntimeCacheChanged`、runtime reconciler revision、frontend bundle invalidation、OpenAPI/management read model invalidation 和 integration enabled snapshot 刷新仍由原编排入口显式触发。缓存一致性语义无变化。
- [x] 5.3 更新或补充现有单元测试，覆盖删除`catalog`副作用回调后的安装、启用、禁用、卸载、同步和资源引用关键路径。

  记录：已更新根门面、`runtime`、`integration`、`lifecycle`、`sourceupgrade`和`store`相关测试。新增`plugin_boundary_test.go`固化依赖边界，`store_startup_snapshot_test.go`承载 registry/release/startup/runtime-upgrade 投影断言，资源引用和菜单同步测试改为通过`store`治理投影准备状态。

## 6. 边界治理与验证

- [x] 6.1 新增或扩展插件服务内部 import 边界治理测试，覆盖`plugintypes`零兄弟依赖、`catalog`不依赖`runtime/integration/dao/do/entity`、`store`不泄漏生成模型。

  记录：新增`apps/lina-core/internal/service/plugin/plugin_boundary_test.go`。`TestPluginInternalImportBoundaries`解析非测试 Go 文件，覆盖`plugintypes`不得导入兄弟包或生成模型、`catalog`不得导入`runtime`、`integration`或生成模型；`TestPluginStoreExportedSurfaceDoesNotLeakGeneratedModels`覆盖`store`导出类型/函数签名不泄漏`DAO/DO/Entity`。
- [x] 6.2 增加静态检索或测试断言，确认`catalog`包内无`Set*`wiring 方法、无`Syncer`/`Dispatcher`回调字段、`plugin.New()`不再调用`catalog.Set*`。

  记录：`TestCatalogSetterWiringRemoved`解析`catalog`生产源码和`plugin.go`，确认`catalog`包内无`Set*`方法、`serviceImpl`无旧回调字段，且`plugin.New()`不调用`catalogSvc.Set*`。手动静态检索仅剩`runtimeSvc.SetHookDispatcher(integrationSvc)`和测试 fixture 中同等 runtime wiring，不属于 catalog setter 环。
- [x] 6.3 运行`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`。

  验证：已通过，覆盖根插件门面、`catalog`、`store`、`runtime`、`integration`、`lifecycle`、`frontend`、`openapi`、`dependency`、`sourceupgrade`、`wasm`和`runtimecache`相关包。
- [x] 6.4 运行`cd apps/lina-core && go test ./internal/cmd -count=1`或记录等价启动绑定覆盖。

  验证：已通过，覆盖宿主启动装配编译和插件服务构造绑定。
- [x] 6.5 运行`openspec validate split-plugin-catalog-store-types --strict`。

  验证：已通过，`Change 'split-plugin-catalog-store-types' is valid`。

## 7. 影响分析与审查

- [x] 7.1 记录本变更无 HTTP API、路由、DTO、OpenAPI、前端页面、插件`plugin.yaml`语义、host service wire 和 SQL schema 影响。

  记录：本变更仅调整`apps/lina-core/internal/service/plugin`内部子组件职责、测试和 OpenSpec 任务记录；未修改`api/`、HTTP 路由、DTO、OpenAPI 源文本、前端页面、插件`plugin.yaml`语义、host service wire 字符串、`manifest/sql/`或生成的`dao/do/entity`。
- [x] 7.2 记录`i18n`影响判断：本变更不修改运行时用户可见文案、API 文档源文本、语言包或翻译缓存；若实现中触及错误消息或文案，必须同步补充治理。

  记录：无`i18n`资源影响。变更未新增或修改用户可见文案、菜单/路由/按钮文案、API 文档元数据、语言包、插件清单文案或翻译缓存语义。
- [x] 7.3 记录数据权限影响判断：治理存储 owner 迁移不得扩大插件列表、详情、授权快照或 host service 数据可见性；若修改查询语义，必须说明过滤边界和测试覆盖。

  记录：数据权限边界无放宽。平台治理入口仍通过`ensurePlatformGovernance`守卫；插件 host service 授权快照迁入`store`后仍校验资源/表属于当前插件命名空间；列表、详情、授权快照和 runtime 路由读取只更换内部 owner，不新增跨租户或跨插件可见性。
- [x] 7.4 记录缓存一致性影响判断：移动写库 owner 后必须保留既有失效触发点、作用域、单机/集群语义和恢复路径；若确认无语义变化，明确记录。

  记录：缓存一致性语义无变化。`store`仅接管权威 DB 写入；原有管理读模型失效、runtime cache revision、frontend bundle invalidation、OpenAPI projection 刷新、integration enabled snapshot 和集群协调触发点仍由显式编排入口负责。未新增本地-only 缓存或新的跨实例失效机制。
- [x] 7.5 完成实现、验证和任务记录后调用`lina-review`进行代码和规范审查。

  Lina 审查记录：

  - 审查范围：`apps/lina-core/internal/service/plugin/**`和`openspec/changes/split-plugin-catalog-store-types/**`，共 151 个已跟踪或未跟踪变更文件；`openspec/changes/archive/**`中的删除/聚合变更为当前工作区无关 churn，未纳入本变更审查范围。
  - 已读取规则入口与规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`。
  - 结论：未发现阻塞问题。`catalog`已收窄为清单读取/校验/资源访问；`store`为治理持久化唯一 owner 且导出 API 不泄漏生成模型；`plugintypes`为叶子包；副作用调用点由编排入口显式触发。
  - 验证证据：`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate split-plugin-catalog-store-types --strict`通过；`git diff --check -- apps/lina-core/internal/service/plugin openspec/changes/split-plugin-catalog-store-types`通过。
  - 影响结论：无 HTTP API、路由、DTO、前端页面、插件`plugin.yaml`语义、host service wire、SQL schema、生成代码、运行时用户文案或语言包影响；纯内部重构未触发 E2E 质量审查。
