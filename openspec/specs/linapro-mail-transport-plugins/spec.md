# linapro-mail-transport-plugins Specification

## Purpose
TBD - created by archiving change plugin-mail-foundation. Update Purpose after archive.
## Requirements
### Requirement: 官方协议插件实现邮件 transport SPI

系统 SHALL 提供官方源码协议插件：`linapro-mail-smtp`（出站 kind=`smtp`）、`linapro-mail-imap`（入站 kind=`imap`）、`linapro-mail-pop3`（入站 kind=`pop3`）。各插件 MUST 向 `linapro-mail-core` 发布的 SPI 注册自身 kind 与 factory，并实现对应方向的协议操作（发送或拉取/同步）及连通性探测。

#### Scenario: SMTP 插件仅出站

- **WHEN** `linapro-mail-smtp` 启用且已配置有效 SMTP Connection
- **THEN** 该插件 MUST 能完成出站发送与 Probe
- **AND** MUST NOT 被要求实现 IMAP/POP3 入站接口

#### Scenario: IMAP 与 POP3 仅入站

- **WHEN** `linapro-mail-imap` 或 `linapro-mail-pop3` 启用且 Connection kind 匹配
- **THEN** 该插件 MUST 能完成入站拉取或同步与 Probe
- **AND** MUST NOT 被要求实现 SMTP 发送接口

### Requirement: 协议插件必须依赖 mail-core 且不拥有 Connection 权威数据

协议插件 MUST 在 `plugin.yaml` 中声明对 `linapro-mail-core` 的插件依赖。协议插件 MUST NOT 将 Connection/Account 权威数据存储在自身业务表中；MUST 使用 mail-core 提供的 Connection 能力或由 mail-core 传入的 endpoint DTO 执行协议。

#### Scenario: 缺少 mail-core 依赖时不可用

- **WHEN** `linapro-mail-core` 未安装或未启用
- **THEN** 协议插件 MUST 不能作为可服务 transport 通过依赖与 enablement 治理
- **AND** 邮件发送/接收 MUST fail-closed

### Requirement: 协议插件不得实现 kind 冲突检测

协议插件 MUST NOT 各自实现同 kind 多插件冲突检测逻辑。冲突检测与启用否决 MUST 由 `linapro-mail-core` 通过全局 lifecycle Hook 与 `Resolve(kind)` 完成。

#### Scenario: 第二 smtp 启用失败不依赖 smtp 插件代码

- **WHEN** 已启用一个 smtp 协议插件并尝试启用另一个同 kind 插件
- **THEN** 否决 MUST 由 mail-core 全局 Hook 产生
- **AND** 协议插件仓库中 MUST NOT 要求重复的冲突检测样板代码作为正确性前提

### Requirement: 协议插件配置面边界

协议插件可为连通性调试提供最小设置页，但用户可见的 Connection 主配置表单 MUST 以 mail-core 为准。协议插件 MUST 启用 i18n，并维护自身 manifest 资源。协议插件 MUST NOT 修改 `lina-core` 宿主领域契约。

#### Scenario: 插件可独立安装

- **WHEN** 管理员仅安装 `linapro-mail-core` 与 `linapro-mail-smtp`
- **THEN** 系统 MUST 支持仅出站邮件能力
- **AND** 未安装 imap/pop3 时入站 API MUST 明确失败而非崩溃

