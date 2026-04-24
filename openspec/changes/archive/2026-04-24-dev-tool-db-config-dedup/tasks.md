## 1. Converge development-only configuration

- [x] 1.1 Update `apps/lina-core/hack/config.yaml` to use YAML anchors for the host development-only database connection settings and remove `multiStatements=true`
- [x] 1.2 Review and update any development-only consumers of `hack/config.yaml` so upgrade tooling and local commands still read the unified connection settings correctly

## 2. Rework the local SQL execution path

- [x] 2.1 Implement SQL file splitting and statement-by-statement execution under `apps/lina-core/internal/cmd/` while preserving ordered execution and fail-fast semantics
- [x] 2.2 Adjust error and log context so statement failures still identify the relevant SQL file

## 3. Testing and verification

- [x] 3.1 Add command-layer unit tests that cover multi-statement splitting, comment/blank skipping, semicolons inside strings, and failure interruption
- [x] 3.2 Run the affected Go unit tests and record the results to confirm stable behavior after the development tooling changes
