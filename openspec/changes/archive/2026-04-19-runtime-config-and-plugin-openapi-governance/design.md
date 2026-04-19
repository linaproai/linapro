## Context

The merged iteration completed two host-governance efforts that were implemented in parallel but should have been managed as one OpenSpec iteration.

The first effort established a governed runtime configuration model for host-consumed settings. The host already depended on settings such as JWT expiry, session timeout, upload limits, and login IP blacklists, but the parameter-management layer did not yet provide a clear protected registry, runtime-safe validation, or multi-instance-friendly cache behavior.

The second effort moved API document generation under host control. GoFrame’s built-in `/api.json` output could not distinguish source-plugin routes from host routes, could not project source-plugin routes by enablement state, and would have required a duplicate route declaration model if the project had tried to solve the problem through `plugin.yaml`.

Several follow-up improvements were also folded into this merged iteration:

- plugin detail dialog and host-service presentation refinements
- OpenSpec language-governance rules
- structured logging and unified log sinks
- configuration extension namespacing
- comment-conformance cleanup for affected backend code paths

## Goals

- Keep a single archived record for the completed active changes.
- Preserve source-plugin route flexibility while giving the host explicit route ownership metadata.
- Ensure runtime configuration actually drives host behavior instead of existing only as editable key-value records.
- Keep hot-path runtime reads on process memory whenever possible, without breaking multi-instance convergence.
- Archive the iteration with English proposal, design, tasks, and delta specs.

## Non-Goals

- No new business capability is introduced during archival.
- No route-prefix constraint is imposed on source plugins.
- No duplicate route declaration model is added to `plugin.yaml`.
- No dynamic-plugin middleware model is forced onto source plugins.

## Decisions

### 1. Archive the two active changes as one merged governance iteration

Instead of archiving two separate completed active changes, the repository now keeps one merged archive entry. This restores the intended workflow constraint that completed work from one implementation window is represented by one archived iteration.

### 2. Runtime configuration is modeled as protected host-owned contract

Protected runtime parameters and protected public frontend settings are registered centrally in the host configuration service. The contract includes:

- stable key ownership
- default values
- validation rules
- runtime override lookup
- protection against rename and deletion

This keeps parameter governance in one place instead of scattering rules across auth, session, upload, UI bootstrap, and import flows.

### 3. Runtime reads use local snapshots plus shared revision convergence

Hot-path host behavior does not read `sys_config` for every request. Instead:

- each process keeps an immutable parsed snapshot in local cache
- writers bump a shared revision and clear their own local snapshot immediately
- other nodes converge through periodic revision synchronization
- single-node mode skips the shared coordination path and keeps only the local invalidation model

This design reduces hot-path overhead while preserving bounded cross-node convergence.

### 4. Public frontend settings are exposed through a whitelist endpoint

Unauthenticated pages and bootstrap flows need a safe subset of host-managed settings. The design therefore keeps:

- a whitelist contract
- structured typed response payloads
- no arbitrary public key lookup

This allows login pages and workspace bootstrap to consume branding and theme settings without exposing generic configuration access.

### 5. The host owns `/api.json`

The host no longer relies on GoFrame’s default OpenAPI output as the source of truth. The host-managed OpenAPI builder now:

- scans real host routes for documentable static APIs
- excludes internal and non-business routes
- excludes plugin routes from the host-static route set
- projects enabled source-plugin routes using captured route bindings
- projects enabled dynamic-plugin routes using runtime route contracts

This gives the host precise control over what appears in system API documentation.

### 6. Source-plugin route ownership is captured at registration time

Source plugins still define routes in code only. The host wraps route registration with an observable facade that records:

- plugin ID
- method
- path
- handler ownership
- DTO `g.Meta` documentation metadata when present

This avoids path-prefix heuristics and avoids duplicating route declarations in plugin manifests.

### 7. Source-plugin middleware remains plugin-owned

Source plugins keep their current flexibility for middleware registration and ordering. The host captures route ownership and performs real binding, but it does not try to reinterpret source-plugin middleware chains as dynamic-plugin-style declarative middleware descriptors.

### 8. Archived artifacts are written in English

The merged archive follows the project governance rule that archived change documents and archived delta specs use English, regardless of the user’s interaction language during implementation.

## Risks and Trade-offs

- A merged archive is less granular than keeping two separate archive entries, but it restores workflow consistency and keeps the implementation window intact.
- Baseline specs touched by this archive may still contain older legacy wording outside the newly synced sections; this archive updates the merged change scope rather than rewriting the entire OpenSpec history.
- Source-plugin raw handlers remain registrable but are not automatically projected into OpenAPI without DTO metadata, which is intentional to avoid a second route-truth system.

## Migration and Archive Plan

1. Create one merged archive directory with English proposal, design, tasks, and delta specs.
2. Sync the merged change requirements into the relevant baseline specs.
3. Remove both completed active change directories.
4. Verify that `openspec/changes/` no longer contains active change directories.
5. Keep the merged archive as the single historical record for this implementation window.
