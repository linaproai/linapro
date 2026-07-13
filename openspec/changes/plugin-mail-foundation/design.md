## Context

### 现状

- 宿主 `notify` 域已有 `sys_notify_channel` / `sys_notify_message` / `sys_notify_delivery`，`channel_type` 注释含 `inbox | email | webhook`，常量含 `ChannelTypeEmail`，但 `Send` 仅实现 `inbox`，`email` 走 unsupported。
- 插件可见通知契约为 `notifycap.Service`；动态插件经授权通道发送，规范已写明邮件通道，但无 SMTP 实现。
- 插件生命周期：Install/Disable/Uninstall 的源码 precondition **仅调用目标插件自身** hook；源码 **Enable 路径无 BeforeEnable**，成功后发 `plugin.enabled` 观察型扩展点，**不可 veto**。
- 租户删除等场景已用「全体 lifecycle 参与者」模型，但语义是自管租户钩子，**不是**「否决他人安装/启用」。
- 对象存储采用「同领域唯一活跃 provider」在 **Resolve** 时冲突；邮件需要 **同 transport kind 唯一**，且用户明确要求 **启用时拦截** 并由 **owner 集中实现**。
- AI 领域已验证 plugin-owned 模式：`linapro-ai-core` + `backend/cap` + SPI，宿主不膨胀领域 facade。

### 约束

- 邮件协议与凭据不得进入 `lina-core/pkg/plugin` 领域契约。
- Connection 统一由 `linapro-mail-core` 维护；协议插件只实现 SPI。
- 多 Account（多厂商邮箱）必须支持；同 kind 多插件同时启用禁止。
- notify 只依赖 mail-core，不依赖 smtp/imap/pop3。
- 全局 Hook 必须是通用机制，宿主不识别 `smtp` 字符串。

## Goals / Non-Goals

**Goals:**

1. 提供通用全局前置 lifecycle Hook + 目标插件 `BeforeEnable`/`AfterEnable`，支持 owner 否决其他插件的安装/启用。
2. 交付 `linapro-mail-core`：Connection、Account、`mailcap`、SPI、kind 单例、管理面最小闭环、全局冲突检测。
3. 交付 `linapro-mail-smtp` / `imap` / `pop3` 协议插件（仅 SPI + 依赖 mail-core）。
4. 接通 notify `email` 通道 → mail-core 出站。
5. 支持仅出站 Account；入站无 SPI/未绑定则明确失败。

**Non-Goals:**

- 完整邮箱客户端（会话线程、标签、规则、全文搜索产品化 UI）。
- 营销批量发送、退订合规产品、反垃圾完整方案。
- 一期动态 transport 插件 + WASM 全局 Hook（源码优先；动态消费 `mailcap` 可作为后续）。
- 将 SMTP/IMAP 配置表建在宿主 `sys_*` 下。
- 按邮箱厂商拆分协议插件（Gmail/QQ 等用多 Connection，而非多插件）。
- 在本变更内实现全部 webhook 通知通道。

## Decisions

### D1：平台先增强全局 lifecycle Hook，再消费于邮件

**决策**：变更内包含平台能力 `plugin-lifecycle-global-hooks`，与邮件交付同一 change 分阶段任务，但规范与实现边界分离。

**理由**：无全局 Hook 则冲突检测只能散落在每个 transport 或退回仅 Resolve；用户明确要求 owner 集中检测。

**备选**：仅 Resolve 冲突（实现轻，体验差）；transport 各自 BeforeInstall 调 Assert（重复代码）。已否决。

### D2：自管 Hook 与全局 Hook 双轨，禁止滥用自管 BeforeInstall 广播

**决策**：

| 类型 | 参与者 | 输入 | 语义 |
|------|--------|------|------|
| Target Before* | 仅被操作插件 | 现有 PluginInput 等 | 我能否被装/启/卸 |
| Global Before* | 显式注册了全局 handler 的插件 | **含 TargetPluginID** | 系统对 target 的动作我是否允许 |

**不**把 Install/Enable 改为 `ListSourcePluginLifecycleParticipants()` 全量跑自管 hook（会误触发「以为在装自己」）。

**一期全局 Hook 集合**（克制）：

- `GlobalBeforeInstall`
- `GlobalBeforeEnable`（主拦截点）
- `GlobalBeforeDisable` / `GlobalBeforeUninstall`（可选实现，供 mail-core 保护依赖；注册面先提供）

**目标侧补齐**：`BeforeEnable` / `AfterEnable`。

**After 全局版一期不做**：已有 `plugin.enabled` 等扩展点。

### D3：Enable 路径必须接入 precondition

**决策**：源码/动态插件 `UpdateStatus(enabled)` 在改状态前跑 Target `BeforeEnable` + Global `BeforeEnable`；失败不改状态。

**理由**：当前 enable 几乎直写 status，全局 Hook 否则无挂载点。

### D4：邮件采用 plugin-owned + 协议 SPI，Connection 在 core

**决策**：

