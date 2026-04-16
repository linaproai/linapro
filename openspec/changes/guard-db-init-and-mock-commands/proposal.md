## Why

当前`lina-core`的`init`和`mock`命令会直接扫描并执行数据库初始化或演示数据`SQL`，而且根目录`make init`、`make mock`也会无条件透传到后端命令。这类命令一旦被误触发，可能在错误环境中写入、覆盖或污染数据库，属于高风险操作。

随着仓库中开发、调试和自动化脚本入口逐渐增多，仅依赖使用者记忆“这些命令要小心”并不可靠。需要给这两个敏感命令补上显式确认护栏，把误操作从“容易发生”降到“必须主动确认才能执行”。

## What Changes

- 为`lina-core`的`init`和`mock`命令增加显式确认参数校验；未提供正确确认值时，命令必须拒绝执行任何`SQL`。
- 为根目录和`apps/lina-core`目录下的`Makefile`入口增加同等确认要求，避免通过`make init`、`make mock`误触发高风险数据库操作。
- 为敏感命令补充清晰的失败提示与正确示例，明确告诉操作者需要使用的确认参数和值。
- 调整命令执行失败语义：当敏感命令进入执行阶段后，如任一`SQL`文件执行失败，命令应返回失败状态，而不是仅记录告警后继续。
- 为命令护栏补充对应的自动化测试，确保命令默认安全、确认后可执行、失败时可观测。

## Capabilities

### New Capabilities
- `database-bootstrap-commands`: 规范宿主数据库初始化与演示数据加载命令的入口、安全确认、执行失败语义和可观测反馈。

### Modified Capabilities

## Impact

- 受影响后端命令实现：`apps/lina-core/internal/cmd/`下的`init`、`mock`相关命令和公共执行逻辑。
- 受影响脚本入口：仓库根`Makefile`与`apps/lina-core/Makefile`中的`init`、`mock`目标。
- 受影响开发者体验：命令使用方式会从直接执行改为显式传入确认参数，例如`make init confirm=init`、`make mock confirm=mock`。
- 受影响测试：需要补充命令级单元测试，覆盖未确认拒绝执行、确认后允许执行、`SQL`执行失败返回非零状态等场景。
