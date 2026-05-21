## ADDED Requirements

### Requirement: Multi-Tenant Terminology i18n Key Complete Coverage
Workbench i18n resources SHALL include complete bilingual translations covering: Tenant, Platform Administrator, Tenant Administrator, Member, Invite, Switch Tenant, impersonation (Acting as Tenant); login selector, tenant selection, suspend, delete flow text; veto reason keys; error code messages.

### Requirement: Translation Completeness Auto-Validation
CI tests SHALL include multi-tenant i18n key coverage validation; any target language missing key blocks release.

### Requirement: Multi-Tenant Menu i18n
Menu i18n projection rules SHALL only expose platform management top-level group's multi-tenant platform management entry; system SHALL NOT create separate "Tenant Workbench" top-level menu group.
