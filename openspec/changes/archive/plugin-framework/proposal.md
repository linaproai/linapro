## 为什么

LinaPro 需要一个正式、稳定且可扩展的插件平台,支持编译到宿主的源码插件、可动态安装的 WASM 运行时插件、前端页面集成、后端钩子与插槽扩展点、权限治理、多节点热升级,以及为动态插件提供完整的宿主服务能力模型。没有统一的插件契约、生命周期管理、运行时加载、宿主服务治理、启动自动化和安装用户体验,系统无法在不侵入性修改核心代码的情况下可持续地扩展业务能力。

## 变更内容

- 定义统一插件契约,以 `plugin.yaml` 作为入口清单,覆盖 `apps/lina-plugins/<plugin-id>/` 下的源码插件和从 `plugin.dynamic.storagePath` 发现的动态插件。
- 建立插件生命周期状态机:发现、安装、启用、禁用、卸载、升级和热更新,源码插件和动态插件具有不同的语义。
- 实现动态 WASM 插件运行时加载,包括清单校验、自定义段产物解析、前端资源提取与托管,以及基于 `wazero` 的执行。
- 构建动态插件 REST 运行时,从 `g.Meta` 提取路由契约,在 `/api/v1/extensions/{pluginId}/...` 进行固定前缀分发,由宿主管理认证和权限检查,使用 protobuf 桥接信封,并通过真实 WASM 桥执行(带 501 降级)。
- 通过 `go:embed` 统一动态插件资源声明,构建器读取嵌入资源并转换为宿主可治理的快照自定义段。
- 将宿主服务能力从离散操作码扩展为结构化宿主服务模型,包括 `runtime`、`storage`、`network`、`data`、`cache`、`lock`、`notify` 和 `config` 服务,每项服务都具有资源授权、执行上下文和审计。
- 通过宿主主配置文件中的 `plugin.autoEnable` 添加启动自动启用,在插件装配前设置专用引导阶段,具有快速失败行为和集群感知的主节点执行。引导阶段必须在任何源码插件生命周期写入后同步启动快照,以便同一启动编排中的后续启用检查读取最新的已安装状态。
- 在插件安装对话框中添加安装并启用快捷方式,具有权限控制、部分成功消息传递和 E2E 覆盖。
- 添加模拟数据安装支持,包括 `installMockData` 选项、事务性模拟 SQL 执行、结构化回滚错误和启动引导集成。
- 在授权审查对话框中与宿主服务授权一起显示动态路由暴露,包括后端路由投影和可折叠路由列表。
- 使插件列表查询只读、为宿主服务表注释提供安全元数据查找,以及会话活动写入节流。
- 收敛集群部署拓扑:`cluster.enabled` 开关、`cluster.Service` 作为唯一拓扑门面、领导选举作为内部实现细节,以及通过世代模型实现插件运行时收敛。
- 使用解析为 `time.Duration` 的时长字符串统一 `jwt.expire`、`session.timeout`、`session.cleanupInterval` 和 `monitor.interval` 的时长配置。
- 使用 `sys_notify_channel`、`sys_notify_message` 和 `sys_notify_delivery` 表重建通知域,替换 `sys_user_message`。
- 为静态 API 建立声明式权限中间件,具有访问上下文缓存和基于拓扑修订的失效。
- 将插件配置服务从特定于插件的 `GetMonitor()` 泛化为业务中性的只读访问器,将每个插件的配置结构、默认值和验证保留在插件内部。为动态插件添加 `config` 宿主服务。
- 规范化插件 ID,具有基本安全边界强制(非空、最大 64 字符、kebab-case)、官方插件 ID 结构化命名约定(`<author>-<domain>-<capability>` 作为推荐),以及将所有 10 个官方插件重命名为 `linapro-*` 前缀的破坏性更改。
- 添加插件清单 `dependencies` 声明,支持框架版本约束和插件依赖约束,具有依赖解析、拓扑自动安装、反向依赖保护,以及在 API 和 UI 中显示依赖检查结果。
- 引入运行时升级状态模型,将文件发现与运行时状态分离,具有 `pending_upgrade`/`abnormal`/`upgrade_failed` 状态、显式升级预览和执行 API、统一生命周期回调(`Before*`/`After*` 替换旧的 `Can*` 守卫),以及集群一致的升级协调。
- 启用官方插件工作区作为可选子模块,具有仅宿主构建/测试验证,并提供插件工作区管理命令(`plugins.init`/`plugins.install`/`plugins.update`/`plugins.status`),基于 `hack/config.yaml` 的源声明。
- 在构建时从访客控制器方法自动发现动态插件生命周期处理器,消除手动 `backend/lifecycle/*.yaml` 声明,同时保留产物嵌入契约作为运行时权威。
- 通过 `HostServices.Cache()` 暴露源码插件作用域缓存门面,为插件私有 KV 缓存提供租户隔离、命名空间隔离和集群后端选择。
- Extend dynamic plugin lifecycle with `Upgrade` and `Uninstall` execution-phase callbacks, `Before*`/`After*` lifecycle for tenant disable/delete, and typed manifest snapshot bridge contract.

