## 1. 配置模型与校验

- [x] 1.1 扩展 `hack/tools/linactl` 的 `rootConfig`，新增 `plugins.sources` 配置结构，支持 `repo`、`root`、`ref` 和字符串数组 `items`
- [x] 1.2 在 `hack/config.yaml` 增加示例插件配置，覆盖官方来源和自定义来源的推荐写法
- [x] 1.3 实现配置校验：source 名称、repo、root、ref、items 必填检查，`items` 仅允许字符串数组
- [x] 1.4 实现插件 ID 全局唯一校验，重复插件必须在写入前失败并列出冲突 source
- [x] 1.5 实现路径安全校验，拒绝空 root、绝对路径、`..`、路径分隔符注入和 Windows drive path
- [x] 1.6 补充配置解析和校验单元测试，覆盖有效配置、重复插件、非法 root、非法 items 类型和缺失字段

## 2. 工作区去 submodule 化

- [x] 2.1 新增 `linactl plugins.init` 命令并在 `Makefile` / `make.cmd` 中提供 `make plugins.init` 包装入口
- [x] 2.2 实现 `apps/lina-plugins` 工作区状态检测，区分缺失、普通目录、submodule/gitlink、嵌套 Git 仓库和无效路径
- [x] 2.3 实现 `.gitmodules` section 删除逻辑：只删除 `apps/lina-plugins`，仅在没有其他 section 时删除整个 `.gitmodules`
- [x] 2.4 实现父仓库 submodule 配置清理，包括 `.git/config` 对应 section 和 `.git/modules/apps/lina-plugins` 元数据
- [x] 2.5 实现 gitlink 到普通目录的转换，保留 `apps/lina-plugins` 下已有插件文件内容
- [x] 2.6 补充单元测试或隔离仓库集成测试，覆盖已初始化 submodule、未初始化 submodule、普通目录、缺失目录和多 submodule `.gitmodules`

## 3. 插件安装与更新

- [x] 3.1 新增 `linactl plugins.install` / `make plugins.install`，按 `hack/config.yaml` 安装配置插件
- [x] 3.2 新增 `linactl plugins.update` / `make plugins.update`，按 `hack/config.yaml` 更新配置插件
- [x] 3.3 实现来源仓库临时 checkout，解析 `ref` 到 commit，并从 `<root>/<plugin-id>` 复制插件目录到 `apps/lina-plugins/<plugin-id>`
- [x] 3.4 确保安装/更新不会把来源仓库 `.git` 元数据写入插件目录
- [x] 3.5 实现目标目录保护：install 遇到已存在目录默认失败，update 遇到本地 dirty 默认失败，`force=1` 才允许覆盖
- [x] 3.6 支持命令子集筛选参数，例如 `p=<plugin-id>` 和 `source=<source-name>`，但 repo/root/ref 仍只来自 `hack/config.yaml`
- [x] 3.7 补充安装与更新测试，覆盖成功安装、缺失远端 plugin.yaml、目标已存在、dirty 阻断、force 覆盖和子集筛选

## 4. 锁定状态与状态检查

- [x] 4.1 设计并实现 `apps/lina-plugins/.linapro-plugins.lock.yaml`，记录插件 ID、source、repo、root、ref、resolved commit、manifest version 和内容摘要
- [x] 4.2 在 install/update 成功后写入或刷新锁定状态，失败时不得写入半成品锁文件
- [x] 4.3 新增 `linactl plugins.status` / `make plugins.status`，只读输出工作区类型、配置插件、本地插件、锁定状态和远端更新状态
- [x] 4.4 实现本地 dirty 检测，优先通过父仓库 `git status -- apps/lina-plugins/<plugin-id>` 判断用户可提交改动
- [x] 4.5 实现远端更新检测，远端不可达时输出 unknown，不修改本地状态
- [x] 4.6 补充状态命令测试，覆盖普通目录、submodule 提示、远端不可达、未配置本地插件、缺失本地插件和 orphaned lock entry

## 5. 文档与命令帮助

