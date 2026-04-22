## 1. 菜单骨架与宿主目录治理

- [x] 1.1 设计并确认新的默认后台一级目录结构与稳定父级 `menu_key`
- [x] 1.2 更新宿主菜单初始化 SQL，创建 `dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer` 等稳定目录
- [x] 1.3 调整宿主菜单投影逻辑，仅消费宿主稳定目录记录并支持新的一级目录结构
- [x] 1.4 实现空父目录自动隐藏规则
- [x] 1.5 补充菜单重构后的后端测试与前端路由装配验证

## 2. 插件清单与挂载规则治理

- [x] 2.1 固化官方源码插件 ID：`org-center`、`content-notice`、`monitor-online`、`monitor-server`、`monitor-operlog`、`monitor-loginlog`
- [x] 2.2 为源码插件菜单挂载规则补充校验，限制其只能挂到宿主稳定目录或插件内部节点
- [x] 2.3 在插件文档和样例中移除 `plugin-` 前缀示例，统一改为领域-能力命名
- [x] 2.4 为插件治理菜单同步与父级挂载补充单元测试

## 3. 宿主核心边界抽离

- [x] 3.1 盘点并标注宿主保留能力与待插件化能力的代码边界
- [x] 3.2 为宿主定义并实现统一登录事件契约与发布器
- [x] 3.3 将登录成功、登录失败、登出成功接入统一登录事件，移除认证链路对具体登录日志落库实现的直接依赖
- [x] 3.4 为宿主定义并实现统一审计事件契约与发布器
- [x] 3.5 将写操作和带 `operLog` 标签的查询接入统一审计事件，移除中间件对具体操作日志落库实现的直接依赖
- [x] 3.6 抽离组织能力接口，避免用户管理直接依赖部门/岗位实现
- [x] 3.7 拆分“认证会话内核”与“在线用户治理能力”的边界，明确宿主负责会话真相源、活跃时间维护、超时判定与清理

## 4. 监控能力源码插件化

- [x] 4.1 创建源码插件 `monitor-operlog`
- [x] 4.2 迁移操作日志查询、详情、导出、清理与页面到 `monitor-operlog`
- [x] 4.3 创建源码插件 `monitor-loginlog`
- [x] 4.4 迁移登录日志查询、详情、导出、清理与页面到 `monitor-loginlog`
- [x] 4.5 创建源码插件 `monitor-server`
- [x] 4.6 迁移服务监控采集、清理、存储、查询与页面到 `monitor-server`
- [x] 4.7 创建源码插件 `monitor-online`
- [x] 4.8 在保留宿主会话内核前提下，迁移在线用户查询与强制下线治理到 `monitor-online`
- [x] 4.9 为 4 个监控插件分别补充安装、启用、停用、卸载和菜单挂载验证

## 5. 组织与内容能力源码插件化

- [x] 5.1 创建源码插件 `org-center`
- [x] 5.2 迁移部门管理到 `org-center`
- [x] 5.3 迁移岗位管理到 `org-center`
- [x] 5.4 实现用户管理在组织插件缺失时的 UI 与接口降级
- [x] 5.5 创建源码插件 `content-notice`
- [x] 5.6 迁移通知公告能力到 `content-notice`

## 6. 插件显式接线与交付文档

- [x] 6.1 更新 `apps/lina-plugins/lina-plugins.go`，为官方源码插件提供显式接线入口
- [x] 6.2 为每个官方源码插件补充 `README.md` 与 `README.zh_CN.md`
- [x] 6.3 更新 `apps/lina-plugins/README.md` 与 `README.zh_CN.md`，说明宿主/插件边界与官方插件列表
- [x] 6.4 补充插件安装、启停、菜单挂载和卸载清理的运维说明

## 7. 前端路由与菜单可见性联动

- [x] 7.1 调整前端静态路由与动态菜单适配，匹配新的一级目录结构
- [x] 7.2 实现插件启停后的菜单与动态路由刷新收敛逻辑回归验证
- [x] 7.3 实现组织插件缺失时用户管理页面字段隐藏
- [x] 7.4 实现监控插件缺失时系统监控空目录隐藏
- [x] 7.5 实现组织插件缺失时用户列表、详情和编辑抽屉的部门/岗位字段按能力探测降级

## 8. E2E 测试

