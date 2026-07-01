## MODIFIED Requirements

### Requirement: 集群启动必须先完成 Redis 探活
系统 SHALL 在 HTTP 服务、定时任务、插件运行时和业务路由启动前完成 Redis coordination 探活。探活失败时，系统 MUST 拒绝以集群模式启动。

#### Scenario: Redis 不可达时拒绝启动
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis 地址不可连接
- **THEN** 宿主启动失败
- **AND** 不注册 HTTP 业务路由
- **AND** 不启动 leader election、cron、插件 runtime reconciler 或缓存 watcher

#### Scenario: Redis 探活成功后继续启动
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis ping 成功
- **THEN** 宿主继续初始化 cluster、coordination、cron 和插件运行时组件
- **AND** 系统信息诊断中显示 coordination backend 为 `redis`
