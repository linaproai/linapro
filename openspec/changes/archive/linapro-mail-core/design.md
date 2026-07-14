# Design

## Context

宿主 `notify` 已有 `sys_notify_channel` / `sys_notify_message` / `sys_notify_delivery`，`channel_type` 含 `inbox | email | webhook`，但发送仅落地 `inbox`。插件可见 `notifycap.Service` 已声明邮件通道语义，缺 SMTP 实现。

插件生命周期在 Install/Disable/Uninstall 上有目标插件自管 precondition；源码 Enable 路径需在写状态前接入 Target `BeforeEnable` + Global `BeforeEnable`。租户删除等场景的「全体 lifecycle 参与者」语义是自管租户钩子，不是「否决他人安装/启用」。

对象存储在 `Resolve` 时做同领域唯一活跃 provider；邮件要求同 transport kind 唯一，并在**启用时**由 owner 集中拦截。AI 领域已验证 plugin-owned 模式：`linapro-ai-core` + `backend/cap` + SPI，宿主不膨胀领域 facade。

## Goals / Non-Goals

**Goals:**

1. 通用全局前置 lifecycle Hook + 目标 `BeforeEnable`/`AfterEnable`，支持 owner 否决其他插件的安装/启用。
2. 交付 `linapro-mail-core`：Connection、Account、`mailcap`、SPI、kind 单例、平台邮件设置管理面、全局冲突检测。
3. 交付 `linapro-mail-smtp` / `imap` / `pop3`（仅 SPI + 硬依赖 mail-core）。
4. 接通 notify `email` 通道 → mail-core 出站，delivery 状态可观测。
5. 支持仅出站 Account；入站无 SPI/未绑定则明确业务错误。

**Non-Goals:**

- 完整邮箱客户端（会话线程、标签、规则、全文搜索产品化 UI）。
- 营销批量发送、退订合规产品、反垃圾完整方案。
- 一期动态 transport 插件与 WASM 全局 Hook（源码优先）。
- 将 SMTP/IMAP 配置表建在宿主 `sys_*` 下。
- 按邮箱厂商拆分协议插件（多 Connection/账号配置，而非多插件）。
- 在本能力内实现全部 webhook 通知通道。

## Decisions

### D1：平台全局 lifecycle Hook 与邮件同能力边界分离交付

全局 Hook 是通用平台机制；邮件只是第一批消费者。规范与实现边界分离：宿主不识别 `smtp` 字符串；冲突规则写在 mail-core。

否决备选：仅 Resolve 冲突（体验差）；各 transport 自管 BeforeInstall 互查（重复、易漏）。

### D2：自管 Hook 与全局 Hook 双轨

| 类型 | 参与者 | 输入 | 语义 |
|------|--------|------|------|
| Target Before* | 仅被操作插件 | 现有 PluginInput 等 | 我能否被装/启/卸 |
| Global Before* | 显式注册全局 handler 的插件 | **含 TargetPluginID** | 系统对 target 的动作我是否允许 |

禁止把 Install/Enable 改为对 `ListSourcePluginLifecycleParticipants()` 全量跑自管 hook。

一期全局集合：`GlobalBeforeInstall`、`GlobalBeforeEnable`（主拦截点）、`GlobalBeforeDisable` / `GlobalBeforeUninstall`（注册面提供，mail-core 按需保护依赖）。目标侧补齐 `BeforeEnable` / `AfterEnable`。After 全局版不做（已有 `plugin.enabled` 等扩展点）。

### D3：Enable 路径必须接入 precondition

源码/动态插件 `UpdateStatus(enabled)` 在改状态前跑 Target `BeforeEnable` + Global `BeforeEnable`；失败不改状态。安装路径聚合 Target `BeforeInstall` + Global `BeforeInstall`。禁用/卸载按需接入 Global Before*，与 force 语义对齐。任一否决/超时/panic → fail-closed。

### D4：plugin-owned + 协议 SPI，Connection 在 core

```
linapro-mail-core
  - Connection 持久化 + CRUD + 密钥引用
  - Account：outboundConnectionId / inboundConnectionId 可选
  - mailcap 公开契约 + spi 注册/Resolve
  - GlobalBeforeEnable 同 kind 冲突检测
  - 平台唯一账号设置 API/页（一期管理面）

linapro-mail-smtp | imap | pop3
  - Register SPI(kind)
  - 接收 core 传入的 Connection endpoint DTO
  - 无 connection 表、无冲突逻辑
```

Connection 公共字段：`kind`、`host`、`port`、`username`、`secretRef`、`tlsMode`、`authMode`、`extraJson`。SPI `Validate`/`Probe` 校验扩展。

表名遵循 `plugin_linapro_mail_core_*`；schema 预留 `tenant_id`，一期管理面按平台配置控制面（`platform_only`）治理。