- [x] 8.1 使用 `openspec-e2e` 规划菜单重构与插件化后的 E2E 用例
- [x] 8.2 新增菜单骨架与空父目录隐藏测试
- [x] 8.3 新增 `monitor-operlog` 生命周期与功能验证测试（`TC0098a` + 既有监控功能用例）
- [x] 8.4 新增 `monitor-loginlog` 生命周期与功能验证测试（`TC0098b` + 既有监控功能用例）
- [x] 8.5 新增 `monitor-server` 生命周期与功能验证测试（`TC0098c` + 既有监控功能用例）
- [x] 8.6 新增 `monitor-online` 生命周期与功能验证测试（`TC0098d` + 既有监控功能用例）
- [x] 8.7 新增 `org-center` 生命周期与用户管理降级测试（`TC0098e`、`TC0081`）
- [x] 8.8 新增 `content-notice` 生命周期测试（`TC0098f`、`TC0037`）
- [x] 8.9 新增 `monitor-online` 缺失或停用时登录、鉴权和会话超时仍正常的回归测试（`TC0099a`）
- [x] 8.10 新增日志插件缺失或停用时登录流程与普通业务请求仍正常的回归测试（`TC0099b`）
- [x] 8.11 运行相关 E2E 回归并记录结果

## 9. 验证与审查

- [x] 9.1 运行宿主与插件相关后端单元测试
- [x] 9.2 运行前端类型检查与构建验证
- [x] 9.3 运行插件启停与菜单刷新相关 E2E 套件
- [x] 9.4 对照规范审查宿主是否仍只保留核心能力
- [x] 9.5 调用 `openspec-review` 进行变更审查

## Feedback

- [x] **FB-1**: 宿主与插件协作改为稳定能力接缝，移除散落的插件占位判断与高耦合分支
- [x] **FB-2**: 6 个官方源码插件必须完整落地到 `apps/lina-plugins/<plugin-id>/`，并对齐 `plugin-demo-source` 目录结构与显式接线方式
- [x] **FB-3**: 删除未被任何插件或宿主代码使用的 `pkg` 桥接模块，避免暴露无效公共接口
- [x] **FB-4**: 将宿主私有菜单挂载键与官方插件治理常量从 `apps/lina-core/pkg/` 回收到 `internal/`，避免 `pkg` 承载宿主治理规则
- [x] **FB-5**: `orgcap` 能力接缝未密封——`internal/service/orgcap/orgcap.go` 直接 `import deptsvc` 并对外暴露 `[]*deptsvc.TreeNode`，`controller/user/user_v1_dept_tree.go` 因此被迫依赖 `deptsvc`；需让 `orgcap` 拥有自有 `DeptTreeNode` 并让宿主业务层彻底不再 `import deptsvc`
- [x] **FB-6**: 清理残留工件——删除空目录 `apps/lina-core/pkg/officialplugin/` 与仅余注释的 SQL 残桩 `apps/lina-core/manifest/sql/003-oper-login-log.sql`、`004-notice-message.sql`
- [x] **FB-7**: 插件前端真实化——6 个插件 `frontend/pages/*.vue` 目前全是 `import HostPage from '#/views/...'` 的薄包装；需把 `apps/lina-vben/apps/web-antd/src/views/{monitor,system}/...` 与 `src/api/{monitor,system}/...` 对应代码物理迁移到各插件 `frontend/` 下，同时在宿主新增 `/user/post-options` 能力端点（复用既有 `/user/dept-tree`）让 `user-drawer` 不再依赖插件前端 API
  - 2026-04-21：宿主已补充 `/user/post-options`，`user-drawer` 已切换为依赖 `#/api/system/user`；`org-center`、`content-notice`、`monitor-online`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 的页面、弹窗与客户端 API 已物理迁入各自 `frontend/pages/`，不再通过 `HostPage` 薄包装依赖宿主视图与宿主前端 API 模块
- [x] **FB-8**: 插件后端真正归属——`dept/post/notice/loginlog/operlog/servermon` 的 `api/`、`internal/controller/`、`internal/service/` 仍在 `apps/lina-core/` 并经 `pkg/plugincontroller/`、`pkg/pluginservice/` 桥接暴露；需参照 `apps/lina-plugins/plugin-demo-source/backend/` 迁入各插件 `backend/api` 与 `backend/internal/{controller,service}`，并删除对应宿主 `internal/` 目录、桥接包与 `dao/sys_{dept,post,user_dept,user_post,notice,login_log,oper_log,server_monitor}.go`
  - 2026-04-21：`org-center` 的 `dept/post`、`content-notice` 的 `notice`、`monitor-online`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 已全部迁入各插件 `backend/api` 与 `backend/internal/{controller,service}`；宿主旧 `api/`、`internal/controller/`、`internal/service/`、`pkg/plugincontroller/`、`pkg/pluginservice/{loginlog,operlog,servermon}` 与对应 `dao/model` 工件已删除，数据库访问模式在后续反馈中继续收敛到插件本地 `gf gen dao` 生成的 `dao/do/entity`
