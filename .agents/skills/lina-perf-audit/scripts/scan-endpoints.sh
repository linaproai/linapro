#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh --run-dir <dir>

Scans LinaPro host and built-in plugin API DTOs and writes <run-dir>/catalog.json.
Run this script from the repository root.
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

if [[ ! -f "apps/lina-core/manifest/config/config.yaml" || ! -d "apps/lina-core/api" ]]; then
  echo "scan-endpoints.sh must be run from the LinaPro repository root" >&2
  exit 1
fi

mkdir -p "$RUN_DIR"

python3 - "$RUN_DIR" <<'PY'
import ast
import json
import re
import sys
from datetime import datetime, timezone
from pathlib import Path

run_dir = Path(sys.argv[1])
root = Path.cwd()

api_prefix = "/api/v1"
meta_re = re.compile(r'(?:^|\s)(?:g|gmeta)\.Meta\s+`([^`]*)`')
struct_re = re.compile(r'type\s+([A-Za-z_]\w*)\s+struct\s*\{(.*?)\n\}', re.S)
field_re = re.compile(r'^\s*([A-Za-z_]\w*)\s+([^`]+?)\s+`([^`]*)`')
tag_re = re.compile(r'([A-Za-z_]\w*):"((?:\\.|[^"\\])*)"')
placeholder_re = re.compile(r'\{([^}]+)\}|\*([A-Za-z_]\w*)')


def decode_go_tag_value(value: str) -> str:
    try:
        return ast.literal_eval('"' + value + '"')
    except Exception:
        return value.replace(r'\"', '"').replace(r'\\', '\\')


def parse_tags(tag_text: str) -> dict:
    parsed = {}
    for key, value in tag_re.findall(tag_text):
        parsed[key] = decode_go_tag_value(value)
    return parsed


def clean_yaml_scalar(value: str) -> str:
    value = value.strip()
    if not value:
        return ""
    if value[0:1] in {"'", '"'} and value[-1:] == value[0:1]:
        return value[1:-1]
    return value


def parse_plugin_yaml(path: Path) -> dict:
    data = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        if raw_line[:1].isspace():
            continue
        match = re.match(r'^([A-Za-z0-9_-]+):\s*(.*?)\s*(?:#.*)?$', raw_line)
        if match:
            data[match.group(1)] = clean_yaml_scalar(match.group(2))
    plugin_id = data.get("id") or path.parent.name
    data["id"] = plugin_id
    data.setdefault("type", "source")
    data["path"] = str(path)
    return data


def normalize_path(path: str) -> str:
    if not path:
        return ""
    return path if path.startswith("/") else "/" + path


def join_paths(prefix: str, path: str) -> str:
    path = normalize_path(path)
    return prefix.rstrip("/") + path


def parse_fields(body: str) -> list:
    fields = []
    for line in body.splitlines():
        if "g.Meta" in line or "gmeta.Meta" in line:
            continue
        match = field_re.match(line)
        if not match:
            continue
        field_name, go_type, tag_text = match.groups()
        tags = parse_tags(tag_text)
        json_name = tags.get("json", "")
        if not json_name or json_name == "-":
            continue
        json_name = json_name.split(",", 1)[0]
        if not json_name:
            continue
        fields.append({
            "name": field_name,
            "json": json_name,
            "goType": " ".join(go_type.strip().split()),
            "dc": tags.get("dc", ""),
            "eg": tags.get("eg", ""),
            "default": tags.get("d", ""),
            "validation": tags.get("v", ""),
        })
    return fields


def path_placeholders(path: str) -> list:
    names = []
    for match in placeholder_re.finditer(path):
        names.append(match.group(1) or match.group(2))
    return names


def infer_resource(route_path: str, module: str) -> str:
    parts = [part for part in normalize_path(route_path).split("/") if part and not part.startswith("{") and not part.startswith("*")]
    if not parts:
        return module
    if parts[0] == "plugins" and len(parts) >= 3:
        return parts[2].replace("-", "_")
    if parts[0] == "extensions" and len(parts) >= 3:
        return parts[2].replace("-", "_")
    if len(parts) >= 2 and parts[0] in {"dict", "job", "user"}:
        return "_".join(parts[:2]).replace("-", "_")
    return parts[0].replace("-", "_")


