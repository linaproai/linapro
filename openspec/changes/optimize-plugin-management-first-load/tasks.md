## 1. 后端接口契约

- [x] 1.1 在 `apps/lina-core/api/plugin/v1/plugin_list.go` 为 `GET /plugins`增加 `pageNum`、`pageSize`请求字段，定义默认值和服务端最大上限，并保持 `plugin:query`只读权限语义。
- [x] 1.2 拆分列表摘要 DTO 与详情 DTO：列表项移除 `dependencyCheck`、`requestedHostServices`、`authorizedHostServices`、`declaredRoutes`和 cron 详情字段，`GET /plugins/{id}`继续返回完整治理详情。
- [x] 1.3 更新接口文档 `dc`、`eg`、时间字段说明和前端契约注释；如 API 文档源文本发生变化，同步宿主 `apidoc` i18n 资源。
- [x] 1.4 执行 `cd apps/lina-core && make ctrl`或确认本次 API 变更不需要生成控制器骨架，并在任务记录中说明判断依据。
  - 记录：本次只修改既有 `GET /plugins`请求/响应 DTO 和现有控制器映射，不新增 controller 方法、路由绑定文件或脚手架骨架；已通过 Go 编译门禁覆盖。

## 2. 后端读模型与缓存

- [x] 2.1 在 `apps/lina-core/internal/service/plugin/`实现插件管理摘要列表路径，先构建轻量 manifest/registry snapshot，再完成过滤、排序、分页和当前页摘要投影。
- [x] 2.2 确保摘要列表路径不调用 `CheckPluginDependencies`、`ListCronDeclarationsByPlugin`、详情投影或完整 host service/route review 转换。
- [x] 2.3 保留或调整详情路径，使 `GET /plugins/{id}`只为目标插件装配依赖检查、host service 授权、动态路由和必要 cron 审查数据。
- [x] 2.4 调整 `buildManagedCronJobMap`使用边界：列表路径不得调用；若仍存在批量治理审查场景，新增一次扫描后按插件 ID 分组的批量契约。
- [x] 2.5 将摘要列表读模型和详情读模型缓存键绑定 locale、runtime bundle version 和 `plugin-runtime` revision；失效继续复用单机本地 revision 或集群 Redis revision/event。
- [x] 2.6 调整 `PrewarmManagementList`为轻量摘要预热或新增明确的摘要预热入口；预热失败不得影响请求正确性，冷启动并发必须通过 singleflight 或等价方式避免重复构建。
- [x] 2.7 记录 DI 来源检查：如新增或修改缓存敏感服务依赖，说明 owner、创建位置、传递路径、共享实例策略；若无新增运行期依赖，明确记录无影响。
  - 记录：未新增运行期依赖或缓存敏感服务实例；复用 `serviceImpl.managementListCache`、既有 `runtimeCacheRevisionCtrl`、`plugin-runtime` revision/event 和启动期注入的共享服务图。

## 3. 后端测试

- [x] 3.1 为列表摘要路径补充单元测试，覆盖默认分页、最大 `pageSize`上限、筛选总数、当前页字段边界和空结果。
- [x] 3.2 补充性能防回归测试或替身断言，证明 `GET /plugins`列表路径不执行依赖检查、不逐插件加载 cron 声明、不为每个插件重复 `ScanManifests()`。
- [x] 3.3 补充详情路径测试，证明详情仍返回依赖检查、host service、动态路由和 cron 审查所需字段，且只装配目标插件。
- [x] 3.4 补充缓存失效测试，覆盖插件安装、启用、禁用、卸载、升级或源码同步后摘要列表与详情缓存失效；记录单机和集群模式策略覆盖。
- [x] 3.5 运行 `cd apps/lina-core && go test ./api/plugin/v1 ./internal/controller/plugin ./internal/service/plugin ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/cmd/internal/httpstartup -count=1`。

## 4. 前端实现

- [x] 4.1 在 `apps/lina-vben/apps/web-antd/src/api/system/plugin/`拆分列表摘要类型和详情类型，新增或调整 `pluginDetail(id)`调用 `GET /plugins/{id}`。
- [x] 4.2 更新 `apps/lina-vben/apps/web-antd/src/views/system/plugin/index.vue`，首屏 grid 仅依赖摘要列表字段，不读取 detail-only 字段。
- [x] 4.3 将详情、动态上传、安装授权、卸载、升级和生命周期前置条件弹窗改为按需异步加载，保持现有 `useVbenModal`交互、权限控制和错误处理。
- [x] 4.4 调整详情、安装授权、卸载和升级工作流，打开弹窗或执行动作前按需请求详情、依赖检查或升级预览；首屏不得为每行自动请求详情。
- [x] 4.5 如新增或修改用户可见文案、表格列、按钮、提示或错误文案，同步宿主前端语言包并运行 `cd apps/lina-vben/apps/web-antd && pnpm i18n:check`。
  - 记录：未新增或修改前端运行时用户可见文案；本次 `i18n`影响为宿主 API 文档源文本变化，已同步 `apps/lina-core/manifest/i18n/zh-CN/apidoc/core-api-plugin.json`。
- [x] 4.6 运行 `cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`。

## 5. E2E 与用户可观察验证

- [x] 5.1 按 `lina-e2e`规范更新宿主插件能力用例 `hack/tests/e2e/extension/plugin/TC014-plugin-management-first-load.ts`；若覆盖无法复用该文件，再新增连续编号 `TC015-*.ts`。
- [x] 5.2 E2E 断言首屏进入插件管理页面时只请求 `GET /plugins`摘要列表，不自动请求 `GET /plugins/{id}`，且页面表格正常渲染插件类型、状态、安装时间和操作列。
- [x] 5.3 E2E 断言打开详情或安装授权弹窗后才请求详情或对应动作接口，并能展示依赖检查、host service 和动态路由审查信息。
- [x] 5.4 E2E 按测试规则捕获首屏和弹窗截图，检查无原始 i18n key、无文本重叠、无异常 toast 或空白状态。
- [x] 5.5 运行 `cd hack/tests && pnpm test:validate && pnpm exec playwright test e2e/extension/plugin/TC014-plugin-management-first-load.ts`。

## 6. 治理验证与审查

- [x] 6.1 运行 `openspec validate optimize-plugin-management-first-load --strict`。
- [x] 6.2 确认本次变更无插件目录结构修改；若后续触碰 `apps/lina-plugins/<plugin-id>/`，先读取该插件根目录 `AGENTS.md`。
- [x] 6.3 记录影响分析：缓存一致性已复用 `plugin-runtime`协调；数据权限为平台治理控制面且首屏减少敏感细节暴露；数据库默认无 SQL 变更；开发工具跨平台无影响；`i18n`按实际文案/API 文档变化记录。
  - 记录：未修改 `apps/lina-plugins/<plugin-id>/`；缓存复用 `plugin-runtime` revision/event 和本地失效；平台治理控制面继续由 `plugin:query`及治理动作权限保护，首屏不再暴露依赖、host service、路由等敏感细节；无 SQL/DAO/索引变更；未修改跨平台开发工具入口；`i18n`仅涉及宿主 API 文档翻译资源更新。
- [x] 6.4 完成实现和验证后调用 `lina-review`进行代码、OpenSpec、E2E 质量和规则合规审查。
  - 记录：`lina-review`未发现阻塞问题；已确认任务状态与实现、验证命令和影响分析一致。
