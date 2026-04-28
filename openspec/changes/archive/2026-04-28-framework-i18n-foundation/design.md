## Context

The project already had scattered prerequisites for internationalization, but it still lacked framework-level i18n capability:

1. The frontend already had a `vue-i18n` base. `apps/lina-vben/packages/locales/` and `apps/lina-vben/apps/web-antd/src/locales/` could load static locale bundles, and workspace requests already carried `Accept-Language`.
2. The backend did not have a locale context. `apps/lina-core` had no unified `LocaleResolver`, translation service, or request-level locale resolution flow, and most APIs still returned single-language database text directly.
3. Dynamic metadata was still single-language. Menus, dictionaries, system parameter names/remarks, public login-page configuration, system information page metadata, and plugin `plugin.yaml` `name/description/menu.name` fields were stored and returned directly.
4. The business content multilingual model was missing. Business projects that needed multilingual titles, descriptions, or body content had to add their own fields or rebuild the same pattern independently.
5. Maintenance was not unified. There was no stable translation-key convention and no "file baseline plus runtime override" resource governance model, making multilingual configuration costly for delivered projects.

Because this is a new project, no historical compatibility layer is required. This design directly establishes unified i18n infrastructure instead of preserving long-term dual-track behavior for old formats.

## Goals / Non-Goals

**Goals:**
- Provide host-level i18n infrastructure so LinaPro delivered projects can support multiple languages as a built-in framework capability.
- Establish a unified locale resolution flow so the frontend and backend operate on the same locale context.
- Establish a unified translation resource organization model covering host, delivered project, and plugin resources, and aggregate them into runtime message bundles by locale.
- Support dynamic metadata localization for menus, dictionaries, system parameter metadata, public login-page configuration, system information, and plugin metadata.
- Provide a common multilingual content storage model for business modules instead of requiring each module to create its own field set.
- Keep configuration and maintenance simple through stable translation-key conventions, file baselines, database overrides, import/export, and missing-translation checks.
- Keep frontend static copy and backend dynamic metadata responsibilities separate.

**Non-Goals:**
- Do not introduce automatic machine translation, AI-generated copy completion, or integration with an external translation platform in this change.
- Do not require every existing business table to be immediately converted to multilingual content storage. This change provides the common capability and integration pattern for modules to adopt when needed.
- Do not build a complete visual i18n management page for every business module in the first phase. The first phase prioritizes a stable resource model, APIs, import/export, and runtime flow.
- Do not convert every existing error message in one pass. The first phase establishes the translation service and the rule that new user-visible copy uses translation keys; historical hard-coded messages can be converged module by module.
- Do not introduce persistent per-user locale preference storage in this scope. The first phase uses the `lang` parameter, `Accept-Language`, and the system default locale.

## Decisions

### Decision 1: Use a three-layer i18n model for static UI copy, dynamic metadata, and business content

**Choice:** Split internationalized text by source and lifecycle:

- **Static UI copy**: frontend framework and page-local static copy continues to be maintained by local `vue-i18n` JSON.
- **Dynamic metadata**: menus, dictionaries, system parameter metadata, system information, plugin descriptions, and similar metadata are projected by the backend according to locale.
- **Business content**: notice titles, body content, descriptions, and similar business content use a common multilingual content table and are adopted by business modules as needed.

**Rationale:**
- Static UI copy and dynamic metadata have different sources. Mixing them blurs maintenance boundaries.
- Menus, dictionaries, and system configuration metadata are backend-governed data and should be projected by the backend rather than hard-coded across frontend combinations.
- Business content usually has data identifiers and publishing lifecycles, so it fits a generic multilingual content model better than `_en/_ja/...` columns on every business table.

**Alternatives:**
- Put all translations into frontend static JSON. Rejected because dynamic menus, dictionaries, plugin metadata, backend errors, and external API consumers cannot be handled reliably that way.
- Add multilingual columns to every business table. Rejected because it inflates models and is not suitable as a framework-level general capability.

### Decision 2: Use a "file baseline plus database override" resource governance model

**Choice:** i18n resources use files as the baseline and database rows as overrides:

- The host and delivered projects maintain base locale resource files.
- Plugins declare translation files in their own directories.
- Runtime maintenance can enable locales, override messages, and maintain business content through the database.
- Message aggregation merges file resources and database resources according to a defined priority.

**Rationale:**
- Files fit code-versioned delivery and code review.
- Database overrides fit online patches, copy updates, and project-specific runtime changes.
- Combining both keeps delivery convenient while avoiding a release for every copy change.

**Resource priority:**
1. Database override messages
2. Project-level file resources
3. Plugin file resources
4. Host default file resources
5. Default fallback values

**Alternatives:**
- Store every translation only in the database. Rejected because delivery and plugin versioning would be awkward.
- Use only files and no database overrides. Rejected because runtime maintenance and small post-release changes would be too expensive.

### Decision 3: Add backend request-level locale resolution and store it in business context

**Choice:** Add a host `LocaleResolver` and request-context injection mechanism. The first-phase resolution priority is:

