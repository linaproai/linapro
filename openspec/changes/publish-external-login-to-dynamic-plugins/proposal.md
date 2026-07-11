## Why

产品原则已明确：**动态插件经安装授权后与源码插件同权、同信任级**。当前 `ExternalLogin` 与 `CreateFromExternal` 对动态插件 fail-closed，文档仍沿用「WASM 不可信」表述，与该原则冲突，也阻碍动态第三方登录插件扩展。

## What Changes

- 规范层：写明「安装授权后源码/动态同权同信」；删除「仅因 type=dynamic 永久拒绝高危能力」的表述。
- 动态 host service 发布：
  - `auth`：`external_login.login_by_verified_identity`
  - `users`：`users.create_from_external`
- guest `domainhostcall`：去掉永久 stub，改为真实 host call。
- WASM dispatcher：实现上述方法；调用方 pluginID 由宿主盖章；provider ownership 对源码插件走 `ProvideExternalIdentity`，对动态插件走 hostServices `resources.ref`（provider ID）授权。
- 测试：fail-closed 用例改为「未授权失败 / 授权成功」；补充动态路径单测。

## Capabilities

### New Capabilities

- `plugin-trust-parity`：安装授权后源码与动态插件同权同信原则。
- `dynamic-external-login`：动态插件外部登录与从外部建号的 host service 契约。

### Modified Capabilities

- （无已归档 baseline 强制 delta；以本变更 specs 为准。）

## Impact

- `apps/lina-core`：protocol 常量、catalog、wasm 分发、domainhostcall、plugin README、`.agents/rules`（若命中）。
- 动态插件作者：可在 `plugin.yaml hostServices` 声明并获授权后使用外部登录相关方法。
- 安全：仍依赖安装治理 + 方法授权 + provider ownership（源码声明 / 动态 resource ref）；宿主继续盖章 pluginID。
