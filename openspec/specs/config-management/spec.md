# 配置管理

## Purpose

定义配置管理行为，包括本地化导入/导出元数据、内置参数展示和系统拥有记录的删除保护。
## Requirements
### Requirement:配置导出和导入表头必须通过翻译键按当前语言解析

系统 SHALL 在配置 Excel 导出和导入流程中通过 `config.field.<name>` 翻译键按当前请求语言解析列头（`name`、`key`、`value`、`remark`、`createdAt`、`updatedAt`）。后端 Go 源码不得维护字面的英文/中文表头映射。新增语言只需在 `apps/lina-core/manifest/i18n/<locale>/*.json` 下添加对应的 `config.field.*` 资源。

#### Scenario:导出使用当前语言的表头
- **当** 管理员以非默认运行时语言导出配置时
- **则** Excel 列头使用该语言的 `config.field.*` 翻译
- **且** 后端源码不包含重复的字面表头映射

#### Scenario:新增语言无需后端代码变更
- **当** 项目启用新的内置语言并提供 `config.field.*` 资源时
- **则** 配置导入和导出表头以该语言显示
- **且** 配置服务无需源码变更

### Requirement:内置系统参数名称和默认文案必须以英文本地化

配置管理页面 SHALL 按当前语言本地化内置系统参数名称、描述和默认显示值，使英文环境不显示默认中文系统文案。投影键 MUST 使用 `config.<config_key>.name` 与 `config.<config_key>.remark`，其中 `<config_key>` 为 `sys_config.key` 原值。

#### Scenario:登录和 IP 黑名单参数显示英文元数据
- **当** 管理员以 `en-US` 打开系统配置时
- **则** 内置登录、页面标题、页面描述、副标题和 IP 黑名单参数元数据以英文显示
- **且** 页面不显示这些参数的中文内置标签

#### Scenario:内置公共前端文案可投射英文显示内容
- **当** 配置列表以 `en-US` 显示默认登录页标题、描述或副标题时
- **则** 可见显示内容使用英文投射或英文默认值
- **且** 编辑详情仍保留稳定的 `configKey` 和实际存储值

#### Scenario:配置本地化资源保持完整
- **当** 内置配置翻译键被添加或更改时
- **则** 宿主全部运行时 locale 的 `config.<config_key>.name` 与 `config.<config_key>.remark` 保持覆盖
- **且** `make i18n.check` 对缺失的内置配置展示键报告失败

### Requirement:内置系统参数必须可编辑但不可删除

系统拥有的配置记录 SHALL 标记为内置。管理员可编辑其可编辑字段和值，但前端和后端都必须阻断内置记录的删除。

#### Scenario:内置系统参数删除操作被禁用
- **当** 管理员查看内置配置行时
- **则** 删除操作被禁用，不打开删除确认
- **且** 悬停文本说明内置系统数据不可删除
- **且** 编辑操作保持可用

#### Scenario:后端拒绝内置系统参数删除
- **当** 调用方绕过前端请求删除内置配置记录时
- **则** 后端返回结构化业务错误并保留记录
- **且** 非内置配置记录在现有权限和验证规则下保持可删除

### Requirement:受保护的运行时参数缓存必须跨节点有界一致

系统 SHALL 通过统一缓存协调机制同步受保护的运行时参数缓存，使集群模式下没有节点无限期使用旧参数快照。

#### Scenario:集群模式下受保护的运行时参数变更

- **当** 管理员更改受保护的运行时参数时
- **则** 系统提交参数变更
- **且** 可靠发布运行时配置缓存修订号
- **且** 其他节点在运行时配置缓存域允许的陈旧窗口内刷新其本地参数快照

#### Scenario:运行时参数修订号发布失败

- **当** 参数变更需要运行时配置缓存刷新但修订号发布失败时
- **则** 系统返回结构化业务错误
- **且** 调用方不得收到静默成功结果
- **且** 系统记录可重试的失败原因

