## 1. Menu Skeleton and Host Directory Management

- [x] 1.1 Design and confirm the new default backend first-level directory structure and stable parent `menu_key`
- [x] 1.2 Update the host menu initialization SQL and create stable directories such as `dashboard`, `iam`, `org`, `setting`, `content`, `monitor`, `scheduler`, `extension`, `developer`
- [x] 1.3 Adjust the host menu projection logic to only consume host stable directory records and support the new first-level directory structure
- [x] 1.4 Implement automatic hiding rules for empty parent directories
- [x] 1.5 Backend testing and frontend routing assembly verification after menu reconstruction

## 2. Plugin List and Mounting Rule Management

- [x] 2.1 Solidify official source plugin ID: `org-center`, `content-notice`, `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog`, `demo-control`
- [x] 2.2 Add verification to the source plugin menu mounting rules, restricting them to the host stable directory or plugin internal nodes
- [x] 2.3 Remove the `plugin-` prefix examples in plugin documents and samples, and change them to domain-capability naming
- [x] 2.4 Supplement unit testing for plugin management menu synchronization and parent mounting

## 3. Host Core Boundary Extraction

- [x] 3.1 Inventory and mark the code boundaries between host retained capabilities and capabilities to be plugged in
- [x] 3.2 Define and implement a unified login event contract and publisher for the host
- [x] 3.3 Integrate login success, login failure, and logout success into unified login events, and remove the direct dependence of the authentication link on the specific login log implementation
- [x] 3.4 Define and implement a unified audit event contract and publisher for the host
- [x] 3.5 Integrate write operations and queries with the `operLog` tag into unified audit events, and remove the middleware's direct dependence on the implementation of specific operation logs
- [x] 3.6 Separate the organizational capability interface to avoid user management directly relying on departments/positions for implementation
- [x] 3.7 Split the boundary between "authentication session kernel" and "online user governance capabilities" and clarify that the host is responsible for the session truth source, active time maintenance, timeout determination, and cleanup

## 4. Monitoring Capability Source Plugins

- [x] 4.1 Create source plugin `monitor-operlog`
- [x] 4.2 Migrate operation log query, details, export, cleaning, and pages to `monitor-operlog`
- [x] 4.3 Create source plugin `monitor-loginlog`
- [x] 4.4 Migrate login log query, details, export, cleaning, and pages to `monitor-loginlog`
- [x] 4.5 Create source plugin `monitor-server`
- [x] 4.6 Migrate service monitoring collection, cleaning, storage, query, and page to `monitor-server`
- [x] 4.7 Create source plugin `monitor-online`
- [x] 4.8 On the premise of retaining the host session kernel, migrate online user query and forced offline management to `monitor-online`
- [x] 4.9 Add installation, activation, deactivation, uninstallation, and menu mounting verification for the 4 monitoring plugins

## 5. Organization and Content Capability Source Plugins

- [x] 5.1 Create source plugin `org-center`
- [x] 5.2 Migrate department management to `org-center`
- [x] 5.3 Migrate position management to `org-center`
- [x] 5.4 Implement user management UI and interface degradation when the organization plugin is missing
- [x] 5.5 Create source plugin `content-notice`
- [x] 5.6 Migrate notification announcement capabilities to `content-notice`

## 6. Demo Control Source Plugin

- [x] 6.1 Add `plugin.autoEnable` examples to the host config template and make it clear that enabling `demo-control` controls the demo-mode switch
- [x] 6.2 Update the plugin workspace wiring entry so `demo-control` can be discovered, installed, and enabled through `plugin.autoEnable`
- [x] 6.3 Add the base source-plugin structure, manifest, embed entry, and documentation for `apps/lina-plugins/demo-control/`
- [x] 6.4 Implement a demo-control service based on global HTTP middleware that blocks write requests across `/*` by HTTP Method while preserving the minimal session and plugin-governance whitelist
- [x] 6.5 Ensure the demo-control plugin does not depend on any extra host boolean config and only takes effect while the plugin is enabled

## 7. Plugin Explicit Wiring and Delivery Documentation

- [x] 7.1 Update `apps/lina-plugins/lina-plugins.go` to provide explicit wiring entry for official source plugins
- [x] 7.2 Supplement `README.md` and `README.zh-CN.md` for each official source plugin
- [x] 7.3 Update `apps/lina-plugins/README.md` and `README.zh-CN.md` to explain the host/plugin boundary and official plugin list
- [x] 7.4 Supplement operation and maintenance instructions for plugin installation, start and stop, menu mounting, and uninstall cleaning

