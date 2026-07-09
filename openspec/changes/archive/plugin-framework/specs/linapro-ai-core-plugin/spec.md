# linapro-ai-core-plugin 规范增量

## ADDED Requirements

### Requirement: 动态插件文本 AI 授权不得授予管理权限

系统 SHALL 确保动态插件获得`ai.text.generate`或其他`ai`host service 方法授权后，只能通过类型化`AI`能力发起对应方法调用。该授权 MUST NOT 授予渠道、模型、档位管理或调用日志管理 API 的访问权限。

#### Scenario: 插件调用不获得管理权限

- **WHEN** 动态插件获得`ai.text.generate`host service 方法授权
- **THEN** 该授权 MUST 只允许调用文本生成能力
- **AND** 插件 MUST NOT 因该授权访问渠道、档位管理或调用日志管理 API
- **AND** 插件请求中的`purpose`、`tier`和其他参数 MUST NOT 被解释为管理权限授权来源
