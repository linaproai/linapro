## Context

产品定调：动态插件安装授权后与源码插件同权同信。现状动态 guest 对 `ExternalLogin` / `CreateFromExternal` 永久 stub，与定调不符。

## Goals / Non-Goals

**Goals:**

- 规范与 README 对齐同权原则。
- 动态插件可经 hostServices 授权调用与源码等价的外部登录换会话、从外部身份建号。
- 保留：宿主盖章 pluginID、插件启用检查、provider ownership、方法级授权。

**Non-Goals:**

- 本变更不强制改为「宿主验 id_token」（可作后续加固）。
- 不实现完整动态 OIDC 协议插件参考实现。
- 不改变源码插件 `ProvideExternalIdentity` / `ProvideExternalIdentityProvider` 模型。

## Decisions

### D1. 同权定义

安装/升级治理通过且处于启用状态的动态插件，与源码插件适用同一能力准入模型：声明 → 授权 → 启用 → 调用。不得仅因 `type=dynamic` 拒绝发布能力。

### D2. 动态 provider ownership

- 源码：`ProvideExternalIdentity(providerID)`（不变）。
- 动态：在 `hostServices` 中声明 `service: auth` 且 `resources[].ref` 列出拥有的 provider ID；dispatcher 校验请求 `provider` 命中授权 resource 后才调用铸会话。

### D3. 方法命名（wire）

- `external_login.login_by_verified_identity`
- `users.create_from_external`

### D4. 调用链

```
dynamic guest
  → domainhostcall host call
  → wasm dispatcher（授权 + ownership）
  → capability.Services.ForPlugin(pluginID).Auth().ExternalLogin()
     或 Users().CreateFromExternal()
  → 宿主 auth / user 实现
```

`ownsProvider` 对非源码插件：若 GetSourcePlugin 失败，则视为动态路径——在 WASM 层做 resource ownership，适配器层对「非源码但已绑定 pluginID」在启用时允许进入（ownership 已在 dispatcher 校验），或扩展 ownsProvider 接受动态 ownership lookup。

实现选择：**dispatcher 校验 resource ownership 后，直接调用 `auth.LoginByExternalIdentity` 并盖章 PluginID**（与 adapter 等价），避免双重 ownership 语义分叉；CreateFromExternal 直接走 users capability。

### D5. 风险

安装授权后的动态插件若被攻破，可铸会话——与源码插件被攻破同等信任模型；靠运营治理与审计缓解。后续可加宿主验签加固。

## Risks / Trade-offs

- 动态 ownership 依赖 manifest resources 正确声明；声明错误会登录失败（fail-closed 于授权）。
- catalog 测试可能依赖方法清单完整性，需同步常量与 registry。
