## ADDED Requirements

### Requirement: 内置运行时参数元数据
系统 SHALL 在初始化数据中提供已接入宿主运行时行为的内置参数元数据，便于管理员直接在参数设置页查看和维护。

#### Scenario: 初始化内置运行时参数
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 中包含 `sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList` 四项内置运行时参数
- **AND** 每条记录包含可读名称、默认值与格式说明备注

### Requirement: 内置运行时参数保护规则
系统 SHALL 对已接入宿主运行时行为的内置参数执行格式校验，并保护稳定键名不被误删或误改。

#### Scenario: 拒绝非法的内置参数值
- **WHEN** user creates, updates, or imports `sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize` 或 `sys.login.blackIPList` with an invalid value format
- **THEN** system rejects the change and returns a validation error

#### Scenario: 拒绝修改或删除内置参数键名
- **WHEN** user attempts to rename or delete one built-in runtime parameter key already consumed by host runtime
- **THEN** system rejects the operation and keeps the parameter record intact

### Requirement: 已纳管的上传大小参数必须驱动宿主行为
系统 SHALL 确保 `sys.upload.maxSize` 不只是参数设置页中的展示元数据，而是宿主文件上传链路实际使用的运行时上限。

#### Scenario: 上传大小上限参数即时生效
- **WHEN** 管理员在参数设置中将 `sys.upload.maxSize` 更新为 `1`
- **THEN** 后续文件上传请求 MUST 按 1MB 上限进行校验
- **AND** 超过上限的上传请求被拒绝

### Requirement: 多实例部署下的运行时参数缓存同步
系统 SHALL 在多实例部署下使用“本地快照 + 共享修订号”的缓存同步策略，避免热点链路每次读取运行时参数都直查 `sys_config`，并保证参数变更可以传播到其他实例。

#### Scenario: 参数读取命中本地快照
- **WHEN** 任一实例在共享 revision 未变化期间多次读取 `sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize` 或 `sys.login.blackIPList`
- **THEN** 系统复用当前实例内存中的运行时参数快照
- **AND** 不要求每次读取都访问 `sys_config`

#### Scenario: 受保护运行时参数变更后传播到其他实例
- **WHEN** 管理员在任一实例成功创建、更新或导入一个受保护运行时参数并改变其运行时值
- **THEN** 当前实例 MUST 立即清空本地快照并递增共享 revision
- **AND** 其他实例 MUST 在下一次 revision 同步周期内重建本地快照
- **AND** 重建后的运行时行为 MUST 使用最新参数值
