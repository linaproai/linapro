# 文本 AI 能力 owner 迁移

## Purpose
保留文本 AI 契约语义，迁移契约包、provider SPI 与动态 SDK owner。

## Requirements

### Requirement: 宿主必须发布文本 AI 抽象能力

系统 SHALL 由`linapro-ai-core/backend/cap/aicap/aitext`发布版本化文本`AI`抽象能力`plugin.linapro-ai-core.ai.text.v1`。该能力 MUST 只定义消费契约、状态查询、降级语义和 provider 接入边界，MUST NOT 在`lina-core`中持有渠道、模型、档位或调用日志业务存储。`lina-core`只通过通用 capability descriptor、依赖治理、授权快照、动态路由和审计 envelope 治理该能力。

#### Scenario: 消费方通过文本能力接口调用

- **WHEN** 宿主模块、源码插件或动态插件需要执行文本生成
- **THEN** 调用方 MUST 通过`linapro-ai-core/backend/cap/aicap/aitext`发布的`plugin.linapro-ai-core.ai.text.v1`消费接口发起调用
- **AND** 调用方 MUST NOT 直接依赖`linapro-ai-core`的`backend/internal/**`、插件表、渠道密钥结构或 provider adapter
- **AND** 调用方 MUST NOT 继续依赖`lina-core/pkg/plugin/capability/aicap/aitext`作为生产契约 owner

#### Scenario: 官方插件提供文本能力实现

- **WHEN** `linapro-ai-core`插件处于平台级可用状态并声明`plugin.linapro-ai-core.ai.text.v1`provider
- **THEN** `plugin.linapro-ai-core.ai.text.v1`的消费 service MUST 将文本生成调用委托给该 provider
- **AND** 返回值 MUST 使用该能力自有 DTO、投影和值对象

#### Scenario: 文本生成映射到固定能力方法

- **WHEN** 调用方通过`plugin.linapro-ai-core.ai.text.v1`执行`GenerateText`
- **THEN** owner 契约 MUST 将该调用视为`capabilityType=text`与`capabilityMethod=generate`
- **AND** Go 契约 MUST 使用命名类型和常量表达该能力方法
- **AND** 调用方 MUST NOT 通过请求字段把`GenerateText`改写为图片、向量、音频或其他方法

#### Scenario: 渠道存储不进入宿主

- **WHEN** 系统实现`plugin.linapro-ai-core.ai.text.v1`
- **THEN** `apps/lina-core` MUST NOT 新增渠道、模型、档位或调用日志业务表
- **AND** owner 公开契约 MUST NOT 暴露插件内部`DAO`、`DO`、`Entity`、缓存快照或密钥明文

### Requirement: 文本能力必须提供可用性和降级状态

系统 SHALL 为`plugin.linapro-ai-core.ai.text.v1`提供`Available(ctx)`、`Status(ctx)`和方法级状态等状态能力。owner 插件禁用、卸载、provider 冲突、档位未配置、模型禁用或密钥不可用时，系统 MUST 返回明确的不可用状态或业务错误，而不是产生宿主 500。动态插件读取状态时 MUST 通过 owner-aware host service 和授权快照进入 owner handler。

#### Scenario: Provider 插件不可用

- **WHEN** `linapro-ai-core`插件被禁用、卸载或启动失败
- **THEN** `Available(ctx)` MUST 返回不可用
- **AND** `Status(ctx)` MUST 返回能力 ID、provider 插件状态和不可用原因
- **AND** 调用`GenerateText` MUST 返回结构化不可用错误

#### Scenario: 档位未配置

- **WHEN** 调用方使用未配置启用主绑定的档位生成文本
- **THEN** 系统 MUST 拒绝调用
- **AND** 错误 MUST 明确指出该档位未配置可用渠道模型

#### Scenario: 可选消费方降级

- **WHEN** 业务功能可选使用文本`AI`能力
- **AND** `Available(ctx)`返回不可用
- **THEN** 业务功能 MUST 隐藏入口、返回零值或按自身规范提示配置缺失
- **AND** 业务功能 MUST NOT 直接暴露 provider 内部错误或空白页面

### Requirement: 文本 AI 能力必须提供方法级状态

系统 SHALL 由`linapro-ai-core/backend/cap/aicap/aitext`为文本`AI`能力提供`Text().MethodStatus`或等价方法级状态读取，并通过 owner descriptor 动态发布为`text.method_status.get`或等价冻结名称。该方法 MUST 与其他`AI`子能力状态语义一致，响应不得暴露 provider 私有模型配置。

#### Scenario: 读取文本生成方法状态

- **WHEN** 插件读取文本生成方法状态
- **THEN** 系统返回该方法是否可用、不可用原因和稳定状态码
- **AND** 不返回 provider 私有模型配置

#### Scenario: 文本 provider 不可用

- **WHEN** 文本 provider 未启用或当前 purpose 不可用
- **THEN** 系统返回结构化 unavailable 状态
- **AND** 插件可以据此降级而不触发 provider 内部错误


### Requirement: 文本 AI 动态 SDK 必须由 owner 插件发布

系统 SHALL 将文本`AI`动态 guest SDK、声明 helper、codec 和 DTO 复用入口迁移到`linapro-ai-core/backend/cap/aicap/bridge`或 owner 插件内等价公开包。动态 SDK MUST 复用 owner 契约 DTO，MUST NOT 在 core `pluginbridge/protocol`中继续维护文本 AI 专属请求别名、codec 或常量作为生产 owner。

#### Scenario: 动态插件调用文本生成

- **WHEN** 动态插件调用 owner SDK 的文本生成 helper
- **THEN** SDK MUST 编码 owner、service、version、method 和 payload
- **AND** SDK MUST 通过通用 host call 进入宿主授权和转发路径
- **AND** SDK 不得直接调用 owner 插件内部实现或绕过授权快照

#### Scenario: 文本 DTO 不分裂

- **WHEN** 源码插件和动态插件同时调用文本生成
- **THEN** 两者 MUST 使用同一 owner 契约 DTO 或由 owner 契约定义的等价投影
- **AND** 动态 DTO 不得在 core 中形成与源码 DTO 并行演化的第二套 owner
