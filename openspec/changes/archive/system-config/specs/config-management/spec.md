## ADDED Requirements

### Requirement: Built-in Runtime Parameter Metadata
The system SHALL provide built-in metadata records for host-consumed runtime parameters so administrators can manage effective host behavior directly from config management.

#### Scenario: Initialize built-in runtime parameters
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains `sys.jwt.expire`, `sys.session.timeout`, `sys.upload.maxSize`, and `sys.login.blackIPList`
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Built-in Runtime Parameter Protection
The system SHALL validate built-in runtime parameter values and SHALL protect stable host-owned keys from rename or deletion.

#### Scenario: Reject invalid built-in runtime parameter values
- **WHEN** a user creates, updates, or imports one of the built-in runtime parameters with an invalid value format
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of built-in runtime parameter keys
- **WHEN** a user attempts to rename or delete a built-in runtime parameter key already consumed by the host
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Upload Size Parameter Must Drive Host Runtime Behavior
The system SHALL ensure that `sys.upload.maxSize` is enforced by the host upload chain instead of existing only as editable metadata.

#### Scenario: Upload size change takes effect immediately
- **WHEN** an administrator updates `sys.upload.maxSize` to `1`
- **THEN** subsequent upload requests are validated against a 1 MB limit
- **AND** uploads above the configured limit are rejected

### Requirement: The default upload size must be unified at 20 MB
The system SHALL set the platform default value of `sys.upload.maxSize` to `20`, and database initialization, config-template defaults, and runtime upload fallbacks SHALL all use that same value unless an administrator explicitly overrides it.

#### Scenario: Host initialization writes the 20 MB default
- **WHEN** an administrator runs the host initialization SQL
- **THEN** the default value of `sys.upload.maxSize` in `sys_config` is `20`
- **AND** the default value read by config management for that built-in parameter is also `20`

#### Scenario: Runtime default remains 20 MB when no override is provided
- **WHEN** the host handles a `multipart` upload request without any administrator override for the upload-size setting
- **THEN** file-upload validation enforces a 20 MB limit
- **AND** the friendly error message triggered by the default limit returns wording equivalent to "file size cannot exceed 20 MB"

### Requirement: All default upload-size sources must stay consistent
The system SHALL keep the database seed value, config-template default, and host static fallback value for `sys.upload.maxSize` consistent so different startup paths do not expose different default upload limits.

#### Scenario: The host starts from the default template
- **WHEN** an operator generates runtime config from the host default `config.template.yaml` and does not change the upload limit separately
- **THEN** the host reads a default upload size of 20 MB
- **AND** that default matches the `sys.upload.maxSize` default written by the host initialization SQL

### Requirement: Multi-Instance Runtime Parameter Cache Synchronization
The system SHALL use a local snapshot plus shared revision strategy for protected runtime parameter reads so hot paths do not query `sys_config` on every request.

#### Scenario: Runtime reads hit the local snapshot
- **WHEN** a node repeatedly reads protected runtime parameters while the shared revision has not changed
- **THEN** the node reuses its local in-memory snapshot
- **AND** it does not need to query `sys_config` on every read

#### Scenario: Parameter changes propagate to other instances
- **WHEN** a protected runtime parameter changes on one instance
- **THEN** the writing instance clears its local snapshot and bumps the shared revision
- **AND** other instances rebuild their local snapshots during the next synchronization cycle

### Requirement: Public Frontend Setting Metadata
The system SHALL provide built-in metadata for safe public frontend settings used by branding, login-page presentation, and workspace theme bootstrap.

#### Scenario: Initialize public frontend settings
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains the built-in public frontend setting keys used by the login page and workspace bootstrap
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Public Frontend Setting Protection
The system SHALL validate built-in public frontend setting values and SHALL protect their stable keys from rename or deletion.

