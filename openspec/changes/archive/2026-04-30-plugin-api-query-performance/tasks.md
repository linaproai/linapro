# Tasks

- [x] Add a read-only plugin list service path that avoids implicit governance synchronization.
- [x] Keep plugin synchronization as an explicit write action and return the synchronized total.
- [x] Add tests proving plugin list queries do not write governance tables.
- [x] Replace unsafe host-service table comment lookup with safe read-only metadata lookup and fallback.
- [x] Add `last_active_time` write throttling for online-session authentication.
- [x] Add backend tests for metadata lookup and session throttling.
- [x] Run affected backend tests and Lina review.
- [x] Record that no i18n resources are needed for this change.
