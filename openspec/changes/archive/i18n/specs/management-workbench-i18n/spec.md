## MODIFIED Requirements

### Requirement: Management workbench first-start language must be automatically identified by browser language

The default management workbench SHALL automatically select the startup language based on the browser's preferred language when no user-saved language preference exists and initialization parameters do not explicitly specify a language. If the browser's preferred language is a Chinese language tag (such as `zh`, `zh-CN`, `zh-TW`, `zh-Hans-CN`), the workbench defaults to `zh-CN`; otherwise it defaults to `en-US`. This automatic identification only affects the first-start default value, must not override language preferences saved by the user through the language switcher or preference settings, and must not override the language explicitly specified by the caller through initialization `overrides.app.locale`. Default delivery no longer provides `zh-TW` static language packs, and Chinese browser language tags must uniformly map to the default built-in Chinese `zh-CN`.

#### Scenario: Chinese browser first visit defaults to Simplified Chinese
- **WHEN** the user's browser language is `zh-CN` or `zh-TW`
- **AND** no saved `preferences` exist in the current workbench namespace
- **AND** initialization parameters do not explicitly specify `app.locale`
- **THEN** the workbench startup language is `zh-CN`
- **AND** login page, runtime language pack requests, and public frontend configuration requests all use `zh-CN`

#### Scenario: Non-Chinese browser first visit defaults to English
- **WHEN** the user's browser language is `en-US`
- **AND** no saved `preferences` exist in the current workbench namespace
- **AND** initialization parameters do not explicitly specify `app.locale`
- **THEN** the workbench startup language is `en-US`
- **AND** login page, runtime language pack requests, and public frontend configuration requests all use `en-US`

#### Scenario: Saved language preference takes priority over browser language
- **WHEN** a saved language preference of `zh-CN` exists in the current workbench namespace
- **AND** the user's browser language is `en-US`
- **THEN** the workbench startup language remains `zh-CN`
- **AND** the browser language does not override the user's saved selection
