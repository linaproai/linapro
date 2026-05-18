## MODIFIED Requirements

### Requirement: Runtime business upgrades remain a directional constraint in this iteration

The framework SHALL keep runtime business-system upgrade capability only as a directional constraint in the current change. Future work still needs versioning, framework-version linkage, upgrade SQL binding, and execution recording, but this iteration MUST prioritize source upgrades and not implement runtime business upgrades.

#### Scenario: The current iteration limits its implementation scope
- **WHEN** the team executes the devops-tooling iteration
- **THEN** runtime business upgrades remain only as a future P1 direction
- **AND** the current implementation scope stays focused on source upgrades
