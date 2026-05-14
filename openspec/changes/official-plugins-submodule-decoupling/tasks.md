## 1. 基线调查与隔离验证

- [x] 1.1 在隔离工作区或可恢复的临时路径中记录当前 `git status --short`、`go.work`、插件目录状态和官方插件列表
- [x] 1.2 临时移出 `apps/lina-plugins`，运行宿主后端构建和宿主后端单元测试，记录所有失败点并按 Go workspace、编译期导入、运行时扫描、测试辅助分类
- [x] 1.3 创建空目录 `apps/lina-plugins`，重复宿主后端构建和宿主后端单元测试，记录与目录缺失状态的差异
- [x] 1.4 在目录缺失和空目录两种状态下运行宿主前端类型检查或构建，记录 Vite 插件页面扫描、路由注册和访问过滤相关失败
- [x] 1.5 在目录缺失和空目录两种状态下运行宿主 E2E 发现/校验命令，记录 Playwright 配置、执行治理脚本和插件测试范围解析失败

## 2. Go Workspace 与宿主编译解耦

- [x] 2.1 调整默认 Go workspace 或生成流程，使 host-only 状态不因 `apps/lina-plugins` 和各 `lina-plugin-*` module 缺失而失败
- [x] 2.2 移除宿主默认入口对官方插件聚合 module 的无条件编译期依赖，并实现显式启用官方源码插件注册的完整构建路径
- [x] 2.3 为官方源码插件注册路径增加缺失 submodule 的可操作错误提示，包含 `git submodule update --init --recursive`
- [x] 2.4 补充或更新 Go 单元测试，覆盖默认宿主构建路径不依赖官方插件 module

## 3. 插件发现与工具入口解耦

- [x] 3.1 调整源码插件 manifest 扫描，在 `apps/lina-plugins` 不存在或为空时返回空源码插件集合，并保留动态插件发现
- [x] 3.2 区分普通读取和显式源码插件操作：插件列表/宿主启动可降级为空集合，插件专属同步、wasm 构建和完整插件验证必须快速失败
- [x] 3.3 更新 `linactl`、`make wasm`、mock SQL 加载和测试辅助中的插件根解析，统一输出插件工作区状态诊断
- [x] 3.4 补充后端单元测试，覆盖源码插件工作区不存在、为空目录、结构无效和 submodule 正常存在四类状态

## 4. 前端 Host-only 支持

- [x] 4.1 调整 Vite 源码插件页面扫描，使插件工作区不存在或为空时返回空模块集合
- [x] 4.2 调整插件页面注册和访问过滤逻辑，使没有源码插件页面时宿主路由、菜单和权限过滤仍正常工作
- [x] 4.3 运行 `pnpm -F @lina/web-antd typecheck` 和必要的前端构建验证，覆盖 host-only 状态
- [x] 4.4 若新增或修改用户可见空状态/错误文案，同步维护前端运行时 i18n 和插件 manifest i18n 资源；若无 i18n 影响，在任务记录中明确说明

## 5. E2E 与验证入口

- [x] 5.1 调整 Playwright 配置和测试治理脚本，使宿主 E2E 范围在 `apps/lina-plugins` 不存在或为空时仍可发现并执行
- [x] 5.2 调整 `plugins` 与 `plugin:<plugin-id>` 范围，使缺少官方插件工作区时快速失败并提示初始化 submodule
- [x] 5.3 新增或更新宿主级 E2E 覆盖 host-only 插件管理空状态，按当前最大 TC ID 分配编号
- [x] 5.4 运行宿主 E2E 校验与目标宿主 E2E，用目录缺失和空目录两种状态验证通过

## 6. Submodule 迁移与完整插件验证

- [x] 6.1 将官方插件仓库作为单个 submodule 挂载到 `apps/lina-plugins`，补齐 `.gitmodules` 和初始化说明
- [x] 6.2 submodule 初始化后运行所有官方插件 Go 单元测试，并记录失败修复结果
- [x] 6.3 submodule 初始化后运行所有官方插件自有 E2E 或插件测试清单中的完整插件范围，并记录失败修复结果
- [x] 6.4 submodule 初始化后运行动态插件 wasm 构建或等价产物验证，确认完整构建路径仍可用

## 7. 文档、治理与审查

