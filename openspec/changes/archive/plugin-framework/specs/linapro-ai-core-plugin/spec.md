# linapro-ai-core-plugin 规范增量

## ADDED Requirements

### Requirement: 动态插件文本 AI 授权不得授予管理权限

系统 SHALL 确保动态插件获得`ai.text.generate`或其他`ai`host service 方法授权后，只能通过类型化`AI`能力发起对应方法调用。该授权 MUST NOT 授予渠道、模型、档位管理或调用日志管理 API 的访问权限。

#### Scenario: 插件调用不获得管理权限

- **WHEN** 动态插件获得`ai.text.generate`host service 方法授权
- **THEN** 该授权 MUST 只允许调用文本生成能力
- **AND** 插件 MUST NOT 因该授权访问渠道、档位管理或调用日志管理 API
- **AND** 插件请求中的`purpose`、`tier`和其他参数 MUST NOT 被解释为管理权限授权来源

### Requirement: AI 契约 owner 扩展

### Requirement: linapro-ai-core 必须拥有 AI 领域公开契约

系统 SHALL 扩展`linapro-ai-core`插件职责，使其除管理页面、provider、模型、档位和调用日志外，还拥有`AI`领域公开契约。插件 MUST 在`backend/cap/aicap`发布普通消费接口、DTO、命名类型、错误语义、方法状态、动态 guest SDK、provider SPI、descriptor helper 和版本策略。`backend/internal`继续承载 provider、模型路由、分层、调用日志、外部协议和业务实现。

#### Scenario: 创建 AI cap 目录

- **WHEN** 实现本变更
- **THEN** `apps/lina-plugins/linapro-ai-core/backend/cap/aicap` MUST 存在并作为`AI`领域公开契约入口
- **AND** 该目录 MUST NOT import `linapro-ai-core/backend/internal/**`
- **AND** 该目录 MUST NOT 暴露`DAO`、`DO`、`Entity`、密钥、模型路由内部配置或调用日志内部结构

#### Scenario: 插件 ID 不再依赖 core aicap

- **WHEN** `linapro-ai-core`声明自身插件 ID、provider ID 或能力 owner ID
- **THEN** 这些稳定标识 MUST 由插件自身公开契约或 manifest 维护
- **AND** `backend/plugin.go`不得通过 core `aicap/aitext`常量间接获得 owner 身份

### Requirement: linapro-ai-core 必须发布 AI capability descriptor

系统 SHALL 要求`linapro-ai-core`在源码插件注册阶段发布`AI`能力 descriptor。descriptor MUST 描述 owner 插件 ID、`ai`能力键、`v1`协议版本、当前已发布方法、风险等级、资源形态、源码契约、动态契约、provider factory、运行依赖和启用策略。descriptor 注册 MUST 通过 owner 插件提供的类型安全 helper 完成。已发布方法 MUST 与真实 invoker 路径一致，不得把仅存在 DTO 的多模态方法提前写入授权 catalog。

#### Scenario: 注册文本 provider

- **WHEN** `linapro-ai-core`注册文本生成 provider
- **THEN** 插件 MUST 使用`aicap.ProviderDescriptor`或等价 helper 将 typed factory 包装为带 invoker 的通用 descriptor
- **AND** invoker MUST 通过`aicap.Service`分发，不得在 SPI 包内复制一套并行业务 switch
- **AND** 不得调用 core `plugin.Providers().ProvideAIText`

#### Scenario: 发布动态方法目录

- **WHEN** `linapro-ai-core`发布`ai.v1`descriptor
- **THEN** descriptor MUST 发布当前可运行的文本和方法状态 methods
- **AND** 管理端和动态授权展示 MUST 能显示这些方法来自`linapro-ai-core`
- **AND** 尚未接线的多模态 methods MUST NOT 出现在 descriptor 授权目录中

### Requirement: linapro-ai-core bridge SDK 必须复用 owner DTO

系统 SHALL 要求`linapro-ai-core/backend/cap/aicap/bridge`或等价公开包提供动态插件 guest SDK。该 SDK MUST 复用`backend/cap/aicap`下普通消费 DTO 或 owner 契约定义的投影，MUST 只负责编码 owner-aware host call、声明 helper 和错误映射，不得包含 provider SPI、源码插件注册 API、宿主 dispatcher 或内部业务实现。

#### Scenario: 动态 SDK 复用文本 DTO

- **WHEN** 动态插件使用 bridge SDK 调用文本生成
- **THEN** 请求响应类型 MUST 与源码插件消费契约保持同一 owner
- **AND** 不得在 core protocol 中维护并行的文本生成 DTO owner

#### Scenario: bridge SDK 不绕过授权

- **WHEN** 动态插件通过 SDK 调用`AI`方法
- **THEN** SDK MUST 通过通用 host call 进入宿主授权和审计
- **AND** SDK 不得直接调用`linapro-ai-core/backend/internal/service`

### Requirement: linapro-ai-core i18n 和文档必须覆盖 owner 契约

系统 SHALL 按`linapro-ai-core/plugin.yaml`中的`i18n.enabled: true`治理插件 owner 契约的用户可见文案、错误 fallback、API 文档源文本、动态授权展示名称和 README 文档。插件自身运行时语言包和`apidoc`翻译 MUST 维护在该插件`manifest/i18n/<locale>/`和`manifest/i18n/<locale>/apidoc/`下，不得集中写入`lina-core`语言资源。

#### Scenario: 新增 owner 错误码

- **WHEN** `linapro-ai-core`新增`CodeCapabilityDenied`、`CodeCapabilityUnavailable`或等价 owner 错误码
- **THEN** 错误必须具备稳定 errorCode、messageKey、messageParams 和英文 fallback
- **AND** 插件必须维护对应目标语言资源

#### Scenario: 更新插件 README

- **WHEN** 完成`AI`契约 owner 迁移
- **THEN** `apps/lina-plugins/linapro-ai-core/README.md`和`README.zh-CN.md` MUST 同步说明`backend/cap/aicap`、源码插件消费、动态 SDK、provider SPI、依赖声明和版本策略
- **AND** 两个 README 的事实内容必须一致
