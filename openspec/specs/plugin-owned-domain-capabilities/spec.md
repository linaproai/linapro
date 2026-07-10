# plugin-owned-domain-capabilities Specification

## Purpose
TBD - created by archiving change plugin-owned-domain-capabilities. Update Purpose after archive.
## Requirements
### Requirement: 非核心领域能力必须由领域 owner 插件拥有

系统 SHALL 将插件可消费领域能力分为 core-owned 与 plugin-owned 两类。`lina-core`只拥有插件内核能力、宿主通用业务能力和通用治理协议；非核心、变化快、领域实现已归属插件的能力 MUST 由对应 owner 插件维护公开契约、源码插件 helper、动态 guest SDK、provider SPI、能力描述符和版本策略。系统 MUST NOT 为每个非核心领域继续在`lina-core/pkg/plugin/capability`、`pluginhost`或`pluginbridge`中追加领域专属 Go 方法、DTO、codec 或 dispatcher 分支。

#### Scenario: 新增非核心领域能力

- **WHEN** 系统新增内容、知识库、工作流、CRM 或其他非核心业务领域能力
- **THEN** 该能力的公开契约 MUST 位于 owner 插件的`backend/cap/<domain>cap`
- **AND** `lina-core`只接收该 owner 发布的通用能力描述符、依赖声明、动态授权和生命周期治理信息

#### Scenario: 宿主内核能力继续归属 core

- **WHEN** 能力属于插件运行、隔离、授权、资源访问或治理必需的宿主内核
- **THEN** 该能力 MAY 继续归属`lina-core/pkg/plugin`
- **AND** 例如`runtime`、`storage`、`cache`、`lock`、`manifest`、`plugins`、`bizctx`和`route`不因本变更迁移到业务插件

### Requirement: owner 插件公开契约必须位于 backend/cap

系统 SHALL 使用`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`作为插件拥有领域能力的唯一公开 Go 契约目录。该目录 MAY 包含普通消费`Service`、DTO、值对象、命名类型、错误码、能力状态投影、动态 guest SDK 子包和 provider SPI 子包。该目录 MUST NOT 包含`DAO`、`DO`、`Entity`、controller、私有缓存结构、provider 密钥、内部 service 实现或宿主内部 dispatcher。

#### Scenario: 源码消费插件 import owner 契约

- **WHEN** 源码消费插件需要调用`linapro-ai-core`发布的`AI`能力
- **THEN** 它只能 import `lina-plugin-linapro-ai-core/backend/cap/aicap/...`
- **AND** 它 MUST NOT import `lina-plugin-linapro-ai-core/backend/internal/...`、`backend/internal/dao`、`backend/internal/model`或 provider 实现包

#### Scenario: backend/pkg 不作为领域能力入口

- **WHEN** owner 插件存在`backend/pkg`或其他公共 helper 目录
- **THEN** 该目录 MUST NOT 成为跨插件领域能力入口
- **AND** 治理扫描 MUST 将跨插件生产 import 限定到`backend/cap/...`

### Requirement: owner 能力必须通过通用 descriptor 注册

系统 SHALL 由 owner 插件发布通用 capability descriptor，描述`ownerPluginId`、`capability`、`version`、方法、风险、资源形态、源码契约、动态契约、运行依赖和启用策略。`lina-core` MUST 使用 descriptor 建立`owner + capability + version + method`索引，并通过该索引完成发现、授权、动态路由、审计和文档投影。descriptor 的具体业务 factory 类型 MUST 由 owner 插件的类型安全 helper 封装，core 不得通过`any`暴露给调用方自行断言。

#### Scenario: owner 插件注册源码 provider

- **WHEN** `linapro-ai-core`声明`AI`文本 provider
- **THEN** 它通过 owner helper 生成通用 capability descriptor 并注册到 core 能力注册表
- **AND** `pluginhost`不得 import `linapro-ai-core/backend/cap/aicap`来定义`ProvideAIText`专属方法

#### Scenario: descriptor 方法重复注册

