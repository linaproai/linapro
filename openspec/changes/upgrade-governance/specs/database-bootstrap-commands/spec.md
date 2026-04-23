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

### Requirement: 数据库启动命令必须按阶段选择 SQL 资产来源
系统 SHALL 让宿主 `init` 和 `mock` 命令支持明确的 SQL 资产来源切换：运行时 `lina init` / `lina mock` 默认读取 embedded FS 中打包的宿主 SQL 资产；开发阶段 `make init` / `make mock` 必须显式切换为源码树中的本地 SQL 文件。实现 MUST 不依赖对当前工作目录的隐式猜测来决定来源。

#### Scenario: 运行时 `init` 默认读取 embedded SQL
- **WHEN** 操作者通过发布后的 `lina init --confirm=init` 执行宿主初始化
- **THEN** 命令必须读取打包进二进制 embedded FS 的 `manifest/sql/` 资产
- **AND** 不得直接要求本地源码树中的 `manifest/sql/` 存在

#### Scenario: 开发阶段 `make mock` 显式读取本地 SQL
- **WHEN** 开发者执行 `make mock confirm=mock`
- **THEN** `Makefile` 必须显式将 `mock` 命令切换到本地 SQL 文件源
- **AND** 命令必须读取源码树中的 `manifest/sql/mock-data/` 目录
