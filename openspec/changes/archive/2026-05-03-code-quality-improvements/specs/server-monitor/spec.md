## ADDED Requirements

### Requirement: Server monitor frontend must support visibility-aware automatic refresh

The server monitor frontend page SHALL start periodic polling after mount to refresh latest metrics. The default refresh interval is `30s`. Polling MUST listen to page visibility events: when the tab is hidden (`document.visibilityState === 'hidden'`), polling MUST pause; when the page becomes visible again, the page MUST refresh immediately once and resume periodic polling. On component unmount, polling MUST be explicitly stopped to avoid memory leaks.

#### Scenario: Page refreshes on schedule while visible

- **WHEN** the user stays on the server monitor page
- **AND** the tab is visible
- **THEN** the frontend MUST request the latest monitoring metrics every 30 seconds
- **AND** the UI updates reactively

#### Scenario: Polling pauses when the tab is hidden

- **WHEN** the user switches to another tab
- **THEN** the server monitor page MUST stop periodic requests
- **AND** no new network requests are produced

#### Scenario: Page refreshes immediately when visible again

- **WHEN** the user switches back to the server monitor page
- **THEN** the frontend MUST trigger one refresh immediately
- **AND** resume periodic polling

#### Scenario: Polling stops on component unmount

- **WHEN** the user leaves the server monitor page through route change or component unmount
- **THEN** the frontend MUST explicitly stop the polling timer and send no further requests