- [x] 7.1 更新 README/README.zh-CN、CONTRIBUTING 和 AGENTS 中的官方插件工作区、submodule 初始化、host-only 验证和 plugin-full 验证说明
- [x] 7.2 更新 OpenSpec 任务记录，明确 i18n 影响、缓存一致性影响、数据权限影响和测试覆盖结论
- [x] 7.3 运行 `openspec validate official-plugins-submodule-decoupling --strict`
- [x] 7.4 运行 `git diff --check -- openspec/changes/official-plugins-submodule-decoupling .gitmodules go.work Makefile hack apps/lina-core apps/lina-vben README.md README.zh-CN.md CONTRIBUTING.md AGENTS.md`
- [x] 7.5 调用 `lina-review` 完成代码和规范审查，修正审查发现后再标记本变更完成

## Feedback

- [x] **FB-1**: 多租户插件 E2E 专属场景 helper 和 fixture 仍维护在宿主测试目录，影响官方插件工作区解耦；已迁移到插件 `hack/tests/support/`，本治理迁移不影响 i18n、缓存和数据权限
- [x] **FB-2**: 运行时生产源码仍判断官方源码插件工作区路径状态，导致宿主业务运行时耦合开发态 submodule 布局；已移除 catalog 生产路径中的 workspace 分类/路径推导，改为只扫描已注册源码插件和动态运行时产物，保留工具、构建和测试入口的 preflight；同步更新接口文档 i18n，未新增缓存或数据权限影响
- [x] **FB-3**: 官方插件工作区测试辅助组件维护在 `internal/testsupport`，作为跨包共享测试支持边界不清；已迁移到 `pkg/testsupport` 并保持仅测试代码引用
- [x] **FB-4**: `pluginhost` 源码插件注册错误仍使用 `fmt.Errorf` 和字符串 panic，未统一使用 GoFrame `gerror` 保留结构化栈信息；已改为 `gerror.New/Newf`，并补充测试覆盖测试辅助注册错误的 GoFrame 栈信息
- [x] **FB-5**: `make dev`、`make build` 等编译入口未显式传入 `plugins` 时默认 host-only，导致已初始化官方插件工作区仍缺失源码插件；已改为按插件目录自动选择编译模式，并在源码插件模式下基于宿主专用根 `go.work` 生成已忽略的 `temp/go.work.plugins`，通过 `GOWORK` 临时使用插件 workspace，根 `go.work` 始终保持 host-only
- [x] **FB-6**: `linactl` 直接运行示例和包装入口仍强制 `GOWORK=off`，与当前根 `go.work` 已包含工具模块、插件完整模式显式管理临时 workspace 的设计不一致；已移除 README 示例、`Makefile` 和 `make.cmd` 的强制 `GOWORK=off`，并更新 smoke 脚本验证直接使用根 `go.work`
- [x] **FB-7**: `hack/tools/linactl/main.go` 单文件内容过大，命令编排、插件 workspace、构建、服务控制和通用工具职责集中在一个源码文件中，影响维护和审查；已按职责拆分为同一 `package main` 下的命令、插件 workspace、构建、服务控制、文件工具和通用工具源码文件，未新增外部可引用包边界
- [x] **FB-8**: `hack/tools/build-wasm/builder` 作为独立可导入包暴露在工具模块根下，组件边界过宽；已迁移到 `hack/tools/build-wasm/internal/builder`，仅允许 `build-wasm` 工具模块内部引用
- [x] **FB-9**: nightly CI 缺少针对 `make build plugins=0` 编译产物的启动与核心功能 smoke 验证，无法证明 host-only 发布产物在不启用官方插件时可独立运行；已新增独立 nightly `host-only-build-smoke` job，显式执行 `make build plugins=0`，启动编译产物并验证健康检查、管理员登录和源码插件列表为空
- [x] **FB-10**: GitHub Actions 未在 host-only 与 plugin-full 路径中一致显式处理官方插件 submodule，导致 nightly full E2E 缺少 submodule 时失败、nightly image 静默退化为 host-only、backend unit 未覆盖插件 Go modules、Windows wasm smoke 默认依赖未初始化 submodule；已改为 reusable backend unit 通过 `linactl test-go plugins=<mode>` 生成对应 Go workspace，Windows wasm smoke 仅在插件模式运行，main/nightly 默认宿主路径显式 `plugins=0`，release 与 nightly full E2E/image 显式初始化 submodule 并强制 `plugins=1`
- [x] **FB-11**: `hack/scripts` 仍保留已可由 `linactl` 覆盖的 Bash 脚本，项目规范和 `lina-review` 也缺少开发工具脚本必须跨平台、优先使用 Go 工具链实现的治理要求；已删除 `prepare-packed-assets.sh` 与 `stop-dev-services.sh`，对应能力由 `linactl prepare-packed-assets` 和 `linactl stop` 承载，`linactl test-scripts` 改为 Go 测试与 Go 静态治理检查，并删除已失效的 Bash smoke 脚本
- [x] **FB-12**: nightly 每日测试与构建未将 host-only 和 plugin-full 测试矩阵单独运行，无法同时证明未初始化 submodule 与完整官方插件工作区两种质量边界；已新增 nightly host-only E2E、plugin-full Windows command smoke、plugin-full Go unit tests、plugin-full frontend unit tests，并将 plugin-full E2E 与 host-only E2E 使用独立 job 和独立报告产物，nightly image 发布依赖完整矩阵通过

