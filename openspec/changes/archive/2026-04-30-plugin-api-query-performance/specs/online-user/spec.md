## MODIFIED Requirements

### Requirement: Session Activity Time Tracking

The system SHALL track online session last active time for timeout decisions. Authentication MUST check session validity and timeout on every protected request, while writes to `last_active_time` MAY be throttled over a short window to reduce database write pressure.

#### Scenario: Initial active time at login
- **WHEN** a user logs in successfully
- **THEN** `sys_online_session.last_active_time` is initialized to the current time

#### Scenario: Active time write is throttled
- **WHEN** a protected request arrives for a valid session whose `last_active_time` was updated inside the throttle window
- **THEN** authentication still checks the session and timeout
- **AND** the system may skip the database update for that request

#### Scenario: Active time updates after throttle window
- **WHEN** a protected request arrives after the throttle window has elapsed
- **THEN** authentication updates `last_active_time`
- **AND** timeout checks continue to reject expired sessions
