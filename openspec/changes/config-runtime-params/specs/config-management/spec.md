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

### Requirement: 品牌与登录展示参数元数据
系统 SHALL 在初始化数据中提供品牌、登录页展示和后台样式相关的内置前端参数元数据，便于管理员在参数设置页统一维护宿主默认展示效果。

#### Scenario: 初始化公开前端展示参数
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 中包含 `sys.app.name`、`sys.app.logo`、`sys.app.logoDark`、`sys.auth.pageTitle`、`sys.auth.pageDesc`、`sys.auth.loginSubtitle`、`sys.ui.theme.mode`、`sys.ui.layout`、`sys.ui.watermark.enabled`、`sys.ui.watermark.content`
- **AND** 每条记录包含可读名称、默认值与格式说明备注

### Requirement: 公开前端展示参数保护规则
系统 SHALL 对已接入登录页和管理后台展示能力的公开前端参数执行格式校验，并保护稳定键名不被误删或误改。

#### Scenario: 拒绝非法的公开前端参数值
- **WHEN** user creates, updates, or imports one of the built-in public frontend settings with an invalid enum, boolean, or required-text value
- **THEN** system rejects the change and returns a validation error

#### Scenario: 拒绝修改或删除公开前端参数键名
- **WHEN** user attempts to rename or delete one built-in public frontend setting key already consumed by login page or admin workspace
- **THEN** system rejects the operation and keeps the parameter record intact

### Requirement: 登录页与工作台消费公开前端展示参数
系统 SHALL 通过公开白名单接口向前端提供安全的品牌与展示配置，并让登录页和管理后台在应用启动后消费这些系统参数。

#### Scenario: 未登录页面读取公开前端配置
- **WHEN** 浏览器在未登录状态访问登录页
- **THEN** 前端可以通过公开接口读取应用名称、Logo、登录页标题、说明、副标题以及安全的主题/布局参数
- **AND** 接口不返回任意 `sys_config` 键值对，只返回白名单字段

#### Scenario: 登录页展示管理员配置的品牌信息
- **WHEN** 管理员将 `sys.app.name`、`sys.auth.pageTitle`、`sys.auth.pageDesc` 或 `sys.auth.loginSubtitle` 更新为新值并刷新登录页
- **THEN** 登录页和浏览器标题 MUST 展示更新后的品牌或说明信息

#### Scenario: 后台工作台应用系统参数主题偏好
- **WHEN** 管理员将 `sys.ui.theme.mode`、`sys.ui.layout` 或水印参数更新为新值并刷新应用
- **THEN** 管理后台 MUST 使用系统参数指定的主题模式、布局和水印偏好作为启动时的默认展示配置
