## Context

`LinaPro`当前根`go.work`只包含`apps/lina-core`和`hack/tools/linactl`，官方插件源码位于`apps/lina-plugins`并作为`submodule`按需参与插件完整模式。现有`CI`已经通过可复用验证套件组织主流程、发布流程、宿主模式和插件完整模式测试，但缺少统一的`Go`静态检查门禁。

`gf`项目的参考做法是根目录维护`.golangci.yml`、通过`make lint`调用`golangci-lint run`，并使用独立`GitHub Actions`工作流运行官方`golangci-lint`action。该模式可以借鉴配置和`CI`工具选择，但不能直接照搬到`LinaPro`，因为本仓库的长期工具逻辑必须收敛到`linactl`，并且需要覆盖官方插件`submodule`下的多个`Go module`。

## Goals / Non-Goals

**Goals:**

- 建立仓库级`Go`静态检查配置，覆盖错误处理、静态分析、基础格式化和`nolint`治理。
- 通过`linactl lint.go`和根`Makefile`提供跨平台本地入口。
- 支持宿主模式和插件完整模式，使`apps/lina-core`、`hack/tools/linactl`和官方插件`Go module`都能被检查。
- 将`Go`静态检查接入现有可复用`CI`验证套件，并作为全量质量门禁阻断。
- 同步更新工具文档和规则描述，明确运行方式、版本锁定、跨平台边界、验证方式和影响分析。

**Non-Goals:**

- 不修改运行时`HTTP API`、数据库结构、前端页面、权限模型、缓存机制或插件运行时契约。
- 不把`golangci-lint`作为运行时依赖引入任何业务模块。
- 不在本次变更中启用高噪声或需要专项治理的安全/复杂度规则，例如`gosec`、`gocognit`、`exhaustruct`和`paralleltest`。
- 不要求一次性重写所有既有代码风格；如果首轮扫描暴露问题，应以修复或少量有理由的`nolint`为主，而不是降低门禁标准。

## Decisions

### 版本固定

仓库新增`.golangci-lint-version`并在`CI`和文档中引用该版本，避免开发者本地和`CI`使用不同`golangci-lint`版本导致结果漂移。`.golangci.yml`使用`golangci-lint`当前`v2`配置结构，便于直接使用`linters`和`formatters`分区。

替代方案是使用`latest`或完全依赖`GitHub Action`默认版本。该方式升级成本低，但会让规则、默认行为和诊断输出随上游变化，和可持续交付门禁的可复现性冲突。

### 命令入口收敛到`linactl`

新增`linactl lint.go [plugins=auto|0|1] [fix=true]`，根`Makefile`只作为薄包装转发到`linactl`。命令实现必须复用已有`prepareOfficialPluginBuildEnv`和`goWorkspaceModules`思路，避免在`Makefile`或`Shell`中复制插件工作区扫描逻辑。

替代方案是像`gf`一样在`Makefile`中直接调用`golangci-lint run -c .golangci.yml`。该方式简单，但无法自然支持`Windows make.cmd`、插件完整模式临时`go.work`和现有工具治理规则。

### 扫描范围按插件模式决定

`plugins=0`使用根`go.work`扫描宿主与`linactl`模块。`plugins=1`要求官方插件工作区可用，生成临时插件完整`go.work`后扫描宿主、`linactl`、源码插件、动态插件构建相关`Go module`和自动生成的官方插件聚合模块。`plugins=auto`沿用现有插件模式解析：发现官方插件清单时启用插件完整模式，否则使用宿主模式。

替代方案是始终遍历所有`go.mod`目录逐个执行`golangci-lint`。该方式覆盖直接，但重复下载、缓存和诊断聚合成本更高，也绕开了仓库已经维护的`go.work`语义。

### CI 作为全量阻断门禁

新增可复用`Go`静态检查工作流并接入现有`reusable-test-verification-suite.yml`。主`CI`和发布验证默认启用宿主模式与插件完整模式检查。`CI`可以使用官方`golangci-lint-action`安装固定版本，但实际执行命令应调用`make lint.go`或`linactl lint.go`，保证本地与`CI`路径一致。

替代方案是新增完全独立的`golangci-lint.yml`。该方式接近`gf`，但会让质量门禁分散在主验证套件之外，后续 nightly、release、host-only 和 plugin-full 配置容易漂移。

### 首批规则控制噪声

首批启用规则以稳定收益为主：`errcheck`、`errchkjson`、`govet`、`staticcheck`、`revive`、`gocritic`、`misspell`、`nolintlint`、`usestdlibvars`、`whitespace`和`goconst`。格式化器首批仅启用`gofmt`，避免把本次治理扩大为全仓导入顺序迁移；`goimports`和`gci`在后续单独规范化导入分组后再作为阻断门禁启用。

替代方案是一次性启用更严格的复杂度、安全和测试风格规则。该方式覆盖更广，但初始噪声更高，容易把本次治理变成大规模代码风格整改。

## Risks / Trade-offs

- `golangci-lint`首轮扫描可能暴露大量既有问题 → 优先修复真实问题；确属误报或生成代码场景时使用路径排除或有理由的`nolint`，不得无差别关闭核心规则。
- 插件完整模式会增加`CI`耗时 → 复用`go.work`和 action 缓存；如果耗时明显上升，再单独评估按变更路径触发，而不是降低默认覆盖。
- `golangci-lint`版本升级可能改变诊断结果 → 通过`.golangci-lint-version`固定版本，并把升级作为独立治理变更处理。
- `fix=true`可能批量改写导入和格式 → 默认检查不自动修复；自动修复仅作为开发者显式入口，任务实现时需记录验证方式。
- 生成代码可能触发不适合人工维护的诊断 → 配置中明确按生成文件标记和生成目录排除，避免要求手改`dao`、`do`、`entity`等生成源码。

## Migration Plan

1. 添加配置、命令入口和文档后，在本地先运行`make lint.go plugins=0`修复宿主与`linactl`问题。
2. 初始化官方插件工作区后运行`make lint.go plugins=1`修复插件完整模式问题。
3. 接入`CI`可复用工作流，并确认主`CI`和发布验证都能通过。
4. 若需要回滚，移除`CI`调用即可临时解除门禁；配置和本地命令可保留为非阻断工具，待问题修复后重新启用。

## Open Questions

- 首次实现时是否命名聚合入口为`make lint`并让其转发到`make lint.go`，或只提供`make lint.go`以便后续扩展前端 lint。默认建议同时提供`lint.go`和聚合`lint`，其中`lint`当前只运行`Go`静态检查。
