# database-bootstrap-commands Specification

## Purpose
TBD - created by archiving change framework-positioning-core-boundary-and-readme-i18n. Update Purpose after archive.
## Requirements
### Requirement: 敏感数据库启动命令必须显式确认
系统 SHALL 要求宿主`init`和`mock`命令在执行任何`SQL`之前接收与命令名一一对应的显式确认值；未提供或提供错误确认值时，命令 MUST 拒绝执行。

#### Scenario: `init`命令缺少确认值
- **WHEN** 操作者执行`go run main.go init`，但未传入`--confirm=init`
- **THEN** 命令拒绝执行数据库初始化`SQL`
- **AND** 命令输出明确的失败原因和正确示例

#### Scenario: `mock`命令收到错误确认值
- **WHEN** 操作者执行`go run main.go mock --confirm=init`
- **THEN** 命令拒绝执行任何`mock-data`目录下的`SQL`
- **AND** 命令要求使用与`mock`命令匹配的确认值

#### Scenario: 命令收到正确确认值
- **WHEN** 操作者执行`go run main.go init --confirm=init`或`go run main.go mock --confirm=mock`
- **THEN** 命令允许进入对应的`SQL`扫描与执行流程

### Requirement: `Makefile`入口必须复用同等确认语义
系统 SHALL 要求仓库根目录和`apps/lina-core`目录下的`make init`、`make mock`入口使用与命令实现一致的`confirm`确认值，并在缺失或错误时提前失败。

#### Scenario: 根目录`make init`缺少确认变量
- **WHEN** 操作者在仓库根目录执行`make init`
- **THEN** `Makefile`拒绝继续调用后端初始化命令
- **AND** 输出正确示例`make init confirm=init`

#### Scenario: 后端`make mock`使用正确确认变量
- **WHEN** 操作者在`apps/lina-core`目录执行`make mock confirm=mock`
- **THEN** `Makefile`将确认值透传给后端命令实现
- **AND** 后端命令继续按`mock`语义校验并执行

### Requirement: 数据库启动`SQL`执行失败必须立即返回失败状态
系统 SHALL 在`init`或`mock`进入执行阶段后，于任一`SQL`文件执行失败时立即停止后续文件执行，并向调用方返回失败状态。

#### Scenario: 某个`SQL`文件执行失败
- **WHEN** `init`或`mock`执行过程中某个`SQL`文件返回执行错误
- **THEN** 系统立即停止执行后续`SQL`文件
- **AND** 命令返回失败状态给`make`或直接调用方
- **AND** 日志包含失败文件名和错误信息

#### Scenario: 所有`SQL`文件执行成功
- **WHEN** `init`或`mock`执行阶段内的所有目标`SQL`文件都成功执行
- **THEN** 命令返回成功状态
- **AND** 日志输出对应的完成提示

