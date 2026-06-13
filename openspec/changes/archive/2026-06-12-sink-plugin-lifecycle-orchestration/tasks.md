## 1. 基线与治理测试

- [x] 1.1 盘点当前`plugin_lifecycle.go`、`plugin_lifecycle_source.go`、`plugin_auto_enable.go`、`internal/lifecycle`、`internal/management`、`plugin_list.go`、`plugin_dependency.go`和缓存失效入口，记录迁移批次、owner、DI 来源、缓存与数据权限影响。

  记录：

  | 区域 | 当前状态 | C 阶段迁移批次与目标 owner |
  |------|----------|-----------------------------|
  | `plugin_lifecycle.go` | 根门面仍承载`Install`、`Uninstall`、`UpdateStatus`、动态 lifecycle precondition/notification、source/dynamic veto 汇总、`withInstallMockData`和`sourceLifecycleInstallOptionsContextKey`。 | 任务 3.1、3.2、3.3、3.4 和 4.4 分批下沉到`internal/lifecycle`；根门面最终只保留平台治理守卫、输入轻量校验和委托。 |
  | `plugin_lifecycle_source.go` | 根门面仍直接导入`dao.SysPlugin`和`do.SysPlugin`，负责 source install/uninstall/rollback、source lifecycle callback 和 mock-data 装饰。 | 任务 4.1 迁入`internal/lifecycle`；根门面 DAO 访问在该任务完成后从静态 allowlist 删除。 |
  | `plugin_auto_enable.go` | `BootstrapAutoEnable`、启动轮询、auto-enable install/enable 和`ReconcileAutoEnabledTenantPlugins`仍在根门面。 | 任务 4.2、4.3 迁入`internal/lifecycle`，startup 语义改为显式 options 或内部启动入口。 |
  | 旧`internal/lifecycle` | 变更前混合动态 install/uninstall 转发与 SQL migration executor。 | 任务 2.1 已把 SQL executor 迁入`internal/migration`；后续 2.2、3.x、4.x 将真实 lifecycle 编排迁入该包。 |
  | `internal/management`与`plugin_list.go` | 仍有`syncAndList`、`buildManagementList`、`buildManagementSummaryList`、`buildManagementDetail`四条投影流水线和`managementListCache`。 | 任务 5.1、5.2、5.3 收敛为统一投影 builder；保留低成本显式深拷贝，不引入 JSON/gob 往返。 |
  | `plugin_dependency.go` | 仍使用`dependencySnapshotCacheContextKey`承载请求内依赖快照，依赖快照装配与列表投影分离。 | 任务 5.1、5.4 将 dependency snapshot mode 并入统一投影 builder；任务 4.4 只清理业务控制流 context key，启动/请求内只读快照需单独记录保留理由。 |
  | 缓存失效入口 | 仍分散在`markRuntimeCacheChanged`、`syncEnabledSnapshotAndPublishRuntimeChange`、`invalidateRuntimeUpgradeCaches`和显式`managementListCache.Invalidate()`调用点。 | 任务 6.1、6.2 收敛为统一`publishPluginChange`或等价入口；任务 6.3 补缓存一致性记录。 |

  DI 来源检查：任务 1.1 只记录基线，不新增运行期依赖。任务 2.1 新增`migration.Service`运行期依赖，由`plugin.New()`和`internal/testutil.NewServices()`创建，owner 为`internal/migration`，依赖`catalog.Service`读取 SQL asset、依赖`store.Service.GetRelease`读取 release 投影；同一实例传给根门面、`runtime`和`sourceupgrade`复用，未新增共享后端或本地-only 缓存。

  缓存影响：任务 1.1 不改变缓存语义；当前分散入口已记录为后续 6.x 收敛对象。数据权限影响：任务 1.1 不改变查询或写入可见性；后续迁移必须保持平台治理守卫、插件 host service 授权和租户供应边界。