### D5：同 kind 插件唯一（启用级）

- Connection/Account 数据模型可多实例；**一期 UI** 仅维护平台默认单一账号。
- Transport kind（`smtp` | `imap` | `pop3`）：可服务插件 **0 或 1**；≥2 enabled → 冲突。
- 允许同时安装多个同 kind 实现，禁止同时启用（先禁 A 再启 B 可切换）。
- 解析：`kind → 唯一 enabled provider`；调用方不手选 pluginId（可审计展示）。
- 不同 kind 可并存：`smtp + imap + pop3` 正常。

冲突检测：

1. 启用时：mail-core `GlobalBeforeEnable` 查 SPI 注册表 + enablement。
2. 运行时：`Resolve(kind)` 0/1/≥2 与 storage 同构（安全网）。

### D6：仅出站与入站失败语义

Account 允许 `inboundConnectionId` 为空；入站 API：无绑定或 kind 无 SPI → 稳定业务错误（如 `InboundTransportUnavailable` / `AccountInboundNotConfigured`），不 500、不静默忽略。

### D7：notify email → mail-core

- `notify.Send` 对 `ChannelTypeEmail`：读通道 `config_json.accountId`（或平台默认 Account）→ mail-core 出站。
- mail-core 未安装/未启用/无默认 Account：fail-closed。
- 依赖策略：notify 对 mail-core 为**运行时可选能力**；transport **硬依赖** mail-core。
- DI：`notifycap.ProvideEmailDelivery` 进程内桥接；mail-core 路由注册时提供实现；notify 只依赖 `notifycap.EmailDelivery`，不 import 协议插件。

### D8：密钥与投递状态

密码等经 secretRef / 插件安全配置存储，Connection 表与 GET 投影不落/不回显明文；空密码提交保留已存密钥。发送维护 delivery 状态（pending/succeeded/failed 与现有模型对齐）；一期允许同步发送但必须写 delivery 状态。

### D9：管理面一期形态

平台唯一邮件账号设置页（Card+Form，对齐系统设置/对象存储设置风格）：

- 顶部：必填「账号」「密码」、可选「发件地址」（空则默认 From=账号）；无独立「账号名称」输入。
- 分区：SMTP 出站；可选 IMAP/POP3 入站。
- 操作：测试连接（失败弹窗展示原因）、测试发送（当前表单 SMTP，非静默用已存默认）、测试接收（当前表单入站，无需先保存；仅发信时明确拒绝）、保存设置。
- 插件管理页「管理」按钮经菜单 + `pluginPageMeta.routePath` 跳转设置页。

### D10：插件命名与分发

| 插件 ID | 角色 | distribution |
|---------|------|--------------|
| `linapro-mail-core` | owner | managed |
| `linapro-mail-smtp` | outbound transport | managed |
| `linapro-mail-imap` | inbound transport | managed |
| `linapro-mail-pop3` | inbound transport | managed |

均为 `type: source`。展示元数据使用 `plugin.<id>.name` / `plugin.<id>.description`（禁止顶层 bare `name`/`description` 误写）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 全局 Hook 滥用导致安装变慢/连环 veto | 仅显式注册参与；超时 fail-closed；快速路径文档 |
| mail-core 未启用时冲突检测缺失 | transport 硬依赖 mail-core；Resolve 运行时兜底 |
| Connection 表单难覆盖全部协议扩展 | 公共字段 + `extraJson` + SPI Validate |
| Enable 路径改动影响面 | 单测覆盖 enable/disable；与 BeforeDisable 对称 |
| notify 异步与邮件重试复杂度 | 一期明确 delivery 状态机；重试可迭代 |
| 动态插件无法注册全局 Hook | 一期源码；动态仅消费或后续扩展 |

## Migration

1. 先交付 lifecycle 全局 Hook + BeforeEnable（可独立验证）。
2. 交付 mail-core（表、cap、SPI、设置页、全局冲突）。
3. 交付 smtp → 接通 notify email。
4. 交付 imap/pop3 入站 SPI 与绑定。
5. 无历史邮件数据迁移；既有 notify 仅 inbox，email 由不可用变为可用。
6. 回滚：禁用/卸载邮件插件后 email fail-closed；无全局 Hook 注册者时零行为变化。

## 数据权限 / 缓存 / i18n

- **数据权限**：邮件设置与 Connection/Account 管理作为平台配置控制面；列表/详情/探测遵守数据权限与平台上下文；不得通过探测/错误泄露范围外资源存在性。
- **缓存**：一期无强制跨实例邮件配置缓存契约；若后续引入解析缓存，须定义权威源、写后失效与故障降级。
- **i18n**：插件启用 i18n；菜单、表单、错误、veto reason、API `dc`、插件展示名/描述均需中英语言包；`make i18n.check` 校验展示元数据 key 结构。
