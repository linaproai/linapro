## ADDED Requirements

### Requirement: The default upload size must be unified at 20 MB
The system SHALL set the platform default value of `sys.upload.maxSize` to `20`, and database initialization, config-template defaults, and runtime upload fallbacks SHALL all use that same value unless an administrator explicitly overrides it.

#### Scenario: Host initialization writes the 20 MB default
- **WHEN** an administrator runs the host initialization SQL
- **THEN** the default value of `sys.upload.maxSize` in `sys_config` is `20`
- **AND** the default value read by config management for that built-in parameter is also `20`

#### Scenario: Runtime default remains 20 MB when no override is provided
- **WHEN** the host handles a `multipart` upload request without any administrator override for the upload-size setting
- **THEN** file-upload validation enforces a 20 MB limit
- **AND** the friendly error message triggered by the default limit returns wording equivalent to "file size cannot exceed 20 MB"

### Requirement: All default upload-size sources must stay consistent
The system SHALL keep the database seed value, config-template default, and host static fallback value for `sys.upload.maxSize` consistent so different startup paths do not expose different default upload limits.

#### Scenario: The host starts from the default template
- **WHEN** an operator generates runtime config from the host default `config.template.yaml` and does not change the upload limit separately
- **THEN** the host reads a default upload size of 20 MB
- **AND** that default matches the `sys.upload.maxSize` default written by the host initialization SQL
