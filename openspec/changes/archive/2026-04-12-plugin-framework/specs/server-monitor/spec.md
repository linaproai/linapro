## MODIFIED Requirements

### Requirement: 服务器指标定时采集

系统 SHALL 在每个 Lina 服务节点上启动定时任务，周期性采集本机服务器指标并写入数据库。采集频率默认 30 秒，可通过配置调整。监控数据清理责任必须根据部署模式确定：单节点模式由当前节点执行，集群模式仅由主节点执行。

#### Scenario: 定时采集写入数据库
- **WHEN** 定时任务触发（默认每 30 秒）
- **THEN** 系统通过 gopsutil 采集当前节点的 CPU、内存、磁盘、网络流量指标，连同 Go 运行时信息和节点标识（hostname + IP），以 JSON 格式写入 `sys_server_monitor` 表的一条记录

#### Scenario: 服务启动后立即采集
- **WHEN** Lina 服务启动
- **THEN** 系统立即执行一次指标采集并写入数据库
- **AND** 不等待第一个定时周期

#### Scenario: 单节点模式执行旧数据清理
- **WHEN** `cluster.enabled=false` 且监控清理任务触发
- **THEN** 当前节点清理超过保留阈值的历史监控数据

#### Scenario: 集群模式由主节点执行旧数据清理
- **WHEN** `cluster.enabled=true` 且监控清理任务触发
- **THEN** 仅主节点执行历史监控数据清理

#### Scenario: 清理过期记录（K8S/动态环境）
- **GIVEN** 监控采集间隔为 N 秒
- **WHEN** 监控清理任务执行
- **THEN** 系统删除 `updated_at < now - N * retention_multiplier` 的记录
- **AND** `retention_multiplier` 默认值为 5
