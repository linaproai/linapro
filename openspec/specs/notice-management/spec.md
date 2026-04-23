# Notice Management

## Purpose

Define the notification announcement data structure, list query, detail maintenance and message distribution behavior provided by the `content-notice` source plugin to ensure that the announcement content can be managed by the plugin and delivered to the target user through the host notification domain.
## Requirements
### Requirement: Notification and announcement database table design
The system SHALL provides the `plugin_content_notice` table to store notification announcement data.

#### Scenario: plugin_content_notice table structure
- **WHEN** View `plugin_content_notice` table structure
- **THEN** table contains: `id` (BIGINT PK AUTO_INCREMENT), `title` (VARCHAR(255) title), `type` (TINYINT type: 1=notification 2=announcement), `content` (LONGTEXT rich text content), `file_ids` (VARCHAR(500) attachment file ID list), `status` (TINYINT status: 0=draft 1=published), `remark` (VARCHAR(500) remarks), `created_by` (BIGINT creator ID), `updated_by` (BIGINT updater ID), `created_at` (DATETIME), `updated_at` (DATETIME), `deleted_at` (DATETIME soft deletion)

### Requirement: Notification announcement list query
System SHALL provides a paging list query interface for notification announcements.

#### Scenario: Query notification announcement list
- **WHEN** Call `GET /api/v1/notice` and pass in the paging parameters `pageNum` and `pageSize`
- **THEN** returns the notification announcement list and total number in the format of `{list: [...], total: number}`
- **THEN** The list is sorted in reverse order of creation time

#### Scenario: The notification list supports conditional filtering
- **WHEN** Pass in the filter parameters `title` (title), `type` (type) or `createdBy` (creator) when querying
- **THEN** `title` uses fuzzy matching (LIKE), `type` uses exact matching
- **THEN** `createdBy` matches the creator username through the associated user table (fuzzy matching)
- **THEN** Returns the list of eligible notification announcements

#### Scenario: Notification announcement list excludes deleted records
- **WHEN** Query notification announcement list
- **THEN** Soft deleted records are not included in the results

#### Scenario: List returns the name of the creator
- **WHEN** Query notification announcement list
- **THEN** Each record contains the `createdByName` field, which is the user name of the creator

### Requirement: Get notification announcement details
System SHALL provides an interface for querying notification and announcement details.

#### Scenario: Query notification announcement details
- **WHEN** calls `GET /api/v1/notice/{id}`
- **THEN** Returns the complete information of the notification announcement, including rich text content

#### Scenario: Query non-existent notification announcements
- **WHEN** calls `GET /api/v1/notice/{id}` and the ID does not exist
- **THEN** The system returns an error message

### Requirement: Create notification announcement
System SHALL provides an interface for creating notification announcements.

#### Scenario: Create notification announcement successfully
- **WHEN** Call `POST /api/v1/notice` and submit the `title`, `type`, `content`, `status` fields
- **THEN** The system creates notification announcements and automatically records `created_by` as the current logged in user ID
- **THEN** returns success

#### Scenario: Create and publish notifications directly
- **WHEN** The notification announcement was created with `status` as 1 (published)
- **THEN** After the system creates the notification announcement, it creates the notification master record and inbox delivery record for the target user through the host unified `notify` service.

#### Scenario: Create draft notifications
- **WHEN** Create notification announcement with `status` as 0 (draft)
- **THEN** Only creates notification announcement records and does not distribute user messages

#### Scenario: Required field verification
- **WHEN** Missing `title`, `type` or `content` when creating notification announcement
- **THEN** The system returns parameter verification error

### Requirement: Update notification announcement
System SHALL provides an interface for update notification announcements.

#### Scenario: Update notification announcement successful
- **WHEN** Call `PUT /api/v1/notice/{id}` and submit the fields to be updated
- **THEN** The system updates corresponding notification and announcement information, and automatically records `updated_by` as the current logged in user ID.

#### Scenario: Draft updated to published
- **WHEN** Change `status` from 0 to 1 when updating notification announcements
- **THEN** After the system updates the notification announcement status, it creates a notification master record and inbox delivery record for the target user through the host unified `notify` service.

