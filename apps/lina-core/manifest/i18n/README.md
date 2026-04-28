# Runtime I18N Resources

This directory stores the delivery baseline i18n resources for `LinaPro`.

The host loads direct JSON files under `manifest/i18n/<locale>/` as runtime UI message bundles, merges them with enabled plugin resources, and exposes the effective result through the runtime i18n APIs.

API-documentation translations are stored under `manifest/i18n/<locale>/apidoc/**/*.json`. They share the same locale directory for discoverability, but stay in a dedicated `apidoc/` subdirectory because OpenAPI documentation is large and is only needed when `/api.json` is rendered.

## Directory Contract

| Path                                                        | Purpose                            |
| ----------------------------------------------------------- | ---------------------------------- |
| `manifest/config/config.template.yaml` `i18n` section         | Default locale, language switch, order, native name |
| `manifest/i18n/<locale>/*.json`                             | Host runtime locale bundle split by semantic domain |
| `manifest/i18n/<locale>/apidoc/**/*.json`                   | Host API-documentation bundle       |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/*.json` | Plugin-owned runtime locale bundle |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/apidoc/**/*.json` | Plugin-owned API-documentation bundle |

Rules:

- The directory name must use the canonical locale code, for example `zh-CN` or `en-US`.
- Built-in runtime locales are discovered from `manifest/i18n/<locale>/` directories with direct runtime JSON files.
- The default config `i18n` section stores metadata that cannot be inferred safely from the locale directory and may be customized by deployments. Runtime text direction is fixed to `ltr` by the current host convention.
- Adding a built-in language must not require changing Go constants, SQL seed files, or frontend TypeScript language lists.
- The host only treats direct `manifest/i18n/<locale>/*.json` files as runtime locale bundles.
- The host treats only `manifest/i18n/<locale>/apidoc/**/*.json` files as API-documentation locale bundles.
- Runtime UI message files may be authored as nested JSON or flat dotted keys.
- The host normalizes runtime UI message files into flat keys for aggregation, missing checks, exports, diagnostics, and plugin packaging.
- The runtime i18n API converts normalized flat keys into nested objects when returning data to the frontend.
- API-documentation bundles may use nested JSON or flat dotted keys, normalize to structured `core.*` and `plugins.*` keys, keep `en-US/apidoc` as an empty placeholder, and never translate `eg/example` values or generated entity metadata.
- API-documentation bundles may use host-owned `core.common.*` fallback keys for repeated standard response, pagination, and timestamp metadata; exact structural keys still take precedence.
- Runtime business files use semantic names such as `framework.json`, `menu.json`, `dict.json`, `config.json`, `error.json`, `artifact.json`, `job.json`, `notify.json`, `role.json`, and `public-frontend.json`. Do not use numeric filename prefixes.
- Host API-documentation files use `common.json` or `core-api-<module>.json`; plugin API-documentation files use `plugin-api-<module>.json` or `<plugin-id>-api-<module>.json`.

## Why JSON And Key Normalization

`JSON` is the canonical delivery format because it matches the existing frontend locale workflow, is easy to review and export through HTTP APIs, and can be embedded into dynamic plugin `Wasm` artifacts without extra conversion layers.

Nested JSON is the recommended file authoring format for runtime UI messages because it reduces repeated prefixes and keeps code review readable. Flat dotted keys are still accepted for small patches and gradual migrations.

Flat keys remain the canonical governance format because they keep resource comparison, missing-translation checks, exports, diagnostics, and plugin packaging simple and deterministic. When one locale JSON file mixes nested JSON and equivalent flat dotted keys, the flat dotted key wins so migrations stay explicit.

Example:

```json
{
  "framework": {
    "description": "AI-driven full-stack development framework"
  },
  "menu": {
    "dashboard": {
      "title": "Workbench"
    }
  },
  "plugin": {
    "org-center": {
      "name": "Organization Center"
    }
  }
}
```

## Key Naming Conventions

| Scope                       | Key pattern                                                     | Example                                   |
| --------------------------- | --------------------------------------------------------------- | ----------------------------------------- |
| Framework metadata          | `framework.<field>`                                             | `framework.description`                   |
| Menu title                  | `menu.<menu_key>.title`                                         | `menu.dashboard.title`                    |
| Dict type name              | `dict.<dict_type>.name`                                         | `dict.sys_normal_disable.name`            |
| Dict option label           | `dict.<dict_type>.<value>.label`                                | `dict.sys_normal_disable.1.label`         |
| Config metadata             | `config.<config_key>.name`                                      | `config.sys.account.captchaEnabled.name`  |
| Public frontend copy        | `publicFrontend.<group>.<field>`                                | `publicFrontend.login.title`              |
| Plugin metadata             | `plugin.<plugin_id>.name`                                       | `plugin.org-center.name`                  |
| Plugin description          | `plugin.<plugin_id>.description`                                | `plugin.org-center.description`           |
| Locale display name         | `locale.<locale>.name`                                          | `locale.en-US.name`                       |
| Locale native name          | `locale.<locale>.nativeName`                                    | `locale.en-US.nativeName`                 |
| Validation or error message | `validation.<module>.<field>.<rule>` or `error.<module>.<code>` | `validation.auth.login.username.required` |

Recommendations:

- Reuse stable business identifiers such as `menu_key`, `dict_type`, `config_key`, and `plugin_id`.
- Keep one semantic message on one key.
- Do not mix display hierarchy with storage hierarchy. The key is a stable identifier, not a UI tree.
- Prefer extending an existing prefix instead of inventing a parallel alias for the same concept.

## Delivery Workflow

1. Add or update the baseline locale files under `manifest/i18n/<locale>/`.
2. Update the default config `i18n.locales` list when the new locale should be enabled, ordered, renamed, or selected as the default.
3. Add plugin locale files under `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/` when a plugin contributes user-visible copy.
4. Add API-documentation locale files under `manifest/i18n/<locale>/apidoc/` and plugin-owned `manifest/i18n/<locale>/apidoc/` directories when API DTO source text changes.
5. Start the host and fetch `GET /api/v1/i18n/runtime/locales?lang=<locale>` to confirm the locale list and metadata.
6. Fetch `GET /api/v1/i18n/runtime/messages?lang=<locale>` to confirm the merged runtime result.
7. Fetch `/api.json?lang=<locale>` to confirm API-documentation localization.
8. Use `GET /api/v1/i18n/messages/missing?locale=<locale>` to detect keys that are still missing compared with the default locale.
9. Use `GET /api/v1/i18n/messages/diagnostics?locale=<locale>` to confirm whether the effective value comes from the host file or plugin file.
10. Use `GET /api/v1/i18n/messages/export?locale=<locale>` when a merged flat-key catalog is needed for offline review or resource-file maintenance.

## Runtime API Caching

`GET /api/v1/i18n/runtime/messages?lang=<locale>` returns the merged runtime bundle with an `ETag` header derived from the locale and runtime bundle version.

Clients should persist `{etag, messages, savedAt}` per locale and send `If-None-Match` on the next request. When the server returns `304 Not Modified`, the response body is empty and the client should reuse the persisted bundle.

Runtime bundle invalidation must always be scoped. Host file, source plugin, and dynamic plugin changes should invalidate only the affected locale, sector, or plugin. Full cache resets are reserved for process-level reloads and tests.

## Source-Text Namespaces

Some runtime keys are owned by code source text, such as built-in scheduler labels. The owning business package must register those prefixes through `i18n.RegisterSourceTextNamespace(prefix, reason)`.

The i18n foundation must not hard-code business prefixes. Missing-message checks only skip source-text-backed keys whose namespace has been registered by the owning module.

## Validation Rules

Before delivery, check the following items:

- Every enabled locale file is valid `JSON`.
- Every locale listed in the default config `i18n.locales` list has a matching `manifest/i18n/<locale>/` runtime directory.
- Every normalized message key is unique across direct runtime JSON files in the same locale directory.
- The target locale passes the missing-translation check against the default locale.
- Plugin-owned keys use the `plugin.<plugin_id>.` prefix unless the plugin is intentionally contributing shared framework metadata.
- New user-visible backend errors and validation messages use translation keys instead of hard-coded literals.

## Authoring Example

```json
{
  "framework": {
    "description": "AI-driven full-stack development framework"
  },
  "menu": {
    "system": {
      "title": "System Management",
      "users": {
        "title": "Users"
      }
    }
  },
  "config": {
    "sys": {
      "account": {
        "captchaEnabled": {
          "name": "Login Captcha"
        }
      }
    }
  },
  "publicFrontend": {
    "login": {
      "title": "Welcome back"
    }
  },
  "plugin": {
    "org-center": {
      "name": "Organization Center",
      "description": "Departments, posts, and hierarchy management"
    }
  }
}
```
