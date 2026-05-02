---
id: "MEDIUM-monitor-operlog-operlog-repeated-same-data-reads-operlog-list-localization-plugin-metadata-api-v1-operlog"
severity: MEDIUM
module: "monitor-operlog:operlog"
endpoint: "GET /api/v1/operlog?pageNum=1&pageSize=3"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "e21f1e57578ffe3adaa720437776d5557a92a7d4a090295a1a4f5ddf5caff49f"
---

# MEDIUM - monitor-operlog:operlog - 请求内重复读取相同数据

## 问题描述

`GET /api/v1/operlog?pageNum=1&pageSize=3` 在同一请求链路中重复读取相同或高度重叠的数据，SQL 总数为 `9`，审计签名为 `repeated-same-data-reads:operlog-list-localization-plugin-metadata`。这类重复读取通常可以在请求内缓存、预加载或合并查询，否则会随着插件数量、菜单数量或日志行数增长而持续增加不必要的数据库访问。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/operlog"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `9`，写入 SQL 数 `0`，审计签名 `repeated-same-data-reads:operlog-list-localization-plugin-metadata`。

## 证据

- Trace-ID：`700a91648b7cab185db5f54172f50241`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/monitor-operlog-operlog.md`
- 源码位置：`apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go:260`
- SQL 总数：`9`
- 写入 SQL 数：`0`
- 指纹输入：`monitor-operlog:operlog:GET:/api/v1/operlog:MEDIUM:repeated-same-data-reads:operlog-list-localization-plugin-metadata`

```sql
SELECT `token_id`,`user_id`,`username`,`dept_name`,`ip`,`browser`,`os`,`login_time`,`last_active_time` FROM `sys_online_session` WHERE `token_id`='...' LIMIT 1
SELECT COUNT(1) FROM `plugin_monitor_operlog`
SELECT `id`,`title`,`oper_summary`,`route_owner`,`route_method`,`route_path`,`route_doc_key`,`oper_type`,`method`,`request_method`,`oper_name`,`oper_url`,`oper_ip`,`oper_param`,`json_result`,`status`,`error_msg`,`cost_time`,`oper_time` FROM `plugin_monitor_operlog` ORDER BY `oper_time` DESC LIMIT 0,3
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
SELECT ... FROM `sys_plugin_release` WHERE `deleted_at` IS NULL
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
SELECT ... FROM `sys_plugin_release` WHERE `deleted_at` IS NULL
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
-- truncated for persistent issue card
```

## 改进方案

1. 在请求链路内复用已读取的数据，或将重复读取合并为一次批量查询，避免相同插件、菜单、发布版本或本地化元数据被反复加载。
2. 复跑 `GET /api/v1/operlog?pageNum=1&pageSize=3`，确认重复读取的 SQL 被折叠，且 GET 请求仍无写入 SQL。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/monitor-operlog-operlog.md`，SQL 总数 `9`，写入 SQL 数 `0`，Trace-ID `700a91648b7cab185db5f54172f50241`。
- 20260502：已通过 `FB-15` 修复，操作日志列表改为批量调用 `ResolveRouteTexts`，一次加载 apidoc catalog 后本地化多条记录；验证命令 `go test ./apps/lina-core/internal/service/apidoc ./apps/lina-core/internal/cmd ./apps/lina-plugins/monitor-operlog/backend/internal/service/operlog` 通过。
