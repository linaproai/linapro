## ADDED Requirements

### Requirement: 静态检查必须支持按组件目录定向到单个 Go module

系统 SHALL 支持通过可选参数`dir=<path>`将`linactl lint.go`与`make lint`/`make lint.go`的扫描范围收敛到目标路径所属的单个`Go module`。未传入`dir`时，命令 MUST 保持现有工作区全量扫描行为（宿主模式或插件完整模式）。`dir`解析、workspace 过滤与错误处理 MUST 在`linactl`的 Go 实现中完成，根`Makefile`与`make.cmd`仅允许透传参数。

#### Scenario: 定向扫描宿主核心模块

- **WHEN** 开发者运行`make lint dir=apps/lina-core`或等价`linactl lint.go dir=apps/lina-core plugins=0`
- **THEN** 命令仅对`apps/lina-core`对应`Go module`执行`golangci-lint`与死代码检查
- **AND** 不对其它工作区 module（例如`hack/tools/linactl`）执行扫描
- **AND** 输出明确标识定向范围（例如`scope=dir`与目标 module）

#### Scenario: 定向扫描工具链模块

- **WHEN** 开发者运行`make lint dir=hack/tools/linactl plugins=0`
- **THEN** 命令仅对`hack/tools/linactl`对应`Go module`执行静态检查

#### Scenario: 插件根目录解析到 backend module

- **WHEN** 开发者运行`make lint dir=apps/lina-plugins/<plugin-id> plugins=1`
- **AND** 该插件目录包含`plugin.yaml`且存在`backend/go.mod`
- **THEN** 命令将扫描目标解析为该插件`backend` module
- **AND** 仅对该 module 执行静态检查

#### Scenario: 子目录向上解析到所属 module

- **WHEN** 开发者传入位于某个`go.mod`之下的子目录作为`dir`
- **THEN** 命令向上查找最近的`go.mod`（不超过仓库根）
- **AND** 仅对该 module 执行静态检查

#### Scenario: 目标 module 不在当前工作区时失败

- **WHEN** 开发者传入的`dir`解析到某个`Go module`
- **AND** 该 module 不在当前`plugins`模式下的工作区 module 列表中
- **THEN** 命令失败并返回非零退出码
- **AND** 错误消息说明 module 不在当前工作区，以及必要时使用`plugins=1`或初始化插件工作区

#### Scenario: 无效目录拒绝执行

- **WHEN** 开发者传入空的`dir`、不存在的路径，或无法在仓库根内解析到`go.mod`的路径
- **THEN** 命令失败并返回非零退出码
- **AND** 不得静默回退为全量工作区扫描

#### Scenario: 未传 dir 时保持全量行为

- **WHEN** 开发者运行`make lint.go plugins=0`且不传`dir`
- **THEN** 命令继续扫描宿主工作区中的全部`Go module`
- **AND** 行为与引入`dir`参数前保持兼容

## MODIFIED Requirements

### Requirement: 静态检查治理必须有文档和验证记录

系统 SHALL 在开发工具文档和规则说明中记录`Go`静态检查入口、参数、版本升级方式、跨平台边界、插件模式覆盖、可选`dir`定向范围和失败处理方式。实现任务 MUST 记录跨平台影响、测试策略、`i18n`影响判断、缓存一致性无影响判断、数据权限无影响判断和实际验证命令。

#### Scenario: 开发者查看工具文档

- **WHEN** 开发者查看`linactl`或仓库开发工具说明文档
- **THEN** 文档说明如何运行宿主模式和插件完整模式`Go`静态检查
- **AND** 文档说明如何使用`dir=<path>`定向到单个组件/`Go module`
- **AND** 文档说明`golangci-lint`和`staticcheck`版本由仓库锁定
- **AND** 文档说明自动修复入口如存在必须由开发者显式触发
- **AND** 文档说明`CI`与审查门禁仍以未传`dir`的工作区扫描为准

#### Scenario: 审查静态检查变更

- **WHEN** `lina-review`审查本变更实现
- **THEN** 审查结论必须包含开发工具跨平台影响和验证方式
- **AND** 审查结论必须记录无运行时`i18n`资源影响、无缓存一致性影响、无数据权限影响和无运行期服务依赖影响
