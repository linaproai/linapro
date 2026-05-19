## 1. 命令实现

- [x] 1.1 新增 `linactl env.check` 命令，检测 Go、Node.js、pnpm、Vite、Playwright、PostgreSQL 并输出包含名称、当前版本、要求版本、是否满足、备注的表格。
- [x] 1.2 将原 `linactl dev.setup` 实现迁移为 `linactl env.setup`，保持前端依赖与 Playwright Chromium 安装行为一致。
- [x] 1.3 从 `linactl` 注册表和根 Make 目标中移除 `dev.setup`，并新增 `env.check` / `env.setup` Make 目标。
- [x] 1.4 更新 Playwright 缺失浏览器提示和相关文档/帮助输出，统一指向 `make env.setup`。

## 2. 治理与验证

- [x] 2.1 新增或更新 `hack/tools/linactl` 单元测试，覆盖环境命令注册、旧命令移除、环境检查表格和命令文件命名治理。
- [x] 2.2 运行 `openspec validate dev-environment-commands --strict` 验证 OpenSpec 文档。
- [x] 2.3 运行 `cd hack/tools/linactl && go test . -count=1` 验证 Go 工具编译与单元测试。
- [x] 2.4 运行 `cd hack/tools/linactl && go run . test.scripts` 验证跨平台工具治理扫描。
- [x] 2.5 记录 i18n、缓存、数据权限、运行时 API 和前端页面影响评估：本次仅修改开发工具入口，不影响这些运行时资源。

执行记录：
- 已新增 `command_env.check.go`，输出带边框的本地开发工具检查表；缺失工具以 `No` 和修复备注展示，不中断其它工具检查。
- 已将原 `command_dev.setup.go` 迁移为 `command_env.setup.go`，保留 `frontend.EnsureDeps` 与 `pnpm exec playwright install --with-deps chromium` 安装流程。
- 已新增 `hack/makefiles/env.mk` 并纳入根 `Makefile`，移除 `hack/makefiles/dev.mk` 中的 `dev.setup` 目标和 `linactl` 注册项。
- 已更新 Playwright 缺失浏览器提示、`AGENTS.md` 与 `lina-review` 命令文件命名示例为 `env.setup`，并同步仍活跃的 `backend-source-readability-governance` 规范示例，避免治理规则继续要求已移除的旧命令。
- 本次仅修改开发工具入口、仓库治理说明和 OpenSpec 文档；不涉及运行时 API、数据库 Schema、权限模型、数据权限、缓存策略、前端页面或 i18n JSON 资源。
- 验证通过：`make env.check`、`openspec validate dev-environment-commands --strict`、`cd hack/tools/linactl && go test . -count=1`、`cd hack/tools/linactl && go run . test.scripts`。

## Feedback

- [x] **FB-1**: `make env.check` 使用带框表格展示，并通过 `apps/lina-core/manifest/config/config.yaml` 的数据库连接检测 PostgreSQL 服务端版本；无法检测时在 Remark 中说明原因。
- [x] **FB-2**: PostgreSQL 版本检测必须使用 Go 数据库连接直接查询，不依赖 `psql` 客户端工具。
- [x] **FB-3**: `TestRunEnvSetupInstallsFrontendAndPlaywright` 在缺少真实 `pnpm` 的 CI 环境中不应因测试替身被 `PATH` 预检提前绕过而失败。

反馈执行记录：
- 已将 `env.check` 输出改为稳定的 ASCII 带框表格，表头包含 `Name`、`Current Version`、`Required Version`、`Satisfied`、`Remark`。
- PostgreSQL 检测改为读取 `apps/lina-core/manifest/config/config.yaml` 中的 `database.default.link`，通过 Go `database/sql` 连接配置的数据库执行 `SHOW server_version`，展示服务端版本而非本地客户端版本。
- 如果配置缺失、不是 PostgreSQL 或 Go 数据库连接/查询失败，PostgreSQL 行会在 `Remark` 中说明无法检测服务端版本的具体原因。
- 本次仅修改跨平台 Go 开发工具与 OpenSpec 任务记录；不涉及运行时 API、数据库 Schema、权限模型、数据权限、缓存策略、前端页面或 i18n JSON 资源。
- 验证通过：`cd hack/tools/linactl && go test . -count=1`、`cd hack/tools/linactl && go run . test.scripts`、`openspec validate dev-environment-commands --strict`、`make env.check`。
- 已按 FB-2 移除 PostgreSQL 检测对 `psql` 客户端工具的依赖，改用 `database/sql` 与 Go PostgreSQL driver 直接连接并执行 `SHOW server_version`；测试覆盖 Go driver 查询成功、查询失败 Remark 和配置类型错误。
- 已按 FB-3 将 `linactl` 子命令的工具查找依赖收敛为 `app.lookPath`，生产默认仍使用 `exec.LookPath`，测试可与 `execCommand` 替身同步注入，避免 CI 缺少真实 `pnpm` 时测试在执行替身前失败；新增 `TestRunCommandReportsMissingToolBeforeExecution` 锁定缺工具时仍在启动子进程前返回可读 PATH 诊断。验证通过：`cd hack/tools/linactl && go test . -run 'TestRunEnvSetupInstallsFrontendAndPlaywright|TestRunCommandReportsMissingToolBeforeExecution' -count=1`、`cd hack/tools/linactl && go test . -count=1`、`cd hack/tools/linactl && go run . test.scripts`、`openspec validate dev-environment-commands --strict`、`git diff --check -- hack/tools/linactl/constants_types.go hack/tools/linactl/app.go hack/tools/linactl/main_test.go openspec/changes/dev-environment-commands/tasks.md`。本次仅修改跨平台 Go 开发工具测试隔离与 OpenSpec 任务记录；不涉及运行时 API、数据库 Schema、权限模型、数据权限、缓存策略、前端页面或 i18n JSON 资源。
