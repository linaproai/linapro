## MODIFIED Requirements

### Requirement: Monitor Configuration
系统 SHALL 支持可配置的监控参数。其中采集周期 MUST 使用 duration 字符串配置 `monitor.interval`；保留倍数 `monitor.retentionMultiplier` 继续使用整数配置。

#### Scenario: 使用新配置采集周期
- **GIVEN** 配置文件包含 `monitor.interval`
- **WHEN** 服务启动
- **THEN** 系统 SHALL 使用该 duration 值作为采集周期
- **OR** 未配置时默认使用 30 秒

#### Scenario: Configure retention multiplier
- **GIVEN** the configuration file contains `monitor.retentionMultiplier`
- **WHEN** the cleanup job runs
- **THEN** the system SHALL use the configured multiplier
- **OR** default to 5 if not configured
