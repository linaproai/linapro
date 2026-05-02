---
id: "MEDIUM-core-jobgroup-small-sample-n-plus-one-job-count-per-group-job-group"
severity: MEDIUM
module: "core:jobgroup"
endpoint: "GET /job-group"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "032fd86a8965583a01621963ce4383f6e03f1f077ce80a961ca30c8cf3e061a0"
---

# MEDIUM - core:jobgroup - 小样本下存在 N+1 查询风险

## 问题描述

`GET /job-group` 在当前样本中已经出现按记录逐项补充统计的查询模式，SQL 总数为 `5`。虽然本次数据规模不大，审查仍将其归类为小样本 N+1 风险；当分组、角色或资源数量增长时，该模式会产生更多数据库往返。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/job-group"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `5`，写入 SQL 数 `0`，审计签名 `small-sample-n-plus-one:job-count-per-group`。

## 证据

- Trace-ID：`88014497507cab1810b5f5415cf1697b`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-jobgroup.md`
- 源码位置：`apps/lina-core/internal/service/jobmgmt/jobmgmt_group.go:65`
- SQL 总数：`5`
- 写入 SQL 数：`0`
- 指纹输入：`core:jobgroup:GET:/job-group:MEDIUM:small-sample-n-plus-one:job-count-per-group`

```sql
SELECT COUNT(1) FROM `sys_job_group` WHERE `deleted_at` IS NULL
SELECT `id`,`code`,`name`,`remark`,`sort_order`,`is_default`,`created_at`,`updated_at`,`deleted_at` FROM `sys_job_group` WHERE `deleted_at` IS NULL ORDER BY `sort_order` ASC LIMIT 0,100
SELECT COUNT(1) FROM `sys_job` WHERE (`group_id`=1) AND `deleted_at` IS NULL
SELECT COUNT(1) FROM `sys_job` WHERE (`group_id`=2) AND `deleted_at` IS NULL
```

## 改进方案

1. 将逐行补充查询改为批量预加载，例如使用 `WHERE IN`、JOIN、聚合子查询或请求内缓存一次性取回关联数据。
2. 使用压力数据复跑 `GET /job-group`，确认 SQL 数量不再随返回行数线性增长。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-jobgroup.md`，SQL 总数 `5`，写入 SQL 数 `0`，Trace-ID `88014497507cab1810b5f5415cf1697b`。
- 20260502：已通过 `FB-11` 修复，作业分组列表使用一次 `GROUP BY group_id` 批量统计作业数量；验证命令 `go test ./apps/lina-core/internal/service/jobmgmt` 通过。
