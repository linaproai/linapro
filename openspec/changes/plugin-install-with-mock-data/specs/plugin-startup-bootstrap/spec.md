## MODIFIED Requirements

### Requirement: The host must provide a simplified plugin auto-enable config in the main config file

宿主 SHALL 在主配置文件 `apps/lina-core/manifest/config/config.yaml` 中提供 `plugin.autoEnable` 列表，用于声明启动期自动安装并启用的插件。该列表 MUST 接受两种条目写法的混用：

- **字符串写法**：直接给出插件 ID，例如 `- "plugin-demo-source"`。该写法语义等价于 `withMockData=false`，启动期 MUST NOT 加载该插件的示例数据。
- **对象写法**：`{id: <pluginID>, withMockData: <bool>}`。`id` 为必填字段，`withMockData` 为可选布尔，缺省值为 `false`；仅当 `withMockData=true` 时启动期 SHALL 加载该插件的 mock 数据并复用与手动安装相同的事务化执行流程。

宿主 SHALL 在配置加载阶段将所有条目标准化为内部的 `(id, withMockData)` 二元组结构供 `BootstrapAutoEnable` 使用，并 MUST 校验：插件 ID 为非空字符串、`withMockData` 为合法布尔值；任一非法值出现时启动 SHALL panic（与"启动期不可恢复错误"语义一致），错误信息 MUST 指明问题条目的位置或值。

#### Scenario: Parse a valid auto-enable list with mixed entry types
- **WHEN** 配置文件 `plugin.autoEnable` 列表包含字符串 `"demo-control"`、对象 `{id: "plugin-demo-source", withMockData: true}`、对象 `{id: "plugin-demo-dynamic"}` 三种写法
- **THEN** 宿主成功解析为 `[("demo-control", false), ("plugin-demo-source", true), ("plugin-demo-dynamic", false)]`
- **AND** 启动期对前两个不加载 mock 数据，对第二个加载 mock 数据
- **AND** 第三个对象因 `withMockData` 缺省而等价于字符串写法

#### Scenario: Reject invalid auto-enable config
- **WHEN** 配置中出现 `{id: ""}`、`{withMockData: true}`（缺 id）、`{id: "x", withMockData: "yes"}`（错误类型）等非法条目
- **THEN** 宿主在配置加载或启动期 panic
- **AND** 错误信息明确指出非法条目的具体位置或键

### Requirement: The auto-enable list must implicitly include install and enable semantics

宿主在 `BootstrapAutoEnable(ctx)` 中 SHALL 对 `plugin.autoEnable` 中每个条目执行"安装（若未安装）+ 启用"的隐式语义。当条目的 `withMockData=true` 时，安装步骤 MUST 复用手动安装路径中的 mock SQL 事务化执行流程；当 `withMockData=false`（包括字符串写法与缺省对象写法）时，启动期 MUST NOT 触发 mock-data 目录的任何扫描或执行。已安装的插件再次出现在该列表中时，启动期 MUST NOT 重复加载 mock 数据（即便 `withMockData=true`）——`withMockData` 仅在第一次安装时生效。

#### Scenario: Auto-enable a newly discovered source plugin without mock data
- **WHEN** `plugin.autoEnable` 包含字符串 `"plugin-demo-source"` 且该插件尚未安装
- **AND** 宿主启动并执行 `BootstrapAutoEnable`
- **THEN** 宿主完成该插件的 install SQL 执行、注册、菜单同步、启用等步骤
- **AND** 完全不扫描该插件的 `manifest/sql/mock-data/` 目录
- **AND** 数据库中不存在任何来自该插件 mock 数据的记录

#### Scenario: Auto-enable a newly discovered plugin with mock data opt-in
- **WHEN** `plugin.autoEnable` 包含 `{id: "plugin-demo-source", withMockData: true}` 且该插件尚未安装
- **AND** 宿主启动并执行 `BootstrapAutoEnable`
- **THEN** 宿主在 install SQL 执行成功后，事务化执行该插件 `manifest/sql/mock-data/*.sql` 全部文件
- **AND** mock 阶段成功后插件被启用，启动继续

#### Scenario: An already installed plugin reappears with withMockData=true
- **WHEN** `plugin.autoEnable` 包含 `{id: "x", withMockData: true}` 且该插件已经安装过（无论上次是否带 mock）
- **AND** 宿主启动并执行 `BootstrapAutoEnable`
- **THEN** 宿主仅确保该插件处于启用状态
- **AND** 不重新执行 install SQL，也不执行 mock-data SQL

### Requirement: Any failure for a listed auto-enable plugin must block host startup

启动期 `BootstrapAutoEnable` 的任一阶段失败 MUST 阻塞宿主启动并 panic（与现有"启动期任意失败 panic"语义一致）。`withMockData=true` 时的 mock 阶段失败 SHALL 同样阻塞启动；mock 事务回滚后宿主 MUST panic 并在错误信息中包含插件 ID、失败的 SQL 文件名以及"已自动回滚 mock 数据"的提示，以便运维介入修复后再重启。

#### Scenario: A missing auto-enable plugin causes startup failure
- **WHEN** `plugin.autoEnable` 列出的某个插件 ID 在 catalog 中不存在
- **THEN** 启动失败 panic，错误信息包含该插件 ID

#### Scenario: An auto-enable plugin install failure causes startup failure
- **WHEN** 某个 auto-enable 插件的 install SQL 执行失败
- **THEN** 启动失败 panic，错误信息包含失败原因

#### Scenario: Mock SQL failure during auto-enable causes startup failure
- **WHEN** `plugin.autoEnable` 包含 `{id: "x", withMockData: true}`
- **AND** 该插件 install SQL 全部执行成功
- **AND** mock-data 目录下任一 SQL 文件执行失败
- **THEN** 宿主回滚 mock 事务后 panic，启动失败
- **AND** panic 错误信息包含插件 ID、失败的 mock SQL 文件名与失败原因
