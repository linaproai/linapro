---
id: "MEDIUM-core-plugin-repeated-read-plugin-list-dynamic-artifact-state-api-v1-plugins"
severity: MEDIUM
module: "core:plugin"
endpoint: "GET /api/v1/plugins"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "6444fbcf0ab0f9b52c86fb020ad45a32a3dade3bab9171cfedc49d27ca366e26"
---

# MEDIUM - core:plugin - 请求内重复读取相同数据

## 问题描述

`GET /api/v1/plugins` 在同一请求链路中重复读取相同或高度重叠的数据，SQL 总数为 `8`，审计签名为 `repeated-read:plugin-list-dynamic-artifact-state`。这类重复读取通常可以在请求内缓存、预加载或合并查询，否则会随着插件数量、菜单数量或日志行数增长而持续增加不必要的数据库访问。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/plugins"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `8`，写入 SQL 数 `0`，审计签名 `repeated-read:plugin-list-dynamic-artifact-state`。

## 证据

- Trace-ID：`e01af9064c7cab1808b5f54106e91fa6`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-plugin.md`
- 源码位置：`apps/lina-core/internal/service/plugin/plugin_list.go:108`
- SQL 总数：`8`
- 写入 SQL 数：`0`
- 指纹输入：`core:plugin:GET:/api/v1/plugins:MEDIUM:repeated-read:plugin-list-dynamic-artifact-state`

```sql
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
SELECT ... FROM `sys_plugin_release` WHERE `deleted_at` IS NULL
SELECT TABLE_NAME AS table_name, TABLE_COMMENT AS table_comment FROM information_schema.TABLES WHERE TABLE_SCHEMA='linapro' AND TABLE_NAME IN('plugin_demo_dynamic_record','sys_plugin_node_state')
SELECT ... FROM `sys_plugin` WHERE (`plugin_id`='plugin-demo-dynamic' AND `type`='dynamic') AND `deleted_at` IS NULL LIMIT 1
SELECT ... FROM `sys_plugin_release` WHERE (`id`=8) AND `deleted_at` IS NULL LIMIT 1
SELECT ... FROM `sys_plugin` WHERE (`plugin_id`='plugin-demo-dynamic' AND `type`='dynamic') AND `deleted_at` IS NULL LIMIT 1
SELECT ... FROM `sys_plugin_release` WHERE (`id`=8) AND `deleted_at` IS NULL LIMIT 1
```

## 改进方案

1. 在请求链路内复用已读取的数据，或将重复读取合并为一次批量查询，避免相同插件、菜单、发布版本或本地化元数据被反复加载。
2. 复跑 `GET /api/v1/plugins`，确认重复读取的 SQL 被折叠，且 GET 请求仍无写入 SQL。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-plugin.md`，SQL 总数 `8`，写入 SQL 数 `0`，Trace-ID `e01af9064c7cab1808b5f54106e91fa6`。
- 20260502：已通过 `FB-13` 修复，动态插件 i18n release 元数据加入可精确失效缓存，插件列表投影不再重复读取同一动态插件与发布版本状态；验证命令 `go test ./apps/lina-core/internal/service/plugin/... ./apps/lina-core/internal/service/i18n` 通过。
