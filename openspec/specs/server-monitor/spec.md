# Server Monitor

## Purpose

Define the server monitoring data acquisition, storage, cleaning and display behavior provided by the `monitor-server` source plugin to ensure that the system can continuously observe the operation status of each node and support troubleshooting and capacity analysis.
## Requirements
### Requirement: Timed collection of server metrics

The system SHALL starts a timing task on each LinaPro service node, periodically collects the local server indicators and writes them to the database. The acquisition frequency defaults to 30 seconds, which can be adjusted through configuration. The responsibility for cleaning up the monitoring data MUST be determined according to the deployment mode: the single node mode is performed by the current node, and the cluster mode is performed by the master node only.

#### Scenario: Timed Acquisition Write Database
- **WHEN** timed task triggers (default every 30 seconds)
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
- **OR** Used by default for 30 seconds when not configured

#### Scenario: Configure retention multiplier
- **GIVEN** the configuration file contains `monitor.retentionMultiplier`
- **WHEN** the cleanup job runs
- **THEN** the system SHALL use the configured multiplier
- **OR** default to 5 if not configured

### Requirement: Multi-node support

The system SHALL supports multiple LinaPro service nodes to independently collect and report monitoring data, and each node only collects its own indicators.

#### Scenario: Multi-Node Data Isolation
- **WHEN** Node A and Node B run the LinaPro service at the same time
- **THEN** There are two node's respective monitoring records in the database, distinguished by the node_name (hostname) and node_ip fields

#### Scenario: New node auto escalation
- **WHEN** Deploy and start the LinaPro service on the new server
- **THEN** This node automatically starts collecting its own metrics and writing them to the database without additional configuration

### Requirement: Service Monitoring API

The system SHALL provides API query server monitoring data, and supports query by node.

#### Scenario: Query a list of all nodes
- **WHEN** Admin calls` get/api/v1/monitor/server `
- **THEN** The system returns the latest monitoring data for all nodes, each containing: node_name, node_ip, cpu information (number of cores, model, usage), memory information (total, used, available, usage), disk information (total, used, available, usage for each partition), network information (number of bytes sent/received, rate), Go runtime information (version, number of goroutines, heap memory allocation), server information (operating system, architecture, startup time), acquisition time

#### Scenario: Query by node
- **WHEN** Admin calls` get/api/v1/monitor/server? nodeName = xxx `
- **THEN** The system only returns the latest monitoring data for the specified node

### Requirement: Collect metric content

The system SHALL collect the following server metrics:

#### Scenario: CPU metrics
- **WHEN** CPU metrics captured by the system
- **THEN** includes: number of CPU cores, CPU model name, CPU usage (percentage)

#### Scenario: Memory Metrics
- **WHEN** system collects memory metrics
- **THEN** includes: total memory, used memory, available memory, memory usage (percentage)

#### Scenario: Disk Metrics
- **WHEN** disk metrics captured by the system
- **THEN** includes each mount point: path, total capacity, used capacity, available capacity, utilization (percentage)

#### Scenario: Network Metrics
- **WHEN** system collects network metrics
- **THEN** includes: the total number of bytes sent, the total number of bytes received; the transmission rate and the reception rate (bytes/second) are calculated by comparing the data collected last time

#### Scenario: Go runtime metrics
- **WHEN** System collects Go runtime metrics
- **THEN** includes: Go version, number of Goroutines, heap memory allocation, GC pause time, GoFrame version

#### Scenario: Server Basics
- **WHEN** System collects basic server information
- **THEN** includes: host name, operating system name, system architecture, service start time, system run time

### Requirement: Service monitoring frontend page

The system SHALL provides a service monitoring page that displays server metrics in the form of cards and tables.

#### Scenario: Overall Page Layout
- **WHEN** Admin access service monitoring page
- **THEN** page displays the following areas: Server basic information card, CPU indicator card (including progress bar), memory indicator card (including progress bar), Go runtime information card, disk usage table (including progress bar), network traffic information

#### Scenario: Multi-Node Switching
- **WHEN** There is monitoring data for multiple nodes in the database
- **THEN** The top of the page shows the node selection drop-down box. After switching nodes, refresh all indicator displays.

#### Scenario: Single node presentation
- **WHEN** monitoring data for only one node in the database
- **THEN** page directly displays the node metric without the node selector

### Requirement: Service monitoring is delivered by an independent source plugin

System SHALL delivers the service monitoring capability as a `monitor-server` source plugin instead of continuing as the host's default built-in module.

#### Scenario: Provides capabilities when the service monitoring plugin is enabled
- **WHEN** `monitor-server` is installed and enabled
- **THEN** Host exposure service monitoring collection, cleaning, query and page capabilities
- **AND** The plugin menu is mounted to the host `system monitoring` directory, and the top-level `parent_key` is `monitor`

#### Scenario: Smooth degradation when service monitoring plugin is missing
- **WHEN** `monitor-server` is not installed or not enabled
- **THEN** The host does not display the service monitoring menu and page entry
- **AND** Other monitoring plugins and host core capabilities continue to operate normally

