## ADDED Requirements

### Requirement: linapro-extid-core 插件拥有外部身份链接存储

`linapro-extid-core` SHALL 是随宿主编译交付的源码插件（`distribution: builtin`），拥有插件私有表 `user_external_identity`（去 `sys_` 前缀），存储 `(provider, subject) → userID` 链接、发起插件 ID 与邮箱快照。该表 SHALL 通过插件安装 SQL 建表、卸载 SQL 清理，插件 MUST 维护自身 `hack/config.yaml` 的 DAO 生成配置，MUST NOT 依赖宿主的 `dao`/`do`/`entity` 生成工件。宿主 SHALL 移除 `sys_user_external_identity` 表及其生成工件与 `013-auth-external-identity.sql`。

#### Scenario: 链接表随插件安装与卸载

- **WHEN** 安装 `linapro-extid-core`
- **THEN** 插件 SHALL 通过安装 SQL 创建 `user_external_identity` 表；卸载时 SHALL 通过卸载 SQL 清理该表

#### Scenario: 宿主不再持有外部身份表

- **WHEN** 检查宿主 `apps/lina-core`
- **THEN** 宿主 SHALL 不含 `sys_user_external_identity` 的 DAO/DO/Entity 与建表 SQL

### Requirement: 插件实现外部身份解析与开户编排

`linapro-extid-core` SHALL 实现宿主 `ExternalIdentityProvider` SPI，提供 `(provider, subject)` 解析为本地 userID 的能力，以及 host-owned 最小权限开户编排。因建号（宿主 `sys_user`）与链接写入（插件 `user_external_identity`）跨模块无法共享单一事务，开户 SHALL 采用"先建号 → 再写链接"顺序，并以链接表 `(provider, subject)` 唯一索引作为权威去重锚点；建号 SHALL 幂等，链接写入失败或并发冲突时不重复建立有效链接。开户策略（含邮箱冲突判定）SHALL 闭环在插件内，宿主不再承载具体开户策略。

#### Scenario: 已链接身份解析为本地用户

- **WHEN** provider 收到已存在 `(provider, subject)` 链接的解析请求
- **THEN** 插件 SHALL 返回对应 userID，MUST NOT 重复建号

#### Scenario: 未链接身份按策略开户

- **WHEN** provider 收到未链接身份且调用方允许自动开户
- **THEN** 插件 SHALL 先通过宿主最小权限建号能力建立账号，再写入 `(provider, subject)` 链接记录，链接写入以唯一索引保证同一外部身份最终只有一条有效链接

#### Scenario: 同邮箱已存在时拒绝静默绑定

- **WHEN** 未链接身份携带的邮箱已被某个启用的本地账号使用
- **THEN** 插件 SHALL 返回邮箱冲突错误，MUST NOT 自动把外部身份链接到该已有账号

### Requirement: 支持无邮箱开户

当外部 IdP 不提供邮箱（如微信）或邮箱可选（如 QQ）时，`linapro-extid-core` SHALL 支持在无邮箱情况下完成开户，采用派生用户名或其他稳定锚点，MUST NOT 因缺少邮箱直接拒绝开户。无邮箱开户 SHALL 不触发同邮箱冲突判定。

#### Scenario: 无邮箱身份成功开户

- **WHEN** 未链接身份不携带邮箱且调用方允许自动开户
- **THEN** 插件 SHALL 使用派生用户名建立最小权限账号并写入链接记录，MUST NOT 返回邮箱无效错误

### Requirement: 已登录用户绑定与解绑外部身份

`linapro-extid-core` SHALL 提供 `BindVerifiedIdentity` 能力，允许已登录用户把已验证的外部身份绑定到当前会话用户，以及解绑与列举当前用户已绑定身份。绑定/解绑动作 SHALL 仅作用于当前会话用户自身的链接记录，MUST NOT 通过 provider、subject 或 userID 参数绕过当前用户边界操作他人链接。已被其他用户绑定的 `(provider, subject)` SHALL 拒绝重复绑定。

#### Scenario: 已登录用户绑定外部身份

