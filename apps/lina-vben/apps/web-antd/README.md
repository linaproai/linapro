# LinaPro Web Workspace

`apps/lina-vben/apps/web-antd` is the default management workspace of `LinaPro`. It is the reference Vue 3 + `Vben 5` application that consumes host APIs, renders built-in system modules, and hosts plugin-aware pages.

## Responsibilities

- Render the host management workspace used for system governance and plugin administration.
- Consume host APIs under `/api/v1` through the shared request client and runtime configuration layer.
- Expose stable routes, pages, forms, and tables that mirror host-side modules.
- Provide the default UI entry for scheduled-job management, including job groups, persisted jobs, and execution logs.

## Scheduled Job Pages

The scheduled-job iteration adds three management entries under `/system`.

| Route | Page | Purpose |
| --- | --- | --- |
| `/system/job` | Task management | Create, edit, enable, disable, trigger, and reset persisted jobs |
| `/system/job-group` | Group management | Manage job groups and protect the default group |
| `/system/job-log` | Execution logs | Inspect execution history, view details, clear logs, and cancel running instances |

The job form supports both `Handler` and `Shell` task types. `Shell` options are hidden when the public frontend runtime config reports that shell execution is disabled or unsupported, or when the current user lacks `system:job:shell`.

## Key Directories

```text
src/api/            Host API clients
src/adapter/        Form and table adapters
src/router/         Workspace routes
src/runtime/        Public runtime config loading
src/views/          Page implementations
```

## Common Commands

```bash
pnpm install
pnpm -F @lina/web-antd dev --host 127.0.0.1
pnpm -F @lina/web-antd build
```

## Related Entry Points

- `src/router/routes/modules/system.ts`: system-management route registration.
- `src/views/system/job/`: scheduled-job pages and task form.
- `src/views/system/job-group/`: job-group list and modal.
- `src/views/system/job-log/`: execution-log list and detail modal.
- `src/runtime/public-frontend.ts`: public runtime capability loading for shell support and UI branding.
