## ADDED Requirements

### Requirement: 服务器指标定时采集

系统 SHALL 在每个 Lina 服务节点上启动定时任务，周期性采集本机服务器指标并写入数据库。采集频率默认 30 秒，可通过配置调整。

#### Scenario: 定时采集写入数据库
- **WHEN** 定时任务触发（默认每 30 秒）
- **THEN** 系统通过 gopsutil 采集当前节点的 CPU、内存、磁盘、网络流量指标，连同 Go 运行时信息和节点标识（hostname + IP），以 JSON 格式写入 `sys_server_monitor` 表的一条记录

#### Scenario: 服务启动后立即采集
- **WHEN** Lina 服务启动
- **THEN** 系统立即执行一次指标采集并写入数据库，不等待第一个定时周期

#### Scenario: 旧数据自动清理
- **WHEN** 定时采集任务执行时
- **THEN** 系统同时清理超过 1 小时的历史监控数据记录

### Requirement: 多节点支持

系统 SHALL 支持多个 Lina 服务节点独立采集并上报监控数据，每个节点仅采集自身指标。

#### Scenario: 多节点数据隔离
- **WHEN** Node A 和 Node B 同时运行 Lina 服务
- **THEN** 数据库中存在两个节点各自的监控记录，通过 node_name（hostname）和 node_ip 字段区分

#### Scenario: 新增节点自动上报
- **WHEN** 在新服务器上部署并启动 Lina 服务
- **THEN** 该节点自动开始采集自身指标并写入数据库，无需额外配置

### Requirement: 服务监控 API

系统 SHALL 提供 API 查询服务器监控数据，支持按节点查询。

#### Scenario: 查询所有节点列表
- **WHEN** 管理员调用 `GET /api/v1/monitor/server`
- **THEN** 系统返回所有节点的最新一条监控数据，每条包含：node_name、node_ip、cpu 信息（核心数、型号、使用率）、内存信息（总量、已用、可用、使用率）、磁盘信息（各分区总量、已用、可用、使用率）、网络信息（发送/接收字节数、速率）、Go 运行时信息（版本、goroutine 数、堆内存分配）、服务器信息（操作系统、架构、启动时间）、采集时间

#### Scenario: 按节点查询
- **WHEN** 管理员调用 `GET /api/v1/monitor/server?nodeName=xxx`
- **THEN** 系统仅返回指定节点的最新监控数据

### Requirement: 采集指标内容

系统 SHALL 采集以下服务器指标：

#### Scenario: CPU 指标
- **WHEN** 系统采集 CPU 指标
- **THEN** 包含：CPU 核心数、CPU 型号名称、CPU 使用率（百分比）

#### Scenario: 内存指标
- **WHEN** 系统采集内存指标
- **THEN** 包含：总内存、已用内存、可用内存、内存使用率（百分比）

#### Scenario: 磁盘指标
- **WHEN** 系统采集磁盘指标
- **THEN** 包含各挂载点的：路径、总容量、已用容量、可用容量、使用率（百分比）

#### Scenario: 网络指标
- **WHEN** 系统采集网络指标
- **THEN** 包含：总发送字节数、总接收字节数；通过与上一次采集数据对比计算发送速率和接收速率（字节/秒）

#### Scenario: Go 运行时指标
- **WHEN** 系统采集 Go 运行时指标
- **THEN** 包含：Go 版本、Goroutine 数量、堆内存分配量、GC 暂停时间、GoFrame 版本

#### Scenario: 服务器基本信息
- **WHEN** 系统采集服务器基本信息
- **THEN** 包含：主机名、操作系统名称、系统架构、服务启动时间、系统运行时长

### Requirement: 服务监控前端页面

系统 SHALL 提供服务监控页面，以卡片和表格形式展示服务器各项指标。

#### Scenario: 页面整体布局
- **WHEN** 管理员访问服务监控页面
- **THEN** 页面展示以下区域：服务器基本信息卡片、CPU 指标卡片（含进度条）、内存指标卡片（含进度条）、Go 运行时信息卡片、磁盘使用表格（含进度条）、网络流量信息

#### Scenario: 多节点切换
- **WHEN** 数据库中存在多个节点的监控数据
- **THEN** 页面顶部显示节点选择下拉框，切换节点后刷新所有指标展示

#### Scenario: 单节点展示
- **WHEN** 数据库中仅有一个节点的监控数据
- **THEN** 页面直接展示该节点指标，不显示节点选择器
