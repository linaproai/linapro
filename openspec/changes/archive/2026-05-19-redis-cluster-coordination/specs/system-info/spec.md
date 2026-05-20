## ADDED Requirements

### Requirement: 系统信息必须暴露 coordination 状态
系统 SHALL 在系统信息或健康诊断中暴露集群 coordination 状态。响应至少包含 cluster enabled、coordination backend、Redis 健康、当前 node ID、primary 状态和最近错误。

#### Scenario: 集群 Redis 健康状态
- **WHEN** 运维查询系统信息
- **AND** 宿主以 `cluster.coordination=redis` 运行
- **THEN** 响应包含 coordination backend `redis`
- **AND** 响应包含 Redis ping 状态
- **AND** 响应包含当前节点 primary 状态

#### Scenario: Redis 最近错误可见
- **WHEN** Redis coordination 最近发生连接错误
- **THEN** 系统信息响应包含最近错误摘要
- **AND** 不暴露 Redis 密码或敏感连接串

### Requirement: 缓存协调诊断必须包含 Redis revision 状态
系统 SHALL 在缓存协调诊断中展示 Redis-backed revision 状态，包括 domain、scope、tenant scope、本地 observed revision、shared revision、last synced at、recent error 和 stale seconds。

#### Scenario: 查询 revision 诊断
- **WHEN** 运维查询缓存协调状态
- **THEN** 响应展示每个已配置或已触及 domain/scope 的 revision 信息
- **AND** 可识别节点是否落后于 Redis shared revision

### Requirement: 诊断字段必须同步 apidoc i18n
如果系统信息或健康 API 新增 coordination 诊断字段，系统 SHALL 同步维护 apidoc i18n JSON。不得只修改响应结构而遗漏接口文档翻译资源。

#### Scenario: 新增 coordination 字段文档
- **WHEN** API 响应新增 `coordination.backend` 或 `coordination.redisHealthy`
- **THEN** 对应 apidoc i18n JSON 包含字段说明
- **AND** `openspec validate` 和静态检查不发现缺失的 apidoc i18n 资源

