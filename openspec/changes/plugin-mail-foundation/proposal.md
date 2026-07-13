## Why

框架与业务插件需要可治理的邮件发送/接收基础能力，但 SMTP/IMAP/POP3 协议细节、连接凭据与多账号管理不应进入 `lina-core` 宿主膨胀 `pkg/plugin`。宿主已有 `notify` 域并预留了 `email` 通道类型，但发送路径尚未落地，且缺少「同协议多实现」的启用期治理。需要以 plugin-owned 方式提供 `linapro-mail-core`，用协议插件承载具体 transport，并通过通用**全局生命周期 Hook** 让 owner 在安装/启用时集中做冲突拦截。

## What Changes

- 扩展插件生命周期契约：为源码插件补齐 `BeforeEnable` / `AfterEnable`（目标插件自管），并新增**全局前置 Hook**（`GlobalBeforeInstall`、`GlobalBeforeEnable`，以及按需的 `GlobalBeforeDisable` / `GlobalBeforeUninstall`），使已注册的 owner 插件可否决**其他**插件的安装/启用，而宿主插件管理模块不感知业务领域。
- 新增官方源码插件 `linapro-mail-core`：拥有 Connection、Account、`mailcap` 公开契约、transport SPI 注册与 kind 级单例解析；Connection 配置与表单统一在 core 维护。
- 新增协议实现插件：`linapro-mail-smtp`（出站）、`linapro-mail-imap` / `linapro-mail-pop3`（入站），仅实现 SPI，不自建 connection 表与冲突检测逻辑。
- 多 Account 支持（如 Gmail/QQ/iCloud 等不同 connection）；同一 transport kind 仅允许一个可服务插件（允许安装多个、禁止同时启用，由 mail-core 全局 Hook 拦截 + Resolve 运行时兜底）。
- Account 可仅出站（只绑 outbound connection）；调用入站 API 时若无 SPI 或未绑定 inbound 则明确失败。
- 宿主 `notify` 的 `email` 通道发送编排依赖 `linapro-mail-core`（`mailcap`），不得直接依赖具体 smtp/imap/pop3 插件。
- 明确不在本变更将完整「邮箱客户端产品」（线程 UI、规则引擎、营销投放等）纳入范围；先交付可治理的发送/接收基础能力与管理面最小闭环。

## Capabilities

### New Capabilities

- `plugin-lifecycle-global-hooks`：定义目标插件 `BeforeEnable`/`AfterEnable` 与全局前置 lifecycle Hook 的注册、参与者范围、输入语义、veto 聚合与启用路径接入要求。
- `linapro-mail-core`：定义 plugin-owned 邮件领域 owner——Connection/Account、`mailcap` 契约、transport SPI、kind 单例解析、管理 API/页面与全局冲突检测职责。
- `linapro-mail-transport-plugins`：定义 smtp/imap/pop3 协议插件边界、对 mail-core 的依赖、SPI 实现与零 connection 自有表要求。

### Modified Capabilities

- `plugin-notify-service`：`email` 通道发送 MUST 通过 `linapro-mail-core` 完成出站，不得直连协议插件；mail-core 不可用时通道 fail-closed。
- `plugin-manifest-lifecycle`：安装/启用/禁用编排 MUST 接入目标与全局前置 Hook（在启用路径补齐 precondition），且不因全局 Hook 改变既有 force/purge/依赖检查语义。
- `core-host-boundary-governance`：邮件协议与连接配置 MUST 不进入宿主 `pkg/plugin` 领域契约；邮件公开契约位于 owner 插件 `backend/cap`。

## Impact

- **宿主插件生命周期**：`pluginhost` lifecycle 注册面、`internal/service/plugin/internal/lifecycle` 的 install/enable 编排、veto 错误投影与 i18n reason。
- **宿主 notify**：`ChannelTypeEmail` 发送路径、delivery 状态、与 mail-core 的可选/硬依赖桥接策略。
- **新增插件目录**：`apps/lina-plugins/linapro-mail-core/`、`linapro-mail-smtp/`、`linapro-mail-imap/`、`linapro-mail-pop3/`（源码插件，`plugin.yaml` / SQL / i18n / 管理页 / E2E）。
- **跨插件依赖**：transport 与业务消费方声明对 `linapro-mail-core` 的 `dependencies.plugins`；动态插件若消费邮件能力则走 owner-aware hostServices（一期可仅源码路径）。
- **数据与权限**：Connection/Account 表、密钥引用、租户边界、管理面数据权限与审计。
- **i18n**：生命周期 veto reason、邮件管理菜单/表单/错误、API 文档源文本与插件语言包。
- **测试**：全局 Hook 单测/集成、kind 冲突、仅出站 Account、notify email 通道与协议 SPI 探测。
