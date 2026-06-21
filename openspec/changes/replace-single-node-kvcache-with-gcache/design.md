## Context

当前宿主共享`kvcache`由`apps/lina-core/internal/cmd/internal/httpstartup/http_runtime.go`在 HTTP 启动期显式创建。现有路径在`cluster.enabled=false`时选择`kvcache.NewSQLTableProvider()`，最终通过`apps/lina-core/internal/service/kvcache/internal/sqltable`访问`sys_kv_cache`；在`cluster.enabled=true`时选择`kvcache.NewCoordinationKVProvider(coordinationSvc)`，通过 Redis coordination KV 承载跨实例缓存。

`sys_kv_cache`名义上是缓存表，但它承载在业务数据库中，会让插件缓存、认证短期状态和 JWT revoke 标记的读写进入数据库路径。对单机部署而言，缓存读写不需要跨进程共享；继续使用数据库表会把“可丢失缓存”变成低性能持久表，并引入额外清理任务和 DAO/Entity 维护成本。

用户会话有效性已经由`sys_online_session`维护。认证中间件在解析 JWT 后调用`SessionStore().TouchOrValidate(...)`，会校验`sys_online_session`中是否存在对应会话、租户是否匹配、会话是否超时，并刷新`last_active_time`。因此服务重启后，普通登录态判断可以依赖`sys_online_session`作为权威来源，而不是依赖`sys_kv_cache`中的 revoke 标记。

本变更属于`apps/lina-core`宿主通用能力调整，影响缓存服务、认证状态、集群拓扑、数据库交付物和内置定时任务，不属于工作台展示适配。实现阶段需要遵守后端 Go、数据库、缓存一致性、测试、`i18n`和数据权限治理规则。

## Goals / Non-Goals

**Goals:**

- 单机模式将宿主共享`kvcache`默认后端从 SQL table 替换为`memory`进程内缓存，当前实现基于 GoFrame`gcache`。
- 集群模式继续使用 Redis coordination KV，保持跨节点共享、TTL 和原子递增语义。
- 删除`sys_kv_cache`作为宿主数据库交付表的运行时依赖，包括 SQL、DAO/DO/Entity、SQL table backend 和 KV 过期清理定时任务。
- 明确完整认证链必须以`sys_online_session`作为 session 有效性的权威来源；单机 JWT revoke 仅作为本进程快速拒绝缓存。
- 保持源码插件、动态插件、WASM host service 和认证短期状态继续复用启动期注入的共享`kvcache.Service`实例。
- 通过单元测试、编译门禁、SQL/DAO 生成和`openspec validate`证明单机、集群、认证和插件缓存路径没有回归。

**Non-Goals:**

- 不为单机模式引入 Redis、独立缓存进程或新的分布式协调组件。
- 不把插件缓存升级为业务权威数据源；插件缓存仍然是有损缓存。
- 不改变`sys_online_session`表的主体模型、在线用户管理数据权限或会话超时策略。
- 不新增面向前端的缓存管理页面、诊断接口或用户可见交互。
- 不为外部直接访问`sys_kv_cache`提供兼容层。当前项目无历史兼容负担，删除该表是有意的破坏性数据库变更。

## Decisions

### 单机`kvcache`后端使用`memory`

`cluster.enabled=false`时，HTTP 启动期 SHALL 选择新的`memory` provider 创建共享`kvcache.Service`。该 provider 使用 GoFrame`gcache`作为进程内存储，缓存键继续复用现有`BuildCacheKey`、`OwnerType`、租户、插件 ID、`namespace`和 logical key 约束，避免改变插件可见契约。

原因：单机部署只有一个宿主进程，进程内缓存足以满足同进程插件、认证和宿主模块的缓存共享；`gcache`已支持 TTL 和按过期时间淘汰，不需要额外 SQL 表或定时清理任务。相比数据库表，进程内缓存降低数据库访问频次，符合“缓存不应成为数据库热路径”的性能目标。

备选方案：

- 继续使用`sys_kv_cache`：实现最少，但性能和缓存语义不匹配，需要维护 SQL 清理任务，因此拒绝。
- 单机也要求 Redis：一致性更强，但提升部署门槛，违背单机模式精简要求，因此拒绝。
- 自研内存 map：可控性强，但需要自行实现 TTL、并发、清理和容量治理；`gcache`是现有框架能力，因此拒绝。

### 集群模式继续使用 Redis coordination KV

