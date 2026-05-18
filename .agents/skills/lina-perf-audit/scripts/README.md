# LinaPro Performance Audit Scripts

This skill-bundled `scripts/` directory contains helper scripts for the manual `lina-perf-audit` skill.
They prepare a local audit-only backend runtime, install built-in plugins, and
restore the host logger configuration afterward.

These scripts are maintained inside `.agents/skills/lina-perf-audit/scripts/` and are intended to be run from the repository root.

## Scripts

- `setup-audit-env.sh`
  - Stops existing local services with `make stop`.
  - Backs up `apps/lina-core/manifest/config/config.yaml` logger settings to
    `<run-dir>/logger-backup.json`.
  - Temporarily patches `logger.path` to the run directory and `logger.file` to
    `server.log`.
  - Builds dynamic plugin Wasm artifacts, prepares embedded host assets, and
    starts only the backend service.
  - Waits for `http://127.0.0.1:8080/api/v1/health`.
  - Logs in as `admin/admin123`, writes the token to `<run-dir>/token.txt`, and
    records the `Trace-ID` check in `<run-dir>/trace-id-check.txt`.

- `prepare-builtin-plugins.sh`
  - Reads `<run-dir>/token.txt`.
  - Scans `apps/lina-plugins/*/plugin.yaml`.
  - Calls the host plugin APIs to sync, install, and enable every discovered
    built-in plugin.
  - Writes progress and failure details to `<run-dir>/plugins.json`.

- `scan-endpoints.sh`
  - Scans host and built-in plugin API DTOs.
  - Parses `g.Meta` / `gmeta.Meta` route metadata.
  - Writes `<run-dir>/catalog.json`.

- `probe-fixtures.sh`
  - Reads `<run-dir>/catalog.json` and the token.
  - Probes safe `GET` list/detail endpoints.
  - Writes `<run-dir>/fixtures.json` and fails if declared DTO routes are not
    accessible.

- `stress-fixture.sh`
  - Inserts audit-only stress data through a direct MySQL connection.
  - Uses idempotent writes and does not modify delivery SQL directories.
  - Writes `<run-dir>/stress-fixture.json`.

- `aggregate-reports.sh`
  - Reads Stage 1 reports from `<run-dir>/audits/*.md`.
  - Writes `<run-dir>/SUMMARY.md` and `<run-dir>/meta.json`.
  - Creates or updates persistent issue cards under root `perf-issues/`.
  - Regenerates `perf-issues/INDEX.md` with open and in-progress cards.

- `restore-audit-env.sh`
  - Restores `logger.path` and `logger.file` from
    `<run-dir>/logger-backup.json` when the backup exists.
  - Calls `make stop`.
  - Is safe to call more than once on success or failure paths.

## Usage

```bash
run_id="$(date +%Y%m%d-%H%M%S)"
run_dir="temp/lina-perf-audit/${run_id}"

bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id "${run_id}"
bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/probe-fixtures.sh --run-dir "${run_dir}"

# After sub agents write reports under "${run_dir}/audits/".
bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir "${run_dir}"

# Always restore after the audit run, including failure paths.
bash .agents/skills/lina-perf-audit/scripts/restore-audit-env.sh --run-dir "${run_dir}"
```

The scripts use only Bash, `curl`, Go tooling, MySQL CLI, and Python 3
standard-library JSON support. They do not depend on `jq` or PyYAML and do not
modify delivery SQL files or application source code.
