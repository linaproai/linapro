## ADDED Requirements

### Requirement: Dictionary capability must return localized names and labels for the current language
The system SHALL return localized dictionary names, labels, and related descriptions in dictionary type lists, dictionary data lists, option-by-type APIs, and dictionary tag display data APIs according to the current request language. Dictionary i18n MUST derive translation keys from stable `dict_type` and `value` anchors.

#### Scenario: Dictionary type list returns localized names
- **WHEN** a user queries the dictionary type list with `en-US`
- **THEN** dictionary names in the response use localized values for that language
- **AND** pagination, filtering, and sorting behavior remain unchanged

#### Scenario: Dictionary options return localized labels
- **WHEN** the frontend calls the option-by-dictionary-type API with the current language
- **THEN** the returned `label` field is the localized label for the current language
- **AND** business `value`, status, and sort order remain unaffected

### Requirement: Dictionary management pages and business display must share the same dictionary translation result
The system SHALL let dictionary management pages, business-page `DictTag` rendering, and other dictionary option consumers share the same backend-localized results so the same dictionary does not display different language text across pages.

#### Scenario: DictTag renders localized labels
- **WHEN** a business page renders `DictTag` from dictionary option data
- **THEN** `DictTag` displays the dictionary label for the current language
- **AND** tag style, CSS class, and status logic keep their existing behavior

#### Scenario: Missing dictionary translations fall back to the default language
- **WHEN** the current language lacks a translation for a dictionary label
- **THEN** the system falls back to the default-language label or original label value
- **AND** the dictionary item can still be selected and rendered normally
