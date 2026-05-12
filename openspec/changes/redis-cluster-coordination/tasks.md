## 1. 配置模型与启动校验

- [x] 1.1 扩展 `internal/service/config` 的 cluster 配置结构，新增 `coordination` 字段和 Redis 配置结构，所有 timeout 字段使用 `time.Duration`
- [x] 1.2 更新 `manifest/config/config.template.yaml`，加入 `cluster.coordination: redis` 和 `cluster.redis` 示例，注释明确单机模式不需要 Redis、集群模式必须配置 Redis
- [x] 1.3 实现 `cluster.enabled=true` 时的配置校验：`coordination` 必填、当前仅允许 `redis`、Redis address 必填、timeout 必须为带单位时长字符串
- [x] 1.4 保持 SQLite 方言强制 `cluster.enabled=false`，并确保 SQLite 模式即使配置 Redis coordination 也不连接 Redis
- [x] 1.5 在 HTTP runtime 启动编排中加入 coordination 初始化阶段，确保 Redis 探活成功后才启动 cluster、cron、plugin runtime 和 HTTP 服务
- [x] 1.6 为配置解析、非法 coordination、Redis address 缺失、timeout 非法、SQLite 忽略 Redis 配置添加单元测试
- [x] 1.7 定义启动期配置错误码或启动诊断错误，确保用户能明确看到失败字段与修复建议

## 2. Coordination Provider 抽象

- [x] 2.1 新增 `internal/service/coordination` 包，定义 `Service`、`Provider`、`LockStore`、`KVStore`、`RevisionStore`、`EventBus`、`HealthChecker` 等窄接口
- [x] 2.2 定义 coordination 通用类型：backend name、lock handle、fencing token、revision key、event payload、tenant invalidation scope、health snapshot
- [x] 2.3 实现集中 Redis key builder，统一 namespace、tenant、domain、scope、owner、plugin、node 维度编码，禁止业务模块手写 Redis key
- [x] 2.4 实现 fake/in-memory provider，用于单元测试覆盖 lock、KV、revision、event、health 语义
- [x] 2.5 定义 coordination 错误分类和 `bizerr` 错误码：配置错误、连接错误、锁未持有、revision 不可用、event 发布失败、KV 操作失败
- [x] 2.6 为 provider 抽象添加接口级单元测试，覆盖 key 生成、tenant 隔离、context cancel、错误分类和 health snapshot

## 3. Redis Provider 实现

- [x] 3.1 选择并接入 Redis Go 客户端，要求支持 context、连接池、超时、`SET NX PX`、Lua 或等价原子 compare 操作、Pub/Sub
- [x] 3.2 实现 Redis 连接初始化、认证、DB 选择、超时配置、Ping 探活和关闭流程
- [x] 3.3 实现 Redis `LockStore`：Acquire、Renew、Release、IsHeld、owner token 校验、TTL、可选 fencing token
- [x] 3.4 实现 Redis `KVStore`：Get、Set、SetNX、Delete、IncrBy、Expire、TTL、CompareAndDelete
- [x] 3.5 实现 Redis `RevisionStore`：按 tenant/domain/scope 原子 Bump 和 Current，支持 cascade metadata
- [x] 3.6 实现 Redis `EventBus`：发布 cache invalidation event、订阅循环、重复事件幂等处理、source node 识别
- [x] 3.7 实现 Redis health snapshot：ping 状态、最近成功时间、最近错误、subscriber 状态、backend 名称
- [x] 3.8 添加 Redis 真连接集成测试，通过 `LINA_TEST_REDIS_ADDR` 显式启用，使用独立 namespace，禁止 `FLUSHDB`
- [x] 3.9 添加 Redis provider 故障测试，覆盖连接失败、超时、owner token 不匹配、event 发布失败和 context cancel

## 4. Cluster 与 Leader Election 迁移

