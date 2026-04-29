## 1. SQL 结构与索引调整

- [ ] 1.1 修改 `apps/lina-core/manifest/sql/001-project-init.sql`：在 `sys_user` 上新增 `KEY idx_status (status)`、`KEY idx_phone (phone)`、`KEY idx_created_at (created_at)`，保持幂等
- [ ] 1.2 修改 `apps/lina-core/manifest/sql/002-dict-dept-post.sql`：为 `sys_dict_type` 与 `sys_dict_data` 新增 `deleted_at DATETIME DEFAULT NULL` 字段，与现有 `dict-management` 规范保持一致
- [ ] 1.3 修改 `apps/lina-core/manifest/sql/008-menu-role-management.sql`：在 `sys_user_role` 上新增 `KEY idx_role_id (role_id)`，在 `sys_role_menu` 上新增 `KEY idx_menu_id (menu_id)`
- [ ] 1.4 定位并修改 `sys_online_session` 所在 SQL 文件，新增 `KEY idx_last_active_time (last_active_time)`
- [ ] 1.5 修改 `apps/lina-core/manifest/sql/014-scheduled-job-management.sql`：移除 `CONSTRAINT fk_sys_job_group_id`，新增 `KEY idx_group_id (group_id)`
- [ ] 1.6 在 `apps/lina-core` 执行 `make init`，验证全部 SQL 幂等且新增结构正确
- [ ] 1.7 在 `apps/lina-core` 执行 `make dao` 重新生成 dao/do/entity，确认 `SysDictType` / `SysDictData` 实体新增 `DeletedAt` 字段

## 2. 后端事务正确性修复

- [ ] 2.1 修改 `apps/lina-core/internal/service/user/user.go` 的 `Delete`：使用 `dao.SysUser.Transaction(ctx, ...)` 包裹用户软删除、组织清理、角色关联清理；事务内任意失败 `return err`；提交后再调用 `NotifyAccessTopologyChanged`
- [ ] 2.2 修改 `apps/lina-core/internal/service/role/role.go` 的 `Delete`：事务内 `sys_role_menu`、`sys_user_role` 清理失败由 `Warningf` 改为 `return err` 触发回滚
- [ ] 2.3 重构 `apps/lina-core/internal/service/role/role.go` 的 `AssignUsers`：在单事务内构造 `[]do.SysUserRole` 后一次 `Insert(slice)` 批量插入，失败整体回滚；移除现有逐条 `Warningf` 兜底
- [ ] 2.4 修改 `apps/lina-core/internal/service/menu/menu.go` 的 `Delete`：事务内 `sys_role_menu` 清理失败由 `Warningf` 改为 `return err` 触发回滚
- [ ] 2.5 在 `apps/lina-core/internal/service/role` 与 `user` 的 `*_code.go` 中确认所需 `bizerr.Code` 已存在（如 `CodeUserBuiltinAdminDeleteDenied`、`CodeUserCurrentDeleteDenied`、`CodeRoleBuiltinDeleteDenied`），缺失则补齐定义并同步 `manifest/i18n/<locale>/error.json`
- [ ] 2.6 增加 GoFrame 单元/集成测试覆盖：`user.Delete` 关联清理失败回滚、`role.Delete` 关联清理失败回滚、`role.AssignUsers` 中途失败整体回滚

## 3. 后端批量删除接口

- [ ] 3.1 在 `apps/lina-core/api/user/v1/` 下新增独立文件定义 `BatchDeleteReq`/`BatchDeleteRes`：`DELETE /api/v1/user`，Query 参数 `Ids []int json:"ids" v:"required|min-length:1"`，英文 `dc`/`eg`，`g.Meta` 标注 `permission:"user:remove"`
- [ ] 3.2 在 `apps/lina-core/api/role/v1/` 下新增独立文件定义 `BatchDeleteReq`/`BatchDeleteRes`：`DELETE /api/v1/role`，权限标签 `role:remove`
- [ ] 3.3 在 `apps/lina-core` 执行 `make ctrl` 重新生成控制器骨架，并在 `controller/user/` 与 `controller/role/` 中实现新的 `BatchDelete` 委托 service
- [ ] 3.4 在 `apps/lina-core/internal/service/user/user.go` 新增 `BatchDelete(ctx, ids []int) error`：事务内复用 `Delete` 的全部保护策略（内置管理员、当前用户），任一拒绝即整体回滚；提交后只发一次 `NotifyAccessTopologyChanged`
- [ ] 3.5 在 `apps/lina-core/internal/service/role/role.go` 新增 `BatchDelete(ctx, ids []int) error`：事务内复用 `Delete` 的全部保护策略，提交后发通知
- [ ] 3.6 增加 service 层批量删除单元测试覆盖：成功批量删除、含内置管理员被拒、含当前用户被拒、空列表参数校验失败

