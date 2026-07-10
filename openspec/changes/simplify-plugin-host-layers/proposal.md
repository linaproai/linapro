## Why

插件宿主实现已经形成稳定的双交付模型（源码插件 + 动态 WASM 插件）和治理能力，但 host service 路径仍存在多层手写同步、dedicated/JSON 双轨编解码、历史别名与升级编排独立包等冗余。这些成本会在每次新增领域方法时线性放大，阻碍“面向可持续交付的 AI 原生全栈框架”对插件扩展面的持续演进。

## What Changes

- 冻结现有 dedicated host-service codec 方法集合；新增 core-owned host service 方法必须使用统一 JSON envelope。
- 在 `hostservices` 单一维护 host service / method wire 常量，catalog 引用常量，治理测试防漂移；不引入编译期 `go generate`。
- 提供 WASM 侧 JSON host-service 分发共用 helper，降低新增 JSON 方法的样板代码。
- 删除 `HostServiceCapabilityJSON*` 历史命名包装，统一到 `HostServiceJSON*`。
- 将 runtime upgrade 编排归属到 lifecycle 门面：根 `plugin` 包只依赖 lifecycle 暴露的安装/启停/卸载/升级入口，不再单独持有 `upgrade.Service`。
- 明确 lifecycle → runtime 单向依赖，runtime 只承担动态执行与 reconcile 副作用，不成为管理生命周期的第二入口。
- 同步更新 `pkg/plugin` 中英文 README 的维护约定。

## Capabilities

### New Capabilities

- `plugin-host-layer-simplification`：插件宿主 host-service 载荷策略、catalog 常量生成、lifecycle 升级归属与协议别名收敛。

### Modified Capabilities

- `plugin-host-service-extension`：新增方法默认走 JSON envelope，dedicated codec 仅允许冻结名单内的存量方法。
- `plugin-manifest-lifecycle`：升级预览/执行统一经 lifecycle 编排入口暴露，根门面不再平行挂接 upgrade 包。

## Impact

- 代码：`apps/lina-core/pkg/plugin/pluginbridge/**`、`apps/lina-core/internal/service/plugin/**`、`apps/lina-core/pkg/plugin/README*.md`
- API/行为：对插件作者运行时协议无破坏性变更；删除仅宿主内部使用的 historical type/func 别名
- 测试：catalog 冻结测试、常量生成一致性测试、lifecycle/upgrade 装配测试、相关包 `go test`
- DI：upgrade 构造从根 `plugin.New` 下沉到 lifecycle 构造；仍复用启动期共享 catalog/store/runtime/migration 等实例
