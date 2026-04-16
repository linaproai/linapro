export const config = {
  adminUser: process.env.E2E_ADMIN_USER ?? 'admin',
  adminPass: process.env.E2E_ADMIN_PASS ?? 'admin123',
  baseURL: process.env.E2E_BASE_URL ?? 'http://localhost:5666',
};
