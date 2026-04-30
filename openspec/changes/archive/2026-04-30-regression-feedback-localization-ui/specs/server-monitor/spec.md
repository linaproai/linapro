## ADDED Requirements

### Requirement: Service monitoring disk table must remain readable in English

The service monitoring page SHALL allocate readable disk table column widths in English so key column headers and values do not wrap unnecessarily.

#### Scenario: English disk table keeps key columns on one line
- **WHEN** an administrator opens service monitoring in `en-US`
- **THEN** `File System`, `Total`, `Used`, and `Available` headers do not wrap
