# Server Monitor

## Overview

The server monitoring feature collects and stores system runtime metrics including CPU, memory, disk, network, and Go runtime data.

## Requirements

### Requirement: Periodic Metrics Collection

The system MUST collect server metrics periodically and store them in the database.

#### Scenario: Collect metrics on startup
- **WHEN** the service starts
- **THEN** the system collects metrics immediately
- **AND** stores them in the `sys_server_monitor` table

#### Scenario: Collect metrics periodically
- **GIVEN** the service is running
- **WHEN** the configured interval has elapsed
- **THEN** the system collects metrics again
- **AND** updates the existing record for the same node

### Requirement: Node-based Record Management

The system MUST maintain exactly one monitoring record per node.

#### Scenario: Upsert on collection
- **WHEN** a node collects and stores metrics
- **THEN** the system uses Save() with unique key (node_name, node_ip)
- **AND** created_at is preserved on update
- **AND** updated_at is automatically updated

### Requirement: Stale Data Cleanup

The system MUST periodically clean up stale monitoring records from inactive nodes.

#### Scenario: Clean up expired records
- **GIVEN** the monitor collection interval is N seconds
- **WHEN** the cleanup cron job runs
- **THEN** the system deletes records where `updated_at` < (now - N * retention_multiplier)
- **AND** the default retention_multiplier is 5

**Why:** In K8S container environments, pod IPs change on each deployment, and old pods may no longer exist. Without cleanup, stale records accumulate indefinitely, degrading query performance and wasting storage.

**How to apply:** Implement a cleanup cron job that runs periodically (e.g., every hour) to delete records where the last update time exceeds the threshold.

### Requirement: Monitor Configuration

The system MUST support configurable monitoring parameters.

#### Scenario: Configure collection interval
- **GIVEN** the configuration file contains `monitor.intervalSeconds`
- **WHEN** the service starts
- **THEN** the system uses the configured interval
- **OR** defaults to 30 seconds if not configured

#### Scenario: Configure retention multiplier
- **GIVEN** the configuration file contains `monitor.retentionMultiplier`
- **WHEN** the cleanup job runs
- **THEN** the system uses the configured multiplier
- **OR** defaults to 5 if not configured
