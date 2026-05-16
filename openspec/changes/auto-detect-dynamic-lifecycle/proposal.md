## Why

动态插件目前需要在 `backend/lifecycle/*.yaml` 中逐个声明生命周期处理器，声明内容与 controller 方法命名高度重复，维护成本较高且容易出现方法已实现但声明遗漏的情况。

本变更通过构建期自动发现动态插件生命周期处理器，降低样例插件和用户插件的重复配置，同时保留 WASM artifact 中的显式生命周期契约，确保宿主运行时仍具备确定、可审计、可校验的调用依据。

## What Changes

- `build-wasm` 在打包动态插件时自动识别 guest controller 中与源码插件一致命名的生命周期方法，例如 `BeforeInstall`、`AfterInstall`、`BeforeUpgrade`、`BeforeUninstall`。
- 自动生成生命周期契约并写入既有 `lina.plugin.backend.lifecycle` WASM custom section，宿主运行时继续只按 artifact 中的 `LifecycleHandlers` 调用，不进行盲目运行时试探。
- `backend/lifecycle/*.yaml` 从必需声明降级为可选 override，仅用于覆盖默认 `requestType`、`internalPath`、`timeoutMs` 等构建期推导值。
- 官方 `plugin-demo-dynamic` 移除重复生命周期 YAML 声明，依赖自动发现生成 14 个生命周期契约。
- 补充构建工具、pluginbridge guest 侧反射元数据、动态 artifact 解析和 demo 插件打包测试。
- 不改变动态插件生命周期运行时调用语义、错误码、缓存失效策略、hostServices 授权边界或 REST API。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `plugin-runtime-loading`: 动态插件生命周期契约来源从手写声明扩展为构建期自动发现，并保留 artifact 内显式契约作为运行时权威输入。

## Impact

- 影响 `hack/tools/build-wasm` 动态插件构建工具、`apps/lina-core/pkg/pluginbridge/guest` controller 反射能力、动态插件 artifact lifecycle contract 生成和相关测试。
- 影响官方动态示例插件 `apps/lina-plugins/plugin-demo-dynamic` 的 lifecycle 声明方式和 README。
- 不涉及前端 UI、REST API、数据库迁移、i18n 运行时语言包、apidoc 资源或缓存/集群策略变更。
