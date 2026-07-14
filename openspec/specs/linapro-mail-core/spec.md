# linapro-mail-core Specification

## Purpose
TBD - created by archiving change plugin-mail-foundation. Update Purpose after archive.
## Requirements
### Requirement: 邮件领域由 linapro-mail-core 作为 plugin-owned owner

系统 SHALL 将邮件领域公开契约、Connection/Account 模型、transport SPI、kind 解析与管理面实现归属官方源码插件 `linapro-mail-core`。宿主 `lina-core` MUST NOT 在 `pkg/plugin/capability` 下新增邮件协议实现或 Connection 存储模型。公开契约 MUST 位于该插件 `backend/cap/mailcap`（或等价包路径）。

#### Scenario: 消费方依赖 owner 契约

- **WHEN** 其他源码插件需要发送或接收邮件
- **THEN** 生产代码 MUST 通过 `linapro-mail-core` 发布的 `mailcap` 契约访问能力
- **AND** MUST 在 `plugin.yaml` 的 `dependencies.plugins` 中声明对 `linapro-mail-core` 的依赖

### Requirement: Connection 统一由 mail-core 维护

系统 SHALL 在 `linapro-mail-core` 中持久化并管理邮件 Connection（协议连接配置），包括创建、更新、删除、列表、密钥引用与连通性探测入口。SMTP/IMAP/POP3 协议插件 MUST NOT 自建 Connection 业务表作为权威源。Connection MUST 至少包含稳定 ID、显示名、`kind`、主机与端口、用户名、密钥引用、TLS/认证相关字段及可选扩展 JSON。

#### Scenario: 协议插件使用 core Connection

- **WHEN** 出站或入站 SPI 被调用
- **THEN** mail-core MUST 将对应 Connection 解析为 endpoint DTO 传给协议插件
- **AND** 协议插件 MUST NOT 要求调用方直接持有其私有连接配置表

#### Scenario: Connection 管理面在 core

- **WHEN** 管理员配置邮件服务器连接
- **THEN** 配置表单与 API MUST 由 `linapro-mail-core` 提供
- **AND** 密钥明文 MUST NOT 写入可导出的普通配置字段

### Requirement: Account 支持出站入站绑定（管理面单账号）

系统 SHALL 提供 Account 模型表示业务邮箱身份。一期管理面 MUST 仅维护**一个**平台默认 Account：管理员在设置页直接填写账号与 SMTP/IMAP/POP3 连接信息并保存/测试；MUST NOT 在默认管理页暴露多账号列表或新增第二个账号。数据模型仍允许 Connection 出站/入站绑定；允许仅出站（仅 outbound）或仅入站（仅 inbound）。调用发送时使用 outbound；调用接收/同步时使用 inbound。

#### Scenario: 单账号设置页保存

- **WHEN** 管理员在邮件设置页填写账号、密码、SMTP 与可选 IMAP/POP3 入站信息并保存
- **THEN** 系统 MUST 将结果收敛为唯一默认 Account 及其绑定 Connection
- **AND** MUST NOT 要求用户分别管理多个 Account 条目
- **AND** 管理页 MUST NOT 要求填写独立的账号显示名称字段；内部 Account/Connection 名称可由账号或发件地址派生

#### Scenario: 可选发件地址默认使用账号

- **WHEN** 管理员保存平台邮件设置且「发件地址」留空
- **THEN** 系统 MUST 将默认 From 解析为「账号」字段值（通常为邮箱登录地址）
- **AND** 出站发送在未显式指定 From 时 MUST 使用该解析结果

#### Scenario: 仅出站 Account

- **WHEN** Account 仅绑定 outbound Connection 且未绑定 inbound
- **THEN** 出站发送在协议可用时 MUST 可成功编排
- **AND** 入站相关方法 MUST 返回明确业务错误，指示未配置入站或入站不可用

### Requirement: Transport kind 单例与解析

系统 SHALL 以 transport `kind`（至少包括 `smtp`、`imap`、`pop3`）为协议能力槽。对每个 kind，同时处于可服务（已安装且启用且已注册 SPI）状态的协议插件数量 MUST 为 0 或 1。`Resolve(kind)`：0 个则该 kind 不可用；1 个则选用该插件；≥2 个 MUST 返回冲突错误且 MUST NOT 静默挑选。出站与入站 SPI MUST 分方向注册，不得强迫 SMTP 实现入站接口。

#### Scenario: 同 kind 多启用在运行时冲突

- **WHEN** 两个插件均注册并处于启用状态且 kind 均为 `smtp`
- **THEN** `Resolve(smtp)` MUST 失败并返回稳定冲突错误
- **AND** MUST 在错误诊断中包含冲突插件 ID 列表（或等价信息）

#### Scenario: 不同 kind 可并存

- **WHEN** `linapro-mail-smtp` 与 `linapro-mail-imap` 同时启用
- **THEN** 系统 MUST 允许该状态
- **AND** 出站与入站解析 MUST 分别命中对应 kind

### Requirement: mail-core 通过全局 Hook 执行启用冲突检测

