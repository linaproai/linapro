// Shared E2E runtime configuration.
// 共享 E2E 运行时配置。
//
// All hard-coded endpoints (frontend dev server, backend HTTP API) are
// centralized here so that test files only depend on this single source of
// truth. Override via environment variables when running tests against a
// non-default deployment:
//   - E2E_BASE_URL          frontend dev server origin (default 127.0.0.1:5666)
//   - E2E_BACKEND_BASE_URL  backend HTTP origin (default 127.0.0.1:9120)
//   - E2E_API_BASE_URL      backend `/api/v1/` base URL; overrides backendBaseURL when set
//   - E2E_PUBLIC_BASE_URL   backend public origin (no `/api/v1/` suffix); overrides backendBaseURL when set
//
// 所有端口（前端开发服务器、后端 HTTP 接口）集中到此处，测试文件不应再单独
// 硬编码端口或回退地址。如需在非默认部署下运行测试，可通过上述环境变量覆盖。

const frontendBaseURL = process.env.E2E_BASE_URL ?? 'http://127.0.0.1:5666';
const backendBaseURL = process.env.E2E_BACKEND_BASE_URL ?? 'http://127.0.0.1:9120';

// Backend API base URL with the /api/v1/ prefix; trailing slash is required so
// relative paths like 'auth/login' resolve correctly under playwrightRequest.
// 后端 API 基础 URL，包含 /api/v1/ 前缀；保留末尾斜杠以便相对路径解析。
const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? `${backendBaseURL.replace(/\/$/, '')}/api/v1/`;

// Public origin used when accessing non-/api/v1 routes on the backend (such as
// /plugin-assets/ or /api.json).
// 直接访问后端非 /api/v1 路径（例如 /plugin-assets/ 或 /api.json）时使用的源。
const publicBaseURL =
  process.env.E2E_PUBLIC_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, '');

// Origin observed by the backend when the frontend dev server proxies a
// request (vite proxy `target` with `changeOrigin: true` rewrites the Host
// header to the backend host). Must stay aligned with the vite config so that
// tests asserting the proxied request origin (e.g. TC0175) remain stable.
// 前端开发服务器以 changeOrigin 方式代理请求时，后端实际看到的 Host。需与
// vite.config.mts 中 proxy target 保持一致，供 TC0175 等断言代理 origin 的用例使用。
const frontendProxyBackendOrigin =
  process.env.E2E_FRONTEND_PROXY_BACKEND_ORIGIN ??
  backendBaseURL.replace(/\/\/127\.0\.0\.1(?=[:/])/, '//localhost');

export const config = {
  adminUser: process.env.E2E_ADMIN_USER ?? 'admin',
  adminPass: process.env.E2E_ADMIN_PASS ?? 'admin123',
  baseURL: frontendBaseURL,
  backendBaseURL,
  apiBaseURL,
  publicBaseURL,
  frontendProxyBackendOrigin,
};
