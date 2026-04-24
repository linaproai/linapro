## Why

The login page still exposes unfinished entry points such as forgot password, registration, mobile login, QR-code login, and other login methods. That can easily mislead users into thinking the system supports those flows even though they are not implemented. At the same time, the default login-panel position needs to be standardized as a right-aligned layout, the login-page description copy needs to be updated to better match the product's positioning, and the final presentation should remain adjustable later through system parameters. We therefore need one formal iteration to complete the login-page presentation strategy.

## What Changes

- Narrow the default visible login capabilities so only the implemented username/password path remains visible, while forgot password, registration, mobile login, QR-code login, and third-party login stay hidden.
- Change the default login-panel position to the right side and align the existing frontend layout-adjustment capability with LinaPro's login-page configuration flow.
- Update the default login-page description copy so it better matches the framework's business-evolution and plugin-extension positioning.
- Add a login-panel position parameter to the host `sys_config` public-frontend configuration set so administrators can maintain the default login layout through system parameters.
- Extend the public-frontend config whitelist response and frontend runtime sync logic so unauthenticated pages can also read and apply the login-panel position.
- Add matching specs and implementation tasks for login-page presentation and system-parameter governance so future login methods can be added behind explicit switches instead of accidental exposure.

## Capabilities

### New Capabilities
- `login-page-presentation`: Define that the login page currently exposes only the username/password path, hides unfinished entry points, and defaults to a right-aligned login panel while still supporting position configuration.

### Modified Capabilities
- `config-management`: Extend built-in public-frontend system-parameter metadata and validation rules with a login-panel position parameter that is exposed to the login page through the public-frontend config endpoint.

## Impact

- **Frontend pages**: affects `apps/lina-vben/apps/web-antd/src/views/_core/authentication/`, `apps/lina-vben/apps/web-antd/src/layouts/auth.vue`, and the public-frontend runtime config sync logic.
- **Frontend shared components**: may affect login-page props or default behavior under `apps/lina-vben/packages/effects/common-ui/src/ui/authentication/`.
- **Backend config service**: affects the public-frontend config model, default values, and parameter validation in `apps/lina-core/internal/service/config/`.
- **System parameter management**: affects the built-in parameter list in host `sys_config`, and how the config-management page displays and maintains the new protected parameter.
