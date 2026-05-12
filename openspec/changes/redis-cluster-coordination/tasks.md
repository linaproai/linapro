## 1. 配置模型与启动校验

- [ ] 1.1 扩展 `internal/service/config` 的 cluster 配置结构，新增 `coordination` 字段和 Redis 配置结构，所有 timeout 字段使用 `time.Duration`
- [ ] 1.2 更新 `manifest/config/config.template.yaml`，加入 `cluster.coordination: redis` 和 `cluster.redis` 示例，注释明确单机模式不需要 Redis、集群模式必须配置 Redis
- [ ] 1.3 实现 `cluster.enabled=true` 时的配置校验：`coordination` 必填、当前仅允许 `redis`、Redis address 必填、timeout 必须为带单位时长字符串
- [ ] 1.4 保持 SQLite 方言强制 `cluster.enabled=false`，并确保 SQLite 模式即使配置 Redis coordination 也不连接 Redis
- [ ] 1.5 在 HTTP runtime 启动编排中加入 coordination 初始化阶段，确保 Redis 探活成功后才启动 cluster、cron、plugin runtime 和 HTTP 服务
- [ ] 1.6 为配置解析、非法 coordination、Redis address 缺失、timeout 非法、SQLite 忽略 Redis 配置添加单元测试
- [ ] 1.7 定义启动期配置错误码或启动诊断错误，确保用户能明确看到失败字段与修复建议

## 2. Coordination Provider 抽象

- [ ] 2.1 新增 `internal/service/coordination` 包，定义 `Service`、`Provider`、`LockStore`、`KVStore`、`RevisionStore`、`EventBus`、`HealthChecker` 等窄接口
- [ ] 2.2 定义 coordination 通用类型：backend name、lock handle、fencing token、revision key、event payload、tenant invalidation scope、health snapshot
- [ ] 2.3 实现集中 Redis key builder，统一 namespace、tenant、domain、scope、owner、plugin、node 维度编码，禁止业务模块手写 Redis key
- [ ] 2.4 实现 fake/in-memory provider，用于单元测试覆盖 lock、KV、revision、event、health 语义
- [ ] 2.5 定义 coordination 错误分类和 `bizerr` 错误码：配置错误、连接错误、锁未持有、revision 不可用、event 发布失败、KV 操作失败
- [ ] 2.6 为 provider 抽象添加接口级单元测试，覆盖 key 生成、tenant 隔离、context cancel、错误分类和 health snapshot

## 3. Redis Provider 实现

- [ ] 3.1 选择并接入 Redis Go 客户端，要求支持 context、连接池、超时、`SET NX PX`、Lua 或等价原子 compare 操作、Pub/Sub
- [ ] 3.2 实现 Redis 连接初始化、认证、DB 选择、超时配置、Ping 探活和关闭流程
- [ ] 3.3 实现 Redis `LockStore`：Acquire、Renew、Release、IsHeld、owner token 校验、TTL、可选 fencing token
- [ ] 3.4 实现 Redis `KVStore`：Get、Set、SetNX、Delete、IncrBy、Expire、TTL、CompareAndDelete
- [ ] 3.5 实现 Redis `RevisionStore`：按 tenant/domain/scope 原子 Bump 和 Current，支持 cascade metadata
- [ ] 3.6 实现 Redis `EventBus`：发布 cache invalidation event、订阅循环、重复事件幂等处理、source node 识别
- [ ] 3.7 实现 Redis health snapshot：ping 状态、最近成功时间、最近错误、subscriber 状态、backend 名称
- [ ] 3.8 添加 Redis 真连接集成测试，通过 `LINA_TEST_REDIS_ADDR` 显式启用，使用独立 namespace，禁止 `FLUSHDB`
- [ ] 3.9 添加 Redis provider 故障测试，覆盖连接失败、超时、owner token 不匹配、event 发布失败和 context cancel

## 4. Cluster 与 Leader Election 迁移

- [ ] 4.1 修改 `internal/service/cluster` 构造函数，集群模式接收 coordination lock store，单机模式保持无 Redis 依赖
- [ ] 4.2 将 leader election 从 `locker.New()` SQL locker 切换到 coordination lock，使用固定 leader lock name、node ID、owner token 和 lease TTL
- [ ] 4.3 实现 leader lock 续约失败、owner token 不匹配、Redis 错误时立即 demote，并停止 primary 状态
- [ ] 4.4 保留单机模式 `IsPrimary=true` 语义，不启动 Redis 选举循环
- [ ] 4.5 保留或调整 SQL locker 仅用于单机/测试/兜底边界，确保集群模式不依赖 `sys_locker`
- [ ] 4.6 更新 cluster/leader election 单元测试，覆盖 Redis primary、follower、续约失败、primary 接管和单机跳过 Redis

