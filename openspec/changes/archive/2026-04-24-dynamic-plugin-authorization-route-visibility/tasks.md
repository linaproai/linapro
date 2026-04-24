## 1. Backend route review projection

- [x] 1.1 Add dynamic-route review fields to the plugin-management API DTOs, including method, real public path, access level, permission key, and summary.
- [x] 1.2 Project the current release route contracts into plugin list and detail data inside `apps/lina-core/internal/service/plugin/` and `apps/lina-core/internal/controller/plugin/`, while reusing the host public-path projection logic.
- [x] 1.3 Add projection tests that cover `public` routes, `login` routes, permission-bound routes, and empty route lists.

## 2. Frontend authorization and detail dialog enhancements

- [x] 2.1 Extend the plugin API client and view models under `apps/lina-vben/apps/web-antd/src/api/system/plugin/` to consume dynamic-route review fields.
- [x] 2.2 Update the dynamic-plugin installation and enablement review dialog to render a read-only route-information section alongside host-service authorization.
- [x] 2.3 Handle empty-route scenarios so plugins without declared routes do not render a redundant route block.

## 3. Regression coverage and E2E

- [x] 3.1 Extend `hack/tests/e2e/extension/plugin/TC0073-plugin-host-service-authorization-review.ts` with assertions for public paths, access levels, and permission keys in the authorization dialog.
- [x] 3.2 Run the related backend tests and `TC0073-plugin-host-service-authorization-review.ts` to confirm the dynamic-plugin authorization review flow still passes.

## Feedback

- [x] **FB-1**: Move the registered-route list below "Host Service Authorization Scope" and collapse routes after the first two items by default.
- [x] **FB-2**: Reuse the same registered-route presentation in the dynamic-plugin detail dialog and support expanding long lists there as well.
- [x] **FB-3**: Improve long-description layout in the plugin detail dialog so the base information table does not collapse.
- [x] **FB-4**: Strengthen governance section-title styling in both the authorization dialog and the detail dialog.
- [x] **FB-5**: Remove the "Authorization Requirement" field from the plugin detail dialog and render the plugin description as a dedicated full-width row in the base information table.
- [x] **FB-6**: Decouple the `demo-control` runtime interceptor from `plugin.autoEnable` and let read-only mode depend only on the plugin's current enabled state.
- [x] **FB-7**: Fix `TC0046` so the user-avatar dropdown regression matches the current topbar rendering structure and restores nickname, email, and avatar assertions.
- [x] **FB-8**: Fix the login recovery path in `TC0066` for the "uninstall while preserving data, then reinstall and recover" source-plugin scenario.
- [x] **FB-9**: Fix `demo-control` write-blocking error serialization so the frontend receives a clear "Demo mode is enabled" message instead of a generic no-permission message.