## Verification Notes

- i18n: 本次变更仅新增/调整 host-only 与 plugin-full 的初始化与错误提示，已同步维护宿主运行时 i18n 错误资源；无新增插件业务翻译面。
- i18n(FB-2): `/plugins/sync` 接口说明不再引用 `apps/lina-plugins`，已同步 `manifest/i18n/zh-CN/apidoc/core-api-plugin.json` 与 packed apidoc 资源。
- i18n(FB-3): 仅迁移 Go 测试支持组件和测试导入路径，未新增或修改用户可见文案、接口文档或运行时翻译资源。
- i18n(FB-4): 仅调整内部开发诊断错误构造方式，错误文本不进入 HTTP/API 用户响应或运行时语言包，不新增 i18n 资源。
- i18n(FB-5): 仅调整开发/构建命令输出与文档说明，未新增前端运行时页面文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-6): 仅调整开发命令示例、包装入口和治理测试脚本，未新增或修改前端运行时页面文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-7): 仅拆分 `linactl` Go 源文件结构，不修改命令输出、用户可见文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-8): 仅迁移 `build-wasm` 内部 Go 包目录并更新导入路径，不修改命令输出、用户可见文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-9): 仅调整 GitHub Actions nightly 验证流程与 CI 步骤名称，不新增或修改前端运行时页面文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-10): 仅调整 GitHub Actions 工作流、开发工具测试入口和测试跳过条件，不新增或修改前端运行时页面文案、接口文档或插件 manifest 翻译资源。
- i18n(FB-11): 仅调整开发工具入口、项目规范和审查标准，不新增或修改前端运行时页面文案、接口文档、manifest i18n 或插件清单文案。
- i18n(FB-12): 仅调整 GitHub Actions nightly job 编排、job 名称和 artifact 名称，不新增或修改前端运行时页面文案、接口文档、manifest i18n 或插件清单文案。
- Cache: 未新增缓存策略或缓存扇区，沿用现有插件发现与运行时缓存机制，未引入新的跨实例失效要求。
- Cache(FB-2): 运行时源码插件发现改为读取进程内已注册插件定义，没有新增缓存、订阅或跨实例失效面。
- Cache(FB-3): 测试支持组件迁移不涉及运行时缓存、订阅、失效或跨实例一致性策略。
- Cache(FB-4): 源码插件注册表既有进程内状态和 listener 失效逻辑未变化，仅错误类型改为 `gerror`。
- Cache(FB-5): 仅调整编译入口、工具环境与临时 `temp/go.work.plugins` 生成逻辑，不新增运行时缓存、缓存失效或跨实例一致性影响。
- Cache(FB-6): 仅调整开发工具启动环境和文档示例，不新增运行时缓存、失效或跨实例一致性影响。
- Cache(FB-7): 仅拆分开发工具源码文件，不新增运行时缓存、失效或跨实例一致性影响。
- Cache(FB-8): 仅收敛开发工具包导入边界，不新增运行时缓存、失效或跨实例一致性影响。
- Cache(FB-9): 仅新增 CI 中的 host-only 编译产物启动验证，不修改运行时缓存、缓存失效或跨实例一致性策略。
- Cache(FB-10): 仅调整 CI 编排和测试入口，不修改运行时缓存、缓存失效或跨实例一致性策略。
- Cache(FB-11): 不新增或修改业务运行时缓存、缓存键、缓存失效路径、跨实例同步机制或故障降级策略。
- Cache(FB-12): 仅调整 CI 测试矩阵和发布门禁，不修改运行时缓存、缓存失效或跨实例一致性策略。
- Data permission: 未新增或扩大数据操作接口，沿用现有数据权限接入方式。
- Data permission(FB-2): 未新增数据查询/写入接口，插件同步仍沿用既有权限 `plugin:install` 和现有 registry 同步路径。
- Data permission(FB-3): 未新增或扩大数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-4): 未新增或修改数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-5): 未新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-6): 未新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-7): 未新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-8): 未新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-9): 未新增或修改 HTTP/API 数据操作接口；CI smoke 仅验证匿名健康检查和管理员登录，不改变角色数据权限边界。
- Data permission(FB-10): 未新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- Data permission(FB-11): 未新增或修改 HTTP/API 数据操作接口、服务数据访问路径或插件宿主数据访问路径，不涉及角色数据权限边界。
- Data permission(FB-12): 未新增或修改 HTTP/API 数据操作接口、服务数据访问路径或插件宿主数据访问路径，不涉及角色数据权限边界。
- Review: 已完成 `lina-review` 审查；修复了 `AGENTS.md` 目录树缩进与 submodule 语义说明、host-only 构建继承 `GOFLAGS=-tags=official_plugins` 时可能错误启用官方插件注册的问题，并补充 `linactl` 环境组装测试；同时显式处理 `build-wasm` 临时 `go.mod`/`go.sum` 清理错误。
- Review(FB-5): 已完成最新 `lina-review` 审查；确认根目录 `go.work` 保持 host-only，插件模式只生成已忽略的 `temp/go.work.plugins` 并通过 `GOWORK` 使用，未发现新的 i18n、缓存一致性或数据权限影响。
- Review(FB-6): 已完成最新 `lina-review` 审查；确认 `linactl` 直接运行示例、`Makefile` 和 `make.cmd` 不再强制 `GOWORK=off`，并将 `make.cmd` 调整为 `pushd/popd` 保留调用方工作目录；本次反馈不涉及后端运行时代码、API、缓存、i18n 资源或数据权限边界。
- Review(FB-7): 已完成最新 `lina-review` 审查；确认拆分仅改变 `linactl` 源码组织方式，`main.go` 保留包入口，其余职责保持同包不可被外部 import，未改变命令行为、运行时依赖、API、缓存、i18n 资源或数据权限边界。
- Review(FB-8): 已完成最新 `lina-review` 审查；确认 `builder` 只移动到 Go `internal` 包边界内，未改变命令行为、运行时依赖、API、缓存、i18n 资源或数据权限边界。
- Review(FB-9): 已完成最新 `lina-review` 审查；确认本次仅调整 GitHub Actions nightly 编排，未修改后端运行时代码、REST API、缓存、i18n 资源或数据权限边界；新增 job 显式使用 `plugins=0`，并在镜像发布前作为 `needs` 门禁。
- Review(FB-10): 已完成最新 `lina-review` 审查；确认 host-only 与 plugin-full CI 路径已显式区分，nightly full E2E 和 image 不再静默依赖插件自动探测，backend unit 的插件模式通过 `linactl` 生成临时 `temp/go.work.plugins`，未新增运行时 API、缓存、i18n 资源或数据权限影响。
- Review(FB-11): 已完成最新 `lina-review` 审查；确认 `hack/scripts` 下两个 Bash 脚本均为已由 `linactl` Go 命令覆盖的遗留入口，删除后 `go run ./hack/tools/linactl prepare-packed-assets`、`cd apps/lina-core && make prepare-packed-assets`、`make stop`、`make status` 和 `make test-scripts` 仍委托 `linactl`；项目规范和审查技能已新增跨平台开发工具脚本要求。现存 `hack/tests/scripts/run-sqlite-smoke.sh`、`hack/tests/scripts/run-redis-cluster-smoke.sh` 与前端 Docker 构建脚本属于本轮未改动的既有平台脚本，后续新增或修改时必须按新规范迁移到 Go/Node 或记录平台边界。
- Review(FB-12): 已完成最新 `lina-review` 审查；确认 nightly 已拆分 host-only 与 plugin-full 测试矩阵，host-only 路径不初始化 submodule 并运行 host-only Windows 命令 smoke、Go 单测、前端单测和 `pnpm test:host`，plugin-full 路径初始化 submodule 并运行 Windows 插件命令 smoke、Go 插件模式单测、前端单测和 full E2E，nightly image 发布前依赖两套测试矩阵全部通过。
- Tests: 已完成宿主后端构建、插件相关单测、前端 typecheck/build、前端 build、E2E 治理校验、宿主 E2E 校验与 plugin-full 构建验证。
- Tests(FB-2): 已通过 `go test ./apps/lina-core/internal/service/plugin/internal/catalog -count=1`、`go test ./apps/lina-core/internal/service/plugin/internal/testutil -count=1`、`go test ./apps/lina-core/pkg/pluginhost -count=1`、`go test ./apps/lina-core/internal/service/plugin -count=1`、`go test ./apps/lina-core/internal/service/plugin/internal/integration -count=1`、`go test ./apps/lina-core/internal/service/plugin/internal/runtime -count=1`、`go test ./apps/lina-core/internal/controller/plugin -count=1`、`go test ./hack/tools/build-wasm/... -count=1`、`go test ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/service/i18n -count=1`、`go build ./apps/lina-core`、`openspec validate official-plugins-submodule-decoupling --strict`。
- Tests(FB-3): 已通过 `go test ./apps/lina-core/pkg/testsupport -count=1`、`go test ./apps/lina-core/pkg/dialect -count=1`、`go test ./apps/lina-core/internal/cmd -count=1`、`go test ./apps/lina-core/internal/service/i18n ./apps/lina-core/internal/service/apidoc -count=1`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- apps/lina-core openspec/changes/official-plugins-submodule-decoupling/tasks.md`；静态扫描确认无 `internal/testsupport` 残留，`pkg/testsupport` 仅被 `_test.go` 文件导入。
- Tests(FB-4): 已通过 `go test ./apps/lina-core/pkg/pluginhost -count=1`、`go test ./apps/lina-core/pkg/pluginhost -run TestRegisterSourcePluginForTestReturnsGoFrameError -count=3`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- apps/lina-core/pkg/pluginhost/pluginhost_registry.go apps/lina-core/pkg/pluginhost/pluginhost_source_plugin_test.go openspec/changes/official-plugins-submodule-decoupling/tasks.md`。
- Tests(FB-5): 已通过 `cd hack/tools/linactl && GOWORK=off go test ./... -count=1`、`go test ./hack/tools/build-wasm/... -count=1`、`go test ./apps/lina-core/internal/service/plugin/internal/testutil -count=1`、`go test ./apps/lina-core/internal/service/plugin/internal/catalog ./apps/lina-core/pkg/pluginhost -count=1`、`bash hack/tests/scripts/test-linactl-cross-platform.sh`、`GOWORK="$PWD/temp/go.work.plugins" GOFLAGS=-tags=official_plugins go build -o /tmp/linapro-plugin-mode-go-build ./apps/lina-core`、`cd hack/tools/linactl && GOWORK=off go run . build --platforms=$(go env GOOS)/$(go env GOARCH) --output-dir=/tmp/linapro-autoplugins-build-test-4 --verbose=0`、`cd hack/tools/linactl && GOWORK=off go run . build plugins=0 --platforms=$(go env GOOS)/$(go env GOARCH) --output-dir=/tmp/linapro-hostonly-build-test-3 --verbose=0`、`test "$(rg -n "apps/lina-plugins" go.work || true)" = ""`、`test -f temp/go.work.plugins`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- Makefile make.cmd hack/makefiles hack/tools/linactl hack/tools/build-wasm apps/lina-core/internal/service/plugin/internal/testutil/testutil_build.go README.md README.zh-CN.md CONTRIBUTING.md AGENTS.md hack/tests/scripts/test-linactl-cross-platform.sh go.work openspec/changes/official-plugins-submodule-decoupling/tasks.md`。
- Tests(FB-6): 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`bash hack/tests/scripts/test-linactl-cross-platform.sh`、`go run ./hack/tools/image-builder --preflight --tag=test-preflight`、`make -n status`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- README.md README.zh-CN.md Makefile make.cmd hack/tools/linactl/README.md hack/tools/linactl/README.zh-CN.md hack/tests/scripts/test-linactl-cross-platform.sh openspec/changes/official-plugins-submodule-decoupling/tasks.md`。
- Tests(FB-7): 已通过 `cd hack/tools/linactl && goimports -w *.go && go test ./... -count=1`、`bash hack/tests/scripts/test-linactl-cross-platform.sh`、`make -n status`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- hack/tools/linactl openspec/changes/official-plugins-submodule-decoupling/tasks.md`；生产 Go 文件函数注释扫描无遗漏，`main.go` 已从 2183 行降至 21 行，拆分后最大生产文件为 398 行。`go run ./hack/tools/image-builder --preflight --tag=test-preflight` 未作为本次通过项记录，因为当前工作区已删除无关的 `hack/tools/image-builder/main.go`，该工具目录没有非测试入口文件。
- Tests(FB-8): 已通过 `go test ./hack/tools/build-wasm/... -count=1`、`cd hack/tools/build-wasm && go test ./... -count=1`；`rg -n '"build-wasm/builder"|go test \./builder|hack/tools/build-wasm/builder' hack/tools/build-wasm README.md README.zh-CN.md CONTRIBUTING.md AGENTS.md .agents --glob '!**/node_modules/**' --glob '!**/.git/**'` 无输出，确认旧路径无代码或项目文档引用残留。
- Tests(FB-9): 已通过 `go run github.com/rhysd/actionlint/cmd/actionlint@latest -no-color .github/workflows/nightly-test-and-build.yml`、`ruby -e 'require "yaml"; YAML.load_file(".github/workflows/nightly-test-and-build.yml"); puts "workflow yaml ok"'`、抽取 `.github/workflows/nightly-test-and-build.yml` 所有 `run: |` block 后执行 `bash -n`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- .github/workflows/nightly-test-and-build.yml openspec/changes/official-plugins-submodule-decoupling/tasks.md`。
- Tests(FB-10): 已通过 `go run github.com/rhysd/actionlint/cmd/actionlint@latest -no-color .github/workflows/main-ci.yml .github/workflows/nightly-test-and-build.yml .github/workflows/release-test-and-build.yml .github/workflows/reusable-backend-unit-tests.yml .github/workflows/reusable-windows-command-smoke.yml`、Ruby YAML 解析上述 5 个 workflow、抽取上述 workflow 所有非 `pwsh`/`cmd` 的 `run` block 后执行 `bash -n`、`go test ./hack/tools/build-wasm/... -count=1`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- .github/workflows/main-ci.yml .github/workflows/nightly-test-and-build.yml .github/workflows/release-test-and-build.yml .github/workflows/reusable-backend-unit-tests.yml .github/workflows/reusable-windows-command-smoke.yml hack/tools/build-wasm/internal/builder/builder_plugin_demo_test.go openspec/changes/official-plugins-submodule-decoupling/tasks.md`。补充验证 `go run ./hack/tools/linactl test-go plugins=0 race=false verbose=false` 已进入 host-only 模式并完成除既有 `apps/lina-core/internal/cmd` panic allowlist 外的包测试；当前失败项为既有未纳入本次变更的 `apps/lina-core/internal/service/hostlock/hostlock.go:New` panic allowlist 缺口。
- Tests(FB-11): 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test-scripts`、`find hack/scripts -maxdepth 1 -type f -print | wc -l` 输出 `0`、旧入口残留静态扫描无输出（`hack/scripts/(prepare-packed-assets|stop-dev-services).sh`、`test-scripts requires POSIX`、`filepath.Glob(...*.sh)`、`"bash", script`）、`rg --files -g '*.sh' -g '*.ps1' -g '*.cmd'` 确认默认仓库工具入口仅剩 `make.cmd`，其余 `.sh` 为既有 CI/前端专用脚本。
- Tests(FB-12): 已通过 `go run github.com/rhysd/actionlint/cmd/actionlint@latest -no-color .github/workflows/nightly-test-and-build.yml`、Ruby YAML 解析 `.github/workflows/nightly-test-and-build.yml`、抽取 `.github/workflows/nightly-test-and-build.yml` 所有 Bash `run` block 执行 `bash -n`、`openspec validate official-plugins-submodule-decoupling --strict`、`git diff --check -- .github/workflows/nightly-test-and-build.yml openspec/changes/official-plugins-submodule-decoupling/tasks.md`。
