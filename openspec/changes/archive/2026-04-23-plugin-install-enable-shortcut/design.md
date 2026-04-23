## Context

The plugin management page already uses one installation review dialog for source-plugin installation confirmation and dynamic-plugin host-service authorization review, but it still forces administrators to return to the list and toggle the enable switch after installation succeeds. The backend already exposes stable, separated install and enable lifecycle actions, and dynamic plugins can already reuse the authorization snapshot captured during installation, so this change is best implemented as a frontend shortcut flow rather than as a new lifecycle or storage model.

The change affects the plugin management list, installation dialog, permission visibility, and E2E coverage, but it does not require database schema changes, plugin manifest changes, or backend governance table changes.

## Goals / Non-Goals

**Goals:**
- Provide an `Install and Enable` shortcut inside the installation dialog to remove the extra round-trip back to the list.
- Preserve the existing `Install Only` path so current governance semantics do not change.
- Reuse the existing install and enable interfaces so source plugins and dynamic plugins continue to follow the same lifecycle order.
- Define the permission boundary and partial-success messaging clearly to avoid accidental misuse or misleading state feedback.
- Add plugin-management automated coverage for dynamic-plugin authorization review, source-plugin shortcut enablement, and permission visibility.

**Non-Goals:**
- Do not add database tables, SQL scripts, or new plugin status fields.
- Do not introduce a new backend composite endpoint or a new lifecycle state.
- Do not change the system-wide default to `enable after install`; administrators still choose explicitly between install-only and install-and-enable.
- Do not enforce strong transactional rollback across install and enable; if enable fails, the successful install is not automatically reverted.

## Decisions

### 1. Use a frontend composite action that calls the existing install and enable APIs in sequence

After adding an `Install and Enable` button to the review dialog, the frontend performs one ordered sequence within a single user action: call install first, then call enable after installation succeeds. This reuses the existing backend install, enable, permission-check, menu-refresh, and runtime-reconciliation logic without introducing extra DTOs, controllers, or service orchestration.

A new backend `install-and-enable` composite endpoint is intentionally not added because the current lifecycle chain is already stable; a new endpoint would increase API surface and regression scope without adding proportional UX value.

### 2. Keep the shortcut inside the installation review dialog instead of adding a new list action

The shortcut remains in the existing installation dialog footer instead of adding a separate `Install and Enable` button to the list action column. This ensures source-plugin detail confirmation and dynamic-plugin authorization review still happen in the same context and prevents the user from bypassing the required plugin-information review flow.

The dialog needs to support both:
- Install Only
- Install and Enable

`Install and Enable` is shown only when the current user has both `plugin:install` and `plugin:enable`. Users with install-only permission can still complete `Install Only`.

### 3. Reuse the authorization snapshot captured during install for dynamic plugins

For dynamic plugins, the authorization result submitted in the installation review dialog is still confirmed only once in the install request. When the administrator chooses `Install and Enable`, the follow-up enable step reuses the authorization snapshot already persisted during install, so no second authorization dialog appears and no duplicate confirmation is required.

This preserves the existing dynamic-plugin authorization model: the authorization snapshot is created during install, enable consumes the final snapshot, and historical releases without a confirmed snapshot still follow the existing enable-time confirmation path.

### 4. Use real-state-first messaging for partial success

If installation succeeds but enablement fails, the system does not roll back the installation result automatically. The frontend must refresh the plugin list and explicitly tell the administrator that the plugin is now installed but disabled and can be enabled again later.

This is intentional because install and enable are already two independent governance actions. An enable failure does not invalidate the installation result, and forced uninstall or rollback would create larger side effects and break the consistency of existing lifecycle events and audit behavior.

### 5. Cover the shortcut flow, permission boundaries, and failure messaging with E2E

The main risks in this change are UI orchestration and permission boundaries, so verification focuses on end-to-end behavior:
- Dynamic plugin: install and enable directly from the authorization-review dialog and verify that no second authorization review appears.
- Source plugin: enter the enabled state immediately after the shortcut flow finishes.
- Permission boundary: users with only install permission must not see `Install and Enable`; users with both permissions must see it.
- Failure feedback: simulate an enable failure in the second step and verify that the UI still shows the plugin as installed but disabled.

## Risks / Trade-offs

- [Partial success adds another messaging branch] -> Use a single `installed successfully, but enable failed` message plus a list refresh so administrators always see the real state.
- [Frontend permission checks may accidentally use any-of semantics] -> Implement a dedicated `has both install and enable permission` check for the shortcut instead of reusing single-permission helpers.
- [Adding another footer action can blur button priority] -> Keep the shortcut wording explicit and preserve `Install Only` as the conservative path.
- [Dynamic-plugin authorization flow could regress into repeated confirmation] -> Make the enable step reuse the install snapshot by default and cover the behavior with E2E regression tests.

## Migration Plan

1. Update the OpenSpec delta specs to define the shortcut action and its permission boundary.
2. Adjust the plugin installation dialog footer to add the `Install and Enable` action and wire the composite submission flow.
3. Extend the plugin page object and plugin-management E2E coverage for the shortcut flow, permission visibility, and dynamic-plugin authorization reuse.
4. Re-run plugin-management regression coverage to confirm the existing `Install Only`, enable-switch, and uninstall flows still work.

## Open Questions

- There are no open questions blocking archive. If button emphasis or priority needs refinement later, that can be handled as a UI polish follow-up.
