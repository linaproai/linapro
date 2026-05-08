## MODIFIED Requirements

### Requirement: Log TraceID output switch is only controlled by static configuration files

The system SHALL turn off TraceID output in the log by default, and only allow the control of whether to output TraceID through the `logger.extensions.traceIDEnabled` static switch in `config.yaml`.

#### Scenario: The log does not output TraceID by default when it is not explicitly enabled.
- **WHEN** `logger.extensions.traceIDEnabled` is not declared in the configuration file
- **THEN** Host logs and HTTP Server logs do not output the TraceID field by default

#### Scenario: Configuration file explicitly enables TraceID output
- **WHEN** `logger.extensions.traceIDEnabled` is explicitly set to `true` in the configuration file
- **THEN** Host log and HTTP Server log output TraceID field

#### Scenario: TraceID system parameters are no longer exposed when initializing built-in parameters.
- **WHEN** Administrator executes host initialization SQL
- **THEN** `sys_config` does not contain `sys.logger.traceID.enabled` records
