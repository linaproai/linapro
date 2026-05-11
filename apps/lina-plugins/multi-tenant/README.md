# multi-tenant

`multi-tenant` is the official LinaPro source plugin for tenant management.

English | [简体中文](README.zh-CN.md)

## Scope

This plugin owns:

- tenant CRUD and lifecycle status changes
- user membership management
- tenant resolution strategies with code-owned built-in policy
- tenant lifecycle guard checks and explicit plugin-governance provisioning
- platform impersonation and lifecycle guard checks

## Host Boundary

The host keeps only stable seams such as `tenantcap`, tenancy middleware, `bizctx` tenant identity, tenant-aware cache scopes, and plugin governance fields. Tenant entities, memberships, resolver policy, lifecycle guards, and tenant administration workflows stay inside this plugin.

When this plugin is not installed or enabled, the host falls back to `tenant_id = 0` and keeps the single-tenant experience.

## Configuration

The multi-tenant defaults are owned by code instead of the host configuration template:

| Default | Value | Meaning |
|---------|-------|---------|
| Isolation model | `pool` | Pool model with `tenant_id` columns. |
| Cardinality | `multi` | Allows one user to join multiple tenants. |
| Resolver chain | `override,jwt,session,header,subdomain,default` | Resolver order. |
| Ambiguous behavior | `prompt` | Returns a tenant selection requirement when no tenant is obvious. |
| Root domain | empty | Reserved for a later settings release; subdomain resolution is disabled for now. |

Resolver policy is fixed in code in the first iteration. There is no resolver configuration table, no runtime mutation path, and no distributed resolver-policy broadcast. `rootDomain` is kept empty until a later settings design explicitly opens it.

## Plugin Governance

The plugin is `platform_only`, `supports_multi_tenant: false`, and `global` because it owns the tenant control plane itself. Tenant-aware business plugins should use `scope_nature: tenant_aware`, declare `supports_multi_tenant: true` when they support tenant-level governance, and choose either `global` or `tenant_scoped` install mode.