### Requirement:运行时参数读取必须执行新鲜度检查

在读取影响认证、会话、上传、调度或其他运行时行为的受保护参数前，系统 SHALL 验证本地快照未超过允许的陈旧窗口。

#### Scenario:本地参数快照已在最新修订号

- **当** 节点读取受保护的运行时参数且其本地修订号已消费共享修订号时
- **则** 系统从本地缓存快照返回参数
- **且** 不重新查询完整的 `sys_config` 参数集

#### Scenario:本地参数快照落后于共享修订号

- **当** 节点读取受保护的运行时参数并观察到更新的共享修订号时
- **则** 系统从 `sys_config` 重建本地参数快照
- **且** 后续读取使用新修订号的快照

#### Scenario:新鲜度无法确认且超过故障窗口

- **当** 节点无法读取共享修订号且其本地运行时参数快照超过故障窗口时
- **则** 系统返回可见错误或按该参数域声明的策略降级
- **且** 系统不得无限期静默使用旧参数快照

### Requirement: 运行时配置 revision 必须使用 Redis coordination
系统 SHALL 在集群模式下通过 Redis revision/event 协调受保护运行时参数变更。`sys_config` 仍为权威数据源，Redis 仅承载 revision 和失效事件。

#### Scenario: 修改 JWT 过期配置
- **WHEN** 管理员修改 `sys.jwt.expire`
- **THEN** 系统提交 `sys_config` 权威数据
- **AND** 发布 `runtime-config` Redis revision
- **AND** 其他节点刷新运行时参数快照

#### Scenario: 修改会话超时配置
- **WHEN** 管理员修改 `sys.session.timeout`
- **THEN** 系统发布 `runtime-config` Redis revision
- **AND** 新请求使用更新后的会话超时策略

### Requirement: 运行时配置 freshness 不可确认时必须返回可见错误
系统 SHALL 在读取受保护运行时参数前确认本地快照 freshness。当 Redis revision 不可读取且本地快照超过最大陈旧窗口时，系统 MUST 返回结构化错误，不得静默使用陈旧配置。

#### Scenario: Redis runtime-config revision 不可读
- **WHEN** 请求路径需要读取受保护运行时参数
- **AND** Redis revision 不可读
- **AND** 本地运行时参数快照超过最大陈旧窗口
- **THEN** 系统返回结构化配置 freshness 错误
- **AND** 记录可观测日志

### Requirement: 单机运行时配置保持本地 revision
系统 SHALL 在单机模式下使用进程内 revision 管理运行时参数快照失效，不得要求 Redis。

#### Scenario: 单机修改运行时配置
- **WHEN** `cluster.enabled=false`
- **AND** 管理员修改受保护运行时参数
- **THEN** 系统更新进程内 revision
- **AND** 当前进程清理本地运行时参数快照

### Requirement: 租户参数 fallback 行必须返回来源和动作元数据

租户上下文查询参数设置列表时，系统 SHALL 对平台默认 fallback 行返回来源和动作元数据，使调用方能区分“当前租户覆盖值”和“继承平台默认值”。元数据至少包含 `sourceTenantId`、`isFallback`、`canEdit`、`canOverride` 和 `overrideMode`。租户上下文不得把平台 fallback 行伪装成可直接编辑的本租户记录。

#### Scenario: 租户看到平台 fallback 参数

- **WHEN** 租户 A 未覆盖某个内置参数
- **AND** 参数列表通过平台默认值 fallback 返回该参数
- **THEN** 响应行包含 `sourceTenantId = 0`
- **AND** `isFallback = true`
- **AND** `canEdit = false`
- **AND** 删除和直接编辑操作不得在前端展示为可执行动作

#### Scenario: 租户看到本租户覆盖参数

- **WHEN** 租户 A 已覆盖某个参数
- **THEN** 响应行包含 `sourceTenantId = A`
- **AND** `isFallback = false`
- **AND** `canEdit` 根据当前用户权限和内置保护规则计算