## 5. Distributed Locker 与插件锁迁移

- [ ] 5.1 修改 `internal/service/locker`，抽象出 coordination lock backed implementation 与现有 SQL implementation 的部署分支
- [ ] 5.2 实现 Redis lock instance 的 Unlock、Renew、IsHeld，确保 release/renew 均校验 owner token
- [ ] 5.3 修改插件 Wasm host lock service，使集群模式下插件锁走 coordination lock
- [ ] 5.4 插件锁 key 必须包含插件 ID、租户维度和逻辑锁名，平台共享锁必须通过显式能力与审计
- [ ] 5.5 为插件锁添加单元测试，覆盖不同插件同名锁隔离、不同租户同名锁隔离、非持有者释放失败、Redis 故障返回错误
- [ ] 5.6 更新 `plugin-lock-service` 相关 apidoc 或 host service 文档，如响应错误语义发生变化则同步 i18n

## 6. Cachecoord Redis Revision/Event 迁移

- [ ] 6.1 修改 `internal/service/cachecoord`，在集群模式下使用 coordination `RevisionStore` 和 `EventBus`
- [ ] 6.2 保留单机模式进程内 revision 分支，不连接 Redis，不访问共享 revision 表
- [ ] 6.3 实现 `MarkTenantChanged` 的 Redis revision bump + event publish + local observed revision 更新流程
- [ ] 6.4 实现 `EnsureFresh` 的本地 revision TTL、Redis current revision 读取、refresher 调用、observed revision 更新和 domain failure strategy
- [ ] 6.5 确保 tenant scope、cascadeToTenants、tenant=-1 运维全清语义在 Redis key/event 中显式表达
- [ ] 6.6 保持 `cachecoord.Snapshot` 可观测字段，增加 Redis backend、event 状态和最近错误
- [ ] 6.7 为 cachecoord 添加双实例 fake provider 测试，覆盖 event 收敛、event 丢失后 revision 兜底、重复事件幂等、权限 fail-closed
- [ ] 6.8 为 Redis 真连接场景添加可选集成测试，覆盖并发 revision bump 和跨实例 event 通知

## 7. Kvcache Redis Backend

- [ ] 7.1 新增 `kvcache` Redis backend provider，通过 coordination KVStore 实现 Get、Set、Delete、Incr、Expire、CleanupExpired
- [ ] 7.2 修改 kvcache 默认构造或启动注入逻辑：单机模式使用 SQL table backend，集群模式使用 Redis backend
- [ ] 7.3 设计并实现 Redis value 编码，保留 `Item` 的 string/int/expireAt 语义，并明确 int/string 类型冲突处理
- [ ] 7.4 使用 Redis 原生 TTL，Redis backend `RequiresExpiredCleanup=false`
- [ ] 7.5 Redis backend 写失败、删除失败、递增失败必须返回结构化错误，不得伪装成功
- [ ] 7.6 更新 cron 内置任务投射逻辑，Redis backend 下不注册 `host:kvcache-cleanup-expired`
- [ ] 7.7 添加 kvcache Redis backend 单元测试，覆盖 string、int、TTL、incr 并发、Expire、Delete、类型冲突、Redis 故障
- [ ] 7.8 更新插件 Wasm host cache service 测试，确认集群模式走 Redis backend 且租户 key 隔离

## 8. Auth Token State 迁移

- [ ] 8.1 修改 JWT revoke store，集群模式使用 coordination KV 写入 revoked token，TTL 等于 JWT 剩余有效期
- [ ] 8.2 保留本地 memory revoke cache 作为当前节点加速层，但集群模式必须以 Redis revoke 状态为跨节点事实源
- [ ] 8.3 实现 revoke 读取失败 fail-closed，禁止 Redis 故障时仅凭 JWT 签名放行
- [ ] 8.4 将 `pre_token`、select-tenant single-use 状态和 replay marker 迁移到 coordination KV
- [ ] 8.5 确认 logout、switch-tenant、force logout 均写 Redis revoke，并在写失败时返回结构化错误或明确部分失败
- [ ] 8.6 添加 auth 单元测试，覆盖 logout revoke、switch-tenant 旧 token 失效、pre-token 单次使用、Redis 读取失败 fail-closed
- [ ] 8.7 若登录/租户选择前端行为受到影响，更新对应 E2E 子断言；否则在验证结论中说明无前端可见变化

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

