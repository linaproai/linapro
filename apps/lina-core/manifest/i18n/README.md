# Runtime I18N Resources

This directory stores the delivery baseline runtime message bundles for `LinaPro`.

The host loads `manifest/i18n/<locale>.json` as the project-level baseline, merges it with enabled plugin resources and database overrides, and finally exposes the effective result through the runtime i18n APIs.

## Directory Contract

| Path                                                        | Purpose                            |
| ----------------------------------------------------------- | ---------------------------------- |
| `manifest/i18n/zh-CN.json`                                  | Simplified Chinese baseline bundle |
| `manifest/i18n/en-US.json`                                  | English baseline bundle            |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>.json` | Plugin-owned locale bundle         |

Rules:

- The filename must use the canonical locale code, for example `zh-CN.json` or `en-US.json`.
- The host only treats top-level `manifest/i18n/<locale>.json` files as runtime locale bundles.
- Runtime messages are maintained with flat keys only.
- The runtime i18n API converts flat keys into nested objects only when returning data to the frontend.

## Why JSON And Flat Keys

`JSON` is the canonical delivery format because it matches the existing frontend locale workflow, is easy to import and export through HTTP APIs, and can be embedded into dynamic plugin `Wasm` artifacts without extra conversion layers.

Flat keys are the canonical management format because they keep backend storage, database overrides, missing-translation checks, and plugin packaging simple and deterministic.

Example:

```json
{
  "framework.description": "AI-driven full-stack development framework",
  "menu.dashboard.title": "Workbench",
  "plugin.org-center.name": "Organization Center"
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

1. Add or update the baseline locale files in `manifest/i18n/`.
2. Add plugin locale files in `apps/lina-plugins/<plugin-id>/manifest/i18n/` when a plugin contributes user-visible copy.
3. Start the host and fetch `GET /api/v1/i18n/runtime/messages?lang=<locale>` to confirm the merged runtime result.
4. Use `GET /api/v1/i18n/messages/missing?locale=<locale>` to detect keys that are still missing compared with the default locale.
5. Use `GET /api/v1/i18n/messages/diagnostics?locale=<locale>` to confirm whether the effective value comes from the host file, plugin file, or database override.
6. If an online hotfix is needed, import database overrides through `POST /api/v1/i18n/messages/import`, then export the effective result through `GET /api/v1/i18n/messages/export` for later code sync.

## Validation Rules

Before delivery, check the following items:

- Every enabled locale file is valid `JSON`.
- Every message key is unique inside one locale file.
- The target locale passes the missing-translation check against the default locale.
- Plugin-owned keys use the `plugin.<plugin_id>.` prefix unless the plugin is intentionally contributing shared framework metadata.
- New user-visible backend errors and validation messages use translation keys instead of hard-coded literals.

## Business Content Contract

`sys_i18n_content` is the shared storage model for business records that need multilingual titles, summaries, descriptions, or body content.

Use the following anchor contract when one business module integrates with it:

| Field           | Rule                                                   | Example                    |
| --------------- | ------------------------------------------------------ | -------------------------- |
| `business_type` | Stable module-level identifier, not a translated label | `notice`, `cms_article`    |
| `business_id`   | Stable primary key or immutable business code          | `42`, `article-homepage`   |
| `field`         | Stable field name from the business aggregate          | `title`, `summary`, `body` |
| `locale`        | Canonical runtime locale code                          | `zh-CN`, `en-US`           |
| `content_type`  | One of `plain`, `markdown`, `html`, `json`             | `markdown`                 |

Recommended access pattern:

1. The business table keeps its source-language field as the canonical fallback value.
2. The business service queries `sys_i18n_content` by `business_type + business_id + field + locale`.
3. If the requested locale is missing, the service falls back to the runtime default locale.
4. If the default locale is also missing, the service falls back to the business table's own source field.

Caching rules:

- Cache by the full anchor tuple `business_type + business_id + field`.
- Store locale variants as the cache payload, not one cache entry per locale, so one invalidation refreshes the whole record.
- Invalidate the anchor cache immediately after create, update, delete, import, or publish actions that touch multilingual content.
- Do not cache only the miss forever without an explicit invalidation strategy; otherwise later writes cannot be observed.

Scope guidance:

- Use `sys_i18n_message` for reusable UI copy or metadata labels.
- Use `sys_i18n_content` only for record-bound business content that belongs to one concrete business entity.
- Keep attachments or rich media references in the business table itself; `sys_i18n_content` stores the localized text payload only.

## Authoring Example

```json
{
  "framework.description": "AI-driven full-stack development framework",
  "menu.system.title": "System Management",
  "menu.system.users.title": "Users",
  "config.sys.account.captchaEnabled.name": "Login Captcha",
  "publicFrontend.login.title": "Welcome back",
  "plugin.org-center.name": "Organization Center",
  "plugin.org-center.description": "Departments, posts, and hierarchy management"
}
```
