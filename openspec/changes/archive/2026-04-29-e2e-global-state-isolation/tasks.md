## 1. Baseline Audit and Conflict Classification

- [x] 1.1 Review the most recent full E2E regression log, catalog shared-state conflict cases, affected test files, root causes, and remediation directions
- [x] 1.2 Scan `hack/tests/e2e` for high-risk operations including plugin lifecycle, system parameters, dictionaries, menu permissions, runtime i18n caches, and public config caches
- [x] 1.3 For each hit, determine the isolation category, whether it must be serial, and whether the conflict can be resolved via fixture or semantic assertion
- [x] 1.4 Record the i18n impact assessment in the change documentation: this change does not introduce product runtime copy; it only governs tests and documentation; i18n cache tests must maintain semantic validation

## 2. Execution Manifest and Runner Governance

- [x] 2.1 Extend `hack/tests/config/execution-manifest.json` to add machine-readable isolation categories for serial entries or high-risk files
- [x] 2.2 Update `hack/tests/scripts/run-suite.mjs` to output parallel file count, serial file count, worker count, and serial isolation category summary in full, smoke, and module modes
- [x] 2.3 Ensure module mode continues to apply the same serial/parallel split so that module regressions cannot bypass isolation rules
- [x] 2.4 Add script-level unit tests or executable verification cases for the runner covering full/module splitting and category summary output

## 3. E2E Validation Gate

- [x] 3.1 Extend `hack/tests/scripts/validate-e2e.mjs` to validate isolation category format, referenced path existence, and serial entry classification completeness
- [x] 3.2 Add high-risk heuristic detection covering plugin install/enable/disable/uninstall/sync/upload/upgrade, system parameter writes, dictionary import/modify, menu permission modifications, and runtime i18n ETag cache assertions
- [x] 3.3 Output actionable error messages for files that hit high-risk patterns but are not in the serial boundary or lack a declared classification
- [x] 3.4 Add an allowlist structure with documented reasons for justified parallel-safe exceptions and enforce reason entry during validation

## 4. Fixture and Test Case Remediation

- [x] 4.1 Consolidate source-plugin state preparation logic to ensure tests depending on plugin pages/APIs/mock data use idempotent fixtures
- [x] 4.2 Review local files, attachments, plugin tables, and mock-data cleanup logic in dynamic-plugin and source-plugin tests to ensure single-file independence
- [x] 4.3 Adjust cache/ETag tests to assert that requests carry conditional headers and accept either `304` or a legitimate `200 + new ETag + body`
- [x] 4.4 Adjust tests that infer business state from localized UI text; switch business assertions to stable IDs, codes, labelKeys, permission keys, or API counters
- [x] 4.5 Perform minimal-scope remediation on known conflict-related files, covering at least `TC0124-runtime-i18n-etag.ts`, plugin lifecycle cases, organization department tree count cases, and content notice dependency cases

## 5. Documentation and Acceptance Materials

- [x] 5.1 Update `hack/tests/README.md` and `hack/tests/README.zh_CN.md` to document isolation categories, serial boundaries, fixture prerequisites, and cache semantic assertion rules
- [x] 5.2 Add an E2E conflict governance record in this change directory listing conflict types, representative cases, remediation approaches, and a checklist for adding new test cases
- [x] 5.3 Confirm that newly added or modified test governance documentation language matches the current change (Chinese); repository-level README additions maintain consistent English/Chinese mirrors

## 6. Verification and Review

- [x] 6.1 Run `cd hack/tests && pnpm test:validate` and confirm E2E naming, directory structure, isolation categories, and high-risk detection all pass
- [x] 6.2 Run script-level verification or related Node tests to confirm that `run-suite.mjs` and `validate-e2e.mjs` new logic is stable
- [x] 6.3 Run the affected E2E subset: `TC0124-runtime-i18n-etag.ts`, plugin lifecycle cases, organization department tree count cases, and content notice dependency cases
- [x] 6.4 Run `cd hack/tests && pnpm test` full regression and record parallel-phase, serial-phase, and skipped-item results
- [x] 6.5 Run `openspec validate e2e-global-state-isolation`
- [x] 6.6 Execute `lina-review`, focusing on E2E isolation boundaries, i18n impact conclusions, fixture independence, documentation Chinese/English consistency, and test coverage

### Verification Notes

- `pnpm test:validate`: Passed; validated 138 E2E files, 28 scopes, 7 smoke files, and 103 serial files.
- `node scripts/run-suite.mjs module i18n --list`: Passed; outputs module split and isolation category summary.
- Affected subset: `TC0066-source-plugin-lifecycle.ts`, `TC0124-runtime-i18n-etag.ts`, `TC0021-user-dept-tree-count.ts`, `TC0037-notice-crud.ts` all passed; `TC0124` passed 5/5 after fix.
- Single-file re-verification: `TC0044c`, `TC0097`, `TC0058f~h`, `TC0012` all passed.
- Full regression: Second round `pnpm test` parallel phase 102/102 passed; serial phase executed 326 assertions, 260 passed, 6 skipped, 4 not run. Actual assertion failures were `TC0044c` and `TC0097`, both fixed and re-verified individually; the remaining 54 failures were cascading `ERR_CONNECTION_REFUSED` errors from a mid-run dev server exit; sampled re-verification of settings/dictionary tail sections passed after restarting the service.