## 8. Frontend Routing and Menu Visibility Linkage

- [x] 8.1 Adjust frontend static routing and dynamic menu adaptation to match the new first-level directory structure
- [x] 8.2 Implement menu and dynamic routing refresh convergence after plugin startup and shutdown
- [x] 8.3 Implement the hiding of fields on the user management page when the organization plugin is missing
- [x] 8.4 Implement system monitoring empty directory hiding when the monitoring plugin is missing
- [x] 8.5 Implement the ability to detect degradation in the department/position fields of user lists, details, and edit drawers when the organization plugin is missing

## 9. E2E Testing

- [x] 9.1 Use `lina-e2e` to plan E2E use cases after menu reconstruction and pluginization
- [x] 9.2 Add menu skeleton and empty parent directory hiding test
- [x] 9.3 Add `monitor-operlog` lifecycle and functional verification test (`TC0098a` + existing monitoring function use cases)
- [x] 9.4 Add `monitor-loginlog` lifecycle and functional verification test (`TC0098b` + existing monitoring function use cases)
- [x] 9.5 Add `monitor-server` lifecycle and functional verification test (`TC0098c` + existing monitoring function use cases)
- [x] 9.6 Add `monitor-online` lifecycle and functional verification test (`TC0098d` + existing monitoring function use cases)
- [x] 9.7 Add `org-center` lifecycle and user management downgrade tests (`TC0098e`, `TC0081`)
- [x] 9.8 Add `content-notice` lifecycle tests (`TC0098f`, `TC0037`)
- [x] 9.9 Add regression test that login, authentication, and session timeout still work when `monitor-online` is missing or disabled (`TC0099a`)
- [x] 9.10 Add regression test that the login process and normal business requests are still normal when the log plugin is missing or disabled (`TC0099b`)
- [x] 9.11 Add `plugin.autoEnable` config tests that cover the default-disabled case and explicit enablement of `demo-control`
- [x] 9.12 Add middleware tests that cover query allow, write rejection, `/*` global scope, login whitelist behavior, and passthrough while the plugin is disabled
- [x] 9.13 Add `TC0105` E2E case to cover `plugin.autoEnable` enablement of `demo-control`, login/logout whitelist behavior, query passthrough, and `/*` global write interception
- [x] 9.14 Run the relevant E2E regression and log the results

## 10. Verification and Review

- [x] 10.1 Run host and plugin related backend unit tests
- [x] 10.2 Run frontend type checking and build verification
- [x] 10.3 Run E2E suite related to plugin start, stop, menu refresh, and demo control
- [x] 10.4 Check whether the host still only retains core capabilities against the specifications
- [x] 10.5 Call `lina-review` for change review

## 11. Pluginbridge Subcomponent Restructuring