系统 SHALL 由 `linapro-mail-core` 注册全局前置生命周期回调（至少 `GlobalBeforeEnable`），在目标插件启用前检测：若目标插件已在邮件 transport SPI 注册表中登记，且其 kind 已存在其他已启用 provider，则 MUST 否决该启用。冲突检测逻辑 MUST NOT 要求每个协议插件各自实现。协议插件在冲突场景下 MUST 无需注册全局 Hook。

#### Scenario: 启用第二个 smtp 被否决

- **WHEN** 已启用 `linapro-mail-smtp`（kind=smtp），管理员尝试启用另一同样注册 kind=smtp 的协议插件
- **THEN** `linapro-mail-core` 的全局 `BeforeEnable` MUST 否决
- **AND** 目标插件 MUST 保持未启用
- **AND** 否决 reason MUST 可本地化且指出 kind 冲突

#### Scenario: 启用非邮件插件放行

- **WHEN** 管理员启用与邮件 transport 注册表无关的插件
- **THEN** mail-core 全局回调 MUST 快速放行
- **AND** MUST NOT 影响该插件启用

### Requirement: mailcap 公开发送与探测能力

系统 SHALL 通过 `mailcap` 提供至少：按 Account（或显式默认 Account 解析）发送邮件、Connection/Account 可见性范围内的连通性探测、以及入站拉取或同步的最小方法集（入站可在协议插件就绪后启用）。未指定 Account 时 MUST 按平台（及若支持的租户）默认 Account 解析；无默认则 fail-closed。

#### Scenario: 显式 Account 发送

- **WHEN** 调用方指定合法且可见的 Account 发送邮件
- **THEN** 系统 MUST 使用该 Account 的 outbound Connection 与对应 kind 的唯一 SPI 发送
- **AND** MUST NOT 要求调用方传入协议插件 ID

#### Scenario: 无默认 Account

- **WHEN** 调用方未指定 Account 且不存在可用默认 Account
- **THEN** 系统 MUST 拒绝发送并返回明确错误

### Requirement: 管理面与资源边界

系统 SHALL 为 `linapro-mail-core` 提供管理 API 与前端页面，并维护安装 SQL、卸载 SQL 与 i18n 资源。插件 MUST 启用 i18n。默认管理页 MUST 采用与宿主/其他插件设置页一致的 Card+Form 布局，提供单账号字段的加载、保存与连通性测试入口。顶部表单字段顺序 MUST 为：必填「账号」、必填「密码」、可选「发件地址」；MUST NOT 展示独立的「账号名称」输入域。密钥明文 MUST NOT 在 GET 投影中回显；空密码提交 MUST 保留已保存密钥。管理操作 MUST 遵守数据权限与审计要求；列表与探测 MUST 避免 N+1 查询。

#### Scenario: 设置页测试与保存

- **WHEN** 管理员在邮件设置页点击测试连接或保存
- **THEN** 系统 MUST 使用当前表单值（密码留空则用已存密钥）执行 SMTP 及已配置入站协议的探测或持久化
- **AND** 页面布局 MUST 与同类设置页（如对象存储设置）在结构上保持一致（说明 Alert、表单项、测试/保存按钮）
- **AND** 页面 MUST 展示「账号」「密码」「发件地址」三字段于 SMTP/收信分区之前，且 MUST NOT 展示「账号名称」

#### Scenario: 测试连接失败以弹窗展示原因

- **WHEN** 管理员点击测试连接且探测失败
- **THEN** 页面 MUST 以弹窗（或等价模态提示）展示具体失败原因，便于排查
- **AND** MUST NOT 仅依赖页面顶部常驻 Alert 区域展示该错误详情

#### Scenario: 使用当前表单配置测试发送

- **WHEN** 管理员在设置页打开「测试发送」并提交收件人与正文
- **THEN** 系统 MUST 使用**当前表单中的 SMTP 配置**（含未保存的主机/端口/账号/密码/TLS 与发件地址）发送一封测试邮件
- **AND** MUST NOT 静默改用已持久化的系统默认 Account/SMTP 连接（当表单与已存配置不一致时）
- **AND** 发送结果（成功或失败原因）MUST 以可观察提示反馈给管理员
- **AND** 页面按钮文案 MUST 为「测试发送」（不得再展示「测试邮件」作为该入口文案）

#### Scenario: 使用当前表单配置测试接收

- **WHEN** 管理员在设置页已选择 IMAP 或 POP3 收信协议，并点击「测试接收」
- **THEN** 系统 MUST 使用**当前表单中的收信配置**（含未保存的协议/主机/端口/账号/密码/TLS）探测入站可达与登录能力
- **AND** MUST NOT 静默改用已持久化的系统默认入站连接（当表单与已存配置不一致时）
- **AND** MUST NOT 要求先保存设置即可执行该探测
- **AND** 探测结果（成功或失败原因）MUST 以可观察提示反馈给管理员
- **WHEN** 收信协议为「无（仅发信）」时管理员点击「测试接收」
- **THEN** 系统 MUST 明确拒绝并提示需先配置收信协议，不得伪装成功

#### Scenario: 卸载清理

- **WHEN** 卸载 `linapro-mail-core`
- **THEN** 卸载 SQL MUST 清理插件自有 Connection/Account 等表（按卸载策略）
- **AND** 依赖该插件的协议插件 MUST 因依赖治理而不可继续作为可服务 transport

