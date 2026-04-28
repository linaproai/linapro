## MODIFIED Requirements

### Requirement: Session Activity Time Tracking

The system SHALL tracks the last active time of each online user and is used to determine if a session has timed out. The authentication middleware MUST validate the session and timeout on every protected request, but MAY skip writing `last_active_time` when the persisted active time is still within the configured short update window.

#### Scenario: Initial active time at login
- **WHEN** User successfully logged in, system created `sys_online_session` session record
- **THEN** `last_active_time` field MUST be set to the current time

#### Scenario: Validate active session on every protected request
- **WHEN** logged in user with a valid token accessing the protected API
- **THEN** The authentication middleware MUST verify that the session record exists
- **AND** The authentication middleware MUST reject the request if the persisted `last_active_time` has exceeded the effective timeout threshold

#### Scenario: Refresh active time after update window
- **WHEN** logged in user with a valid token accesses a protected API and the persisted `last_active_time` is older than the short update window
- **THEN** The authentication middleware MUST update `last_active_time` to the current time
- **AND** The request is processed normally when the update succeeds

#### Scenario: Skip duplicate active time write within update window
- **WHEN** logged in user with a valid token repeatedly accesses protected APIs within the short update window
- **THEN** The authentication middleware MAY skip the duplicate `last_active_time` update
- **AND** The request is still processed normally if the session has not timed out
