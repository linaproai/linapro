## ADDED Requirements

### Requirement: 失效消息携带 tenant_id
`cachecoord` 失效广播消息 SHALL 包含 `tenant_id` 字段;每条失效消息必须显式声明作用域:
- `tenant_id = T`(单租户级):仅失效 `(tenant=T, scope, key)` 缓存。
- `tenant_id = 0` + `cascade_to_tenants = true`(平台默认变更):失效所有租户的 `(tenant=*, scope, key)` 缓存。
- `tenant_id = -1`(全租户清空):罕见使用,必须有审计日志。

#### Scenario: 单租户失效
- **WHEN** 租户 A 写入字典覆盖
- **THEN** 广播 `{tenant_id: A, scope: dict, key: type1}`
- **AND** 各节点仅失效 `tenant=A` 行

#### Scenario: 平台默认级联失效
- **WHEN** 平台管理员修改字典平台默认
- **THEN** 广播 `{tenant_id: 0, scope: dict, key: type1, cascade_to_tenants: true}`
- **AND** 各节点失效所有租户的对应行

### Requirement: 禁止无理由全租户清空
普通业务路径 SHALL 不允许发起 `tenant_id = -1` 的全清失效;只有平台管理员的显式运维操作可,且需 operlog 审计。

#### Scenario: 业务路径误用被拒
- **WHEN** 业务 service 尝试 `cache.InvalidateAllTenants(...)`
- **THEN** API 设计上无此调用入口(仅 `/platform/cache/admin/*` 暴露)

### Requirement: 失效消息幂等
节点收到失效消息 SHALL 幂等处理(可重试不报错);重复消息不影响最终一致性。

#### Scenario: 重复广播
- **WHEN** 集群中失效消息因网络抖动被重复发送
- **THEN** 节点处理两次但缓存最终状态一致
