## ADDED Requirements

### Requirement: Plugin installation review flow supports direct chained enablement

The system SHALL allow administrators to trigger enablement directly from the plugin installation review flow, while the host still follows the existing `install -> enable` lifecycle order instead of collapsing both actions into a new implicit state transition.

#### Scenario: Choose install and enable from the installation review dialog
- **WHEN** an administrator chooses `Install and Enable` in the installation review dialog for a plugin that is not installed
- **THEN** the host runs the plugin install lifecycle first
- **AND** after installation succeeds, the host continues with the enable lifecycle
- **AND** when both steps succeed, the plugin ends in the `installed and enabled` state

#### Scenario: Dynamic plugin composite action reuses the authorization snapshot captured during install
- **WHEN** a dynamic plugin completes host-service authorization confirmation in the installation review dialog and the administrator continues with `Install and Enable`
- **THEN** the host persists the authorization snapshot for the current release during install
- **AND** the subsequent enable step reuses that snapshot directly
- **AND** the composite action MUST NOT open a second authorization-confirmation dialog

#### Scenario: Enablement failure in the composite action does not roll back a successful install
- **WHEN** an administrator runs `Install and Enable`, the install lifecycle succeeds, and the enable lifecycle fails
- **THEN** the host keeps the plugin in the real `installed but disabled` state
- **AND** the host MUST NOT automatically undo the completed installation because enablement failed
- **AND** the administrator can still trigger enablement again later from the existing installed state
