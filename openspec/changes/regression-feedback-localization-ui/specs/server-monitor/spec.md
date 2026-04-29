## ADDED Requirements

### Requirement: Service monitoring disk table must remain readable in English
服务监控页面 SHALL 在英文环境下为磁盘使用情况表格提供可读的列宽分配，避免关键列标题和数值换行。

#### Scenario: English disk table keeps key columns on one line
- **WHEN** 管理员在 `en-US` 环境下打开服务监控页面
- **THEN** `File System`、`Total`、`Used`、`Available` 列标题不换行
- **AND** 这些列的常见值不因 `Mount Path` 列过宽而被挤压换行
- **AND** `Mount Path` 列可通过截断、tooltip 或横向滚动展示长路径

## MODIFIED Requirements

### Requirement: Timed collection of server metrics

The system SHALL starts a timing task on each LinaPro service node, periodically collects the local server indicators and writes them to the database. The acquisition frequency defaults to 1 minute, which can be adjusted through configuration. The responsibility for cleaning up the monitoring data MUST be determined according to the deployment mode: the single node mode is performed by the current node, and the cluster mode is performed by the master node only.

#### Scenario: Timed Acquisition Write Database
- **WHEN** timed task triggers (default every 1 minute)
- **THEN** The system collects CPU, memory, disk, network traffic indicators of the current node through gopsutil, along with Go runtime information and node identification (hostname + IP), and writes a record of the `plugin_monitor_server` table in JSON format

#### Scenario: Collect immediately after the service starts
- **WHEN** LinaPro service startup
- **THEN** The system immediately performs an index acquisition and writes to the database
- **AND** Don't wait for the first timing period

#### Scenario: Single node mode for old data cleanup
- **WHEN** `cluster.enabled = false` and monitor cleanup task triggered
- **THEN** The current node cleans up the historical monitoring data that exceeds the retention threshold

#### Scenario: Cluster Mode Old Data Cleanup Performed by Masternode
- **WHEN** `cluster.enabled = true` and monitor cleanup task triggered
- **THEN** Historical monitoring data cleansing performed by masternodes only

#### Scenario: Clean up expired records (K8S/dynamic environment)
- **GIVEN** Monitoring acquisition interval is N seconds
- **WHEN** monitoring cleanup task execution
- **THEN** System deletes` updated_at < now - N * retention_multiplier `record
- **AND** `retention_multiplier` defaults to 5

### Requirement: Monitor Configuration
System SHALL supports configurable monitoring parameters. where the acquisition period MUST uses the duration string to configure `monitor.interval`; the retention multiple `monitor.retentionMultiplier` continues to use the integer configuration.

#### Scenario: Use the new configuration acquisition cycle
- **GIVEN** Configuration file contains` monitor.interval `
- **WHEN** service start
- **THEN** system SHALL use this duration value as the acquisition period
- **OR** Used by default for 1 minute when not configured

#### Scenario: Configure retention multiplier
- **GIVEN** the configuration file contains `monitor.retentionMultiplier`
- **WHEN** the cleanup job runs
- **THEN** the system SHALL use the configured multiplier
- **OR** default to 5 if not configured