- [x] 11.1 Create `pkg/pluginbridge/{contract,codec,artifact,hostcall,hostservice,guest}` subcomponent directories with compliant package-level and file-level documentation comments
- [x] 11.2 Define subcomponent dependency direction; migrate low-level contract/artifact/codec capabilities first, ensuring no subcomponent imports the root `pluginbridge` package
- [x] 11.3 Move protobuf wire, WASM section low-level reading, and other pure implementation details into corresponding subcomponent `internal` packages, avoiding new cross-domain `util/common/helper` packages
- [x] 11.4 Migrate `BridgeSpec`, `RouteContract`, `BridgeRequestEnvelopeV1`, `BridgeResponseEnvelopeV1`, `IdentitySnapshotV1`, `CronContract`, `ExecutionSource`, and other stable contracts to the `contract` subcomponent
- [x] 11.5 Migrate bridge request/response/route/identity/HTTP snapshot codec to the `codec` subcomponent, preserving existing round-trip tests
- [x] 11.6 Migrate WASM section constants, `RuntimeArtifactMetadata`, `ReadCustomSection`, `ListCustomSections` to the `artifact` subcomponent, updating i18n, apidoc, and runtime call paths
- [x] 11.7 Add facade and subcomponent consistency tests covering bridge envelope and WASM section representative entry points
- [x] 11.8 Migrate host call opcode, status codes, `HostCallResponseEnvelope`, and generic host call codec to the `hostcall` subcomponent
- [x] 11.9 Migrate `HostServiceSpec`, capability derivation, host service manifest codec, and service/method constants to the `hostservice` subcomponent
- [x] 11.10 Migrate runtime, storage, network, data, cache, lock, config, notify, cron host service payload codec to the `hostservice` subcomponent, preserving field numbering and default value semantics
- [x] 11.11 Update Wasm host function, runtime, and plugindb host code to prefer importing `hostcall`/`hostservice`/`codec` and other precise subcomponents
- [x] 11.12 Migrate guest runtime, guest controller dispatcher, context response helper, `BindJSON`/`WriteJSON`, `ErrorClassifier` to the `guest` subcomponent
- [x] 11.13 Migrate guest host service client helpers to the `guest` subcomponent, preserving `Runtime()`, `Storage()`, `HTTP()`, `Data()`, `Cache()`, `Lock()`, `Config()`, `Notify()`, `Cron()` compatibility entry points
- [x] 11.14 Converge the root `pluginbridge` package into a thin facade using type aliases, const aliases, and wrapper functions delegating to subcomponents; constrain root-directory production source files to 1-3 files
- [x] 11.15 Update dynamic plugin samples or add compatibility tests ensuring both root-package old entry points and `guest` subcomponent entry points compile and function correctly
- [x] 11.16 Run and fix `go test ./pkg/pluginbridge/...`
- [x] 11.17 Run and fix plugin runtime, WASM host function, and plugindb related tests: `go test ./internal/service/plugin/internal/runtime/... ./internal/service/plugin/internal/wasm/... ./pkg/plugindb/...`
- [x] 11.18 Execute ordinary Go tests and `GOOS=wasip1 GOARCH=wasm go build ./...` on `apps/lina-plugins/plugin-demo-dynamic`
- [x] 11.19 Run `openspec validate pluginbridge-subcomponent-refactor --strict` to ensure proposal, design, specs, and tasks are archivable
- [x] 11.20 Record i18n impact assessment: this change does not add, modify, or delete runtime language packs, plugin manifest i18n resources, or apidoc i18n JSON resources
- [x] 11.21 Record cache consistency assessment: this change does not add business caches or alter the authoritative data sources, cache keys, invalidation triggers, cross-instance synchronization mechanisms, or failure degradation strategies for plugin runtime cache, i18n resource cache, or WASM compilation cache
- [x] 11.22 Call `lina-review` for code and specification review

## Feedback

