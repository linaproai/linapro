## ADDED Requirements

### Requirement: New Active Change Artifacts Follow User Language
The system SHALL generate new active change artifacts in the user's current request language unless the user explicitly requests another language.

#### Scenario: Chinese request context creates a new active change
- **WHEN** the user requests a new active change primarily in Simplified Chinese
- **THEN** the generated proposal, design, tasks, and delta specs use Simplified Chinese

#### Scenario: English request context creates a new active change
- **WHEN** the user requests a new active change primarily in English
- **THEN** the generated proposal, design, tasks, and delta specs use English

### Requirement: Archived Change Documents Use English
The system SHALL archive change documents and archived delta specs in English, regardless of the current conversation language.

#### Scenario: Archive is executed from a Chinese conversation
- **WHEN** a completed change is archived from a Chinese conversation context
- **THEN** the archived proposal, design, tasks, and archived delta specs are written in English
- **AND** any synced baseline spec updates introduced by the archive are written in English
