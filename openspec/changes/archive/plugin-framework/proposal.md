## Why

`apps/lina-core/internal/service/plugin` 同时承载插件清单扫描、治理表读写、生命周期编排、升级治理、运行时认证、缓存协调和 WASM 路由分发。早期方案同时牵引了集群、发布、菜单、角色、认证、监控、E2E 和开发工具等能力；后续通过 Phase A-D 系统治理，将插件框架自身作为长期历史 owner，只保留插件平台的核心演进和决策。

Phase A-D 治理的核心问题包括：catalog 反向回调环、setter 注入链、包级可变状态、生命周期长流程滞留根门面、升级逻辑分散在多个平行包、动态路由认证绕过 role 模块权限边界、清单解析无读模型缓存、delegate 未绑定静默成功、route.go 接近千行。这些问题互相耦合，需要按演进路线系统治理。

## What Changes

- 建立统一插件契约，以 `plugin.yaml`、源码插件目录和动态插件发布产物作为插件身份、资源、依赖、菜单、权限和生命周期的事实入口。
- 定义源码插件和动态插件生命周期，覆盖发现、安装、启用、禁用、卸载、升级、同版本刷新、租户级生命周期和失败诊断。
- 构建动态 WASM 插件运行时，包括自定义段解析、运行时资源视图、前端资产、动态路由、桥接协议、生命周期自动发现和执行资源边界。
- 建立统一宿主服务模型，通过 `hostServices` 授权快照和 `pkg/plugin/capability` 能力目录向动态插件和源码插件暴露受治理的宿主能力。
- 建立插件宿主领域能力模型，覆盖 usercap、authzcap、dictcap、filecap、sessioncap、plugincap、jobcap、infracap，并重整 orgcap、tenantcap 和 ai 能力。
- 将插件公开能力组件统一为 `*cap` 命名体系，删除 `capability/contract` 作为具体能力聚合包的角色。
- **拆解 catalog/store/types**：将 `catalog` 收窄为清单事实源，新建 `store` 接管治理持久化，新建 `plugintypes` 承载纯类型和值对象，消除 catalog 反向回调环。
- **构造函数直化**：删除所有内部 service 的 wiring setter 和 `ValidateRequiredDependencies`，改为构造函数逐项显式注入；切断 runtime/integration/lifecycle 互持宽 service 的循环；消除包级可变运行期状态；将 WASM host service 从包级配置快照改为显式实例；将 runtime revision controller 迁入缓存协调边界。
- **生命周期编排下沉**：将 Install/Uninstall/UpdateStatus/源码生命周期/自动启用/租户钩子编排从根门面下沉到 `internal/lifecycle`；SQL migration executor 独立为 `internal/migration`；收敛列表投影为单一构建入口；统一 `publishPluginChange` 缓存失效入口。
- **升级编排统一**：新建 `internal/upgrade` 吸收 `sourceupgrade` 和 `runtimeupgrade`，统一 source/dynamic 升级的 preview、execute、失败记账、release 提升和缓存发布。
- **运行时认证快照**：动态路由认证改为通过 role 模块发布的访问投影契约读取权限，不再直接查询角色治理表。
- **读模型性能优化**：新增清单读模型缓存，将 ScanManifests 稳态成本收敛为目录枚举加 stat 守卫，优化 WASM 编译缓存失效粒度。
- **运行时组合简化**：收紧 delegate 未绑定语义为 fail-fast，将 kvcache 后端选择改为显式创建。
- **WASM 路由瘦身**：将 route.go 按职责拆分为路由匹配、鉴权、请求封装和响应写回，公共 host call helper 迁回 wasm 公共层。
- **移除动态 i18n host service**：从动态插件 host service catalog 和 WASM dispatcher 中移除 i18n 服务，多语言资源由宿主统一管理。
- 将 `Cache`、`Lock` 和 `Storage` 统一为插件领域能力。
- 将插件资源读取统一为 `Manifest()`，删除旧 `Metadata` 服务语义。
- 收敛插件公共包边界：`pluginhost` 负责源码插件贡献，`pluginbridge` 负责动态插件 ABI 与 transport，`capability` 负责插件消费宿主能力。
- 引入 `plugin.autoEnable` 启动引导、事务性 mock data 安装、安装并启用快捷操作、插件依赖检查、运行时升级预览和显式升级执行。
- 在插件 manifest 中新增`distribution`字段，合法值为`managed|builtin`（缺省归一化为`managed`，拒绝旧值`marketplace`）；支持`builtin`声明项目内建源码插件；启动期自动安装、启用和安全升级 builtin 源码插件，普通插件管理入口隐藏并拒绝写操作。
- 冻结插件领域能力扩展的阶段 0 门禁和阶段 1 第一批高频只读能力（`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`、`Sessions.Current`），同步动态 host service catalog、guest client、WASM dispatcher 和 README。
- 继续完成阶段 1.5 至阶段 5 的剩余能力：候选搜索、组织/租户/插件治理投影、插件私有资源批量、通知类型化和 AI 状态，覆盖 40+ 个新方法的动态发布矩阵、错误语义、规模上限和授权资源。
- 废除`capability.AdminServices`和各领域`AdminService`，每个领域只保留一个插件可见`Service`入口；动态 wire method 一次性标准化，不保留旧方法兼容别名。
- 收敛`plugin.Service`方法集合，删除仅作为语义包装或无生产入口的方法；插件 job 查询合并为`ListManagedJobs`，状态变更合并为`UpdateStatus`。
- 将生产者包中仅用于自组合的分类接口合并回默认`Service`接口，将跨包窄接口移动到消费者包，收敛`i18n.Service`并删除无业务入口的管理诊断 API。

