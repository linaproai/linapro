# User Message Specification

## Purpose

Define user message data structure, unread-count query, list query, read handling, deletion, notification bell behavior, message panel behavior, notice detail navigation, and frontend store polling.

## Requirements

### Requirement: User message database table design

The system SHALL provide `sys_user_message` to store user message data.

#### Scenario: sys_user_message table structure

- **WHEN** the `sys_user_message` table structure is inspected
- **THEN** the table includes `id` (BIGINT PK AUTO_INCREMENT), `user_id` (BIGINT recipient user ID), `title` (VARCHAR(255) message title), `type` (TINYINT message type: 1=notification, 2=announcement), `source_type` (VARCHAR(50) source type), `source_id` (BIGINT source ID), `is_read` (TINYINT read flag: 0=unread, 1=read), `read_at` (DATETIME read time), and `created_at` (DATETIME creation time)
- **AND** a composite index is created on `user_id` and `is_read`

### Requirement: Query unread message count

The system SHALL provide an endpoint to query the unread message count for the current user.

#### Scenario: Get unread message count

- **WHEN** the caller invokes `GET /api/v1/user/message/count`
- **THEN** the system returns the unread message count for the current logged-in user as `{count: number}`

#### Scenario: No unread messages

- **WHEN** the current user has no unread messages
- **THEN** the system returns `{count: 0}`

### Requirement: Query message list

The system SHALL provide an endpoint to query the current user's message list.

#### Scenario: Get message list

- **WHEN** the caller invokes `GET /api/v1/user/message` with pagination parameters
- **THEN** the system returns the current user's messages sorted by creation time descending
- **AND** each message includes `id`, `title`, `type`, `sourceType`, `sourceId`, `isRead`, `readAt`, and `createdAt`

### Requirement: Mark messages as read

The system SHALL provide endpoints to mark messages as read.

#### Scenario: Mark a single message as read

- **WHEN** the caller invokes `PUT /api/v1/user/message/{id}/read`
- **THEN** the system sets `is_read` to 1 and `read_at` to the current time for that message
- **AND** only the current logged-in user's own message can be modified

#### Scenario: Mark all messages as read

- **WHEN** the caller invokes `PUT /api/v1/user/message/read-all`
- **THEN** the system marks all unread messages for the current logged-in user as read

### Requirement: Delete messages

The system SHALL provide endpoints to delete messages.

#### Scenario: Delete one message

- **WHEN** the caller invokes `DELETE /api/v1/user/message/{id}`
- **THEN** the system physically deletes that message record
- **AND** only the current logged-in user's own message can be deleted

#### Scenario: Clear all messages

- **WHEN** the caller invokes `DELETE /api/v1/user/message/clear`
- **THEN** the system physically deletes all messages for the current logged-in user

### Requirement: Message notification bell component

The system SHALL provide a message notification bell component in the top navigation bar.

#### Scenario: Bell icon display

- **WHEN** a user signs in and views the top navigation bar
- **THEN** the bell icon is displayed
- **AND** when unread messages exist, an unread-count badge is displayed at the top-right of the icon

#### Scenario: No unread messages

- **WHEN** the user has no unread messages
- **THEN** the bell icon does not display a count badge

#### Scenario: Periodic polling

- **WHEN** the user signs in
- **THEN** the frontend periodically calls the unread-count endpoint to update the badge

### Requirement: Message panel

The system SHALL provide a message panel that opens when the user clicks the bell icon.

#### Scenario: Open message panel

- **WHEN** the user clicks the bell icon
- **THEN** a popover message panel opens and displays the message list
- **AND** each message displays title, message type, and time
- **AND** unread messages have a visual marker such as bold text or a dot

#### Scenario: Click message to navigate to detail

- **WHEN** the user clicks a message in the message panel
- **THEN** the message is automatically marked as read
- **AND** the user navigates to notice detail page `/system/notice/detail/{sourceId}`

#### Scenario: Mark all as read

- **WHEN** the user clicks Mark All Read in the message panel
- **THEN** the system calls the mark-all-read endpoint and updates all message states in the panel

#### Scenario: Clear messages

- **WHEN** the user clicks Clear in the message panel
- **THEN** a confirmation dialog is shown, and after confirmation the clear endpoint is called and the panel becomes empty

#### Scenario: Delete one message

- **WHEN** the user clicks the delete button for a message in the message panel
- **THEN** the system calls the delete-one endpoint and removes that message from the panel

### Requirement: Notice detail page

The system SHALL provide a notice detail page.

#### Scenario: Detail page display

- **WHEN** the user opens `/system/notice/detail/{id}`
- **THEN** the page displays notice title, dictionary-rendered type, creator, and creation time
- **AND** displays notice content as rich text

#### Scenario: Navigate from message panel

- **WHEN** the user clicks a message in the message panel to navigate to detail
- **THEN** that message is automatically marked as read

### Requirement: Message Store in Pinia

The system SHALL provide a Pinia store to manage user message state.

#### Scenario: Initialize polling

- **WHEN** the user signs in and enters the host workspace
- **THEN** the message store starts interval polling to fetch unread message count

#### Scenario: Stop polling

- **WHEN** the user signs out
- **THEN** the message store stops polling

#### Scenario: Unread count updates reactively

- **WHEN** polling fetches a new unread count
- **THEN** `unreadCount` in the store updates reactively, and the bell badge updates accordingly

#### Scenario: Reserve SSE extension point

- **WHEN** the message store polling implementation is inspected
- **THEN** polling logic is encapsulated in an independent method so future SSE listening can replace polling without changing the external store interface

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
