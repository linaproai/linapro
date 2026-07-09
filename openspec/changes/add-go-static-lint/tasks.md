## 1. 静态检查配置

- [x] 1.1 新增仓库根目录`.golangci.yml`，配置首批`linters`、`formatters`、生成代码排除和`nolint`治理策略。
- [x] 1.2 新增仓库根目录`.golangci-lint-version`，固定`golangci-lint`版本，并确保`CI`和文档引用同一版本来源。
- [x] 1.3 本地运行配置语法检查或等价`golangci-lint`空跑验证，确认配置可被当前固定版本解析。

## 2. `linactl`和`Makefile`入口

- [x] 2.1 在`hack/tools/linactl`新增`lint.go`命令注册和命令实现，支持省略`plugins`自动探测、显式`plugins=0|1`与显式`fix=true`参数。
- [x] 2.2 复用现有官方插件工作区和临时`go.work`准备逻辑，使`plugins=0`覆盖宿主工作区，`plugins=1`覆盖官方插件完整工作区。
- [x] 2.3 在根`Makefile`拆分片段中新增`lint.go`薄转发目标，并按最终决策新增或保留聚合`lint`目标。
- [x] 2.4 为`linactl lint.go`补充单元测试，覆盖命令注册、参数解析、宿主模式分发、插件完整模式分发和插件工作区缺失错误。

## 3. `CI`集成

- [x] 3.1 新增可复用`GitHub Actions`工作流，用固定版本安装`golangci-lint`并通过仓库`make`或`linactl`入口执行检查。
- [x] 3.2 将`Go`静态检查接入`reusable-test-verification-suite.yml`，并在主`CI`和发布验证中默认启用宿主模式与插件完整模式。
- [x] 3.3 确认`CI`不以`only-new-issues`作为长期默认豁免策略，静态检查失败必须阻断主验证和发布验证。

## 4. 文档和规范

- [x] 4.1 更新`hack/tools/linactl/README.md`和`README.zh-CN.md`，说明`lint.go`命令、参数、插件模式、版本锁定、自动修复入口和验证方式。
- [x] 4.2 更新相关开发工具或后端治理规则，记录`Go`静态检查门禁、跨平台要求、`CI`执行策略和审查要求。
- [x] 4.3 记录本变更影响判断：无运行时`i18n`资源影响、无缓存一致性影响、无数据权限影响、无运行期服务依赖影响、无需`E2E`测试。

## 影响判断记录

