---
name: lina-review
description: >-
  Code and specification review for OpenSpec workflow. Triggers automatically after /opsx:apply
  task completion, after /opsx:feedback task completion, and before /opsx:archive. Use when
  user requests code review, spec compliance check, or when explicitly invoked via /lina-review.
compatibility: Requires OpenSpec CLI, GoFrame v2 skill, lina-e2e skill.
---

# Lina Review

Structured code and specification review for the OpenSpec development workflow.

**Spec Source**: `CLAUDE.md` is the single source of truth for all review criteria.

---

## When This Skill Activates

**Automatic triggers:**
- After completing each task in `/opsx:apply`
- After completing each task in `lina-feedback`
- Before executing `/opsx:archive`

**Manual trigger:**
- User explicitly requests: "review this code", "check spec compliance", "/lina-review"

---

## Review Workflow

### 1. Identify Scope

Determine what needs to be reviewed:

1. **After task completion** — Review files modified/created by the completed task
2. **Before archive** — Review all changes in the current OpenSpec change
3. **Manual invocation** — Ask user to specify scope or use current change

**Mandatory scope collection rules:**

1. Start with repository status, not `git diff` alone:
   ```bash
   git status --short
   git ls-files --others --exclude-standard
   ```
2. Treat **all tracked and untracked changes** as review candidates, including:
   - staged files
   - unstaged files
   - untracked files shown as `??`
   - untracked directories shown as `?? path/`
3. When `git status --short` reports an untracked directory, expand it to concrete files before review:
   ```bash
   find <path> -type f
   # Or prefer:
   rg --files <path>
   ```
4. If the task ran generators such as `make ctrl`, `make dao`, codegen scripts, or produced new test files, explicitly include the generated untracked files in review scope even if they do not appear in `git diff`.
5. `git diff` may be used only as a secondary narrowing aid after status collection. It is **never sufficient by itself** for review scope definition.

Run `openspec status --change "<name>" --json` to understand the current change state.

### 2. Load Specifications

Read `CLAUDE.md` to load all specifications. This is the single source of truth.

### 3. Backend Code Review

**Trigger**: Changes to any `*.go` files

1. Invoke `goframe-v2` skill for GoFrame framework conventions
2. Check against `CLAUDE.md` backend code specifications

### 4. RESTful API Review

**Trigger**: Any API endpoint changes

Check against `CLAUDE.md` API design specifications, including:
1. HTTP method and resource path compliance with RESTful rules
2. API DTO documentation metadata completeness
3. **API documentation i18n compliance** — `g.Meta` and hand-authored DTO documentation tags must use readable English source text, must not use Chinese source text or opaque i18n placeholders, must keep apidoc localization in dedicated apidoc i18n JSON resources separate from runtime UI i18n bundles, must use stable structured apidoc keys instead of source-text keys, must keep host `core.*` apidoc keys in lina-core resources and plugin `plugins.*` apidoc keys in each plugin's own `manifest/i18n/apidoc/<locale>.json`, must keep `en-US` apidoc JSON as an empty placeholder because English docs use source text directly, must not add service-layer Chinese-to-English fallback maps or apidoc JSON mappings for generated/framework/database schema metadata such as `internal.model.entity.*`, must display generated metadata as supplied by its source, must leave `eg/example` values untranslated and out of apidoc i18n resources, and must include tests or review checks that prevent missing non-English apidoc translations when English source text changes

### 5. Project Specification Review

**Trigger**: Any implementation changes

Check against `CLAUDE.md` architecture design specifications and code development specifications.

For every functional change, also perform an **i18n impact review**:
1. Identify whether the change adds, modifies, or removes user-visible text, menus, routes, buttons, form fields, table columns, status labels, prompts, validation/errors, API documentation, plugin manifests, seed data labels, or other localized content.
2. Verify the corresponding i18n JSON resources were added, updated, or deleted as needed, including frontend runtime locale files, host/plugin `manifest/i18n` resources, and dedicated `apidoc i18n JSON` files.
3. Flag hard-coded user-facing text, missing translation keys, stale/orphaned translation entries left after feature removal, and changes whose i18n impact was not explicitly evaluated.
4. If the change has no i18n impact, require the review result to state that conclusion explicitly.

For every backend service, the component main file may be named `<component>.go`; every non-main Go source file in that component directory MUST use the `<component>_*.go` prefix, and tests must follow the same prefix convention.

