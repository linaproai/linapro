## Requirements

### Requirement: Standard Build Must Distinguish Host-Only and Full Build with Official Plugins

Release build SHALL support host-only build and full build including official source plugin submodule. Host-only build must not require `apps/lina-plugins` to exist; full build must verify the submodule is initialized when official source plugin backend, frontend pages, or dynamic plugin artifacts are needed.

#### Scenario: Host-only release build
- **WHEN** the build environment has not initialized the `apps/lina-plugins` submodule
- **AND** operations staff runs the host-only build entry
- **THEN** the build output does not contain official source plugin pages and official plugin backend registration
- **AND** the build must not fail due to missing official plugin workspace

#### Scenario: Full release build missing submodule
- **WHEN** operations staff runs the full build entry including official plugins
- **AND** `apps/lina-plugins` does not exist or is empty
- **THEN** the build fails fast
- **AND** the error prompts to initialize the official plugin submodule

#### Scenario: Full release build includes official source plugins
- **WHEN** `apps/lina-plugins` submodule initialization is complete
- **AND** operations staff runs the full release build entry
- **THEN** the build includes official source plugin backend registration and frontend pages
- **AND** dynamic plugin artifacts are generated per existing output directory conventions
