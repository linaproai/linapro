## ADDED Requirements

### Requirement: linapro-extlogin-core 插件拥有外部身份链接存储

`linapro-extlogin-core` SHALL 是可按需安装的源码插件（`type: source`，`distribution: managed`），拥有插件私有表 `plugin_linapro_extlogin_core_user_external_identity`（`plugin_linapro_extlogin_core_*` 前缀），存储 `(provider, subject) → userID` 链接、发起插件 ID 与邮箱/档案快照。该表 SHALL 通过插件安装 SQL 建表、卸载 SQL 清理，插件 MUST 维护自身 `hack/config.yaml` 的 DAO 生成配置，MUST NOT 依赖宿主的 `dao`/`do`/`entity` 生成工件。宿主 SHALL 不含 `sys_user_external_identity` 表及其生成工件。

#### Scenario: 链接表随插件安装与卸载

- **WHEN** 安装 `linapro-extlogin-core`
- **THEN** 插件 SHALL 通过安装 SQL 创建 `plugin_linapro_extlogin_core_user_external_identity` 表；卸载时 SHALL 通过卸载 SQL 清理该表

#### Scenario: 宿主不再持有外部身份表

- **WHEN** 检查宿主 `apps/lina-core`
- **THEN** 宿主 SHALL 不含 `sys_user_external_identity` 的 DAO/DO/Entity 与建表 SQL

### Requirement: 插件实现外部身份解析与开户编排

`linapro-extlogin-core` SHALL 实现宿主 `ExternalIdentityProvider` SPI，提供 `(provider, subject)` 解析为本地 userID 的能力，以及 host-owned 最小权限开户编排。因建号（宿主 `sys_user`）与链接写入（插件表）跨模块无法共享单一事务，开户 SHALL 采用"先建号 → 再写链接"顺序，并以链接表 `(provider, subject)` 唯一索引作为权威去重锚点；建号 SHALL 幂等，链接写入失败或并发冲突时不重复建立有效链接。开户策略（含邮箱冲突判定）SHALL 闭环在插件内，宿主不再承载具体开户策略。

#### Scenario: 已链接身份解析为本地用户

- **WHEN** provider 收到已存在 `(provider, subject)` 链接的解析请求
- **THEN** 插件 SHALL 返回对应 userID，MUST NOT 重复建号

#### Scenario: 未链接身份按策略开户

- **WHEN** provider 收到未链接身份且调用方允许自动开户
- **THEN** 插件 SHALL 先通过宿主最小权限建号能力建立账号，再写入 `(provider, subject)` 链接记录

#### Scenario: 同邮箱已存在时拒绝静默绑定

- **WHEN** 未链接身份携带的邮箱已被某个启用的本地账号使用
- **THEN** 插件 SHALL 返回邮箱冲突错误，MUST NOT 自动把外部身份链接到该已有账号

### Requirement: 支持无邮箱开户

当外部 IdP 不提供邮箱时，`linapro-extlogin-core` SHALL 支持在无邮箱情况下完成开户，采用确定性用户名 anchor，MUST NOT 因缺少邮箱直接拒绝开户。无邮箱开户 SHALL 不触发同邮箱冲突判定。

#### Scenario: 无邮箱身份成功开户

- **WHEN** 未链接身份不携带邮箱且调用方允许自动开户
- **THEN** 插件 SHALL 使用派生用户名建立最小权限账号并写入链接记录，MUST NOT 返回邮箱无效错误

### Requirement: 已登录用户绑定与解绑外部身份

插件 SHALL 提供绑定/解绑/列举能力；绑定公开路径 SHALL 基于已验证 ticket。动作仅作用于当前会话用户，已被其他用户占用的 `(provider, subject)` SHALL 拒绝重复绑定。

#### Scenario: 绑定动作不越权

- **WHEN** 绑定/解绑请求目标用户不是当前会话用户
- **THEN** 插件 SHALL 拒绝该操作

#### Scenario: 重复绑定已占用身份被拒绝

- **WHEN** 用户尝试绑定一个已被其他账号链接的 `(provider, subject)`
- **THEN** 插件 SHALL 拒绝绑定并返回冲突错误

### Requirement: 并发开户与链接的正确性

权威去重锚点是链接表 `(provider, subject)` 唯一索引。系统 SHALL 保证同一 `(provider, subject)` 最终只有一条有效链接；唯一索引冲突 MUST NOT 冒泡为 500。无邮箱 anchor 派生 SHALL 确定性可复现。

#### Scenario: 并发未链接登录只产生一条有效链接

- **WHEN** 两个携带相同 `(provider, subject)` 的登录请求并发触发自动开户
- **THEN** 系统 SHALL 最终只保留一条有效链接，冲突请求复用该账号，MUST NOT 返回 500

### Requirement: 插件禁用与卸载的数据处置

禁用时链接表 SHALL 保留，external login fail-closed，重新启用后链接恢复。卸载时 SHALL 清理链接表，MUST NOT 级联删除宿主 `sys_user` 中经外部登录建立的账号。

#### Scenario: 禁用保留链接数据

- **WHEN** 插件被禁用后重新启用
- **THEN** 之前的 `(provider, subject)` 链接 SHALL 仍然有效

#### Scenario: 卸载断开链接但保留账号

- **WHEN** 插件被卸载
- **THEN** 卸载 SQL SHALL 删除链接表，宿主 `sys_user` 对应账号 SHALL 保留

### Requirement: 外部身份链接数据权限边界

外部身份链接记录 SHALL 作为用户自隔离资源：登录解析走 `(provider, subject)` 唯一键，不泄露范围外账号存在性；绑定/解绑/列举仅返回当前会话用户自身链接。

#### Scenario: 列举只返回自身链接

- **WHEN** 用户请求列举已绑定外部身份
- **THEN** 插件 SHALL 仅返回当前会话用户自身的链接记录