def classify_probe(method: str, route_path: str, summary: str, dc: str) -> str:
    if method.lower() != "get":
        return "none"
    lowered = " ".join([route_path, summary, dc]).lower()
    if any(token in lowered for token in ["export", "download", "template", "upload", "avatar"]):
        return "none"
    if "*" in route_path:
        return "none"
    if "{" in route_path:
        return "detail"
    if any(token in lowered for token in [
        " list", "list ", "query", "options", "option", "tree", "count",
        "monitor", "health", "information", "runtime", "diagnostic",
        "get all", "get host", "get current",
    ]):
        return "list"
    return "none"


def expects_read_only(method: str, route_path: str, summary: str, dc: str) -> bool:
    if method.upper() == "GET":
        return True
    lowered = " ".join([route_path, summary, dc]).lower()
    if any(token in lowered for token in [
        "create", "add ", "update", "delete", "remove", "clear", "reset",
        "mark ", "install", "enable", "disable", "upload", "import", "sync", "login", "logout",
    ]):
        return False
    return any(token in lowered for token in [
        " list", "list ", "query", "search", "options", "option", "tree",
        "count", "detail", "information", "get all", "get current", "current user",
    ])


def endpoint_from_struct(file_path: Path, struct_name: str, body: str, owner: dict) -> dict | None:
    meta_match = meta_re.search(body)
    if not meta_match:
        return None
    meta = parse_tags(meta_match.group(1))
    if "path" not in meta or "method" not in meta:
        return None

    route_path = normalize_path(meta.get("path", ""))
    method = meta.get("method", "").upper()
    fields = parse_fields(body)
    placeholders = path_placeholders(route_path)
    path_field_set = {name.lower() for name in placeholders}
    query_fields = [field for field in fields if field["json"].lower() not in path_field_set]

    plugin_id = owner.get("pluginId")
    plugin_type = owner.get("pluginType")
    if owner["owner"] == "dynamic-plugin":
        public_path = join_paths(f"{api_prefix}/extensions/{plugin_id}", route_path)
    else:
        public_path = join_paths(api_prefix, route_path)

    resource = infer_resource(route_path, owner["module"])
    probe_class = classify_probe(method, route_path, meta.get("summary", ""), meta.get("dc", ""))
    read_only_expected = expects_read_only(method, route_path, meta.get("summary", ""), meta.get("dc", ""))
    relative = file_path.relative_to(root).as_posix()
    return {
        "id": f'{owner["moduleKey"]}:{method}:{route_path}:{struct_name}',
        "owner": owner["owner"],
        "pluginId": plugin_id,
        "pluginType": plugin_type,
        "module": owner["module"],
        "moduleKey": owner["moduleKey"],
        "resource": resource,
        "requestStruct": struct_name,
        "sourceFile": relative,
        "method": method,
        "path": route_path,
        "publicPath": public_path,
        "tags": meta.get("tags", ""),
        "summary": meta.get("summary", ""),
        "dc": meta.get("dc", ""),
        "permission": meta.get("permission", ""),
        "access": meta.get("access", ""),
        "eg": meta.get("eg", meta.get("example", "")),
        "operLog": meta.get("operLog", ""),
        "mime": meta.get("mime", ""),
        "meta": meta,
        "pathParams": placeholders,
        "queryFields": query_fields,
        "fields": fields,
        "probeClass": probe_class,
        "readOnlyExpected": read_only_expected,
        "writeSqlForbidden": read_only_expected,
        "destructive": method in {"DELETE"} or any(word in meta.get("summary", "").lower() for word in ["clear", "reset", "uninstall", "force logout"]),
    }


