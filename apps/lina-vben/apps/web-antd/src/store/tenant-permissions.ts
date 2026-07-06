export const tenantAccessCodes = {
  listLoginTenants: 'system:tenant:auth:login-tenants',
  listPlatformTenants: 'system:tenant:list',
} as const;

export function hasTenantAccessCode(codes: string[] = [], code: string) {
  return codes.includes('*') || codes.includes(code);
}

export function canListLoginTenants(codes: string[] = []) {
  return hasTenantAccessCode(codes, tenantAccessCodes.listLoginTenants);
}

export function canListPlatformTenants(codes: string[] = []) {
  return hasTenantAccessCode(codes, tenantAccessCodes.listPlatformTenants);
}