```
linapro-mail-core
  - Connection 持久化 + CRUD + 密钥引用 + 表单
  - Account：outboundConnectionId / inboundConnectionId 可选
  - mailcap 公开契约 + spi 注册/Resolve
  - GlobalBeforeEnable 冲突检测

linapro-mail-smtp | imap | pop3
  - Register SPI(kind)
  - 接收 core 传入的 Connection endpoint DTO
  - 无 connection 表、无冲突逻辑
```

Connection 公共字段：`kind`、`host`、`port`、`username`、`secretRef`、`tlsMode`、`authMode`、`extraJson`（一期可先公共字段，SPI `Validate`/`Probe` 校验）。

### D5：多 Account + 同 kind 插件唯一（启用级）

**决策**：

- **Account / Connection**：任意多个（Gmail/QQ/iCloud…）。
- **Transport kind**（`smtp` | `imap` | `pop3`）：可服务插件 **0 或 1**；≥2 enabled → 冲突。
- **允许**同时安装多个同 kind 实现，**禁止同时启用**（先禁 A 再启 B 可切换）。
- 解析：`kind → 唯一 enabled provider`；调用方与 Account **不**手选 pluginId（可审计展示 pluginId）。
- 不同 kind 可并存：`smtp + imap + pop3` 正常。

**冲突检测**：

1. **启用时**：mail-core `GlobalBeforeEnable` 查 SPI 注册表 + enablement。
2. **运行时**：`Resolve(kind)` 0/1/≥2 与 storage 同构（安全网）。

### D6：仅出站与入站失败语义

**决策**：Account 允许 `inboundConnectionId` 为空；入站 API：无绑定或 kind 无 SPI → 稳定业务错误（如 `InboundTransportUnavailable` / `AccountInboundNotConfigured`），不 500、不静默忽略。

### D7：notify email → mail-core

**决策**：

- `notify.Send` 对 `ChannelTypeEmail`：读通道 `config_json.accountId`（或平台默认 Account）→ `mailcap.Send`。
- mail-core 未安装/未启用/无默认 Account：fail-closed，明确错误。
- 依赖策略：notify 对 mail-core 为**运行时可选能力**（未装则 email 通道不可用）；transport 插件 **硬依赖** mail-core。

### D8：密钥与异步

**决策**：密码等经 secretRef / 插件安全配置能力存储，Connection 表不落明文。发送默认异步友好：delivery 先 `pending`，成功/失败更新（可与现有 job 能力对接；一期允许同步发送但必须写 delivery 状态）。

### D9：插件命名与分发

| 插件 ID | 角色 | distribution |
|---------|------|--------------|
| `linapro-mail-core` | owner | managed |
| `linapro-mail-smtp` | outbound transport | managed |
| `linapro-mail-imap` | inbound transport | managed |
| `linapro-mail-pop3` | inbound transport | managed |

均为 `type: source`，参考 `linapro-demo-source` / `linapro-ai-core` 结构。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 全局 Hook 被滥用导致安装变慢/连环 veto | 仅显式注册参与；超时 fail-closed；文档要求快速路径 |
| mail-core 未启用时冲突检测缺失 | transport 硬依赖 mail-core；Resolve 运行时兜底 |
| Connection 表单在 core 难覆盖全部协议扩展 | 公共字段 + `extraJson` + SPI Validate；二期 schema 贡献 |
| Enable 路径改动影响面大 | 单测覆盖 enable/disable；与现有 BeforeDisable 对称 |
| notify 异步与 mail 失败重试复杂度 | 一期明确状态机；重试策略可迭代 |
| 动态插件无法注册全局 Hook | 一期源码；动态仅消费或后续扩展 |

## Migration Plan

1. 先合并/交付 **lifecycle 全局 Hook + BeforeEnable**（平台可独立验证）。
2. 交付 **mail-core**（表结构、cap、SPI、管理面、全局冲突）。
3. 交付 **smtp**（最小可用出站）→ 接通 **notify email**。
4. 交付 **imap/pop3** 入站 SPI 与 Account 绑定。
5. 无历史邮件数据迁移；既有 notify 仅 inbox，email 通道从不可用变为可用。
6. 回滚：禁用/卸载邮件插件后 email 通道恢复 fail-closed；全局 Hook 若回滚需保证无注册参与者时零行为变化。

## Open Questions

1. **Install 是否禁止同 kind 第二套实现**：当前设计为「可装不可同时启」。若产品要求仓库内唯一实现，可在 `GlobalBeforeInstall` 收紧（实现期确认）。
2. **平台默认 Account 创建引导**：装完 smtp 后是否向导一键创建系统发信 Account（体验项，不阻塞契约）。
3. **动态插件一期是否暴露 mail hostServices**：建议源码契约优先，动态列为后续任务可选。
4. **Connection 是否支持租户级**：建议 schema 预留 `tenant_id`；一期可先平台级（`platform_only`），与 AI/storage 插件 scope 对齐后再开租户覆盖。
