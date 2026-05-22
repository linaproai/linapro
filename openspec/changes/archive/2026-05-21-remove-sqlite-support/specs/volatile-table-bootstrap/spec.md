## MODIFIED Requirements

### Requirement: 易失性表 MUST 使用普通持久表存储而非引擎特定的临时存储

系统 SHALL 要求 `sys_online_session`、`sys_locker`、`sys_kv_cache` 三张原 MySQL `ENGINE=MEMORY` 语义表在 PostgreSQL 上使用普通持久表存储。SQL 源 DDL MUST NOT 包含 `ENGINE=MEMORY`、`UNLOGGED`、`TEMPORARY` 等任何引擎或临时表声明。PostgreSQL 不再提供“进程重启即清”语义，这三张表 SHALL 分别依赖业务层 `sys_online_session.last_active_time`、`sys_locker.expire_time`、`sys_kv_cache.expire_at`、TTL 清理任务与锁过期抢占自然收敛。

#### Scenario: sys_online_session 在 PostgreSQL 上为持久表

- **WHEN** 在 PostgreSQL 上执行宿主初始化 SQL
- **THEN** `sys_online_session` 表 MUST 创建为普通持久表
- **AND** 表数据在数据库连接断开后 MUST 持久化保留

#### Scenario: sys_locker 在 PostgreSQL 上为持久表

- **WHEN** 在 PostgreSQL 上执行宿主初始化 SQL
- **THEN** `sys_locker` 表 MUST 创建为普通持久表
- **AND** 表 DDL MUST NOT 包含任何引擎或临时表声明

#### Scenario: sys_kv_cache 在 PostgreSQL 上为持久表

- **WHEN** 在 PostgreSQL 上执行宿主初始化 SQL
- **THEN** `sys_kv_cache` 表 MUST 创建为普通持久表
- **AND** 表 DDL MUST NOT 包含任何引擎或临时表声明

### Requirement: 宿主启动期 MUST NOT 清空易失性表

系统 SHALL 在宿主进程启动、重启、滚动发布、集群 leader 切换和插件运行时启动过程中保留 `sys_online_session`、`sys_locker`、`sys_kv_cache` 的现有数据，不得执行 `TRUNCATE`、全表 `DELETE` 或重置自增序列等清空操作。表内记录的可用性 SHALL 分别由 `last_active_time`、`expire_time`、`expire_at` 与业务读取/清理逻辑判断。

#### Scenario: 单节点启动期保留未过期数据

- **WHEN** 宿主以单节点模式（`cluster.enabled=false`）启动且数据库为 PostgreSQL
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