- [x] 4.1 修改 `internal/service/cluster` 构造函数，集群模式接收 coordination lock store，单机模式保持无 Redis 依赖
- [x] 4.2 将 leader election 从 `locker.New()` SQL locker 切换到 coordination lock，使用固定 leader lock name、node ID、owner token 和 lease TTL
- [x] 4.3 实现 leader lock 续约失败、owner token 不匹配、Redis 错误时立即 demote，并停止 primary 状态
- [x] 4.4 保留单机模式 `IsPrimary=true` 语义，不启动 Redis 选举循环
- [x] 4.5 保留或调整 SQL locker 仅用于单机/测试/兜底边界，确保集群模式不依赖 `sys_locker`
- [x] 4.6 更新 cluster/leader election 单元测试，覆盖 Redis primary、follower、续约失败、primary 接管和单机跳过 Redis

## 5. Distributed Locker 与插件锁迁移

- [x] 5.1 修改 `internal/service/locker`，抽象出 coordination lock backed implementation 与现有 SQL implementation 的部署分支
- [x] 5.2 实现 Redis lock instance 的 Unlock、Renew、IsHeld，确保 release/renew 均校验 owner token
- [x] 5.3 修改插件 Wasm host lock service，使集群模式下插件锁走 coordination lock
- [x] 5.4 插件锁 key 必须包含插件 ID、租户维度和逻辑锁名，平台共享锁必须通过显式能力与审计
- [x] 5.5 为插件锁添加单元测试，覆盖不同插件同名锁隔离、不同租户同名锁隔离、非持有者释放失败、Redis 故障返回错误
- [x] 5.6 更新 `plugin-lock-service` 相关 apidoc 或 host service 文档，如响应错误语义发生变化则同步 i18n

## 6. Cachecoord Redis Revision/Event 迁移

- [x] 6.1 修改 `internal/service/cachecoord`，在集群模式下使用 coordination `RevisionStore` 和 `EventBus`
- [x] 6.2 保留单机模式进程内 revision 分支，不连接 Redis，不访问共享 revision 表
- [x] 6.3 实现 `MarkTenantChanged` 的 Redis revision bump + event publish + local observed revision 更新流程
- [x] 6.4 实现 `EnsureFresh` 的本地 revision TTL、Redis current revision 读取、refresher 调用、observed revision 更新和 domain failure strategy
- [x] 6.5 确保 tenant scope、cascadeToTenants、tenant=-1 运维全清语义在 Redis key/event 中显式表达
- [x] 6.6 保持 `cachecoord.Snapshot` 可观测字段，增加 Redis backend、event 状态和最近错误
- [x] 6.7 为 cachecoord 添加双实例 fake provider 测试，覆盖 event 收敛、event 丢失后 revision 兜底、重复事件幂等、权限 fail-closed
- [x] 6.8 为 Redis 真连接场景添加可选集成测试，覆盖并发 revision bump 和跨实例 event 通知

## 7. Kvcache Coordination KV Backend

- [x] 7.1 新增 `kvcache` coordination KV backend provider，通过 coordination KVStore 实现 Get、Set、Delete、Incr、Expire、CleanupExpired
- [x] 7.2 修改 kvcache 默认构造或启动注入逻辑：单机模式使用 SQL table backend，集群模式使用 coordination KV backend
- [x] 7.3 设计并实现 Redis value 编码，保留 `Item` 的 string/int/expireAt 语义，并明确 int/string 类型冲突处理
- [x] 7.4 使用 coordination KV 后端原生 TTL，coordination KV backend `RequiresExpiredCleanup=false`
- [x] 7.5 coordination KV backend 写失败、删除失败、递增失败必须返回结构化错误，不得伪装成功
- [x] 7.6 更新 cron 内置任务投射逻辑，coordination KV backend 下不注册 `host:kvcache-cleanup-expired`
- [x] 7.7 添加 kvcache coordination KV backend 单元测试，覆盖 string、int、TTL、incr 并发、Expire、Delete、类型冲突、Redis 故障
- [x] 7.8 更新插件 Wasm host cache service 测试，确认集群模式走 coordination KV backend 且租户 key 隔离

## 8. Auth Token State 迁移