- **插件拥有的非核心领域能力**：非核心领域契约由领域 owner 插件在`backend/cap`维护；`lina-core`只保留内核、依赖治理、通用 descriptor 注册、动态路由、授权、审计与生命周期治理。
- **AI owner 试点**：以`linapro-ai-core`为第一批 owner，将`AI`契约、DTO、SPI、动态 guest SDK 与能力方法状态迁出 core；删除`ProvideAIText`、AI 专属 codec/dispatcher，改为通用 descriptor + owner-aware dispatcher。
- **owner-aware hostServices**：动态插件申请 owner 能力时显式声明`owner`/`version`/`service`/`methods`/资源范围，并在`dependencies.plugins`硬依赖 owner 插件；安装/启用/升级/卸载路径阻断缺失依赖、版本不满足与反向破坏。
- **跨插件 import 边界**：生产代码只能依赖其他插件的`backend/cap/...`，禁止`backend/internal`、DAO/DO/Entity、controller/service 或`backend/pkg`领域入口。
- **宿主分层简化**：新增 core-owned host service 方法强制 JSON envelope；dedicated codec 冻结为方法级名单；wire 常量单一维护在`hostservices`；upgrade 编排归 lifecycle 拥有，根 facade 不再平行持有`upgrade.Service`。

## Capabilities

### New Capabilities

