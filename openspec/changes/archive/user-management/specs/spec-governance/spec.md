## ADDED Requirements

### Requirement: Multi-Tenant Sensitive Capability Specs Must Explicitly Declare Four Behavior Categories
Any capability spec involving tenant-sensitive business data, cache, audit SHALL explicitly declare `tenant_id` handling in Requirements covering: read path, write path, cache, audit four categories.

### Requirement: lina-review Audit Checklist Includes tenancy Check
`/lina-review` SHALL check: DAO calls pass `tenantcap.Apply`; tenant-sensitive cache keys carry tenant dimension; tenant-sensitive business table writes fill tenant_id; audit logs record tenant fields. Any violation is review failure.
