---
id: "HIGH-core-joblog-n-plus-one-joblog-list-dynamic-plugin-i18n-api-v1-job-log"
severity: HIGH
module: "core:joblog"
endpoint: "GET /api/v1/job/log?pageNum=1&pageSize=100"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "43c12c7cb2ea795463a7005b5a7cac5183ba8ab5590efcca19b32084d90a863f"
---

# HIGH - core:joblog - 列表查询存在 N+1 查询

## 问题描述

`GET /api/v1/job/log?pageNum=1&pageSize=100` 在压力数据下执行了 `80` 条 SQL，审计签名为 `n-plus-one:joblog-list-dynamic-plugin-i18n`。SQL 调用数量会随返回行数或需要本地化的行数增长，说明列表查询路径存在 N+1 查询或逐行补充查询风险。当数据量继续增长时，该接口的数据库往返次数和响应耗时会同步上升。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/job/log"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `80`，写入 SQL 数 `0`，审计签名 `n-plus-one:joblog-list-dynamic-plugin-i18n`。

## 证据

- Trace-ID：`e0fa7427487cab18feb4f541c6e5debe`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-joblog.md`
- 源码位置：`apps/lina-core/internal/service/jobmgmt/jobmgmt_log.go:68`
- SQL 总数：`80`
- 写入 SQL 数：`0`
- 指纹输入：`core:joblog:GET:/api/v1/job/log:HIGH:n-plus-one:joblog-list-dynamic-plugin-i18n`

```sql
SELECT COUNT(1) FROM `sys_job_log`
SELECT `id`,`job_id`,`job_snapshot`,`node_id`,`trigger`,`params_snapshot`,`start_at`,`end_at`,`duration_ms`,`status`,`err_msg`,`result_json`,`created_at` FROM `sys_job_log` ORDER BY `start_at` DESC LIMIT 0,100
SELECT `id`,`name`,`handler_ref`,`is_builtin` FROM `sys_job` WHERE (`id` IN (...100 log rows...)) AND `deleted_at` IS NULL
-- repeated 36 times in this trace:
SELECT `id`,`plugin_id`,`name`,`version`,`type`,`installed`,`status`,`desired_state`,`current_state`,`generation`,`release_id`,`manifest_path`,`checksum`,`installed_at`,`enabled_at`,`disabled_at`,`remark`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL LIMIT 1
-- repeated 36 times in this trace:
SELECT `id`,`plugin_id`,`release_version`,`type`,`runtime_kind`,`schema_version`,`min_host_version`,`max_host_version`,`status`,`manifest_path`,`package_path`,`checksum`,`manifest_snapshot`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin_release` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL ORDER BY `id` DESC LIMIT 1
```

## 改进方案

1. 将逐行补充查询改为批量预加载，例如使用 `WHERE IN`、JOIN、聚合子查询或请求内缓存一次性取回关联数据。
2. 使用压力数据复跑 `GET /api/v1/job/log?pageNum=1&pageSize=100`，确认 SQL 数量不再随返回行数线性增长。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-joblog.md`，SQL 总数 `80`，写入 SQL 数 `0`，Trace-ID `e0fa7427487cab18feb4f541c6e5debe`。
- 20260502：已通过 `FB-8` 修复，作业日志列表与详情投影复用请求内 handler source text 缓存；验证命令 `go test ./apps/lina-core/internal/service/jobmgmt` 通过。
