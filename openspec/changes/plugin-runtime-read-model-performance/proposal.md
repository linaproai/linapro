## Why

插件运行时已经具备`O(1)`动态路由分发、`revision-gated`收敛和管理列表读模型缓存，但在 30 个插件规模下，动态路由热路径、治理操作和高频只读投影仍会反复执行完整`WASM artifact`解析、全量`ScanManifests`和全局`WASM`编译缓存失效。该变更先收敛最确定的 P0/P1 性能瓶颈，让插件数量增长时请求热路径和单插件治理操作保持有界成本。

## What Changes

- 新增插件清单读模型缓存，覆盖源码插件清单、动态插件 desired manifest、release manifest 和 YAML 快照解析结果。
- 将`ScanManifests`稳态成本从“逐插件整文件读取、哈希和解析”收敛为“目录枚举加文件`stat`守卫”，并为单插件读取提供索引化路径。
- 在动态路由同一请求内复用已解析 manifest，避免`matchDynamicRoute`和`prepareDynamicRouteRuntime`重复解析同一产物。
- 将已知单插件变更的`WASM`编译缓存失效收敛为按插件或 artifact 路径失效，并把编译过程移出全局缓存写锁。
- 在集群 peer 观察到`plugin-runtime`修订号变化时执行有界差异对账，而不是无条件清空全部派生缓存。
- 让插件详情、依赖检查、OpenAPI 文档投影和 hook 分发复用同一批清单快照；反向依赖检查改为索引直查。
- 不改变`plugin.yaml`、动态插件交付形态、host service 协议、HTTP API 契约或数据库结构。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-manifest-lifecycle`：补充清单读取和单插件查询的有界读模型缓存要求。
- `plugin-runtime-loading`：补充动态请求热路径 manifest 复用、`WASM`编译缓存精细失效和差异对账要求。
- `plugin-service-layout`：补充治理读投影、依赖检查和 OpenAPI 投影复用同一批快照的性能要求。

## Impact

- 影响`apps/lina-core/internal/service/plugin/internal/catalog`、`store`、`runtime`、`wasm`、`integration`、`dependency`、插件管理列表投影和 OpenAPI 投影路径。
- 涉及缓存一致性，必须复用`plugin-runtime`缓存协调域，覆盖单机和集群两种分支。
- 涉及后端 Go 生产代码和测试，实施阶段必须执行覆盖变更包的`go test`和必要的启动绑定测试。
- 不新增 HTTP API，不新增 SQL 迁移，不改变数据权限边界。
- `i18n`影响限于 OpenAPI/管理投影缓存键必须包含 locale 和运行时翻译包版本；如同步更新`apps/lina-core/pkg/plugin`说明文档，需同时维护`README.md`和`README.zh-CN.md`。
