## ADDED Requirements

### Requirement: linapro-extlogin-core 为 managed 外部身份领域 owner

`linapro-extlogin-core` SHALL 使用固定插件 ID `linapro-extlogin-core`，`type: source`，`distribution: managed`。宿主启动 MUST NOT 因 builtin 引导自动安装或启用该插件。该插件 SHALL 拥有外部身份链接存储与 `extidcap` 领域契约实现。

#### Scenario: 未安装时宿主不强制装入

- **WHEN** 宿主启动且用户从未安装 `linapro-extlogin-core`
- **THEN** 系统 MUST NOT 自动安装该插件，且外部身份领域能力不可用

#### Scenario: 禁用保留数据

- **WHEN** 已安装的 `linapro-extlogin-core` 被禁用
- **THEN** 链接表数据 SHALL 保留，外部登录 fail-closed，重新启用后链接恢复可用

### Requirement: 完整 extidcap 领域契约

`linapro-extlogin-core` SHALL 在 `backend/cap/extidcap` 发布 plugin-owned 领域契约：以单一宽入口 `Service` 聚合 `TicketService`、`LoginService`、`LinkageService`、`ProviderService` 子面，至少覆盖 Verified ticket 签发/消费、LoginPrepare、BindByTicket、Unbind、ListByUser、GetLinkage、ListProviders 与 provision policy 查询。Catalog 注册与 SPA handoff SHALL 通过独立 `CatalogService` / `HandoffService` 门面发布。契约 MUST NOT 通过插件可见 `AdminService` 目录表达风险边界；未发布能力 MUST NOT 以 silent no-op 成功占位。

#### Scenario: 协议插件可依赖 cap 契约

- **WHEN** 协议源码插件需要签发 ticket 或查询链接
- **THEN** 它 SHALL 通过 `extidcap` 或宿主 SPI/HTTP 受治理入口访问，MUST NOT import core 的 `internal` 包

#### Scenario: 消费方依赖宽入口子面

- **WHEN** 源码插件或 HTTP 控制器需要外部身份领域操作
- **THEN** 它 SHALL 依赖 `extidcap.Service`（或其子面），并通过 `Ticket()` / `Login()` / `Linkage()` / `Providers()` 访问对应能力

### Requirement: 绑定必须基于已验证 ticket

公开绑定 API SHALL 仅接受 ticket 标识，MUST NOT 接受客户端自报的裸 `provider`+`subject` 作为唯一证明完成绑定。

#### Scenario: 裸 subject 绑定被拒绝

- **WHEN** 客户端对绑定接口提交无 ticket 的 provider/subject
- **THEN** 系统 SHALL 拒绝绑定

#### Scenario: 有效 ticket 绑定成功

- **WHEN** 已登录用户提交未过期、未消费且 subject 未被他人占用的 ticket
- **THEN** 系统 SHALL 将身份链接到当前会话用户并消费 ticket

### Requirement: VerifiedIdentity 扩展模型

领域模型 SHALL 支持 `Provider`、`Subject`、`SubjectKind`、可选 `SecondarySubjects`、`AppContext`、邮箱/手机号快照与展示字段，以便微信/QQ 等协议扩展。权威链接键仍为 `(provider, subject)`。

#### Scenario: 不同 provider 同 subject 独立

- **WHEN** 两个 provider 使用相同 subject 字符串
- **THEN** 系统 SHALL 视为两个独立外部身份

### Requirement: 未安装时降级

当 `linapro-extlogin-core` 未安装或未启用时，外部登录与绑定 SHALL fail-closed 或不可达；依赖该能力的 UI SHALL 隐藏而非报错空白；密码登录与其它宿主能力 MUST NOT 受影响。

#### Scenario: 无 core 时密码登录仍可用

- **WHEN** `linapro-extlogin-core` 未安装
- **THEN** 用户仍能使用宿主密码登录完成会话
