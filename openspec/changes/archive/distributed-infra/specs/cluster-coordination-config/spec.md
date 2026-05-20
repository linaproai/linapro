## ADDED Requirements

### Requirement: 集群模式必须声明协调后端
系统 SHALL 在 `cluster.enabled=true` 时要求配置 `cluster.coordination`。当前版本唯一合法值 MUST 为 `redis`。当 `cluster.enabled=false` 时，系统 MUST 不要求 `cluster.coordination` 存在，也不得因为 Redis 配置缺失而影响单机启动。

#### Scenario: 集群模式缺少 coordination
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** 未声明 `cluster.coordination`
- **THEN** 宿主启动失败
- **AND** 错误信息明确指出集群模式必须配置 `cluster.coordination=redis`

#### Scenario: 集群模式配置非法 coordination
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination=postgres`
- **THEN** 宿主启动失败
- **AND** 错误信息明确指出当前仅支持 `redis`

#### Scenario: 单机模式不要求 Redis
- **WHEN** 配置文件声明 `cluster.enabled=false`
- **AND** 未声明 `cluster.coordination`
- **AND** 未声明 `cluster.redis`
- **THEN** 宿主按单机模式启动成功
- **AND** 系统不得尝试连接 Redis

### Requirement: Redis 配置必须使用集群命名空间
系统 SHALL 在 `cluster.coordination=redis` 时从 `cluster.redis` 读取 Redis 连接配置。配置 MUST 支持 `address`、`db`、`password`、`connectTimeout`、`readTimeout`、`writeTimeout`。所有时间长度 MUST 使用带单位的时长字符串并解析为 `time.Duration`。

#### Scenario: Redis 配置解析成功
- **WHEN** 配置文件声明 `cluster.coordination=redis`
- **AND** `cluster.redis.address="127.0.0.1:6379"`
- **AND** `cluster.redis.connectTimeout=3s`
- **AND** `cluster.redis.readTimeout=2s`
- **AND** `cluster.redis.writeTimeout=2s`
- **THEN** 配置服务返回 Redis 配置对象
- **AND** 超时字段均为 `time.Duration`

#### Scenario: Redis address 缺失
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** `cluster.redis.address` 为空
- **THEN** 宿主启动失败
- **AND** 错误信息包含缺失字段 `cluster.redis.address`

#### Scenario: Redis timeout 格式非法
- **WHEN** 配置文件声明 `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** `cluster.redis.readTimeout=2000`
- **THEN** 宿主启动失败
- **AND** 错误信息要求使用带单位的时长字符串

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
- **AND** 健康诊断中显示 coordination backend 为 `redis`

### Requirement: SQLite 方言必须禁止集群 coordination
当数据库链接为 SQLite 方言时，系统 SHALL 在内存层强制 `cluster.enabled=false`。即使配置文件同时声明 `cluster.coordination=redis`，系统也 MUST 不连接 Redis，不启动集群 coordination。

#### Scenario: SQLite 配置了 Redis coordination
- **WHEN** `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明 `cluster.enabled=true`
- **AND** 配置文件声明 `cluster.coordination=redis`
- **THEN** `IsClusterEnabled` 返回 `false`
- **AND** 系统输出 SQLite 单机模式警告
- **AND** 系统不得连接 Redis

### Requirement: 配置模板必须展示 Redis 集群模式
系统 SHALL 在 `manifest/config/config.template.yaml` 中提供 Redis coordination 配置示例，并明确注释单机模式不需要 Redis、集群模式必须配置 `cluster.coordination=redis`。

#### Scenario: 配置模板包含 Redis coordination
- **WHEN** 开发者查看 `config.template.yaml`
- **THEN** 文件包含 `cluster.coordination: redis` 示例
- **AND** 文件包含 `cluster.redis.address`、`db`、`password`、`connectTimeout`、`readTimeout`、`writeTimeout` 字段说明
- **AND** 注释说明 `cluster.enabled=false` 时不需要 Redis

