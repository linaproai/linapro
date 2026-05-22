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
- [x] **FB-2**: `make init` 在 clean checkout 中因 `internal/packed/public` 没有被跟踪的嵌入文件而无法编译 `go:embed all:public`。
- [x] **FB-3**: `plugindb/host` 的元数据读与 schema probe 判定包含 PostgreSQL 专属 SQL 字符串，导致插件数据治理层与具体数据库实现强绑定。

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

- FB-2 修复：为 `apps/lina-core/internal/packed/public` 增加被 Git 跟踪的 `.gitkeep` 占位文件，确保 clean checkout 中 `//go:embed all:public all:manifest` 至少能匹配一个 public 文件；同时让 `linactl build` 刷新前端嵌入目录后自动重建该占位文件。
- 新增测试：`apps/lina-core/internal/packed` 增加占位文件嵌入断言；`hack/tools/linactl` 增加 `linactl build` 占位文件重建断言。
- 验证通过：`cd apps/lina-core && go test ./internal/packed -run '^TestFilesEmbedFrontendPlaceholder$' -count=1`。
- 验证通过：`cd hack/tools/linactl && go test ./... -run '^TestEnsurePackedPublicPlaceholderCreatesGitkeep$' -count=1`。
- 验证通过：仅保留 `internal/packed/public/.gitkeep` 的 clean-checkout 模拟下执行 `cd apps/lina-core && go run main.go init --help`，确认原始 `go:embed all:public` 编译错误消失。
- 验证通过：`cd apps/lina-core && go test ./internal/cmd ./internal/packed -count=1`。
- 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- i18n 影响：本次不新增或修改用户可见文案、运行时语言包、插件 `manifest/i18n` 或 apidoc i18n JSON。
- 缓存影响：本次不新增或修改运行时缓存、缓存键、失效触发点或跨实例同步机制。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口、数据库查询路径或角色数据权限边界。
- 审查：`/lina-review` 已完成；审查范围为 `internal/packed` 嵌入占位、`linactl build` 占位重建逻辑、相关测试与 `remove-sqlite-support` 反馈记录，未发现阻断问题。

- FB-3 修复：将插件数据治理层的读 SQL 分类改为依赖 `pkg/dialect` 暴露的 `ClassifyReadSQL` 抽象；`plugindb/host` 只判断“元数据读/无表 schema probe”语义，不再包含 PostgreSQL catalog/schema probe 具体字符串，PostgreSQL 专属识别逻辑收敛到 `pkg/dialect/internal/postgres`；同步更新 `database-dialect-abstraction` 增量规范，明确驱动/ORM 只读 SQL 分类属于方言边界。
- 新增测试：`pkg/dialect/internal/postgres` 覆盖 PostgreSQL catalog lookup、schema probe 与普通应用读分类；`pkg/dialect` 覆盖通过 GoFrame driver type 解析方言；`pkg/plugindb/host` 覆盖 table guard 继续允许 ORM 元数据读和只读 schema probe，同时新增源码扫描断言防止 `plugindb/host/db.go` 再引入 PostgreSQL catalog 字符串。
- 验证通过：`cd apps/lina-core && go test ./pkg/dialect/internal/postgres ./pkg/dialect ./pkg/plugindb/host -count=1`。
- 验证通过：`cd apps/lina-core && go test ./pkg/dialect/... ./pkg/plugindb/... -count=1`。
- 验证通过：`openspec validate remove-sqlite-support --strict`。
- 静态扫描通过：`rg -n "information_schema|pg_catalog|current_schema|pg_class|pg_namespace|version\\(\\)" apps/lina-core/pkg/plugindb/host/db.go` 无匹配。
- i18n 影响：本次仅调整后端治理抽象和单元测试，不新增、修改或删除用户可见文案、运行时语言包、插件 `manifest/i18n` 或 apidoc i18n JSON。
- 缓存影响：本次不新增或修改运行时缓存、缓存键、失效触发点、跨实例同步机制或缓存一致性模型。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口；插件数据服务原有表级治理继续在 `DoCommit` 阶段校验授权资源表，本次只移动数据库方言分类边界。
- 开发工具影响：本次不新增或修改开发工具、脚本、CI 默认入口或平台相关命令。
- 审查：`/lina-review` 已完成；审查范围为 `pkg/dialect` 只读 SQL 分类抽象、`pkg/plugindb/host` 表级治理调用点、相关单元测试与 `remove-sqlite-support` 反馈记录，未发现阻断问题。
