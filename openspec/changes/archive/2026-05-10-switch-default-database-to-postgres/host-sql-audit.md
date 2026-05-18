# Host SQL PostgreSQL Rewrite Audit

Scope: host SQL only under `apps/lina-core/manifest/sql/**`.

## Idempotency Basis

`ON CONFLICT DO NOTHING` is used only where a real primary key or unique index can trigger the conflict:

- `sys_user`: `uk_sys_user_username(username)`
- `sys_dict_type`: `uk_sys_dict_type_type("type")`
- `sys_dict_data`: `uk_sys_dict_data_dict_type_value(dict_type, "value")`
- `sys_config`: `uk_sys_config_key("key")`
- `sys_role`: `uk_sys_role_key("key")`
- `sys_menu`: `uk_sys_menu_menu_key(menu_key)`
- `sys_role_menu`: primary key `(role_id, menu_id)`
- `sys_user_role`: primary key `(user_id, role_id)`
- `sys_online_session`: primary key `(token_id)`
- `sys_job_group`: `uk_sys_job_group_code(code)`
- `sys_job`: `uk_sys_job_group_id_name(group_id, name)`
- `sys_notify_channel`: `uk_sys_notify_channel_channel_key(channel_key)`

## Static History Mock Decisions

The following mock tables model notification and execution history. They do not use `ON CONFLICT DO NOTHING` or artificial unique keys because that would constrain legitimate production writes. Their static mock rows instead use exact `WHERE NOT EXISTS` checks over the demonstration identity fields so repeated mock loading leaves the same final rows:

- `sys_notify_message`
- `sys_notify_delivery`
- `sys_job_log`

Repeated host mock loading must keep these table row counts stable for the bundled static records.

## Volatile Table Decision

The host SQL currently defines `sys_online_session`, `sys_locker`, and `sys_kv_cache`. All three are PostgreSQL normal persistent tables with no engine, unlogged, or temporary-table clause.

No host SQL or host code reference to `sys_session` was found in this branch. This rewrite did not invent a new `sys_session` table because the allowed ownership is host SQL only and the active host session truth table is `sys_online_session`.