1. Query parameter `lang`
2. `Accept-Language` request header
3. System default locale

The resolved `locale` is available to controllers, services, and plugin host bridges for menus, dictionaries, configs, system information, error messages, and runtime message bundle aggregation.

**Rationale:**
- The frontend already sends `Accept-Language`, so adding backend resolution completes the frontend/backend loop.
- The `lang` query parameter is useful for downloads, public pages, debugging, and embedded contexts, and should override request headers.
- Writing the locale into business context prevents each service from reading headers independently.

**Alternatives:**
- Depend only on the frontend local language without backend awareness. Rejected because dynamic metadata and errors come from the backend.
- Persist user locale preference into `sys_user` in the first phase. Deferred because it is not required to establish the foundation.

### Decision 4: Define stable translation-key conventions and avoid manual mapping tables where possible

**Choice:** Derive translation keys from stable business keys. Runtime UI resource files may be authored as nested JSON or flat dotted keys; the host normalizes them to flat keys for governance:

- Menu title: `menu.<menu_key>.title`
- Dictionary type name: `dict.<dict_type>.name`
- Dictionary label: `dict.<dict_type>.<value>.label`
- Config name/remark: `config.<config_key>.name`, `config.<config_key>.remark`
- Plugin name/description: `plugin.<plugin_id>.name`, `plugin.<plugin_id>.description`
- System information component description: `systemInfo.component.<section>.<name>.description`
- Public frontend configuration copy: `publicFrontend.<group>.<field>`

**Rationale:**
- The project already has stable business keys or derivable anchors in menus, plugins, configs, and similar models.
- Convention-based key generation is easier to adopt than explicit mapping tables and works better for plugins and delivered projects.
- Nested JSON is easier for humans to maintain and review.
- Flat keys are better for database modeling, missing translation checks, import/export, and backend translation components such as GoFrame `gi18n`.
- Runtime APIs can still convert flat keys into nested message objects for frontend consumption.
- API documentation translation resources also support nested JSON and may be split under `manifest/i18n/<locale>/apidoc/**/*.json`; the host normalizes them to stable structured keys.
- Repeated OpenAPI metadata such as standard response wrappers, pagination fields, and common time fields use `core.common.*` fallback keys, while specific structure keys still take precedence when present.

**Alternatives:**
- Add a separate `i18n_key` field to every business table. Deferred because most scenarios already have stable anchors and extra fields would increase maintenance cost.

### Decision 5: Add unified translation resource tables for messages and business content

**Choice:** Add three core tables:

- `sys_i18n_locale`: locale registry with locale code, name, enabled state, and default marker.
- `sys_i18n_message`: message translation table with `key + locale + value + scope/source`, used for UI/metadata translation and database overrides.
- `sys_i18n_content`: multilingual business content table with `business_type + business_id + field + locale + content`, used for business body content.

**Rationale:**
- Message translations and business content translations have different lifecycles, lookup patterns, and cache behavior. Separate tables keep the model clear.
- `sys_i18n_content` gives business projects a shared extension model instead of requiring notices, articles, page configs, and other modules to each invent their own multilingual structure.

**Alternatives:**
- Store all translations in one large table. Rejected because message-key lookups and content-anchor lookups have very different access patterns.

### Decision 6: Backend projects dynamic metadata, while the frontend merges static copy and runtime messages

**Choice:**

- The backend localizes dynamic data such as menus, dictionaries, config metadata, system information, plugin manifests, and public frontend configuration.
- The frontend keeps the existing `vue-i18n` stack and adds a runtime loader that merges host runtime bundles, plugin bundles, and local static bundles.
- On language switch, the frontend refreshes runtime messages, public frontend configuration, dynamic menus/routes, and pages that depend on dynamic metadata.
- Translation files may use nested JSON or flat dotted keys. Database overrides, import/export, missing checks, and source diagnostics use flat keys, and the runtime interface converts them to nested message objects for the frontend.

**Rationale:**
- Dynamic metadata is owned by the backend, so backend projection lets the web app, plugin pages, and OpenAPI consumers share the same result.
- The frontend already has `vue-i18n`; extending its loading flow is safer than rebuilding it.

**Frontend/backend responsibilities:**
- **Backend**: resolve locale, aggregate translation resources, provide runtime message bundle APIs, output localized dynamic data, and provide import/export plus missing translation checks.
- **Frontend**: load and cache message bundles, switch languages, refresh dependent pages, and ensure components consistently use `$t` and runtime localized data.

