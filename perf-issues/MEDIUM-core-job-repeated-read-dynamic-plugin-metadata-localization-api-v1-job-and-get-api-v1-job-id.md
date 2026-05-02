---
id: "MEDIUM-core-job-repeated-read-dynamic-plugin-metadata-localization-api-v1-job-and-get-api-v1-job-id"
severity: MEDIUM
module: "core:job"
endpoint: "GET /api/v1/job?pageNum=1&pageSize=100`, `GET /api/v1/job/6"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "25e6ff1f3cffff215ac3fd05c300227fb6f0a3044fb8e52cb86452df40b4590a"
---

# MEDIUM - core:job - 请求内重复读取相同数据

## 问题描述

`GET /api/v1/job?pageNum=1&pageSize=100`, `GET /api/v1/job/6` 在同一请求链路中重复读取相同或高度重叠的数据，SQL 总数为 `26`, `7`，审计签名为 `repeated-read:dynamic-plugin-metadata-localization`。这类重复读取通常可以在请求内缓存、预加载或合并查询，否则会随着插件数量、菜单数量或日志行数增长而持续增加不必要的数据库访问。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/api/v1/job and GET /api/v1/job/{id}"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `26`, `7`，写入 SQL 数 `1`, `0`，审计签名 `repeated-read:dynamic-plugin-metadata-localization`。

## 证据

- Trace-ID：`005a5b233c7cab18ecb4f541215b30f3`, `80e54b273c7cab18eeb4f5413e285420`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-job.md`
- 源码位置：`apps/lina-core/internal/service/jobmgmt/jobmgmt_i18n.go:149`
- SQL 总数：`26`, `7`
- 写入 SQL 数：`1`, `0`
- 指纹输入：`core:job:GET:/api/v1/job and GET /api/v1/job/{id}:MEDIUM:repeated-read:dynamic-plugin-metadata-localization`

```sql
SELECT `id`,`plugin_id`,`name`,`version`,`type`,`installed`,`status`,`desired_state`,`current_state`,`generation`,`release_id`,`manifest_path`,`checksum`,`installed_at`,`enabled_at`,`disabled_at`,`remark`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL LIMIT 1
SELECT `id`,`plugin_id`,`release_version`,`type`,`runtime_kind`,`schema_version`,`min_host_version`,`max_host_version`,`status`,`manifest_path`,`package_path`,`checksum`,`manifest_snapshot`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin_release` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL ORDER BY `id` DESC LIMIT 1
SELECT `id`,`plugin_id`,`name`,`version`,`type`,`installed`,`status`,`desired_state`,`current_state`,`generation`,`release_id`,`manifest_path`,`checksum`,`installed_at`,`enabled_at`,`disabled_at`,`remark`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL LIMIT 1
SELECT `id`,`plugin_id`,`release_version`,`type`,`runtime_kind`,`schema_version`,`min_host_version`,`max_host_version`,`status`,`manifest_path`,`package_path`,`checksum`,`manifest_snapshot`,`created_at`,`updated_at`,`deleted_at` FROM `sys_plugin_release` WHERE (`plugin_id`='monitor-server' AND `type`='dynamic') AND `deleted_at` IS NULL ORDER BY `id` DESC LIMIT 1
```

## 改进方案

1. 在请求链路内复用已读取的数据，或将重复读取合并为一次批量查询，避免相同插件、菜单、发布版本或本地化元数据被反复加载。
2. 复跑 `GET /api/v1/job?pageNum=1&pageSize=100`, `GET /api/v1/job/6`，确认重复读取的 SQL 被折叠，且 GET 请求仍无写入 SQL。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-job.md`，SQL 总数 `26`, `7`，写入 SQL 数 `1`, `0`，Trace-ID `005a5b233c7cab18ecb4f541215b30f3`, `80e54b273c7cab18eeb4f5413e285420`。
- 20260502：已通过 `FB-10` 修复，作业列表、详情与本地化关键字搜索复用 handler source text 缓存，并由动态插件 i18n release 缓存折叠重复发布版本读取；验证命令 `go test ./apps/lina-core/internal/service/jobmgmt ./apps/lina-core/internal/service/i18n` 通过。