#### Scenario: 平台上下文查看平台默认参数

- **WHEN** 平台管理员在平台上下文查询参数列表
- **THEN** 平台默认参数行不标记为租户 fallback
- **AND** `canEdit` 按平台参数管理权限和内置保护规则计算

### Requirement: 参数 fallback 动作必须避免必失败详情请求

前端 SHALL 使用参数行的动作元数据决定操作按钮。对 `isFallback = true` 且 `canEdit = false` 的行，前端不得显示会调用当前租户详情编辑接口且必然返回 not found 的编辑入口。若 `canOverride = true`，前端 MAY 显示创建租户覆盖入口，但该入口 MUST 调用明确的覆盖创建流程。

#### Scenario: fallback 参数行不显示直接编辑

- **WHEN** 租户用户在参数列表看到平台 fallback 行
- **AND** 该行 `canEdit = false`
- **THEN** 前端不显示直接编辑按钮
- **AND** 用户不会触发返回“参数设置不存在”的详情请求

#### Scenario: fallback 参数行允许创建覆盖

- **WHEN** 租户用户看到 `canOverride = true` 且 `overrideMode = createTenantOverride` 的 fallback 参数行
- **THEN** 前端显示的覆盖入口必须表达为创建租户覆盖
- **AND** 保存后创建或更新当前租户的参数覆盖记录
- **AND** 缓存失效范围限定为当前租户和该参数键

### Requirement: HostConfig 原始读取必须使用统一来源优先级

系统 SHALL 将`sys_config`作为宿主运行时系统配置的权威数据源。宿主配置服务 MUST 基于共享 revision 和本地快照缓存读取当前上下文可见的`sys_config`有效 key，而不是仅依赖 Go 代码中的硬编码 key 白名单。`sys_config`创建、更新、导入或删除导致有效配置变化后，系统 MUST 推进 runtime-config revision 并使后续读取重建快照。

#### Scenario: 读取自定义系统配置

- **WHEN** `sys_config`中存在 key 为`custom.feature.limit`的当前上下文可见记录
- **THEN** 宿主配置服务通过`GetRaw(ctx, "custom.feature.limit")`返回该记录的值
- **AND** 不需要为该 key 新增 Go 常量或修改硬编码白名单

#### Scenario: 静态配置 fallback

- **WHEN** 当前上下文可见的`sys_config`中不存在`workspace.basePath`
- **AND** 静态`config.yaml`中存在`workspace.basePath`
- **THEN** 宿主配置服务通过`GetRaw(ctx, "workspace.basePath")`返回静态配置值

#### Scenario: 租户覆盖优先于平台默认

- **WHEN** 当前上下文为租户`tenant-a`
- **AND** `sys_config`同时存在平台 key`custom.feature.limit=100`和租户 key`custom.feature.limit=50`
- **THEN** 宿主配置服务返回`50`

#### Scenario: 配置更新后刷新快照

- **WHEN** `sys_config`中`custom.feature.limit`从`100`更新为`200`
- **THEN** 系统推进 runtime-config revision
- **AND** 后续宿主配置读取返回`200`

#### Scenario: 配置删除后刷新快照

- **WHEN** `sys_config`中`custom.feature.limit`被删除
- **THEN** 系统推进 runtime-config revision
- **AND** 后续宿主配置读取不再返回被删除值

#### Scenario: 内置运行时参数使用系统命名空间

- **WHEN** 主框架新增或维护内置运行时参数
- **THEN** 参数 key MUST 使用`sys.`前缀
- **AND** 调度模块内置运行时参数 MUST 使用`sys.cron.shell.enabled`和`sys.cron.log.retention`

### Requirement: 系统默认值必须由通用元数据提供

系统 SHALL 将宿主已有硬编码默认值维护为可按 key 查询的通用默认值元数据或等价 resolver。`HostConfig`通用读取流程 MUST 只调用默认值查询入口，不得在读取流程中直接判断具体配置键。新增宿主默认值时，系统 MUST 更新默认值元数据和测试，而不是在`GetRaw()`中增加新的 key 分支。

