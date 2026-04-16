## ADDED Requirements

### Requirement: Import dictionary types from Excel
The system SHALL support importing dictionary type records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download dictionary type import template
- **WHEN** user requests GET `/dict/type/import-template`
- **THEN** system returns an Excel template with example data showing required columns: 字典名称、字典类型、状态、备注

#### Scenario: Import dictionary types with valid data
- **WHEN** user uploads a valid Excel file to POST `/dict/type/import`
- **THEN** system validates all rows, creates records, and returns success count

#### Scenario: Import dictionary types with validation errors
- **WHEN** user uploads an Excel file with invalid data (missing required fields, duplicate type values)
- **THEN** system rejects the entire import and returns error details with row numbers and reasons

#### Scenario: Import dictionary types with overwrite mode
- **WHEN** user uploads an Excel file with `updateSupport=true` and the file contains type values that already exist
- **THEN** system updates existing records with the imported values

#### Scenario: Import dictionary types with ignore mode
- **WHEN** user uploads an Excel file with `updateSupport=false` (default) and the file contains type values that already exist
- **THEN** system skips existing records and only creates new records

#### Scenario: Dictionary type import modal UI
- **WHEN** user clicks "导入" button on the dictionary type management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch

### Requirement: Import dictionary data from Excel
The system SHALL support importing dictionary data records from an Excel file. The system SHALL provide a template download endpoint and validate imported data before persisting.

#### Scenario: Download dictionary data import template
- **WHEN** user requests GET `/dict/data/import-template`
- **THEN** system returns an Excel template with example data showing required columns: 字典类型、字典标签、字典键值、排序、Tag样式、CSS类名、状态、备注

#### Scenario: Import dictionary data with valid data
- **WHEN** user uploads a valid Excel file to POST `/dict/data/import`
- **THEN** system validates all rows, creates records, and returns success count

#### Scenario: Import dictionary data with validation errors
- **WHEN** user uploads an Excel file with invalid data (missing required fields, non-existent dict type)
- **THEN** system rejects the entire import and returns error details with row numbers and reasons

#### Scenario: Import dictionary data with overwrite mode
- **WHEN** user uploads an Excel file with `updateSupport=true` and the file contains records with matching dictType+value combination
- **THEN** system updates existing records with the imported values

#### Scenario: Import dictionary data with ignore mode
- **WHEN** user uploads an Excel file with `updateSupport=false` (default) and the file contains records with matching dictType+value combination
- **THEN** system skips existing records and only creates new records

#### Scenario: Dictionary data import modal UI
- **WHEN** user clicks "导入" button on the dictionary data management page
- **THEN** system displays a modal with template download link, drag-and-drop upload area, file type hint (xlsx/xls), and overwrite/ignore mode switch