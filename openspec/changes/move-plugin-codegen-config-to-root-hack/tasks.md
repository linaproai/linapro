## Implementation

- [x] **T-1**：更新 OpenSpec delta specs，覆盖`linactl-build-tool-consolidation`、`module-decoupling`和`plugin-capability-boundary-governance`的新路径、目标模型与治理行为。
- [x] **T-2**：改造`hack/tools/linactl/internal/goframecli`，将目标解析升级为`workDir`和`configDir`，并保持非插件目标读取`dir/hack`。
- [x] **T-3**：收敛`linactl ctrl`和`linactl dao`参数，只允许`dir=`，删除`p=`、`plugin=`和`target=`，并补充失败用例。
- [x] **T-4**：改造`plugins.check`扫描插件根`hack/config.yaml`，阻断`backend/hack/config.yaml`，并覆盖已有`backend/internal/dao`但缺少根配置的失败用例。
- [x] **T-5**：迁移所有官方插件的`backend/hack/config.yaml`到插件根`hack/config.yaml`，不为无 DAO 配置的插件创建空配置。
- [x] **T-6**：同步更新`.agents/rules/plugin.md`、`apps/lina-plugins/README.md`、`apps/lina-plugins/README.zh-CN.md`、`hack/tools/linactl/README.md`和`hack/tools/linactl/README.zh-CN.md`。
- [x] **T-7**：审查`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`是否需要同步更新，并记录结论。结论：两份文档描述插件公开契约和 host service 边界，不描述插件本地工具配置、代码生成或自定义构建指令，无需修改。
- [x] **T-8**：运行验证：`openspec validate move-plugin-codegen-config-to-root-hack --strict`、`cd hack/tools/linactl && go test ./... -count=1`、`make plugins.check`、代表性插件`make dao`和`make ctrl`烟测、旧路径静态检索。

## Impact Records

- [x] **IR-1**：记录无运行时 HTTP API、DTO、前端 UI、SQL schema、运行时配置、缓存、数据权限和业务服务语义变化。
- [x] **IR-2**：记录开发工具跨平台影响：路径解析、文件检查和测试夹具使用 Go 标准库路径 API，不引入平台专属脚本语义。
- [x] **IR-3**：记录`i18n`影响：无运行时用户可见文案和语言包变化；仅工具错误提示和文档文本变化。
- [x] **IR-4**：记录数据库规则影响：不修改 SQL；仅迁移 DAO 生成配置，并通过生成路径、治理检查和 Go 编译门禁验证。

## Feedback

- [x] **FB-1**：移除`ctrl`和`dao`参数校验中对`p=`、`plugin=`和`target=`的专门分支，统一只允许`dir=`。
- [x] **FB-2**：移除根`Makefile`的`ctrl`和`dao`对`p=`的包装兼容，只允许通过`dir=`转发显式目标。
- [x] **FB-3**：移除`apps/lina-plugins`工作区级`ctrl`和`dao`对`p=`的包装兼容，只允许通过`dir=`转发显式目标。
- [x] **FB-4**：移除`hack/makefiles/database.mk`中对旧参数的显式识别和报错分支，使该文件只表达`dir=`转发。
- [x] **FB-5**：移除`apps/lina-plugins/Makefile`中`ctrl`和`dao`对旧`p=`参数的显式识别和报错分支，确保该入口在删除前不保留兼容逻辑。
- [x] **FB-6**：删除`apps/lina-plugins/Makefile`工作区级重复入口，统一由仓库根目录维护`make wasm`，由插件目录本地`Makefile`维护`make ctrl`和`make dao`。
- [x] **FB-7**：将插件自定义构建指令从插件`Makefile`变量收敛到插件根`hack/config.yaml`，并确保根目录`make build`未指定`dir`时遍历所有插件，指定插件`dir`时只构建该插件。
- [x] **FB-8**：补充`apps/lina-plugins`中英文 README，说明插件自定义构建指令位于插件根`hack/config.yaml`的`build.commands`下。
- [x] **FB-9**：修正`hack/config.yaml`的文档定位，明确其是插件本地工具配置入口，而不是围绕单一代码生成场景描述。
- [x] **FB-10**：删除`apps/lina-plugins`根`go.mod`、`go.sum`、`lina-plugins.go`和`package.json`，改为由`linactl`在插件完整构建时自动生成源码插件聚合模块和临时`go.work`。
- [x] **FB-11**：修复`make dev`在已有旧版动态插件 release snapshot 包含独立`config`或`cron`宿主服务声明时启动失败的问题。

