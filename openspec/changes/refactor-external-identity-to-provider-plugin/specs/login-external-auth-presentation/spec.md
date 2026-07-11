## ADDED Requirements

### Requirement: 登录页外部登录入口按形态分槽展示
宿主登录页 SHALL 将外部登录入口按展示形态拆分为两个独立宿主槽位，避免平台账号与协议/目录登录混用同一布局：

- `auth.login.after`：通用协议或目录登录（如通用 OIDC、LDAP）→ 全宽单行按钮纵向排列。
- `auth.login.social`：第三方平台账号登录（如 Google、Discord、QQ、微信、GitHub）→ 「其他登录方式」分隔线 + 横向圆形图标按钮行。

两区域 MUST 在各自无已启用插件注入时整块隐藏，且互不影响。

#### Scenario: 仅启用协议/目录登录插件
- **WHEN** 仅有通用 OIDC 或 LDAP 等协议/目录插件向 `auth.login.after` 注入入口
- **THEN** 登录页 MUST 以全宽按钮纵向展示这些入口
- **AND** MUST NOT 展示「其他登录方式」分隔线或社交图标区域

#### Scenario: 仅启用平台账号登录插件
- **WHEN** 仅有 Google 或 Discord 等平台插件向 `auth.login.social` 注入入口
- **THEN** 登录页 MUST 展示「其他登录方式」分隔线
- **AND** MUST 以横向图标按钮聚合展示这些入口
- **AND** MUST NOT 以全宽协议按钮样式展示这些平台入口

#### Scenario: 同时启用两类插件
- **WHEN** 协议/目录插件与平台账号插件均已启用
- **THEN** 登录页 MUST 先展示 `auth.login.after` 全宽按钮区域
- **AND** 再展示 `auth.login.social` 分隔线与图标行