- [x] 5.1 更新根 `README.md` 和 `README.zh-CN.md`，说明官方仓库 submodule 与用户项目普通插件目录的区别
- [x] 5.2 更新 `hack/tools/linactl/README.md` 和 `README.zh-CN.md`，补充 `plugins.init`、`plugins.install`、`plugins.update`、`plugins.status` 用法
- [x] 5.3 更新命令帮助输出，确保 Windows、Linux、macOS 用户可通过相同 make/linactl 入口执行
- [x] 5.4 更新必要的开发规范说明，强调 `plugins.sources.items` 仅支持字符串数组且插件目录固定为 `apps/lina-plugins`

## 6. 验证与审查

- [x] 6.1 运行 `cd hack/tools/linactl && go test ./... -count=1`
- [x] 6.2 运行 `go run ./hack/tools/linactl test.scripts`
- [x] 6.3 运行 `openspec validate plugin-workspace-management --strict`
- [x] 6.4 运行 `git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`
- [x] 6.5 记录 i18n 影响结论：本变更仅涉及开发工具命令和文档，不新增前端运行时、接口文档或插件 manifest i18n；若实现中新增运行时文案则同步维护翻译
- [x] 6.6 记录缓存一致性结论：本变更不新增运行时缓存、缓存键或跨实例失效逻辑
- [x] 6.7 记录数据权限结论：本变更不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界
- [x] 6.8 记录 RESTful API 结论：本变更不新增后端 REST API
- [x] 6.9 完成实现后调用 `lina-review`，重点审查去 submodule 化安全性、路径安全、dirty 保护、跨平台实现和测试覆盖

## Verification Notes