- [x] 8.1 修改 JWT revoke store，集群模式使用 coordination KV 写入 revoked token，TTL 等于 JWT 剩余有效期
- [x] 8.2 保留本地 memory revoke cache 作为当前节点加速层，但集群模式必须以 Redis revoke 状态为跨节点事实源
- [x] 8.3 实现 revoke 读取失败 fail-closed，禁止 Redis 故障时仅凭 JWT 签名放行
- [x] 8.4 将 `pre_token`、select-tenant single-use 状态和 replay marker 迁移到 coordination KV
- [x] 8.5 确认 logout、switch-tenant、force logout 均写 Redis revoke，并在写失败时返回结构化错误或明确部分失败
- [x] 8.6 添加 auth 单元测试，覆盖 logout revoke、switch-tenant 旧 token 失效、pre-token 单次使用、Redis 读取失败 fail-closed
- [x] 8.7 若登录/租户选择前端行为受到影响，更新对应 E2E 子断言；否则在验证结论中说明无前端可见变化

## 9. Session Hot State 迁移

- [ ] 9.1 设计 session hot state payload，包含 token ID、tenant ID、user ID、username、login time、last active、IP/browser/os 等必要字段
- [ ] 9.2 修改 session store，集群模式登录时同时写 Redis hot state 和 `sys_online_session` PostgreSQL 投影
- [ ] 9.3 修改认证中间件，请求路径先验证 JWT/revoke，再读取 Redis session hot state，Redis 不可读时 fail-closed
- [ ] 9.4 实现 Redis session TTL 刷新和 last active 热状态更新
- [ ] 9.5 实现 PostgreSQL `last_active_time` 节流写回，避免每个请求写主库
- [ ] 9.6 修改强制下线流程，先校验 PostgreSQL 投影可见性，再删除 Redis session、写 revoke、删除或标记投影
- [ ] 9.7 保留 PostgreSQL 投影清理任务，清理 Redis 已过期或长时间不活跃的投影行
- [ ] 9.8 添加 session 单元测试，覆盖登录双写、请求校验、租户不匹配、节流写回、强退、Redis 故障 fail-closed、投影 cleanup
- [ ] 9.9 更新 `monitor-online` 插件测试或 E2E，确认在线列表和强退在 Redis hot state 模型下仍符合数据权限

## 10. Role Permission Cache 集成

- [x] 10.1 修改 role access revision controller，集群模式使用 Redis-backed cachecoord revision/event
- [ ] 10.2 确认角色、菜单、用户角色、角色菜单、插件权限治理写路径均发布 `permission-access` revision
- [ ] 10.3 确认 token access cache key 和反向索引包含租户维度
- [ ] 10.4 实现权限 revision 读取失败且超过陈旧窗口时 fail-closed
- [x] 10.5 添加 role 单元测试，覆盖跨节点 revision/event 失效、同一用户多租户权限隔离、Redis 失败 fail-closed

## 11. Runtime Config Cache 集成

- [x] 11.1 修改 runtime param revision controller，集群模式使用 Redis-backed cachecoord revision/event
- [ ] 11.2 确认 `sys.jwt.expire`、`sys.session.timeout`、登录黑名单、cron 配置等受保护参数写路径均发布 `runtime-config` revision
- [ ] 11.3 实现 runtime-config Redis revision 不可读且超过陈旧窗口时返回结构化可见错误
- [x] 11.4 保持单机模式进程内 revision 和本地 gcache 快照行为
- [x] 11.5 添加 config 单元测试，覆盖跨节点快照刷新、Redis 事件丢失后 revision 兜底、Redis 故障错误传播、单机模式无 Redis

## 12. Plugin Runtime Cache 集成

- [x] 12.1 修改 `pluginruntimecache` controller，使集群模式底层使用 Redis-backed cachecoord
- [ ] 12.2 确认插件 install、enable、disable、uninstall、upgrade、active release 切换、dynamic artifact 更新均发布 `plugin-runtime` revision/event
- [ ] 12.3 修改 dynamic plugin reconciler wake-up，使用 Redis revision/event 触发并保留 safety sweep 兜底
- [ ] 12.4 确保收到 plugin-runtime event 后刷新 enabled snapshot、frontend bundle、runtime i18n 和 Wasm 派生缓存
- [ ] 12.5 实现 plugin-runtime freshness 不可确认时 conservative-hide 或结构化错误，不暴露可能已禁用/卸载插件
- [x] 12.6 添加 plugin runtime 单元测试，覆盖跨节点启用/禁用、event 丢失兜底、reconciler revision、frontend/i18n/wasm cache 失效

