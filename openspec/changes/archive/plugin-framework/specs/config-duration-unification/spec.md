## Requirements

### Requirement: Duration Configuration Unification

The system SHALL use duration strings for all time-length configuration: `jwt.expire`, `session.timeout`, `session.cleanupInterval`, `monitor.interval`. Parsed to `time.Duration` at the config layer. No legacy integer key compatibility.
