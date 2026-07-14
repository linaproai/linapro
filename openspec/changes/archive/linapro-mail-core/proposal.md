## Why

框架与业务插件需要可治理的邮件发送与接收基础能力，但 SMTP/IMAP/POP3 协议细节、连接凭据与账号配置不得进入 `lina-core` 宿主并膨胀 `pkg/plugin`。宿主 `notify` 域已预留 `email` 通道，但发送路径需落地为可观测投递；同协议多实现还需在启用期由领域 owner 集中冲突治理，而非散落在各协议插件或仅依赖运行时 `Resolve`。

## What Changes

- 扩展插件生命周期：目标插件补齐 `BeforeEnable` / `AfterEnable`；新增通用全局前置 Hook（`GlobalBeforeInstall`、`GlobalBeforeEnable`，以及 `GlobalBeforeDisable` / `GlobalBeforeUninstall` 注册面），使已注册 owner 可否决**其他**插件的安装/启用，宿主插件管理不感知业务领域。
- 新增官方源码插件 `linapro-mail-core`（`distribution: managed`）：拥有 Connection、Account、`mailcap` 公开契约、transport SPI 注册与 kind 级单例解析；Connection 配置与平台邮件设置管理面统一在 core。
- 新增协议实现插件：`linapro-mail-smtp`（出站）、`linapro-mail-imap` / `linapro-mail-pop3`（入站），仅实现 SPI，不自建 Connection/Account 权威表，不实现 kind 冲突检测。
- 同 transport kind 仅允许一个可服务插件（可安装多个、禁止同时启用）；由 mail-core 全局 Hook 拦截 + `Resolve` 运行时兜底。
- 管理面一期收敛为平台唯一邮件账号设置（SMTP/IMAP/POP3 直填、保存、测试连接、测试发送、测试接收）；数据模型仍支持出站/入站可选绑定与仅出站语义。
- 宿主 `notify` 的 `email` 通道发送编排依赖 `linapro-mail-core`（`mailcap` / 进程内 email delivery 桥接），不得直连 smtp/imap/pop3；mail-core 不可用时 fail-closed。
- 明确不纳入完整邮箱客户端产品（线程 UI、规则引擎、营销投放等）；先交付可治理发送/接收基础能力与管理面闭环。

## Capabilities

### New Capabilities

- `plugin-lifecycle-global-hooks`：目标插件 `BeforeEnable`/`AfterEnable` 与全局前置 lifecycle Hook 的注册、参与者范围、输入语义、veto 聚合与启用路径接入。
- `linapro-mail-core`：plugin-owned 邮件领域 owner——Connection/Account、`mailcap`、transport SPI、kind 单例解析、管理 API/设置页与全局冲突检测。
- `linapro-mail-transport-plugins`：smtp/imap/pop3 协议插件边界、对 mail-core 的依赖、SPI 实现与零 connection 自有表要求。

### Modified Capabilities

- `plugin-notify-service`：`email` 通道发送 MUST 通过 `linapro-mail-core` 完成出站，不得直连协议插件；mail-core 不可用时通道 fail-closed。
- `plugin-manifest-lifecycle`：安装/启用/禁用编排 MUST 接入目标与全局前置 Hook；不得因全局 Hook 改变 force/purge/依赖检查语义。
- `core-host-boundary-governance`：邮件协议与连接配置 MUST 不进入宿主 `pkg/plugin` 领域契约；邮件公开契约位于 owner 插件 `backend/cap`。

## Impact

- **宿主插件生命周期**：`pluginhost` lifecycle 注册面、`internal/service/plugin/internal/lifecycle` 的 install/enable 编排、veto 错误投影与 i18n reason；`pkg/plugin` README。
- **宿主 notify**：`ChannelTypeEmail` 发送路径、delivery 状态、与 mail-core 的进程内 email delivery 桥接（可选能力，未装则 email 不可用）。
- **插件目录**：`apps/lina-plugins/linapro-mail-core/`、`linapro-mail-smtp/`、`linapro-mail-imap/`、`linapro-mail-pop3/`（源码插件：`plugin.yaml`、SQL、i18n、管理页、E2E）。
- **跨插件依赖**：transport 硬依赖 `linapro-mail-core`；业务消费方声明 `dependencies.plugins`；动态插件消费邮件能力走 owner-aware hostServices 为后续可选。
- **数据与权限**：Connection/Account 表（含 `tenant_id` 预留）、密钥引用、平台配置控制面数据权限与审计。
- **i18n**：生命周期 veto reason、邮件管理菜单/表单/错误、API 文档源文本、插件展示元数据与语言包；`make i18n.check` 校验 `plugin.<id>.name/description`。
- **测试**：全局 Hook 单测、kind 冲突、仅出站入站错误、notify email 通道、协议 SPI、插件 E2E（Connection/Account API 与设置页）。
