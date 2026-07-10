# 运行时错误文案能力化

## Purpose
要求宿主运行时业务错误用能力语义描述框架约束，不得把官方插件品牌 ID 写入用户可见错误文案。

## Requirements

### Requirement: 宿主能力约束类错误文案不得绑定官方插件品牌 ID

系统 SHALL 在宿主运行时业务错误的英文源消息与运行时语言包中，用能力语义描述框架级约束（例如 multi-tenant governance、tenant-scoped 安装模式），不得将官方插件品牌 ID（如 `linapro-tenant-core`）写入用户可见错误文案来表达该约束。错误码与 `messageKey` 保持稳定；调用方 MUST 依赖 `errorCode` / `messageKey` 判断语义，不得依赖自然语言 `message` 中的插件品牌名。

#### Scenario: 新租户自动启用策略错误使用能力语义

- **当** 平台配置某插件为新租户自动启用，但该插件不支持多租户治理或未以 `tenant_scoped` 安装时
- **则** 返回错误码 `PLUGIN_TENANT_PROVISIONING_POLICY_INVALID`（或当前等价稳定码）
- **且** 默认英文源消息与 `manifest/i18n/**/error.json` 对应翻译描述 multi-tenant / framework tenant governance 约束
- **且** 用户可见消息中不得出现 `linapro-tenant-core` 品牌 ID

#### Scenario: 错误码与翻译键保持稳定

- **当** 仅调整上述错误的自然语言文案以去除官方插件品牌名时
- **则** `errorCode` 与 `messageKey` 不得变更
- **且** 宿主 `zh-CN`、`en-US`、`zh-TW`（若已维护该键）运行时语言包同步更新同一语义
