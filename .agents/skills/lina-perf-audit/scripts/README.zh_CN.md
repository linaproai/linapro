# LinaPro 性能审查脚本

该 skill 内置的 `scripts/` 目录包含手动 `lina-perf-audit` skill 使用的辅助脚本。它们负责准备本地审计专用后端运行环境、安装内置插件，并在审计结束后恢复宿主 logger 配置。

这些脚本在 `.agents/skills/lina-perf-audit/scripts/` 中维护，并应从仓库根目录运行。

## 脚本

- `setup-audit-env.sh`
  - 通过 `make stop` 停止已有本地服务。
  - 将 `apps/lina-core/manifest/config/config.yaml` 中的 logger 配置备份到 `<run-dir>/logger-backup.json`。
  - 临时把 `logger.path` patch 为 run 目录，把 `logger.file` patch 为 `server.log`。
  - 构建动态插件 Wasm 产物、准备宿主嵌入资源，并只启动后端服务。
  - 等待 `http://127.0.0.1:8080/api/v1/health` 就绪。
  - 使用 `admin/admin123` 登录，把 token 写入 `<run-dir>/token.txt`，并把 `Trace-ID` 验证写入 `<run-dir>/trace-id-check.txt`。

- `prepare-builtin-plugins.sh`
  - 读取 `<run-dir>/token.txt`。
  - 扫描 `apps/lina-plugins/*/plugin.yaml`。
  - 调用宿主插件 API，同步、安装并启用所有发现的内置插件。
  - 将进度和失败详情写入 `<run-dir>/plugins.json`。

- `scan-endpoints.sh`
  - 扫描宿主与内置插件 API DTO。
  - 解析 `g.Meta` / `gmeta.Meta` 路由元数据。
  - 写入 `<run-dir>/catalog.json`。

- `probe-fixtures.sh`
  - 读取 `<run-dir>/catalog.json` 与 token。
  - 探测安全的 `GET` 列表/详情接口。
  - 写入 `<run-dir>/fixtures.json`，并在声明的 DTO 路由不可访问时失败。

- `stress-fixture.sh`
  - 通过 MySQL 直连插入审计专用压力数据。
  - 使用幂等写入，不修改交付 SQL 目录。
  - 写入 `<run-dir>/stress-fixture.json`。

- `aggregate-reports.sh`
  - 读取 `<run-dir>/audits/*.md` 中的 Stage 1 报告。
  - 写入 `<run-dir>/SUMMARY.md` 与 `<run-dir>/meta.json`。
  - 在仓库根目录 `perf-issues/` 下创建或更新持久问题卡片。
  - 重新生成包含 open 与 in-progress 卡片的 `perf-issues/INDEX.md`。

- `restore-audit-env.sh`
  - 从 `<run-dir>/logger-backup.json` 恢复 `logger.path` 和 `logger.file`。
  - 调用 `make stop`。
  - 可在成功路径或失败路径重复调用。

## 使用方式

```bash
run_id="$(date +%Y%m%d-%H%M%S)"
run_dir="temp/lina-perf-audit/${run_id}"

bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id "${run_id}"
bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/scan-endpoints.sh --run-dir "${run_dir}"
bash .agents/skills/lina-perf-audit/scripts/probe-fixtures.sh --run-dir "${run_dir}"

# 子 agent 在 "${run_dir}/audits/" 下写入报告后执行汇总。
bash .agents/skills/lina-perf-audit/scripts/aggregate-reports.sh --run-dir "${run_dir}"

# 审计结束后必须恢复，包括失败路径。
bash .agents/skills/lina-perf-audit/scripts/restore-audit-env.sh --run-dir "${run_dir}"
```

脚本仅依赖 Bash、`curl`、Go 工具链、MySQL CLI 和 Python 3 标准库 JSON 支持，不依赖 `jq` 或 PyYAML，也不会修改交付 SQL 文件或应用源码。
