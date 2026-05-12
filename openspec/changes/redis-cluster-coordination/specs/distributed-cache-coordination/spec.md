## ADDED Requirements

### Requirement: 集群模式缓存协调必须使用 Redis revision
系统 SHALL 在 `cluster.enabled=true` 时通过 Redis revision store 协调关键缓存域修订号。系统 MUST 不依赖 `sys_cache_revision` 表完成跨节点一致性。

#### Scenario: 集群模式发布缓存 revision
- **WHEN** 集群模式下业务写路径发布 `runtime-config` 变更
- **THEN** 系统递增 Redis revision key
- **AND** 返回新的 shared revision
- **AND** 不通过 `sys_cache_revision` 行锁递增作为主实现

#### Scenario: 单机模式本地 revision
- **WHEN** `cluster.enabled=false` 且业务写路径发布缓存变更
- **THEN** 系统仅更新进程内 revision
- **AND** 不连接 Redis

### Requirement: 集群模式缓存失效必须发布 Redis event
系统 SHALL 在集群模式下为缓存 revision 变更发布跨节点 event。event MUST 携带 tenant ID、cascade 标记、domain、scope、revision、reason、source node 和创建时间。

#### Scenario: 权限拓扑失效事件
- **WHEN** 管理员修改角色权限
- **THEN** 系统递增 `permission-access` revision
- **AND** 发布 `cache.invalidate` event
- **AND** 其他节点收到事件后清理本地 token access snapshot

#### Scenario: 重复事件幂等
- **WHEN** 节点收到相同 event 两次
- **THEN** 节点最多执行一次有效刷新
- **AND** 最终本地 observed revision 与 shared revision 收敛

### Requirement: revision 读取必须兜底 Pub/Sub 丢失
系统 SHALL 保留请求路径或 watcher 的 revision check。即使 Redis event 没有被某个节点收到，该节点也 MUST 能通过读取 Redis revision 判断本地缓存是否陈旧。

#### Scenario: 节点错过失效事件
- **WHEN** 节点 B 在节点 A 发布失效事件时短暂离线
- **AND** 节点 B 恢复后处理请求
- **THEN** 节点 B 读取 Redis revision
- **AND** 发现本地 observed revision 落后
- **AND** 刷新对应本地缓存后继续处理请求

### Requirement: 租户范围必须在 Redis revision 中显式表达
系统 SHALL 在 Redis revision key 和 event payload 中显式表达 tenant scope。单租户失效、平台默认级联失效和全租户运维失效 MUST 可区分。

#### Scenario: 单租户失效
- **WHEN** 租户 A 修改租户级字典覆盖
- **THEN** Redis revision key 包含租户 A ID
- **AND** event payload `tenantId=A`
- **AND** 其他租户缓存不被失效

#### Scenario: 平台默认级联失效
- **WHEN** 平台管理员修改平台默认配置
- **THEN** event payload 包含 `tenantId=0`
- **AND** event payload 包含 `cascadeToTenants=true`
- **AND** 各节点按平台 fallback 语义清理受影响租户视图

### Requirement: 关键缓存故障必须遵循域策略
系统 SHALL 在 Redis revision 不可读或刷新失败时按缓存域策略处理。权限缓存 MUST fail-closed；插件运行时缓存 MUST conservative-hide；运行时配置 MUST 返回可见错误或等价结构化错误。

#### Scenario: 权限 revision 超过陈旧窗口
- **WHEN** 节点无法读取 `permission-access` Redis revision
- **AND** 本地 observed revision 已超过最大陈旧窗口
- **THEN** 受保护 API 权限校验失败
- **AND** 系统不得静默放行请求

#### Scenario: 插件 runtime revision 不可确认
- **WHEN** 节点无法确认 `plugin-runtime` revision freshness
- **THEN** 动态插件能力按 conservative-hide 处理
- **AND** 不得暴露可能已禁用或已卸载的插件入口

