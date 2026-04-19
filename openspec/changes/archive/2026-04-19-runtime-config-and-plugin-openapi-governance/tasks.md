## 1. Change Governance

- [x] 1.1 Review the two completed active changes and confirm that both artifacts and task sets are complete.
- [x] 1.2 Merge `config-runtime-params` and `host-managed-plugin-openapi` into one archived iteration record.
- [x] 1.3 Archive the merged iteration in a single dated archive directory and remove both active change directories.

## 2. Protected Runtime Configuration

- [x] 2.1 Register built-in runtime parameters for JWT expiry, session timeout, upload size, and login IP blacklist.
- [x] 2.2 Register protected public frontend settings for branding, login-page copy, theme mode, layout, and watermark behavior.
- [x] 2.3 Add value validation, protected-key safeguards, SQL seed metadata, and import/update protection for host-owned settings.
- [x] 2.4 Wire runtime configuration into host behavior for authentication, online sessions, file upload, and frontend bootstrap.

## 3. Runtime Cache and Cluster Strategy

- [x] 3.1 Add process-local runtime snapshot caches for protected configuration reads.
- [x] 3.2 Add shared revision synchronization for multi-instance propagation.
- [x] 3.3 Optimize single-node mode to skip unnecessary shared-KV and watcher behavior.
- [x] 3.4 Refactor cache and watcher control flow toward constructor-time strategy wiring and add supporting tests.

## 4. Plugin Governance and API Documentation

- [x] 4.1 Replace direct reliance on GoFrame’s built-in `/api.json` output with a host-managed OpenAPI builder.
- [x] 4.2 Capture source-plugin route ownership and DTO documentation metadata at registration time through a host-observable route facade.
- [x] 4.3 Keep source-plugin middleware registration plugin-owned and unrestricted by route-prefix rules.
- [x] 4.4 Project enabled source-plugin routes and enabled dynamic-plugin routes into the host-managed API document.
- [x] 4.5 Exclude internal and non-business routes from the system API document.

## 5. Plugin UI and Operational Follow-up

- [x] 5.1 Add plugin detail dialog support and refine host-service presentation semantics in the default admin workspace.
- [x] 5.2 Improve plugin resource grouping, labels, empty-state behavior, and layout consistency between detail and authorization dialogs.
- [x] 5.3 Add pagination behavior for the dynamic plugin demo record list.
- [x] 5.4 Add structured logging switch support and align HTTP server logs with business log sinks.
- [x] 5.5 Move host-specific server and logger extensions under explicit `extensions` namespaces.

## 6. Comment Conformance and Review

- [x] 6.1 Review comment coverage for configuration services, permission caches, command boot paths, host-managed OpenAPI code, source-plugin route capture code, and related tests.
- [x] 6.2 Complete backend comment-conformance cleanup across the affected host services and plugin backend samples.
- [x] 6.3 Run targeted Go verification for the touched host packages and plugin backend packages.
- [x] 6.4 Run an archive-time OpenSpec review and confirm that no critical issues remain for the merged iteration.