#### Scenario: Notification posted for editing again
- **WHEN** Update the content of a published notification announcement (do not change status)
- **THEN** Only updates notification announcement records and does not repeatedly distribute user messages

#### Scenario: Update non-existent notification announcements
- **WHEN** Update a non-existent notification announcement ID
- **THEN** The system returns an error message

### Requirement: Delete notification announcement
System SHALL provides an interface for deleting notification announcements and supports batch deletion.

#### Scenario: Deletion notification announcement successful
- **WHEN** Call `DELETE /api/v1/notice` and pass in the `ids` parameter (comma separated list of IDs)
- **THEN** The corresponding notification announcement is soft deleted (set `deleted_at`)

### Requirement: Notification announcement dictionary data
System SHALL provides dictionary data related to notification announcements.

#### Scenario: Initialize notification type dictionary
- **WHEN** Execute v0.4.0 database migration script
- **THEN** Create dictionary type `sys_notice_type` (notification type), including dictionary data: notification (1), announcement (2)

#### Scenario: Initialize announcement status dictionary
- **WHEN** Execute v0.4.0 database migration script
- **THEN** Create dictionary type `sys_notice_status` (notice status), including dictionary data: draft (0), published (1)

### Requirement: Notification and announcement management frontend list page
System SHALL provides a notification and announcement management list page.

#### Scenario: List page display
- **WHEN** The user enters the notification and announcement management page
- **THEN** Display the notification announcement list in VXE-Grid form, supporting paging
- **THEN** Display columns: announcement title, announcement type (dictionary rendering), status (dictionary rendering), creator, creation time
- **THEN** supports multiple selection of check boxes

#### Scenario: Search filters
- **WHEN** The user enters the title, selects the type or enters the creator in the search bar and clicks search
- **THEN** The table refreshes to display eligible notification announcements

#### Scenario: New notification announcement
- **WHEN** User clicks the "Add" button
- **THEN** Pop-up window (800px wide), including title, status (RadioButton), type (RadioButton), content (Tiptap editor) fields

#### Scenario: Editorial Notice Announcement
- **WHEN** The user clicks the "Edit" button of a record
- **THEN** Pops up a pop-up window and echoes the notification announcement information. Submit the update after modification.

#### Scenario: Deletion notification announcement
- **WHEN** The user clicks the "Delete" button of a record
- **THEN** A confirmation dialog box pops up. After confirmation, the notification announcement is deleted and the list is automatically refreshed.

#### Scenario: Batch delete
- **WHEN** The user selects multiple records and clicks the "Delete" button on the toolbar.
- **THEN** A confirmation dialog box will pop up. After confirmation, the selected notification announcements will be deleted in batches.

### Requirement: Notification announcement menu and permissions
The system SHALL mounts the notification announcement menu as the `content-notice` source plugin menu to the host `content management` directory instead of to `system management`.

#### Scenario: Menu display
- **WHEN** `content-notice` is installed, enabled and the current user has menu access
- **THEN** `Content Management` displays the `Notification Announcement` menu item under the group
- **AND** Plugin governance is still the responsibility of `Extension Center / Plugin Management`

#### Scenario: Plug-in missing or disabled
- **WHEN** `content-notice` is not installed, not enabled, or the current user does not have access to its menu
- **THEN** The host does not display the `Notification Announcement` menu entry
- **AND** If there are no other visible submenus in `Content Management`, the parent directory will be hidden as well.

### Requirement: Notification announcements are delivered by the content source plugin

The The system SHALL deliver the notification announcement capability as a `content-notice` source plugin, rather than continuing as the host's default built-in module.

#### Scenario: Provides notification announcement capability when the content plugin is enabled
- **WHEN** `content-notice` is installed and enabled
- **THEN** The host exposes notification announcement related APIs, pages and menus
- **AND** This plugin continues to host the announcement content management and publishing process

#### Scenario: Hide the notification announcement entrance when the content plugin is missing
- **WHEN** `content-notice` is not installed or enabled
- **THEN** The host does not display the notification announcement menu and page entry
- **AND** The remaining core capabilities of the host continue to operate normally