For every frontend change that introduces or modifies an enumerated business value (status, type, scope, mode, source, etc.), also perform a **dictionary-vs-locale single-source review**:
1. If a backend `sys_*` dictionary already owns the same enumeration (registered via `manifest/sql` or runtime dict registration), the frontend MUST consume that dictionary through `useDictStore().getDictOptions(...)` / `getDictOptionsAsync(...)` rather than maintaining a parallel `options: [...]` literal array with `$t(...)` labels in `frontend/pages/data.ts`, `*.vue` form schema, or sibling files. Static `options` arrays are only acceptable when no backend dictionary owns the enumeration.
2. When the same field is rendered as a `DictTag` in the table column, the search form, the create/edit form, and any preview/detail surface, every surface MUST use the same dictionary as its single source of truth. Flag any surface that injects literal label/value pairs alongside a sibling `DictTag` consumer of the same dictionary.
3. The shared frontend `pages.*` namespace (e.g., `apps/lina-vben/apps/web-antd/src/locales/langs/<locale>/pages.json`) MUST NOT carry translations that duplicate `sys_*` dictionary labels. If a host UI legitimately renders an enumeration owned by a plugin dictionary, the host backend (e.g., `usermsg`, `notify`) must surface a localized label field on its API response and the host frontend must consume that label directly; do not add `pages.status.<enum>` keys that mirror dict labels as a workaround for cross-module coupling.
4. When a backend-owned data field needs localized display in the frontend, prefer adding a localized field (e.g., `typeLabel`, `statusLabel`) on the backend service/API output and consume it directly. The frontend must not maintain `type === N ? $t(...) : $t(...)` mapping helpers that mirror dictionary semantics.
5. If the change removes a frontend `options` literal, also confirm any orphaned `pages.*` keys, fallback arrays, and `getXxxLabel` helpers are deleted in the same change so stale translation keys do not remain.

For every backend change that localizes export/import headers, field labels, or metadata projections, also perform a **hard-coded bilingual map review**:
1. Business modules MUST NOT maintain Go maps such as `englishLabels`, `chineseLabels`, or equivalent locale-to-label tables for user-visible text.
2. Such labels MUST be resolved through runtime i18n keys (for example `config.field.<name>`) and corresponding host/plugin `manifest/i18n/<locale>.json` resources, including the packed manifest copy when the host embeds delivery resources.

For every change that touches source-text-backed missing-message behavior, also perform a **source-text namespace registration review**:
1. `apps/lina-core/internal/service/i18n/` MUST NOT hard-code business-owned prefixes such as `job.handler.` or `job.group.default.` when deciding whether a key is source-text-backed.
2. Business modules that own source-text-backed runtime keys MUST register their namespaces through `i18n.RegisterSourceTextNamespace(prefix, reason)` from their own package initialization or wiring path.
3. Missing-message tests MUST cover both unregistered namespaces still appearing as missing and registered namespaces disappearing from the missing result.

For every backend change that adds localization projection logic, also perform an **i18n foundation boundary review**:
1. `apps/lina-core/internal/service/i18n/` is a foundational translation service and MUST NOT own business entity projection rules, business protection rules, or business translation-key derivation for modules such as menu, dict, sysconfig, jobmgmt, role, usermsg, notify, or plugin runtime.
2. Flag any i18n package API named like `ProjectMenu`, `ProjectDictType`, `ProjectBuiltinJob`, `ProjectPluginMeta`, or any other method that accepts business entities to mutate display fields.
3. Business modules MUST keep their own localized projection rules inside their package boundary and depend only on narrow i18n capabilities such as `ResolveLocale`, `Translate`, and `TranslateSourceText`.
4. Shared resource-loading utilities used by both apidoc and runtime i18n MUST live in a stable public component such as `pkg/i18nresource`; do not place them in `internal/service/i18n` if that would make unrelated services depend on the runtime i18n service package.

For every backend change that touches i18n service consumers, also perform a **minimal i18n interface dependency review**:
1. Business services, controllers, middleware, and plugin adapters MUST NOT declare fields as `i18n.Service` / `i18nsvc.Service` / `internali18n.Service` when they use only a subset of methods.
2. Prefer package-local narrow interfaces when the caller uses one or two methods; otherwise depend on the exported small interfaces `LocaleResolver`, `Translator`, `BundleProvider`, `ContentProvider`, and `Maintainer`, or an explicit composition of those interfaces.
3. The complete `i18n.Service` composite is reserved for constructors, service factories, and rare integration points that genuinely need the full surface; such uses must be justified in review.

