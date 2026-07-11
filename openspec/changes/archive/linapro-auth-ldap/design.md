# Design

## 插件与 provider

- ID：`linapro-auth-ldap`；`provider = "ldap:default"`；`SubjectKind = custom`。
- `ProvideExternalIdentity("ldap:default")` + catalog；硬依赖 `linapro-extlogin-core`。

## 验真流程

用户弹层提交 username + password → TLS 连接 → Search（服务账号 bind + filter）或 Template DN → 用户 DN + password bind → 读 subject/email/displayName → `LoginByVerifiedIdentity` → `CreateLoginHandoffFromHost` → JSON `{handoff}`。

## 配置与安全

- sys_config：host/port/tls_mode、bind 凭证脱敏、base_dn/filter/template、属性映射、display_name、allow_auto_provision 默认关、connection_key=default。
- 生产必须 `ldaps` 或 `starttls`；`plain` 仅 localhost/127.0.0.1。
- 密码不写日志、不进 err 文案、不落库；超时；统一失败消息防枚举。

## HTTP 与前端

- 公开 `POST /portal/linapro-auth-ldap/login`；受保护 settings；无 OAuth redirect。
- settings 页 + 登录 slot 弹层；统一 Vben 按钮/弹层/表单与主题 token。

## 风险与迁移

- subject 避免默认可变邮箱；文档强调 GUID/UUID。
- 新插件；安装 core → ldap → 配置 → 启用；卸载不删已开户用户。