- [ ] 10.1 修改 role access revision controller，集群模式使用 Redis-backed cachecoord revision/event
- [ ] 10.2 确认角色、菜单、用户角色、角色菜单、插件权限治理写路径均发布 `permission-access` revision
- [ ] 10.3 确认 token access cache key 和反向索引包含租户维度
- [ ] 10.4 实现权限 revision 读取失败且超过陈旧窗口时 fail-closed
- [ ] 10.5 添加 role 单元测试，覆盖跨节点 revision/event 失效、同一用户多租户权限隔离、Redis 失败 fail-closed

## 11. Runtime Config Cache 集成

- [ ] 11.1 修改 runtime param revision controller，集群模式使用 Redis-backed cachecoord revision/event
- [ ] 11.2 确认 `sys.jwt.expire`、`sys.session.timeout`、登录黑名单、cron 配置等受保护参数写路径均发布 `runtime-config` revision
- [ ] 11.3 实现 runtime-config Redis revision 不可读且超过陈旧窗口时返回结构化可见错误
- [ ] 11.4 保持单机模式进程内 revision 和本地 gcache 快照行为
- [ ] 11.5 添加 config 单元测试，覆盖跨节点快照刷新、Redis 事件丢失后 revision 兜底、Redis 故障错误传播、单机模式无 Redis

## 12. Plugin Runtime Cache 集成

- [ ] 12.1 修改 `pluginruntimecache` controller，使集群模式底层使用 Redis-backed cachecoord
- [ ] 12.2 确认插件 install、enable、disable、uninstall、upgrade、active release 切换、dynamic artifact 更新均发布 `plugin-runtime` revision/event
- [ ] 12.3 修改 dynamic plugin reconciler wake-up，使用 Redis revision/event 触发并保留 safety sweep 兜底
- [ ] 12.4 确保收到 plugin-runtime event 后刷新 enabled snapshot、frontend bundle、runtime i18n 和 Wasm 派生缓存
- [ ] 12.5 实现 plugin-runtime freshness 不可确认时 conservative-hide 或结构化错误，不暴露可能已禁用/卸载插件
- [ ] 12.6 添加 plugin runtime 单元测试，覆盖跨节点启用/禁用、event 丢失兜底、reconciler revision、frontend/i18n/wasm cache 失效

## 13. Cron 与内置任务调整

- [ ] 13.1 修改 Master-Only 任务判定，确保基于 Redis leader lock primary 状态
- [ ] 13.2 Redis kvcache backend 下不投射 KV SQL expired cleanup job
- [ ] 13.3 保留 access topology sync、runtime param sync、plugin runtime sync watcher 作为 Redis event 的 revision 兜底
- [ ] 13.4 确保 session cleanup 继续清理 PostgreSQL 投影，而 Redis session hot state 由 TTL 过期
- [ ] 13.5 添加 cron 单元测试，覆盖 primary 执行、follower 跳过、Redis backend 不注册 KV cleanup、watcher 使用 Redis revision

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
- [ ] 15.3 更新配置示例和注释，说明当前 coordination 仅支持 `redis`，未来可扩展其他 backend
- [ ] 15.4 检查文档新增或修改时是否需要中英文 README 同步，遵守 markdown 格式规范

## 16. 回归测试与验证

- [ ] 16.1 运行 `cd apps/lina-core && go test ./internal/service/config ./internal/service/coordination ./internal/service/cluster ./internal/service/locker ./internal/service/cachecoord ./internal/service/kvcache -count=1`
- [ ] 16.2 运行 `cd apps/lina-core && go test ./internal/service/auth ./internal/service/session ./internal/service/role ./internal/service/config ./internal/service/cron -count=1`
- [ ] 16.3 运行 `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/pluginruntimecache ./internal/service/i18n ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`
- [ ] 16.4 在 Redis 可用环境下运行显式集成测试：`cd apps/lina-core && LINA_TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service/coordination ./internal/service/cachecoord ./internal/service/kvcache ./internal/service/session -run Redis -count=1`
- [ ] 16.5 如实现影响在线用户、登录、系统信息页面，按 `lina-e2e` 规范新增或更新对应 TC，并运行相关 E2E
- [ ] 16.6 运行 `openspec validate redis-cluster-coordination --strict`
- [ ] 16.7 运行 `git diff --check -- openspec/changes/redis-cluster-coordination apps/lina-core`
- [ ] 16.8 完成实现后调用 `lina-review` 进行代码、规范、i18n、缓存一致性和数据权限审查

