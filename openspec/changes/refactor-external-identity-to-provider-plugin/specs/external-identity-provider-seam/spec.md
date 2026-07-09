## ADDED Requirements

### Requirement: 宿主外部身份 Provider SPI 契约

宿主 SHALL 在 `pkg/plugin/capability/authcap/externallogin/externalidentityspi` 定义 `ExternalIdentityProvider` 能力接缝，用于把"外部身份 → 本地用户"的解析、开户与绑定编排委托给源码插件实现。宿主 SHALL 持有 provider manager 实例并在启动装配阶段注入 auth 服务，其模式 MUST 与既有 `orgspi`/`tenantspi` 的 provider 管理与注入方式一致。SPI 契约 MUST 仅承载能力接缝方法，不得暴露宿主 auth 的 token、session、租户或私有配置内部结构。

#### Scenario: SPI 契约由宿主定义并注入

- **WHEN** 宿主启动装配 auth 服务
- **THEN** 宿主 SHALL 通过构造函数注入 `ExternalIdentityProvider`，且注入路径与 `orgspi`/`tenantspi` 的 provider manager 注入方式一致

#### Scenario: SPI 契约不泄露宿主内部结构

- **WHEN** 插件实现 `ExternalIdentityProvider`
- **THEN** 契约入参与返回值 SHALL 仅使用契约自有 DTO 或值对象，MUST NOT 出现宿主 `dao`/`do`/`entity`、token 铸造器、会话存储或租户内部类型

### Requirement: Provider 缺失时外部登录 fail-closed

当没有源码插件注册 `ExternalIdentityProvider` 时，宿主 `externallogin.Service.LoginByVerifiedIdentity` SHALL 走 fail-closed 路径，返回 not-provisioned 语义错误，MUST NOT 创建任何本地账号或会话。该降级行为 MUST 与 tenant/org 能力缺失时返回中性值的处理保持语义一致（不抛 500、不空白）。

#### Scenario: 无 provider 时登录被拒绝

- **WHEN** 宿主未注入任何 `ExternalIdentityProvider` 且插件调用 `LoginByVerifiedIdentity`
- **THEN** 宿主 SHALL 返回 not-provisioned 错误，MUST NOT 铸造 token、创建会话或建号

#### Scenario: provider 存在但身份未链接时不泄露

- **WHEN** provider 已注入但 `(provider, subject)` 未链接且未开启自动开户
- **THEN** 宿主 SHALL 返回统一的 not-provisioned 错误，MUST NOT 泄露捕获邮箱是否已存在于其他账号

### Requirement: Token 与会话铸造保留在宿主

外部登录路径中的 IP 黑名单检查、禁用账号检查、租户解析、多租户 pre-token 生成、access/refresh token 铸造、会话持久化、登录时间更新与登录成功/失败 hook 派发 SHALL 全部保留在宿主 auth 服务内，MUST NOT 下放给 provider 插件。Provider 的职责边界 SHALL 仅限于"外部身份解析为本地 userID"与"host-owned 开户/绑定编排"。

#### Scenario: provider 只返回 userID，宿主铸 token

- **WHEN** provider 成功把 `(provider, subject)` 解析为本地 userID
- **THEN** 宿主 SHALL 使用该 userID 继续执行租户解析与 token 铸造，且 token/session 逻辑与密码登录共用同一 seam 保持行为兼容

#### Scenario: provider 不参与 token 铸造

- **WHEN** 审查 `ExternalIdentityProvider` 契约方法
- **THEN** 契约 SHALL 不含任何铸造 token、创建 session 或选择租户的方法

### Requirement: 宿主提供插件可调的最小权限建号能力

宿主 SHALL 通过插件可消费的能力（`usercap` 宽接口的最小权限外部建号方法）让 `ExternalIdentityProvider` 实现方能反向发起 host-owned 建号，区别于操作员建号（`usercap.Create`，带租户/角色/创建边界校验）。该建号能力 SHALL 接受可选用户名 anchor：当邮箱为空时用 anchor 派生用户名，MUST NOT 因缺邮箱直接拒绝；邮箱非空时按邮箱派生。建号 SHALL 幂等：以 `(provider, subject)` 链接为权威去重锚点，无邮箱 anchor 派生用户名须确定性可复现，MUST NOT 假设 `sys_user` email 唯一（email 无唯一索引）。建号 shape（用户名派生、不可用密码、最小权限）SHALL 保留在宿主用户域。

#### Scenario: 插件通过注入能力建号

- **WHEN** provider 插件需要为未链接身份建号
- **THEN** 插件 SHALL 通过注入的 `usercap` 最小权限外部建号方法发起，宿主内部委托建号 shape 逻辑，MUST NOT 要求插件直接操作 `sys_user`

#### Scenario: 无邮箱时用 anchor 建号

- **WHEN** 建号请求邮箱为空但提供了用户名 anchor
- **THEN** 宿主 SHALL 用 anchor 派生用户名完成建号，MUST NOT 返回邮箱无效错误

### Requirement: 跨 provider 的 subject 不冲突

不同 provider 下的相同 subject 值 SHALL 被视为不同外部身份。解析、开户与绑定 SHALL 以 `(provider, subject)` 组合为唯一键，MUST NOT 仅凭 subject 相等把两个 provider 的身份解析到同一账号。

#### Scenario: 同 subject 不同 provider 互不干扰

- **WHEN** provider A 与 provider B 各有一条 subject 值相同的身份
- **THEN** 系统 SHALL 把它们解析为两个独立外部身份，MUST NOT 交叉链接到对方账号

### Requirement: Provider ownership 治理不变

宿主 SHALL 继续基于 `ProvideExternalIdentity` 声明的 provider 归属，从源码插件作用域能力目录盖章调用方插件身份，MUST NOT 信任插件自述的 provider 归属。请求非本插件拥有的 provider、或来自已禁用插件的请求 SHALL 被拒绝。该治理边界在 provider SPI 重构后 MUST 保持不变。

#### Scenario: 请求未拥有的 provider 被拒绝

- **WHEN** 插件 A 请求登录时声明的 provider 归属于插件 B
- **THEN** 宿主 SHALL 拒绝该请求，MUST NOT 委托 provider 解析或建号
