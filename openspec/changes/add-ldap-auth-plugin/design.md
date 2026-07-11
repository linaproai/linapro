## Context

已确认决议：单 connection、平台全局、独立入口弹层、JIT 默认关、managed source、OIDC 优先（已交付 `linapro-oidc-generic`）。本变更交付 LDAP 第二波。

## Goals / Non-Goals

**Goals:**

- `linapro-auth-ldap`：可配置单 LDAP/AD 目录，登录页弹层账号密码，服务端 bind 验真后走 `extlogin` + handoff。
- TLS：默认要求 LDAPS 或 StartTLS；明文仅允许 localhost 开发（与 generic issuer http 策略同类）。
- 稳定 subject：可配置 subject 属性；缺省推荐 `entryUUID`（OpenLDAP）/`objectGUID`（AD）回退 `uid`。
- 设置页 + 双语 i18n + README + 单元测试（mock dialer/bind）+ E2E 骨架。

**Non-Goals:**

- 多目录、租户级目录、组→角色同步、密码同步入库、宿主密码路径 hook、SAML。

## Decisions

### D1. 插件与 provider

- ID：`linapro-auth-ldap`
- `provider = "ldap:default"`
- `SubjectKind = custom`（GUID/UUID/uid 均视为目录稳定键）
- `ProvideExternalIdentity("ldap:default")` + catalog

### D2. 验真流程

```
用户弹层提交 username + password
  → 插件建立 TLS 连接
  → （可选）服务账号 bind + search 得用户 DN
  → 或 username→DN 模板直接替换
  → 用户 DN + password bind 验真
  → 读取 subject/email/displayName 属性
  → LoginByVerifiedIdentity
  → CreateLoginHandoffFromHost
  → JSON { handoff } 给 SPA
```

两种 DN 解析模式（设置二选一或并存，search 优先）：

1. **Search**：`bindDN`/`bindPassword`（服务账号）+ `baseDN` + `userFilter`（含 `{username}`）
2. **Template**：`userDNTemplate` 如 `uid={username},ou=people,dc=example,dc=com`

### D3. 配置键（sys_config）

- host, port, tls_mode (`ldaps`|`starttls`|`plain` 仅 localhost)
- bind_dn, bind_password（脱敏）
- base_dn, user_filter, user_dn_template
- subject_attr, email_attr, display_name_attr
- display_name（登录按钮文案）
- allow_auto_provision（默认关）
- connection_key 固定 `default`

### D4. HTTP

- PUBLIC `POST /portal/linapro-auth-ldap/login`：JSON `{username,password}` → `{handoff}` 或业务错误码
- PROTECTED settings GET/PUT（权限 `linapro-auth-ldap:settings:*`）
- 无 OAuth redirect

### D5. 前端

- settings.vue：目录连接与属性映射
- slot：圆形/文案按钮 → Modal 用户名密码 → 调 portal login → `authStore.completeExternalLoginFromHandoff(handoff)`

### D6. 安全

- 密码不写日志、不进 err 文案、不落库
- 连接/操作超时（如 10s）
- 统一失败消息防枚举（「账号或密码错误」）
- bind 失败与用户不存在对外同一错误码

## Risks / Trade-offs

| 风险 | 缓解 |
| --- | --- |
| 明文 LDAP | 默认拒绝非 localhost plain |
| subject 用可变 uid | 文档强调 GUID/UUID；设置可配 |
| go-ldap 依赖 | 仅插件 go.mod |
| 弹层 CORS/cookie | portal 同源 POST，与宿主同 origin |

## Migration Plan

新插件；安装 core → ldap 插件 → 配置 → 启用。卸载不删已开户用户。

## Open Questions

无阻塞；subject 默认属性：`entryUUID`，AD 部署可改为 `objectGUID`。