- 跨平台影响：默认本地入口收敛到`linactl lint.go`，根`Makefile`和`make.cmd`保持薄包装；命令实现使用`Go`路径、环境变量和子进程能力，支持`Windows`、`Linux`和`macOS`。`golangci-lint`和`staticcheck`由`linactl`按仓库锁定版本自动检测和安装，`CI`与本地复用同一路径。
- `i18n`影响：本变更只新增开发工具命令、`CI`门禁和技术文档，不新增运行时用户可见文案、菜单、路由、接口文档源文本或翻译资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：不新增运行时缓存、缓存失效、订阅状态、权限快照或分布式协调路径，确认无缓存一致性影响。
- 数据权限影响：不新增或修改列表、详情、导出、下拉、批量接口或数据可见性逻辑，确认无数据权限影响。
- 运行期服务依赖影响：不新增`Controller`、`Middleware`、`Service`、插件宿主服务适配器或`WASM host service`运行期依赖，确认无需`DI`来源检查。
- 测试策略影响：本变更属于开发工具与`CI`治理，不改变前端页面、用户交互或端到端业务流程，确认无需新增`E2E`测试；验证聚焦`linactl`单元测试、`golangci-lint`配置检查、宿主/插件模式静态扫描和`OpenSpec`严格校验。
- FB-1影响：根据`temp/lint.log`审查首批严格`linter`的实际诊断后，关闭`contextcheck`、`dupl`、`tagliatelle`、`intrange`、`perfsprint`、`prealloc`、`mirror`、`gocyclo`、`funlen`、`nestif`和`maintidx`；保留`errorlint`、`noctx`、`ineffassign`、`dupword`、`errname`、`containedctx`和`cyclop`等更接近真实缺陷、错误处理、资源释放或复杂度基线的门禁。已修复`temp/lint.log`中`dupword`、`errname`、`errorlint`、`ineffassign`、`noctx`、`predeclared`和`containedctx`诊断；对既有高复杂度函数保留精准`nolint:cyclop`说明，避免本轮扩大为大型重构。仅修改`.golangci.yml`、后端`Go`源码和本任务记录，不新增运行时`i18n`资源、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖；开发工具跨平台入口不变。已运行受影响`Go`包测试并通过；已运行`make lint`，`golangci-lint`阶段为`0 issues`，命令随后被独立`staticcheck U1000`死代码门禁中的既有未使用符号阻断，需作为死代码治理任务单独处理。
- FB-4影响：仅修改`linactl lint.go`开发工具入口、可复用`CI`工作流和工具文档；不新增运行时`i18n`资源、缓存、数据权限、`HTTP API`、数据库或运行期服务依赖。已运行`cd hack/tools/linactl && go test ./... -count=1`和`openspec validate add-go-static-lint --strict`并通过；已重新运行`make lint.go plugins=0`并通过，验证自动检测到`golangci-lint v2.12.2`。
- FB-5影响：仅修改`linactl lint.go`开发工具入口、可复用`CI`工作流和工具文档；不新增运行时`i18n`资源、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖。`staticcheck`安装流程与`golangci-lint`共用跨平台`Go`工具安装路径，使用`GOWORK=off`并移除`GOFLAGS`、`GOOS`、`GOARCH`等构建变量，避免插件完整模式污染工具构建。已运行`cd hack/tools/linactl && go test ./... -count=1`、`openspec validate add-go-static-lint --strict`和`git diff --check`并通过；`make lint.go plugins=0`已验证工具探测阶段自动检测到`staticcheck v0.7.0`并使用二进制路径，不再通过`go run honnef.co/go/tools/cmd/staticcheck@<version>`执行，命令后续在`golangci-lint`扫描`apps/lina-core`时被32条既有源码质量诊断阻断，尚未进入包级`staticcheck`扫描。
- FB-6影响：仅修改`linactl env.setup`开发工具入口、`make env.setup`帮助文案、`linactl`工具文档和本变更增量规范；不新增运行时`i18n`资源、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖。`env.setup`在前端依赖和`Playwright`安装前复用`lint.go`同一套`golangci-lint`与`staticcheck`锁定版本检测安装路径，安装环境继续使用`GOWORK=off`并移除构建标签和交叉编译变量，保持`Windows`、`Linux`、`macOS`默认入口跨平台。已运行`go test . -run 'TestRunEnvSetup' -count=1`、`cd hack/tools/linactl && go test ./... -count=1`、`make -n env.setup`、`git diff --check`和`openspec validate add-go-static-lint --strict`并通过；无需新增`E2E`测试。
- FB-7影响：修复`staticcheck U1000`死代码治理误删动态插件示例`BeforeInstall`生命周期超时常量导致`GitHub Actions`插件完整模式失败的回归；仅修改`apps/lina-plugins/linapro-demo-dynamic/backend/internal/controller/dynamic/dynamic_lifecycle.go`，恢复动态插件 artifact builder 通过`AST`读取的`BeforeInstallTimeoutMs`导出契约。该修复不新增运行时`i18n`文案、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖；开发工具跨平台入口不变，插件本地`AGENTS.md`不存在。已运行`cd hack/tools/linactl && go test ./internal/wasmbuilder -run TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets -count=1 -v`、`cd hack/tools/linactl && go test ./internal/wasmbuilder -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`和`make lint`并通过；本地未重跑完整`Playwright`分片，原因是需要重置数据库并启动完整服务，已由 artifact 内容回归单测覆盖原失败断言。
- FB-8影响：将动态插件生命周期超时从 Go 源码常量和`backend/lifecycle/*.yaml`隐式/分散约定迁移到插件根`hack/config.yaml`的`wasm.lifecycle.timeouts`配置；`wasmbuilder`使用跨平台 Go 文件读取和`time.ParseDuration`解析带单位时长，旧`backend/lifecycle/*.yaml`不再作为配置来源。仅修改`hack/tools/linactl/internal/wasmbuilder`、`apps/lina-plugins/linapro-demo-dynamic/hack/config.yaml`、动态示例插件生命周期控制器和本任务记录；不新增运行时`i18n`文案、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖，确认无需`DI`来源检查；插件本地`AGENTS.md`不存在。该变更属于内部构建工具和 artifact 元数据治理，不改变用户可观察页面或端到端业务流程，确认无需新增`E2E`测试。已运行`cd hack/tools/linactl && go test ./internal/wasmbuilder -count=1`、`cd hack/tools/linactl && go test ./... -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`、`make lint`和`openspec validate add-go-static-lint --strict`并通过。
- FB-9影响：将动态插件 builder 专用的`hooks`、`resources`和生命周期 timeout 配置统一收敛到插件根`hack/config.yaml`的`wasm.*`配置，并删除`wasmbuilder`与宿主本地动态插件目录加载逻辑中对`backend/hooks/*.yaml`、`backend/resources/*.yaml`、`backend/lifecycle/*.yaml`及其他`backend/*/*.yaml`的读取路径。`wasm.hooks`中的`timeout`和`sleep`必须使用带单位字符串，`timeoutMs`、`sleepMs`等旧字段直接报错。仅修改`hack/tools/linactl/internal/wasmbuilder`、`apps/lina-core/internal/service/plugin/internal/catalog`、`integration`、`runtime`相关测试和本变更增量规范；`apps/lina-core/pkg/plugin` README 未记录旧`backend/*/*.yaml`配置来源，确认无需同步。该变更不新增运行时`i18n`文案、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖，确认无需`DI`来源检查；开发工具跨平台入口仍使用 Go 文件系统和 YAML 解析，支持`Windows`、`Linux`和`macOS`；属于内部构建与 manifest 水合治理，不改变用户可观察页面，确认无需新增`E2E`测试。已运行`cd hack/tools/linactl && go test ./... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`、`openspec validate add-go-static-lint --strict`、`git diff --check`、`git -C apps/lina-plugins diff --check`和`make lint`并通过。
- FB-10影响：删除上一版中旧`backend/*/*.yaml`配置的 fail-fast 检测函数、调用点和对应失败断言测试；`wasmbuilder`与宿主本地动态插件目录加载逻辑仅读取插件根`hack/config.yaml`，不再扫描、检测或报错处理旧`backend/hooks/*.yaml`、`backend/resources/*.yaml`、`backend/lifecycle/*.yaml`及其他`backend/*/*.yaml`。仅修改`hack/tools/linactl/internal/wasmbuilder`、`apps/lina-core/internal/service/plugin/internal/catalog`和本变更增量规范；不新增运行时`i18n`文案、缓存、数据权限、`HTTP API`、数据库、前端`UI`或运行期服务依赖，确认无需`DI`来源检查；插件本地`AGENTS.md`不存在。该变更属于内部构建工具和 manifest 水合治理，不改变用户可观察页面，确认无需新增`E2E`测试。已运行`cd hack/tools/linactl && go test ./internal/wasmbuilder -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog -count=1`、`cd hack/tools/linactl && go test ./... -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./... -count=1`、`openspec validate add-go-static-lint --strict`、`git diff --check`、`git -C apps/lina-plugins diff --check`和`make lint`并通过。

