## Why

当前项目已有多个进程内缓存与共享修订号机制，但共享修订号复用了 `sys_kv_cache`。该表本身是基于 `MEMORY` 的可丢失缓存表，适合承载插件 host-cache，不适合作为权限、运行时配置、插件运行时等关键缓存域的持久修订号来源。权限、运行时配置、插件运行时和 Wasm 模块等关键派生缓存需要独立的集群缓存协调能力，以明确权威数据源、一致性窗口、失效触发点和故障降级策略。

## What Changes

- 新增宿主统一的分布式缓存协调能力，区分 `cluster.enabled=false` 的本地失效策略与 `cluster.enabled=true` 的共享修订号/跨实例同步策略。
- 将关键缓存域纳入统一协调：权限拓扑、受保护运行时参数、插件运行时启用快照、插件前端包、插件 i18n 包和动态插件 Wasm 编译缓存。
- 修复共享修订号发布的可靠性问题，要求修订号递增具备原子性、持久性、幂等性和可观测性。
- 保留插件 host-cache 的可丢失缓存语义，不将 `sys_kv_cache` 改为持久业务存储；同时修复 `incr` 在同一数据库存活期间的并发原子性问题。
- 调整动态插件同版本刷新策略，避免跨节点继续命中旧 Wasm 编译缓存或旧前端/i18n 派生缓存。
- 明确关键写路径的失效发布失败处理：权限与配置类关键缓存不得静默吞掉跨节点失效失败。

## Capabilities

### New Capabilities

- `distributed-cache-coordination`: 定义宿主统一缓存协调、修订号发布、自由缓存域、可选策略配置、跨节点同步、陈旧窗口、故障降级和观测要求。

### Modified Capabilities

- `plugin-cache-service`: 插件 host-cache 的可丢失缓存边界、TTL、`incr` 原子性和过期清理要求发生变化。
- `plugin-runtime-loading`: 动态插件运行时缓存、Wasm 编译缓存、前端包和 i18n 派生缓存的跨节点失效要求发生变化。
- `config-management`: 受保护运行时参数缓存需要通过统一协调机制保证跨节点可见性和有界陈旧。
- `role-management`: 角色、菜单、用户角色与插件权限拓扑变更需要通过统一协调机制可靠失效 token 权限快照。

## Impact

- 后端：`internal/service/kvcache`、`internal/service/config`、`internal/service/role`、`internal/service/pluginruntimecache`、`internal/service/plugin/internal/runtime`、`internal/service/plugin/internal/frontend`、`internal/service/i18n`、`internal/service/sysconfig`、`internal/service/menu`、`internal/service/plugin`。
- 数据库：新增 `sys_cache_revision` 持久修订号表；`sys_kv_cache` 保持缓存表语义与 `MEMORY` 引擎，不再承担关键缓存修订号职责。当前项目是全新项目，可直接修改既有 SQL 并重新初始化数据库。
- 运行时：新增缓存域状态查询或健康诊断输出，用于暴露本地修订号、共享修订号、最后同步时间、最近错误和陈旧秒数。
- 测试：新增并发单元测试、双实例服务级测试和动态插件同版本刷新测试，覆盖修订号不丢失、跨节点失效、有界陈旧和故障降级。
- i18n：本次不新增用户可见 UI 文案；若健康诊断或 API 文档新增字段说明，需要同步维护 apidoc i18n 资源。
