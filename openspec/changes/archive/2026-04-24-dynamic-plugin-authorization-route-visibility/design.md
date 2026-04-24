## Context

The current dynamic-plugin authorization dialog reads directly from plugin list data. It already renders `requestedHostServices` from the `SystemPlugin` row object, which keeps the interaction lightweight and avoids a second request when the dialog opens. That also means any new governance information shown in the dialog must be prepared during backend projection.

Dynamic-route contracts already exist inside dynamic-plugin manifest and release snapshots, and the host already knows how to project their real public paths into OpenAPI output. What is missing is the administrator-facing projection inside plugin management. This iteration spans backend manifest projection, plugin-management DTOs, frontend dialog view models, and regression coverage. The design must preserve three things: route exposure and `hostServices` remain semantically distinct, the dialog still opens without extra round trips, and the route path shown to administrators is the real host public path rather than the guest-side `route.Path`.

## Goals / Non-Goals

**Goals**
- Add an independent route-information section to the dynamic-plugin installation and enablement review dialog.
- Let administrators review both host-resource access and externally exposed routes in a single governance flow.
- Keep host-service authorization visually first and route exposure second, while only showing the first two routes by default.
- Give governance section titles a clear visual hierarchy in both dialogs.
- Let the backend project route display fields so the frontend consumes a ready-to-render structure.
- Keep the existing authorization snapshot model unchanged; the new route data is review-only.
- Reuse the same route presentation in the plugin detail dialog.
- Show long plugin descriptions as a dedicated full-width row and remove the redundant authorization-requirement field from the detail dialog.
- Cover `public`, `login`, permission-bound, and empty-route scenarios with tests.

**Non-Goals**
- Do not change runtime route authorization, permission enforcement, OpenAPI projection, or the guest executor.
- Do not introduce per-route approval, filtering, or pruning.
- Do not expand static source-plugin route review in this iteration.
- Do not add a new detail endpoint or a dedicated route-review endpoint.

## Decisions

### 1. Render routes as a separate governance section

Route exposure and host-service authorization answer different governance questions. The dialog therefore renders routes in an independent review section rather than mixing them into the host-service cards. This prevents administrators from misreading routes as authorization items.

### 2. Project route display data on the backend

The backend projects method, real public path, access level, permission key, summary, and optional description from the current release snapshot. This avoids duplicating host-path assembly rules in the frontend and preserves the existing "open dialog from row data" interaction.

### 3. Keep the route list read-only and preserve declaration order

Both dialogs show the current release route list in declaration order without per-route confirmation. Only the first two items are shown initially, with an explicit expand action for longer lists.

### 4. Reuse the same route review structure in the detail dialog

The installation review dialog and the detail dialog serve the same governance inspection purpose before and after installation. Reusing the same route block reduces implementation drift and cognitive overhead.

### 5. Strengthen section-title hierarchy

Governance titles such as "Host Service Authorization Scope", "Host Service Information", and "Registered Routes" use stronger type, weight, and accent treatment so they act as obvious section dividers rather than blending into body text.

### 6. Move long descriptions back into the base information table

The detail dialog removes the standalone long-description block and renders the plugin description as a full-width row inside the base information table. This preserves readability without distorting the rest of the grid.

### 7. Hide empty route blocks

When the current release declares no dynamic routes, the frontend does not render a redundant empty route section. The backend may still return an empty collection; the UI simply treats it as "nothing to show."

## Risks / Trade-offs

- The plugin list payload grows because dynamic-plugin rows now include compact route review data. The payload stays limited to fields actually required by the dialogs.
- If the frontend accidentally frames routes as authorization items, the governance model becomes misleading. The section naming and structure therefore avoid terms such as "route authorization."
- Public-path assembly could drift if implemented in multiple places. The host must remain the single source of truth for that projection.
- Empty-route scenarios are easy to miss during implementation, so both backend tests and UI/E2E coverage must include them.

## Migration Plan

1. Update the `plugin-manifest-lifecycle` delta spec to define route exposure review requirements for dynamic plugins.
2. Add route review DTOs and backend projection logic for the current dynamic-plugin release.
3. Update the frontend authorization dialog and plugin detail dialog to consume the projected structure.
4. Add regression coverage for empty routes, `public` routes, `login` routes, and permission-bound routes.
5. Keep rollback simple: if necessary, remove the new frontend presentation and stop projecting the route fields because this change does not alter persisted storage.

## Open Questions

- None.