## 13. Cron 与内置任务调整

- [x] 13.1 修改 Master-Only 任务判定，确保基于 Redis leader lock primary 状态
- [x] 13.2 Redis kvcache backend 下不投射 KV SQL expired cleanup job
- [x] 13.3 保留 access topology sync、runtime param sync、plugin runtime sync watcher 作为 Redis event 的 revision 兜底
- [ ] 13.4 确保 session cleanup 继续清理 PostgreSQL 投影，而 Redis session hot state 由 TTL 过期
- [x] 13.5 添加 cron 单元测试，覆盖 primary 执行、follower 跳过、coordination KV backend 不注册 KV cleanup、watcher 使用 Redis revision

## 14. 系统信息、健康检查与可观测性

- [ ] 14.1 扩展 system info 或 health response，暴露 coordination backend、Redis ping 状态、node ID、primary 状态、最近错误
- [ ] 14.2 扩展 cachecoord snapshot，暴露 Redis shared revision、event subscriber 状态、最近同步时间、stale seconds
- [ ] 14.3 确保诊断响应不暴露 Redis 密码、完整敏感连接串或 token key
- [ ] 14.4 同步 apidoc i18n JSON，覆盖新增 coordination/redis 诊断字段
- [ ] 14.5 如前端展示新增诊断字段，同步维护前端运行时语言包和宿主 manifest i18n
- [ ] 14.6 添加 sysinfo/health 单元测试或接口测试，覆盖 Redis healthy/unhealthy、敏感信息脱敏和 apidoc i18n 存在性

## 15. 文档与部署说明

- [ ] 15.1 更新相关 README/README.zh-CN 或部署文档，说明单机模式无需 Redis、集群模式必须配置 Redis
- [ ] 15.2 更新开发环境说明，说明 Redis 集成测试通过 `LINA_TEST_REDIS_ADDR` 显式启用
- [x] 15.3 更新配置示例和注释，说明当前 coordination 仅支持 `redis`，未来可扩展其他 backend
- [ ] 15.4 检查文档新增或修改时是否需要中英文 README 同步，遵守 markdown 格式规范

## 16. 回归测试与验证

- [ ] 16.1 运行 `cd apps/lina-core && go test ./internal/service/config ./internal/service/coordination ./internal/service/cluster ./internal/service/locker ./internal/service/cachecoord ./internal/service/kvcache -count=1`
- [ ] 16.2 运行 `cd apps/lina-core && go test ./internal/service/auth ./internal/service/session ./internal/service/role ./internal/service/config ./internal/service/cron -count=1`
- [ ] 16.3 运行 `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/pluginruntimecache ./internal/service/i18n ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`
- [ ] 16.4 在 Redis 可用环境下运行显式集成测试：`cd apps/lina-core && LINA_TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service/coordination ./internal/service/cachecoord ./internal/service/kvcache ./internal/service/session -run Redis -count=1`
- [ ] 16.5 如实现影响在线用户、登录、系统信息页面，按 `lina-e2e` 规范新增或更新对应 TC，并运行相关 E2E
- [x] 16.6 运行 `openspec validate redis-cluster-coordination --strict`
- [x] 16.7 运行 `git diff --check -- openspec/changes/redis-cluster-coordination apps/lina-core`
- [ ] 16.8 完成实现后调用 `lina-review` 进行代码、规范、i18n、缓存一致性和数据权限审查

## Feedback

