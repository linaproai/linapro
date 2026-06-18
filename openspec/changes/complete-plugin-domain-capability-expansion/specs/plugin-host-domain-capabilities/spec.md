## ADDED Requirements

### Requirement: 阶段一后续候选与搜索能力必须有界发布
系统 SHALL 补齐阶段 1.5 的普通领域候选、搜索和可见性能力，包括扩展`Users.Search`稳定过滤、新增`Dict.ListValues`、`Files.Search`、`Jobs.Search`、`Jobs.EnsureVisible`、`Sessions.BatchGetUserOnlineStatus`和`Sessions.EnsureVisible`。这些方法 MUST 定义分页或数量上限、稳定排序、数据库侧租户和数据权限过滤、动态授权资源和超限错误。

#### Scenario: 用户搜索使用稳定过滤
- **WHEN** 插件按关键字、状态、租户或启用状态搜索用户候选
- **THEN** 系统在数据库查询阶段应用租户和数据权限过滤
- **AND** 响应只返回有界用户投影，不诱导逐项用户详情补查

#### Scenario: 字典候选分页
- **WHEN** 插件按字典类型读取字典值候选
- **THEN** 系统按类型、租户覆盖、状态和排序返回有界分页或 limit 结果
- **AND** 不得无上限返回全部字典值

#### Scenario: 在线状态批量读取
- **WHEN** 插件批量判断多个用户是否在线
- **THEN** 系统使用 session owner 的批量路径或投影查询完成
- **AND** 不得对每个用户执行一次在线会话查询

### Requirement: 组织租户和插件治理投影必须安全降级
系统 SHALL 补齐阶段 2 的`Org`、`Tenant`和`Plugins`投影能力。组织能力 MUST 支持批量用户组织档案、受限部门树、部门搜索、岗位候选和部门/岗位可见性校验；租户能力 MUST 支持当前租户详情、批量租户投影、租户搜索、批量用户租户列表和批量租户可见性校验；插件治理能力 MUST 支持当前插件投影、插件搜索、分页租户插件列表和能力状态批量读取。Provider 缺失或禁用时 MUST 返回规范定义的空投影、空页、不可用错误或中性状态，不得放开全量数据。多个同一 singleton 能力 provider 同时可服务时 MUST 返回`CodeCapabilityProviderConflict`，能力状态 reason MUST 为`provider_conflict`。

#### Scenario: 组织 provider 禁用
- **WHEN** 插件读取用户组织档案且组织 provider 未启用
- **THEN** 系统返回空组织档案或明确能力不可用结果
- **AND** 不得通过宿主内部组织表或未授权 provider 数据兜底

#### Scenario: 领域 provider 多实例冲突
- **WHEN** 同一 singleton 领域能力存在多个可服务 provider
- **THEN** 系统返回`CodeCapabilityProviderConflict`
- **AND** 能力状态 reason 为`provider_conflict`

#### Scenario: 租户批量投影
- **WHEN** 插件按租户 ID 集合读取租户投影
- **THEN** 系统只返回当前 actor 可见的租户
- **AND** 不存在、不可见和租户外目标统一进入`MissingIDs`

#### Scenario: 插件治理分页搜索
- **WHEN** 插件按插件 ID、名称、类型或启用状态搜索插件
- **THEN** 系统使用插件治理读模型、缓存快照或集合化查询返回分页结果
- **AND** 不得为每个插件重复扫描 manifest 或重复读取启用状态

### Requirement: 插件私有资源批量能力必须保持插件和租户作用域
系统 SHALL 补齐阶段 3 的插件私有资源批量能力，包括`Storage.BatchStat`、`Storage.ListCursor`、`Storage.DeleteMany`、`Cache.GetMany`、`Cache.SetMany`、`Cache.DeleteMany`、`Manifest.GetMany`、`Manifest.List`以及动态 runtime state 多键读写删。所有方法 MUST 受当前插件 ID、租户、资源授权、路径或 key 数量上限约束，并复用既有共享后端或 owner 实例。

#### Scenario: Storage 批量元数据读取
- **WHEN** 动态插件批量读取多个私有对象元数据
- **THEN** 系统只返回当前插件和租户作用域下且资源授权允许的路径
- **AND** 不得暴露宿主物理路径或 provider 私有 key

#### Scenario: Cache 多键写入
- **WHEN** 插件批量写入缓存键
- **THEN** 系统复用既有缓存后端和命名空间隔离
- **AND** 缓存仍为非权威加速数据，不改变权限、配置或业务记录权威来源

#### Scenario: Runtime state 多键删除
- **WHEN** 动态插件批量删除运行态 key
- **THEN** 系统只删除当前插件和租户作用域下的 key
- **AND** 删除不存在 key 不得泄露其他插件状态空间

### Requirement: 通知和 AI 状态能力必须类型化并可批量降级
系统 SHALL 补齐阶段 4 的通知类型化和`AI`状态能力。通知读取 MUST 返回稳定类型化投影，支持按业务来源批量读取和可见性校验；`AI`能力 MUST 支持文本方法级状态和跨子能力方法状态批量读取。结果不得暴露 provider 密钥、模型映射、供应商配置或通知内部存储结构。

#### Scenario: 通知按来源批量读取
- **WHEN** 插件按`SourceType + SourceIDs`批量读取消息
- **THEN** 系统返回当前 actor 可见的类型化消息投影
- **AND** 不得用`map[string]any`暴露未治理字段

#### Scenario: AI 方法状态批量读取
- **WHEN** 插件批量读取多个`AI`子能力方法状态
- **THEN** 系统返回可用性、禁用原因或能力不可用的结构化状态
- **AND** 不返回 provider 配置、API key 或模型路由内部细节

### Requirement: 剩余阶段完成记录必须覆盖影响和验证证据
系统 SHALL 在任务记录和审查结论中证明阶段 1.5 至阶段 5 全部已实现或被规范明确排除，并记录`i18n`、缓存一致性、数据权限、数据库、开发工具、测试和 E2E 的影响判断。

#### Scenario: 完整方案完成审计
- **WHEN** 本变更准备标记完成
- **THEN** 任务记录必须逐项列出`localdocs`阶段 1.5 至阶段 5 的实现、排除或延后依据
- **AND** 验证证据必须覆盖 OpenSpec strict、目标 Go 测试、启动绑定 smoke、静态边界检索和`lina-review`
