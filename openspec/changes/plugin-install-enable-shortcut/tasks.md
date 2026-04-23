## 1. 插件安装弹窗交互改造

- [ ] 1.1 在 `apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-auth-modal.vue` 中新增“安装并启用”动作状态，保留“仅安装”路径，并根据当前模式动态输出按钮文案与提交逻辑。
- [ ] 1.2 在 `apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue` 中补充组合动作入口所需的权限判断与弹窗入参，确保只有同时具备 `plugin:install` 和 `plugin:enable` 的用户才可使用“安装并启用”。
- [ ] 1.3 完成组合动作的成功、部分成功与失败提示，并在每种结果后刷新插件列表状态，确保页面展示真实安装/启用状态。

## 2. 生命周期复用与回归保护

- [ ] 2.1 复核并接线现有 `pluginInstall` / `pluginEnable` 调用顺序，确保组合动作继续复用现有 install -> enable 生命周期，而不新增复合接口。
- [ ] 2.2 验证动态插件在组合动作中复用安装阶段授权快照、不重复弹出授权审查窗口；如实现中存在阻碍，最小化调整前后端接口编排代码以满足该约束。
- [ ] 2.3 复核源码插件与动态插件在启用失败后的状态落点，确保系统保留“已安装、未启用”的真实结果，且后续仍可手动启用。

## 3. 自动化验证

- [ ] 3.1 扩展 `hack/tests/pages/PluginPage.ts` 页面对象，补充安装弹窗“安装并启用”与部分成功提示相关操作封装。
- [ ] 3.2 新增 `hack/tests/e2e/extension/plugin/TC0103-plugin-install-enable-shortcut.ts`，覆盖动态插件授权审查下直接安装并启用、源码插件快捷启用以及权限可见性边界。
- [ ] 3.3 运行新增用例及受影响的插件管理回归用例，确认“仅安装”“启用开关”“卸载”现有流程未被破坏。
