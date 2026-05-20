## ADDED Requirements

### Requirement: 易失性表 MUST 使用普通持久表存储而非引擎特定的临时存储

系统 SHALL 要求 `sys_online_session`、`sys_locker`、`sys_kv_cache` 三张原 MySQL `ENGINE=MEMORY` 语义表在所有支持的数据库方言上使用普通持久表存储。SQL 源 DDL MUST NOT 包含 `ENGINE=MEMORY`、`UNLOGGED`、`TEMPORARY` 等任何引擎或临时表声明。PostgreSQL 与 SQLite 均不再提供"进程重启即清"语义，这三张表 SHALL 分别依赖业务层 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at`、TTL 清理任务与锁过期抢占自然收敛。

#### Scenario: sys_online_session 在 PG 与 SQLite 上为持久表
- **WHEN** 在 PG 或 SQLite 上执行宿主初始化 SQL
- **THEN** `sys_online_session` 表 MUST 创建为普通持久表
- **AND** 表数据在数据库连接断开后 MUST 持久化保留

#### Scenario: sys_locker 在 PG 与 SQLite 上为持久表
- **WHEN** 在 PG 或 SQLite 上执行宿主初始化 SQL
- **THEN** `sys_locker` 表 MUST 创建为普通持久表
- **AND** 表 DDL MUST NOT 包含任何引擎或临时表声明

#### Scenario: sys_kv_cache 在 PG 与 SQLite 上为持久表
- **WHEN** 在 PG 或 SQLite 上执行宿主初始化 SQL
- **THEN** `sys_kv_cache` 表 MUST 创建为普通持久表
- **AND** 表 DDL MUST NOT 包含任何引擎或临时表声明

### Requirement: 宿主启动期 MUST NOT 清空易失性表

系统 SHALL 在宿主进程启动、重启、滚动发布、集群 leader 切换和插件运行时启动过程中保留 `sys_online_session`、`sys_locker`、`sys_kv_cache` 的现有数据，不得执行 `TRUNCATE`、全表 `DELETE` 或重置自增序列等清空操作。表内记录的可用性 SHALL 分别由 `last_active_time`、`expire_time`、`expire_at` 与业务读取/清理逻辑判断。

#### Scenario: 单节点启动期保留未过期数据
- **WHEN** 宿主以单节点模式（`cluster.enabled=false`）启动且数据库为 PG 或 SQLite
- **THEN** 启动流程 MUST NOT 清空 `sys_online_session`、`sys_locker`、`sys_kv_cache`
- **AND** 未过期的会话、锁和 KV cache 记录在启动后仍可按业务规则继续生效

#### Scenario: 集群模式 leader 切换不清空数据
- **WHEN** 宿主以集群模式（`cluster.enabled=true`）启动
- **THEN** leader 节点与 follower 节点均 MUST NOT 清空 `sys_online_session`、`sys_locker`、`sys_kv_cache`
- **AND** leader 重新选举或节点滚动重启不得导致这三张表的数据被删除
- **AND** 过期数据仍由 TTL 清理和业务过期判断自然收敛

#### Scenario: 启动路径不包含清空 SQL
- **WHEN** 检查宿主启动 bootstrap、cluster leader 回调、HTTP runtime 启动和插件 runtime 启动路径
- **THEN** 代码 MUST NOT 对 `sys_online_session`、`sys_locker`、`sys_kv_cache` 执行 `TRUNCATE TABLE`
- **AND** 代码 MUST NOT 对这三张表执行无条件全表 `DELETE`
- **AND** 代码 MUST NOT 重置这三张表的自增序列

### Requirement: 易失性表过期判断 MUST 早于使用过期数据

系统 SHALL 保证任何读取 `sys_online_session`、`sys_locker`、`sys_kv_cache` 的业务路径在使用记录前先校验对应过期字段：`sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at`。过期记录不得被视为有效会话、有效锁或有效缓存命中。

#### Scenario: 会话读取忽略过期记录
- **WHEN** 会话服务读取 `sys_online_session` 中 `last_active_time` 早于配置会话超时时间窗口的记录
- **THEN** 该记录 MUST 被视为无效会话
- **AND** 后续清理任务可以异步删除该记录

#### Scenario: 分布式锁服务抢占过期锁
- **WHEN** 业务模块首次调用 `locker.Acquire()`
- **THEN** 锁服务 MUST 基于 `expire_time` 判断已有锁是否过期
- **AND** 对于已过期锁，锁服务 MUST 允许新请求按既有抢占或覆盖策略获取锁

#### Scenario: KV cache 读取忽略过期记录
- **WHEN** 业务模块首次调用 `kvcache.Get()` 或 `kvcache.Set()`
- **THEN** KV cache 服务 MUST 基于 `expire_at` 判断缓存记录是否过期
- **AND** 过期记录 MUST 被视为缓存未命中

### Requirement: 易失性表 TTL 兜底 MUST 保证自然过期

系统 SHALL 保证易失性表的过期数据由业务层对应过期字段驱动的 TTL 清理任务持续维护，并与读取时过期判断、锁过期抢占配合形成完整的自然过期语义。会话表使用 `sys_online_session.last_active_time` 与配置超时窗口，锁表使用 `sys_locker.expire_time`，KV cache 表使用 `sys_kv_cache.expire_at`；TTL 清理任务 SHALL 周期性删除已过期记录，独立于宿主启动流程。

#### Scenario: sys_online_session 过期清理
- **WHEN** `session.CleanupInactive` 定时任务触发
- **THEN** 任务 MUST 删除 `sys_online_session` 中 `last_active_time` 早于配置会话超时时间窗口的所有记录
- **AND** 删除 MUST 通过 GoFrame DAO 完成，不绕过软删除/审计逻辑

#### Scenario: sys_locker 过期锁回收
- **WHEN** 业务模块尝试获取已被持有但 `expire_time < NOW()` 的锁
- **THEN** 锁服务 MUST 视该锁为已释放，允许新请求获取
- **AND** 过期锁记录 MUST 由后续清理任务或抢占式覆盖处理

#### Scenario: sys_kv_cache 过期记录清理
- **WHEN** KV cache 服务读取一个 `expire_at < NOW()` 的记录
- **THEN** 服务 MUST 视该记录为不存在
- **AND** 过期记录 MUST 由后续清理任务或抢占式覆盖处理

### Requirement: 易失性表 MUST 通过统一注册点维护自然过期清单

系统 SHOULD 在易失性表治理实现或测试中维护一个明确的易失性表清单（`volatileTables` 或等价命名），用于集中说明哪些表依赖自然过期语义。新增易失性表 MUST 通过修改该清单并补充对应过期判断/清理测试完成，禁止在业务代码中分散维护隐式过期规则。

#### Scenario: 清单显式列出
- **WHEN** 检查易失性表治理实现或测试代码
- **THEN** 易失性表清单 SHOULD 以常量或显式数组形式集中定义（如 `volatileTables = []string{"sys_online_session", "sys_locker", "sys_kv_cache"}`）
- **AND** 清单 MUST NOT 通过遍历整个数据库或动态推断

#### Scenario: 新增易失性表通过清单维护
- **WHEN** 新业务需要一张依赖自然过期语义的表
- **THEN** 新增 SQL 源 DDL（普通持久表）后，必须在自然过期清单中显式追加表名
- **AND** 必须补充过期判断与清理测试
- **AND** 不得在业务代码中独立写启动期清空逻辑