- [x] **FB-1**: Change the collaboration between host and plugin to stabilize the capability seam, and remove scattered plugin occupancy judgment and high-coupling branches.
- [x] **FB-2**: The 6 official source plugins must be completely implemented in `apps/lina-plugins/<plugin-id>/`, and aligned with the `plugin-demo-source` directory structure and explicit wiring method.
- [x] **FB-3**: Remove the `pkg` bridge module that is not used by any plugin or host code to avoid exposing invalid public interfaces.
- [x] **FB-4**: Recycle host private menu mount keys and official plugin management constants from `apps/lina-core/pkg/` to `internal/` to avoid `pkg` from carrying host management rules.
- [x] **FB-5**: `orgcap` capability seams are not sealed -- `internal/service/orgcap/orgcap.go` directly imports `deptsvc` and exposes `[]*deptsvc.TreeNode`; need to let `orgcap` own `DeptTreeNode` and let the host business layer completely stop importing `deptsvc`.
- [x] **FB-6**: Clean up residual artifacts -- remove empty directory `apps/lina-core/pkg/officialplugin/` and commented-only SQL stubs.
- [x] **FB-7**: Plugin frontend realization -- 6 plugins `frontend/pages/*.vue` are currently thin wrappers; need to physically migrate pages, pop-up windows, and client APIs to each plugin `frontend/`.
- [x] **FB-8**: The `api/`, `internal/controller/` and `internal/service/` of `dept/post/notice/loginlog/operlog/servermon` are still in `apps/lina-core/`; need to migrate into each plugin `backend/api` and `backend/internal/{controller,service}`.
- [x] **FB-9**: Explicit host boundary exception for `orgcap` -- the org capability seam continues to install by the host consumer plugin into `sys_dept/sys_post/sys_user_dept/sys_user_post` tables as a read-only/associative source of truth, but the exception must be explicitly declared.
- [x] **FB-10**: Stable `pkg/pluginservice/session` external contract -- remove the direct alias publishing of the `internal/service/session` type, replace it with the host's own independent DTO.
- [x] **FB-11**: Block the default fallback of the user page to the organization plugin frontend API.
- [x] **FB-12**: The system information page E2E needs to follow the backend dynamic metadata to verify current component descriptions.
- [x] **FB-13**: `TC0021c` depends on the fixed `testuser`; needs to be changed to a self-created user in the test and cleaned up.
- [x] **FB-14**: The backend database access of official source plugins needs to be uniformly switched to the `dao/do/entity` generated by the plugin's local `gf gen dao`.
- [x] **FB-15**: `middleware_request_body_limit` still uses bare `g.Map` for unified error response; needs to be changed to `ghttp.DefaultHandlerResponse`.
- [x] **FB-16**: Remove the DML for organization, announcement, and other source plugin business tables in `apps/lina-core/manifest/sql/mock-data/`.
- [x] **FB-17**: Delete the host's `dao/do/entity` and direct table lookup logic for the organization plugin business table; provide implementation by `org-center` through the stable capability provider.
- [x] **FB-18**: Official source plugin business tables will be uniformly migrated to the `plugin_<plugin_id_snake_case>_` scope prefix.
- [x] **FB-19**: The official organization source plugin naming is adjusted from `org-management` to `org-center`.
- [x] **FB-20**: Single-table official source plugin business table names should not repeat the resource suffix.
- [x] **FB-21**: Synchronize the main specifications to the final table name and capability boundary description after plugin.
- [x] **FB-22**: Merge the mistakenly created active iteration `host-plugin-boundary-followup` back to the current iteration.
- [x] **FB-23**: Change the variadic constructor used for "optional dependencies" in the host business component to an explicit parameter signature.
- [x] **FB-24**: Change the permission menu filtering dependency of `role` on `plugin` to narrow interface injection.
- [x] **FB-25**: Host controller constructor does not synchronize explicit dependency injection signatures, causing `go test ./...` to fail.
- [x] **FB-26**: `TC0069` still uses the old organization association table for cleaning, causing E2E failures.
- [x] **FB-27**: Operation log needs to be changed to a source plugin that self-registers the audit link through the global HTTP middleware registrar.
- [x] **FB-28**: `pluginbridge`/`pluginhost` behavioral objects exposed to plugins need to be unified interfaces.
- [x] **FB-29**: Delete the `plugin.jobs` plugin general task processor capability; retain only the `plugin.cron` built-in scheduled task projection link.
- [x] **FB-30**: Delete the `http.request.after-auth` / `RegisterAfterAuthHandler` plugin extension points.
- [x] **FB-31**: `monitor-operlog` audit middleware needs to converge to the plugin `service` layer.
- [x] **FB-32**: Source plugin `service` directory hierarchy is not standardized; uniformly migrate to `backend/internal/service/`.
- [x] **FB-33**: Supplement the source plugin directory structure to the project specification.
- [x] **FB-34**: Log TraceID should be turned off by default, and support both `config.yaml` static switch and system parameter dynamic override.
- [x] **FB-35**: System menu types still use `"D"` / `"M"` / `"B"` strings scatteredly; need to converge into strongly typed constants.
- [x] **FB-36**: `monitor-operlog` still uses `1~6` integers to express operation types; need to change to semantic strings.
- [x] **FB-37**: The host workbench still has unreferenced old versions of `monitor-operlog` frontend pages and API copies.
- [x] **FB-38**: `pluginbridge` exposes unified guest controller scaffolding with generics and error classifiers.
- [x] **FB-39**: `plugin-demo-dynamic` controller template slimming down using `pluginbridge` new helpers.
- [x] **FB-40**: The dynamic plugin guest controller supports the typed req/res signature of the API DTO driver.
- [x] **FB-41**: Remove the `apimeta.Meta` compatibility layer of the dynamic plugin API DTO after upstream GoFrame fixes.
- [x] **FB-42**: Remove the `sys.logger.traceID.enabled` parameter configuration; the log TraceID switch is only controlled by `config.yaml`.
- [x] **FB-43**: The unified task scheduling interface document group name is changed to `Task Scheduling/Timed Tasks`, etc.
- [x] **FB-44**: Remove the extra `demo.control.enabled` switch and use whether `plugin.autoEnable` contains `demo-control` as the demo-mode switch.
- [x] **FB-45**: Expand the demo-control global middleware scope to `/*`, cover the full system request chain, and preserve the login whitelist.
- [x] **FB-46**: Fix the issue where removing `demo-control` from `plugin.autoEnable` still left write interception active at runtime.
- [x] **FB-47**: Allow install, uninstall, enable, and disable operations for plugins other than `demo-control` while demo mode is active.
- [x] **FB-48**: Add bugfix feedback test coverage requirements to the project specification and `lina-review` skill.
