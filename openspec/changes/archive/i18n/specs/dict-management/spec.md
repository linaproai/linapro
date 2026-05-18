## ADDED Requirements

### Requirement: Dictionary capability must return localized names and labels for the current language
The system SHALL return localized dictionary names, labels, and related descriptions in dictionary type lists, dictionary data lists, option-by-type APIs, and dictionary tag display data APIs according to the current request language. Dictionary i18n MUST derive translation keys from stable `dict_type` and `value` anchors.

#### Scenario: Dictionary type list returns localized names
- **WHEN** a user queries the dictionary type list with `en-US`
- **THEN** dictionary names in the response use localized values for that language

#### Scenario: Dictionary options return localized labels
- **WHEN** the frontend calls the option-by-dictionary-type API with the current language
- **THEN** the returned `label` field is the localized label for the current language

### Requirement: Dictionary management pages and business display must share the same dictionary translation result
The system SHALL let dictionary management pages, business-page `DictTag` rendering, and other dictionary option consumers share the same backend-localized results.

#### Scenario: DictTag renders localized labels
- **WHEN** a business page renders `DictTag` from dictionary option data
- **THEN** `DictTag` displays the dictionary label for the current language

### Requirement: Dictionary form layout must remain readable in English

Dictionary create and edit forms SHALL provide enough label width in English so labels such as `Dictionary Type` and `Tag Style` do not wrap awkwardly.

#### Scenario: English dictionary labels stay readable
- **WHEN** an administrator opens dictionary forms in `en-US`
- **THEN** long labels remain readable and aligned

### Requirement: Tag Style dropdown must show readable localized options

The dictionary data form Tag Style dropdown SHALL display human-readable option text in the current language and MUST NOT expose runtime i18n keys.

#### Scenario: English tag style dropdown labels
- **WHEN** an administrator opens the Tag Style dropdown in `en-US`
- **THEN** options display labels such as `Default`, `Primary`, and `Success`

### Requirement: Built-in dictionary types and data must be editable but not deletable

System-owned dictionary types and dictionary data SHALL be marked built-in. They remain editable where allowed, but deletion MUST be blocked in both frontend and backend.

#### Scenario: Backend rejects built-in dictionary deletion
- **WHEN** a caller requests deletion of built-in dictionary type or data
- **THEN** the backend returns a structured business error and preserves the record
