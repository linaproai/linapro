## MODIFIED Requirements

### Requirement: Query Unread Message Count

The system SHALL provide an endpoint for querying the current user's unread message count. This endpoint SHALL always use the current logged-in user as the sole data boundary, and MUST NOT count other users' messages even if the user's role has all data permission.

#### Scenario: Get Unread Message Count

- **WHEN** caller calls `GET /api/v1/user/message/count`
- **THEN** the system returns the current logged-in user's unread message count `{count: number}`

#### Scenario: No Unread Messages

- **WHEN** the current user has no unread messages
- **THEN** the system returns `{count: 0}`

#### Scenario: All Data Permission Does Not Widen Unread Count Boundary

- **WHEN** the current user's role data scope is all data
- **AND** other users have unread messages
- **THEN** the unread count endpoint still only counts the current logged-in user's own unread messages

### Requirement: Query Message List

The system SHALL provide an endpoint for querying the current user's message list. The message list SHALL always use the current logged-in user as the sole data boundary, and MUST NOT change this boundary regardless of whether the user's role has all data, department data, or self-only data.

#### Scenario: Get Message List

- **WHEN** caller calls `GET /api/v1/user/message` with pagination parameters
- **THEN** the system returns the current user's messages sorted by creation time descending
- **AND** each message contains `id`, `title`, `type`, `sourceType`, `sourceId`, `isRead`, `readAt`, and `createdAt`

#### Scenario: All Data Permission Does Not Read Others' Messages

- **WHEN** the current user's role data scope is all data
- **AND** other users have messages
- **THEN** the message list still does not return other users' messages

### Requirement: Mark Message as Read

The system SHALL provide an endpoint for marking messages as read. The mark-read operation SHALL only act on the current logged-in user's own messages, and MUST NOT be widened to other users' messages by role data scope.

#### Scenario: Mark Single Message as Read

- **WHEN** caller calls `PUT /api/v1/user/message/{id}/read`
- **THEN** the system sets that message's `is_read` to 1 and `read_at` to the current time
- **AND** only the current logged-in user's own messages can be modified

#### Scenario: Mark All Messages as Read

- **WHEN** caller calls `PUT /api/v1/user/message/read-all`
- **THEN** the system marks all of the current logged-in user's unread messages as read

#### Scenario: Reject Marking Others' Messages

- **WHEN** the current user's role data scope is all data
- **AND** requests to mark another user's message as read
- **THEN** the system rejects the operation or treats as not found
- **AND** the target message remains unread

### Requirement: Delete Message

The system SHALL provide an endpoint for deleting messages. The delete message operation SHALL only act on the current logged-in user's own messages, and MUST NOT be widened to other users' messages by role data scope.

#### Scenario: Delete Single Message

- **WHEN** caller calls `DELETE /api/v1/user/message/{id}`
- **THEN** the system physically deletes that message record
- **AND** only the current logged-in user's own messages can be deleted

#### Scenario: Clear All Messages

- **WHEN** caller calls `DELETE /api/v1/user/message/clear`
- **THEN** the system physically deletes all of the current logged-in user's messages

#### Scenario: Reject Deleting Others' Messages

- **WHEN** the current user's role data scope is all data
- **AND** requests to delete another user's message
- **THEN** the system rejects the operation or treats as not found
- **AND** the target message is not deleted
