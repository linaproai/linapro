#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir <dir>

Adds idempotent audit-only stress data through MySQL and writes
<run-dir>/stress-fixture.json. Run this script from the repository root.
USAGE
}

RUN_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir)
      if [[ $# -lt 2 ]]; then
        echo "missing value for --run-dir" >&2
        exit 2
      fi
      RUN_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$RUN_DIR" ]]; then
  echo "--run-dir is required" >&2
  usage >&2
  exit 2
fi

if [[ ! -f "apps/lina-core/manifest/config/config.yaml" || ! -d "apps/lina-core/manifest/sql" ]]; then
  echo "stress-fixture.sh must be run from the LinaPro repository root" >&2
  exit 1
fi

mkdir -p "$RUN_DIR"

python3 - "$RUN_DIR" <<'PY'
import hashlib
import json
import re
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

run_dir = Path(sys.argv[1])
root = Path.cwd()
output_path = run_dir / "stress-fixture.json"
manifest_dirs = [
    root / "apps/lina-core/manifest/sql",
    root / "apps/lina-core/manifest/sql/mock-data",
]


def now_utc() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def digest_dirs(paths: list[Path]) -> str:
    hasher = hashlib.sha256()
    for base in paths:
        if not base.exists():
            continue
        for path in sorted(item for item in base.rglob("*") if item.is_file()):
            relative = path.relative_to(root).as_posix()
            hasher.update(relative.encode("utf-8") + b"\0")
            hasher.update(hashlib.sha256(path.read_bytes()).hexdigest().encode("ascii") + b"\0")
    return hasher.hexdigest()


def parse_link() -> dict:
    config_path = root / "apps/lina-core/manifest/config/config.yaml"
    lines = config_path.read_text(encoding="utf-8").splitlines()
    in_database = False
    in_default = False
    link = ""
    for raw_line in lines:
        line = raw_line.split("#", 1)[0].rstrip()
        if not line.strip():
            continue
        if re.match(r"^database:\s*$", line):
            in_database = True
            in_default = False
            continue
        if in_database and re.match(r"^[A-Za-z0-9_-]+:\s*", line):
            break
        if in_database and re.match(r"^\s{2}default:\s*$", line):
            in_default = True
            continue
        if in_database and in_default and re.match(r"^\s{2}[A-Za-z0-9_-]+:\s*", line):
            break
        if in_database and in_default:
            match = re.match(r'^\s{4}link:\s*(.*?)\s*$', line)
            if match:
                link = match.group(1).strip()
                if len(link) >= 2 and link[0] == link[-1] and link[0] in ("'", '"'):
                    link = link[1:-1]
                break
    if not link:
        raise RuntimeError("database.default.link not found in config.yaml")
    dsn = re.match(r'^mysql:([^:]+):([^@]*)@tcp\(([^:)]+):([0-9]+)\)/([^?]+)', link)
    if not dsn:
        raise RuntimeError(f"unsupported MySQL link format: {link}")
    return {
        "driver": "mysql",
        "user": dsn.group(1),
        "password": dsn.group(2),
        "host": dsn.group(3),
        "port": dsn.group(4),
        "database": dsn.group(5),
        "link": link,
    }


def values_rows(rows: list[tuple]) -> str:
    rendered = []
    for row in rows:
        parts = []
        for value in row:
            if value is None:
                parts.append("NULL")
            elif isinstance(value, (int, float)):
                parts.append(str(value))
            else:
                escaped = str(value).replace("\\", "\\\\").replace("'", "''")
                parts.append("'" + escaped + "'")
        rendered.append("(" + ", ".join(parts) + ")")
    return ",\n".join(rendered)


def sequence_select(limit: int, alias: str = "seq") -> str:
    rows = " UNION ALL ".join(f"SELECT {index} AS n" for index in range(1, limit + 1))
    return f"({rows}) AS {alias}"


class Mysql:
    def __init__(self, cfg: dict):
        mysql_bin = shutil.which("mysql")
        if not mysql_bin:
            raise RuntimeError("mysql CLI is required for direct stress-fixture inserts")
        self.base_cmd = [
            mysql_bin,
            "--protocol=tcp",
            "-h", cfg["host"],
            "-P", cfg["port"],
            "-u", cfg["user"],
            f'--password={cfg["password"]}',
            "--batch",
            "--raw",
            "--skip-column-names",
            cfg["database"],
        ]

    def run(self, sql: str) -> str:
        completed = subprocess.run(
            self.base_cmd,
            input=sql,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )
        if completed.returncode != 0:
            raise RuntimeError(completed.stderr.strip() or "mysql command failed")
        return completed.stdout.strip()

    def scalar(self, sql: str) -> str:
        out = self.run(sql)
        return out.splitlines()[0].strip() if out.splitlines() else ""

    def count(self, table: str, where: str = "1=1") -> int:
        value = self.scalar(f"SELECT COUNT(*) FROM `{table}` WHERE {where};")
        return int(value or "0")

    def table_exists(self, table: str) -> bool:
        escaped = table.replace("'", "''")
        value = self.scalar(
            "SELECT COUNT(*) FROM information_schema.tables "
            f"WHERE table_schema = DATABASE() AND table_name = '{escaped}';"
        )
        return int(value or "0") > 0


def record_result(results: list, mysql: Mysql, resource: str, table: str, target: int, marker_where: str, insert_sql: str):
    if not mysql.table_exists(table):
        results.append({
            "resource": resource,
            "table": table,
            "status": "skipped",
            "reason": "table not found",
            "targetRows": target,
        })
        return
    before_total = mysql.count(table)
    before_stress = mysql.count(table, marker_where)
    mysql.run(insert_sql)
    after_total = mysql.count(table)
    after_stress = mysql.count(table, marker_where)
    results.append({
        "resource": resource,
        "table": table,
        "status": "ok",
        "targetRows": target,
        "beforeTotalRows": before_total,
        "afterTotalRows": after_total,
        "beforeStressRows": before_stress,
        "afterStressRows": after_stress,
    })


def insert_sys_user(mysql: Mysql, results: list):
    rows = []
    password = "$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe"
    for n in range(1, 101):
        rows.append((
            f"perf_audit_user_{n:03d}",
            password,
            f"Perf Audit User {n:03d}",
            f"perf_audit_user_{n:03d}@example.invalid",
            f"1390000{n:04d}",
            n % 3,
            1,
            "lina-perf-audit stress fixture",
        ))
    values = []
    for row in rows:
        values.append(row + ("__NOW__", "__NOW__"))
    value_sql = values_rows(values).replace("'__NOW__'", "NOW()")
    sql = (
        "INSERT IGNORE INTO sys_user "
        "(username, password, nickname, email, phone, sex, status, remark, created_at, updated_at) VALUES\n"
        + value_sql
        + ";"
    )
    record_result(results, mysql, "sys_user", "sys_user", 100, "username LIKE 'perf_audit_user_%'", sql)


def insert_dict(mysql: Mysql, results: list):
    type_rows = [
        (f"Perf Audit Type {n:03d}", f"perf_audit_type_{n:03d}", 1, 0, "lina-perf-audit stress fixture")
        for n in range(1, 51)
    ]
    type_sql = (
        "INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at) VALUES\n"
        + values_rows(row + ("__NOW__", "__NOW__") for row in type_rows).replace("'__NOW__'", "NOW()")
        + ";"
    )
    record_result(results, mysql, "sys_dict_type", "sys_dict_type", 50, "type LIKE 'perf_audit_type_%'", type_sql)

    data_rows = []
    for type_index in range(1, 51):
        for value_index in range(1, 21):
            data_rows.append((
                f"perf_audit_type_{type_index:03d}",
                f"Perf Audit Label {type_index:03d}-{value_index:03d}",
                f"perf_audit_value_{value_index:03d}",
                value_index,
                "primary",
                "",
                1,
                0,
                "lina-perf-audit stress fixture",
            ))
    data_sql = (
        "INSERT IGNORE INTO sys_dict_data "
        "(dict_type, label, value, sort, tag_style, css_class, status, is_builtin, remark, created_at, updated_at) VALUES\n"
        + values_rows(row + ("__NOW__", "__NOW__") for row in data_rows).replace("'__NOW__'", "NOW()")
        + ";"
    )
    record_result(results, mysql, "sys_dict_data", "sys_dict_data", 1000, "dict_type LIKE 'perf_audit_type_%'", data_sql)


def insert_role_menu(mysql: Mysql, results: list):
    role_rows = [
        (f"Perf Audit Role {n:03d}", f"perf_audit_role_{n:03d}", n, 1, 1, "lina-perf-audit stress fixture")
        for n in range(1, 31)
    ]
    role_sql = (
        "INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at) VALUES\n"
        + values_rows(row + ("__NOW__", "__NOW__") for row in role_rows).replace("'__NOW__'", "NOW()")
        + ";"
    )
    record_result(results, mysql, "sys_role", "sys_role", 30, "`key` LIKE 'perf_audit_role_%'", role_sql)

    menu_rows = []
    for n in range(1, 51):
        menu_rows.append((
            0,
            f"perf_audit_menu_{n:03d}",
            f"Perf Audit Menu {n:03d}",
            f"/perf-audit/{n:03d}",
            "system/perf-audit",
            f"perf:audit:{n:03d}",
            "lucide:gauge",
            "M",
            n,
            0,
            1,
            0,
            0,
            "",
            "lina-perf-audit stress fixture",
        ))
    menu_sql = (
        "INSERT IGNORE INTO sys_menu "
        "(parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at) VALUES\n"
        + values_rows(row + ("__NOW__", "__NOW__") for row in menu_rows).replace("'__NOW__'", "NOW()")
        + ";"
    )
    record_result(results, mysql, "sys_menu", "sys_menu", 50, "menu_key LIKE 'perf_audit_menu_%'", menu_sql)


def insert_user_messages(mysql: Mysql, results: list):
    if not mysql.table_exists("sys_notify_message") or not mysql.table_exists("sys_notify_delivery"):
        results.append({
            "resource": "sys_user_msg",
            "table": "sys_user_msg",
            "status": "skipped",
            "reason": "legacy sys_user_msg table not found and current notify tables are incomplete",
            "targetRows": 100,
        })
        return
    message_sql = (
        "INSERT IGNORE INTO sys_notify_message "
        "(plugin_id, source_type, source_id, category_code, title, content, payload_json, sender_user_id, created_at)\n"
        f"SELECT '', 'system', CONCAT('perf-audit-', LPAD(seq.n, 3, '0')), 'other', "
        f"CONCAT('Perf Audit Message ', LPAD(seq.n, 3, '0')), "
        f"CONCAT('lina-perf-audit stress fixture message ', seq.n), '{{}}', 1, NOW()\n"
        f"FROM {sequence_select(100)}\n"
        "WHERE NOT EXISTS ("
        "SELECT 1 FROM sys_notify_message m "
        "WHERE m.source_type = 'system' AND m.source_id = CONCAT('perf-audit-', LPAD(seq.n, 3, '0'))"
        ");"
    )
    delivery_sql = (
        "INSERT IGNORE INTO sys_notify_delivery "
        "(message_id, channel_key, channel_type, recipient_type, recipient_key, user_id, delivery_status, is_read, created_at, updated_at)\n"
        "SELECT m.id, 'inbox', 'inbox', 'user', '1', 1, 1, 0, NOW(), NOW()\n"
        "FROM sys_notify_message m\n"
        "WHERE m.source_type = 'system' AND m.source_id LIKE 'perf-audit-%'\n"
        "AND NOT EXISTS ("
        "SELECT 1 FROM sys_notify_delivery d "
        "WHERE d.message_id = m.id AND d.user_id = 1 AND d.channel_key = 'inbox'"
        ");"
    )
    before_total = mysql.count("sys_notify_delivery")
    before_stress = mysql.count(
        "sys_notify_message",
        "source_type = 'system' AND source_id LIKE 'perf-audit-%'",
    )
    mysql.run(message_sql)
    mysql.run(delivery_sql)
    after_total = mysql.count("sys_notify_delivery")
    after_stress = mysql.count(
        "sys_notify_message",
        "source_type = 'system' AND source_id LIKE 'perf-audit-%'",
    )
    results.append({
        "resource": "sys_user_msg",
        "table": "sys_notify_message/sys_notify_delivery",
        "status": "ok",
        "targetRows": 100,
        "beforeTotalRows": before_total,
        "afterTotalRows": after_total,
        "beforeStressRows": before_stress,
        "afterStressRows": after_stress,
    })


def insert_jobs(mysql: Mysql, results: list):
    job_sql = (
        "INSERT IGNORE INTO sys_job "
        "(group_id, name, description, task_type, handler_ref, params, timeout_seconds, shell_cmd, work_dir, env, "
        "cron_expr, timezone, scope, concurrency, max_concurrency, max_executions, status, is_builtin, seed_version, created_by, updated_by, created_at, updated_at)\n"
        "SELECT COALESCE((SELECT id FROM sys_job_group WHERE code = 'default' AND deleted_at IS NULL LIMIT 1), 0), "
        "CONCAT('Perf Audit Job ', LPAD(seq.n, 3, '0')), "
        "'lina-perf-audit stress fixture', 'shell', '', NULL, 300, 'echo perf-audit', '', NULL, "
        "'# */5 * * * *', 'UTC', 'master_only', 'singleton', 1, 0, 'disabled', 0, 0, 1, 1, NOW(), NOW()\n"
        f"FROM {sequence_select(50)};"
    )
    record_result(results, mysql, "sys_job", "sys_job", 50, "name LIKE 'Perf Audit Job %'", job_sql)

    log_sql = (
        "INSERT IGNORE INTO sys_job_log "
        "(job_id, job_snapshot, node_id, `trigger`, params_snapshot, start_at, end_at, duration_ms, status, err_msg, result_json, created_at)\n"
        "SELECT COALESCE((SELECT MIN(id) FROM sys_job WHERE name LIKE 'Perf Audit Job %'), 0), "
        "CONCAT('{\"name\":\"Perf Audit Job ', LPAD(seq.n, 3, '0'), '\"}'), "
        "'perf-audit-node', 'manual', '{}', DATE_SUB(NOW(), INTERVAL seq.n MINUTE), "
        "DATE_SUB(NOW(), INTERVAL seq.n MINUTE), seq.n, 'success', '', "
        "CONCAT('{\"perfAudit\":', seq.n, '}'), NOW()\n"
        f"FROM {sequence_select(200)}\n"
        "WHERE NOT EXISTS ("
        "SELECT 1 FROM sys_job_log l WHERE l.result_json = CONCAT('{\"perfAudit\":', seq.n, '}')"
        ");"
    )
    record_result(results, mysql, "sys_job_log", "sys_job_log", 200, "result_json LIKE '{\"perfAudit\":%'", log_sql)


def insert_org_plugin(mysql: Mysql, results: list):
    dept_sql = (
        "INSERT IGNORE INTO plugin_org_center_dept "
        "(parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at) VALUES\n"
        + values_rows((
            0,
            "",
            f"Perf Audit Dept {n:03d}",
            f"perf_audit_dept_{n:03d}",
            n,
            0,
            f"0210000{n:04d}",
            f"perf_audit_dept_{n:03d}@example.invalid",
            1,
            "lina-perf-audit stress fixture",
            "__NOW__",
            "__NOW__",
        ) for n in range(1, 101)).replace("'__NOW__'", "NOW()")
        + ";"
    )
    record_result(results, mysql, "plugin:sys_dept", "plugin_org_center_dept", 100, "code LIKE 'perf_audit_dept_%'", dept_sql)

    post_sql = (
        "INSERT IGNORE INTO plugin_org_center_post "
        "(dept_id, code, name, sort, status, remark, created_at, updated_at)\n"
        "SELECT COALESCE((SELECT MIN(id) FROM plugin_org_center_dept WHERE code LIKE 'perf_audit_dept_%'), 0), "
        "CONCAT('perf_audit_post_', LPAD(seq.n, 3, '0')), "
        "CONCAT('Perf Audit Post ', LPAD(seq.n, 3, '0')), seq.n, 1, "
        "'lina-perf-audit stress fixture', NOW(), NOW()\n"
        f"FROM {sequence_select(50)};"
    )
    record_result(results, mysql, "plugin:sys_post", "plugin_org_center_post", 50, "code LIKE 'perf_audit_post_%'", post_sql)


def insert_notice_plugin(mysql: Mysql, results: list):
    notice_sql = (
        "INSERT IGNORE INTO plugin_content_notice "
        "(title, type, content, file_ids, status, remark, created_by, updated_by, created_at, updated_at)\n"
        "SELECT CONCAT('Perf Audit Notice ', LPAD(seq.n, 3, '0')), "
        "IF(MOD(seq.n, 2) = 0, 2, 1), "
        "CONCAT('lina-perf-audit stress fixture notice ', seq.n), '', 1, "
        "'lina-perf-audit stress fixture', 1, 1, NOW(), NOW()\n"
        f"FROM {sequence_select(100)}\n"
        "WHERE NOT EXISTS ("
        "SELECT 1 FROM plugin_content_notice n "
        "WHERE n.title = CONCAT('Perf Audit Notice ', LPAD(seq.n, 3, '0'))"
        ");"
    )
    record_result(results, mysql, "plugin:sys_notice", "plugin_content_notice", 100, "title LIKE 'Perf Audit Notice %'", notice_sql)


def insert_loginlog_plugin(mysql: Mysql, results: list):
    login_sql = (
        "INSERT IGNORE INTO plugin_monitor_loginlog "
        "(user_name, status, ip, browser, os, msg, login_time)\n"
        "SELECT CONCAT('perf_audit_user_', LPAD(seq.n, 3, '0')), "
        "IF(MOD(seq.n, 5) = 0, 1, 0), "
        "CONCAT('10.96.', FLOOR(seq.n / 255), '.', MOD(seq.n, 255)), "
        "'Chrome', 'macOS', 'lina-perf-audit stress fixture', "
        "DATE_SUB(NOW(), INTERVAL seq.n MINUTE)\n"
        f"FROM {sequence_select(200)}\n"
        "WHERE NOT EXISTS ("
        "SELECT 1 FROM plugin_monitor_loginlog l "
        "WHERE l.user_name = CONCAT('perf_audit_user_', LPAD(seq.n, 3, '0')) "
        "AND l.ip = CONCAT('10.96.', FLOOR(seq.n / 255), '.', MOD(seq.n, 255))"
        ");"
    )
    record_result(results, mysql, "plugin:sys_login_log", "plugin_monitor_loginlog", 200, "user_name LIKE 'perf_audit_user_%'", login_sql)


def main():
    before_hash = digest_dirs(manifest_dirs)
    cfg = parse_link()
    mysql = Mysql(cfg)
    results = []

    insert_sys_user(mysql, results)
    insert_dict(mysql, results)
    insert_role_menu(mysql, results)
    insert_user_messages(mysql, results)
    insert_jobs(mysql, results)
    insert_org_plugin(mysql, results)
    insert_notice_plugin(mysql, results)
    insert_loginlog_plugin(mysql, results)

    after_hash = digest_dirs(manifest_dirs)
    status = "ok" if before_hash == after_hash else "failed"
    payload = {
        "schemaVersion": 1,
        "generatedAt": now_utc(),
        "status": status,
        "database": {
            "host": cfg["host"],
            "port": cfg["port"],
            "name": cfg["database"],
            "user": cfg["user"],
        },
        "manifestSqlHashBefore": before_hash,
        "manifestSqlHashAfter": after_hash,
        "resources": results,
    }
    if status != "ok":
        payload["error"] = "apps/lina-core/manifest/sql hash changed during stress-fixture execution"
    output_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(str(output_path))
    if status != "ok":
        sys.exit(1)


try:
    main()
except Exception as exc:
    failure = {
        "schemaVersion": 1,
        "generatedAt": now_utc(),
        "status": "failed",
        "error": str(exc),
    }
    output_path.write_text(json.dumps(failure, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(str(output_path), file=sys.stderr)
    print(str(exc), file=sys.stderr)
    sys.exit(1)
PY
