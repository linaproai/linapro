---
id: "HIGH-plugin-demo-dynamic-dynamic-read-write-side-effect-dynamic-host-call-demo-api-v1-extensions-plugin-demo-dynamic-host-call-de"
severity: HIGH
module: "plugin-demo-dynamic:dynamic"
endpoint: "GET /api/v1/extensions/plugin-demo-dynamic/host-call-demo?skipNetwork=1"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "11d81e3c9a6b76fabe41a99fe23a1a4fec8ae640d2f1b6ea95422347e9d6294a"
---

# HIGH - plugin-demo-dynamic:dynamic - GET 演示接口执行持久化写入

## 问题描述

`GET /api/v1/extensions/plugin-demo-dynamic/host-call-demo?skipNetwork=1` 属于 GET 或读语义接口，但本次审查在 Trace-ID ``00273989df7cab18afb5f54108a1f508` (fallback from `server.log`; response header empty)` 中观察到 `6` 条写入 SQL，主要来源是动态插件演示逻辑中的运行时状态、节点状态或插件数据写入。该接口样本中的总 SQL 数为 `28`；即使响应耗时不高，读请求写库也会放大刷新、轮询、导出或详情访问的数据库写压力，并破坏 GET 请求应无副作用的接口约定。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/extensions/plugin-demo-dynamic/host-call-demo"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `28`，写入 SQL 数 `6`，审计签名 `read-write-side-effect:dynamic-host-call-demo`。

## 证据

- Trace-ID：``00273989df7cab18afb5f54108a1f508` (fallback from `server.log`; response header empty)`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/plugin-demo-dynamic-dynamic.md`
- 源码位置：`apps/lina-plugins/plugin-demo-dynamic/backend/internal/service/dynamic/dynamic_host_call_demo.go:36`
- SQL 总数：`28`
- 写入 SQL 数：`6`
- 指纹输入：`plugin-demo-dynamic:dynamic:GET:/api/v1/extensions/plugin-demo-dynamic/host-call-demo:HIGH:read-write-side-effect:dynamic-host-call-demo`

```sql
UPDATE `sys_online_session` SET `last_active_time`='2026-05-02 00:09:37' WHERE `token_id`='...'
INSERT INTO sys_plugin_state (...) VALUES ('plugin-demo-dynamic', 'host_call_demo_visit_count', '4', NOW(), NOW()) ON DUPLICATE KEY UPDATE ...
INSERT INTO `sys_plugin_node_state`(...)
UPDATE `sys_plugin_node_state` SET `current_state`='running',`error_message`='',`updated_at`='2026-05-02 00:09:37' WHERE `id`=4
DELETE FROM `sys_plugin_node_state` WHERE `id`=4
INSERT INTO `plugin_monitor_operlog`(...) VALUES('Dynamic Plugin Demo','Host calling capability demonstration',...)
```

## 改进方案

1. 将会产生运行时状态、节点状态、存储或结构化数据写入的演示逻辑改为 `POST`，保留 GET 仅返回只读摘要。
2. 分别复跑 `skipNetwork=1` 与正常路径，确认 GET 请求写入 SQL 数为 `0`。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/plugin-demo-dynamic-dynamic.md`，SQL 总数 `28`，写入 SQL 数 `6`，Trace-ID ``00273989df7cab18afb5f54108a1f508` (fallback from `server.log`; response header empty)`。
- 20260502：已通过 `FB-9` 修复，`host-call-demo` 改为 `POST` 并同步插件 apidoc `zh-CN`/`zh-TW` 翻译路径；验证命令 `go test ./apps/lina-plugins/plugin-demo-dynamic/backend/...` 与 apidoc JSON 语法校验通过。
