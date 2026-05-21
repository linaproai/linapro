## 1. Resource and Configuration Cleanup

- [x] 1.1 Delete `zh-TW` default i18n resource directories from host, source plugins, default management workbench, and shared frontend packages
- [x] 1.2 Converge default configuration and frontend i18n static checks to `zh-CN` and `en-US` bilingual
- [x] 1.3 Clean up default text descriptions about trilingual or Traditional Chinese default support

## 2. Test and Verification Adjustments

- [x] 2.1 Remove Traditional Chinese-specific E2E, and adjust general i18n E2E to only cover default bilingual
- [x] 2.2 Adjust frontend and backend unit tests that depend on default `zh-TW` resource assertions
- [x] 2.3 Run OpenSpec, JSON, frontend i18n/typecheck, and related backend test verification

## 3. Review

- [x] 3.1 Execute lina-review, confirm i18n, cache, data permission, API, and test governance conclusions

## Records

- Verified through static scanning of `zh-TW`, `Traditional Chinese` keywords that default resources and test references were cleaned up.
- Frontend `i18n:check` and typecheck confirmed only `zh-CN` alignment with `en-US` is required.
- No new REST APIs, database schema, SQL seed, permission boundaries, or runtime cache mechanisms were added.