- [x] 1.2 新增或扩展 C 阶段静态边界测试：根门面迁移完成后不得直接导入`dao/do/entity`，迁移后的 lifecycle/投影函数长度受控，列表投影入口唯一，插件治理变化通过统一 cache publisher，D/E 范围不被误纳入。

  记录：已扩展`apps/lina-core/internal/service/plugin/plugin_boundary_test.go`，新增`TestPluginLifecycleOrchestrationStaticBoundaries`。该测试当前固化三类 C 阶段护栏：旧`internal/lifecycle`不得重新出现 SQL executor 方法或`SysPluginMigration`账本访问；`internal/migration/migration.go`必须存在并作为 SQL migration owner；C 阶段不得提前引入`internal/upgrade`统一升级包。根门面 DAO/DO/Entity 检查使用过渡 allowlist，当前仅允许既有`plugin_lifecycle_source.go`、`plugin_install_mock_data.go`和现有 route/menu 投影桥接文件继续存在；后续 4.1 完成后必须从 allowlist 删除 source lifecycle DAO/DO 例外。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run TestPluginLifecycleOrchestrationStaticBoundaries -count=1`通过。函数长度、列表入口唯一和统一 cache publisher 的最终严格断言将在 3.x、5.x、6.x 代码迁移完成后收紧，当前不以失败测试阻塞过渡状态。
- [x] 1.3 记录本变更无 HTTP API、DTO、SQL schema、前端页面、插件 manifest wire 和`apps/lina-plugins/*`文件影响；记录`i18n`、数据权限、缓存一致性、开发工具跨平台和测试策略影响判断。

  记录：本变更 C 当前只修改`apps/lina-core/internal/service/plugin/**`内部 service 组织、测试和 OpenSpec 任务记录；未修改`api/`、HTTP 路由、DTO、OpenAPI 文档源文本、数据库 schema、宿主或插件 SQL 文件、前端页面、插件 manifest wire、动态插件 guest 协议或`apps/lina-plugins/*`文件。`i18n`资源无新增文案；MockDataLoadError 类型 owner 从`lifecycle`迁到`migration`后仍由根门面的既有 bizerr code 和既有 message key 包装。数据权限语义无变化，插件列表、详情、host service 授权和租户供应边界未放宽。缓存一致性在 2.1 中无语义变化，未新增缓存域；后续 6.x 统一发布入口时需记录权威源、一致性模型、触发点、跨实例同步、最大陈旧窗口、故障降级、可观测性和恢复路径。开发工具跨平台无影响，未修改脚本、Makefile、CI 或 linactl。测试策略为后端单元测试、静态治理测试、OpenSpec 严格校验；无 UI 或用户可观察页面变化，未触发 E2E。

## 2. lifecycle 与 migration 组件重建

- [x] 2.1 将旧`internal/lifecycle` SQL migration executor 迁移为`internal/migration`组件，保持动态 install/uninstall/migration 既有测试通过。

  记录：已新增`apps/lina-core/internal/service/plugin/internal/migration`，迁入`ExecuteManifestSQLFiles`、`ExecuteManifestMockSQLFilesInTx`、`ResolveSQLAssets`、`ResolvePluginSQLAssets`、`SQLAsset`、`MockSQLExecutionResult`和`MockDataLoadError`。`internal/lifecycle.Service`已移除 SQL executor 方法，只保留动态 install/uninstall 编排入口；`plugin.New()`和`internal/testutil.NewServices()`显式创建`migrationSvc := migration.New(catalogSvc, storeSvc)`，并传给根门面、`runtime.New`和`sourceupgrade.New`。`runtime`、`sourceupgrade`和 source lifecycle 调用点改为通过`migration.Service`执行 SQL 阶段，行为不改变；`internal/lifecycle`将在任务 2.2 重建真实编排 service 时再接收 migration 窄接口。

  DI 来源检查：`migration.Service`owner 为`internal/migration`；创建位置为`plugin.New()`和`internal/testutil.NewServices()`；传递路径为`plugin.serviceImpl.migrationSvc`、`runtime.serviceImpl.migrationSvc`和`sourceupgrade.serviceImpl.migrationSvc`。依赖为`catalog.Service`和`store.Service`窄能力，复用启动期同一 catalog/store 实例，不引入`Deps`/`Options`聚合结构体，不创建新的缓存或协调后端。

  数据库与 SQL 记录：未新增或修改 SQL 文件、schema、DAO/DO/Entity 生成文件；迁移执行仍使用原有`dao.SysPluginMigration.Transaction`、DO 写入、方言转译和账本记录逻辑。SQL 幂等性和数据分类语义无变化；本任务仅变更 Go owner。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/migration -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过。
- [x] 2.2 新建真实`internal/lifecycle`编排 service，构造函数逐项接收 catalog、store、runtime、integration、migration、dependency、i18n、cache publisher、topology 和必要 capability 窄接口。

  记录：已重建`apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle.go`主契约，`Service`继续保留当前已迁移的动态`Install`/`Uninstall`入口，新增后续 3.x/4.x 下沉所需的显式 orchestration seams：`RuntimeOrchestrator`、`IntegrationOrchestrator`、`DependencyResolver`、`I18nTranslator`、`CachePublisher`、`TopologyProvider`和`TenantProvisioningService`。`lifecycle.New(...)`现在逐项接收`catalog.Service`、`store.Service`、runtime、integration、`migration.Service`、dependency resolver、i18n、cache publisher、topology 和 tenant provisioning，不使用`Deps`/`Options`聚合结构体，也不保留旧 dynamic-only `ReconcileProvider`。

  DI 来源检查：`internal/lifecycle`owner 为 lifecycle orchestration；catalog/store/migration/i18n/topology/tenant provisioning 均复用`plugin.New()`启动期同一实例或启动期传入能力。runtime owner 为`internal/runtime`，integration owner 为`internal/integration`，dependency resolver owner 为`internal/dependency`并由组合根创建一个无状态实例。cache publisher 当前为组合根创建的`runtimeCacheChangeNotifierProvider`，后续 6.x 统一`publishPluginChange`时继续收敛到单一发布入口；本任务不创建新的缓存后端、协调器或本地-only 状态。

  影响判断：本任务只建立 lifecycle 编排 owner 和构造契约，不改变 HTTP API、DTO、SQL schema、SQL 文件、前端页面或`apps/lina-plugins/*`。`i18n`无新增文案或资源；数据权限语义无变化，写操作仍由根门面平台治理守卫保护；数据库/SQL 幂等性无影响；缓存语义无变化，仅把既有 cache publisher 显式注入 lifecycle 以便后续迁移。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/testutil -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run TestPluginLifecycleOrchestrationStaticBoundaries -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过。
- [x] 2.3 更新`plugin.New()`分阶段构造顺序，将 migration、lifecycle、runtime、integration 和 sourceupgrade 依赖接入新的 owner，不引入 setter 或 deps/options 聚合结构体。

  记录：`plugin.New()`已按 C 阶段拓扑调整为先构造 catalog/store/migration/frontend/openapi/sourceServices 等叶子和共享能力，再构造`runtimeSvc`、`integrationSvc`，随后以完整显式依赖构造`lifecycleSvc`，最后组装根`serviceImpl`和`sourceUpgradeSvc`。旧`lifecycleReconcilerProvider`已删除，动态 install/uninstall 不再通过构造后 provider 绑定 runtime；`internal/testutil.NewServices()`同步采用同样的显式构造顺序。`sourceupgrade`仍保持 D 阶段前的独立 owner，但继续显式接收 store、migration、runtime、integration、i18n 和 dependency validator，不反向持有根门面。

  DI 来源检查：`lifecycleSvc`创建位置为`plugin.New()`和`internal/testutil.NewServices()`；传递路径为`plugin.serviceImpl.lifecycleSvc`和`testutil.Services.Lifecycle`。runtime/integration/migration/sourceupgrade 均由组合根按 owner 构造后传递，未新增 setter、`Deps`、`Options`或服务定位器。启动期共享实例策略保持不变：cachecoord、i18n、locker、session、capability services、tenant/org 能力和 WASM runtime 继续复用组合根传入或创建的同一实例。

  影响判断：本任务修改启动装配和测试装配，不改变业务行为、接口契约、数据库结构或插件 wire 格式。`i18n`、数据权限、数据库/SQL、前端/E2E 无新增影响；缓存一致性不新增缓存域，后续 6.x 再统一实际发布入口。开发工具跨平台无影响，未修改脚本、Makefile、CI 或 linactl。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/testutil -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run TestPluginLifecycleOrchestrationStaticBoundaries -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过。

## 3. 安装、卸载与状态变更下沉

- [x] 3.1 迁移 Install 编排到`internal/lifecycle`，保留依赖检查、host service authorization、SQL migration、资源/菜单同步、hook 分发、runtime/cache 语义。

  记录：已新增`internal/lifecycle`完整 Install 编排入口，根门面`Install`/`install`收缩为平台治理守卫、`InstallOptions`到`lifecycle.InstallOptions`的值映射、framework version 值输入和 mock-data 错误包装。依赖检查、install mode 校验、source/dynamic 分支、动态 host service authorization、动态 BeforeInstall/AfterInstall、source BeforeInstall/AfterInstall、source SQL install、菜单/权限同步、resource refs 同步、hook 分发、runtime install reconciliation、enabled snapshot 刷新、runtime cache revision 发布和 observer install 通知已由`internal/lifecycle`承接。低层动态安装入口保留为`InstallDynamic`，供 runtime/integration 内部测试验证动态 reconciler install 语义；完整管理安装路径统一走 lifecycle `Install`。`migration.Service`新增`ExecuteManifestMockSQLFiles`事务包装，source/dynamic mock-data load 都复用该入口，根门面仅保留 typed mock-data error 到 bizerr 的包装。

  DI 来源检查：本任务没有新增构造函数级运行期依赖，也没有改变`lifecycle.New(...)`的 10 个显式依赖。Install 所需的 framework version 作为单次请求纯值由根门面读取后传入 lifecycle options，不创建新 service graph。observer registry 挂到 lifecycle 实例内部，根`RegisterLifecycleObserver`委托到`lifecycle.Service.RegisterLifecycleObserver`并保留未迁移启停/卸载流程的过渡根 observer registry；后续 3.2/3.3 迁移完成后再删除根过渡 registry。cache publisher 继续复用`runtimeCacheChangeNotifierProvider`和启动期同一 runtime revision controller；未新增缓存后端。

  影响判断：本任务只修改后端内部 service owner、测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，迁入 lifecycle 的 install-mode、dependency blocked、lifecycle veto 和 source install 错误码保持原 errorCode/messageKey/fallback，并在根包保留别名供现有调用和测试使用。数据权限语义无变化，HTTP 可达 install 仍先通过根门面平台治理守卫；动态 host service 授权仍只暴露本次确认的资源范围。缓存一致性语义无变化，install 成功后仍刷新本地 enabled snapshot 并发布`plugin-runtime`revision；统一发布入口仍留给 6.x。数据库/SQL schema 无变化；mock-data 事务 owner 从根门面收敛到`internal/migration`，仍使用同一账本事务和 typed rollback error。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestInstall|TestRegisterLifecycleObserver|TestLifecycleObserver|TestPluginLifecycleOrchestrationStaticBoundaries' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestListRuntimeStates|TestRuntime|TestDynamic' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run 'TestSyncSourcePluginMenus|Test.*Menu' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 3.2 迁移 Uninstall 编排到`internal/lifecycle`，抽取统一卸载收尾 helper，保留反向依赖、veto、force、purge storage、uninstall SQL 和回滚语义。

  记录：已新增`internal/lifecycle`完整 Uninstall 编排入口，根门面`Uninstall`收缩为平台治理守卫、读取`AllowForceUninstall`配置快照、映射`UninstallOptions`纯值并委托 lifecycle。反向依赖检查、source/dynamic 分支、source BeforeUninstall/AfterUninstall、dynamic active manifest 回退、force 缺失 artifact 处理、purge storage 策略、runtime/integration 清理、resource refs 清理、uninstall SQL、enabled snapshot 刷新、runtime cache revision 发布和 observer uninstall 通知已由`internal/lifecycle`承接。低层动态卸载入口重命名为`UninstallDynamic`，仅保留给 runtime/integration 内部测试验证动态 reconciler uninstall 语义；完整管理卸载路径统一走 lifecycle `Uninstall`。根包 source lifecycle 文件已移除本批迁走后不再需要的 source uninstall DAO/DO 事务 helper，静态边界 allowlist 同步删除`plugin_lifecycle_source.go`例外。

  DI 来源检查：本任务没有新增构造函数级运行期依赖，也没有改变`lifecycle.New(...)`的 10 个显式依赖。`AllowForceUninstall`作为单次请求纯值由根门面从既有 config service 快照读取后传入 lifecycle options，不创建新 service graph，不引入 setter、`Deps`、`Options`或接口型依赖聚合。cache publisher 继续复用`runtimeCacheChangeNotifierProvider`和启动期同一 runtime revision controller；未新增缓存后端、协调器或本地-only 状态。

  影响判断：本任务只修改后端内部 service owner、测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，反向依赖、force disabled、dynamic artifact missing、uninstall execution failed 和 source registry after-uninstall missing 错误码迁入 lifecycle owner 后在根包保留别名，既有 message key/fallback 语义不变。数据权限语义无变化，HTTP 可达 uninstall 仍先通过根门面平台治理守卫；反向依赖和 lifecycle veto 均在副作用前执行。缓存一致性语义无变化，uninstall 成功后仍刷新本地 enabled snapshot 并发布`plugin-runtime`revision；统一发布入口仍留给 6.x。数据库/SQL schema 无变化；source/dynamic uninstall SQL 仍由既有 migration 组件执行，不新增迁移文件或 DAO/DO/Entity 生成物。测试策略为后端单元测试、静态治理测试、OpenSpec 严格校验和 diff 检查；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestInstall|TestUninstall|TestPluginLifecycleOrchestrationStaticBoundaries|TestRegisterLifecycleObserver|TestLifecycleObserver|TestSourceLifecycle|TestDynamic' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestListRuntimeStates|TestRuntime|TestDynamic' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run 'TestSyncSourcePluginMenus|Test.*Menu' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 3.3 迁移 UpdateStatus/Enable/Disable 编排到`internal/lifecycle`，将 source/dynamic 与 enable/disable 四象限拆成命名窄函数。

  记录：已新增`internal/lifecycle/lifecycle_status.go`，由 lifecycle 承接`UpdateStatus`完整编排。根门面`UpdateStatus`、`Enable`、`Disable`仍保留平台治理守卫，私有`updateStatus`收缩为 4.2 自动启用路径可复用的轻量委托。lifecycle 内将状态变更拆为`enableDynamicPlugin`、`disableDynamicPlugin`、`enableSourcePlugin`、`disableSourcePlugin`四个命名分支，保留动态启用 artifact 检查、启用依赖检查、动态 host service authorization 持久化、动态 BeforeDisable/AfterDisable、source BeforeDisable/AfterDisable、runtime reconcile、source registry status 写入、enabled snapshot 刷新、runtime cache revision 发布和 observer enable/disable 通知语义。状态错误码`PLUGIN_STATUS_INVALID`与`PLUGIN_NOT_INSTALLED`迁入 lifecycle owner，根包保留别名。

  DI 来源检查：本任务没有新增构造函数级运行期依赖，也没有改变`lifecycle.New(...)`的 10 个显式依赖。`UpdateStatusOptions`只携带动态启用授权快照和 framework version 纯值，不承载接口型依赖，不引入 setter、`Deps`或接口聚合结构体。状态变更复用 lifecycle 已有 catalog/store/runtime/integration/dependency/i18n/cache publisher/topology/tenant provisioning 显式依赖；cache publisher 继续复用启动期同一 runtime revision controller，未新增缓存后端、协调器或本地-only 状态。

  影响判断：本任务只修改后端内部 service owner、测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，状态错误码迁入 lifecycle 后 message key/fallback 不变。数据权限语义无变化，HTTP 可达状态变更仍先通过根门面平台治理守卫；启用依赖检查和禁用 lifecycle veto 均在副作用前执行。缓存一致性语义无变化，状态变更成功后仍刷新 enabled snapshot 并发布`plugin-runtime`revision；统一发布入口仍留给 6.x。数据库/SQL schema 无变化；source 状态写入仍通过 store 组件，未新增迁移文件或 DAO/DO/Entity 生成物。测试策略为后端单元测试、静态治理测试、OpenSpec 严格校验和 diff 检查；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecycleBeforeDisableBlocksDisable|TestEnableWithAuthorizationAppliesConfirmedHostServiceSnapshot|TestPluginRuntime|TestSourceProviderAvailabilityFollowsEnabledSnapshot|TestBootstrapAutoEnable|TestRegisterLifecycleObserver|TestLifecycleObserver|TestPluginLifecycleOrchestrationStaticBoundaries' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestInstall|TestUninstall|TestUpdate|TestEnable|TestDisable|TestSourceLifecycle|TestDynamicLifecycle|TestPluginLifecycleOrchestrationStaticBoundaries|TestRegisterLifecycleObserver|TestLifecycleObserver|TestBootstrapAutoEnable' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestListRuntimeStates|TestRuntime|TestDynamic' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run 'TestSyncSourcePluginMenus|Test.*Menu' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 3.4 统一 source lifecycle decision 与 dynamic lifecycle decision 的 veto 汇总 helper，删除成对重复计数和本地化逻辑。

  记录：已新增`internal/lifecycle/lifecycle_veto.go`，通过内部`lifecycleVetoDecision`投影统一 source `pluginhost.LifecycleDecision`和 dynamic `runtime.DynamicLifecycleDecision`的 veto reason 汇总、本地化、默认 reason 和插件前缀判断。`summarizeLifecycleVetoReasons`、`summarizeLocalizedLifecycleVetoReasons`、`summarizeDynamicLifecycleVetoReasons`和`summarizeLocalizedDynamicLifecycleVetoReasons`保留原函数入口作为薄适配，调用点不需要理解 source/dynamic 决策结构差异；原成对`countLifecycleVetoes`/`countDynamicLifecycleVetoes`已删除，输出格式保持不变。

  DI 来源检查：本任务不新增运行期依赖，不改变构造函数和`lifecycle.New(...)`显式依赖数量。通用 formatter 只接收纯数据切片和可选翻译函数；翻译仍由既有 lifecycle `i18nSvc`提供，不创建新 service graph，不引入 setter、`Deps`或接口聚合结构体。

  影响判断：本任务只修改 lifecycle 内部 helper 和 OpenSpec 任务记录；不修改 HTTP API、DTO、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，既有 message key、本地化 fallback 和 reason 拼接格式不变。数据权限、数据库/SQL 和缓存一致性无语义变化；本任务不新增缓存域或数据访问路径。测试策略为后端单元测试、OpenSpec 严格校验和 diff 检查；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecyclePreconditionLocalizesReasonParams|TestSourceLifecycleBeforeInstall|TestSourceLifecycleBeforeDisable|TestSourceLifecycleBeforeUninstall|TestDynamicLifecycle|TestUninstallForce|TestEnable|TestDisable|TestPluginLifecycleOrchestrationStaticBoundaries' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。

## 4. 源码生命周期、自动启用与租户钩子下沉

- [x] 4.1 迁移源码插件安装、卸载和回滚事务编排到`internal/lifecycle`，根门面不再直接触达`dao.SysPlugin`。

  记录：源码插件 install、uninstall、rollback 事务编排已归属`apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle_source_install.go`。根门面`plugin_lifecycle.go`和`plugin_lifecycle_source.go`不再导入`dao`、`do`或`entity`，source install/uninstall/rollback 不再直接触达`dao.SysPlugin`；根门面仅保留平台治理守卫、选项纯值映射、生命周期 callback 过渡入口和后续 4.3 租户钩子仍需迁移的代码。`plugin_boundary_test.go`的根门面生成模型 allowlist 已不包含 source lifecycle 文件，只保留 route/projection 过渡项。

  函数长度治理：`installSourcePlugin`、`uninstallSourcePlugin`已拆成 source install/uninstall 目标解析、可回滚治理副作用、storage purge、hook 分发等命名窄函数；本批 source lifecycle 迁移函数均小于 60 行。该拆分不改变原有步骤顺序：install SQL 失败前不触发 rollback，菜单/registry/release/resource refs 等可回滚副作用失败时仍执行`rollbackSourcePluginInstall`，install/uninstall hook 分发失败仍按原语义直接返回。

  DI 来源检查：本任务不新增构造函数级运行期依赖，不改变`lifecycle.New(...)`的 10 个显式依赖。source install/uninstall/rollback 继续复用 lifecycle 已有的 catalog、store、migration、integration、i18n 和 cache publisher 等启动期共享实例；未引入 setter、`Deps`、`Options`或接口型依赖聚合。`dao.SysPlugin`和`dao.SysPluginResourceRef`访问仅保留在 lifecycle owner 内部的 source stable state 与 resource ref cleanup 边界，后续 store/projection 深化不属于本任务范围。

  影响判断：本任务只修改后端内部 lifecycle 编排、OpenSpec 任务记录和静态边界；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，源码生命周期错误码、message key 和 fallback 不变。数据权限语义无变化，HTTP 可达 install/uninstall 仍先通过根门面平台治理守卫；source lifecycle veto 仍在副作用前执行。缓存一致性语义无变化，source install/uninstall 成功后的 enabled snapshot 刷新和`plugin-runtime`revision 发布仍由 lifecycle 调用既有 cache publisher；统一发布入口仍留给 6.x。数据库/SQL schema 无变化，source install/uninstall SQL 仍由`internal/migration`执行；本任务不新增迁移文件或 DAO/DO/Entity 生成物。无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecycle|TestInstall|TestUninstall|TestPluginLifecycleOrchestrationStaticBoundaries' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 4.2 迁移`BootstrapAutoEnable`和启动自动启用逻辑到`internal/lifecycle`，将 startup auto-enable 和 mock-data 控制从 context key 改为显式 options 或内部启动入口。

  记录：启动自动启用状态机已归属`apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle_auto_enable.go`。根门面`BootstrapAutoEnable`保留 5.x 投影收敛前的`syncAndList(ctx)`启动同步入口，随后只读取`configSvc.GetPluginAutoEnableEntries(ctx)`、映射为 lifecycle 纯值`AutoEnableEntry`并委托`lifecycleSvc.BootstrapAutoEnable`。lifecycle 内承接单项 auto-enable 分派、启动依赖检查、source/dynamic install/enable、动态 host service 授权快照复用、单机/集群 primary 等待和超时错误格式化；启动成功后仍刷新 enabled snapshot。

  显式控制参数：`plugin.autoEnable`中的 mock-data opt-in 映射为`lifecycle.AutoEnableEntry.WithMockData`，source startup-only 语义通过`lifecycle.InstallOptions.StartupAutoEnable`传入 source lifecycle callback。根包已删除`sourceLifecycleInstallOptionsContextKey`、`withSourceLifecycleInstallOptions`和`sourceLifecycleStartupAutoEnable`，普通 HTTP install 无法通过伪造 context key 获得 startup-only 行为。`catalog.WithInstallMockData`仍仅作为 install 内部向 source/dynamic SQL 执行层传递单次 mock-data opt-in 的过渡输入，后续 4.4 统一清理和记录剩余业务 context key。

  DI 来源检查：本任务没有新增构造函数级运行期依赖。`lifecycle.TopologyProvider`补充`IsClusterModeEnabled()`窄方法，由`plugin.New()`中已存在的`Topology`通过`lifecycleTopologyAdapter`适配传入；owner 为启动装配层的 topology，创建位置仍是`plugin.New()`和`internal/testutil.NewServices()`，传递路径为`plugin.Topology -> lifecycleTopologyAdapter -> lifecycle.New(...)`。该适配复用启动期同一 topology 实例，不创建新 service graph、共享后端或本地-only 缓存。`BootstrapAutoEnableOptions`只携带 entries 和 framework version 纯值，不承载接口型依赖。

  影响判断：本任务只修改后端内部 lifecycle 编排、根门面委托、错误码 owner、测试装配和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，startup auto-enable 错误码迁入 lifecycle owner 后在根包保留别名，既有 errorCode/messageKey/fallback 不变。数据权限语义无变化，auto-enable 属于宿主启动治理路径，不新增用户数据读取接口；普通 HTTP 生命周期写入仍由根门面平台治理守卫保护。缓存一致性语义无变化，仍复用`plugin-runtime`revision 相关 cache publisher 和 enabled snapshot 刷新，不新增缓存域；统一发布入口仍留给 6.x。数据库/SQL schema 无变化，install/mock-data SQL 仍由既有 migration 组件执行；本任务不新增迁移文件或 DAO/DO/Entity 生成物。无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestBootstrapAutoEnable|TestReconcileAutoEnabledTenantPlugins|TestSourceLifecycleBeforeInstallRejectsManualWhenStartupAutoEnableRequired|TestSourceLifecycleBeforeInstallReceivesStartupAutoEnableFlag|TestPluginLifecycleOrchestrationStaticBoundaries|TestBootstrapAutoEnableBlocksUninstalledDependency' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 4.3 迁移`ReconcileAutoEnabledTenantPlugins`、租户删除/禁用 precondition 和 notification 到`internal/lifecycle`，复用统一 veto 汇总。

  记录：已新增`apps/lina-core/internal/service/plugin/internal/lifecycle/lifecycle_tenant.go`，由 lifecycle 承接`ReconcileAutoEnabledTenantPlugins`、租户插件禁用前置检查、租户插件禁用通知、租户删除前置检查和租户删除通知。根门面`plugin_auto_enable.go`仅负责读取`plugin.autoEnable`配置并映射为 lifecycle 纯值输入；根门面`plugin_lifecycle.go`中的四个租户生命周期公共方法已收缩为一行委托，不再复制 source/dynamic 扫描、租户 veto 汇总或 notification 逻辑。source tenant lifecycle 继续复用统一`lifecycle_veto.go`汇总 helper；dynamic tenant lifecycle 统一复用 lifecycle owner 内的`dynamicLifecycleError`和`runtime.DynamicLifecycleDecision`格式化路径。

  DI 来源检查：本任务没有新增构造函数级运行期依赖，也没有改变`lifecycle.New(...)`的显式依赖列表。租户供应策略继续复用任务 2.2 已注入的`TenantProvisioningService`，owner 为租户能力 provider，创建位置和传递路径仍为组合根`plugin.New()`或`internal/testutil.NewServices()`传入 lifecycle；source/dynamic tenant hook 复用 lifecycle 已有 catalog、store、runtime、integration、i18n、tenant provisioning 等启动期共享实例，不引入 setter、`Deps`、`Options`或接口型依赖聚合。

  影响判断：本任务只修改后端内部 lifecycle owner、根门面委托、错误码 owner 和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源，租户 auto-enable provisioning 错误码迁入 lifecycle owner 后在根包保留别名，既有 errorCode/messageKey/fallback 不变；source/dynamic veto reason 继续走既有本地化与 fallback。数据权限语义无变化，租户 capability 回调路径保持原边界，不新增 HTTP/API 读写接口，也不扩大租户/插件可见性。缓存一致性语义无变化，租户供应策略对账后仍刷新 enabled snapshot，未新增缓存域；统一插件变化发布入口仍留给 6.x。数据库/SQL schema 无变化，不新增迁移文件或 DAO/DO/Entity 生成物。无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestReconcileAutoEnabledTenantPlugins|TestTenant|TestDynamicLifecycleBeforeTenant|TestPluginLifecycleOrchestrationStaticBoundaries' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 4.4 清理本变更迁移后不再需要的业务控制流 context key，仅保留启动只读大快照所需 context 输入，并记录保留理由。

  记录：已删除`internal/catalog/install_context.go`中的`installMockDataContextKey`、`WithInstallMockData`和`ShouldInstallMockData`，并移除根门面`withInstallMockData`/`shouldInstallMockData`薄包装。`InstallOptions.InstallMockData`现在作为显式纯值传入 lifecycle；source install 通过`SourceInstallOptions.InstallMockData`决定是否加载 mock-data，dynamic install 通过`runtime.DynamicReconcileOptions.InstallMockData`传入 runtime reconciler。动态 runtime 的后台 reconcile、卸载和状态切换路径均使用零值 options，避免普通业务控制语义继续依赖 context key。

  同步清理：根包旧 install 编排迁移后遗留的`dependencyInstallContextKey`、`dependencyInstallContext`、`dependencyContextFrom`和根包私有`prepareInstallDependencies`已删除；依赖检查 owner 已在 lifecycle 中显式执行，不再通过请求 context 记录递归安装状态。`dependencySnapshotCacheContextKey`仍保留到 5.x：它只服务列表/依赖投影中的请求内只读快照复用，属于投影 builder 收敛对象，不在 4.4 提前拆除。

  保留理由：`store.WithStartupDataSnapshot`和`integration.WithStartupDataSnapshot`仍保留，它们承载启动一致性与列表同步期间的只读大快照，并在写入后有对应刷新逻辑；这是设计中允许保留的 startup orchestration 输入。`management.WithManifestSnapshot`和`WithDependencySnapshotCache`仍保留到 5.x 统一投影 builder，当前用于避免列表/摘要/详情/依赖检查重复扫描。`integration.WithAuthoritativeEnablement`是只读查询一致性 marker，不改变生命周期写入语义；WASM host call 和 datahost audit context 不属于本次 lifecycle 业务控制流。

  DI 来源检查：本任务没有新增运行期服务依赖。新增的`runtime.DynamicReconcileOptions`是单次调用纯值结构，只携带 mock-data opt-in 布尔值，不承载接口型依赖，不引入 setter、`Deps`或服务定位器。runtime 仍复用启动期同一 store、migration、integration、cache publisher 和 reconciler revision controller。

  影响判断：本任务只修改后端内部显式参数传递、删除业务控制 context key、测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或文案。数据权限语义无变化，未新增读写接口或租户/插件可见性路径。缓存一致性语义无变化，动态 install 仍通过既有 runtime/cache publisher 发布，未新增缓存域；统一插件变化发布入口仍留给 6.x。数据库/SQL schema 无变化，不新增迁移文件或 DAO/DO/Entity 生成物。无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -run TestNonExistent -count=0`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run TestNonExistent -count=0`通过；`cd apps/lina-core && go test ./internal/service/plugin -run TestNonExistent -count=0`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestBootstrapAutoEnableHonorsPerEntryMockDataOptIn|TestInstall|TestPluginLifecycleOrchestrationStaticBoundaries|TestCheckPluginDependencies|TestBootstrapAutoEnable' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/lifecycle -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestDynamic|TestRuntime|TestReconcile|TestInstall|TestUninstall' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。

## 5. 列表投影与管理读模型收敛

- [x] 5.1 设计并实现统一插件投影 builder，覆盖 list、summary、detail 和 dependency snapshot mode，集中 manifest/store/runtime/dependency/i18n/tenant provisioning 批量装配。

  记录：已新增`apps/lina-core/internal/service/plugin/plugin_projection.go`作为 C 阶段统一投影 builder。`buildPluginProjection`通过`projectionMode`覆盖`list`、`summary`、`detail`和`dependency_snapshot`四种模式，统一完成 runtime cache freshness、manifest snapshot 获取、store startup snapshot、manifest context snapshot、dependency snapshot context、registry 批量读取或同步、runtime registry-only 动态插件追加、i18n 展示字段和租户供应策略字段投影。`plugin_list.go`中的`syncAndList`、`buildManagementList`、`buildManagementSummaryList`和`buildManagementDetail`已收缩为 builder 委托；`plugin_dependency.go`中的`buildDependencySnapshots`已改为 dependency snapshot mode 委托，实际快照组装集中到`buildDependencySnapshotsForProjection`，避免列表详情路径各自维护 manifest/store/dependency 组装入口。

  性能边界：list/summary/detail/dependency snapshot 均由同一入口控制 manifest 扫描；只读模式若 context 已有`management.WithManifestSnapshot`则复用，不重复扫描；sync 模式始终重新扫描当前 manifest，避免启动或管理同步复用陈旧快照。list/full 和 summary 只执行一次`ListAllRegistries`批量读取，并通过内存 map 合并 manifest 与 registry；sync 模式按 manifest 范围执行既有`SyncManifest`写入语义并保留 integration startup snapshot。summary mode 继续省略 dependency check、host service detail 和 route payload；detail mode 仅对目标插件执行单个 registry 查找并复用同一 manifest/dependency context。依赖快照仍通过请求内`dependencySnapshotCacheContextKey`缓存一次，列表 detail dependency check 不会为每个插件重复扫描 manifest 或 registry。

  静态治理：`plugin_boundary_test.go`已补充投影 builder 边界断言，要求`syncAndList`、`buildManagementList`、`buildManagementSummaryList`、`buildManagementDetail`和`buildDependencySnapshots`必须调用`buildPluginProjection`，且列表四条入口不得重新拥有独立`ScanManifests`流水线。5.2 后续可继续在此基础上细化缓存键和字段完整性测试，5.3 再收敛`internal/management`剩余缓存/深拷贝工具箱。

  DI 来源检查：本任务未新增运行期 service 依赖、构造函数参数、共享后端、setter、`Deps`或`Options`聚合结构体。builder 归属根插件门面包内部私有投影层，复用`serviceImpl`已有 catalog、store、runtime、integration、dependency、i18n、cache revision controller 和 tenant provisioning registry 字段来源；无新增服务图或本地-only 缓存域。

  影响判断：本任务只修改后端内部投影 owner、静态治理测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或文案，插件名称/描述仍由 runtime 投影内既有`localizePluginMetadata`处理，缓存键仍包含 locale 与 bundle version。数据权限语义无变化，插件管理列表/详情仍属于平台插件治理投影，不新增用户数据读取或租户数据暴露路径；租户供应策略仅沿用 registry 字段投影。缓存一致性语义无变化，仍复用既有`plugin-runtime`revision、management list cache key 和显式 invalidation，不新增缓存域；统一发布入口仍留给 6.x。数据库/SQL schema 无变化，不新增迁移文件或 DAO/DO/Entity 生成物。无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestManagementListCacheAvoidsRepeatedManifestScans|TestPrewarmManagementListPopulatesCache|TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary|TestGetReturnsStableNotFoundBizerr|TestDependencySnapshotCacheReusesCatalogSnapshot' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -run 'TestManagementList|TestList|TestCheckPluginDependencies|TestDependencySnapshot|TestPluginLifecycleOrchestrationStaticBoundaries|TestGet|TestSyncAndList' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 5.2 将`plugin_list.go`四条同构流水线迁移到统一 builder，保留管理弹窗依赖字段和缓存键语义。

  记录：5.1 已完成`plugin_list.go`中`syncAndList`、`buildManagementList`、`buildManagementSummaryList`和`buildManagementDetail`四个入口到`buildPluginProjection`的委托迁移；本任务在该基础上补充语义固化。`plugin_boundary_test.go`继续阻断四个入口恢复独立`ScanManifests`流水线。`plugin_list_test.go`已扩展`TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary`，明确区分三种投影模式：`List`摘要缓存路径保留`AuthorizationRequired`和基础状态字段，但不携带`DependencyCheck`、`RequestedHostServices`、`AuthorizedHostServices`和`DeclaredRoutes`；`ReadOnlyList`完整管理读模型继续携带管理弹窗依赖字段，包括依赖检查、host service 申请和动态路由声明；`Get`详情路径继续只投影目标插件，并在动态插件安装后从 release snapshot 回退读取`AuthorizedHostServices`和`DeclaredRoutes`。该测试同时覆盖未安装 staging artifact 和已安装 release snapshot 两类管理弹窗字段来源。

  缓存键语义：已扩展`TestManagementListCacheIsLocaleScoped`，断言`managementListCacheKey`仍由 locale、runtime bundle version 和`plugin-runtime`runtime revision 组成，且默认语言与英文请求形成不同缓存分区。`managementSummaryList`仍先按当前 key 构建缓存，构建后再次读取 latest key 并在 revision 或 bundle version 变化时把结果存入最新 key，保留原缓存热身和运行期 revision 语义。

  性能边界：四个列表入口不再维护各自 manifest/store/runtime 组装流水线；summary、full list 和 detail 通过同一 builder 复用 manifest snapshot、store startup snapshot 和 dependency snapshot context。summary 模式不做逐插件 dependency check；full list/detail 模式在同一请求上下文内复用 dependency snapshot cache，管理弹窗字段不会触发固定外的逐插件 manifest 扫描。5.4 后续继续补充更完整的批量装配和扫描次数覆盖。

  DI 来源检查：本任务只补充后端单元测试和任务记录，不新增运行期 service 依赖、构造函数参数、共享后端、setter、`Deps`或`Options`聚合结构体。统一 builder 仍复用`serviceImpl`已有 catalog、store、runtime、integration、dependency、i18n、cache revision controller 和 tenant provisioning registry 字段来源。

  影响判断：本任务不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或用户可见文案；缓存一致性无新增缓存域，继续复用`plugin-runtime`revision、management list cache key 和显式 invalidation；数据权限语义无变化，插件管理投影仍属于平台插件治理读模型，不新增用户数据读取或租户数据暴露路径；数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestManagementListCacheIsLocaleScoped|TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过。
- [x] 5.3 收敛`internal/management`私有工具箱职责；能并入投影所在层的 helper 直接合并，必须保留的缓存深拷贝集中到单文件维护且不使用 JSON/gob 往返。

  记录：已将`apps/lina-core/internal/service/plugin/internal/management/management.go`收敛为管理读模型类型、`ListCache`并发构建和`ListCacheKey`职责。分页默认值、分页切片和插件类型匹配 helper 已迁回投影调用所在的`plugin_list.go`私有函数；registry map 和投影排序 helper 已迁回`plugin_projection.go`私有函数，避免`internal/management`继续作为跨层工具箱暴露。必须保留的管理列表缓存深拷贝集中到`management_cache_clone.go`，manifest 请求内只读快照集中到`management_manifest_snapshot.go`；缓存深拷贝继续显式复制 host service、route、string map、dependency projection 和 last upgrade failure，不使用 JSON/gob 往返。静态检索`rg "json\\.|gob\\." apps/lina-core/internal/service/plugin/internal/management -g'*.go'`无命中。

  导出面与职责边界：`internal/management`当前导出面仅保留`PluginItem`、`ListOutput`、`ListInput`、`ListCache`、`NewListCache`、`ListCacheKey`、manifest snapshot helpers 和缓存深拷贝 helpers，均被根投影、依赖快照、lifecycle 只读快照或缓存路径使用；本任务未新增跨模块契约、运行期服务依赖、构造函数参数、共享后端、setter、`Deps`或`Options`聚合结构体。

  影响判断：本任务只调整后端内部 helper 归属和文件组织，不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或用户可见文案；数据权限语义无变化，插件管理投影仍是平台插件治理读模型；缓存一致性语义无变化，仍复用`plugin-runtime`revision、management list cache key 和显式 invalidation，不新增缓存域；数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/management -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 5.4 补充或更新列表/详情/摘要/依赖投影测试，验证字段完整性、缓存命中、批量装配和不存在固定外的逐插件重复扫描。

  记录：已在`apps/lina-core/internal/service/plugin/plugin_list_test.go`新增`TestReadOnlyListProjectionBatchesDependencyChecks`，构造两个 source 插件和一条 hard dependency，验证完整管理投影会给多个列表行附加`DependencyCheck`，目标插件能暴露缺失依赖 blocker，同时通过`startupstats`断言整个`ReadOnlyList`只执行一次 manifest scan、一次 store startup snapshot，且不会构建 startup-only integration snapshot。该测试与既有`TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary`、`TestManagementListCacheAvoidsRepeatedManifestScans`和`TestDependencySnapshotCacheReusesCatalogSnapshot`共同覆盖：summary 与 full/detail 字段边界、缓存命中、dependency snapshot clone、防止列表路径恢复固定外逐插件重复扫描。

  影响判断：本任务只新增后端单元测试和 OpenSpec 任务记录，不修改生产代码、HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或用户可见文案；数据权限语义无变化；缓存一致性无新增缓存域，测试只验证现有`plugin-runtime`/management projection 快照复用行为；数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestReadOnlyListProjectionBatchesDependencyChecks|TestListPaginatesAndKeepsGovernanceDetailsOutOfSummary|TestManagementListCacheAvoidsRepeatedManifestScans|TestDependencySnapshotCacheReusesCatalogSnapshot' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。

## 6. 缓存失效入口收敛

- [x] 6.1 新增统一`publishPluginChange(ctx, pluginID, pluginType, reason)`或等价入口，封装`plugin-runtime`revision 发布、管理读模型失效和 runtime/frontend/i18n/WASM 派生缓存失效。

  记录：已在`apps/lina-core/internal/service/plugin/plugin_runtime_cache.go`新增根门面内部`publishPluginChange(ctx, pluginChangePublishInput)`，统一封装本节点派生缓存失效、管理读模型失效和`plugin-runtime`revision 发布。空`pluginID`表示全量插件治理变化，失效全部 runtime frontend bundle、WASM cache 和 source/dynamic i18n runtime bundle；带`pluginID`和`pluginType`表示插件级变化，失效对应 frontend bundle、source/dynamic i18n scope，动态插件额外失效 WASM cache。旧`markRuntimeCacheChanged`和`syncEnabledSnapshotAndPublishRuntimeChange`已委托到统一入口，`UploadDynamicPackage`、`UpdateTenantProvisioningPolicy`、source upgrade 失败、source upgrade 成功、dynamic runtime upgrade 失败和成功路径已接入统一入口；列表同步、安装、卸载、启用、禁用、启动自动启用等经 lifecycle cache publisher 的剩余路径将在 6.2 继续迁移或收紧调用约束。

  缓存一致性记录：权威数据源仍是插件 registry/release、manifest artifact/source tree 和 runtime revision controller；一致性模型仍是本节点同步失效加`plugin-runtime`revision 跨实例广播；失效触发点为治理写入成功后；跨实例同步继续复用既有`revisionctrl.Controller`和`cachecoord`协调，不新增本地-only 缓存域；最大陈旧窗口仍由各读路径`ensureRuntimeCacheFresh`和 revision 观察窗口约束；缓存后端不可用时保持原错误返回或日志降级语义；可观测性继续通过 reason 和 revision debug 日志保留。6.3 将在 6.2 迁移完成后补完整最终记录。

  静态治理：已扩展`TestPluginLifecycleOrchestrationStaticBoundaries`，新增`checkPluginChangePublisherBoundary`，要求`publishPluginChange`同时调用派生缓存失效、管理读模型失效和 shared revision 发布，并要求`markRuntimeCacheChanged`、`syncEnabledSnapshotAndPublishRuntimeChange`委托统一入口。

  DI 来源检查：本任务未新增运行期 service 依赖、构造函数参数、共享后端、setter、`Deps`或`Options`聚合结构体。统一入口复用`serviceImpl`既有`runtimeCacheRevisionCtrl`、`managementListCache`、`frontendSvc`、`i18nSvc`和`wasmRuntime`字段。

  影响判断：本任务只修改后端内部缓存发布入口、相关 root 写路径、静态边界测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或用户可见文案，仅复用既有 runtime bundle cache scope；数据权限语义无变化；数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestPluginLifecycleOrchestrationStaticBoundaries|TestRuntimeCacheChangeInvalidatesManagementList|TestManagementListCacheIsLocaleScoped|TestUpdateTenantProvisioningPolicy|TestListMarksInstalledDynamicPluginWithHigherArtifactPendingUpgrade|TestListMarksInstalledDynamicPluginWithFailedTargetReleaseUpgradeFailed' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 6.2 将同步、上传、安装、卸载、启用、禁用、源码升级、动态升级、租户供应策略和启动自动启用成功路径迁移到统一发布入口。

  记录：已在根门面保留兼容`MarkRuntimeCacheChanged(ctx, reason)` wrapper，并新增作用域化`PublishPluginChange(ctx, pluginID, pluginType, reason)`，两者均委托`publishPluginChange`。`runtimeCacheChangeNotifierProvider`同步转发新方法。`plugin_list.go`中的`SyncSourcePlugins`、`SyncSourcePluginsStrict`和`SyncAndList`已从`markRuntimeCacheChanged`加手工`managementListCache.Invalidate()`改为直接调用`publishPluginChange`，管理读模型失效统一由发布入口负责。`internal/lifecycle`的 cache publisher 已扩展为`PublishPluginChange`，安装、卸载、启用、禁用和启动自动启用成功路径通过`syncEnabledSnapshotAndPublishRuntimeChange`读取 registry 类型后携带`pluginID/pluginType/reason`发布。`internal/runtime`的 cache notifier 已扩展为`PublishPluginChange`，动态安装、卸载、force orphan uninstall、升级、启停、同版本 refresh 和 artifact missing reconciliation 均通过`notifyRuntimeCacheChanged(ctx, manifest, reason)`携带动态插件作用域发布。6.1 已迁移的上传、源码升级、动态升级成功/失败和租户供应策略路径继续保持在统一入口内。

  静态治理：已扩展`TestPluginLifecycleOrchestrationStaticBoundaries`，要求根`MarkRuntimeCacheChanged`、`markRuntimeCacheChanged`、`PublishPluginChange`和`syncEnabledSnapshotAndPublishRuntimeChange`必须委托`publishPluginChange`；要求同步列表入口必须直接调用`publishPluginChange`且不得自行调用旧发布或手工 cache invalidation；要求 lifecycle/runtime 的内部发布 helper 必须使用`PublishPluginChange`而不是旧`MarkRuntimeCacheChanged`。

  DI 来源检查：本任务未新增运行期 service 依赖、构造函数参数、共享后端、setter、`Deps`或`Options`聚合结构体。变更只扩展既有 cache publisher/cache notifier seam 的方法集；owner 仍为根`serviceImpl`的`publishPluginChange`，创建位置仍为`plugin.New()`中的`runtimeCacheChangeNotifierProvider`，绑定路径仍为`cacheChangeNotifier.BindService(service)`，runtime/lifecycle 继续复用启动期同一 provider 和同一`runtimeCacheRevisionCtrl`、`managementListCache`、`frontendSvc`、`i18nSvc`、`wasmRuntime`。

  影响判断：本任务只修改后端内部发布 seam、同步/lifecycle/runtime 调用点、测试替身、静态治理测试和 OpenSpec 任务记录；不修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无新增资源或用户可见文案，仅按插件作用域复用既有 runtime bundle cache invalidation。数据权限语义无变化，平台治理守卫和租户供应边界未改变。缓存一致性不新增缓存域，仍复用`plugin-runtime`revision 和`cachecoord`；本任务只让成功写路径统一通过同一发布入口。数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。

  验证：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestPluginLifecycleOrchestrationStaticBoundaries|TestRuntimeCacheChangeInvalidatesManagementList|TestManagementListCacheIsLocaleScoped|TestInstall|TestUninstall|TestUpdateStatus|TestBootstrapAutoEnable|TestUpdateTenantProvisioningPolicy|TestSyncAndList' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestDynamic|TestRuntime|TestReconcile|TestInstall|TestUninstall|TestRuntimeWiring' -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。
- [x] 6.3 补充缓存一致性记录：权威源、一致性模型、失效触发点、跨实例同步、最大陈旧窗口、故障降级、可观测性和恢复路径未因入口收敛改变。

  记录：6.1 和 6.2 完成后，插件治理变化的统一发布入口为根门面内部`publishPluginChange(ctx, pluginChangePublishInput)`；旧`MarkRuntimeCacheChanged(ctx, reason)`、`markRuntimeCacheChanged(ctx, reason)`、根`syncEnabledSnapshotAndPublishRuntimeChange(ctx, pluginID, reason)`、lifecycle cache publisher 和 runtime cache notifier 均只作为该入口的调用 seam 或兼容 wrapper。权威数据源仍是`store`持久化的插件 registry/release/runtime upgrade state、source tree 或动态 artifact 解析得到的 desired manifest、runtime active release archive、integration enabled snapshot 的可重建投影，以及`revisionctrl.Controller`维护的`plugin-runtime`共享 revision；本变更不创建新的缓存权威源。

  一致性模型：单节点内，成功写入或治理状态收敛后同步失效本节点 frontend bundle、WASM runtime cache、runtime i18n bundle scope 和`managementListCache`，再通过`runtimeCacheRevisionCtrl.MarkChanged(ctx)`发布共享 revision。集群模式下，其他节点通过`EnsureFresh(ctx)`观察共享 revision，并在 refresh callback 中重建或失效 integration enabled snapshot、frontend bundle、management read model、WASM cache 和 source/dynamic runtime i18n bundle。单机模式下，`revisionctrl.Controller`保留本地 revision 和同步失效语义，不强制依赖分布式协调组件。

  失效触发点：列表同步、源码插件同步、动态包上传、安装、卸载、启用、禁用、source upgrade 成功或失败、dynamic runtime upgrade 成功或失败、runtime reconciler install/uninstall/status/refresh/artifact missing、租户供应策略更新和启动自动启用成功路径，均在对应 store/runtime/integration 写入成功或状态收敛后发布。入口不在事务可能回滚前发布不可恢复事件；mock-data 可选阶段失败但安装已提交时沿用既有“安装有效、mock 装饰失败”的 typed error 语义，并已在 3.1 记录。

  跨实例同步机制：继续复用既有`cachecoord.Service`、`revisionctrl.Controller`、部署拓扑开关和`plugin-runtime`domain revision；未新增 Redis topic、DB 表、本地轮询、后台 goroutine 或仅当前节点可见的默认实例。最大陈旧窗口仍由读路径调用`ensureRuntimeCacheFresh(ctx)`/`ensureRuntimeCacheFreshBestEffort(ctx, operation)`的时机、共享 revision 可见性和 cachecoord 后端可用性决定；管理列表缓存 key 继续包含 locale、runtime bundle version 和 runtime revision，避免旧读模型跨 revision 命中。

  故障降级与恢复：发布入口在本节点先做幂等 cache invalidation；若`MarkChanged(ctx)`失败，调用方继续收到错误并可重试，未把失败隐藏为成功。远端节点遗漏失效、重启或暂时落后时，后续读路径会通过 revision 对账触发 refresh callback；本地派生缓存也可由显式发布、bundle version、runtime revision 或重新构建恢复。可观测性继续由发布 reason、revision debug 日志和各失效 callback 的错误返回/日志承担；本任务未降低原有错误可见性。

  影响判断：本任务只补充 OpenSpec 缓存一致性记录，不修改生产代码、测试、HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`。`i18n`无资源变化；数据权限语义无变化；数据库/SQL 幂等性、自增主键和软删除语义无影响；开发工具跨平台无影响；无 UI 或用户可观察页面变化，未触发 E2E。DI 来源无新增运行期依赖，继续沿用 6.2 记录的 owner、创建位置、绑定路径和共享实例策略。

  验证：`openspec validate sink-plugin-lifecycle-orchestration --strict`通过；`git diff --check`通过。该任务为缓存一致性治理记录补充，未修改 Go 生产代码或测试，Go 编译门禁沿用 6.2 当前工作区验证结果，最终完整套件在 7.x 统一运行并记录。

## 7. 验证与审查

- [x] 7.1 运行`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`。

  记录：已运行`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`并通过，覆盖插件根 service、capabilityhost、catalog、datahost、dependency、frontend、integration、lifecycle、management、migration、openapi、runtime、sourceupgrade、store、testutil、wasm 和 hostservicedispatch 等包；`governance`、`plugintypes`、`resourcefs`、`runtimeupgrade`当前无测试文件。该任务只记录验证结果，不修改生产代码；无新增 DI、缓存、数据权限、`i18n`、SQL、前端或 E2E 影响。
- [x] 7.2 运行`cd apps/lina-core && go test ./internal/service/i18n/... -count=1`。

  记录：已运行`cd apps/lina-core && go test ./internal/service/i18n/... -count=1`并通过，覆盖宿主`i18n`服务包。该任务只记录验证结果，不修改生产代码；无新增 DI、缓存域、数据权限、SQL、前端或 E2E 影响，`i18n`资源本身无新增或修改。
- [x] 7.3 运行`cd apps/lina-core && go test ./internal/service/cachecoord/... -count=1`。

  记录：已运行`cd apps/lina-core && go test ./internal/service/cachecoord/... -count=1`并通过，覆盖`cachecoord`服务和`revisionctrl`共享 revision 控制器。该任务只记录验证结果，不修改生产代码；无新增 DI、缓存域、数据权限、`i18n`、SQL、前端或 E2E 影响。
- [x] 7.4 运行`cd apps/lina-core && go test ./internal/cmd -count=1`。

  记录：已运行`cd apps/lina-core && go test ./internal/cmd -count=1`并通过，覆盖宿主启动装配和相关绑定测试。该任务只记录验证结果，不修改生产代码；无新增 DI、缓存域、数据权限、`i18n`、SQL、前端或 E2E 影响。
- [x] 7.5 运行`openspec validate sink-plugin-lifecycle-orchestration --strict`和`git diff --check`。

  记录：已运行`openspec validate sink-plugin-lifecycle-orchestration --strict`和`git diff --check`并通过。该任务只记录治理验证结果，不修改生产代码；无新增 DI、缓存域、数据权限、`i18n`、SQL、前端或 E2E 影响。
- [x] 7.6 完成实现、验证和影响记录后调用`lina-review`审查；审查通过后再进入后续变更 D。

  记录：已完成 6.1-7.5 当前阶段最终`lina-review`审查，审查范围覆盖统一发布入口收敛、缓存一致性记录和验证任务记录。审查结论为未发现阻塞问题；规则域覆盖 OpenSpec、后端 Go、架构、插件、缓存一致性、数据权限、测试、`i18n`和文档治理；确认未修改 HTTP API、DTO、OpenAPI 文案、SQL schema、SQL 文件、DAO/DO/Entity 生成物、前端页面、插件 manifest wire 或`apps/lina-plugins/*`，未触发 E2E。

  验证：最终审查前已重新运行`openspec validate sink-plugin-lifecycle-orchestration --strict`和`git diff --check`并通过；7.1-7.5 已分别记录`plugin/...`、`i18n/...`、`cachecoord/...`、`internal/cmd`和治理校验通过结果。
