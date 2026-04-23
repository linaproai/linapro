## Why

插件管理页当前要求管理员先在安装弹窗中完成安装，再回到列表手动启用插件。对于大多数“安装后立即启用”的治理场景，这会制造额外点击、状态切换和认知中断，尤其在动态插件授权审查后体验更割裂，因此有必要补上一条更顺手的一键路径。

## What Changes

- 在插件安装弹窗中新增“安装并启用”快捷动作，同时保留现有“仅安装”路径。
- 前端在管理员确认后按既有生命周期顺序串行执行安装和启用，而不是引入新的插件状态机。
- 当安装成功但启用失败时，界面明确提示插件当前处于“已安装、未启用”状态，便于管理员后续重试或排查。
- “安装并启用”动作必须同时满足安装权限与启用权限；若仅具备安装权限，则界面只展示“仅安装”。
- 补充插件管理相关的 E2E 用例，覆盖动态插件授权审查、源码插件快捷安装启用以及权限可见性边界。

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `plugin-ui-integration`: 调整插件安装弹窗交互，支持在同一弹窗内执行“仅安装”或“安装并启用”，并补充相应权限可见性与结果提示。
- `plugin-manifest-lifecycle`: 明确插件治理允许管理员在安装链路中直接触发后续启用动作，但宿主仍按既有 install -> enable 生命周期顺序执行，并保留部分成功后的真实状态反馈。

## Impact

- 前端插件管理页与安装授权弹窗：`apps/lina-vben/apps/web-antd/src/views/system/plugin/`
- 前端插件 API 调用封装：`apps/lina-vben/apps/web-antd/src/api/system/plugin/index.ts`
- 现有插件生命周期与状态切换接口语义校验：`apps/lina-core/api/plugin/v1/`、`apps/lina-core/internal/controller/plugin/`、`apps/lina-core/internal/service/plugin/`
- 插件管理与授权审查 E2E：`hack/tests/pages/PluginPage.ts`、`hack/tests/e2e/extension/plugin/`
- OpenSpec 增量规范：`plugin-ui-integration`、`plugin-manifest-lifecycle`