- [x] **FB-9**: 明确 `orgcap` 的宿主边界例外——组织能力接缝继续由宿主消费插件安装到宿主库中的 `sys_dept/sys_post/sys_user_dept/sys_user_post` 表作为只读/关联真相源，但该例外必须在 OpenSpec 文档与代码注释中显式声明，避免 `FB-8` 目标与实际实现语义不一致
- [x] **FB-10**: 稳定 `pkg/pluginservice/session` 对外契约——移除对 `internal/service/session` 类型的直接别名发布，改为宿主自有的独立 `Session/ListFilter/ListResult` DTO，并让 `monitor-online` 仅依赖该稳定契约
- [x] **FB-11**: 封死用户页对组织插件前端 API 的默认回退——`user/dept-tree` 与 `user/post-options` 已是宿主能力端点，`system/user/dept-tree.vue` 需默认依赖宿主用户 API/类型而非 `#/api/system/dept`，避免后续复用时重新引入插件前端耦合
- [x] **FB-12**: 系统信息页 E2E 需跟随后端动态元数据校验当前组件描述，避免继续断言已变更的静态 GoFrame 文案
- [x] **FB-13**: `TC0021c` 依赖固定 `testuser` 导致完整回归出现 skipped，需改为测试内自建用户并清理，保证用例独立可重复运行
  - 2026-04-21：`hack/tests/e2e/system/TC0021-user-dept-tree-count.ts` 已改为测试内创建唯一用户名用户、通过 API 完成部门变更并在 `finally` 中清理，去除对固定 `testuser` 的前置依赖
- [x] **FB-14**: 官方源码插件与 `plugin-demo-source` 的后端数据库访问需统一切换到插件本地 `gf gen dao` 生成的 `dao/do/entity`，并为每个源码插件 `backend/` 补齐 `hack/config.yaml`
  - 2026-04-21：已为 `content-notice`、`monitor-loginlog`、`monitor-online`、`monitor-operlog`、`monitor-server`、`org-center`、`plugin-demo-source` 的 `backend/` 补齐 `hack/config.yaml`；并为当前存在数据库访问的源码插件生成本地 `internal/dao`、`internal/model/do`、`internal/model/entity` 工件，相关 service 已改造为通过插件本地 `dao/do/entity` 访问数据库
- [x] **FB-15**: `middleware_request_body_limit` 在手写统一错误响应时仍使用裸 `g.Map`，需改为 `ghttp.DefaultHandlerResponse` 或宿主自有 typed DTO，保持与统一返回结构一致
  - 2026-04-21：`apps/lina-core/internal/service/middleware/middleware_request_body_limit.go` 已统一改为输出 `ghttp.DefaultHandlerResponse`，保持超限上传错误响应与宿主默认 JSON 结构一致
- [x] **FB-16**: 宿主默认数据库装载边界仍泄漏到插件——移除 `apps/lina-core/manifest/sql/mock-data/` 中对组织、公告等源码插件业务表的 DML，并补齐插件自有演示数据装载方案
  - 2026-04-21：宿主已删除 `001-mock-depts.sql`、`002-mock-posts.sql`、`004-mock-associations.sql`、`005-mock-notices.sql`，并将组织/公告演示数据收敛到各插件 `manifest/sql/` 生命周期资源中随插件安装装载
- [x] **FB-17**: 组织能力仍保留宿主持有插件表的存储例外——删除宿主对组织插件业务表的 `dao/do/entity` 与直接查表逻辑（含 `orgcap`、插件资源范围等路径），改为由 `org-center` 通过稳定 capability provider 提供实现，宿主仅保留接口、DTO 与空实现
  - 2026-04-21：宿主已删除 `sys_dept`、`sys_post`、`sys_user_dept`、`sys_user_post` 的 `dao/do/entity` 与直接查表实现；`orgcap`、插件资源范围与用户管理路径统一改为通过 `pkg/orgcap` provider 获取组织能力，`FB-9` 中记录的宿主直查组织表例外已被移除