**Boundary notes:**
- Backend-owned data includes host governance metadata, plugin governance metadata, built-in cron jobs, job groups, execution logs, protected built-in roles, and audit-log route metadata. These must be localized in backend APIs or backend export paths.
- The operation log plugin collects audit records through its own global middleware and writes through its own persistence service. The host only exposes route registration context, generic dynamic-route metadata reads, and apidoc text resolution. It does not define operation-log table-bound audit event payloads or maintain operation-log-specific metadata structures in the dynamic-route runtime.
- Plugin-specific declarations in dynamic plugin route contracts pass through the generic `meta` map. The host must not define or validate operation-log-specific fields such as `operLog` or operation type enums; the operation log plugin reads and interprets the `meta` keys it needs in its own middleware.
- The frontend must not maintain backend-data translation mappings based on Chinese source text, stable business keys, database IDs, `handlerRef`, `dict_type`, `config_key`, or other backend anchors. Such anchors are backend i18n key derivation inputs.
- The frontend may still format stable codes for pure UI concerns, such as permission string grouping, table width adaptation, status tag colors, and static UI `$t` copy, but it must not change business display values returned by APIs.
- Code-registered built-in cron job source copy uses English. Non-English display values come from backend translation resources, keeping English source copy, API documentation source text, and code registration aligned.

### Decision 7: Define a standard plugin i18n resource directory and manage it through lifecycle

**Choice:** Plugins add a standard `manifest/i18n/<locale>/` resource directory, with runtime JSON files maintained directly under each locale directory. The host handles plugin i18n resources during:

- source-plugin synchronization and discovery;
- dynamic plugin installation or upgrade, where resource snapshots are written for the release;
- plugin enablement, where resources join runtime message aggregation;
- plugin disablement or uninstallation, where resources are removed from runtime aggregation.

**Rationale:**
- Plugin i18n resources should be delivered with the plugin version instead of being manually maintained in the host project.
- Host-managed resource lifecycle keeps plugin visibility and language bundle visibility consistent.

**Alternatives:**
- Embed long multilingual structures directly in `plugin.yaml`. Rejected because the manifest would become large and hard to maintain.

### Decision 8: Provide a runtime message bundle API instead of having the frontend assemble module resources one by one

**Choice:** The host provides one runtime message bundle API that returns aggregated results for the current or requested locale. On startup and language switch, the frontend requests this bundle and layers it with local static bundles. The aggregation flow normalizes file resources and database overrides to flat keys, then converts them to nested objects for API output.

**Rationale:**
- One API can cover host, delivered project, plugin, and database override messages, avoiding frontend-side manual merging across menus, configs, and plugin resources.
- This shape makes caching, version stamps, and missing-key diagnostics easier.

**Alternatives:**
- Provide a separate translation API per module. Rejected because it increases frontend coordination complexity.

### Decision 9: Favor convention-based maintenance and provide import/export plus missing checks

**Choice:**

- Use stable key conventions, directory conventions, and template JSON to reduce the cost of adding a locale.
- Provide host APIs for locale enablement, message import/export, missing-key checks, and fallback reports.
- A management UI can be completed in a later iteration, but the first phase must establish the interfaces and model.

**Rationale:**
- The main delivery need is convenient multilingual configuration. Convention over configuration is more effective here than building a heavy CMS first.
- Import/export and missing checks are common delivery-stage maintenance needs and have higher priority than a complex editor.

## Risks / Trade-offs

- [Risk] The i18n scope spans host, frontend, plugins, and business models. Mitigation: implement in phases from infrastructure to dynamic metadata, plugin integration, and business content, with clear task boundaries.
- [Risk] Dynamic metadata and runtime bundles are cacheable. If cache invalidation lags after language changes or plugin lifecycle changes, copy can become inconsistent. Mitigation: define invalidation/versioning for runtime bundles, public frontend config, and menu data, and refresh them actively on frontend language switch.
- [Risk] Without stable translation-key conventions, delivered projects may duplicate equivalent copy in many places. Mitigation: specify key derivation rules and provide missing/duplicate checks.
- [Risk] Existing hard-coded error messages cannot all be converged immediately. Mitigation: add the translation service first, require new user-visible copy to use translation keys, and converge historical modules incrementally.
- [Risk] Plugin and host resources come from different sources. If merge priority is unclear, overrides can become confusing. Mitigation: define resource priority and expose actual source hits in export/diagnostic results.

## Migration Plan

1. Add host i18n infrastructure, including locale resolution, translation service, resource aggregation, and database model.
2. Add runtime message bundle, locale list, import/export, and missing-check APIs.
3. Localize dynamic metadata projection for menus, dictionaries, config, public login-page configuration, system information, and plugin manifests.
4. Extend the default workspace `vue-i18n` loading flow, connect it to runtime message bundles, and refresh public config, menus, and plugin pages on language switch.
5. Add plugin `manifest/i18n/` resource organization rules and include them in host-managed install, upgrade, enable, disable, and uninstall flows.
6. Provide the `sys_i18n_content` integration contract for business modules so later business projects can adopt multilingual content fields as needed.
7. Add import/export, missing-check, and regression tests so delivered projects can continuously maintain multilingual resources.

## Open Questions

- Should the first phase provide a complete i18n management UI, or should it first deliver APIs plus import/export capability? The current design prioritizes the latter so the foundation stabilizes first.
- Should multilingual business content support a unified structure for rich text, Markdown, and attachment references? The first phase focuses on text and long text; complex rich-text constraints can be added when business modules integrate.
- Should user locale preference later be persisted in the user profile and included in resolution priority? This change does not make it mandatory and leaves it for a later user-center expansion.
