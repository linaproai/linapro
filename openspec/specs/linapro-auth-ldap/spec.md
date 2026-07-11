# linapro-auth-ldap Specification

## Purpose
TBD - created by archiving change add-ldap-auth-plugin. Update Purpose after archive.
## Requirements
### Requirement: 插件标识、分发与依赖

系统 SHALL 提供 managed 源码插件 `linapro-auth-ldap`。MUST 声明对 `linapro-extlogin-core` 的版本依赖（`>=0.1.0`）。未满足依赖时 MUST 无法成功启用。MUST NOT 作为宿主 builtin 强制安装。

#### Scenario: 缺少 core 时无法启用

- **WHEN** 管理员尝试启用 `linapro-auth-ldap` 且 `linapro-extlogin-core` 未启用
- **THEN** 宿主依赖治理 MUST 阻止启用

#### Scenario: 不影响密码登录

- **WHEN** 未安装本插件
- **THEN** 宿主本地密码登录 MUST 仍可用

### Requirement: Provider 与 subject

插件 SHALL 声明 provider `ldap:default`。权威键 MUST 为 `(provider, subject)`。subject MUST 来自目录属性（可配置），MUST NOT 使用密码，MUST NOT 默认使用可变邮箱作为解析键。

#### Scenario: 登录提交 provider

- **WHEN** bind 成功并调用 `LoginByVerifiedIdentity`
- **THEN** `provider` MUST 为 `ldap:default` 且 subject 非空

### Requirement: LDAP bind 登录 API

插件 SHALL 提供公开 `POST /portal/linapro-auth-ldap/login`，接受用户名与密码，在服务端完成 LDAP 连接与 bind 验真，成功后经宿主外部登录 seam 与 extlogin-core handoff 返回 **handoff 码**（JSON）。响应 MUST NOT 包含 access/refresh JWT。密码 MUST NOT 被持久化或写入日志。

#### Scenario: 成功返回 handoff

- **WHEN** 用户提交正确目录凭证且身份允许登录
- **THEN** 响应 MUST 包含 handoff 且 MUST NOT 包含 accessToken/refreshToken

#### Scenario: 失败统一对外语义

- **WHEN** 用户不存在、密码错误或 bind 失败
- **THEN** 插件 MUST 返回统一业务错误（不区分是否存在账号的细节），且 MUST NOT 在错误消息中包含密码

#### Scenario: 未配置 fail-closed

- **WHEN** host/base 等必要配置缺失
- **THEN** 登录 API MUST 拒绝请求且不尝试明文不可信连接

### Requirement: TLS

生产配置 MUST 使用 `ldaps` 或 `starttls`。`plain` 仅当 host 为 localhost/127.0.0.1 时允许。

#### Scenario: 非本机明文拒绝

- **WHEN** tls_mode=plain 且 host 非本机
- **THEN** 保存配置或登录 MUST 失败并给出明确错误

### Requirement: 自动开户默认关闭

AllowAutoProvision MUST 默认关闭；仅管理员显式开启后登录请求才允许 `AllowAutoProvision=true`。

#### Scenario: 默认不自动开户

- **WHEN** 新安装且未改自动开户，未链接身份登录
- **THEN** MUST 不创建新本地用户

### Requirement: 管理设置与入口

插件 SHALL 提供受权限保护的设置 API/页面与 `auth.login.after` 登录入口弹层。插件禁用或依赖不满足时入口 MUST 不展示。菜单 MUST 挂在 `plugin:linapro-extlogin-core:auth-login` 下。

#### Scenario: 禁用后入口消失

- **WHEN** 插件被禁用
- **THEN** 登录页 MUST NOT 展示目录登录入口

#### Scenario: 登录入口使用工作台统一认证样式

- **WHEN** 插件在登录页渲染目录登录入口并打开凭证弹层
- **THEN** 入口按钮、弹层、表单与校验反馈 MUST 复用工作台统一前端组件和主题`token`，且在桌面与移动视口中 MUST NOT 出现溢出或元素重叠

### Requirement: 协议边界与 i18n

插件 MUST NOT 持有外部身份链接表，MUST NOT 铸造宿主 JWT，MUST NOT 修改宿主密码校验路径。MUST 提供 en-US/zh-CN 双语资源与 README。

#### Scenario: 无链接表

- **WHEN** 审查插件 SQL/DAO
- **THEN** MUST NOT 新增外部身份链接业务表

