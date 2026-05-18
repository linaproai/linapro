## ADDED Requirements

### Requirement: Directory-level READMEs must use bilingual mirrors
The system SHALL require all directory-level main documentation in the repository to simultaneously provide an English primary document `README.md` and a Chinese mirror document `README.zh-CN.md`.

#### Scenario: Directory already has a main README
- **WHEN** a directory exists with main documentation
- **THEN** the directory simultaneously has English `README.md` and Chinese `README.zh-CN.md`
- **AND** `README.md` is written in English
- **AND** `README.zh-CN.md` is written in Chinese

#### Scenario: Adding new directory documentation
- **WHEN** a developer adds main documentation for a directory
- **THEN** both `README.md` and `README.zh-CN.md` are created in the same change
- **AND** creating only a single-language directory-level README is not allowed

### Requirement: Bilingual README content must remain synchronized
The system SHALL require the English primary document and Chinese mirror document to maintain consistent structure and information, differing only in language.

#### Scenario: Updating English README
- **WHEN** a developer updates a directory's `README.md`
- **THEN** the corresponding `README.zh-CN.md` is updated in the same change
- **AND** both documents maintain the same section structure and technical facts

#### Scenario: Repository root must provide bilingual entry README
The system SHALL provide an English primary entry README and Chinese mirror README in the repository root directory as the unified project entry documentation.

- **WHEN** a user accesses the repository root directory
- **THEN** they can see English `README.md` and Chinese `README.zh-CN.md`
- **AND** both documents can completely introduce project positioning, directory structure, development methods, and key specification entry points
