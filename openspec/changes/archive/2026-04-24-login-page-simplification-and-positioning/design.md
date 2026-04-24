## Context

The current LinaPro login page is implemented by composing `@vben/common-ui`'s `AuthenticationLogin` component from `apps/lina-vben/apps/web-antd/src/views/_core/authentication/login.vue`. That shared component already includes switches for forgot password, registration, mobile login, QR-code login, and third-party login, but the current business page does not explicitly disable those entry points. At the same time, the frontend preference system already supports `authPageLayout`, which can switch among `panel-left`, `panel-center`, and `panel-right`, yet the repository does not fully route that default through host-managed public-frontend settings.

This change therefore is not inventing a brand-new login layout. It completes three tasks on top of the existing host config flow and Vben layout capabilities: stop exposing unfinished login methods, standardize the default LinaPro login panel to the right-aligned layout, and govern that default through host-managed `sys_config` public-frontend parameters.

## Goals / Non-Goals

**Goals:**
- Make the login page expose only username/password login in the current stage and stop showing unfinished authentication entry points.
- Use the existing `authPageLayout` capability instead of creating another layout enum, and make LinaPro default to `panel-right`.
- Manage the login-panel position through a built-in host public-frontend parameter so administrators can maintain the default layout through system parameters.
- Let unauthenticated pages read and apply the login-panel position through the existing public-frontend config endpoint.
- Add frontend and backend regression protection so unfinished entry points are not exposed again later.

**Non-Goals:**
- Do not implement real forgot-password, registration, SMS-code login, QR-code login, or third-party login flows in this iteration.
- Do not change the login API, JWT behavior, session governance, or permission payload structure.
- Do not add a separate login-page config center or a new public config API.
- Do not redesign the login-page visual style; only adjust exposure strategy and default layout behavior.

## Decisions

### 1. Reuse Vben's existing `authPageLayout` enum and standardize LinaPro on `panel-right`

The underlying preference system already supports three layout values, `panel-left`, `panel-center`, and `panel-right`, and `AuthPageLayout` already renders the left, center, and right variants. This change therefore does not invent a new login-position field or a CSS-specific custom string. It reuses that existing enum and standardizes LinaPro's default at `panel-right`.

Why:
1. It lets the implementation reuse the current layout switcher and rendering logic directly instead of maintaining a second mapping owned by the host.
2. A right-aligned default better matches the current visual focus of the login page, and the same semantics can be enforced both through frontend defaults and through the host system parameter.

**Alternatives considered:**
- Hard-code the login panel to the center and remove the enum entirely: smaller change, but it fails the requirement that the frontend already has a position option and that the default should be configurable through system parameters.
- Add a new custom position field and map it back to CSS classes: gives a business-specific name, but duplicates existing capabilities and increases adaptation cost across host and frontend.

### 2. Close unfinished login methods at both the page-entry and route-entry layers

Hiding buttons in the login card is not enough, because the router still registers pages such as `code-login`, `qrcode-login`, `forget-password`, and `register`. Users could still navigate to those unfinished pages directly by URL. To stop exposing a capability that looks supported but is not actually delivered, this iteration uses a two-layer shutdown strategy:

- explicitly disable forgot-password, registration, mobile-login, QR-code, and third-party entry points when assembling `AuthenticationLogin`;
- remove or redirect unfinished auth routes so those addresses are no longer formal public entry points.

That keeps both the visible page entry points and the accessible routing surface aligned with the product statement that only username/password login is supported today.

**Alternatives considered:**
- Hide only the buttons: simplest implementation, but dead pages remain reachable.
- Delete the unfinished page files entirely: gives the strongest closure, but increases rework cost when those capabilities are implemented later. Keeping the files while disconnecting them is the better trade-off for now.

### 3. Put login-panel position into the existing public-frontend whitelist through `sys.auth.loginPanelLayout`

Login-panel position is presentation config that unauthenticated pages must read, so it belongs in the existing `config.GetPublicFrontend()` whitelist instead of a dedicated endpoint. The new protected key should be `sys.auth.loginPanelLayout`, with allowed values `panel-left`, `panel-center`, and `panel-right`, grouped under `auth`, and returned through the same public-frontend payload.

Frontend runtime sync then maps that value into `preferences.app.authPageLayout`, which keeps it on the same application path as other public settings such as `themeMode` and `layout`. After administrators update the system parameter, the login page can use the new default on the next load.

The key belongs under the `auth` group rather than `ui`, because it only affects unauthenticated auth pages and does not affect the workspace layout. That keeps the semantics cohesive and avoids confusion with `sys.ui.layout`.

**Alternatives considered:**
- Put the field under `ui`: technically possible, but too easy to confuse with workspace layout settings.
- Change only the frontend default and add no host parameter: does not satisfy the requirement that administrators can configure the default position through system parameters.

### 4. Keep the existing frontend layout switcher and use host system parameters as the default source

The Vben login toolbar already includes a layout switcher that can move among left, center, and right layouts. This change keeps that capability instead of removing it, but lets the host system parameter define the initial default state. In practice that means:

- LinaPro's built-in default becomes `panel-right`.
- If the host maintains `sys.auth.loginPanelLayout` in `sys_config`, the login page uses the host value as its startup default.
- The toolbar switcher still works as an immediate frontend-only preview and interaction tool.

That matches both the current frontend capability and the governance requirement that the default source of truth is a system parameter.

### 5. Cover regression protection on both backend and frontend planes

This change crosses frontend page assembly, frontend runtime config sync, backend protected-parameter metadata, and whitelist projection. Verification therefore needs to cover at least two categories:

- **backend config tests** for allowed values, invalid values, whitelist projection, and default output;
- **frontend runtime/page tests** for hiding unfinished entry points, default right-aligned layout, and layout switching based on public-frontend config payloads.

That avoids one side changing without the other, such as a parameter existing without any page effect or a page default changing without the backend metadata being updated.

## Risks / Trade-offs

- **[Risk] Bookmarks or external links may still point to old auth sub-routes** -> **Mitigation**: redirect unfinished auth routes back to the standard login page instead of leaving them as 404s or misleading pages.
- **[Risk] Both host system parameters and frontend local layout switching can affect panel position, so priority must stay stable** -> **Mitigation**: treat public-frontend config as the startup default and keep runtime switching as an immediate per-session browser behavior only.
- **[Risk] Adding a new public-frontend parameter without updating SQL seed data or built-in metadata would make it unmanageable in the system-parameter page** -> **Mitigation**: update the built-in parameter list, validation logic, and public-frontend response tests together.
- **[Risk] Future reopening of multiple login methods may forget to re-register routes** -> **Mitigation**: keep the files but define clearly in the spec that the current stage only exposes username/password login, so future enablement must happen through a new change.

## Migration Plan

1. Update host public-frontend parameter metadata and default values, adding `sys.auth.loginPanelLayout=panel-right`.
2. Extend the public-frontend whitelist response and frontend runtime sync logic so `auth.panelLayout` maps to `preferences.app.authPageLayout`.
3. Adjust login-page assembly and auth-route registration so unfinished entry points are hidden and unfinished auth pages all redirect back to the standard login page.
4. Add backend config tests and frontend runtime/page tests that verify the default right-aligned layout and system-parameter-driven switching behavior.

## Open Questions

- Should a future iteration persist toolbar layout changes as a user-specific preference that overrides the system default? For now, the implementation keeps the host default as the page-load source of truth and leaves personalized preference governance out of scope.