- [x] **FB-18**: 官方源码插件业务表仍沿用 `sys_*` 命名——将 `org-center`、`content-notice`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 的插件自有表统一迁移到 `plugin_<plugin_id_snake_case>_` 作用域前缀，并同步更新安装/卸载 SQL、插件本地 `gf gen dao` 配置、后端实现、测试与文档
  - 2026-04-21：`org-center`、`content-notice`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 的业务表已统一迁移为 `plugin_*` 前缀，插件安装/卸载 SQL、`backend/hack/config.yaml`、本地 `gf gen dao` 工件、后端实现与变更文档已同步更新
- [x] **FB-19**: 官方组织源码插件命名从 `org-management` 调整为 `org-center`，并同步更新目录名、Go 模块名、插件治理常量、权限键、测试与文档引用
- [x] **FB-20**: 单表官方源码插件的业务表名不应重复资源后缀——将 `content-notice`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 的物理表分别收敛为 `plugin_content_notice`、`plugin_monitor_loginlog`、`plugin_monitor_operlog`、`plugin_monitor_server`，并同步调整 `backend/hack/config.yaml` 的 `removePrefix` 避免 codegen 留下空表名或不必要的重复后缀
  - 2026-04-21：已将 `content-notice`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` 的安装/卸载 SQL、本地数据库表名、插件 `backend/hack/config.yaml` 与 `gf gen dao` 工件统一收敛到单表命名；其中 `content-notice` 使用 `removePrefix: "plugin_content_"` 生成 `Notice`，3 个监控插件使用 `removePrefix: "plugin_monitor_"`，对应生成 `Loginlog`、`Operlog`、`Server`
- [x] **FB-21**: `openspec/specs/` 主规范仍残留旧 `sys_*` 插件表名和旧宿主边界表述——同步更新用户、组织、公告、日志、服务监控和通知域主规范，改为插件化后的最终表名与能力边界描述
  - 2026-04-21：已更新 `openspec/specs/{user-management,dept-management,post-management,notice-management,login-log,oper-log,server-monitor,plugin-notify-service}/spec.md`，将旧 `sys_*` 插件表名、宿主直持有插件表描述与旧菜单边界表述统一收敛为当前插件化后的最终规范
- [x] **FB-22**: 合并误创建的活跃迭代 `host-plugin-boundary-followup` 回当前 `host-plugin-boundary-modularization`，恢复单一活跃变更治理状态
  - 2026-04-21：已删除误创建且未承载有效文档的 `openspec/changes/host-plugin-boundary-followup/` 目录，并继续将后续反馈统一追加到当前未归档的 `host-plugin-boundary-modularization`
- [x] **FB-23**: 将宿主业务组件中用于“可选依赖”的 variadic 构造函数改为显式参数签名；未显式注入时统一传 `nil`，并在构造函数注释中说明默认行为
  - 2026-04-21：`auth.New`、`user.New`、`menu.New`、`orgcap.New`、`plugin.New` 与 `controller/plugin.NewV1` 已统一改为显式参数签名；宿主调用点与相关测试统一改为按需传入依赖或显式传 `nil`，并补充了默认行为注释
- [x] **FB-24**: 将 `role` 对 `plugin` 的权限菜单过滤依赖改为窄接口注入，并在循环依赖解除后将菜单治理元数据从 `menu/metadata` 回收到 `menu` 根包内聚维护
  - 2026-04-21：`role` 已改为依赖窄接口 `PermissionMenuFilter`，移除对 `plugin.Service` 的直接依赖；`menu` 稳定目录与官方插件挂载元数据已回收到 `internal/service/menu/menu_metadata.go`，`plugin/internal/{catalog,integration}`、`orgcap` 与相关测试已同步切换
- [x] **FB-25**: 宿主控制器构造函数未同步显式依赖注入签名，导致 `go test ./...` 与后端构建在 `auth`、`role`、`joblog` 控制器处直接编译失败
  - 2026-04-22：`apps/lina-core/internal/controller/{auth,role,joblog}/*_new.go` 已同步改为显式构造 `pluginSvc` / `orgCapSvc` 并传入 `auth.New(...)`、`role.New(...)`；随后 `apps/lina-core/go test ./...`、插件/构建模块 `go test ./...` 与前端 `pnpm test:unit` 全部通过
- [x] **FB-26**: `TC0069` 仍使用旧组织关联表 `sys_user_dept/sys_user_post` 做清理，导致全量 E2E 在动态插件权限治理用例收尾阶段失败
  - 2026-04-22：`hack/tests/e2e/plugin/TC0069-plugin-permission-governance.ts` 已改为清理 `plugin_org_center_user_dept`、`plugin_org_center_user_post`；单测 `TC0069` 单独回归通过，随后 `npx playwright test` 全量 341 条全部通过
- [x] **FB-27**: 操作日志需改为源码插件通过宿主封装的全局 HTTP middleware 注册器自注册审计链路，插件停用时完全旁路采集逻辑并移除宿主专用 `OperLog` 业务中间件
  - 2026-04-22：宿主已发布统一 `HTTPRegistrar`（路由注册器 + 全局 middleware 注册器），`monitor-operlog` 改为自注册 `/api/v1/*` 审计 middleware 并通过稳定 `pkg/pluginservice/audit` 接缝发射审计事件；宿主静态/动态路由链路已移除专用 `OperLog` 中间件，插件启停运行时通过集成层内存快照即时旁路对应全局 middleware / 路由守卫 / cron 守卫；已回归 `apps/lina-core` 与受影响源码插件后端 `go test ./...`，并通过 Playwright 用例 `TC0026`、`TC0098`、`TC0099`
- [x] **FB-28**: `pkg/pluginbridge` / `pkg/pluginhost` 暴露给插件的行为型对象需统一接口化，`NewSourcePlugin` 不再返回具体 struct，并将源码插件能力注册改为分组接口对象
  - 2026-04-22：`pluginhost.NewSourcePlugin` 已改为返回 `SourcePlugin` 接口，并新增 `Assets/Lifecycle/Hooks/HTTP/Cron/Jobs/Governance` 分组注册 facade；宿主注册表与 manifest 已切换为 `SourcePluginDefinition` 读取视图，`pluginbridge.NewGuestRuntime` 与 `NewGuestControllerRouteDispatcher` 也已改为返回接口；相关源码插件接线与单元测试已同步更新，并完成 `apps/lina-core` 相关包、`plugin-demo-dynamic` 与受影响源码插件 `go test ./...` 回归
- [x] **FB-29**: 删除 `plugin.jobs` 插件通用任务处理器能力，仅保留 `plugin.cron` 内置定时任务投影链路，并移除 `plugin-demo-source` 中的示例注册代码
  - 2026-04-22：已删除 `pkg/pluginhost` 对 `plugin.jobs` / `JobHandlerRegistration` 的公开契约与实现、移除 `jobhandler` 中对通用插件任务处理器的注册逻辑、删除 `pluginbridge.BuildPluginHandlerRef` 与 `plugin-demo-source` 示例注册代码；并回归 `cd apps/lina-core && go test ./pkg/pluginhost ./pkg/pluginbridge ./internal/service/jobhandler ./internal/service/plugin/... ./internal/service/jobmgmt/...`、`cd apps/lina-plugins/plugin-demo-source/backend && go test ./...`、`cd hack/tests && npx playwright test e2e/system/job/TC0090-job-plugin-cascade.ts`
- [x] **FB-30**: 删除 `http.request.after-auth` / `RegisterAfterAuthHandler` 插件扩展点，移除宿主分发链路与源码插件残留调用
  - 2026-04-22：已从 `openspec/specs/plugin-hook-slot-extension/spec.md` 与当前变更增量规范中移除 `http.request.after-auth`；宿主 `pkg/pluginhost` 已删除 `RegisterAfterAuthHandler`、`AfterAuthInput`、`ExtensionPointHTTPRequestAfterAuth` 与对应注册存储；`middleware`、`plugin/internal/{integration,runtime}`、`plugin` facade 与 `monitor-operlog` 源码插件中的 after-auth 分发/调用链已同步清理。该反馈不涉及用户可观察行为变化，未新增 E2E；已回归 `cd apps/lina-core && go test ./pkg/pluginhost ./internal/service/middleware ./internal/service/plugin/...`，以及各源码插件 backend 的 `go test ./...`
- [x] **FB-31**: `monitor-operlog` 审计中间件需参照 `lina-core/internal/service/middleware` 收敛到插件 `service` 层，避免 `backend/audit_middleware.go` 在根目录承载厚编排逻辑
  - 2026-04-22：已新增 `apps/lina-plugins/monitor-operlog/backend/internal/service/middleware/` 组件，按宿主 `middleware.Service` 模式承载审计 middleware 编排、元数据归一化、敏感字段脱敏与异步审计事件分发；`backend/plugin.go` 改为注册 `middlewaresvc.New().Audit`，并删除根目录 `audit_middleware.go` 与对应测试文件。该反馈属于内部架构收敛，未引入用户可观察行为变更，因此补充并运行了插件后端 `go test ./...` 与新的 middleware 单元测试，未新增 E2E。
- [x] **FB-32**: 源码插件 `service` 目录层级不规范——统一迁移到 `backend/internal/service/`，禁止继续直接放在 `backend/service/`
  - 2026-04-22：已将 `content-notice`、`monitor-loginlog`、`monitor-online`、`monitor-operlog`、`monitor-server`、`org-center` 与 `plugin-demo-source` 的 `service` 组件统一迁移到 `backend/internal/service/`，并同步修复控制器、provider、插件注册入口与测试中的全部导入路径；随后完成上述插件模块与 `apps/lina-plugins` 聚合模块的 `go test ./...` 回归
- [x] **FB-33**: 将源码插件目录结构补充为项目规范，显式约束 `backend/internal/service/`、`backend/internal/controller/`、`backend/hack/config.yaml` 与插件资源目录
  - 2026-04-22：已更新 `CLAUDE.md`（即根 `AGENTS.md`）、`apps/lina-plugins/README.md`、`apps/lina-plugins/README.zh_CN.md`、`apps/lina-plugins/plugin-demo-source/README.md`、`apps/lina-plugins/plugin-demo-source/README.zh_CN.md`，并在当前活跃变更的 `plugin-manifest-lifecycle` 增量规范中补充源码插件标准目录约束
- [x] **FB-34**: 系统菜单类型在宿主代码中仍散落使用 `"D"` / `"M"` / `"B"` 字符串，需收敛为强类型常量并统一复用
  - 2026-04-22：已新增 `apps/lina-core/pkg/menutype` 定义宿主/插件共享的菜单类型强类型常量，并将宿主菜单校验、菜单/用户控制器及相关插件单测统一切换为常量复用；`go build ./pkg/menutype ./internal/service/menu ./internal/controller/menu ./internal/controller/user ./internal/service/plugin/...` 编译通过
- [x] **FB-35**: `monitor-operlog` 仍使用 `1~6` 整数表达操作类型，需改为 `create/update/delete/export/import/other` 语义字符串并同步宿主事件、插件接口、字典与前端契约
  - 2026-04-22：已新增 `apps/lina-core/pkg/audittype` 统一宿主审计事件、动态路由 `operLog` 校验与插件落库类型；`monitor-operlog` SQL、DAO/DO/Entity、API DTO、控制器、服务、前端客户端与字典值均已切换为语义字符串类型，并已更新本地 MySQL `plugin_monitor_operlog` 表结构后执行 `gf gen dao`。验证通过：`cd apps/lina-core && go test ./pkg/menutype ./pkg/audittype ./pkg/pluginbridge ./pkg/pluginhost`、`cd apps/lina-core && go build ./pkg/menutype ./pkg/audittype ./pkg/pluginbridge ./pkg/pluginhost ./pkg/pluginservice/audit ./internal/service/menu ./internal/controller/menu ./internal/controller/user ./internal/service/plugin/...`、`cd apps/lina-plugins/monitor-operlog/backend && go test ./api/operlog/... ./internal/controller/operlog ./internal/service/operlog`、`cd apps/lina-plugins/monitor-operlog/backend && go build ./...`
- [ ] **FB-36**: 宿主工作台仍残留未引用的旧版 `monitor-operlog` 前端页面与 API 副本，需删除重复实现并回归操作日志相关 E2E，避免后续误接回旧契约
  - 2026-04-22：已删除 `apps/lina-vben/apps/web-antd/src/{api/monitor/operlog,views/monitor/operlog}/` 旧宿主页副本，当前仅保留 `apps/lina-plugins/monitor-operlog/frontend/pages/` 作为前端事实来源；`rg` 已确认宿主源码中无残留引用。已尝试运行受影响 Playwright 用例 `TC0026`、`TC0027`、`TC0028`、`TC0029`、`TC0035` 与 `TC0098`，但当前环境缺少可用的 Playwright 浏览器产物，`chromium`/`webkit` 项目均在浏览器安装阶段失败；作为替代，已通过 `playwright-cli` 真浏览器冒烟验证 `admin/admin123` 登录后可正常打开 `/monitor/operlog`，列表、筛选区、工具栏与详情抽屉均可渲染。该项待具备可用 Playwright 浏览器环境后补跑正式 E2E 再勾选完成