## 4. 后端性能与默认值改造

- [ ] 4.1 重写 `apps/lina-core/internal/service/menu/menu.go` 的 `isDescendant`：一次性 `dao.SysMenu.Ctx(ctx).Scan(&all)` 加载父子映射后内存 BFS/DFS；保持函数签名不变
- [ ] 4.2 在 `apps/lina-core/internal/service/cron/cron_managed_jobs.go` 删除 `defaultManagedJobTimezone = "Asia/Shanghai"` 常量；改为读取 `g.Cfg().Get(ctx, "scheduler.defaultTimezone", "UTC")`
- [ ] 4.3 在 `apps/lina-core/manifest/config/config.yaml` 下新增 `scheduler.defaultTimezone: "UTC"` 配置项及英文注释
- [ ] 4.4 增加 service 层单测覆盖菜单 `isDescendant` 内存判定的正确性边界（自身不算后代、跨层判定、不存在 id）

## 5. 上传路由鉴权与运维端点

- [ ] 5.1 在 `apps/lina-core/internal/cmd/cmd_http.go` 中将 `bindUploadRoutes` 的注册移入受保护路由分组（与 `bindProtectedStaticAPIRoutes` 同层），挂载 `Auth` 与 `Permission` 中间件
- [ ] 5.2 确定上传路由的最终权限标签（推荐 `file:read`，apply 阶段对照 `file` 模块现有菜单/按钮权限确认），通过路由元数据声明权限要求
- [ ] 5.3 在 `bindPublicStaticAPIRoutes` 中新增 `GET /api/v1/health` 路由：执行 `dao.SysUser.Ctx(ctx).Limit(1).Count()` 探活；正常返回 `{status:"ok", mode:"<single|master|slave>"}`，失败返回 `503 {status:"unavailable", reason}`
- [ ] 5.4 在 `apps/lina-core/manifest/config/config.yaml` 新增 `health.timeout: "2s"` 与 `shutdown.timeout: "30s"`；后端读取时统一解析为 `time.Duration`
- [ ] 5.5 在 `apps/lina-core/internal/cmd/cmd.go`（或 `cmd_http.go`）入口注册 `signal.NotifyContext` 监听 `SIGTERM` / `SIGINT`；按"HTTP Server `Shutdown` → cron 调度器 `Stop` → 数据库连接池 `Close`"顺序优雅关停，受 `shutdown.timeout` 约束，超时 warning 后非零退出
- [ ] 5.6 如 `internal/service/cron` 不存在 `Stop(ctx)` 方法，则在该组件主文件补齐，确保已注册任务在收到关停信号时停止接受新触发并等待在途任务完成

## 6. 残留空包清理

- [ ] 6.1 在仓库内 grep 确认 `apps/lina-core/pkg/auditi18n` 与 `apps/lina-core/pkg/audittype` 当前没有任何 import 引用
- [ ] 6.2 删除空目录 `apps/lina-core/pkg/auditi18n/` 与 `apps/lina-core/pkg/audittype/`

## 7. 前端批量操作与适配

