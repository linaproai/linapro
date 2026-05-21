## ADDED Requirements

### Requirement: Tenant Capability as Host Stable Seam
`pkg/tenantcap` SHALL be listed alongside `pkg/orgcap` as host "Stable Capability Seam"; seam must meet: interface contract stable, versioned changes; default no-op implementation; Provider + Register pattern for plugin integration; host does not hold specific business implementation.

### Requirement: Workbench Display Changes Must Not Sink to Core
Changes only from workbench display needs SHALL not be resolved by modifying `tenantcap.Provider` interface or core service; should be handled through workbench adaptation layer.

### Requirement: Seam Review Included in lina-review
`/lina-review` SHALL check any modifications to `pkg/tenantcap` / `internal/service/tenantcap` conform to "stable seam" four elements.
