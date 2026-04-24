## 1. Host configuration and seam updates

- [x] 1.1 Add `plugin.autoEnable` examples to the host config template and make it clear that enabling `demo-control` controls the demo-mode switch
- [x] 1.2 Update the plugin workspace wiring entry so `demo-control` can be discovered, installed, and enabled through `plugin.autoEnable`

## 2. Demo-control source plugin implementation

- [x] 2.1 Add the base source-plugin structure, manifest, embed entry, and documentation for `apps/lina-plugins/demo-control/`
- [x] 2.2 Implement a demo-control service based on global HTTP middleware that blocks write requests across `/*` by `HTTP Method` while preserving the minimal session whitelist
- [x] 2.3 Ensure the demo-control plugin does not depend on any extra host boolean config and only takes effect while the plugin is enabled

## 3. Verification and regression protection

- [x] 3.1 Add `plugin.autoEnable` config tests that cover the default-disabled case and explicit enablement of `demo-control`
- [x] 3.2 Add middleware tests that cover query allow, write rejection, `/*` global scope, login whitelist behavior, and passthrough while the plugin is disabled

## Feedback

- [x] **FB-1**: Remove the extra `demo.control.enabled` switch and use whether `plugin.autoEnable` contains `demo-control` as the demo-mode switch
- [x] **FB-2**: Expand the demo-control global middleware scope to `/*`, cover the full system request chain, and preserve the login whitelist
- [x] **FB-3**: Add the `TC0105` E2E case to cover `plugin.autoEnable` enablement of `demo-control`, login/logout whitelist behavior, query passthrough, and `/*` global write interception
- [x] **FB-4**: Fix the issue where removing `demo-control` from `plugin.autoEnable` still left write interception active at runtime, ensuring the auto-enable list is the single source of truth for demo mode
- [x] **FB-5**: Allow install, uninstall, enable, and disable operations for plugins other than `demo-control` while demo mode is active, and add matching unit-test plus `TC0105` coverage
