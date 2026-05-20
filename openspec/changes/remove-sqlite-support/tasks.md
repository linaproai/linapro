## 1. 后端数据库支持收敛

- [x] 1.1 从 `pkg/dbdriver` 移除 SQLite 驱动注册和支持类型，只保留 PostgreSQL 驱动。
- [x] 1.2 从 `pkg/dialect` 移除 SQLite 方言实现、转译器、错误分类和相关测试，确保 `sqlite:` 链接返回明确不支持错误。
- [x] 1.3 调整依赖文件并运行 `go mod tidy`，移除 `apps/lina-core` 中 SQLite 驱动链路依赖残留。
- [x] 1.4 更新后端测试，删除 SQLite 专属用例，保留或补充 PostgreSQL-only 的方言、驱动、插件数据服务、系统信息和缓存单测覆盖。

## 2. 工具、CI 与 E2E 清理

- [x] 2.1 删除 SQLite smoke workflow、CI 输入和 main/nightly/release 调用参数。
- [x] 2.2 删除 `hack/tests` 中 SQLite 专属 E2E 用例、support、脚本和 package scripts，并更新执行清单。
- [x] 2.3 更新 `linactl`、makefile 注释和测试中 SQLite 配置示例，确保默认开发工具链只表达 PostgreSQL。

## 3. 配置与文档同步

- [x] 3.1 更新运行时配置模板、packed 配置模板和镜像运行配置，移除 SQLite 示例并改用 PostgreSQL-only 注释。
- [x] 3.2 同步更新英文/中文 README 和测试文档中的数据库支持说明。
- [x] 3.3 更新现行 OpenSpec 基线中 SQLite 支持口径，确保非归档规范与本变更一致。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate remove-sqlite-support --strict`。
- [x] 4.2 运行至少覆盖变更包的 Go 编译/测试门禁，包括 `cd apps/lina-core && go test ./pkg/dbdriver ./pkg/dialect ./pkg/plugindb/host -count=1` 和受影响宿主包测试。
- [x] 4.3 运行工具链验证和静态扫描，确认非归档代码、CI、脚本、配置和文档不再包含受支持 SQLite 入口。
- [x] 4.4 记录 i18n 与缓存影响判断：本次不修改运行时翻译资源，不新增缓存，仅移除 SQLite 特殊缓存语义。
- [x] 4.5 执行 `/lina-review` 审查，修复发现的问题后再完成任务。

## 5. 验证记录

- [x] OpenSpec：`openspec validate remove-sqlite-support --strict`
- [x] 核心 Go：`cd apps/lina-core && go test ./pkg/dbdriver ./pkg/dialect ./pkg/plugindb/host ./internal/service/config ./internal/service/cachecoord ./internal/service/sysinfo ./internal/cmd ./internal/service/kvcache/internal/sqltable ./internal/service/plugin/internal/testutil -count=1`
- [x] 插件 Go：`cd apps/lina-plugins && GOWORK=off go test lina-plugin-linapro-monitor-server/backend/internal/service/monitor -count=1`
- [x] 工具链：`cd hack/tools/linactl && go test ./... -count=1`
- [x] E2E 清单：`cd hack/tests && pnpm test:validate`
- [x] 静态扫描：确认非归档代码中无 `include-sqlite`、`sqlite-smoke`、`run-sqlite`、`test:sqlite` 入口；确认 `go.mod` / `go.sum` / `go.work.sum` 无 SQLite 驱动链路依赖。
- [x] i18n 影响：本次没有新增、修改或删除运行时语言包、插件 `manifest/i18n` 或 apidoc i18n JSON；仅调整文档和测试说明中的数据库支持口径。
- [x] 缓存影响：本次不新增缓存；移除 SQLite 单机特殊缓存语义后，单机仍使用 PostgreSQL SQL table 后端，集群仍要求 coordination KV/Redis 语义。
- [x] 审查：`/lina-review` 已完成；未发现阻断项，审查范围限定在 `remove-sqlite-support` 相关代码、配置、CI、E2E 入口、文档与规范，未纳入并行存在的 E2E 重组变更。

## Feedback

- [x] **FB-1**: `linapro-monitor-server` PostgreSQL 单测未注册 `pgsql` 驱动导致 `g.DB()` 初始化 panic。

## Feedback 验证记录

- FB-1 修复：`linapro-monitor-server` monitor 单测包显式导入宿主统一 `pkg/dbdriver` 注册入口，并新增不依赖外部 PostgreSQL 服务的驱动注册 smoke，确保 `pgsql` 类型可在 GoFrame ORM 初始化阶段被发现。
- 验证通过：`cd apps/lina-plugins && GOWORK=off go test -run '^TestPostgreSQLDriverRegisteredForMonitorTests$' -count=1 -v lina-plugin-linapro-monitor-server/backend/internal/service/monitor`。
- 验证通过：`cd apps/lina-plugins && GOWORK=off go test -run 'TestUpsertMonitorSnapshotWorksOnPostgreSQL|TestGetDBInfoReturnsPostgreSQLVersion' -count=1 -v lina-plugin-linapro-monitor-server/backend/internal/service/monitor`；本机未设置 `LINA_TEST_PGSQL_LINK`，真实 PostgreSQL 集成断言按既有逻辑跳过。
- 验证通过：`cd apps/lina-plugins && GOWORK=off go test -count=1 lina-plugin-linapro-monitor-server/backend/internal/service/monitor`。
- 验证通过：`cd apps/lina-plugins && GOWORK=off go test -p=1 -race -count=1 -v lina-plugin-linapro-monitor-server/backend/internal/service/monitor`。
- 验证通过：`openspec validate remove-sqlite-support --strict`。
- 验证通过：`git diff --check -- openspec/changes/remove-sqlite-support/tasks.md` 与 `cd apps/lina-plugins && git diff --check -- linapro-monitor-server/backend/internal/service/monitor/monitor_upsert_test.go`。
- i18n 影响：本次仅修改后端测试和 OpenSpec 反馈记录，不新增或修改用户可见文案、运行时语言包、插件 `manifest/i18n` 或 apidoc i18n JSON。
- 缓存影响：本次不新增或修改运行时缓存、缓存键、失效触发点或跨实例同步机制。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口、数据库查询路径或角色数据权限边界。
- 审查：`/lina-review` 已完成；审查范围为 `linapro-monitor-server` monitor 单测驱动注册修复与 `remove-sqlite-support` 反馈记录，未发现阻断问题。
