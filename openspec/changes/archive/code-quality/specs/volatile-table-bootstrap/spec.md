## MODIFIED Requirements

### Requirement: 易失性表 MUST 使用普通持久表存储

系统 SHALL 要求 `sys_online_session`、`sys_locker`、`sys_kv_cache` 三张表在 PostgreSQL 上使用普通持久表存储。SQL 源 DDL MUST NOT 包含 `ENGINE=MEMORY`、`UNLOGGED`、`TEMPORARY` 等声明。

### Requirement: 宿主启动期 MUST NOT 清空易失性表

系统 SHALL 在宿主进程启动、重启、滚动发布和集群 leader 切换过程中保留这三张表的现有数据，不得执行 TRUNCATE 或全表 DELETE。
