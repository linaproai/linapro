## Why

The frontend already had static locale bundle support through `vue-i18n`, and browser requests carried `Accept-Language`, but the host had not yet formed project-level i18n capability. Menus, dictionaries, system parameter metadata, public login-page configuration, system information, plugin `plugin.yaml` metadata, and many backend-returned labels were still stored and returned as single-language text. Delivery teams therefore had to edit copy manually in many places, which could not reliably support multilingual project delivery.

LinaPro is positioned as an AI-driven full-stack development framework. Internationalization should be framework infrastructure, not something every delivered project rebuilds from scratch. This change establishes unified i18n capability across the host, default workspace, and plugins, enabling low-cost convention-based multilingual configuration and aligning frontend and backend behavior around shared language negotiation, resource organization, and dynamic copy loading.

## What Changes

- Add the host-level `framework-i18n-foundation` capability, defining unified locale resolution, translation resource organization, runtime message bundle aggregation, fallback strategy, and the multilingual business content storage model.
- Establish a three-layer i18n model for static UI copy, dynamic metadata, and business content, avoiding a design where all copy is maintained inside frontend static JSON.
- Define unified i18n resource conventions for the host and plugins, using a "file baseline plus database override" governance model that supports both delivery-time configuration and runtime maintenance.
- Extend backend dynamic metadata projection so menus, dictionaries, system parameter metadata, public frontend configuration, system information, and plugin metadata can return localized results according to the request language.
- Clarify backend-owned data localization boundaries: built-in roles, cron jobs, job groups, execution logs, audit logs, and other backend governance data must be projected by backend APIs according to the request language. The frontend no longer maintains translation mappings based on Chinese source text or backend anchors.
- Extend the default workspace language-loading flow by layering host runtime message bundles and plugin message bundles on top of the existing `vue-i18n` setup, and refresh menus, public configuration, and dynamic page content when the language changes.
- Add i18n resource loading rules to plugin declarations and lifecycle management, ensuring plugin resources are managed consistently by the host during installation, enablement, disablement, and uninstallation.
- Define i18n maintenance behavior, including locale enablement, translation import/export, missing translation checks, and stable translation key conventions to reduce maintenance cost for delivered projects.

## Capabilities

### New Capabilities
- `framework-i18n-foundation`: Provides host-level locale resolution, translation resource aggregation, runtime message bundle distribution, i18n maintenance, and multilingual business content modeling.

### Modified Capabilities
- `menu-management`: Menu management and menu route responses support localized titles according to the current language, with translations managed from stable business keys.
- `dict-management`: Dictionary types and entries support localized names, labels, and descriptions according to the current language for consistent consumption by management and business pages.
- `config-management`: System parameter metadata and public frontend configuration copy support i18n projection and provide a unified maintenance entry point for delivered projects.
- `login-page-presentation`: The login page displays localized title, description, and subtitle by combining host public configuration with frontend static locale bundles.
- `plugin-manifest-lifecycle`: Plugin manifests and plugin lifecycle flows support i18n resource declaration, loading, refresh, and removal.
- `plugin-ui-integration`: Plugin pages hosted by the workspace participate in the host locale context and runtime message refresh flow.
- `system-info`: Project introduction, component descriptions, and other display copy on the system information page support localized API responses and rendering.
- `role-management`: Protected built-in roles and other framework governance data return backend-projected read-only display copy according to the current language, while user-editable role fields keep their database value semantics.
- `cron-job-management`: Built-in cron jobs, job groups, and execution log display names are returned by the backend according to the current language. Code-registered source copy uses English, and user-created job data keeps its original values.

## Impact

- **Backend capabilities**: Affects request context, shared middleware, config service, menu/dictionary/plugin/system information services, the new i18n resource model, runtime message bundle APIs, and maintenance APIs in `apps/lina-core`.
- **Database model**: Adds locale, translation message, and multilingual business content tables, and establishes translation mapping conventions for stable business keys such as menus, dictionaries, configs, and plugins.
- **Frontend capabilities**: Affects `vue-i18n` initialization, runtime message loading, language switch refresh behavior, public frontend config synchronization, dynamic menu refresh, and plugin page refresh in `apps/lina-vben`.
- **Plugin ecosystem**: Affects resource organization around `plugin.yaml` in `apps/lina-plugins`, adding plugin translation resource directories and including them in host lifecycle management.
- **Delivery and maintenance flow**: Delivered projects use unified multilingual resource directories, import/export conventions, and missing-translation checks, reducing hard-coded copy and manual inspection.
