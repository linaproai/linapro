## Context

当前缓存能力已经分成两层：

- `internal/service/kvcache` 是宿主内部 KV cache facade，负责 owner、key、TTL、SQL table backend 和 coordination KV backend 的统一抽象。
- 动态插件已经通过 WASM host service 使用受治理缓存能力，调用路径会根据插件授权 namespace、插件 ID 和租户上下文生成内部 cache key。

源码插件通过 `pluginhost.HostServices` 获得宿主能力，但这个服务目录目前没有 cache 入口。若源码插件直接导入 `internal/service/kvcache` 或自行创建缓存实例，会引入三个问题：

1. 插件可以绕过宿主对插件 ID、namespace 和租户上下文的统一 key 约束。
2. 插件调用路径可能创建新的 `kvcache.New()` 实例，破坏启动期共享后端和 `cluster.enabled` 分支。
3. 插件与宿主内部 owner/key 编码耦合，后续调整 coordination backend、限额、观测或授权模型时兼容成本高。

因此，本变更采用源码插件专用 cache contract，由宿主在服务目录中提供按插件作用域绑定的 facade。

## Goals / Non-Goals

**Goals:**

- 向源码插件公开稳定的缓存服务契约，覆盖 `Get`、`Set`、`Delete`、`Incr`、`Expire`。
- 保证源码插件缓存 key 始终由宿主使用当前 `pluginID`、`namespace`、逻辑 key 和租户上下文生成。
- 复用 HTTP 启动期显式注入的同一个 `kvCacheSvc`，不在插件调用路径构造独立缓存服务图。
- 保持源码插件缓存与动态插件缓存一致的有损缓存语义、TTL 语义和集群后端选择。
- 让官方源码插件和测试替身在编译期显式感知 `HostServices.Cache()` 契约变化。

**Non-Goals:**

- 不向源码插件暴露 Redis client、coordination KV、SQL table backend、provider、cleanup 或内部 owner/key 编码。
- 不为源码插件新增跨插件共享缓存或平台级全局缓存写入能力。
- 不把缓存作为权限、配置、租户、插件状态、业务数据或关键缓存修订号的权威存储。
- 不修改动态插件 WASM cache host service wire 协议。
- 不新增数据库表、SQL 迁移或前端交互。

## Decisions

### 1. 在 `pluginservice/contract` 中新增源码插件 cache contract

新增 `contract.CacheService` 和插件可见 `CacheItem`，由源码插件依赖公共 contract，而不是依赖 `internal/service/kvcache`。

建议契约形态：

```go
type CacheService interface {
    Get(ctx context.Context, namespace string, key string) (*CacheItem, bool, error)
    Set(ctx context.Context, namespace string, key string, value string, ttl time.Duration) (*CacheItem, error)
    Delete(ctx context.Context, namespace string, key string) error
    Incr(ctx context.Context, namespace string, key string, delta int64, ttl time.Duration) (*CacheItem, error)
    Expire(ctx context.Context, namespace string, key string, ttl time.Duration) (bool, *gtime.Time, error)
}
```

Rationale:

- `namespace` 和 `key` 是插件可理解的最小输入。
- `time.Duration` 符合后端 Go 时间长度规范，避免源码插件传递裸秒数。
- 返回 `CacheItem` 可隐藏内部 `OwnerType`、encoded key 和 backend name。

Alternative considered:

- 直接暴露 `kvcache.Service`：实现最少，但插件会看到 `OwnerType` 和 encoded key，并可绕过隔离边界，放弃。
- 让插件自行创建 `kvcache.New()`：简单但违反共享实例和集群后端复用要求，放弃。

### 2. `HostServices` 必须支持按插件 ID 绑定的 scoped 目录

`pluginhost.HostServices` 增加 `Cache() contract.CacheService`。宿主 `pluginhostservices` 不应把同一个未绑定服务直接传给所有源码插件，而应在每个插件注册、cron 收集和 hook 分发时创建 plugin-scoped service view。

推荐形态：

```text
base directory
  └─ ForPlugin(pluginID) / ScopedHostServices(pluginID, base)
       └─ Cache() returns cache adapter bound to pluginID
```

HTTP registrar、Cron registrar、managed cron collector 和 hook payload 传入 scoped view：