### FB-7 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | 从`agentbox`合入的`linactl build`实现读取插件根`Makefile`中的`PLUGIN_BUILD_STEP_N`变量，与项目将插件开发期工具配置集中到插件根`hack/config.yaml`的治理方向不一致。 |
| 实现 | `linactl build`改为读取插件根`hack/config.yaml`中的`build.commands`，支持`$(PLUGIN_ROOT)`和`$(REPO_ROOT)`变量展开；未指定`dir=`时遍历`apps/lina-plugins`下所有包含`plugin.yaml`的直接插件目录，指定`dir=apps/lina-plugins/<plugin-id>`时只执行该插件构建。 |
| 测试 | 新增和更新`hack/tools/linactl/main_test.go`中的构建测试，覆盖默认遍历所有插件、定向构建单个插件、配置变量展开、未声明`build.commands`时跳过自定义构建。 |
| 文档与规范 | 更新`linactl-build-tool-consolidation`和`module-decoupling`增量规范、`.agents/rules/plugin.md`以及`hack/tools/linactl`中英文 README，统一说明插件自定义构建指令位于插件根`hack/config.yaml`。 |
| 开发工具跨平台影响 | 变更位于`linactl` Go 工具链，路径解析、YAML 解析、插件目录遍历和命令执行均使用 Go 标准库与现有跨平台命令入口，不新增平台专属脚本。 |
| `i18n`影响 | 无运行时 UI 文案、语言包、插件清单多语言或 API 文档本地化资源变更；仅修改开发工具文档和 OpenSpec 文本。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更。 |
| 数据权限影响 | 无运行时数据读取、写入、列表、详情、导出或权限边界变更。 |
| DI 来源检查 | 无新增运行期依赖、服务构造函数、启动装配、插件宿主服务适配器或`WASM host service`依赖。 |
| 验证 | `go test ./hack/tools/linactl/... -count=1`通过；`openspec validate move-plugin-codegen-config-to-root-hack --strict`通过；`git diff --check`通过；静态检索确认无`PLUGIN_BUILD_STEP`、`build hook`、`lina-core-api`或`default-admin-workspace`残留。 |

### FB-8 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | FB-7 初次更新文档时只覆盖主框架`linactl`README、规则文件和 OpenSpec 文档，遗漏了插件工作区自身的目录级 README。 |
| 实现 | 更新`apps/lina-plugins/README.md`和`README.zh-CN.md`，说明插件自定义构建指令通过插件根`hack/config.yaml`的`build.commands`声明，根目录`make build`默认遍历插件，传入插件`dir`时只构建指定插件。 |
| 测试策略 | 纯文档治理补充，无运行时行为变化，不新增单元测试或 E2E；使用 OpenSpec strict 校验、Markdown 静态检索和`git diff --check`验证。 |
| `i18n`影响 | 无运行时用户可见文案、语言包、插件清单多语言或 API 文档本地化资源变更。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更。 |
| 数据权限影响 | 无运行时数据读取、写入、列表、详情、导出或权限边界变更。 |
| 开发工具跨平台影响 | 无开发工具或脚本行为变更，仅补充跨平台构建入口说明。 |

### FB-9 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | FB-8 文档把代码生成场景放在插件根`hack/config.yaml`的主定位上，容易误导读者认为该文件主要服务单一工具链；实际定位应是插件本地工具配置入口。 |
| 实现 | 更新`.agents/rules/plugin.md`、`apps/lina-plugins`中英文 README 和`module-decoupling`增量规范，将该文件描述为插件本地工具配置文件；GoFrame 相关说明仅保留在`gfcli.gen.dao`代码生成子场景中。 |
| 测试策略 | 纯文档和规范措辞修正，无运行时行为变化，不新增单元测试或 E2E；使用 OpenSpec strict 校验、静态检索和`git diff --check`验证。 |
| `i18n`影响 | 无运行时用户可见文案、语言包、插件清单多语言或 API 文档本地化资源变更。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更。 |
| 数据权限影响 | 无运行时数据读取、写入、列表、详情、导出或权限边界变更。 |
| 开发工具跨平台影响 | 无开发工具或脚本行为变更，仅修正文档对插件工具配置入口的描述。 |

