## 1. Installation dialog interaction updates

- [x] 1.1 Add an `Install and Enable` action state in `apps/lina-vben/apps/web-antd/src/views/system/plugin/plugin-host-service-auth-modal.vue`, keep the `Install Only` path, and switch button copy plus submit logic by mode.
- [x] 1.2 Extend `apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue` with the permission guard and modal payload needed by the composite action so only users with both `plugin:install` and `plugin:enable` can use `Install and Enable`.
- [x] 1.3 Finish success, partial-success, and failure messaging for the composite action, and refresh the plugin list after each outcome so the page shows the real install/enable state.

## 2. Lifecycle reuse and regression protection

- [x] 2.1 Review and wire the existing `pluginInstall` / `pluginEnable` call order so the composite action continues to reuse the current `install -> enable` lifecycle without adding a composite API.
- [x] 2.2 Verify that dynamic plugins reuse the authorization snapshot captured during install and do not open a duplicate authorization-review dialog during the composite action; if any blocker appears, keep the front-end and back-end orchestration changes minimal.
- [x] 2.3 Recheck the resulting state for source plugins and dynamic plugins after enablement failure so the system preserves the real `installed but disabled` result and still allows manual enablement later.

## 3. Automated validation

- [x] 3.1 Extend the page object in `hack/tests/pages/PluginPage.ts` with helpers for the installation dialog `Install and Enable` action and the related partial-success flow.
- [x] 3.2 Add `hack/tests/e2e/extension/plugin/TC0103-plugin-install-enable-shortcut.ts` to cover direct install-and-enable from the dynamic-plugin authorization review, the source-plugin shortcut flow, and the permission-visibility boundary.
- [x] 3.3 Run the new case plus the affected plugin-management regression cases to confirm the existing `Install Only`, enable-switch, and uninstall flows remain intact.