## Capabilities

### New Capabilities
- `plugin-manifest-lifecycle`: Unified plugin directory structure, manifest schema, resource ownership, install/enable/disable/uninstall/upgrade lifecycle, manifest-driven menu governance, plugin ID safety validation, and dependency declaration recognition.
- `plugin-runtime-loading`: Dynamic WASM plugin discovery, validation, loading, hot-switch, generation propagation, multi-node convergence, and build-time lifecycle contract auto-discovery from guest controller methods.
- `plugin-hook-slot-extension`: Backend hooks, frontend slots, callback registration extension points, execution order, failure isolation, and observability.
- `plugin-ui-integration`: Plugin page mounting (iframe, new-tab, embedded-mount), frontend resource hosting, slot outlet rendering, generation-aware refresh prompts, and host-only empty workspace tolerance.
- `plugin-permission-governance`: Plugin menu and permission reuse of Lina governance modules, role authorization persistence across disable/enable cycles, and runtime permission context.
- `plugin-embed-snapshot-packaging`: Dynamic plugin `go:embed` resource declaration, builder snapshot generation, and directory-scan fallback compatibility.
- `plugin-host-service-extension`: Structured host-service protocol, capability auto-derivation from `hostServices`, resource authorization at install/enable time, and execution context with audit.
- `plugin-storage-service`: Logical storage space isolation, path-prefix authorization, and `put/get/delete/list/stat` methods.
- `plugin-network-service`: Outbound HTTP via authorized URL patterns with scheme/host/port/path matching and default-deny.
- `plugin-data-service`: Table-level data access via structured CRUD/transaction methods, DAO/ORM execution, `DoCommit` interception, and `plugindb` guest SDK.
- `plugin-cache-service`: Distributed KV cache via MySQL `MEMORY` table with namespace isolation, strict length validation, and expiry cleanup. Extended with source-plugin scoped facade through `HostServices.Cache()` for plugin-private KV cache with tenant isolation, namespace isolation, and cluster backend selection.
- `plugin-lock-service`: Named lock resources reusing host distributed lock with ticket-based renew/release.
- `plugin-notify-service`: Unified notification domain with channel-based send, message records, and delivery tracking.
- `plugin-config-service`: Business-neutral read-only configuration access for plugins, including arbitrary key reads, section scanning, basic type parsing, `time.Duration` parsing, and a `config` host service for dynamic plugins.
- `plugin-startup-bootstrap`: `plugin.autoEnable` config, startup bootstrap phase, source/dynamic branching, fail-fast, cluster-aware primary execution, startup snapshot synchronization after source plugin lifecycle writes, and dependency-aware auto-enable with deterministic topological ordering.
- `plugin-mock-data-installation`: Optional mock-data loading during install, transactional mock SQL, structured rollback errors, and startup bootstrap integration.
- `plugin-api-query-performance`: Read-only plugin list queries, safe metadata lookup, and session activity write throttling.
- `plugin-install-enable-shortcut`: Install-and-enable shortcut in the installation dialog with permission gating and partial-success messaging.
- `demo-control-guard`: Demo read-only mode controlled by plugin enabled state (`linapro-ops-demo-guard`), with clear write-blocking messages and minimal session whitelist.
- `system-api-docs`: OpenAPI projection of dynamic plugin routes with runtime-aware response semantics.
- `cluster-deployment-mode`: `cluster.enabled` switch, single-node default, and cluster-aware plugin lifecycle.
- `cluster-topology-boundaries`: `cluster.Service` as sole topology facade with election encapsulation.
- `config-duration-unification`: Unified duration-string configuration for `jwt.expire`, `session.timeout`, `session.cleanupInterval`, and `monitor.interval`.
- `plugin-id-governance`: Plugin ID basic safety boundary enforcement (non-empty, 64-char max, kebab-case), official plugin ID normalization mapping, runtime identity consistency, and governance validation for directory names, manifest IDs, source registration IDs, menu keys, i18n namespaces, and apidoc namespaces.
- `plugin-dependency-management`: Plugin manifest `dependencies` declaration with framework version constraints and plugin dependency constraints, dependency resolution with topological sorting, automatic installation of discovered hard dependencies, reverse-dependency protection on uninstall, and dependency check results exposed via API and UI.
- `plugin-runtime-upgrade`: Runtime upgrade state model (`normal`, `pending_upgrade`, `abnormal`, `upgrade_running`, `upgrade_failed`), startup version drift scanning with status marking (not fail-fast), explicit upgrade preview and execution APIs, unified lifecycle callback model (`Before*`/`After*` replacing old `Can*` guards), upgrade failure diagnostics and retry, and cluster-consistent cache invalidation.
- `plugin-workspace-management`: Plugin workspace de-submodulization, `hack/config.yaml` based plugin source declaration, `plugins.init`/`plugins.install`/`plugins.update`/`plugins.status` cross-platform commands, lock file state tracking, and local dirty protection.
- `official-plugin-workspace-decoupling`: Official source plugin workspace as optional submodule, host-only build/test verification without plugin workspace, plugin-full verification with submodule initialized, and CI matrix separation.