def discover_host_endpoints() -> list:
    endpoints = []
    for path in sorted((root / "apps/lina-core/api").glob("**/v1/*.go")):
        relative = path.relative_to(root / "apps/lina-core/api")
        module = relative.parts[0]
        owner = {
            "owner": "core",
            "pluginId": None,
            "pluginType": None,
            "module": module,
            "moduleKey": f"core:{module}",
        }
        text = path.read_text(encoding="utf-8")
        for struct_name, body in struct_re.findall(text):
            endpoint = endpoint_from_struct(path, struct_name, body, owner)
            if endpoint:
                endpoints.append(endpoint)
    return endpoints


def discover_plugin_endpoints(plugins: dict) -> tuple[list, list]:
    endpoints = []
    skipped = []
    for plugin_dir in sorted((root / "apps/lina-plugins").glob("*")):
        if not plugin_dir.is_dir():
            continue
        plugin_yaml = plugin_dir / "plugin.yaml"
        if not plugin_yaml.exists():
            continue
        plugin = plugins.get(plugin_dir.name) or parse_plugin_yaml(plugin_yaml)
        plugin_id = plugin.get("id", plugin_dir.name)
        plugin_type = plugin.get("type", "source")
        owner_kind = "dynamic-plugin" if plugin_type == "dynamic" else "source-plugin"
        api_root = plugin_dir / "backend/api"
        route_count_before = len(endpoints)
        if api_root.exists():
            for path in sorted(api_root.glob("**/v1/*.go")):
                relative = path.relative_to(api_root)
                module = relative.parts[0]
                owner = {
                    "owner": owner_kind,
                    "pluginId": plugin_id,
                    "pluginType": plugin_type,
                    "module": module,
                    "moduleKey": f"{plugin_id}:{module}",
                }
                text = path.read_text(encoding="utf-8")
                for struct_name, body in struct_re.findall(text):
                    endpoint = endpoint_from_struct(path, struct_name, body, owner)
                    if endpoint:
                        endpoints.append(endpoint)
        if len(endpoints) == route_count_before:
            skipped.append({
                "pluginId": plugin_id,
                "pluginType": plugin_type,
                "path": plugin_dir.relative_to(root).as_posix(),
                "reason": "no backend API",
            })
    return endpoints, skipped


def group_modules(endpoints: list) -> list:
    grouped = {}
    for endpoint in endpoints:
        key = endpoint["moduleKey"]
        module = grouped.setdefault(key, {
            "key": key,
            "owner": endpoint["owner"],
            "pluginId": endpoint["pluginId"],
            "pluginType": endpoint["pluginType"],
            "module": endpoint["module"],
            "resources": sorted({endpoint["resource"]}),
            "endpointCount": 0,
            "endpoints": [],
        })
        module["endpoints"].append(endpoint)
        module["endpointCount"] += 1
        module["resources"] = sorted(set(module["resources"]) | {endpoint["resource"]})
    return [grouped[key] for key in sorted(grouped)]


plugins = {}
for plugin_yaml in sorted((root / "apps/lina-plugins").glob("*/plugin.yaml")):
    plugin = parse_plugin_yaml(plugin_yaml)
    plugins[plugin_yaml.parent.name] = plugin

host_endpoints = discover_host_endpoints()
plugin_endpoints, skipped_plugins = discover_plugin_endpoints(plugins)
all_endpoints = sorted(
    host_endpoints + plugin_endpoints,
    key=lambda item: (item["moduleKey"], item["publicPath"], item["method"], item["requestStruct"]),
)

catalog = {
    "schemaVersion": 1,
    "generatedAt": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    "runDir": str(run_dir),
    "apiVersionPrefix": api_prefix,
    "endpointCount": len(all_endpoints),
    "moduleCount": len({endpoint["moduleKey"] for endpoint in all_endpoints}),
    "modules": group_modules(all_endpoints),
    "endpoints": all_endpoints,
    "skippedPlugins": skipped_plugins,
}

catalog_path = run_dir / "catalog.json"
catalog_path.write_text(json.dumps(catalog, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(str(catalog_path))
PY