## 5. 验证和审查

- [x] 5.1 运行`go test`覆盖`hack/tools/linactl`新增或修改的命令实现和测试。
- [x] 5.2 运行`make lint.go plugins=0`并修复宿主模式静态检查问题。
- [x] 5.3 在官方插件工作区可用时运行`make lint.go plugins=1`并修复插件完整模式静态检查问题；如环境不可用，记录阻断原因和替代验证。
- [x] 5.4 运行`openspec validate add-go-static-lint --strict`。
- [x] 5.5 完成实现后调用`lina-review`进行代码和规范审查，审查通过后再标记全部任务完成。

## Feedback

- [x] **FB-1**: 扩展更严格的`Go`质量类`linters`，在配置中以注释形式保留完整可用`linter`清单，记录新门禁发现的问题，待审查确认后再修复源码。
- [x] **FB-2**: 修正`unused`门禁对`wasip1` guest 专属代码的误报，改为按宿主与`wasip1/wasm`目标归并死代码检查结果。
- [x] **FB-3**: 注释独立`unused` linter，统一由固定版本`staticcheck U1000`承担死代码门禁，并保留`wasip1/wasm`矩阵归并。
- [x] **FB-4**: 在`linactl lint.go`中按`.golangci-lint-version`自动检测和安装`golangci-lint`，避免本机缺少二进制或版本漂移导致`make lint.go`失败。
- [x] **FB-5**: 在`linactl lint.go`中按`.staticcheck-version`自动检测和安装`staticcheck`，避免死代码检查依赖`go run honnef.co/go/tools/cmd/staticcheck@<version>`临时执行路径。
- [x] **FB-6**: 在`linactl env.setup`中优先按仓库锁定版本检测和安装`golangci-lint`与`staticcheck`，使开发环境初始化提前准备`Go`静态检查工具。
- [x] **FB-7**: 恢复动态插件示例`BeforeInstall`生命周期超时常量的 artifact builder 导出契约，避免`staticcheck U1000`死代码治理误删后导致插件完整模式`Go`单测和`E2E`安装流程失败。
- [x] **FB-8**: 将动态插件生命周期超时等 builder 专用配置统一迁移到插件根`hack/config.yaml`，删除`backend/lifecycle/*.yaml`和源码常量隐式读取路径。
- [x] **FB-9**: 将`wasmbuilder`读取的`backend/hooks/*.yaml`和`backend/resources/*.yaml`统一迁移到插件根`hack/config.yaml`，删除`backend/*/*.yaml`构建配置读取路径。
- [x] **FB-10**: 删除旧`backend/*/*.yaml`配置的 fail-fast 检测和对应报错测试，不保留任何兼容判断逻辑。