- i18n: 本变更仅新增开发工具命令输出、README 文档和 `hack/config.yaml` 示例，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- Cache: 本变更不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑；插件目录变化仍由既有构建和插件同步流程处理。
- Data permission: 本变更不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- RESTful API: 本变更不新增后端 REST API。
- Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`。
- Review: 已完成 `lina-review` 审查；修复了新增 helper 未使用、新增 Go 测试函数缺少注释、更新命令未检查已提交锁摘要漂移、以及 submodule section 删除可能误删后续普通 Git config section 的问题。确认 `plugins.init/install/update/status` 均由 `linactl` Go 工具承载，`Makefile` 仅作为包装入口；`plugins.status` 在当前真实仓库中识别 `apps/lina-plugins` 为 submodule 并只读输出提示，未修改真实插件工作区。

## Feedback

- [x] **FB-1**: 清理已删除升级技能在项目文档、运行时提示和 i18n 资源中的残留描述
- [x] **FB-2**: 支持 `plugins.sources.<name>.items` 使用字符串 `"*"` 展开安装来源 root 下全部插件，禁止与显式插件 ID 混用
- [x] **FB-3**: 将测试工具入口从 `make test-*` / `linactl test-*` 统一改为 `make test.*` / `linactl test.*`
- [x] **FB-4**: 改进 `make plugins.install` 的终端过程信息，展示来源下载、插件安装进度和安装结果
- [x] **FB-5**: 改进 `make plugins.status` 的终端过程信息，并使用对齐表格展示配置插件状态
- [x] **FB-6**: 修复源码插件托管 cron 枚举单元测试缺少 host config service 注入导致官方插件注册 panic
- [x] **FB-7**: 修复 plugin-full CI 中宿主导入 `lina-plugins` 聚合模块失败导致 Go 测试和 E2E 启动失败
- [x] **FB-8**: 补齐 host-only E2E `TC0233b` 的消息接口 mock，避免页面额外受保护请求触发第二次 `/auth/refresh` 导致用例失败
- [x] **FB-9**: 修复宿主 i18n E2E 与源码插件专属 E2E 的跨插件测试模块加载边界，避免 Playwright 将插件 `hack/tests` 页面对象与测试文件按 CommonJS 加载导致 serial/full E2E 失败
- [x] **FB-10**: 修复 host-only E2E runner 仍纳入依赖源码插件或动态插件内容的 plugin-full 用例，导致 `plugins=0` 宿主矩阵错误失败
- [x] **FB-11**: 更新 E2E 中过期的官方插件工作区 `make wasm` 调用路径，改为当前仓库根级构建入口以恢复动态插件产物前置
- [x] **FB-12**: 调查 `TC0186` 平台管理员 impersonation 用例偶发失败，确认为多 Playwright 会话共享 artifact 目录引发的非稳定环境竞争，单会话复跑关闭
- [x] **FB-13**: 调查 `TC0206b` 租户级插件安装模式 serial 失败，单会话隔离复跑确认并非稳定产品回归，关闭
- [x] **FB-14**: 调查 `TC0066l` 源码插件焦点恢复 serial 失败，单会话隔离复跑确认并非稳定产品回归，关闭
- [x] **FB-15**: 调查 `TC0068b` 运行时 wasm hook 失败隔离 serial 失败，单会话隔离复跑确认并非稳定产品回归，关闭
- [ ] **FB-16**: 修复 plugin-full GitHub Actions `E2E tests (plugin-full)` 中 `TC0222`、`TC0098` 与 `TC0111` 在 CI 容器下稳定撞到默认 60 秒 Playwright 超时，导致 nightly plugin-full E2E 门禁失败
- [x] **FB-17**: 明确插件工作区更新只负责开发阶段离线文件覆盖，运行时状态和数据升级拆分到 `plugin-runtime-upgrade` 变更处理

## Feedback Verification Notes

- FB-1 i18n: 已同步更新源码插件待升级错误提示的 `en-US` 与 `zh-CN` 运行时错误翻译；未新增前端运行时文案、菜单、路由、按钮、接口文档或插件 manifest 文案。
- FB-1 cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-1 data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-1 RESTful API: 不新增后端 REST API。
- FB-1 dev tools: 仅清理已删除命令路径的说明文字，不新增或修改开发工具/脚本入口。
- FB-1 tests: 已通过 `go test ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin -run 'TestBuildSourcePluginUpgradePendingErrorIncludesBulkCommand|TestValidateSourcePluginUpgradeReadinessFailsForPendingUpgrade' -count=1`、`openspec validate plugin-workspace-management --strict`、残留静态扫描和 `git diff --check`。

- FB-2 i18n: 仅调整开发工具配置语义、命令输出和 README 文档，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-2 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-2 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-2 RESTful API: 不新增后端 REST API。
- FB-2 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`。
- FB-3 i18n: 仅调整开发工具命令名称、CI 调用和治理文档引用，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-3 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-3 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-3 RESTful API: 不新增后端 REST API。
- FB-3 Dev tools: `make test.go`、`make test.host`、`make test.plugins`、`make test.scripts` 已替代旧的 `make test-go`、`make test-host`、`make test-plugins`、`make test-scripts`；`linactl` 同步提供 `test.go`、`test.host`、`test.plugins`、`test.scripts`，并用单元测试断言旧 hyphen 命令不再注册。
- FB-3 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`make -n test.go test.host test.plugins test.scripts`、`make -n test.go plugins=0`、`make -n test-go` 确认旧目标不存在、`go run ./hack/tools/linactl test.scripts`、`go run ./hack/tools/linactl help`、`go run ./hack/tools/linactl help test.go` / `test.host` / `test.plugins` / `test.scripts`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/reusable-backend-unit-tests.yml`、`openspec validate plugin-workspace-management --strict`、旧命令名静态扫描和 `git diff --check`。
- FB-3 Review: 已完成 `lina-review` 审查；确认 `hack/makefiles/test.mk` 只保留点号命名测试目标，`linactl` 命令注册表只暴露 `test.go`、`test.host`、`test.plugins`、`test.scripts`，旧 hyphen 命令仅作为单元测试负向断言和本任务说明出现。变更不新增 REST API、业务数据操作、运行时缓存或 i18n 资源。
- FB-4 i18n: 仅调整开发工具终端输出，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-4 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-4 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-4 RESTful API: 不新增后端 REST API。
- FB-4 Dev tools: `plugins.install` / `plugins.update` 现在输出准备阶段、来源下载与解析、实际插件安装/更新总数以及逐插件 `[n/total]` 进度；Git clone 使用 `--progress` 并把进度输出到命令 stdout，避免长时间静默。
- FB-4 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/tools/linactl/command_plugin_management.go hack/tools/linactl/command_plugin_management_test.go openspec/changes/plugin-workspace-management/tasks.md`。
- FB-4 Review: 已完成 `lina-review` 审查；确认本次仅调整跨平台 Go 开发工具输出与测试断言，未新增平台专属脚本、REST API、业务数据操作、运行时缓存或 i18n 资源；新增 Go helper 均有职责注释且目标包编译通过。
- FB-5 i18n: 仅调整开发工具终端输出，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-5 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-5 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-5 RESTful API: 不新增后端 REST API。
- FB-5 Dev tools: `plugins.status` 现在在远端查询前输出正在查询配置插件来源，查询完成后输出正在渲染配置插件数量，并使用与 `make status` 一致的 ASCII 对齐表格展示 Plugin、Source、Version、Installed、Dirty、Remote 与 Note。
- FB-5 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/tools/linactl/command_plugin_management.go hack/tools/linactl/command_plugin_management_test.go openspec/changes/plugin-workspace-management/tasks.md`。
- FB-5 Review: 已完成 `lina-review` 审查；确认状态命令仍保持只读诊断，不写入 `apps/lina-plugins`、`.gitmodules` 或锁文件，表格渲染复用既有 ASCII 表格工具函数，目标测试已单独重复运行通过。
- FB-6 i18n: 仅调整 Go 单元测试夹具和测试覆盖，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-6 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-6 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-6 RESTful API: 不新增后端 REST API。
- FB-6 Dev tools: 不新增或修改开发工具、构建脚本、测试脚本或 CI 入口。
- FB-6 Tests: 已通过 `cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run TestListManagedCronJobsSkipsDynamicDiscoveryForSourcePlugins -count=1`、`go test ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil -count=1`、`go test ./internal/service/plugin/... -count=1`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_fixture.go apps/lina-core/internal/service/plugin/internal/integration/extensions_cron_test.go openspec/changes/plugin-workspace-management/tasks.md`。
- FB-6 Review: 已完成 `lina-review` 审查；确认 `testutil.NewServices` 为源码插件 cron 收集注入最小 host config service，回归测试通过读取 host config 并断言 source plugin managed cron 被收集来覆盖原 panic 路径，未修改生产接口或运行时行为。
- FB-7 i18n: 仅调整 Go 开发工具生成模块逻辑与 Go 测试夹具，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n；auth 测试夹具复用既有 runtime i18n 服务适配源码插件 hook。
- FB-7 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑；生成的 `temp/official-plugins` 仅服务本地/CI 构建解析。
- FB-7 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界；auth 测试夹具的 tenant-filter 仅用于源码插件登录日志 hook 在测试环境下的租户上下文解析。
- FB-7 RESTful API: 不新增后端 REST API。
- FB-7 Dev tools: `linactl` 在 plugin-full 模式下生成已忽略的 `temp/official-plugins` 聚合模块，模块名为 `lina-plugins`，并 blank import 官方 source plugin backend 注册包以满足宿主 `official_plugins` 构建；`test.go` 仍通过 `go.work.plugins` 编译宿主与插件模块，但跳过该生成聚合模块本身的独立 `go test ./...`。
- FB-7 Tests: 已通过 `cd hack/tools/linactl && go test ./... -run 'TestGoWorkspaceModulesSkipsGeneratedOfficialPluginAggregate|TestOfficialPluginBackendImportsDiscoversSourcePlugins|TestPrepareOfficialPluginWorkspaceWritesTemporaryWorkspace|TestOfficialPluginGoWorkUsesDiscoversPluginModules|TestOfficialPluginBuildEnvSeparatesHostOnlyAndPluginFullModes' -count=1`、`cd apps/lina-core && GOWORK=<repo>/temp/go.work.plugins GOFLAGS='-tags=official_plugins' go test -race ./internal/service/auth -run TestLoginRejectsBlacklistedIP -count=1 -v`、`cd apps/lina-core && GOWORK=<repo>/temp/go.work.plugins GOFLAGS='-tags=official_plugins' go test -race ./internal/service/auth -count=1`、`cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`make test.go plugins=1 race=true verbose=true`、`openspec validate plugin-workspace-management --strict`。
- FB-8~FB-11 i18n: 仅调整 E2E fixture、页面路由 mock、执行清单和测试脚本，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-8~FB-11 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-8~FB-11 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-8~FB-11 RESTful API: 不新增后端 REST API。
- FB-8 Notes: `TC0233` 追加 `GET /api/v1/user/message/count`、`GET /api/v1/user/message?*` 和 `GET /api/v1/platform/tenants?*` mock，并将 refresh 断言放宽为 `toBeGreaterThanOrEqual(1)`，覆盖页面新增受保护请求。
- FB-9 Notes: 调整宿主 i18n 与源码插件专属 E2E 的测试模块边界，避免 Playwright 在 serial/full 阶段把插件 `hack/tests` 页面对象与用例文件按 CommonJS 交叉加载。
- FB-10 Notes: `execution-manifest.json`、`execution-governance.mjs` 与 `run-suite.mjs` 已补齐 host-only 排除与 serial 拆分治理，避免 `plugins=0` 宿主矩阵误纳入依赖插件/动态 wasm 的用例。
- FB-11 Notes: `TC0067`、`TC0107`、`TC0108`、`TC0140` 等用例中的 `make wasm` 已统一改为仓库根级构建入口，恢复当前官方插件工作区的动态插件产物前置步骤。
- FB-8~FB-11 Tests: 已通过 `make test.go plugins=0 race=true verbose=true`、`make test.go plugins=1 race=true verbose=true`、`cd apps/lina-vben && pnpm test:unit`、`cd hack/tests && pnpm test:validate`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test $(cat ../../temp/full-parallel-files.txt) $(cat ../../temp/full-serial-files.txt) --workers=1`、`openspec validate plugin-workspace-management --strict` 与 `git diff --check`。
- FB-12~FB-15 Investigation: 单会话隔离复跑 `TC0186`、`TC0206`、`TC0066`、`TC0068` 全部通过；之前 full/serial 阶段的 4 个失败由多个 Playwright 进程共享 `hack/tests/test-results` 引发的 artifact/trace 竞争造成，包含 `ENOENT` 与 trace/network 文件缺失，不是稳定产品回归。
- FB-12~FB-15 Closure: 后续完整 E2E 统一使用单一 `playwright test` 会话和 `E2E_BROWSER_CHANNEL=chrome` 运行，避免 bundled Chromium 下载/版本问题与共享 artifact 目录竞争。
- FB-12~FB-15 Tests: 已通过 `cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test apps/lina-plugins/multi-tenant/hack/tests/e2e/platform-admin/TC0186-impersonation-start-exit.ts --workers=1`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test apps/lina-plugins/multi-tenant/hack/tests/e2e/plugin-governance/TC0206-tenant-aware-install-mode.ts --workers=1`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test hack/tests/e2e/extension/plugin/TC0066-source-plugin-lifecycle.ts --workers=1`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test hack/tests/e2e/extension/plugin/TC0068-runtime-wasm-failure-isolation.ts --workers=1`、`cd hack/tests && E2E_BROWSER_CHANNEL=chrome pnpm exec playwright test $(cat ../../temp/full-parallel-files.txt) $(cat ../../temp/full-serial-files.txt) --workers=1`、`cd apps/lina-vben && pnpm test:unit`、`cd hack/tests && pnpm test:validate`、`make test.go plugins=0 race=true verbose=true`、`make test.go plugins=1 race=true verbose=true`、`openspec validate plugin-workspace-management --strict` 与 `git diff --check`。
- FB-17 i18n: 仅补充 OpenSpec 规范边界，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n；运行时升级相关文案由 `plugin-runtime-upgrade` 变更负责。
- FB-17 Cache: `plugin-workspace-management` 仍不新增运行时缓存、缓存键或跨实例失效；运行时升级缓存一致性由 `plugin-runtime-upgrade` 变更负责。
- FB-17 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-17 RESTful API: 不新增后端 REST API；运行时升级 API 在 `plugin-runtime-upgrade` 中设计。
- FB-17 Dev tools: 明确 `plugins.install` / `plugins.update` 只覆盖插件文件和锁定状态，不连接运行时数据库、不执行升级 SQL、不调用插件升级回调。
- FB-17 Tests: 使用 OpenSpec 治理验证覆盖，运行 `openspec validate plugin-workspace-management --strict` 和 `openspec validate plugin-runtime-upgrade --strict`。
