---
id: "MEDIUM-core-menu-repeated-read-full-menu-plugin-runtime-api-v1-menu"
severity: MEDIUM
module: "core:menu"
endpoint: "GET /api/v1/menu"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "ce1b6e99f2214e31c0e9b642fc32804f8b6c3c76c1ebbf8b37fbc9b1fb38ddc9"
---

# MEDIUM - core:menu - 请求内重复读取相同数据

## 问题描述

`GET /api/v1/menu` 在同一请求链路中重复读取相同或高度重叠的数据，SQL 总数为 `8`，审计签名为 `repeated-read-full-menu-plugin-runtime`。这类重复读取通常可以在请求内缓存、预加载或合并查询，否则会随着插件数量、菜单数量或日志行数增长而持续增加不必要的数据库访问。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/menu"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `8`，写入 SQL 数 `0`，审计签名 `repeated-read-full-menu-plugin-runtime`。

## 证据

- Trace-ID：`b025b64de67bab1844b4f5413db6ce97`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-menu.md`
- 源码位置：`apps/lina-core/internal/service/role/role_user_access.go:204`, `apps/lina-core/internal/service/menu/menu.go:101`, `apps/lina-core/internal/service/plugin/internal/integration/integration.go:380`
- SQL 总数：`8`
- 写入 SQL 数：`0`
- 指纹输入：`core:menu:GET:/api/v1/menu:MEDIUM:repeated-read-full-menu-plugin-runtime`

```sql
SELECT ... FROM `sys_menu` WHERE (`status`=1) AND `deleted_at` IS NULL ORDER BY `id` ASC
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
SELECT ... FROM `sys_menu` WHERE `deleted_at` IS NULL ORDER BY `parent_id` ASC,`sort` ASC,`id` ASC
SELECT ... FROM `sys_plugin` WHERE `deleted_at` IS NULL ORDER BY `plugin_id` ASC
```

## 改进方案

1. 在请求链路内复用已读取的数据，或将重复读取合并为一次批量查询，避免相同插件、菜单、发布版本或本地化元数据被反复加载。
2. 复跑 `GET /api/v1/menu`，确认重复读取的 SQL 被折叠，且 GET 请求仍无写入 SQL。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-menu.md`，SQL 总数 `8`，写入 SQL 数 `0`，Trace-ID `b025b64de67bab1844b4f5413db6ce97`。
- 20260502：已通过 `FB-12` 修复，插件过滤层优先复用已加载启用状态快照，冷路径读取后回填共享快照以避免重复读取插件运行时状态；验证命令 `go test ./apps/lina-core/internal/service/plugin/internal/integration ./apps/lina-core/internal/service/middleware ./apps/lina-core/internal/controller/menu` 通过。
