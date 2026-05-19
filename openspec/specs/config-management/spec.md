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

