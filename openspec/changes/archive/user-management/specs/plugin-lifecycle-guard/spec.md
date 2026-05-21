## ADDED Requirements

### Requirement: LifecycleGuard Interface Family
Host SHALL define independent interfaces for plugins to implement as needed:
- `CanUninstaller.CanUninstall(ctx) (ok bool, reason string, err error)`
- `CanDisabler.CanDisable(ctx) (ok, reason, err)`
- `CanTenantDisabler.CanTenantDisable(ctx, tenantID int) (ok, reason, err)`
- `CanTenantDeleter.CanTenantDelete(ctx, tenantID int) (ok, reason, err)`

### Requirement: Veto Aggregation and Display
Host SHALL concurrently invoke all relevant plugins' corresponding hooks before executing protected action (timeout 5s/hook); any returning `ok=false` rejects; **no short-circuit**, all reasons aggregated and returned.

#### Scenario: Multi-plugin multi-veto
- **WHEN** Platform admin attempts to uninstall `multi-tenant`
- **AND** Hook A returns `(false, "tenants_exist", nil)`
- **AND** Hook B returns `(false, "billing_pending", nil)`
- **THEN** System rejects uninstall
- **AND** UI simultaneously shows A and B reasons

#### Scenario: Hook timeout
- **WHEN** Some plugin `CanUninstall` does not return within 5s
- **THEN** System treats as `ok=false`
- **AND** reason = `"plugin.<id>.guard.timeout"` (i18n key)

#### Scenario: Hook panic
- **WHEN** Some plugin hook panics
- **THEN** System recovers and treats as `ok=false`
- **AND** reason = `"plugin.<id>.guard.panic"`

### Requirement: reason Must Be i18n Key
All veto reasons SHALL return i18n key strings, prohibit hardcoded Chinese/English text; frontend renders by `bizctx.Locale`.

### Requirement: Platform Admin --force Channel
Platform admin SHALL bypass veto via `?force=true` query parameter with UI second confirmation; force operations must: UI require text input plugin ID for second confirmation; operation log write with `oper_type='other'` marking `platform_force_action`; config `plugin.allow_force_uninstall: false` disables force channel.

### Requirement: Hook Concurrency and Timeout
All veto hooks SHALL execute concurrently, single hook timeout 5s, total timeout 10s; exceeding total timeout treated as veto.

### Requirement: Hook Invocation Observability
All veto hook invocations SHALL be recorded in `monitor-operlog` with `oper_type='other'` marking `lifecycle_guard`, including plugin_id, hook_method, ok, reason, elapsed_ms, force, timeout, panic flags.
