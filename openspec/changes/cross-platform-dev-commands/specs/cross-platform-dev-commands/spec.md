## ADDED Requirements

### Requirement: 跨平台命令主入口

系统 MUST 提供不依赖 GNU Make 或 POSIX Shell 的跨平台开发命令主入口，用于执行项目常用开发、构建、初始化和验证任务。

#### Scenario: 使用 Go CLI 执行常用命令

- **WHEN** 开发者在项目根目录执行跨平台命令入口并指定 `dev`、`stop`、`status`、`build`、`wasm`、`init`、`mock`、`test`、`test-go` 或 `help` 等受支持命令
- **THEN** 系统 MUST 通过跨平台实现执行对应任务，而不是要求调用者安装 GNU Make 或 POSIX Shell 工具链

#### Scenario: 缺失外部专业工具

- **WHEN** 受支持命令需要调用 `go`、`pnpm`、`docker`、`kubectl`、`gf` 或 `playwright` 等外部专业工具但当前环境缺失该工具
- **THEN** 系统 MUST 输出明确的缺失工具名称、当前命令上下文和可执行的安装或修复提示

### Requirement: Windows make.cmd 兼容入口

系统 MUST 在项目根目录提供 Windows `make.cmd` 薄包装入口，并将所有真实任务逻辑委托给跨平台命令主入口。

#### Scenario: cmd.exe 使用 make 风格命令

- **WHEN** Windows `cmd.exe` 用户在项目根目录执行 `make dev`
- **THEN** `make.cmd` MUST 将命令和参数转发给跨平台命令主入口执行 `dev` 任务

#### Scenario: PowerShell 使用当前目录 make 入口

- **WHEN** Windows PowerShell 用户在项目根目录执行 `.\make dev` 或 `.\make.cmd dev`
- **THEN** `make.cmd` MUST 将命令和参数转发给跨平台命令主入口执行 `dev` 任务，且不得要求维护额外的 `make.ps1`

#### Scenario: 参数透传

- **WHEN** Windows 用户执行 `make init confirm=init rebuild=true`
- **THEN** `make.cmd` MUST 原样透传 `init`、`confirm=init` 和 `rebuild=true` 参数，并返回跨平台命令主入口的退出码

### Requirement: Makefile 兼容层

系统 MUST 保留现有 `Makefile` 入口作为 Linux/macOS 与既有 CI 的兼容层，并避免在已迁移目标中继续维护复杂 shell 编排逻辑。

#### Scenario: 既有 make 目标继续可用

- **WHEN** Linux/macOS 开发者执行现有 `make dev`、`make build`、`make init confirm=init` 等命令
- **THEN** 系统 MUST 保持目标语义兼容，并通过跨平台命令主入口或等价跨平台工具实现任务编排

#### Scenario: 薄包装一致性

- **WHEN** 同一个目标同时支持 `make <target>`、`make.cmd <target>` 和跨平台命令主入口
- **THEN** 三种入口 MUST 调用同一套任务实现，并保持参数语义、退出码和关键错误信息一致

### Requirement: Make 风格参数兼容

跨平台命令主入口 MUST 支持现有 make 风格 `key=value` 参数，并可将其映射到内部命令选项。

#### Scenario: 初始化参数兼容

- **WHEN** 开发者执行跨平台命令主入口 `init confirm=init rebuild=true`
- **THEN** 系统 MUST 将 `confirm=init` 和 `rebuild=true` 识别为初始化任务参数，并保持与现有 `make init confirm=init rebuild=true` 等价的行为

#### Scenario: 构建参数兼容

- **WHEN** 开发者执行跨平台命令主入口 `build platforms=linux/amd64,linux/arm64 verbose=1`
- **THEN** 系统 MUST 将 `platforms` 和 `verbose` 识别为构建任务参数，并按指定平台和日志模式执行构建流程

#### Scenario: 插件参数兼容

- **WHEN** 开发者执行跨平台命令主入口 `wasm p=plugin-demo-dynamic`
- **THEN** 系统 MUST 仅对指定插件执行动态 Wasm 构建，并保持与现有 `make wasm p=plugin-demo-dynamic` 等价的行为

### Requirement: Shell 依赖治理

系统 MUST 将常用命令路径中的文件复制、目录清理、进程启动、端口检测、HTTP readiness、插件发现和帮助输出等逻辑迁移到跨平台实现，避免依赖 Linux/macOS 专属命令。

#### Scenario: 文件复制和目录清理

- **WHEN** 构建流程需要准备嵌入资源或清理输出目录
- **THEN** 系统 MUST 使用跨平台文件系统实现完成目录创建、删除和复制，而不是依赖 `rm`、`cp`、`mkdir` 或 `touch`