`cluster.enabled=true`时，系统 MUST 继续要求`cluster.coordination=redis`且 coordination 服务初始化成功。共享`kvcache.Service`继续使用`NewCoordinationKVProvider(coordinationSvc)`，不得退回`memory`或任何节点本地缓存。

原因：集群模式下插件缓存、认证短期状态和 token revoke 状态必须跨节点可见。进程内缓存只能在当前节点生效，会导致节点间认证和插件行为不一致。

### `sys_online_session`是完整登录态权威来源

完整认证链 SHALL 包含两个阶段：

1. JWT 签名、token kind、`clientType`和 revoke 快速检查。
2. 基于`sys_online_session`的`TouchOrValidate`会话存在性、租户归属和超时校验。

低层 JWT 解析不得被业务调用方误认为“用户已完整登录”的唯一判断，也不得作为公开`Service`契约暴露。实现阶段应提供更明确的完整认证方法供中间件和插件路由认证使用；所有完整认证入口都必须校验`sys_online_session`。

原因：单机进程重启会丢失`memory`中的 revoke 标记，但`sys_online_session`会保留仍有效的 session。用户退出、强制下线或租户切换必须删除旧 session；重启后旧 token 是否有效由 session 表记录决定。这样可以接受进程内 revoke 缓存丢失，同时不牺牲普通 API 的登录态正确性。

### 单机 JWT revoke 是快速拒绝缓存，不是权威状态

单机模式下，JWT revoke store SHALL 使用进程内缓存快速拒绝当前进程已撤销的 token。退出、强制下线和租户切换仍 MUST 删除`sys_online_session`记录并清理 token 相关访问上下文。服务重启后，revoke 缓存丢失可接受，因为完整认证入口会因`sys_online_session`缺失拒绝已退出或已强制下线 token。

集群模式下，JWT revoke store 继续使用 Redis coordination KV，读取失败仍必须 fail-closed。该要求不因单机`memory`调整而放宽。

### 删除`sys_kv_cache`数据库表和 SQL table backend

实现阶段 SHALL 直接从宿主 SQL 基线中删除`sys_kv_cache`相关建表、索引和注释。当前项目无历史兼容负担，不通过新增后续 SQL 文件执行“先建再删”的过渡清理。随后运行`make db.init`和`make dao`，由生成流程删除或更新`SysKvCache`相关 DAO/DO/Entity 生成物，不手工维护生成代码。

`apps/lina-core/internal/service/kvcache/internal/sqltable`和`BackendSQLTable`应被删除或彻底下线。`kvcache.New()`的本地默认 provider 应固定为`memory`，不得保留进程级可变默认 provider；生产路径仍应优先通过 HTTP 启动期显式注入 provider，避免调用方依赖包级默认状态。

原因：删除表后保留 SQL table backend 会形成不可用分支和误用入口；项目无历史兼容负担，应一次性移除低性能路径。

### 移除 KV cache 过期清理定时任务

`memory`和 Redis coordination KV 都有后端 TTL 能力，`RequiresExpiredCleanup()`在这两个后端下均应为`false`。内置`host:kvcache-cleanup-expired`handler、默认作业投影和相关用户可见 job 文案应删除或停止注册。保留`host:session-cleanup`，因为`sys_online_session`仍是持久 session 表，需要按`last_active_time`自然清理。

### 容量和故障语义

`memory`后端 SHALL 要求`set`、`incr`和`expire`调用传入正 TTL，不支持创建永不过期缓存条目，也不通过任意固定 LRU 容量承担缓存生命周期语义。缓存值长度、命名空间、key 校验、负 TTL 和缺失 TTL 错误应继续保持结构化失败，不得静默截断、创建无过期条目或伪装写入成功。

`memory`后端写入失败通常只来自校验或进程内资源错误；发生失败时 SHALL 返回错误。集群 coordination KV 写入、删除、递增或过期失败仍 SHALL 返回错误。

## Risks / Trade-offs

