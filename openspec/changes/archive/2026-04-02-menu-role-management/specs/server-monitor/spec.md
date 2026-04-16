# Server Monitor

## Overview

服务器监控功能负责采集和存储系统运行指标，包括 CPU、内存、磁盘、网络、Go 运行时等数据。

## Requirements

### Requirement: Periodic Metrics Collection

The system SHALL collect server metrics periodically and store them in the database.

#### Scenario: Collect metrics on startup
WHEN the service starts
THEN the system SHALL collect metrics immediately
AND store them in `sys_server_monitor` table

#### Scenario: Collect metrics periodically
GIVEN the service is running
WHEN the configured interval has elapsed
THEN the system SHALL collect metrics again
AND update the existing record for the same node

### Requirement: Node-based Record Management

The system SHALL maintain exactly one monitoring record per node.

#### Scenario: Upsert on collection
WHEN a node collects and stores metrics
THEN the system SHALL use Save() with unique key (node_name, node_ip)
AND created_at SHALL be preserved on update
AND updated_at SHALL be automatically updated

### Requirement: Stale Data Cleanup

<!-- ADDED: Monitor data cleanup for K8S/dynamic environments -->
The system SHALL periodically clean up stale monitoring records from inactive nodes.

#### Scenario: Clean up expired records
GIVEN the monitor collection interval is N seconds
WHEN the cleanup cron job runs
THEN the system SHALL delete records where `updated_at` < (now - N * retention_multiplier)
AND the default retention_multiplier SHALL be 5

**Why:** In K8S container environments, pod IPs change on each deployment, and old pods may no longer exist. Without cleanup, stale records accumulate indefinitely, degrading query performance and wasting storage.

**How to apply:** Implement a cleanup cron job that runs periodically (e.g., every hour) to delete records where the last update time exceeds the threshold.

### Requirement: Monitor Configuration

The system SHALL support configurable monitoring parameters.

#### Scenario: Configure collection interval
GIVEN the configuration file contains `monitor.intervalSeconds`
WHEN the service starts
THEN the system SHALL use the configured interval
OR default to 30 seconds if not configured

#### Scenario: Configure retention multiplier
<!-- ADDED: Configurable retention multiplier -->
GIVEN the configuration file contains `monitor.retentionMultiplier`
WHEN the cleanup job runs
THEN the system SHALL use the configured multiplier
OR default to 5 if not configured
