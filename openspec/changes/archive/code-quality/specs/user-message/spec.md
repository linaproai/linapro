## ADDED Requirements

### Requirement: User-message polling must be page-visibility aware

Unread-count polling for user messages SHALL listen to page visibility events. When `document.visibilityState === 'hidden'`, polling MUST pause. When the page becomes visible again, the message store MUST immediately refresh unread count once and resume periodic polling. The underlying timer MUST be explicitly stopped when the user logs out or the store is disposed.

#### Scenario: Unread-count polling pauses while page is hidden

- **WHEN** the user switches to another tab
- **THEN** the message store MUST pause unread-count polling
- **AND** no `GET /api/v1/user/message/count` requests are produced

#### Scenario: One refresh runs immediately when the page is visible again

- **WHEN** the user switches back to the current application
- **THEN** the message store MUST immediately trigger one unread-count refresh
- **AND** periodic polling resumes afterward

#### Scenario: Polling timer stops on logout

- **WHEN** the user logs out
- **THEN** the message store MUST explicitly stop the polling timer
- **AND** no unread-count requests are produced afterward
