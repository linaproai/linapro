## Context

The current upload default is spread across multiple layers: `apps/lina-core/manifest/sql/007-config-management.sql` initializes `sys.upload.maxSize` to 16, `apps/lina-core/manifest/config/config.template.yaml` also exposes 16, and the static fallback in `apps/lina-core/internal/service/config/config_upload.go` is still 10. The upload chain uses runtime config, config-file fallback, and request-body protection together, so inconsistent defaults make fresh initialization, fallback runtime behavior, and user-facing error messages diverge. The project is still new, so we can normalize the existing default sources directly and validate them through re-initialization instead of carrying compatibility migration baggage.

## Goals / Non-Goals

**Goals:**
- Unify the host upload-size default at 20 MB.
- Make database initialization, config templates, static fallback values, and upload validation all agree on one default.
- Keep automated tests and user-visible error messages aligned with the new baseline.

**Non-Goals:**
- Do not remove the administrator's ability to override `sys.upload.maxSize`.
- Do not redesign the file-upload module or add other runtime parameters.
- Do not change the existing request-body protection scope for non-`multipart` requests.

## Decisions

### 1. Normalize existing default sources directly instead of adding a compatibility migration layer

This change updates the existing host initialization SQL, config template, and backend static fallback directly so 20 MB becomes the only default baseline. That keeps the implementation aligned with the project rule that this is a new project and existing SQL can be updated directly with re-initialization.

- Alternative: add a new SQL iteration file that only overrides the old default value.
- Why not: that would keep default-value ownership split across multiple places and would not fit the current no-compatibility-overhead expectation.

### 2. Treat `sys.upload.maxSize` as the single business source of truth and align every fallback with it

`sys.upload.maxSize` already represents host upload-size governance, so the 20 MB default must live not only in config-management seed data but also in the config template and the static fallback inside `config_upload.go`. That guarantees the database default, direct config-file reads, and upload-chain validation all see the same baseline.

- Alternative: change only the initialization SQL.
- Why not: if only SQL changes, any uninitialized or non-overridden path can still fall back to 10 MB or 16 MB and the split behavior remains.

### 3. Update user-facing error copy and derived artifacts together

Friendly messages shown when uploads exceed the limit, request-body protection assertions, and any embedded or packaged manifest artifact derived from the manifest must all move to 20 MB together. That avoids a state where source code has been updated but build outputs or test baselines still expose the old default.

- Alternative: change only source code and ignore derived artifacts or baseline assertions.
- Why not: that leaves a mismatch between build outputs and the source baseline and makes regressions likely.

## Risks / Trade-offs

- [Risk] Derived package artifacts may still contain the old default. -> Mitigation: update or regenerate embedded resources as part of the implementation and check both source files and derived outputs.
- [Risk] Updating only one source of truth would keep split behavior alive. -> Mitigation: verify SQL seed values, config-template defaults, static fallback constants, and user-facing error assertions one by one.
- [Risk] Local environments that already changed `sys.upload.maxSize` manually will not automatically become 20 MB. -> Mitigation: this change targets the default baseline; validation is done in a clean reinitialized environment or after explicitly resetting the config.

## Migration Plan

- Update the existing default sources directly and verify through local host re-initialization that the initial value of `sys.upload.maxSize` is 20.
- Update and run the affected backend automated tests so upload validation and friendly error messages both use 20 MB as the default baseline.
- If rollback is required, restore the previous defaults and reinitialize the verification environment; no extra upgrade metadata is needed.

## Open Questions

- None.
