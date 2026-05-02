#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/probe-fixtures.sh --run-dir <dir> [--base-url <url>]

Reads <run-dir>/catalog.json and a token from <run-dir>, probes safe GET list/detail
routes, and writes <run-dir>/fixtures.json. Run from the repository root.
USAGE
}

RUN_DIR=""
BASE_URL="${LINA_PERF_AUDIT_BASE_URL:-http://127.0.0.1:8080}"

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
    --base-url)
      if [[ $# -lt 2 ]]; then
        echo "missing value for --base-url" >&2
        exit 2
      fi
      BASE_URL="$2"
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

if [[ ! -f "apps/lina-core/manifest/config/config.yaml" || ! -d "apps/lina-core/api" ]]; then
  echo "probe-fixtures.sh must be run from the LinaPro repository root" >&2
  exit 1
fi

if [[ ! -f "$RUN_DIR/catalog.json" ]]; then
  echo "missing catalog: $RUN_DIR/catalog.json" >&2
  exit 1
fi

python3 - "$RUN_DIR" "$BASE_URL" <<'PY'
import json
import os
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

run_dir = Path(sys.argv[1])
base_url = sys.argv[2].rstrip("/")
catalog = json.loads((run_dir / "catalog.json").read_text(encoding="utf-8"))
fixtures_path = run_dir / "fixtures.json"


def now_utc() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")


def find_token_in_object(value):
    if isinstance(value, dict):
        for key in ("accessToken", "access_token", "token", "jwt"):
            token = value.get(key)
            if isinstance(token, str) and token.strip():
                return token.strip()
        for child in value.values():
            token = find_token_in_object(child)
            if token:
                return token
    if isinstance(value, list):
        for child in value:
            token = find_token_in_object(child)
            if token:
                return token
    return ""


def read_token() -> tuple[str, str]:
    env_token = os.environ.get("LINA_PERF_AUDIT_TOKEN", "").strip()
    if env_token:
        return env_token, "env:LINA_PERF_AUDIT_TOKEN"

    text_candidates = [
        "token",
        "token.txt",
        "access-token",
        "access-token.txt",
        "access_token",
        "access_token.txt",
        "admin.token",
    ]
    for name in text_candidates:
        path = run_dir / name
        if path.exists():
            token = path.read_text(encoding="utf-8").strip()
            if token:
                return token, str(path)

    json_candidates = [
        "token.json",
        "auth-token.json",
        "login.json",
        "setup-audit-env.json",
        "meta.json",
    ]
    for name in json_candidates:
        path = run_dir / name
        if not path.exists():
            continue
        try:
            payload = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError:
            continue
        token = find_token_in_object(payload)
        if token:
            return token, str(path)

    return "", ""


def request_json(method: str, path: str, token: str, query: dict | None = None) -> dict:
    query = query or {}
    encoded_query = urllib.parse.urlencode({key: value for key, value in query.items() if value not in ("", None)}, doseq=True)
    url = base_url + path
    if encoded_query:
        url += "?" + encoded_query
    headers = {
        "Accept": "application/json",
        "User-Agent": "lina-perf-audit-probe/1",
    }
    if token:
        headers["Authorization"] = "Bearer " + token
    request = urllib.request.Request(url, headers=headers, method=method)
    started = time.monotonic()
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            raw = response.read()
            status = response.status
            response_headers = dict(response.headers.items())
    except urllib.error.HTTPError as exc:
        raw = exc.read()
        status = exc.code
        response_headers = dict(exc.headers.items())
    except urllib.error.URLError as exc:
        elapsed_ms = int((time.monotonic() - started) * 1000)
        return {
            "url": url,
            "status": 0,
            "headers": {},
            "elapsedMs": elapsed_ms,
            "body": None,
            "bodyText": str(exc.reason if hasattr(exc, "reason") else exc)[:1000],
        }
    elapsed_ms = int((time.monotonic() - started) * 1000)
    text = raw.decode("utf-8", errors="replace")
    parsed = None
    if text.strip():
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError:
            parsed = None
    return {
        "url": url,
        "status": status,
        "headers": response_headers,
        "elapsedMs": elapsed_ms,
        "body": parsed,
        "bodyText": text[:1000],
    }


def login_for_token() -> tuple[str, str]:
    payload = json.dumps({"username": "admin", "password": "admin123"}).encode("utf-8")
    request = urllib.request.Request(
        base_url + "/api/v1/auth/login",
        data=payload,
        headers={"Content-Type": "application/json", "Accept": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=20) as response:
        data = json.loads(response.read().decode("utf-8"))
    token = find_token_in_object(data)
    if not token:
        raise RuntimeError("login fallback did not return an access token")
    return token, "fallback-login:/api/v1/auth/login"


def success_body(response: dict) -> tuple[bool, str]:
    if response["status"] < 200 or response["status"] >= 300:
        return False, f'HTTP {response["status"]}'
    body = response.get("body")
    if isinstance(body, dict) and "code" in body and body.get("code") not in (0, "0"):
        return False, f'application code {body.get("code")}: {body.get("message", "")}'
    return True, ""


def is_missing_generic_plugin_resource(endpoint: dict, response: dict, reason: str) -> bool:
    if "/plugins/{id}/resources/{resource}" not in endpoint.get("publicPath", ""):
        return False
    body = response.get("body")
    message = ""
    if isinstance(body, dict):
        message = str(body.get("message", ""))
    return "plugin resource does not exist" in " ".join([reason, message]).lower()


def response_data(response: dict):
    body = response.get("body")
    if isinstance(body, dict) and "data" in body:
        return body.get("data")
    return body


def find_first_list(value):
    if isinstance(value, dict):
        for key in ("list", "items", "records", "rows", "data"):
            child = value.get(key)
            if isinstance(child, list):
                return child
        for child in value.values():
            found = find_first_list(child)
            if found is not None:
                return found
    elif isinstance(value, list):
        return value
    return None


def pick_sample_id(item):
    if not isinstance(item, dict):
        return "", ""
    for key in ("id", "Id", "ID", "tokenId", "key", "type", "code", "ref"):
        value = item.get(key)
        if value not in (None, ""):
            return str(value), key
    for key, value in item.items():
        if key.lower().endswith("id") and value not in (None, ""):
            return str(value), key
    return "", ""


def compact_sample(item):
    if not isinstance(item, dict):
        return item
    compact = {}
    for key, value in item.items():
        if len(compact) >= 12:
            break
        if isinstance(value, (str, int, float, bool)) or value is None:
            compact[key] = value
    return compact


def query_defaults(endpoint: dict) -> dict:
    query = {}
    for field in endpoint.get("queryFields", []):
        name = field.get("json")
        if not name:
            continue
        lower = name.lower()
        if lower in {"pagenum", "page", "current"}:
            query[name] = field.get("default") or "1"
        elif lower in {"pagesize", "size", "limit"}:
            query[name] = "100"
    return query


def endpoint_sort_key(endpoint: dict):
    probe_order = {"list": 0, "detail": 1, "none": 2}
    return (
        probe_order.get(endpoint.get("probeClass", "none"), 9),
        endpoint.get("moduleKey", ""),
        endpoint.get("publicPath", ""),
    )


def infer_resource_from_path(path: str, default_resource: str) -> str:
    parts = [part for part in path.split("/") if part and not part.startswith("{")]
    if not parts:
        return default_resource
    if "api" in parts and "v1" in parts:
        parts = parts[parts.index("v1") + 1:]
    if parts and parts[0] == "extensions" and len(parts) >= 3:
        return parts[2].replace("-", "_")
    if parts and parts[0] == "plugins" and len(parts) >= 3:
        return parts[2].replace("-", "_")
    return (parts[0] if parts else default_resource).replace("-", "_")


def path_param_values(endpoint: dict, module_samples: dict) -> tuple[dict, list]:
    missing = []
    values = {}
    fields_by_json = {field.get("json", "").lower(): field for field in endpoint.get("fields", [])}
    sample = module_samples.get(endpoint.get("resource", "")) or module_samples.get(infer_resource_from_path(endpoint.get("publicPath", ""), endpoint.get("resource", "")))
    prefer_examples = "/plugins/{id}/resources/{resource}" in endpoint.get("publicPath", "")
    for name in endpoint.get("pathParams", []):
        lower = name.lower()
        field = fields_by_json.get(lower)
        value = ""
        if prefer_examples and field:
            value = field.get("eg") or field.get("default") or ""
        if not value and lower in {"id", "ids"} and sample:
            value = sample.get("sampleId", "")
        if not value and field:
            value = field.get("eg") or field.get("default") or ""
        if not value and lower.endswith("id") and sample:
            value = sample.get("sampleId", "")
        if not value:
            missing.append(name)
        else:
            values[name] = value
    return values, missing


def substitute_path(path: str, values: dict) -> str:
    result = path
    for name, value in values.items():
        encoded = urllib.parse.quote(str(value), safe="")
        result = result.replace("{" + name + "}", encoded)
    return result


token, token_source = read_token()
token_error = ""
if not token:
    try:
        token, token_source = login_for_token()
    except Exception as exc:
        token_error = str(exc)

fixtures = {
    "schemaVersion": 1,
    "generatedAt": now_utc(),
    "baseUrl": base_url,
    "catalog": "catalog.json",
    "tokenSource": token_source,
    "modules": {},
    "probes": [],
    "failures": [],
    "skipped": [],
}

if token_error:
    fixtures["failures"].append({
        "endpointId": "auth-token",
        "moduleKey": "setup",
        "method": "POST",
        "publicPath": "/api/v1/auth/login",
        "status": 0,
        "reason": "unable to read token from run-dir or login fallback: " + token_error,
        "body": "",
    })
    fixtures_path.write_text(json.dumps(fixtures, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(str(fixtures_path))
    sys.exit(1)

endpoints = sorted(catalog.get("endpoints", []), key=endpoint_sort_key)

for endpoint in endpoints:
    if endpoint.get("method") != "GET":
        continue
    probe_class = endpoint.get("probeClass", "none")
    if probe_class not in {"list", "detail"}:
        continue
    module_key = endpoint.get("moduleKey", "unknown")
    module = fixtures["modules"].setdefault(module_key, {
        "owner": endpoint.get("owner"),
        "pluginId": endpoint.get("pluginId"),
        "module": endpoint.get("module"),
        "resources": {},
    })

    if probe_class == "detail":
        path_values, missing = path_param_values(endpoint, module["resources"])
        if missing:
            fixtures["skipped"].append({
                "endpointId": endpoint.get("id"),
                "method": endpoint.get("method"),
                "publicPath": endpoint.get("publicPath"),
                "reason": "missing fixture for path params: " + ",".join(missing),
            })
            continue
        public_path = substitute_path(endpoint.get("publicPath", ""), path_values)
        query = {}
    else:
        public_path = endpoint.get("publicPath", "")
        query = query_defaults(endpoint)

    response = request_json("GET", public_path, token, query)
    ok, reason = success_body(response)
    probe_record = {
        "endpointId": endpoint.get("id"),
        "moduleKey": module_key,
        "resource": endpoint.get("resource"),
        "probeClass": probe_class,
        "method": "GET",
        "publicPath": public_path,
        "query": query,
        "status": response["status"],
        "elapsedMs": response["elapsedMs"],
        "traceId": response["headers"].get("Trace-ID", response["headers"].get("Trace-Id", "")),
    }
    fixtures["probes"].append(probe_record)
    if not ok:
        if is_missing_generic_plugin_resource(endpoint, response, reason):
            fixtures["skipped"].append({
                "endpointId": endpoint.get("id"),
                "method": "GET",
                "publicPath": public_path,
                "reason": "no installed plugin declares a generic backend resource for this route",
            })
            continue
        failure = {
            "endpointId": endpoint.get("id"),
            "moduleKey": module_key,
            "method": "GET",
            "publicPath": public_path,
            "status": response["status"],
            "reason": reason,
            "body": response.get("body") if isinstance(response.get("body"), dict) else response.get("bodyText", ""),
        }
        fixtures["failures"].append(failure)
        continue

    if probe_class == "list":
        items = find_first_list(response_data(response))
        if not items:
            fixtures["skipped"].append({
                "endpointId": endpoint.get("id"),
                "method": "GET",
                "publicPath": public_path,
                "reason": "list probe returned no sample rows",
            })
            continue
        sample_id, sample_field = pick_sample_id(items[0])
        if not sample_id:
            fixtures["skipped"].append({
                "endpointId": endpoint.get("id"),
                "method": "GET",
                "publicPath": public_path,
                "reason": "list sample row has no usable id field",
            })
            continue
        resource = endpoint.get("resource") or infer_resource_from_path(public_path, endpoint.get("module", "resource"))
        module["resources"][resource] = {
            "sampleId": sample_id,
            "sampleField": sample_field,
            "sourceEndpointId": endpoint.get("id"),
            "sourcePublicPath": public_path,
            "sample": compact_sample(items[0]),
        }

fixtures_path.write_text(json.dumps(fixtures, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(str(fixtures_path))

if fixtures["failures"]:
    print(f'{len(fixtures["failures"])} probed DTO route(s) were not accessible', file=sys.stderr)
    sys.exit(1)
PY