- `plugin-owned-domain-capabilities`：非核心领域 owner 模型、`backend/cap`、descriptor、动态 owner 声明与缓存一致性。
- `plugin-host-layer-simplification`：JSON 载荷政策、dedicated codec 冻结、wire 常量单一来源、lifecycle 升级归属。
- `framework-ai-capability-namespace` / `framework-ai-text-capability` / `framework-ai-multimodal-capabilities`：AI 命名空间与子能力 owner 迁移语义。
- `plugin-manifest-lifecycle`：插件清单发现、生命周期资源同步、只读治理查询、SQL 资源分类、语言资源扩展和依赖声明识别。清单读取和单插件查询的有界读模型缓存。
- `plugin-runtime-loading`：动态插件运行时加载、WASM 自定义段解析、跨节点派生缓存失效、WASM 编译缓存、产物刷新一致性、协调器、保守隐藏策略和批量 manifest 读取。动态请求热路径 manifest 复用和差异对账。
- `plugin-host-service-extension`：动态插件版本化宿主服务协议、CapabilityContext、领域方法授权、安装授权模型、源码插件 hostservices 子组件和适配器复用。WASM host service dispatch 由显式 registry 驱动，公共 helper 归属公共层。动态插件不开放 i18n host service。
- `plugin-capability-boundary-governance`：插件公共组件单一职责、internal 边界、插件间运行时调用接缝、能力公开面治理验证、领域能力语义共享和插件不得依赖宿主核心表实现。
- `plugin-package-boundary-governance`：`pkg/plugin` 公共命名空间收敛、`pluginhost`/`pluginbridge`/`capability` 职责分离、`*cap` 组件包命名、公共原语包、旧包删除和包边界迁移。
- `pluginbridge-subcomponent-architecture`：`pluginbridge` 子组件化拆分、依赖方向、协议行为不变、自动化验证、根包不发布业务能力 client 和公开协议出口唯一。
- `plugin-config-service`：插件配置服务通用只读配置访问、结构体扫描、插件作用域目录约定、宿主公开配置独立服务和源码插件不受 key 白名单限制。
- `plugin-data-service`：动态插件宿主数据服务、数据权限、原始 SQL 禁止、DAO/ORM 执行、DoCommit 拦截、受限 ORM SDK、枚举语义值和只访问当前插件自有表。datahost 表契约缓存。
- `plugin-cache-service`：插件缓存领域能力契约。宿主共享 `kvcache.Service` 必须由 HTTP 启动期按拓扑显式构造。
- `plugin-lock-service`：插件锁领域能力契约。
- `plugin-network-service`：动态插件出站 HTTP、授权 URL 模式、default-deny。
- `plugin-notify-service`：动态插件通知服务、授权通道。
- `plugin-storage-service`：插件存储领域能力契约。Storage 新增 Provider 扩展机制和分片上传。
- `plugin-hook-slot-extension`：宿主后端扩展点、前端 slot 扩展点。
- `plugin-ui-integration`：插件页面多模式集成、前端资产托管。
- `plugin-embed-snapshot-packaging`：动态插件 go:embed 资源声明。
- `plugin-id-governance`：插件 ID 安全边界。
- `plugin-dependency-management`：插件依赖声明、结构化校验。
- `plugin-startup-bootstrap`：plugin.autoEnable 结构化条目。
- `plugin-mock-data-installation`：安装请求可选 mock-data 加载。
- `plugin-install-enable-shortcut`：安装弹窗安装并启用快捷操作。
- `plugin-runtime-upgrade`：插件运行时升级状态标记。
- `plugin-upgrade-governance`：源码插件有效版本和发现版本分离、统一升级模型。
- `plugin-workspace-management`：插件工作区管理。
- `framework-capability-registry`：框架能力按领域归属独立 `*cap` 组件。
- `plugin-host-domain-capabilities`：插件宿主数据领域能力总边界。主资源能力方法使用动作式短名。
- `plugin-service-layout`：插件服务内部构造必须无 setter 回注、组件不得互持宽 service、包级可变状态必须清零、lifecycle 编排归属、upgrade 编排归属、列表投影单一入口、根门面不得直接访问治理 DAO、delegate 未绑定必须 fail-fast。动态路由分发职责拆分。
- `distributed-cache-coordination`：runtime revision controller 必须属于缓存协调边界、插件生命周期缓存失效必须通过单一变化发布入口。集群和单机缓存后端选择必须在拓扑感知构造边界完成。
- `role-management`：role 模块必须发布 token access snapshot 投影给动态插件运行时使用。
- `service-dependency-injection-governance`：Capability Provider Manager 必须由宿主显式持有；缓存敏感服务的启动装配必须显式选择共享后端。
- `plugin-api-query-performance`：插件列表查询无副作用。
- `plugin-permission-governance`：插件菜单和权限复用宿主体系。
- `plugin-upgrade-governance`：插件升级必须由统一升级编排组件执行、失败诊断必须使用单一账本约定、治理守卫必须只在公开入口执行一次、缓存发布必须复用插件变化发布入口。

### Modified Capabilities

- `menu-management`、`role-management`、`user-auth`、`cron-jobs`、`cluster-deployment-mode`、`distributed-locker`、`leader-election`、`project-setup`、`e2e-suite-organization`、`server-monitor`、`online-user`、`core-host-boundary-governance`、`module-decoupling`、`source-upgrade-governance` 和 `system-api-docs` 只保留插件相关交叉影响摘要，当前契约以 `openspec/specs` 为准。

## Impact

- 后端影响集中在插件注册、运行时加载、生命周期编排、WASM 桥接、host service、能力目录、领域能力适配、启动引导、升级治理、缓存一致性、插件管理读模型、协议样板治理、内部文件组织、`pkg/plugin` 包边界治理、host service registry dispatch、领域能力 SPI 分离、provider 注册机制收敛、AI 授权简化、guest 传输单轨化和资源能力领域能力化。
- 前端影响集中在插件管理摘要列表、详情按需加载、动态页面承载、插件菜单和路由刷新、公开资产引用、安装与升级弹窗。
- 数据和配置影响集中在插件治理表、发布快照、迁移账本、资源引用、plugin.autoEnable、插件运行期配置、manifest 资源和动态产物快照。
- 不涉及数据库迁移、HTTP API 路由、前端 UI 或运行时用户可见文案变更。
- 数据权限影响：租户与组织 scope 过滤只迁移类型归属和注入路径，过滤语义和拒绝策略不变；动态插件缓存、锁和存储仍按插件 ID 和租户上下文隔离。
- 缓存一致性影响：Cache 继续复用启动期共享后端；统一 `publishPluginChange` 入口复用 `plugin-runtime` revision controller。
- i18n 影响：动态插件 i18n host service 移除属于边界调整，无运行时用户可见文案或语言包新增。
- 本归档压缩不修改运行时代码、数据库、API、前端页面或插件源码；当前能力契约以 `openspec/specs` 为准，归档仅保留历史设计和治理原因。
