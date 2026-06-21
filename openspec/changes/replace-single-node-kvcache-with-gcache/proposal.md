## Why

当前单机模式下宿主通用`kvcache`默认使用`sys_kv_cache`数据库表承载缓存数据，导致高频缓存读写进入数据库路径，缓存语义和性能目标不匹配。单机部署本身只有一个宿主进程，适合使用进程内`memory`后端提供 TTL、过期淘汰和原子递增，同时继续使用`sys_online_session`作为用户会话有效性的权威来源。

## What Changes

- 将`cluster.enabled=false`时的宿主共享`kvcache`后端从 SQL table backend 调整为`memory`单进程内存后端，当前实现基于 GoFrame`gcache`。
- 保留`cluster.enabled=true`时的 coordination KV backend，集群模式仍通过 Redis coordination 承载跨实例 KV cache、TTL 和原子递增。
- 将单机模式下 JWT revoke 标记降级为本进程快速拒绝缓存；用户 session 的重启后有效性以`sys_online_session`记录是否存在且未超时为权威。
- 要求所有完整认证入口在 JWT 签名、类型和撤销快速检查后，继续校验`sys_online_session`；低层 JWT 解析不得作为公开`Service`契约暴露或被误用为完整登录态判断。
- 移除单机 SQL table cache 对`sys_kv_cache`的运行时依赖，删除该表的宿主 SQL、DAO/DO/Entity、过期清理定时任务和相关测试路径。
- **BREAKING**：`sys_kv_cache`不再作为宿主数据库交付表存在；依赖直接查询或写入该表的外部脚本、诊断流程或插件实现必须改为通过受治理的 cache host service 或宿主诊断接口访问缓存。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-cache-service`：单机默认缓存后端从`sql-table`改为`memory`内存后端，并更新 TTL、`incr`、故障和清理语义。
- `user-auth`：明确用户登录态的权威判断由`sys_online_session`承担，单机进程内 JWT revoke 仅作为快速拒绝缓存。
- `cluster-deployment-mode`：更新单机模式缓存拓扑，明确单机不依赖 Redis 或`sys_kv_cache`，集群继续使用 Redis coordination。
- `volatile-table-bootstrap`：从易失性表清单中移除`sys_kv_cache`，保留`sys_online_session`和`sys_locker`的持久表与自然过期语义。

## Impact

- 后端 Go：影响`apps/lina-core/internal/service/kvcache`后端实现、HTTP 启动装配、认证 revoke/pre-token 存储、插件 cache facade、WASM cache host service 和相关单元测试。
- 数据库：直接从宿主 SQL 基线删除`sys_kv_cache`建表 SQL、索引和注释，并删除生成的 DAO/DO/Entity；本项目无兼容性负担，不新增过渡清理 SQL 文件，随后运行`make db.init`与`make dao`。
- 定时任务：移除 SQL table cache 过期清理 handler、内置任务投影、i18n job 文案和相关测试。
- 安全边界：完整认证必须继续校验`sys_online_session`，服务重启后依赖会话表拒绝已退出或强制下线 token。
- 性能：单机缓存读写从数据库访问变为进程内访问；数据装配不引入额外 N+1 查询。
- i18n：若移除内置 KV cache cleanup 定时任务的用户可见名称或描述，需同步清理或调整宿主`manifest/i18n`资源；若仅删除后台不可见路径，记录无运行时 UI 文案影响。
- 数据权限：本变更不新增数据读取或写操作接口；认证会话有效性继续使用已有`sys_online_session`边界，不改变在线用户管理的数据权限过滤。
