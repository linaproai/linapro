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

配置管理页面 SHALL 按当前语言本地化内置系统参数名称、描述和默认显示值，使英文环境不显示默认中文系统文案。

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
- **则** `zh-CN`、`en-US` 和 `zh-TW` 运行时资源保持匹配的键覆盖
- **且** 缺失翻译检查不报告新缺失的内置配置键

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

