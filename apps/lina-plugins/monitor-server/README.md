# monitor-server

`monitor-server` is the official LinaPro source plugin for server monitoring.

## Scope

This plugin owns:

- server-monitor data collection
- stale snapshot cleanup
- server-monitor query APIs and workspace entry

The host keeps the cron and plugin lifecycle substrate; this plugin supplies the monitoring capability itself.