- **WHEN** 两个插件或同一插件两次注册相同`owner + capability + version + method`
- **THEN** core MUST 拒绝重复注册或进入规范定义的 provider 冲突状态
- **AND** 错误或状态必须包含 owner 插件、capability、version 和 method

### Requirement: 动态 owner 能力必须显式声明 owner 和版本

系统 SHALL 要求动态插件在`plugin.yaml hostServices`中为 plugin-owned 能力显式声明`owner`和`version`。动态插件申请 owner 能力时，清单 MUST 同时在`dependencies.plugins`中声明对应 owner 插件和版本约束。宿主在构建、安装、启用、升级和运行时 MUST 校验 owner 依赖、方法注册、授权快照、资源范围、payload 结构和 owner 插件启用状态。

#### Scenario: 动态插件申请 AI 文本生成

- **WHEN** 动态插件声明`service: ai`、`owner: linapro-ai-core`、`version: v1`和`methods: [text.generate]`
- **THEN** 该插件 MUST 同时声明`dependencies.plugins`依赖`linapro-ai-core`
- **AND** 宿主仅在 owner 已安装、已启用、版本满足且方法授权确认后允许运行时调用

#### Scenario: 未声明 owner 依赖

- **WHEN** 动态插件声明 owner 能力方法但未在`dependencies.plugins`中依赖 owner 插件
- **THEN** manifest 校验、catalog 校验或安装启用检查 MUST 失败
- **AND** 运行时不得为该调用创建授权快照

### Requirement: owner 能力调用必须保持宿主治理边界

系统 SHALL 为源码插件和动态插件的 owner 能力调用保留宿主治理边界。每次调用 MUST 携带调用插件 ID、actor、tenant、授权快照、资源声明、执行来源和审计信息。owner 能力读取或操作宿主数据时 MUST 在 owner 查询阶段注入数据权限；执行、更新、删除和批量动作 MUST 在操作前校验目标可见性。owner 插件不得通过公开契约暴露 provider 密钥、内部模型路由、私有配置、调用日志详情或其他插件数据。

#### Scenario: owner 能力读取宿主资源

- **WHEN** `AI`能力未来按文件 ID、知识库 ID 或业务记录 ID 读取上下文
- **THEN** owner 方法 MUST 在调用 provider 前校验资源对当前 actor、tenant 和调用插件可见
- **AND** 不存在、不可见和租户外目标 MUST 使用统一拒绝语义

#### Scenario: 调用日志不跨插件泄露

- **WHEN** 消费插件调用`AI`生成能力
- **THEN** 该插件不得通过能力响应读取 provider 密钥、模型路由内部配置或其他插件调用日志
- **AND** 管理调用日志只能通过`linapro-ai-core`自身管理 API 和权限边界读取

### Requirement: owner 能力运行时缓存必须满足关键数据一致性

系统 SHALL 将能力注册表、授权快照、owner 可用性和 SDK catalog 视为关键运行时数据。设计和实现 MUST 明确权威数据源、一致性模型、失效触发点、跨实例同步机制、最大可接受陈旧时间、故障降级策略、可观测性和恢复路径。插件安装、启用、禁用、升级和卸载事务提交成功后，系统 MUST 幂等失效或重建受影响的能力注册表与授权快照。

#### Scenario: 集群模式 owner 插件禁用

- **WHEN** `cluster.enabled=true`且某节点禁用 owner 插件
- **THEN** 该节点 MUST 在事务提交后发布插件运行时修订、共享修订号或等价事件
- **AND** 其他节点观察到事件后刷新 owner 能力注册表和授权快照
- **AND** 动态插件不得在过期授权下无限期继续调用该 owner 能力

#### Scenario: 缓存后端不可用

- **WHEN** 能力注册表或授权快照缓存后端不可用
- **THEN** 系统 MUST 从已发现插件、源码注册和动态 artifact 描述符等权威源重建，或返回明确的能力不可用错误
- **AND** 集群模式下不得退化为仅当前节点可见的本地状态作为长期授权依据

