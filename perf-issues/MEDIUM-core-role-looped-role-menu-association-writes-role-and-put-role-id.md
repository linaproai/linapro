---
id: "MEDIUM-core-role-looped-role-menu-association-writes-role-and-put-role-id"
severity: MEDIUM
module: "core:role"
endpoint: "POST /role`, `PUT /role/{id}"
status: fixed
first_seen_run: "20260501-233924"
last_seen_run: "20260501-233924"
seen_count: 1
fingerprint: "b4d67400649c61880da8f3638ad9ea90f8ad70ecfc66c5aad3583a668f434e79"
---

# MEDIUM - core:role - 角色菜单关联逐条写入

## 问题描述

`POST /role`, `PUT /role/{id}` 在角色菜单关联写入时按菜单逐条执行写操作，本次样本 SQL 总数为 `16`, `18`，写入 SQL 数为 `5`, `6`。当角色关联的菜单数量变多时，该实现会产生更多数据库写入往返，应改为批量插入或批量替换。

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id 20260501-233924`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/20260501-233924`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/20260501-233924`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url>/role and PUT /role/{id}"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/20260501-233924/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：SQL 总数 `16`, `18`，写入 SQL 数 `5`, `6`，审计签名 `looped-role-menu-association-writes`。

## 证据

- Trace-ID：`4899be1bfa7bab1887b4f541317cd1e7`, `08ddda1efa7bab188ab4f541fbd40ed0`
- 审计文件：`temp/lina-perf-audit/20260501-233924/audits/core-role.md`
- 源码位置：`apps/lina-core/internal/service/role/role.go:359`, `apps/lina-core/internal/service/role/role.go:446`
- SQL 总数：`16`, `18`
- 写入 SQL 数：`5`, `6`
- 指纹输入：`core:role:POST:/role and PUT /role/{id}:MEDIUM:looped-role-menu-association-writes`

```sql
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(37,1)
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(37,2)
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(37,3)
DELETE FROM `sys_role_menu` WHERE `role_id`=38
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(38,1)
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(38,2)
INSERT INTO `sys_role_menu`(`role_id`,`menu_id`) VALUES(38,3)
```

## 改进方案

1. 将角色菜单关联的逐条写入改为批量插入、批量替换或一次性差异同步。
2. 使用包含大量 `menuIds` 的审计夹具复跑创建和更新接口，确认关联写入次数保持稳定。

## 历史记录

- 20260501-233924：本次审查发现，审计文件 `temp/lina-perf-audit/20260501-233924/audits/core-role.md`，SQL 总数 `16`, `18`，写入 SQL 数 `5`, `6`，Trace-ID `4899be1bfa7bab1887b4f541317cd1e7`, `08ddda1efa7bab188ab4f541fbd40ed0`。
- 20260502：已通过 `FB-14` 修复，角色创建与更新批量插入 `sys_role_menu` 关联并过滤非法/重复菜单 ID；验证命令 `go test ./apps/lina-core/internal/service/role` 通过。