- 进程重启会丢失单机缓存 → 这是有损缓存的预期行为；完整认证以`sys_online_session`为权威，插件和宿主缓存读取按未命中重建。
- 单机`incr`在重启后会重新开始 → 插件缓存`incr`不得用于业务权威计数；规范明确其只在进程存活期间原子。
- 低层 JWT 解析路径可能绕过 session 校验 → 实现阶段必须移除公开`Service`契约中的低层解析入口，并通过测试覆盖中间件、插件 route auth、租户切换、刷新和 impersonation revoke 等入口。
- 删除`sys_kv_cache`会破坏外部脚本或插件直接表访问 → 项目无兼容负担；插件生产代码本就不得直接依赖宿主核心表，迁移路径是改用 cache host service 或受治理的诊断能力。
- 内存缓存可能增长过快 → 通过强制正 TTL、key/value 校验和有损缓存定位限制占用；不允许无限期保存大 payload，也不依赖任意固定 LRU 容量作为缓存语义兜底。
- 集群误配置可能退回单机缓存 → 启动期必须 fail-fast，`cluster.enabled=true`缺失 coordination 时不得静默启动。
- 删除内置 KV 清理作业可能影响已有 job 投影 → 当前后端不需要外部清理；实现阶段应同步清理 job handler、默认作业和`i18n`资源，避免 UI 中出现不可执行任务。

## Migration Plan

1. 新增`memory` backend/provider，实现`Get`、`GetInt`、`Set`、`Delete`、`Incr`、`Expire`和`CleanupExpired`空操作，保持`kvcache.Service`契约；当前实现基于 GoFrame`gcache`。
2. 修改 HTTP 启动期 provider 选择：单机选择`memory`，集群选择 coordination KV；补充单机和集群 provider 选择测试。
3. 调整 auth revoke store：单机路径使用进程内快速拒绝缓存；集群路径继续使用 coordination KV。确认完整认证入口统一校验`sys_online_session`。
4. 删除 SQL table backend、`BackendSQLTable`默认路径和相关测试，迁移插件 cache、WASM cache、auth pre-token/revoke 测试到`memory`或 fake backend。
5. 直接从宿主 SQL 基线删除`sys_kv_cache`交付 DDL，不新增过渡删除 SQL；运行`make db.init`与`make dao`同步生成物。
6. 移除 KV cache cleanup handler、默认 job、相关错误码和`i18n`资源；保留 session cleanup。
7. 更新 OpenSpec 任务记录中的 DI 来源检查、缓存一致性、数据库幂等性、数据权限、`i18n`和测试验证结果。
8. 运行至少覆盖变更包的 Go 测试、宿主启动绑定包测试、SQL/DAO 生成验证、静态检索和`openspec validate replace-single-node-kvcache-with-gcache --strict`。

回滚策略：如果实现阶段发现`memory`后端存在不可接受的功能缺口，应回滚本变更的实现提交并保留当前 SQL table 方案；由于本提案明确删除`sys_kv_cache`，一旦数据库迁移已执行，回滚需要通过后续 SQL 重新创建表和生成 DAO。当前项目无历史兼容负担，因此不为生产外部数据提供自动恢复脚本。

## Impact Analysis

| 领域 | 影响判断 |
| --- | --- |
| 宿主边界 | 属于`lina-core`宿主通用缓存与认证基础能力，不与工作台页面绑定。 |
| 缓存一致性 | 单机为进程内有损缓存；集群为 Redis coordination KV；权威源按缓存域分别为业务数据库、`sys_online_session`或插件自身数据源。 |
| 数据库 | 直接从宿主 SQL 基线删除`sys_kv_cache`表、索引和注释，并删除 DAO/DO/Entity；不新增过渡删除 SQL，运行生成流程验证 fresh init 基线干净。 |
| 数据权限 | 不新增 HTTP 数据接口；在线用户管理继续使用既有`sys_online_session`数据权限边界。 |
| `i18n` | 若删除 KV cleanup 内置任务的名称、描述或错误码，需要同步清理宿主`manifest/i18n`和接口文档资源；无前端新文案。 |
| DI | 新`memory` provider 由 HTTP 启动期创建并注入共享`kvcache.Service`；插件 facade、WASM host service、auth 状态均复用该共享实例。 |
| 性能 | 单机缓存读写不再访问数据库；session 权威校验继续按 token_id 主键查询，不引入随返回行数增长的`N+1`。 |
| 测试 | 需要覆盖单机`memory`TTL、`incr`、负 TTL、类型错误、重启丢失语义；集群 coordination KV 不注册清理任务；认证重启后依赖 session 表拒绝无效 token。 |

## Open Questions

无待用户决策问题。实现阶段如发现`memory`容量治理需要暴露配置项，应优先使用保守默认值；只有确需用户可配置时再补充配置规范。
