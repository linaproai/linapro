# LinaPro 性能问题索引

- 生成来源 run：`20260501-233924`
- 活跃问题卡片数：`0`
- 已修复问题卡片数：`8`

| 严重度 | 模块 | 接口 | 状态 | 最近发现 | 出现次数 | 卡片 |
|---|---|---|---|---|---:|---|
| HIGH | `core:joblog` | `GET /api/v1/job/log?pageNum=1&pageSize=100` | 已修复 | `20260501-233924` | 1 | [perf-issues/HIGH-core-joblog-n-plus-one-joblog-list-dynamic-plugin-i18n-api-v1-job-log.md](perf-issues/HIGH-core-joblog-n-plus-one-joblog-list-dynamic-plugin-i18n-api-v1-job-log.md) |
| HIGH | `plugin-demo-dynamic:dynamic` | `GET /api/v1/extensions/plugin-demo-dynamic/host-call-demo?skipNetwork=1` | 已修复 | `20260501-233924` | 1 | [perf-issues/HIGH-plugin-demo-dynamic-dynamic-read-write-side-effect-dynamic-host-call-demo-api-v1-extensions-plugin-demo-dynamic-host-call-de.md](perf-issues/HIGH-plugin-demo-dynamic-dynamic-read-write-side-effect-dynamic-host-call-demo-api-v1-extensions-plugin-demo-dynamic-host-call-de.md) |
| MEDIUM | `core:job` | `GET /api/v1/job?pageNum=1&pageSize=100`, `GET /api/v1/job/6` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-core-job-repeated-read-dynamic-plugin-metadata-localization-api-v1-job-and-get-api-v1-job-id.md](perf-issues/MEDIUM-core-job-repeated-read-dynamic-plugin-metadata-localization-api-v1-job-and-get-api-v1-job-id.md) |
| MEDIUM | `core:jobgroup` | `GET /job-group` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-core-jobgroup-small-sample-n-plus-one-job-count-per-group-job-group.md](perf-issues/MEDIUM-core-jobgroup-small-sample-n-plus-one-job-count-per-group-job-group.md) |
| MEDIUM | `core:menu` | `GET /api/v1/menu` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-core-menu-repeated-read-full-menu-plugin-runtime-api-v1-menu.md](perf-issues/MEDIUM-core-menu-repeated-read-full-menu-plugin-runtime-api-v1-menu.md) |
| MEDIUM | `core:plugin` | `GET /api/v1/plugins` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-core-plugin-repeated-read-plugin-list-dynamic-artifact-state-api-v1-plugins.md](perf-issues/MEDIUM-core-plugin-repeated-read-plugin-list-dynamic-artifact-state-api-v1-plugins.md) |
| MEDIUM | `core:role` | `POST /role`, `PUT /role/{id}` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-core-role-looped-role-menu-association-writes-role-and-put-role-id.md](perf-issues/MEDIUM-core-role-looped-role-menu-association-writes-role-and-put-role-id.md) |
| MEDIUM | `monitor-operlog:operlog` | `GET /api/v1/operlog?pageNum=1&pageSize=3` | 已修复 | `20260501-233924` | 1 | [perf-issues/MEDIUM-monitor-operlog-operlog-repeated-same-data-reads-operlog-list-localization-plugin-metadata-api-v1-operlog.md](perf-issues/MEDIUM-monitor-operlog-operlog-repeated-same-data-reads-operlog-list-localization-plugin-metadata-api-v1-operlog.md) |