#### Scenario: Reject invalid public frontend setting values
- **WHEN** a user creates, updates, or imports a built-in public frontend setting with an invalid enum, boolean, or required-text value
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of public frontend setting keys
- **WHEN** a user attempts to rename or delete a built-in public frontend setting key already consumed by the login page or admin workspace
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Login and Workspace Consume Public Frontend Settings
The system SHALL expose a safe whitelist endpoint for public frontend settings and SHALL let the login page and admin workspace consume that contract.

#### Scenario: Public frontend settings are available before login
- **WHEN** a browser loads the login page without an authenticated session
- **THEN** the frontend can read the whitelisted branding and presentation settings through the public endpoint
- **AND** the endpoint does not expose arbitrary `sys_config` keys

#### Scenario: Updated branding is reflected after refresh
- **WHEN** an administrator updates public frontend settings and a user refreshes the login page or workspace
- **THEN** the refreshed UI shows the updated branding, copy, and theme defaults

### Requirement: The config-management component must have a unit-test coverage gate
The system SHALL maintain repeatable unit tests for the `apps/lina-core/internal/service/config` config-management component, and SHALL use package-level coverage verification as a delivery gate before that component is considered ready.

#### Scenario: Package-level coverage meets the delivery bar
- **WHEN** a maintainer runs `go test ./internal/service/config -cover` from `apps/lina-core`
- **THEN** the command succeeds
- **AND** the reported package-level statement coverage is not lower than `80%`

### Requirement: Critical config-management branches must have automated regression protection
The system SHALL add automated unit tests for critical helper logic inside the config-management component, including high-risk branches around defaults and fallbacks, cache or snapshot reuse, and invalid input or error propagation.

#### Scenario: Plugin and public-frontend config helper logic changes
- **WHEN** a change touches plugin dynamic storage paths, protected public-frontend config key checks, or the shared validation entry point
- **THEN** unit tests cover the normal read path
- **AND** cover default-value or compatibility-fallback behavior
- **AND** cover invalid input or empty-value defensive behavior

#### Scenario: Runtime-parameter cache and revision synchronization logic changes
- **WHEN** a change touches runtime-parameter snapshot caches, the revision controller, or shared-KV synchronization logic
- **THEN** unit tests cover cache-hit or local-reuse behavior
- **AND** cover rebuilds after revision changes
- **AND** cover error propagation and defensive behavior for shared-KV read failures, invalid cached values, or equivalent exceptional cases

### Requirement: 内置系统参数名称和描述必须按请求语言投影

配置管理页面 SHALL 按当前语言本地化内置系统参数名称与描述，使英文环境不显示默认中文系统文案。投影键 MUST 使用 `config.<config_key>.name` 与 `config.<config_key>.remark`，其中 `<config_key>` 为 `sys_config.key` 原值。

列表、按 key 查询与按 ID 的编辑详情 MUST 对 `name`/`remark` 返回当前请求语言投影。编辑详情的 `value` MUST 保持库内实际存储值，不得用公共前端默认文案投影覆盖可编辑值。前端编辑弹窗对内置参数 MUST 将 `name`/`remark` 设为只读。

#### Scenario: 登录和 IP 黑名单参数显示英文元数据
- **WHEN** 管理员以 `en-US` 打开系统配置
- **THEN** 内置登录、页面标题、页面描述、副标题和 IP 黑名单参数元数据以英文显示
- **AND** 页面不显示这些参数的中文内置标签

#### Scenario: 内置公共前端文案可投射英文显示内容
- **WHEN** 配置列表以 `en-US` 显示默认登录页标题、描述或副标题
- **THEN** 可见显示内容使用英文投射或英文默认值
- **AND** 编辑详情仍保留稳定的 `configKey` 和实际存储的 `value`

#### Scenario: 英文环境编辑内置参数时元数据为英文
- **WHEN** 管理员以 `en-US` 打开某内置参数的编辑详情
- **THEN** 详情中的 `name` 与 `remark` 为英文投影
- **AND** `value` 等于库内存储值（含管理员自定义后的原文）
- **AND** 编辑表单不展示该参数的中文 seed 名称或描述
- **AND** 内置参数的 `name`/`remark` 输入为只读