#### Scenario: 服务状态和停止

- **WHEN** 开发者执行 `status` 或 `stop`
- **THEN** 系统 MUST 使用跨平台方式检查端口、PID 文件、进程启动记录和服务 readiness，并避免依赖 `lsof`、`pgrep`、`ps`、`kill`、`sed`、`head` 或 `tr`

#### Scenario: 帮助输出

- **WHEN** 开发者执行 `help`
- **THEN** 系统 MUST 从跨平台命令定义或等价结构输出可用命令说明，而不是依赖 `awk`、`sort` 或 POSIX 管道

### Requirement: 文档和验证

系统 MUST 为跨平台命令入口提供中英文文档和自动化验证，确保 Windows、Linux、macOS 使用方式清晰可执行。

#### Scenario: README 展示跨平台入口

- **WHEN** 开发者阅读根目录 README 和中文镜像 README
- **THEN** 文档 MUST 明确展示跨平台推荐命令、Windows `cmd.exe` 中省略 `.cmd` 后缀的 `make` 使用方式、PowerShell 的 `.\make` 使用方式、必要时可显式使用的 `.\make.cmd` 方式，以及 Linux/macOS 的兼容 `make` 使用方式

#### Scenario: 工具文档同步

- **WHEN** 新增跨平台工具目录
- **THEN** 该目录 MUST 同步提供英文 `README.md` 和中文 `README.zh-CN.md`，并说明命令、参数、示例、输出和验证方式

#### Scenario: 自动化验证

- **WHEN** 完成跨平台命令入口实现
- **THEN** 系统 MUST 通过 Go 单元测试、命令级 smoke 测试、静态扫描或等价治理验证覆盖参数解析、入口转发、文件操作、插件发现、帮助输出和 Makefile 薄包装一致性

### Requirement: GitHub Actions Windows 验证

系统 MUST 在 `.github/workflows/` 下与构建、测试或工具验证相关的 GitHub Actions 中增加 Windows runner 基本命令验证，确保 Windows 下跨平台命令入口和 `make.cmd` 可用。

#### Scenario: Windows runner 验证 Go CLI 基本命令

- **WHEN** GitHub Actions 在 `windows-latest` runner 上执行跨平台命令验证
- **THEN** workflow MUST 至少验证 `go run ./hack/tools/linactl help`、`go run ./hack/tools/linactl status` 和一个不依赖外部服务的文件或插件工具命令能够成功执行

#### Scenario: Windows runner 验证 make.cmd

- **WHEN** GitHub Actions 在 `windows-latest` runner 上执行 Windows 入口验证
- **THEN** workflow MUST 验证 `make.cmd` 能够透传参数到跨平台命令主入口，并至少覆盖 `help` 或 `status` 这类轻量命令

#### Scenario: cmd.exe 与 PowerShell 使用方式

- **WHEN** GitHub Actions 验证 Windows 命令入口
- **THEN** workflow MUST 覆盖 `cmd.exe` 中的 `make <target>` 或 `make.cmd <target>` 用法，以及 PowerShell 中的 `.\make <target>` 或 `.\make.cmd <target>` 用法，确保文档声明的两类 Windows 终端入口均可执行

#### Scenario: Windows CI 成本控制

- **WHEN** 基本命令需要外部数据库、Docker daemon、长驻开发服务或完整前端依赖才能执行
- **THEN** workflow MUST 使用轻量 smoke、dry-run、跳过重型路径或拆分后续验证，避免把 Windows 基本命令验证绑定到不稳定或高成本的外部依赖

### Requirement: 运行时影响隔离

跨平台命令入口变更 MUST 不改变 LinaPro 后端运行时 API、数据库 Schema、权限模型、数据权限、插件运行时契约、前端用户界面、运行时 i18n 资源或业务缓存行为。

#### Scenario: 运行时资源不受影响

- **WHEN** 跨平台命令入口变更完成
- **THEN** 系统 MUST 不新增或修改运行时菜单、按钮、表单、接口消息、插件 manifest 文案、运行时 i18n JSON、manifest i18n、apidoc i18n 或业务缓存失效逻辑

#### Scenario: 外部工具只包装不重写

- **WHEN** 跨平台命令入口需要执行构建、代码生成、容器镜像、部署或测试任务
- **THEN** 系统 MUST 通过清晰的命令包装调用 `go`、`pnpm`、`docker`、`kubectl`、`gf`、`playwright` 等专业工具，而不得在本次变更中重写这些工具的核心能力
