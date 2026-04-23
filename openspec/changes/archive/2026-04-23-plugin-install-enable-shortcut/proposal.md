## Why

The plugin management page currently requires administrators to finish installation in the review dialog, return to the list, and then enable the plugin manually. For the common governance flow where a plugin should be enabled immediately after installation, this creates unnecessary clicks, state switching, and context interruption, especially after dynamic-plugin authorization review, so the workflow needs a direct shortcut.

## What Changes

- Add an `Install and Enable` shortcut action to the plugin installation dialog while preserving the existing `Install Only` path.
- Keep the frontend implementation aligned with the existing lifecycle order by calling install and enable sequentially instead of introducing a new plugin state machine.
- When installation succeeds but enablement fails, surface the real state as `installed but disabled` so administrators can retry or investigate.
- Require both install and enable permissions for the shortcut action; users with install-only permission continue to see only the install path.
- Add plugin-management E2E coverage for dynamic-plugin authorization review, source-plugin shortcut enablement, and permission-visibility boundaries.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `plugin-ui-integration`: Update the plugin installation dialog so administrators can choose `Install Only` or `Install and Enable` in the same review flow, with matching permission visibility and result messaging.
- `plugin-manifest-lifecycle`: Clarify that plugin governance may trigger enablement directly from the installation flow while the host still executes the existing `install -> enable` lifecycle sequence and preserves the real partial-success state.

## Impact

- Frontend plugin management page and installation review dialog: `apps/lina-vben/apps/web-antd/src/views/system/plugin/`
- Frontend plugin API client wiring: `apps/lina-vben/apps/web-antd/src/api/system/plugin/index.ts`
- Existing plugin lifecycle and state-transition API semantics: `apps/lina-core/api/plugin/v1/`, `apps/lina-core/internal/controller/plugin/`, `apps/lina-core/internal/service/plugin/`
- Plugin management and authorization-review E2E coverage: `hack/tests/pages/PluginPage.ts`, `hack/tests/e2e/extension/plugin/`
- OpenSpec delta specs: `plugin-ui-integration`, `plugin-manifest-lifecycle`
