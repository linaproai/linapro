## ADDED Requirements

### Requirement: 权限拓扑 revision 必须使用 Redis coordination
系统 SHALL 在集群模式下通过 Redis revision/event 协调权限拓扑变更。角色、菜单、用户角色、角色菜单、插件权限治理等影响权限快照的写路径 MUST 发布 `permission-access` revision。

#### Scenario: 修改角色菜单后跨节点失效
- **WHEN** 管理员修改角色菜单关联
- **THEN** 系统发布 `permission-access` Redis revision
- **AND** 系统发布 cache invalidation event
- **AND** 其他节点清理本地 token access snapshot

#### Scenario: 修改用户角色后跨节点失效
- **WHEN** 管理员修改用户角色绑定
- **THEN** 系统发布 `permission-access` Redis revision
- **AND** 目标用户在任意节点的后续权限校验使用新权限拓扑

### Requirement: 权限 freshness 不可确认时必须 fail-closed
系统 SHALL 在集群模式下对权限拓扑 revision 做 freshness 检查。当 Redis revision 不可读取且本地权限缓存超过最大陈旧窗口时，权限校验 MUST fail-closed。

#### Scenario: Redis 权限 revision 不可读
- **WHEN** 受保护 API 执行权限校验
- **AND** Redis `permission-access` revision 读取失败
- **AND** 本地权限 revision 已超过最大陈旧窗口
- **THEN** 系统拒绝请求
- **AND** 不得使用可能陈旧的权限快照放行

### Requirement: 权限缓存 key 必须包含租户维度
系统 SHALL 在 token access snapshot、用户索引和权限 revision 中显式携带租户维度，避免跨租户权限缓存复用。

#### Scenario: 同一用户不同租户权限隔离
- **WHEN** 用户 U 同时在租户 A 和租户 B 有会话
- **THEN** 租户 A token 的权限快照不得被租户 B token 复用
- **AND** 修改租户 A 权限不得失效租户 B 的无关权限快照

