# 开发工具与脚本规则

## 适用范围

本规则约束开发、构建、测试、代码生成、资源打包、服务启停、CI 辅助、仓库治理入口、`linactl`命令实现、脚本目录和跨平台执行。

## 跨平台执行要求

- 所有新增或修改的开发、构建、测试、代码生成、资源打包、服务启停、CI 辅助和仓库治理入口，都必须能在 Windows、Linux、macOS 上执行。
- 禁止默认开发路径依赖单一平台默认存在的命令或语义，例如`bash`、`sh`、`sed`、`awk`、`grep`、`perl`、`lsof`、`pgrep`、`xargs`、`kill`、`rm`、`cp`、`mv`、`mkdir -p`、POSIX 路径分隔符、Unix 信号或 PowerShell 专属语法。
- 确实只能在特定平台运行的操作必须写明平台边界，提供等价跨平台入口，或在 CI/文档中显式标注为平台专属运维步骤。
- 平台专属操作不得作为默认开发和测试入口。

## Go 工具链优先要求

- 长期维护的开发工具和脚本应优先实现为 Go 工具，放在`hack/tools/<tool>/`。
- 工具应通过`go run ./hack/tools/<tool>`、`linactl`或薄包装入口调用。
- 文件复制、目录遍历、配置改写、进程启停、端口探测、HTTP smoke、压缩/解压、模板渲染和静态扫描等逻辑应使用 Go 标准库或项目已有 Go 组件实现。
- 避免用 Shell 管道拼接系统命令承载业务逻辑。
- 根`Makefile`和`make.cmd`只允许作为兼容包装层，业务逻辑必须收敛到跨平台工具中。

## linactl 命令文件命名要求

- `hack/tools/linactl/`下承载具体`make`或`linactl`命令实现的源码文件，必须按对应命令名称命名为`command_<command>.go`。
- `<command>`保持命令的点分段语义。例如`make dev`对应`command_dev.go`，`make build`对应`command_build.go`，`make env.setup`对应`command_env.setup.go`。
- 若命令名与 Go 工具链文件后缀规则冲突，必须使用明确的命令专属后缀并记录原因，例如`test`命令使用`command_testcmd.go`，`wasm`命令使用`command_wasmcmd.go`。

## linactl 子组件组织要求

- `hack/tools/linactl/`根目录应尽可能只保留`command_*.go`指令入口、`command.go`注册与参数解析、`app.go`/`main.go`启动装配、基础类型和必要的平台适配文件。
- 开发服务、插件工作区、GoFrame CLI、前端依赖、Playwright、镜像构建、仓库治理扫描、文件系统工具等跨命令或较复杂实现逻辑必须迁移到`hack/tools/linactl/internal/<组件名称>/`子组件中。
- 根目录命令文件必须通过明确的包接口引用内部组件。
- 禁止在根目录继续新增`*_ops.go`、`*_management.go`、`*_workspace.go`、`*_util.go`这类承载复杂共享实现的兜底文件。
- 确需保留根目录非命令文件时，必须说明其属于启动装配、命令注册、基础类型或平台边界。

## 脚本目录治理要求

- `hack/scripts/`不再作为长期维护开发工具目录。
- 已有能力应迁移到`hack/tools/linactl`或独立 Go 工具。
- 测试辅助入口若必须保留在`hack/tests/scripts/`，应优先使用 Node 或 Go 编写。
- 新增或修改`.sh`、`.ps1`等平台脚本必须在变更记录和审查结论中说明无法使用 Go 工具链的原因、受支持平台、等价入口和验证方式。

## 验证要求

- 涉及开发工具或脚本的变更必须运行对应 Go 工具测试或跨平台 smoke。
- 若本次变更确认不影响开发工具或脚本，应在任务记录或审查结论中明确说明。