#### Scenario: 配置本地化资源保持完整
- **WHEN** 内置配置翻译键被添加或更改
- **THEN** 宿主全部运行时 locale 的 `config.<config_key>.name` 与 `config.<config_key>.remark` 保持覆盖
- **AND** `make i18n.check` 对缺失的内置配置展示键报告失败

### Requirement: 内置系统参数更新不得写回本地化名称与描述

系统 SHALL 在更新内置（`isBuiltin` 或受管系统键）配置记录时忽略请求中的 `name` 与 `remark`，仅允许在既有规则下更新可编辑字段（至少包含 `value`）。非内置自定义参数仍可更新 `name` 与 `remark`。mutation 路径 MUST 使用未投影实体加载，避免投影值进入比较或写回。

#### Scenario: 内置参数保存不污染 name/remark
- **WHEN** 调用方以 `en-US` 获取内置参数详情并将投影后的英文 `name`/`remark` 连同新 `value` 提交更新
- **THEN** 系统更新 `value`
- **AND** 库内 `name` 与 `remark` 仍为更新前的存储原文
- **AND** 后续中文环境下列表/详情仍可从 i18n 或库内 fallback 得到正确展示

#### Scenario: 自定义参数仍可修改名称与备注
- **WHEN** 管理员更新非内置参数的 `name` 或 `remark`
- **THEN** 系统按请求写入对应字段

### Requirement: 参数管理 API 必须暴露 valueType 与 options

参数设置的列表、详情、按键查询、创建与更新契约 SHALL 包含 `valueType` 与 `options` 字段。创建与更新请求可提交上述字段；响应 MUST 返回持久化后的有效值。`options` 在 JSON 中为对象数组；空选项以空数组表示。

#### Scenario: 创建带类型的参数
- **WHEN** 管理员调用创建接口提交 `name`、`key`、`value`、`valueType=select` 与非空 `options`
- **THEN** 系统创建成功
- **AND** 详情响应包含相同的 `valueType` 与 `options`

#### Scenario: 列表项携带类型元数据
- **WHEN** 管理员查询参数列表
- **THEN** 每条 `ConfigItem` 包含 `valueType`
- **AND** 对枚举类型包含可用于编辑的 `options`

### Requirement: 配置导入导出必须包含类型与选项列

配置 Excel 导出、导入与导入模板 SHALL 在既有列基础上包含 `valueType` 与 `options` 列，列头通过 `config.field.valueType` 与 `config.field.options` 翻译键按当前语言解析。导入时缺失 `valueType` 视为 `text`；`options` 单元格存 JSON 文本。

#### Scenario: 导出包含类型列
- **WHEN** 管理员导出参数设置
- **THEN** Excel 包含本地化的 valueType 与 options 列头
- **AND** 行数据写出对应字段

#### Scenario: 导入非法 options JSON 行失败
- **WHEN** 管理员导入一行 `valueType=select` 且 options 单元格不是合法 JSON 数组
- **THEN** 该行计入失败列表并给出可理解原因
- **AND** 其他合法行仍可成功导入

### Requirement: 内置参数种子必须携带类型与选项元数据

宿主初始化 SQL SHALL 为已有内置运行时与公共前端参数写入匹配语义的 `value_type` 与（如需要）`options`，使管理面开箱按正确组件编辑。至少覆盖：

- 布尔开关类（如忘记密码入口、注册入口、水印开关）→ `boolean`
- 布局/主题等有限枚举 → `select` 或 `radio`，并给出完整 options
- 上传上限、日志保留天数等 → `number`
- 隐私政策/服务条款等长文 → `textarea` 或 `richtext`
- 其余短文案/路径/duration → `text` 或 `textarea`

#### Scenario: 登录框位置参数为下拉类型
- **WHEN** 管理员完成宿主数据库初始化后打开参数设置并编辑 `sys.auth.loginPanelLayout`
- **THEN** 该参数 `valueType` 为 `select`（或 `radio`）
- **AND** options 包含 `panel-left`、`panel-center`、`panel-right`
- **AND** 编辑界面以下拉或单选方式选择，而非自由文本猜测

