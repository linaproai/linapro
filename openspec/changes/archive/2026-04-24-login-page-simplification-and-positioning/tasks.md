## 1. Wire host login-page config parameters

- [x] 1.1 Add protected public-frontend metadata, default values, domain validation, and whitelist projection for `sys.auth.loginPanelLayout` under `apps/lina-core/internal/service/config/`
- [x] 1.2 Update the built-in system-parameter seed list in the host initialization SQL so the login-panel position defaults to `panel-right` and remains manageable through the system-parameter page
- [x] 1.3 Add backend config unit tests that cover valid login-panel values, invalid value rejection, and `GET /config/public/frontend` output

## 2. Adjust login-page entry points and layout defaults

- [x] 2.1 Adjust frontend default preferences and public-frontend runtime sync so the login page defaults to `panel-right` and can consume the host's `auth.panelLayout`
- [x] 2.2 Adjust login-page assembly so forgot password, registration, mobile login, QR-code login, and third-party login entry points are explicitly hidden while username/password login remains
- [x] 2.3 Adjust auth-route registration and fallback logic so unfinished auth subpages all redirect to `/auth/login` and no longer appear as valid public entry points

## 3. Frontend interaction tests and regression verification

- [x] 3.1 Update the relevant frontend runtime tests so they cover how public-frontend config synchronizes the default login-panel position
- [x] 3.2 Create `hack/tests/e2e/auth/TC0102-login-page-presentation.ts` with these assertions: `TC-102a` unfinished entry points stay hidden; `TC-102b` the login page defaults to the right-aligned layout; `TC-102c` changing the system parameter switches the layout accordingly
- [x] 3.3 Run this iteration's relevant verification set, including backend config tests, frontend tests, and `TC0102`, and confirm that login-page presentation matches system-parameter behavior

## Feedback

- [x] **FB-1**: Expand the length limit for `sys.auth.pageDesc` to 500 characters and add matching system-parameter validation plus public-frontend regression coverage
- [x] **FB-2**: Change the default login-page description to `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model` and standardize the default login-panel layout to the right side
