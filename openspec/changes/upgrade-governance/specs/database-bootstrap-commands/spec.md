## MODIFIED Requirements

### Requirement: 敏感数据库启动命令必须显式确认
系统 SHALL 要求宿主 `init` 和 `mock` 命令在执行任何 `SQL` 之前接收与命令名一一对应的显式确认值；未提供或提供错误确认值时，命令 MUST 拒绝执行。`init` 与 `mock` 只允许作用于宿主 bootstrap 初始化资产，不承担正式升级语义，也不得替代 `upgrade` 命令。

#### Scenario: `init` 命令缺少确认值
- **WHEN** 操作者执行 `go run main.go init`，但未传入 `--confirm=init`
- **THEN** 命令拒绝执行数据库初始化 `SQL`
- **AND** 命令输出明确的失败原因和正确示例

#### Scenario: `mock` 命令收到错误确认值
- **WHEN** 操作者执行 `go run main.go mock --confirm=init`
- **THEN** 命令拒绝执行任何 `mock-data` 目录下的 `SQL`
- **AND** 命令要求使用与 `mock` 命令匹配的确认值

#### Scenario: `init` 不建立框架升级账本
- **WHEN** 操作者执行 `go run main.go init --confirm=init` 且全部宿主 SQL 成功执行
- **THEN** 命令只完成宿主 bootstrap 初始化
- **AND** 不得写入任何框架升级状态、升级记录或 SQL 游标元数据