#### Scenario: 忘记密码入口参数为布尔类型
- **WHEN** 管理员编辑 `sys.auth.forgetPasswordEnabled`
- **THEN** 该参数 `valueType` 为 `boolean`
- **AND** 编辑界面提供开关或二值选择组件

### Requirement: 系统参数管理面仅展示可系统维护的配置

系统 SHALL 为每条 `sys_config` 记录持久化 `system_manageable` 字段（`SMALLINT`，`1` 表示允许在系统参数设置管理面维护，`0` 表示不允许）。系统参数管理面的 List 与 Export MUST 仅返回 `system_manageable = 1` 的可见行。运行时配置读取 MUST 不受该字段影响。宿主 seed 与管理面 Create 默认 `1`。

#### Scenario: 列表不返回插件闭环配置
- **WHEN** 管理员打开系统参数设置列表
- **AND** 存在 `system_manageable = 0` 的行
- **THEN** 列表不包含这些行

### Requirement: 系统参数管理面不得变更不可系统维护的配置

对 `system_manageable = 0` 的行，管理面 Get/Update/Delete/Import 覆盖 MUST 拒绝或视为不存在。管理面 Create MUST 写入 `system_manageable = 1`。

#### Scenario: 管理面更新被拒绝
- **WHEN** 调用方对 `system_manageable = 0` 的配置请求管理面更新
- **THEN** 系统返回错误且 value 不变

### Requirement: 插件 SetValue 支持显式 SystemManageable

插件经 `HostConfig.SysConfig().SetValue(ctx, key, value, options)` 或 `BatchSetValue(ctx, items, options)` 写入时，`options` 可为 nil 或 `*SetSysConfigValueOptions`。当 `options` 为 nil 或 `options.SystemManageable` 为 nil 且首次插入时 MUST 写 `0`；更新时 MUST 保持原标记；当 `SystemManageable` 非 nil 时 MUST 写入对应标记。仅在插件入口维护的业务配置 MUST 传 `false`。

#### Scenario: 插件闭环写入
- **WHEN** 插件 `SetValue`/`BatchSetValue` 且 `options.SystemManageable = false` 或未指定（首次插入）
- **THEN** 行的 `system_manageable = 0` 且不出现在系统参数列表

#### Scenario: 插件显式进入系统参数页
- **WHEN** 插件 `SetValue` 且 `options.SystemManageable = true`
- **THEN** 行的 `system_manageable = 1`

### Requirement: 插件批量设置必须单事务单 revision

系统 SHALL 提供 `BatchSetValue`，在一次事务中写入全部 items，并在全部成功后仅推进一次 runtime-config revision。空 items MUST 成功且无副作用。多字段插件 settings 保存 MUST 使用 `BatchSetValue` 而非循环 `SetValue`。

#### Scenario: 批量写入多键
- **WHEN** 插件一次 `BatchSetValue` 写入多个 key
- **THEN** 所有 key 在同一事务中落库
- **AND** runtime-config revision 仅推进一次

### Requirement: 内置参数 sys.auth.sloganImage

系统 SHALL 提供内置参数 `sys.auth.sloganImage`，用于配置登录页 slogan 插画图片地址。默认值为 `/slogan.svg`（Vben 内置插画）。空值表示不使用插画。该键使用允许空串的读取路径：库内已存在的空值不得回退到默认值，以支持“清空=隐藏插画”。参数纳入公共前端白名单并投影为 `auth.sloganImage`。

#### Scenario: 参数设置页可见 slogan 参数
- **当** 管理员打开参数设置并搜索 `sys.auth.sloganImage` 时
- **则** 列表显示该内置参数
- **且** 参数名称标识为登录展示相关的 slogan 插画配置

#### Scenario: 允许清空 slogan 地址以隐藏插画
- **当** 管理员将 `sys.auth.sloganImage` 保存为空值时
- **则** 系统接受该值
- **且** 公共前端配置返回空串而非默认 `/slogan.svg`
