## Requirements

### Requirement: Server Metrics Collection

The system SHALL periodically collect CPU, memory, disk, network metrics per node. Cleanup is primary-only in cluster mode, direct execution in single-node mode. Interval configured via `monitor.interval` duration string.