### FB-10 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | `apps/lina-plugins`根`go.mod`、`go.sum`和`lina-plugins.go`把源码插件聚合关系作为手写事实源维护，和`make build`、`make dev`已自动生成临时`go.work`的方向不一致；根`package.json`仅保存前端工作区元数据，无脚本、依赖或活动引用。 |
| 实现 | `linactl`插件完整模式改为始终生成`temp/official-plugins`聚合模块，并忽略官方插件工作区根 Go module；`temp/go.work.plugins`只包含宿主模块、生成聚合模块和各插件自身 Go module。删除`apps/lina-plugins`根`go.mod`、`go.sum`、`lina-plugins.go`和`package.json`，同步更新插件工作区 README 和相关 OpenSpec delta。 |
| 测试 | 更新`hack/tools/linactl/main_test.go`，覆盖根`go.mod`残留时仍生成聚合模块并自动导入源码插件 backend；更新`apps/lina-core/internal/service/plugin/internal/testutil`，让动态插件构建测试 helper 统一生成聚合模块。 |
| 开发工具跨平台影响 | 变更位于`linactl` Go 工具链和 Go 测试 helper，目录扫描、临时文件生成和路径渲染继续使用 Go 标准库；未新增平台专属脚本或 shell 语义。 |
| `i18n`影响 | 无运行时 UI 文案、菜单、按钮、语言包、插件清单多语言或 API 文档本地化资源变更；仅调整开发工具行为、测试和文档说明。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更。 |
| 数据权限影响 | 无运行时数据读取、写入、列表、详情、导出或租户/组织边界变更。 |
| DI 来源检查 | 无新增运行期依赖、服务构造函数、启动装配、插件宿主服务适配器或`WASM host service`依赖。 |
| 测试策略 | 属于开发工具和构建接线行为变化，不涉及用户可观察 UI 或端到端业务流程；使用`linactl`单元测试、宿主相关 Go 编译烟测、OpenSpec strict 校验和静态检索验证。 |
| 验证 | `go test ./hack/tools/linactl/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/testutil -count=1`通过；`GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS='-tags=official_plugins' go test ./internal/cmd -count=1`通过；`openspec validate move-plugin-codegen-config-to-root-hack --strict`通过；`git diff --check`和插件子仓库`git diff --check`通过；静态检索确认无活动文档继续要求维护根`lina-plugins.go`、根`go.mod`、根`go.sum`或根`package.json`。 |

### FB-11 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | `make dev`的 Go 编译和动态 WASM 打包均成功，失败发生在后端启动期`BootstrapAutoEnable`同步插件 manifest。当前动态插件 artifact 已声明合法的`plugins.config.get`和`jobs.jobs.register`，但本地数据库中同版本旧`sys_plugin_release.manifest_snapshot`仍保存历史`service: config`和`service: cron`声明；同步新 artifact 前会先解析旧 snapshot 以判断授权是否可继承，旧名称未被迁移导致解析失败并中止启动。 |
| 实现 | 扩展旧 release snapshot 迁移逻辑，仅在解析持久化 snapshot 时将`config.get`迁移为`plugins.config.get`，将`cron.register`或`register`迁移为`jobs.jobs.register`；新 manifest 和新 WASM artifact 继续使用严格校验，不重新允许独立`config`或`cron`服务声明。 |
| 测试 | 新增`TestSyncManifestMigratesLegacyStandaloneConfigSnapshot`和`TestSyncManifestMigratesLegacyCronSnapshot`，覆盖旧 snapshot 可解析、启动同步可重写为当前声明、同版本 release 不再阻断后端启动。 |
| 架构影响 | 变更局限在插件治理 store 的持久化快照兼容解析，不改变插件宿主公开契约、HTTP API、模块边界、启动装配或动态插件 host service 新声明规则。 |
| 开发工具跨平台影响 | 无开发工具或脚本实现变更；实际验证覆盖`make dev`默认入口，服务启动成功后已执行`make stop`确认本地服务停止。 |
| `i18n`影响 | 无运行时 UI 文案、菜单、按钮、语言包、插件清单多语言或 API 文档源文本变更。 |
| 缓存一致性影响 | 未新增缓存层；继续使用`manifest_snapshot`内容哈希缓存，权威数据源仍为`sys_plugin_release.manifest_snapshot`。同步新 manifest 成功后既有写入路径会刷新 release snapshot 并复用既有失效/刷新逻辑。 |
| 数据权限影响 | 无业务数据读取、列表、详情、导出、租户/组织可见性或权限边界变更；仅解析和重写插件治理元数据。 |
| DI 来源检查 | 无新增运行期依赖、服务构造函数、启动装配、插件宿主服务适配器或`WASM host service`依赖。 |
| 测试策略 | 属于后端启动期插件治理兼容修复，不涉及用户可观察 UI 或 E2E 业务流程；使用 store 单元测试、`internal/cmd`插件完整编译烟测、实际`make dev`启动验证和 OpenSpec strict 校验。 |
| 验证 | `cd apps/lina-core && go test ./internal/service/plugin/internal/store -count=1`通过；`GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins GOFLAGS='-tags=official_plugins' go test ./internal/cmd -count=1`通过；`go test ./hack/tools/linactl/... -count=1`通过；`make dev`通过并已`make stop`；`openspec validate move-plugin-codegen-config-to-root-hack --strict`通过。 |
