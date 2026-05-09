## ADDED Requirements

### Requirement: 多租户相关术语 i18n key 完整覆盖
工作台 i18n 资源 SHALL 包含完整的中英双语翻译,涵盖:
- 租户(Tenant)、平台管理员(Platform Administrator)、租户管理员(Tenant Administrator)、成员(Member)、邀请(Invite)、切换租户(Switch Tenant)、impersonation(Acting as Tenant)。
- 登录挑选器、租户选择、租户挂起、租户归档、租户删除等流程文案。
- 否决理由 reason key(plugin lifecycle guard 类)。
- 错误码消息(`bizerr.Code*` 中所有多租户相关码)。

#### Scenario: 双语完整
- **WHEN** 检查 `manifest/i18n/zh-CN/*.json` 与 `manifest/i18n/en-US/*.json`
- **THEN** 多租户相关 key 在两种语言中均存在
- **AND** 无遗漏

### Requirement: 翻译完整性自动校验
CI 测试 SHALL 包含针对多租户 i18n key 的覆盖校验;任意目标语言缺失某 key 时阻断发布。

#### Scenario: 缺失阻断
- **WHEN** 开发提交时遗漏 `tenant.suspended.confirm` 的 zh-CN 翻译
- **THEN** CI 测试失败,提示具体缺失 key

### Requirement: 平台/租户菜单分组 i18n
菜单 i18n 投影规则 SHALL 区分"平台管理"与"租户管理"两个顶级分组;每个分组下条目命名约定遵循已有 i18n module projection 规范。

#### Scenario: 分组顶层 key 唯一
- **WHEN** 检查工作台菜单 i18n 资源
- **THEN** 平台管理项位于 `menu.platform.*` 命名空间
- **AND** 租户管理项位于 `menu.tenant.*` 命名空间