- `RegisterHTTPRoutes`: 使用当前 manifest ID 包装后再创建 registrar。
- `RegisterCrons`: 使用当前 manifest ID 包装后再创建 registrar。
- `executeSourcePluginHookHandler`: 使用被调插件 ID 包装后创建 hook payload。
- 测试替身补齐 `Cache()`，缺省可返回 nil；涉及缓存能力的测试使用 fake cache。

Rationale:

- 源码插件没有动态插件的 host service authorization envelope，插件 ID 绑定必须由宿主传递服务目录时完成。
- 编译期显式接口变化能发现所有测试和插件调用点。

Alternative considered:

- 在 `CacheService` 每个方法中要求插件传 `pluginID`：会把隔离字段交给调用方，不可靠，放弃。
- 通过 context 注入 pluginID：隐式且易丢失，不利于审计和测试，放弃。

### 3. 适配器内部复用 `kvcache.Service` 并统一 key 映射

源码插件 cache adapter 接收启动期 `kvCacheSvc` 和绑定的 `pluginID`。每次调用：

1. 校验 adapter、`kvCacheSvc`、`pluginID`、`namespace`、`key`。
2. 从 `ctx` 读取当前业务上下文中的租户 ID。
3. 有租户时调用 `kvcache.BuildTenantCacheKey(tenantID, "plugin-cache", pluginID, namespace, key)`。
4. 无租户时调用 `kvcache.BuildCacheKey(pluginID, namespace, key)`。
5. 使用 `kvcache.OwnerTypePlugin` 调用内部服务。
6. 将内部 `kvcache.Item` 映射为 `contract.CacheItem`。

Rationale:

- 与动态插件 cache host service 的 key 语义保持一致。
- 复用 `kvcache` 已有长度校验、TTL、类型错误、coordination KV 和 SQL table backend 行为。
- 所有 backend 选择继续由启动期 `configureDistributedKVCache` / `configureLocalKVCache` 和注入的 `kvCacheSvc` 控制。

Alternative considered:

- 为源码插件单独实现 SQL/Redis cache：会重复后端逻辑并增加一致性风险，放弃。

### 4. 错误和降级语义保持诚实

源码插件 cache 写入、删除、递增和过期失败时必须返回错误，不得伪装成功。读取失败默认返回错误，而不是默认当作 miss；业务插件可以在调用方自行选择降级策略。

Rationale:

- coordination backend 故障时，伪造成功会导致插件业务形成错误假设。
- 将降级策略留给插件业务更清晰，也与已有动态插件 host service 的“写失败返回错误”方向一致。

### 5. i18n、数据权限和缓存一致性影响

- i18n：本变更默认不新增用户可见文案；若实现新增 `bizerr.Code`，必须补齐中英文 error i18n。
- 数据权限：本变更不新增 REST 数据操作接口，也不允许缓存绕过数据权限；插件若缓存由数据权限控制的数据，必须自行使用已有 host services 或业务查询结果构造可见范围内的缓存。
- 缓存一致性：缓存权威数据源是调用方业务数据，不是 cache 本身。`cluster.enabled=false` 时使用单机 SQL/local 路径；`cluster.enabled=true` 时使用启动期注入的 coordination KV backend。最大可接受陈旧时间由插件传入 TTL 决定，故障时写操作返回错误，读操作返回错误或 miss 由插件调用方显式处理。

## Risks / Trade-offs

- `HostServices` 接口扩展会导致较多测试替身编译失败 -> 在任务中显式更新所有实现和官方源码插件测试替身，利用编译错误定位遗漏。
- 源码插件可能把缓存误用为权威业务状态 -> 在 contract 注释、规范和测试中强调有损缓存语义，并禁止暴露全局共享写入。
- hook/cron 路径如果继续传未绑定 base directory，会造成插件间 key 串扰 -> 将 scoped host services 作为单独任务覆盖 HTTP、Cron、managed cron 和 hook，并添加插件 ID 隔离测试。
- 租户上下文缺失时会写入平台级插件缓存 -> 保持与动态插件当前语义一致；需要租户隔离的调用必须在带业务上下文的请求或任务中执行。
- 读取 coordination backend 故障时返回错误可能要求插件调用方处理更多分支 -> 这是更诚实的失败语义，插件可按业务场景选择降级为 miss。