- [x] **FB-1**: Redis provider 包边界需要收敛到 `coordination/internal/redis`，`kvcache` 仅保留 coordination KV 适配层，避免业务缓存层直接表达 Redis 后端
- [x] **FB-2**: 项目介绍文档仍将 `OpenSpec` 描述为内置必需工作流，应调整为可选但推荐的依赖组件，并说明框架提供良好支持
- [x] **FB-3**: `lina-archive-consolidate` 未指定变更列表时应只读取以日期开头命名的归档变更，避免重复聚合已生成的聚合归档目录
- [x] 2026-05-12: `lina-archive-consolidate` 默认归档读取边界修正验证通过:`rg -n '所有已归档|所有子目录|处理 .* 下的所有|日期开头|YYYY-MM-DD|非日期|显式指定' .agents/skills/lina-archive-consolidate/SKILL.md`;`openspec validate redis-cluster-coordination --strict`;`git diff --check -- .agents/skills/lina-archive-consolidate/SKILL.md openspec/changes/redis-cluster-coordination/tasks.md`。确认技能说明已将未指定变更列表时的默认输入集合收敛为目录名匹配 `^[0-9]{4}-[0-9]{2}-[0-9]{2}-` 的原始归档变更,并明确排除不以日期开头的既有聚合目录、手工汇总目录和临时目录;显式指定列表时仍允许处理用户点名的非日期目录。i18n 影响:本轮仅修改 agent skill 文档和 OpenSpec 任务记录,不新增或修改前端运行时语言包、manifest i18n、apidoc i18n、菜单、按钮、表单、接口文档或用户可见运行时文案。缓存一致性影响:本轮不修改运行时代码、不新增缓存、不改变缓存失效或跨实例同步逻辑。
- [x] 2026-05-12: 本轮继续实现验证通过:`gofmt -w apps/lina-core/internal/service/config/config_cluster.go apps/lina-core/internal/service/config/config_cluster_test.go apps/lina-core/internal/service/coordination/coordination_test.go apps/lina-core/internal/service/coordination/coordination_redis_integration_test.go apps/lina-core/internal/service/cachecoord/cachecoord.go apps/lina-core/internal/service/cachecoord/cachecoord_revision.go apps/lina-core/internal/service/cachecoord/cachecoord_test.go apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache.go apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache_test.go`;`cd apps/lina-core && go test ./internal/service/config ./internal/cmd -run 'TestGetCluster|TestProductionPanicGovernance' -count=1`;`cd apps/lina-core && go test ./internal/service/coordination ./internal/service/cachecoord ./internal/service/kvcache ./internal/service/plugin/internal/wasm -count=1`。确认启动期 cluster 配置诊断错误包含失败字段与修复建议;Redis provider 新增默认可运行的连接关闭/超时故障测试,可选真连接集成测试通过 `LINA_TEST_REDIS_ADDR` 显式启用、独立 namespace 且只删除精确测试 key;`cachecoord.Snapshot` 暴露 coordination backend、health、event subscriber 和最后事件时间;Wasm cache host service 在 coordination KV backend 下按租户 identity 构造 tenant-aware cache key。i18n 影响:本轮仅修改后端诊断错误文本、内部测试和 OpenSpec 任务记录,不新增前端菜单/按钮/表单/接口文档字段、插件 manifest 文案、运行时 i18n JSON、manifest i18n 或 apidoc i18n。缓存一致性影响:本轮补强 cachecoord 可观测性和插件 cache 租户隔离测试;不改变 cachecoord revision/event 一致性模型,Redis 真连接集成测试保持显式启用且禁止 `FLUSHDB`。
- [x] 2026-05-12: 插件锁 coordination 迁移验证通过:`gofmt -w apps/lina-core/internal/cmd/cmd_http_runtime.go apps/lina-core/internal/service/coordination/internal/core/core_memory.go apps/lina-core/internal/service/hostlock/hostlock.go apps/lina-core/internal/service/hostlock/hostlock_code.go apps/lina-core/internal/service/hostlock/hostlock_ticket.go apps/lina-core/internal/service/locker/locker.go apps/lina-core/internal/service/locker/locker_instance.go apps/lina-core/internal/service/locker/locker_coordination_test.go apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_lock.go apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_lock_test.go`;`python3 -m json.tool apps/lina-core/manifest/i18n/zh-CN/error.json >/dev/null && python3 -m json.tool apps/lina-core/manifest/i18n/en-US/error.json >/dev/null`;`cd apps/lina-core && go test ./internal/service/coordination ./internal/service/locker ./internal/service/hostlock ./internal/service/plugin/internal/wasm ./internal/cmd -count=1`。确认集群启动时 `locker` 切换到 coordination lock,单机恢复 SQL lock;coordination lock instance 的 `Unlock`/`Renew`/`IsHeld` 均使用 owner token 校验;Wasm 插件锁名包含 plugin ID、tenant ID 和逻辑锁名;不同插件同名锁、不同租户同名锁可并行持有,同插件同租户重复获取会 miss,跨租户 ticket release 被拒绝。i18n 影响:新增 `HOST_LOCK_TICKET_TENANT_MISMATCH` 业务错误码并同步 `manifest/i18n/zh-CN/error.json` 与 `manifest/i18n/en-US/error.json`;本轮未新增前端页面文案、apidoc 字段或插件 manifest 文案。缓存一致性影响:本轮不新增业务缓存;锁状态在 cluster 模式下以 coordination lock/Redis TTL 为权威,release/renew 幂等并按 owner token fail-closed,单机模式仍使用 `sys_locker`。
- [x] 2026-05-12: Redis cachecoord 真连接集成测试验证通过:`gofmt -w apps/lina-core/internal/service/cachecoord/cachecoord_redis_integration_test.go`;`cd apps/lina-core && go test ./internal/service/cachecoord -count=1`。新增 `TestRedisCacheCoordIntegrationConcurrentRevisionAndEvent`,通过 `LINA_TEST_REDIS_ADDR` 显式启用,使用独立 Redis key namespace 和精确 revision key 删除,不使用 `FLUSHDB`;测试两个独立 Redis coordination service 模拟两个节点,覆盖并发 `MarkChanged` revision bump 不丢失、跨节点 `cache.invalidate` Pub/Sub 事件通知、事件驱动 `EnsureFresh` 收敛到最新 revision,并校验 `Snapshot` 中 Redis backend、订阅状态、最近事件时间、本地/共享 revision。i18n 影响:本轮仅新增后端可选集成测试和 OpenSpec 任务记录,不新增用户可见文案、前端语言包、manifest i18n 或 apidoc i18n。缓存一致性影响:本轮不改变生产一致性模型,补充验证 Redis revision 为权威事实源、Pub/Sub 为低延迟通知、`EnsureFresh` 读取 shared revision 完成事件丢失/乱序兜底;单机路径不受影响。
- [x] 2026-05-12: Auth token state 迁移收敛验证通过:`gofmt -w apps/lina-core/internal/service/auth/auth.go apps/lina-core/internal/service/auth/auth_tenant_flow_test.go apps/lina-core/internal/controller/auth/auth_v1_logout.go`;`python3 -m json.tool apps/lina-core/manifest/i18n/en-US/error.json >/dev/null && python3 -m json.tool apps/lina-core/manifest/i18n/zh-CN/error.json >/dev/null && python3 -m json.tool apps/lina-core/internal/packed/manifest/i18n/en-US/error.json >/dev/null && python3 -m json.tool apps/lina-core/internal/packed/manifest/i18n/zh-CN/error.json >/dev/null`;`cd apps/lina-core && go test ./internal/service/auth ./internal/controller/auth ./pkg/pluginservice/session ./internal/service/middleware -count=1`;`cd apps/lina-plugins/monitor-online && go test ./backend/internal/service/monitor ./backend/internal/controller/monitor -count=1`。确认 `ParseToken` 在 shared revoke 读取失败时返回 `AUTH_TOKEN_STATE_UNAVAILABLE` 并 fail-closed;`logout`、`switch-tenant` 和 monitor-online force logout 所用 `RevokeSession` 均先写 shared revoke,成功后删除在线会话投影,写失败时返回结构化错误且保留 session projection 供重试;pre-token 单次使用已有共享 KV 测试覆盖。i18n 影响:补齐 `error.auth.token.state.unavailable` 中英文翻译并同步 packed manifest;未新增前端页面文案、按钮、表单、apidoc 字段或插件 manifest 文案。缓存一致性影响:cluster 模式下 revoked token 以 coordination KV/Redis TTL 为跨节点事实源,本地 memory revoke 仅作加速层;读取失败 fail-closed,写失败不删除 session projection 以避免“本地下线但共享 revoke 未写入”的不一致窗口。前端影响:登录、租户选择和切换租户响应结构未变化,只在 Redis token-state 故障时返回已有结构化错误,本轮不新增 E2E 子断言。
