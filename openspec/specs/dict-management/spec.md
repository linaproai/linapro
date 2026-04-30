# Dict Management

## Purpose

Define dictionary type and dictionary data management, frontend dictionary rendering, import/export behavior, initialization data, built-in protection, and English layout requirements.

## Requirements

### Requirement: Dictionary type management APIs

The system SHALL provide RESTful APIs to list, create, update, delete, export, and read dictionary types.

#### Scenario: Query dictionary type list
- **WHEN** a caller requests dictionary types with pagination and optional name, type, or status filters
- **THEN** the system returns matching non-deleted dictionary types and total count

#### Scenario: Create or update dictionary type
- **WHEN** a caller submits valid dictionary type fields
- **THEN** the system creates or updates the dictionary type
- **AND** duplicate dictionary type keys are rejected

#### Scenario: Delete dictionary type
- **WHEN** a caller deletes a dictionary type
- **THEN** the system soft-deletes the type according to existing governance rules

### Requirement: Dictionary data management APIs

The system SHALL provide RESTful APIs to list, create, update, delete, export, and query dictionary data by dictionary type.

#### Scenario: Query dictionary data list
- **WHEN** a caller requests dictionary data with optional dict type, label, value, or status filters
- **THEN** the system returns matching non-deleted data rows ordered by sort

#### Scenario: Create or update dictionary data
- **WHEN** a caller submits valid dict type, label, value, sort, tag style, CSS class, status, and remark fields
- **THEN** the system creates or updates the dictionary data row

#### Scenario: Get dictionary options by type
- **WHEN** a caller requests options for a dictionary type
- **THEN** the system returns enabled rows with label, value, tag style, and CSS class

### Requirement: Dictionary tables must support type and data records

The system SHALL provide `sys_dict_type` and `sys_dict_data` tables for dictionary governance.

#### Scenario: Dictionary type table structure
- **WHEN** the dictionary type table is inspected
- **THEN** it includes identifier, name, unique type key, status, remark, timestamps, and soft-delete fields

#### Scenario: Dictionary data table structure
- **WHEN** the dictionary data table is inspected
- **THEN** it includes identifier, dict type key, label, value, sort, tag style, CSS class, status, remark, timestamps, and soft-delete fields

### Requirement: Tag style rendering must support preset and custom styles

The system SHALL provide dictionary tag style configuration with preset colors, custom colors, and CSS class support.

#### Scenario: Preset color rendering
- **WHEN** dictionary data has a preset tag style such as `cyan`, `green`, `orange`, `purple`, `red`, `success`, `warning`, or `default`
- **THEN** `DictTag` renders the matching Ant Design tag color

#### Scenario: Custom color rendering
- **WHEN** dictionary data has a hex color tag style
- **THEN** `DictTag` renders the tag with that custom color

#### Scenario: TagStylePicker component
- **WHEN** a user edits dictionary data tag style
- **THEN** the UI offers default color and custom color modes
- **AND** custom color mode uses a hex color picker without transparency

### Requirement: Dictionary management frontend must use a two-panel layout

The dictionary management page SHALL use a left dictionary-type panel and a right dictionary-data panel.

#### Scenario: Two-panel interaction
- **WHEN** a user opens dictionary management
- **THEN** the left panel shows dictionary types and the right panel shows data for the selected type
- **AND** the right panel starts empty until a type is selected

#### Scenario: Responsive layout
- **WHEN** the page is viewed on desktop
- **THEN** panels display side by side
- **WHEN** the page is viewed on mobile
- **THEN** panels stack vertically

### Requirement: Global dictionary components and store must cache dictionary data

The system SHALL provide a reusable `DictTag` component and Pinia dictionary store so dictionary options are loaded once and reused across pages.

#### Scenario: DictTag renders dictionary value
- **WHEN** `DictTag` receives dictionary options and a value
- **THEN** it displays the matching label and configured tag style
- **AND** unmatched values display fallback text

#### Scenario: Dictionary store deduplicates requests
- **WHEN** multiple components request the same dictionary type concurrently
- **THEN** only one API request is issued
- **AND** all callers share the same result

#### Scenario: Dictionary cache refresh
- **WHEN** a user refreshes dictionary cache
- **THEN** cached dictionary data is cleared and later requests reload it from the API

### Requirement: Dictionary initialization data must be delivered

The system SHALL provide baseline dictionary types and data needed by framework modules, including normal/disabled status and user gender values.

#### Scenario: Initialize common dictionaries
- **WHEN** database initialization runs
- **THEN** common framework dictionary types and data are created idempotently

### Requirement: Dictionary import and export must be supported

The system SHALL provide Excel templates, import flows, and export files for dictionary types and dictionary data.

#### Scenario: Export dictionary data
- **WHEN** a user exports dictionaries
- **THEN** the system returns an Excel file containing dictionary type and dictionary data sheets

#### Scenario: Import dictionary type or data
- **WHEN** a user imports a valid template
- **THEN** the system validates required columns and writes dictionary records according to existing idempotent rules

### Requirement: Dictionary form layout must remain readable in English

Dictionary create and edit forms SHALL provide enough label width in English so labels such as `Dictionary Type` and `Tag Style` do not wrap or harm visual alignment.

#### Scenario: Dictionary type label stays readable
- **WHEN** an administrator opens dictionary type add or edit form in `en-US`
- **THEN** `Dictionary Type` remains readable and is not forced into an awkward line break
- **AND** input fields align clearly with labels

#### Scenario: Dictionary data tag style label stays readable
- **WHEN** an administrator opens dictionary data add or edit form in `en-US`
- **THEN** `Tag Style` remains readable
- **AND** the tag style selector does not overlap other fields

### Requirement: Tag Style dropdown must show readable localized options

The dictionary data form Tag Style dropdown SHALL display human-readable option text in the current language and MUST NOT expose runtime i18n keys.

#### Scenario: English tag style dropdown labels
- **WHEN** an administrator opens the Tag Style dropdown in `en-US`
- **THEN** options display labels such as `Default`, `Primary`, and `Success`
- **AND** no raw `pages.*` or `component.*` key is shown

### Requirement: Built-in dictionary types and data must be editable but not deletable

System-owned dictionary types and dictionary data SHALL be marked built-in. They remain editable where allowed, but deletion MUST be blocked in both frontend and backend.

#### Scenario: Built-in dictionary type delete action is disabled
- **WHEN** an administrator views a built-in dictionary type
- **THEN** the delete action is disabled and does not open confirmation
- **AND** hover text explains that built-in data cannot be deleted
- **AND** edit remains available

#### Scenario: Built-in dictionary data delete action is disabled
- **WHEN** an administrator views built-in dictionary data
- **THEN** the delete action is disabled and does not open confirmation
- **AND** edit remains available

#### Scenario: Backend rejects built-in dictionary deletion
- **WHEN** a caller bypasses the frontend and requests deletion of built-in dictionary type or data
- **THEN** the backend returns a structured business error and preserves the record
- **AND** non-built-in dictionary records remain deletable under existing rules