For every change that touches the host i18n service or any caller of it, also perform a **runtime i18n cache hygiene review**:
1. Hot-path translation calls (`Translate`, `TranslateSourceText`, `TranslateOrKey`, `TranslateWithDefaultLocale`) MUST NOT clone the runtime message catalog. Flag any code that introduces `cloneFlatMessageMap` or equivalent full-map copies on the per-key lookup path; the cache returns a read-only merged view and direct `merged[key]` access is the contract.
2. Code outside `apps/lina-core/internal/service/i18n/` MUST NOT clone runtime message catalogs returned by the i18n service. The service is responsible for cloning before it hands maps to external consumers (`BuildRuntimeMessages`, `ExportMessages`); business modules and controllers must treat returned maps as read-only.
3. Every call to `InvalidateRuntimeBundleCache` MUST pass an explicit `i18n.InvalidateScope`. Flag any call that omits scope or uses a zero-value scope without justification — wiping every locale and every sector is reserved for full process-level reload paths and must include a comment explaining why a narrower scope is not feasible. Plugin lifecycle invalidations MUST set `Sectors: []Sector{SectorDynamicPlugin}` together with `DynamicPluginID`; database imports MUST set `Sectors: []Sector{SectorDatabase}` together with the affected `Locales`.
4. Every call to `InvalidateContentCache` MUST pass an explicit `i18n.ContentInvalidateScope`. Pure `ContentInvalidateScope{}` (wipe-all) is permitted only for test cleanup or for full reload paths; production callers must scope by `BusinessType` and/or `Locale`.
5. Any new sector contributing to the runtime cache MUST be registered in `apps/lina-core/internal/service/i18n/i18n_cache.go` (the `Sector` enum and the merge order in `mergeLocaleSectors`). Do not introduce ad-hoc sectors in business modules.

### 6. SQL Review

**Trigger**: New or modified files under `apps/lina-core/manifest/sql/`、`apps/lina-core/manifest/sql/mock-data/`、`apps/lina-plugins/**/manifest/sql/` or SQL snippets embedded in related delivery docs

Check against `CLAUDE.md` SQL file management specifications, at minimum covering:
1. File naming, versioning, and single-iteration single-file rules
2. Seed DML vs mock data separation
3. **Idempotent execution safety** — SQL must be safe to run multiple times without duplicate-object errors or duplicate seed data; verify use of `IF [NOT] EXISTS`, `IF EXISTS`, `INSERT IGNORE`, or equivalent safe re-entry patterns
4. **Seed write style compliance** — delivered SQL must reject `INSERT ... ON DUPLICATE KEY UPDATE` and reject explicit writes to `AUTO_INCREMENT` `id` columns in seed/mock/install data
5. Whether schema/data changes still match the current change scope and deployment path

### 7. E2E Test Review

**Trigger**: New or modified E2E test files in `hack/tests/e2e/` directory

1. Invoke `lina-e2e` skill for test conventions
2. Check against `CLAUDE.md` E2E test specifications

### 8. Generate Review Report

```markdown
## Lina Review Report

**Change:** <change-name>
**Scope:** <task-specific / full change>
**Files Reviewed:** <count>
**Scope Source:** `git status --short` + `git ls-files --others --exclude-standard` + task/change context

### Backend Code Review
✓ All checks passed / ⚠ N issues found

### RESTful API Review
✓ All endpoints compliant / ⚠ N violations found
✓ API documentation i18n compliant / ⚠ N apidoc i18n issues found

### Project Spec Review
✓ Compliant with CLAUDE.md / ⚠ N violations found
✓ i18n impact reviewed / ⚠ N i18n governance issues found
✓ Dictionary-vs-locale single-source compliant / ⚠ N dict double-source issues found

### SQL Review
✓ No SQL changes / ✓ SQL changes compliant / ⚠ N SQL issues found

### E2E Test Review
✓ Tests follow conventions / ⚠ N issues found

### Summary
- **Critical:** N (must fix before archive)
- **Warnings:** N (recommended to fix)

### Recommended Actions
1. [Specific action with CLAUDE.md reference]
```

---

## Issue Severity

| Level | Behavior |
|-------|----------|
| **Critical** | Block archive, must fix |
| **Warning** | Show but allow proceed |

---

## Integration Points

| Workflow Step | Behavior |
|---------------|----------|
| `/opsx:apply` task done | Review, offer to fix issues before next task |
| `lina-feedback` task done | Review, fix before marking complete |
| `/opsx:archive` | Review all changes, block on critical issues |

---

## Guardrails

- **CLAUDE.md is the single source of truth** — All spec references point to it
- Only check categories relevant to changed files
- Scope identification MUST include untracked files and expanded untracked directories; never rely on `git diff` alone
- Don't block on warnings — only critical issues block archive
- Include file paths and line numbers in issue reports
- Offer to fix issues automatically when straightforward
