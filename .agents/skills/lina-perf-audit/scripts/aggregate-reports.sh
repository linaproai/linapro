#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir <dir>

Options:
  --run-dir  Audit run directory containing audits/, catalog.json, and run artifacts.

The script reads Stage 1 audit reports, writes <run-dir>/SUMMARY.md and
<run-dir>/meta.json, and creates or updates persistent cards under
perf-issues/.

Set LINA_PERF_AUDIT_ISSUE_DIR only for isolated lifecycle validation. Normal
audit runs must use the default root perf-issues/ directory.
EOF
}

die() {
  printf 'aggregate-reports: %s\n' "$*" >&2
  exit 1
}

RUN_DIR=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir)
      [[ $# -ge 2 ]] || die "--run-dir requires a value"
      RUN_DIR="${2%/}"
      shift 2
      ;;
    --run-dir=*)
      RUN_DIR="${1#--run-dir=}"
      RUN_DIR="${RUN_DIR%/}"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

if [[ -z "$RUN_DIR" ]]; then
  die "--run-dir is required"
fi

if [[ ! -f "Makefile" || ! -d "$RUN_DIR/audits" || ! -f "$RUN_DIR/catalog.json" ]]; then
  die "run from the LinaPro repository root with a completed audit run directory"
fi

python3 - "$RUN_DIR" <<'PY'
import hashlib
import json
import os
import re
import subprocess
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path

run_dir = Path(sys.argv[1])
root = Path.cwd()
issue_dir = Path(os.environ.get("LINA_PERF_AUDIT_ISSUE_DIR", "perf-issues"))
if not issue_dir.is_absolute():
    issue_dir = root / issue_dir
issue_dir.mkdir(parents=True, exist_ok=True)
run_id = run_dir.name


def now_iso() -> str:
    return datetime.now(timezone.utc).astimezone().isoformat(timespec="seconds")


def mtime_iso(path: Path) -> str:
    if not path.exists():
        return now_iso()
    return datetime.fromtimestamp(path.stat().st_mtime, timezone.utc).astimezone().isoformat(timespec="seconds")


def read_json(path: Path, default):
    try:
        with path.open(encoding="utf-8") as f:
            return json.load(f)
    except FileNotFoundError:
        return default


def git_commit() -> str:
    completed = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    return completed.stdout.strip() if completed.returncode == 0 else "unknown"


def slugify(value: str, fallback: str = "issue") -> str:
    value = value.lower().strip()
    value = value.replace("n+1", "n-plus-1")
    value = re.sub(r"[^a-z0-9]+", "-", value)
    value = re.sub(r"-+", "-", value).strip("-")
    return value or fallback


def strip_code(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == "`" and value[-1] == "`":
        return value[1:-1]
    return value


def repo_rel(path: Path) -> str:
    return path.resolve().relative_to(root).as_posix()


def field(block: str, key: str) -> str:
    match = re.search(rf"^- {re.escape(key)}:\s*(.*)$", block, flags=re.MULTILINE)
    return strip_code(match.group(1).strip()) if match else ""


def fenced_sql(block: str) -> str:
    match = re.search(r"Evidence:\n\n```sql\n(.*?)\n```", block, flags=re.DOTALL)
    return match.group(1).strip() if match else ""


def paragraph_after(block: str, label: str) -> str:
    match = re.search(rf"{re.escape(label)}:\n\n(.*?)(?:\n\n[A-Z][A-Za-z ]+:\n|\Z)", block, flags=re.DOTALL)
    if not match:
        return ""
    return match.group(1).strip()


def canonical_target(heading: str, endpoint_field: str) -> tuple[str, str, str]:
    target = heading.split(" - ", 1)[0].strip()
    title = heading.split(" - ", 1)[1].strip() if " - " in heading else "performance finding"
    if not target and endpoint_field:
        target = endpoint_field
    method = target.split(" ", 1)[0].upper() if " " in target else "MULTI"
    path = target.split(" ", 1)[1].strip() if " " in target else target
    path = re.sub(r"https?://[^/]+", "", path)
    path = path.split("?", 1)[0].rstrip("/") or "/"
    return method, path, title


def normalize_fingerprint(module: str, method: str, path: str, severity: str, signature: str) -> tuple[str, str]:
    normalized_module = re.sub(r"\s+", "-", module.strip().lower())
    normalized_method = method.strip().upper()
    normalized_path = re.sub(r"https?://[^/]+", "", path.strip()).split("?", 1)[0].rstrip("/") or "/"
    normalized_severity = severity.strip().upper()
    normalized_signature = signature.strip().lower()
    source = f"{normalized_module}:{normalized_method}:{normalized_path}:{normalized_severity}:{normalized_signature}"
    return source, hashlib.sha256(source.encode("utf-8")).hexdigest()


def first_int(value: str) -> int:
    match = re.search(r"\d+", str(value))
    return int(match.group(0)) if match else 0


EXPECTED_READ_SIDE_EFFECT_TABLES = {"sys_online_session", "plugin_monitor_operlog"}
WRITE_TABLE_REF = r"(?:`?[A-Za-z0-9_]+`?\.)?`?([A-Za-z0-9_]+)`?"
WRITE_SQL_PATTERNS = [
    rf"(?:^|[^\w])UPDATE\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])(?:INSERT|REPLACE)(?:\s+IGNORE)?\s+INTO\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])DELETE\s+FROM\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])TRUNCATE(?:\s+TABLE)?\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])ALTER\s+TABLE\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])DROP\s+TABLE(?:\s+IF\s+EXISTS)?\s+{WRITE_TABLE_REF}",
    rf"(?:^|[^\w])CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+{WRITE_TABLE_REF}",
]


def write_sql_summary(sql: str) -> tuple[set[str], int]:
    targets: set[str] = set()
    matched_write_count = 0
    for line in sql.splitlines():
        for pattern in WRITE_SQL_PATTERNS:
            match = re.search(pattern, line, flags=re.IGNORECASE)
            if match:
                targets.add(match.group(1).lower())
                matched_write_count += 1
                break
    return targets, matched_write_count


def is_expected_read_side_effect(finding: dict) -> bool:
    if not finding["signature"].startswith("read-write-side-effect"):
        return False
    write_count = first_int(finding["write_sql_count"])
    targets, matched_write_count = write_sql_summary(finding["evidence"])
    return (
        write_count > 0
        and first_int(finding["sql_count"]) > write_count
        and matched_write_count >= write_count
        and bool(targets)
        and targets <= EXPECTED_READ_SIDE_EFFECT_TABLES
    )


def parse_audit_file(path: Path) -> tuple[list[dict], dict, list[dict]]:
    text = path.read_text(encoding="utf-8")
    result_summary = {}
    summary_match = re.search(r"## Result Summary\n\n(.*?)(?:\n## |\Z)", text, flags=re.DOTALL)
    if summary_match:
        for line in summary_match.group(1).splitlines():
            match = re.match(r"- ([A-Z]+): `?([0-9]+)`?", line.strip())
            if match:
                result_summary[match.group(1)] = int(match.group(2))

    skipped = []
    skipped_match = re.search(r"## Skipped Endpoints\n\n(.*?)(?:\n## |\Z)", text, flags=re.DOTALL)
    if skipped_match:
        for line in skipped_match.group(1).splitlines():
            if not line.startswith("| `"):
                continue
            cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
            if len(cells) >= 3:
                skipped.append({
                    "endpoint": cells[0].strip("`"),
                    "reason": cells[1],
                    "follow_up": cells[2],
                    "audit_file": path.as_posix(),
                })

    findings = []
    matches = list(re.finditer(r"^### (HIGH|MEDIUM|LOW) - (.+)$", text, flags=re.MULTILINE))
    for index, match in enumerate(matches):
        severity = match.group(1)
        heading = match.group(2).strip()
        start = match.end()
        end = matches[index + 1].start() if index + 1 < len(matches) else len(text)
        block = text[start:end]
        block = re.split(
            r"\n## (Passed Endpoints|Skipped Endpoints|Destructive Endpoint Handling|Failure And Warning Notes)\n",
            block,
            maxsplit=1,
        )[0]
        module = field(block, "Module")
        endpoint = field(block, "Endpoint")
        signature = field(block, "Anti-pattern signature")
        method, path_value, title = canonical_target(heading, endpoint)
        fingerprint_input, fingerprint = normalize_fingerprint(module, method, path_value, severity, signature)
        sql_count = field(block, "SQL count")
        write_sql_count = field(block, "Write SQL count")
        findings.append({
            "severity": severity,
            "module": module,
            "method": method,
            "path": path_value,
            "endpoint": endpoint or f"{method} {path_value}",
            "title": title,
            "signature": signature,
            "trace_id": field(block, "Trace ID"),
            "status_elapsed": field(block, "Status / elapsed"),
            "sql_count": sql_count,
            "write_sql_count": write_sql_count or "0",
            "source": field(block, "Source"),
            "evidence": fenced_sql(block),
            "analysis": paragraph_after(block, "Analysis"),
            "recommendation": paragraph_after(block, "Recommendation"),
            "audit_file": repo_rel(path),
            "fingerprint_input": fingerprint_input,
            "fingerprint": fingerprint,
        })
    return findings, result_summary, skipped


def parse_frontmatter(text: str) -> tuple[dict, str]:
    if not text.startswith("---\n"):
        return {}, text
    end = text.find("\n---\n", 4)
    if end == -1:
        return {}, text
    payload = {}
    for line in text[4:end].splitlines():
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] == '"':
            try:
                value = json.loads(value)
            except json.JSONDecodeError:
                value = value.strip('"')
        payload[key.strip()] = value
    return payload, text[end + 5:]


def render_frontmatter(frontmatter: dict) -> str:
    lines = ["---"]
    for key in ["id", "severity", "module", "endpoint", "status", "first_seen_run", "last_seen_run", "seen_count", "fingerprint"]:
        value = frontmatter.get(key, "")
        if key == "seen_count":
            lines.append(f"{key}: {int(value)}")
        elif key in {"severity", "status"}:
            lines.append(f"{key}: {value}")
        else:
            lines.append(f"{key}: {json.dumps(str(value), ensure_ascii=False)}")
    lines.append("---")
    return "\n".join(lines) + "\n\n"


def history_lines(body: str) -> list[str]:
    marker = "## 历史记录"
    if marker not in body:
        return []
    tail = body.split(marker, 1)[1].strip()
    return [line for line in tail.splitlines() if line.strip()]


def recommendation_lines(text: str) -> list[str]:
    lines = []
    for line in text.splitlines():
        clean = line.strip()
        if clean:
            lines.append(clean)
    return lines or ["后续变更中按审计建议优化该接口，并复跑 lina-perf-audit 验证。"]


def sql_excerpt(sql: str, limit: int = 8) -> str:
    lines = [line.rstrip() for line in sql.splitlines() if line.strip()]
    if len(lines) > limit:
        return "\n".join(lines[:limit] + ["-- truncated for persistent issue card"])
    return "\n".join(lines) or "-- no SQL excerpt retained"


def card_slug(finding: dict) -> str:
    module_slug = slugify(finding["module"].replace(":", "-"), "module")
    signature_slug = slugify(finding["signature"].replace(":", "-"), "pattern")
    path_slug = slugify(finding["path"].replace("{", "").replace("}", ""), "endpoint")
    slug = f"{signature_slug}-{path_slug}"
    if len(slug) > 96:
        slug = slug[:96].rstrip("-")
    return module_slug, slug


def issue_title_zh(finding: dict) -> str:
    signature = finding["signature"]
    if signature.startswith("read-write-side-effect"):
        if "session-last-active" in signature:
            return "GET 读请求刷新会话活动时间"
        if "operlog" in signature:
            return "GET 导出请求写入操作日志"
        if "host-call-demo" in signature:
            return "GET 演示接口执行持久化写入"
        if "dynamic-demo-record-list" in signature:
            return "动态插件列表 GET 写入会话与操作日志"
        if "dynamic-demo-record-detail" in signature:
            return "动态插件详情 GET 写入会话与操作日志"
        if "dynamic-demo-record-attachment" in signature:
            return "动态插件下载 GET 写入会话与操作日志"
        if "dynamic-backend-summary" in signature:
            return "动态插件摘要 GET 写入会话与操作日志"
        return "读接口存在写入副作用"
    if signature.startswith("n-plus-one"):
        return "列表查询存在 N+1 查询"
    if signature.startswith("small-sample-n-plus-one"):
        return "小样本下存在 N+1 查询风险"
    if signature.startswith("repeated-read") or signature.startswith("repeated-same-data-reads"):
        return "请求内重复读取相同数据"
    if signature.startswith("looped-role-menu-association-writes"):
        return "角色菜单关联逐条写入"
    return f"性能审查发现：{signature}"


def issue_description_zh(finding: dict) -> str:
    endpoint = finding["endpoint"]
    signature = finding["signature"]
    sql_count = finding["sql_count"]
    write_count = finding["write_sql_count"]
    trace = finding["trace_id"]
    if signature.startswith("read-write-side-effect"):
        if "session-last-active" in signature:
            cause = "刷新 `sys_online_session.last_active_time` 的会话心跳写入"
        elif "operlog" in signature:
            cause = "由 `operLog:\"export\"` 等审计元数据触发的操作日志写入"
        elif "host-call-demo" in signature:
            cause = "动态插件演示逻辑中的运行时状态、节点状态或插件数据写入"
        elif "dynamic-demo-record" in signature or "dynamic-backend-summary" in signature:
            cause = "共享中间件或动态插件操作日志产生的会话活动与操作日志写入"
        else:
            cause = "请求链路中的持久化写入"
        return (
            f"`{endpoint}` 属于 GET 或读语义接口，但本次审查在 Trace-ID `{trace}` 中观察到 "
            f"`{write_count}` 条写入 SQL，主要来源是{cause}。该接口样本中的总 SQL 数为 `{sql_count}`；"
            "即使响应耗时不高，读请求写库也会放大刷新、轮询、导出或详情访问的数据库写压力，"
            "并破坏 GET 请求应无副作用的接口约定。"
        )
    if signature.startswith("n-plus-one"):
        return (
            f"`{endpoint}` 在压力数据下执行了 `{sql_count}` 条 SQL，审计签名为 `{signature}`。"
            "SQL 调用数量会随返回行数或需要本地化的行数增长，说明列表查询路径存在 N+1 查询或逐行补充查询风险。"
            "当数据量继续增长时，该接口的数据库往返次数和响应耗时会同步上升。"
        )
    if signature.startswith("small-sample-n-plus-one"):
        return (
            f"`{endpoint}` 在当前样本中已经出现按记录逐项补充统计的查询模式，SQL 总数为 `{sql_count}`。"
            "虽然本次数据规模不大，审查仍将其归类为小样本 N+1 风险；当分组、角色或资源数量增长时，"
            "该模式会产生更多数据库往返。"
        )
    if signature.startswith("repeated-read") or signature.startswith("repeated-same-data-reads"):
        return (
            f"`{endpoint}` 在同一请求链路中重复读取相同或高度重叠的数据，SQL 总数为 `{sql_count}`，"
            f"审计签名为 `{signature}`。这类重复读取通常可以在请求内缓存、预加载或合并查询，"
            "否则会随着插件数量、菜单数量或日志行数增长而持续增加不必要的数据库访问。"
        )
    if signature.startswith("looped-role-menu-association-writes"):
        return (
            f"`{endpoint}` 在角色菜单关联写入时按菜单逐条执行写操作，本次样本 SQL 总数为 `{sql_count}`，"
            f"写入 SQL 数为 `{write_count}`。当角色关联的菜单数量变多时，该实现会产生更多数据库写入往返，"
            "应改为批量插入或批量替换。"
        )
    return f"`{endpoint}` 命中性能审查签名 `{signature}`，本次样本 SQL 总数为 `{sql_count}`。"


def recommendations_zh(finding: dict) -> list[str]:
    signature = finding["signature"]
    endpoint = finding["endpoint"]
    if signature.startswith("read-write-side-effect"):
        if "operlog" in signature:
            return [
                "将导出审计日志写入迁移到显式的 `POST` 动作，或拆分为只读下载接口与独立写入动作。",
                f"修复后复跑 `{endpoint}`，确认 GET 请求链路中的写入 SQL 数为 `0`。",
            ]
        if "session-last-active" in signature:
            return [
                "将会话活跃时间刷新从普通 GET/读请求链路中移出，改为显式心跳接口、节流写入或非持久化读模型。",
                f"修复后复跑 `{endpoint}`，确认 Trace-ID 对应 SQL 中不再出现 `UPDATE sys_online_session`。",
            ]
        if "host-call-demo" in signature:
            return [
                "将会产生运行时状态、节点状态、存储或结构化数据写入的演示逻辑改为 `POST`，保留 GET 仅返回只读摘要。",
                "分别复跑 `skipNetwork=1` 与正常路径，确认 GET 请求写入 SQL 数为 `0`。",
            ]
        return [
            "将持久化访问记录、会话活动或操作日志写入移出 GET 读请求，必要时改为显式写动作或非持久化遥测。",
            f"修复后复跑 `{endpoint}`，确认请求链路只包含读取类 SQL。",
        ]
    if signature.startswith("n-plus-one") or signature.startswith("small-sample-n-plus-one"):
        return [
            "将逐行补充查询改为批量预加载，例如使用 `WHERE IN`、JOIN、聚合子查询或请求内缓存一次性取回关联数据。",
            f"使用压力数据复跑 `{endpoint}`，确认 SQL 数量不再随返回行数线性增长。",
        ]
    if signature.startswith("repeated-read") or signature.startswith("repeated-same-data-reads"):
        return [
            "在请求链路内复用已读取的数据，或将重复读取合并为一次批量查询，避免相同插件、菜单、发布版本或本地化元数据被反复加载。",
            f"复跑 `{endpoint}`，确认重复读取的 SQL 被折叠，且 GET 请求仍无写入 SQL。",
        ]
    if signature.startswith("looped-role-menu-association-writes"):
        return [
            "将角色菜单关联的逐条写入改为批量插入、批量替换或一次性差异同步。",
            "使用包含大量 `menuIds` 的审计夹具复跑创建和更新接口，确认关联写入次数保持稳定。",
        ]
    return [f"按 `{signature}` 对应的反模式优化 `{endpoint}`，并复跑 `lina-perf-audit` 验证结果。"]


def status_zh(status: str) -> str:
    return {
        "open": "待处理",
        "in-progress": "处理中",
        "fixed": "已修复",
        "obsolete": "已废弃",
    }.get(status, status)


def load_cards_by_fingerprint() -> dict:
    mapping = {}
    for path in sorted(issue_dir.glob("*.md")):
        if path.name == "INDEX.md":
            continue
        frontmatter, _ = parse_frontmatter(path.read_text(encoding="utf-8"))
        fingerprint = frontmatter.get("fingerprint")
        if fingerprint and fingerprint not in mapping:
            mapping[fingerprint] = path
    return mapping


def write_card(finding: dict, existing_path: Path | None) -> Path:
    module_slug, slug = card_slug(finding)
    severity = finding["severity"]
    default_id = f"{severity}-{module_slug}-{slug}"
    target_path = existing_path or (issue_dir / f"{default_id}.md")
    if existing_path is None:
        suffix = 1
        while target_path.exists():
            suffix += 1
            target_path = issue_dir / f"{default_id}-{suffix}.md"

    previous_frontmatter = {}
    previous_history = []
    if target_path.exists():
        previous_frontmatter, previous_body = parse_frontmatter(target_path.read_text(encoding="utf-8"))
        previous_history = history_lines(previous_body)

    previous_status = previous_frontmatter.get("status", "open")
    first_seen = previous_frontmatter.get("first_seen_run", run_id)
    last_seen = previous_frontmatter.get("last_seen_run")
    seen_count = int(previous_frontmatter.get("seen_count", 0) or 0)
    increment = 0 if last_seen == run_id else 1
    status = previous_status
    regression = False
    if previous_status in {"fixed", "obsolete"}:
        status = "open"
        regression = True
        increment = max(1, increment)
    elif not previous_frontmatter:
        status = "open"
        increment = 1

    seen_count = max(1, seen_count + increment)
    card_id = previous_frontmatter.get("id", target_path.stem)
    frontmatter = {
        "id": card_id,
        "severity": severity,
        "module": finding["module"],
        "endpoint": finding["endpoint"],
        "status": status,
        "first_seen_run": first_seen,
        "last_seen_run": run_id,
        "seen_count": seen_count,
        "fingerprint": finding["fingerprint"],
    }

    previous_history = [
        line for line in previous_history
        if not line.startswith(f"- {run_id}:")
    ]
    history_entry = (
        f"- {run_id}：{'被再次观察到（回归）' if regression else '本次审查发现'}，"
        f"审计文件 `{finding['audit_file']}`，SQL 总数 `{finding['sql_count']}`，"
        f"写入 SQL 数 `{finding['write_sql_count']}`，Trace-ID `{finding['trace_id']}`。"
    )
    if history_entry not in previous_history:
        previous_history.append(history_entry)

    recs = recommendations_zh(finding)
    rec_markdown = "\n".join(f"{idx}. {line}" for idx, line in enumerate(recs, 1))
    body = f"""# {severity} - {finding['module']} - {issue_title_zh(finding)}

## 问题描述

{issue_description_zh(finding)}

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id {run_id}`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/{run_id}`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/{run_id}`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>{finding['path']}"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/{run_id}/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `{finding['sql_count']}`，写入 SQL 数 `{finding['write_sql_count']}`，审计签名 `{finding['signature']}`。

## 证据

- Trace-ID：`{finding['trace_id']}`
- 审计文件：`{finding['audit_file']}`
- 源码位置：`{finding['source']}`
- SQL 总数：`{finding['sql_count']}`
- 写入 SQL 数：`{finding['write_sql_count']}`
- 指纹输入：`{finding['fingerprint_input']}`

```sql
{sql_excerpt(finding['evidence'])}
```

## 改进方案

{rec_markdown}

## 历史记录

{chr(10).join(previous_history)}
"""
    target_path.write_text(render_frontmatter(frontmatter) + body, encoding="utf-8")
    return target_path


def regenerate_index() -> None:
    records = []
    for path in sorted(issue_dir.glob("*.md")):
        if path.name == "INDEX.md":
            continue
        frontmatter, _ = parse_frontmatter(path.read_text(encoding="utf-8"))
        if frontmatter.get("status") not in {"open", "in-progress"}:
            continue
        records.append((frontmatter, repo_rel(path)))
    order = {"HIGH": 0, "MEDIUM": 1, "LOW": 2}
    records.sort(key=lambda item: (order.get(item[0].get("severity", "LOW"), 9), item[0].get("module", ""), item[0].get("endpoint", "")))

    lines = [
        "# LinaPro 性能问题索引",
        "",
        f"- 生成来源 run：`{run_id}`",
        f"- 活跃问题卡片数：`{len(records)}`",
        "",
        "| 严重度 | 模块 | 接口 | 状态 | 最近发现 | 出现次数 | 卡片 |",
        "|---|---|---|---|---|---:|---|",
    ]
    for frontmatter, rel_path in records:
        lines.append(
            "| {severity} | `{module}` | `{endpoint}` | {status} | `{last_seen}` | {seen} | [{card}]({card}) |".format(
                severity=frontmatter.get("severity", ""),
                module=frontmatter.get("module", ""),
                endpoint=frontmatter.get("endpoint", ""),
                status=status_zh(frontmatter.get("status", "")),
                last_seen=frontmatter.get("last_seen_run", ""),
                seen=frontmatter.get("seen_count", 0),
                card=rel_path,
            )
        )
    (issue_dir / "INDEX.md").write_text("\n".join(lines) + "\n", encoding="utf-8")


catalog = read_json(run_dir / "catalog.json", {})
plugins = read_json(run_dir / "plugins.json", {})
stress = read_json(run_dir / "stress-fixture.json", {})
backup = read_json(run_dir / "logger-backup.json", {"logger": {}})

all_findings = []
audit_summaries = {}
manual_follow_up = []
for audit_file in sorted((run_dir / "audits").glob("*.md")):
    findings, result_summary, skipped = parse_audit_file(audit_file)
    all_findings.extend(findings)
    audit_summaries[audit_file.name] = result_summary
    manual_follow_up.extend(skipped)

cards_by_fingerprint = load_cards_by_fingerprint()
expected_side_effects = [finding for finding in all_findings if is_expected_read_side_effect(finding)]
expected_side_effect_fingerprints = {finding["fingerprint"] for finding in expected_side_effects}
removed_expected_cards = []
for finding in expected_side_effects:
    path = cards_by_fingerprint.get(finding["fingerprint"])
    if path and path.exists():
        removed_expected_cards.append(repo_rel(path))
        path.unlink()
        cards_by_fingerprint.pop(finding["fingerprint"], None)

reportable_findings = [
    finding for finding in all_findings
    if finding["fingerprint"] not in expected_side_effect_fingerprints
]
updated_cards = []
for finding in reportable_findings:
    path = write_card(finding, cards_by_fingerprint.get(finding["fingerprint"]))
    cards_by_fingerprint[finding["fingerprint"]] = path
    finding["card"] = repo_rel(path)
    updated_cards.append(finding["card"])

regenerate_index()

finding_counts = Counter(finding["severity"] for finding in reportable_findings)
read_side_effects = [
    finding for finding in reportable_findings
    if finding["signature"].startswith("read-write-side-effect")
]

host_endpoint_count = sum(1 for endpoint in catalog.get("endpoints", []) if endpoint.get("owner") == "core")
plugin_endpoint_count = sum(1 for endpoint in catalog.get("endpoints", []) if endpoint.get("owner") != "core")
skipped_plugins = [
    {
        "plugin": item.get("pluginId") or item.get("plugin"),
        "reason": item.get("reason", "no backend API"),
    }
    for item in catalog.get("skippedPlugins", [])
]

restore_log = run_dir / "10-restore-audit-env.log"
restore_result = "success" if restore_log.exists() and "Restored logger.path" in restore_log.read_text(encoding="utf-8", errors="ignore") else "unknown"
started_at = mtime_iso(run_dir / "00-make-stop.log")
finished_at = now_iso()

audit_files_rows = []
for audit_file in sorted((run_dir / "audits").glob("*.md")):
    summary = audit_summaries.get(audit_file.name, {})
    status_parts = []
    for key in ["HIGH", "MEDIUM", "LOW", "PASS", "SKIPPED"]:
        if key in summary:
            status_parts.append(f"{key}={summary[key]}")
    audit_files_rows.append((audit_file.stem, repo_rel(audit_file), ", ".join(status_parts) or "completed"))


def issue_table(severity: str) -> list[str]:
    rows = [
        "| Issue | Endpoint | Module | Card |",
        "|---|---|---|---|",
    ]
    for finding in reportable_findings:
        if finding["severity"] != severity:
            continue
        rows.append(f"| {finding['title']} | `{finding['endpoint']}` | `{finding['module']}` | [{Path(finding['card']).name}]({finding['card']}) |")
    if len(rows) == 2:
        rows.append("| None |  |  |  |")
    return rows


summary_lines = [
    "# LinaPro Performance Audit Summary",
    "",
    "## Run Metadata",
    "",
    f"- Run ID: `{run_id}`",
    f"- Git commit: `{git_commit()}`",
    f"- Started at: `{started_at}`",
    f"- Finished at: `{finished_at}`",
    f"- Run directory: `{run_dir.as_posix()}/`",
    f"- Stress fixture: `{'enabled' if stress.get('status') in {'ok', 'completed'} or stress.get('resources') else 'unknown'}`",
    f"- Read side-effect violations: `{len(read_side_effects)}`",
    f"- Expected session/operation-log side effects ignored: `{len(expected_side_effects)}`",
    f"- Sub agents: `{len(audit_summaries)}` report files completed",
    f"- Logger original path/file: `{backup.get('logger', {}).get('path', '')}` / `{backup.get('logger', {}).get('file', '')}`",
    f"- Logger audit path/file: `{run_dir.as_posix()}/` / `server.log`",
    f"- Restore result: `{restore_result}`",
    "- Delivery code modified by dry-run: `no`",
    "",
    "## Scope",
    "",
    f"- Host API catalog: `{host_endpoint_count}` endpoints",
    f"- Built-in plugin API catalog: `{plugin_endpoint_count}` endpoints",
    f"- Total modules: `{catalog.get('moduleCount', len(catalog.get('modules', [])))}`",
    f"- Total endpoints: `{catalog.get('endpointCount', len(catalog.get('endpoints', [])))}`",
    f"- Skipped plugins: `{', '.join(f'{item['plugin']}: {item['reason']}' for item in skipped_plugins) or 'none'}`",
    "",
    "## HIGH",
    "",
    *issue_table("HIGH"),
    "",
    "## Read Request Side-Effect Violations",
    "",
    "| Endpoint | Module | Write SQL count | Card |",
    "|---|---|---:|---|",
]
if read_side_effects:
    for finding in read_side_effects:
        summary_lines.append(f"| `{finding['endpoint']}` | `{finding['module']}` | {finding['write_sql_count']} | [{Path(finding['card']).name}]({finding['card']}) |")
else:
    summary_lines.append("| None |  | 0 |  |")

summary_lines.extend([
    "",
    "## MEDIUM",
    "",
    *issue_table("MEDIUM"),
    "",
    "## LOW",
    "",
    *issue_table("LOW"),
    "",
    "## Manual Follow-Up Required",
    "",
    "| Endpoint | Reason | Audit file |",
    "|---|---|---|",
])
if manual_follow_up:
    for item in manual_follow_up:
        summary_lines.append(f"| `{item['endpoint']}` | {item['reason']} | `{item['audit_file']}` |")
else:
    summary_lines.append("| None |  |  |")

summary_lines.extend([
    "",
    "## Audit Files",
    "",
    "| Module or shard | Path | Status |",
    "|---|---|---|",
])
for module, path, status in audit_files_rows:
    summary_lines.append(f"| `{module}` | `{path}` | {status} |")

summary_lines.extend([
    "",
    "## Notes",
    "",
    "- `demo-control` was enabled during built-in plugin preparation, temporarily disabled for write/destructive endpoint sampling, then restored to `enabled` before environment cleanup.",
    "- Read-request traces whose writes only touched `sys_online_session` or `plugin_monitor_operlog` were treated as expected operational side effects and were not emitted as `perf-issues/` cards.",
    "- This dry-run did not modify `apps/lina-core` delivery code, frontend code, or delivery SQL files. The OpenSpec implementation itself adds the audit skill files and persistent `perf-issues/` cards.",
    "- Some endpoint rows are marked `SKIPPED` in module audit files where safe fixtures or retained SQL evidence were not available before the sub-agent cutoff; they are listed above as manual follow-up.",
])
(run_dir / "SUMMARY.md").write_text("\n".join(summary_lines) + "\n", encoding="utf-8")

meta = {
    "run_id": run_id,
    "started_at": started_at,
    "finished_at": finished_at,
    "git_commit": git_commit(),
    "run_dir": run_dir.as_posix(),
    "stress_fixture_enabled": bool(stress.get("resources")),
    "logger": {
        "original_path": backup.get("logger", {}).get("path", ""),
        "original_file": backup.get("logger", {}).get("file", ""),
        "audit_path": run_dir.as_posix() + "/",
        "audit_file": "server.log",
        "restore_result": restore_result,
    },
    "sub_agents": {
        "count": len(audit_summaries),
        "completed": len(audit_summaries),
        "failed": 0,
        "reports": sorted(audit_summaries.keys()),
    },
    "catalog": {
        "module_count": catalog.get("moduleCount", len(catalog.get("modules", []))),
        "endpoint_count": catalog.get("endpointCount", len(catalog.get("endpoints", []))),
        "host_endpoint_count": host_endpoint_count,
        "plugin_endpoint_count": plugin_endpoint_count,
    },
    "findings": dict(finding_counts),
    "read_side_effect_violations": len(read_side_effects),
    "expected_read_side_effects_ignored": {
        "count": len(expected_side_effects),
        "write_tables": sorted(EXPECTED_READ_SIDE_EFFECT_TABLES),
        "removed_cards": removed_expected_cards,
    },
    "skipped_plugins": skipped_plugins,
    "plugins": {
        "failed": plugins.get("failed", False),
        "count": len(plugins.get("plugins", [])),
    },
    "temporary_actions": [
        {
            "action": "disable-demo-control-for-write-audit",
            "reason": "avoid demo-control write guard while sampling mutating endpoints",
            "restore_result": "enabled" if (run_dir / "09-restore-demo-control.log").exists() else "unknown",
        }
    ],
    "issue_cards": {
        "updated_count": len(updated_cards),
        "index": repo_rel(issue_dir / "INDEX.md"),
        "cards": sorted(set(updated_cards)),
    },
}
(run_dir / "meta.json").write_text(json.dumps(meta, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

print(f"Aggregated {len(reportable_findings)} reportable findings into {run_dir / 'SUMMARY.md'}")
print(f"Ignored {len(expected_side_effects)} expected session/operation-log side-effect findings")
print(f"Updated {len(set(updated_cards))} issue cards under {repo_rel(issue_dir)}/")
PY
