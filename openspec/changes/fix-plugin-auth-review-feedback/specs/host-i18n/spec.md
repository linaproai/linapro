## ADDED Requirements

### Requirement: 宿主新增公开认证 API 必须有 apidoc 翻译条目

系统 SHALL 为 `/auth/providers` 接口的请求与响应 DTO 在 `manifest/i18n/<locale>/apidoc/core-api-auth.json` 中提供完整翻译条目。`zh-CN/apidoc/core-api-auth.json` MUST 包含 `ListProvidersReq` 的 meta 摘要、`ListProvidersRes.fields.providers` 描述，以及 `ProviderEntity` 全部字段的 `dc` 翻译。`en-US/apidoc/core-api-auth.json` MUST 至少存在为空对象占位文件，以保持与其他 host apidoc 模块一致的目录形态。

#### Scenario: zh-CN apidoc 覆盖 ListProviders 投影

- **WHEN** 宿主在 `api/auth/v1/auth_provider.go` 声明 `ListProvidersReq`、`ListProvidersRes` 与 `ProviderEntity`
- **THEN** `apps/lina-core/manifest/i18n/zh-CN/apidoc/core-api-auth.json` 同步包含 `core.api.auth.v1.ListProvidersReq.meta` 与 `schema`
- **AND** 同时包含 `core.api.auth.v1.ListProvidersRes.fields.providers`
- **AND** 同时包含 `core.api.auth.v1.ProviderEntity.fields` 下每个字段的 `dc`

#### Scenario: en-US apidoc 保持占位文件

- **WHEN** 宿主在 `manifest/i18n/zh-CN/apidoc/` 维护非空 `core-api-auth.json`
- **THEN** `manifest/i18n/en-US/apidoc/core-api-auth.json` 至少存在
- **AND** 文件内容为合法 JSON
- **AND** 文件内容可以是空对象 `{}` 占位

### Requirement: 宿主新增 bizerr 错误必须有运行时翻译条目

系统 SHALL 在新增 `bizerr.Code` 时同步为其 `MessageKey` 在所有已启用语言的 `manifest/i18n/<locale>/error.json` 中提供翻译条目。`AUTH_EXTERNAL_IDENTITY_INVALID` 与 `AUTH_EXTERNAL_USER_NOT_PROVISIONED` MUST 在 `zh-CN/error.json` 和 `en-US/error.json` 的 `error.auth.external.identityInvalid` 与 `error.auth.external.userNotProvisioned` 路径下各自具备真实文案，而不是占位字符串。

#### Scenario: zh-CN 运行时错误本地化

- **WHEN** 宿主 `internal/service/auth/auth_code.go` 中 `CodeAuthExternalIdentityInvalid` 与 `CodeAuthExternalUserNotProvisioned` 已注册
- **THEN** `apps/lina-core/manifest/i18n/zh-CN/error.json` 中 `error.auth.external.identityInvalid` 存在非空文案
- **AND** `error.auth.external.userNotProvisioned` 存在非空文案

#### Scenario: en-US 运行时错误本地化

- **WHEN** 宿主 `internal/service/auth/auth_code.go` 中 `CodeAuthExternalIdentityInvalid` 与 `CodeAuthExternalUserNotProvisioned` 已注册
- **THEN** `apps/lina-core/manifest/i18n/en-US/error.json` 中 `error.auth.external.identityInvalid` 存在非空文案
- **AND** `error.auth.external.userNotProvisioned` 存在非空文案