#### Scenario: 新增默认值不修改通用读取分支

- **WHEN** 主框架为新的宿主配置 key 增加系统默认值
- **THEN** 开发者在默认值元数据或等价 resolver 中登记该 key 和默认值
- **AND** 不在`GetRaw()`读取流程中增加该 key 的专用判断

#### Scenario: 专用 getter 保留类型校验

- **WHEN** 专用 getter 读取具有系统默认值的配置键
- **AND** `sys_config`和静态`config.yaml`都没有提供该值
- **THEN** 专用 getter 使用系统默认值作为输入
- **AND** 继续执行该 getter 已有的类型解析、归一化和业务校验

#### Scenario: sys_config freshness 错误不被 fallback 掩盖

- **WHEN** 宿主读取非 root 配置键时无法确认`sys_config`快照 freshness
- **THEN** 系统返回可见错误
- **AND** 不继续回退到静态配置或系统默认值来掩盖运行时配置一致性故障

### Requirement:内置系统参数更新不得写回本地化名称与描述

系统 SHALL 在更新内置（`isBuiltin` 或受管系统键）配置记录时忽略请求中的 `name` 与 `remark`，仅允许在既有规则下更新可编辑字段（至少包含 `value`）。非内置自定义参数仍可更新 `name` 与 `remark`。

#### Scenario:内置参数保存不污染 name/remark
- **当** 调用方以 `en-US` 获取内置参数详情并将投影后的英文 `name`/`remark` 连同新 `value` 提交更新
- **则** 系统更新 `value`
- **且** 库内 `name` 与 `remark` 仍为更新前的存储原文
- **且** 后续中文环境下列表/详情仍可从 i18n 或库内 fallback 得到正确展示

#### Scenario:自定义参数仍可修改名称与备注
- **当** 管理员更新非内置参数的 `name` 或 `remark`
- **则** 系统按请求写入对应字段

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

系统 SHALL 为每条 `sys_config` 记录持久化 `system_manageable` 字段（`SMALLINT`，`1` 表示允许在系统参数设置管理面维护，`0` 表示不允许）。系统参数管理面的 List 与 Export MUST 仅返回 `system_manageable = 1` 的可见行。运行时配置读取 MUST 不受该字段影响。

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

### Requirement:启用 i18n 的插件 sys_config 展示键必须可本地化

对 `plugin.yaml` 中 `i18n.enabled: true` 且以 `SysConfigKey` 常量声明的插件配置键，系统 SHALL 在插件 `manifest/i18n/<locale>/` 中提供 `config.<config_key>.name` 与 `config.<config_key>.remark`，使参数设置页列表投影不为技术 key 裸展示。

#### Scenario:插件 SysConfigKey 缺译被门禁阻断
- **当** 启用 i18n 的插件新增 `hostconfigcap.SysConfigKey = "plugin.<id>...."` 常量且未补齐对应 `config.<key>.name/remark`
- **则** `make i18n.check` 失败并指出缺失键与 locale
- **且** 未启用 i18n 的插件不要求上述插件侧资源

### Requirement:内置参数 sys.auth.sloganImage

系统 SHALL 提供内置参数 `sys.auth.sloganImage`，用于配置登录页 slogan 插画图片地址。默认值为 `/slogan.svg`（Vben 内置插画）。空值表示不使用插画。

#### Scenario:参数设置页可见 slogan 参数
- **当** 管理员打开参数设置并搜索 `sys.auth.sloganImage` 时
- **则** 列表显示该内置参数
- **且** 参数名称标识为登录展示相关的 slogan 插画配置

#### Scenario:允许清空 slogan 地址以隐藏插画
- **当** 管理员将 `sys.auth.sloganImage` 保存为空值时
- **则** 系统接受该值
- **且** 登录页不展示 slogan 插画