- **WHEN** 已登录用户提交一个未被任何账号占用的已验证外部身份
- **THEN** 插件 SHALL 把该 `(provider, subject)` 链接到当前会话用户

#### Scenario: 绑定动作不越权

- **WHEN** 绑定/解绑请求携带的目标用户不是当前会话用户
- **THEN** 插件 SHALL 拒绝该操作，MUST NOT 修改他人链接记录

#### Scenario: 重复绑定已占用身份被拒绝

- **WHEN** 用户尝试绑定一个已被其他账号链接的 `(provider, subject)`
- **THEN** 插件 SHALL 拒绝绑定并返回冲突错误

### Requirement: 并发开户与链接的正确性

权威去重锚点是 `user_external_identity` 的 `(provider, subject)` 唯一索引（`sys_user` 仅 `username` 唯一、`email` 无唯一索引，因此不能以 email 唯一性做去重）。当同一 `(provider, subject)` 的多个登录请求并发到达且身份未链接时，为避免竞态产生游离 `sys_user`，实现 SHOULD 先以 `(provider, subject)` 抢占链接（如先插入占位链接或使用唯一索引冲突语义）再补建号，或在补偿路径中容忍未链接的建号孤儿。系统 SHALL 保证同一 `(provider, subject)` 最终只有**一条有效链接、指向一个账号**；唯一索引冲突 SHALL 被捕获并转为复用已链接账号或返回冲突错误，MUST NOT 冒泡为 500。无邮箱开户的用户名 anchor 派生 SHALL 确定性可复现（同一 anchor 复用同一账号），MUST NOT 因数字后缀去重导致同一外部身份重复建号。并发竞态下可能出现的未链接建号孤儿 SHALL 按插件禁用与卸载数据处置规则容忍（不影响有效链接唯一性）。

#### Scenario: 并发未链接登录只产生一条有效链接

- **WHEN** 两个携带相同 `(provider, subject)` 的登录请求并发触发自动开户
- **THEN** 系统 SHALL 最终只保留一条 `(provider, subject)` 有效链接指向同一账号，冲突请求复用该账号，MUST NOT 返回 500；竞态产生的未链接账号视为可容忍孤儿

#### Scenario: 链接写入失败后幂等复用

- **WHEN** 建号成功但链接写入失败，随后同一 `(provider, subject)` 再次登录
- **THEN** 系统 SHALL 通过确定性 anchor 复用已建账号完成链接，MUST NOT 对同一外部身份重复建立有效链接

### Requirement: 插件禁用与卸载的数据处置

`linapro-extid-core` 被**禁用**时，`user_external_identity` 表 SHALL 保留，external login 走 fail-closed，重新启用后链接 SHALL 恢复可用。被**卸载**时，卸载 SQL SHALL 清理 `user_external_identity` 表，但 MUST NOT 级联删除宿主 `sys_user` 中经外部登录建立的账号——这些账号 SHALL 作为无外部链接的孤儿账号保留，可由管理员重置密码后以密码登录。

#### Scenario: 禁用保留链接数据

- **WHEN** 插件被禁用后重新启用
- **THEN** 之前的 `(provider, subject)` 链接 SHALL 仍然有效，用户可再次外部登录

#### Scenario: 卸载断开链接但保留账号

- **WHEN** 插件被卸载
- **THEN** 卸载 SQL SHALL 删除 `user_external_identity` 表，宿主 `sys_user` 中对应账号 SHALL 保留且不被级联删除

### Requirement: 外部身份链接数据权限边界

外部身份链接记录 SHALL 作为用户自隔离资源处理：登录解析走 `(provider, subject)` 唯一键，不通过关联名称、数量或候选项泄露范围外账号存在性；绑定/解绑/列举仅返回当前会话用户自身的链接。该例外 SHALL 在设计与审查结论中记录权威边界与拒绝策略。

#### Scenario: 列举只返回自身链接

- **WHEN** 用户请求列举已绑定外部身份
- **THEN** 插件 SHALL 仅返回当前会话用户自身的链接记录