### Modified Capabilities
- `menu-management`: Plugin menu ownership, `menu_key` stability, manifest-driven sync, visibility linkage with plugin state, and plugin autonomous parent mount point selection.
- `role-management`: Plugin menu and permission authorization with persistence across disable/enable cycles.
- `user-auth`: Authentication lifecycle hooks for plugins with failure isolation.
- `module-decoupling`: Plugin dimension extension for graceful degradation when disabled, missing, or upgrading.
- `online-user`: Duration-string session config and throttled `last_active_time` writes.
- `server-monitor`: Duration-string monitor interval and cluster-aware cleanup.
- `cron-jobs`: Primary-node-only vs all-node task classification with cluster mode awareness; cron declaration visibility split into executable handler publishing, authorization preview, and installed declaration projection.
- `leader-election`: Cluster-mode-only election with single-node bypass.
- `source-upgrade-governance`: Removed old development-time upgrade entry requirements; retained framework metadata display requirements.
- `plugin-upgrade-governance`: Source plugin upgrade moved from development-time command to runtime explicit upgrade model; unified dynamic plugin upgrade boundary.
- `project-setup`: Host initialization commands must support host-only workspace; plugin source management through `hack/config.yaml` and `linactl` commands.
- `e2e-suite-organization`: E2E test suite must support host-only and plugin-full separation; plugin workspace missing does not block host test discovery.
- `release-image-build`: Standard build must distinguish host-only build from full build with official plugin submodule.

## Impact

- Backend: New plugin registration, lifecycle management, runtime loading, hook bus, resource indexing, host-service dispatch, multi-node convergence, startup bootstrap with snapshot synchronization, declarative permission middleware, notification domain, cluster topology infrastructure, generalized plugin configuration service, plugin ID governance, dependency resolution engine, runtime upgrade orchestration, unified lifecycle callbacks, and source-plugin cache facade.
- Frontend: Plugin page mounting protocol, resource access mechanism, slot extension registry, generation-aware refresh prompts, install-and-enable shortcut, mock-data checkbox, route exposure review in authorization dialog, dynamic routing adjustments, dependency plan display, runtime upgrade state and actions, and host-only empty workspace tolerance.
- Data layer: New tables for `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, `sys_plugin_state`, `sys_kv_cache`, `sys_notify_channel`, `sys_notify_message`, `sys_notify_delivery`, and removal of `sys_user_message`.
- Build and delivery: `apps/lina-plugins/` source scanning, `hack/build-wasm` builder for WASM artifacts, `go:embed` resource declaration, unified output directory, build-time lifecycle auto-discovery, host-only vs plugin-full build modes, and CI matrix separation.
- Configuration: `plugin.autoEnable`, `cluster.enabled`, `cluster.election.*`, duration-string keys, host-service authorization snapshots, `hack/config.yaml` plugin sources, and workspace management lock files.
- Developer tools: `linactl plugins.init`/`plugins.install`/`plugins.update`/`plugins.status` commands, `linactl test.go`/`test.host`/`test.plugins`/`test.scripts` test matrix, and cross-platform `make` wrappers.