- [ ] 7.1 在 `apps/lina-vben/apps/web-antd/src/api/system/user/index.ts` 新增 `userBatchDelete(ids: number[])`，使用 `requestClient.delete('/user', { params: { ids: ids.join(',') } })`
- [ ] 7.2 在 `apps/lina-vben/apps/web-antd/src/api/system/role/index.ts` 新增 `roleBatchDelete(ids: number[])`
- [ ] 7.3 修改 `apps/lina-vben/apps/web-antd/src/views/system/user/index.vue` 的批量删除处理：把 `for (const id of ids) await userDelete(id)` 替换为单次 `await userBatchDelete(ids)`，错误展示沿用现有 bizerr 通用处理
- [ ] 7.4 修改 `apps/lina-vben/apps/web-antd/src/views/system/role/index.vue` 的批量删除处理：替换为 `roleBatchDelete(ids)`

## 8. 前端轮询、缓存与语言切换优化

- [ ] 8.1 修改 `apps/lina-vben/apps/web-antd/src/views/monitor/server/index.vue`：使用 `@vueuse/core` 的 `useIntervalFn` + `useDocumentVisibility` 实现 30s 自动刷新；标签页隐藏时 `pause`，可见时立即触发一次刷新；组件卸载时显式 `stop`
- [ ] 8.2 修改 `apps/lina-vben/apps/web-antd/src/store/message.ts`：把原始 `setInterval` 改造为可见性感知轮询，hidden 时暂停、visible 时立即刷新一次；登出或 store 销毁时显式停止
- [ ] 8.3 修改 `apps/lina-vben/apps/web-antd/src/router/guard.ts` 的 `loadedPaths`：改为有界 LRU（默认上限 50 条），命中时移到队尾，超出时淘汰最旧
- [ ] 8.4 修改 `apps/lina-vben/apps/web-antd/src/bootstrap.ts` 中 `watch(preferences.app.locale, ...)`：保留 `syncPublicFrontendSettings` 与 `useDictStore().resetCache()`，移除 `refreshAccessibleState(router)` 调用
- [ ] 8.5 在 `apps/lina-vben/apps/web-antd/src/router/routes/modules/` 下扫描所有 `meta.title`，确认全部使用 i18n key 或函数返回 `$t(...)` 形式；如有静态字符串 hardcode 即修复

## 9. 单测、E2E 与回归验证

- [ ] 9.1 创建 `hack/tests/e2e/auth/TC0145-health-endpoint-anonymous-access.ts`：匿名访问 `GET /api/v1/health` 验证 200 + `status="ok"`；按 `lina-e2e` 技能约定生成
- [ ] 9.2 创建 `hack/tests/e2e/iam/user/TC0146-user-batch-delete-single-request.ts`：在用户列表选中多个用户执行批量删除，断言只发起一次 `DELETE /api/v1/user?ids=...` 请求且数据被批量移除
- [ ] 9.3 创建 `hack/tests/e2e/iam/role/TC0147-role-batch-delete-single-request.ts`：角色列表批量删除走单次接口
- [ ] 9.4 创建 `hack/tests/e2e/monitor/server/TC0148-server-monitor-visibility-aware-polling.ts`：验证页面可见时 30s 自动刷新、隐藏时停止、恢复后立即刷新
- [ ] 9.5 创建 `hack/tests/e2e/settings/file/TC0149-uploads-route-requires-auth.ts`：未登录访问 `/api/v1/uploads/<path>` 应返回未认证；已登录无权限返回无权限；登录有权限可下载
- [ ] 9.6 在 `hack/tests/e2e/i18n/` 下新增或扩展用例：切换语言后菜单标题响应式更新且不触发 `/api/v1/user/info` 重新拉取（按 `lina-e2e` TC ID 约定继续编号）
- [ ] 9.7 在 `apps/lina-core` 运行 `go test ./...`，确认服务层单测全部通过
- [ ] 9.8 在 `hack/tests` 运行 `pnpm test`，确认 E2E 全部通过

## 10. 审查与归档准备

- [ ] 10.1 调用 `/lina-review` 执行变更全面审查，覆盖代码、SQL、E2E 与规范遵循
- [ ] 10.2 根据审查结果在 `tasks.md` 内追加修复任务并完成；如有规范偏差同步更新 `specs/**/spec.md`
- [ ] 10.3 在归档前再次运行 `openspec validate code-quality-improvements --strict` 与 `make test`，确认无回归
