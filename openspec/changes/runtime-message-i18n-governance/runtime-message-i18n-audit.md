# 运行时文案中英文混合盘点

## 扫描范围

本次盘点聚焦会进入运行时响应、导入导出文件、插件桥接协议、前端页面展示或运维/审计输出的源代码文案。

当前扫描命令基线：

```bash
make check-runtime-i18n
make check-runtime-i18n-messages
go run ./hack/tools/runtime-i18n scan --format json
```

初始扫描结果：

- 后端宿主、公共包和源码插件中，排除测试、DAO、DO、Entity 后仍有 127 个 Go 生产文件包含中文字符。
- 其中 117 个 Go 生产文件命中错误、返回消息、失败原因、导出 fallback、状态文本或插件桥接错误等高风险模式。
- 前端主应用和插件前端仍有中文字符，但大量是注释、已通过 `$t` 使用的文案或 i18n 资源之外的类型说明；运行时高风险集中在监控页、在线用户页和请求错误透传。

当前工具基线：

- `go run ./hack/tools/runtime-i18n messages` 已通过，宿主和插件运行时语言包 key 覆盖没有缺失。
- `hack/tools/runtime-i18n/allowlist.json` 当前为空，扫描结果没有依赖 allowlist 豁免。
- `go run ./hack/tools/runtime-i18n scan` 当前仍报告 732 个高风险候选项，后续任务继续按模块清理。
- 最近一批已清理 41 个插件生命周期、自动启用、源码升级和动态产物生命周期守卫候选项；`plugin_auto_enable.go`、`plugin_lifecycle.go`、`plugin_lifecycle_source.go`、`internal/sourceupgrade/sourceupgrade.go`、`internal/lifecycle/lifecycle.go` 已不再命中扫描。
- 按规则统计仍以 `go-error-han` 为主，剩余候选集中在插件 catalog/spec、runtime/upload/reconciler、wasm host service、pluginbridge、plugindb、pluginfs 和源码插件后端。

## 问题分类

### 1. 后端 API 与业务错误直接返回中文

本轮清理前的代表性文件：

- `apps/lina-core/internal/service/dict/dict_data.go`
  - `DataGetById` 返回 `字典数据不存在`。
- `apps/lina-core/internal/service/dict/dict_type.go`
  - `Create`、`Update`、`GetById` 返回 `字典类型已存在`、`字典类型不存在`。
- `apps/lina-core/internal/service/user/user.go`
  - 用户不存在、用户状态、用户删除等业务错误仍存在中文返回。
- `apps/lina-core/internal/service/usermsg/usermsg.go`
  - `未登录` 等用户可见错误未结构化。
- `apps/lina-plugins/content-notice/backend/internal/service/notice/notice.go`
  - `通知公告不存在`、`请选择要删除的记录` 等插件业务错误由源码插件直接返回。
- `apps/lina-plugins/org-center/backend/internal/service/dept/dept.go`
  - `部门不存在`、`存在子部门，不允许删除`、`部门编码已存在`。
- `apps/lina-plugins/org-center/backend/internal/service/post/post.go`
  - `岗位不存在`、`请选择要删除的岗位`、`岗位编码已存在`。

影响：

- 前端请求拦截器会优先展示后端 `error` 或 `message` 字段，导致当前语言环境无法保证展示语言一致。
- 业务错误没有稳定错误码和参数，前端无法可靠二次本地化，也不利于测试断言。

### 2. 导入、导出和模板文件硬编码中文

代表性文件：

- `apps/lina-core/internal/service/user/user_excel.go`
  - 导出表头：`用户名`、`昵称`、`手机号码`、`邮箱`、`性别`、`状态`、`备注`、`创建时间`。
  - 枚举展示：`未知`、`男`、`女`、`正常`、`停用`。
  - 导入失败原因：`无法解析 Excel 文件`、`用户名和密码不能为空`、`密码加密失败`、`插入失败`。
  - 模板示例：`张三`、`男`、`正常`、`示例用户`。
- `apps/lina-core/internal/service/dict/dict_data.go`
  - 字典数据导出表头和状态文本：`字典标签`、`状态`、`正常`、`停用`。
- `apps/lina-core/internal/service/dict/dict_type.go`
  - 字典类型导出表头和状态文本：`字典名称`、`状态`、`正常`、`停用`。
- `apps/lina-core/internal/service/dict/dict_export.go`
  - 多 sheet 名称：`字典类型`、`字典数据`。
- `apps/lina-core/internal/service/dict/dict_import.go`
  - 大量导入失败原因：`数据不完整`、`字典名称不能为空`、`字典类型格式错误`、`查询失败`、`插入失败`。
- `apps/lina-core/internal/service/sysconfig/sysconfig_import.go`
  - 系统参数导入失败原因未进入 i18n 资源。
- `apps/lina-plugins/org-center/backend/internal/service/post/post.go`
  - 岗位导出状态 `正常`、`停用`。
- `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go`
  - 操作类型和状态映射包含 `导出`、`成功`、`失败`，导出字段 fallback 为中文。

影响：

- Excel 文件是用户可见交付物，当前会固定输出中文。
- 导入失败结果往往返回给页面或作为结果文件下载，必须按请求语言渲染。
- 逐行循环内直接拼接中文和底层错误会形成中英混排，例如 `数据库查询错误: <driver error>`。

### 3. 插件平台、桥接协议和宿主服务错误中英混排

代表性文件：

- `apps/lina-core/pkg/pluginbridge/pluginbridge_codec.go`
  - `动态路由 path 不能为空`、`动态路由 access 仅支持 public/login`、`动态路由 bridge runtimeKind 仅支持 wasm`。
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_data_codec.go`
  - `解析 data list request tag 失败`、`跳过未知 data list request 字段失败`。
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_notify_codec.go`
  - `解析 notify send request title 失败` 等协议解析错误。
- `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_network.go`
  - `network request URL 非法`、`network request URL scheme 不支持`、`network request header 不允许设置`。
- `apps/lina-core/pkg/pluginfs/pluginfs.go`
  - `SQL 资源路径不能为空`、`SQL 文件名必须使用 {序号}-{当前迭代名称}.sql`。
- `apps/lina-core/pkg/plugindb/host/db.go`
  - `plugin data service 暂不支持数据库类型`、`plugin data service SQL 未命中授权表`。
- `apps/lina-core/internal/service/plugin/internal/catalog/*.go`
  - 插件清单、菜单、资源、授权、动态产物校验错误大量包含中文和英文术语混排。

影响：

- 这些错误同时面向开发者、插件运行时和管理端用户；如果没有区分协议错误、开发者诊断和用户展示，会导致多语言策略不一致。
- 插件宿主服务错误需要保持机器可读状态码，不能只依赖本地化字符串。

### 4. 插件生命周期、升级和治理返回消息硬编码中文

代表性文件：

- `apps/lina-core/internal/service/plugin/internal/sourceupgrade/sourceupgrade.go`
  - `源码插件未安装，跳过升级。`
  - `当前源码插件已是最新版本，无需升级。`
  - `源码插件已从 %s 升级到 %s。`
  - 升级阻断提示多行中文说明。
- `apps/lina-core/internal/service/plugin/plugin_auto_enable.go`
  - 自动启用、跳过、失败等治理结果。
- `apps/lina-core/internal/service/plugin/plugin_lifecycle*.go`
  - 插件安装、卸载、启用、停用的业务错误和结果说明。

影响：

- 这类消息通常会展示在管理端、命令输出或升级诊断中，未来多语言环境下应根据操作者语言或命令 locale 输出。
- 结果对象应存储 `messageKey` 和参数，避免只存一段已经本地化的字符串。
- 当前已将上述核心文件改为结构化 `bizerr` 或 `messageKey/messageParams` 结果对象；剩余插件平台候选继续由任务 `4.1`、`4.2`、`4.4` 跟踪。

### 5. 前端页面仍有直接中文展示

代表性文件：

- `apps/lina-vben/apps/web-antd/src/views/monitor/online/data.ts`
  - 查询和表格列：`用户账号`、`IP地址`、`登录账号`、`部门名称`、`浏览器`、`操作系统`、`登录时间`、`操作`。
- `apps/lina-vben/apps/web-antd/src/views/monitor/server/index.vue`
  - 页面标题、指标名和空状态：`数据库信息`、`服务器信息`、`服务信息`、`服务 CPU`、`系统运行时长`、`暂无监控数据，请等待数据采集...`。
  - 时间格式：`天`、`小时`、`分钟`、`刚启动`。
- `apps/lina-vben/apps/web-antd/src/api/request.ts`
  - 错误拦截器直接展示 `responseData.error ?? responseData.message ?? msg`，后端硬编码文案会原样进入 UI。

影响：

- 前端主体已大量使用 `$t`，但这些残留页面会在切换语言后继续显示中文。
- 请求错误透传是后端硬编码问题的放大器，需要和后端结构化错误模型一起治理。

### 6. 日志、审计记录和运行结果边界不清

当前状态：

- 许多 `logger.Warningf` 已经是英文运维日志，例如 `session cleanup error`、`sync builtin scheduled jobs failed`。
- 但部分错误通过 `gerror.Wrap` 使用中文，随后可能被日志或响应复用。
- 操作日志、登录日志、任务日志、插件升级结果等数据同时承担审计、展示和导出职责，部分字段存的是中文展示值而不是稳定代码。

影响：

- 运维日志不应按用户语言本地化，否则会降低检索和聚合效率。
- 面向用户展示的审计/操作日志应存稳定代码和参数，展示或导出时按请求语言投影。

## 当前高风险候选文件清单

以下文件命中错误、返回消息、失败原因、导出 fallback、状态文本或插件桥接错误等模式，后续实施应逐批清理：

- `apps/lina-core/internal/cmd/cmd.go`
- `apps/lina-core/internal/cmd/cmd_init.go`
- `apps/lina-core/internal/cmd/cmd_mock.go`
- `apps/lina-core/internal/controller/config/config_v1_config_import.go`
- `apps/lina-core/internal/controller/dict/dict_v1_data_import.go`
- `apps/lina-core/internal/controller/dict/dict_v1_type_import.go`
- `apps/lina-core/internal/controller/jobhandler/jobhandler_v1_detail.go`
- `apps/lina-core/internal/controller/joblog/joblog_v1_cancel.go`
- `apps/lina-core/internal/controller/plugin/plugin_v1_resource_list.go`
- `apps/lina-core/internal/controller/user/user_v1_import.go`
- `apps/lina-core/internal/service/config/config_cron.go`
- `apps/lina-core/internal/service/config/config_duration.go`
- `apps/lina-core/internal/service/config/config_i18n.go`
- `apps/lina-core/internal/service/config/config_metadata.go`
- `apps/lina-core/internal/service/config/config_plugin.go`
- `apps/lina-core/internal/service/config/config_public_frontend.go`
- `apps/lina-core/internal/service/config/config_runtime_params.go`
- `apps/lina-core/internal/service/cron/cron_managed_jobs.go`
- `apps/lina-core/internal/service/dict/dict_data.go`
- `apps/lina-core/internal/service/dict/dict_export.go`
- `apps/lina-core/internal/service/dict/dict_import.go`
- `apps/lina-core/internal/service/dict/dict_type.go`
- `apps/lina-core/internal/service/file/file.go`
- `apps/lina-core/internal/service/file/file_storage_local.go`
- `apps/lina-core/internal/service/hostlock/hostlock.go`
- `apps/lina-core/internal/service/hostlock/hostlock_ticket.go`
- `apps/lina-core/internal/service/i18n/i18n_plugin_dynamic.go`
- `apps/lina-core/internal/service/jobhandler/jobhandler.go`
- `apps/lina-core/internal/service/jobhandler/jobhandler_host.go`
- `apps/lina-core/internal/service/jobhandler/jobhandler_plugin.go`
- `apps/lina-core/internal/service/jobhandler/jobhandler_schema.go`
- `apps/lina-core/internal/service/jobmeta/jobmeta.go`
- `apps/lina-core/internal/service/jobmgmt/internal/scheduler/scheduler.go`
- `apps/lina-core/internal/service/jobmgmt/internal/scheduler/scheduler_cancel.go`
- `apps/lina-core/internal/service/jobmgmt/internal/scheduler/scheduler_register.go`
- `apps/lina-core/internal/service/jobmgmt/internal/shellexec/shellexec.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_builtin.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_cron_validate.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_group.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_job_crud.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_job_status.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_log.go`
- `apps/lina-core/internal/service/jobmgmt/jobmgmt_registry.go`
- `apps/lina-core/internal/service/kvcache/kvcache_key.go`
- `apps/lina-core/internal/service/kvcache/kvcache_ops.go`
- `apps/lina-core/internal/service/menu/menu.go`
- `apps/lina-core/internal/service/menu/menu_validation.go`
- `apps/lina-core/internal/service/notify/notify_inbox.go`
- `apps/lina-core/internal/service/notify/notify_send.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/authorization.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/embedded.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/manifest.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/manifest_access.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/manifest_validate.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/manifest_validation.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/release.go`
- `apps/lina-core/internal/service/plugin/internal/catalog/spec.go`
- `apps/lina-core/internal/service/plugin/internal/datahost/datahost.go`
- `apps/lina-core/internal/service/plugin/internal/datahost/datahost_plan.go`
- `apps/lina-core/internal/service/plugin/internal/datahost/datahost_scope.go`
- `apps/lina-core/internal/service/plugin/internal/datahost/datahost_table.go`
- `apps/lina-core/internal/service/plugin/internal/frontend/bundle.go`
- `apps/lina-core/internal/service/plugin/internal/frontend/contract.go`
- `apps/lina-core/internal/service/plugin/internal/frontend/frontend.go`
- `apps/lina-core/internal/service/plugin/internal/integration/backend.go`
- `apps/lina-core/internal/service/plugin/internal/integration/extensions.go`
- `apps/lina-core/internal/service/plugin/internal/integration/extensions_cron_managed.go`
- `apps/lina-core/internal/service/plugin/internal/integration/menu.go`
- `apps/lina-core/internal/service/plugin/internal/lifecycle/migration.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/artifact.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/reconciler.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/release_artifact.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/route.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/runtime_cron.go`
- `apps/lina-core/internal/service/plugin/internal/runtime/upload.go`
- `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_network.go`
- `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_storage.go`
- `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_storage_cleanup.go`
- `apps/lina-core/internal/service/plugin/internal/wasm/wasm.go`
- `apps/lina-core/internal/service/role/role.go`
- `apps/lina-core/internal/service/sysconfig/sysconfig.go`
- `apps/lina-core/internal/service/sysconfig/sysconfig_import.go`
- `apps/lina-core/internal/service/user/user.go`
- `apps/lina-core/internal/service/user/user_excel.go`
- `apps/lina-core/internal/service/usermsg/usermsg.go`
- `apps/lina-core/internal/service/usermsg/usermsg_get.go`
- `apps/lina-core/pkg/excelutil/excelutil.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_cron_contract.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_guest_helpers.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_guest_hostcall.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostcall_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_cache_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_cron_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_data_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_lock_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_network_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_notify_codec.go`
- `apps/lina-core/pkg/pluginbridge/pluginbridge_hostservice_storage_codec.go`
- `apps/lina-core/pkg/plugindb/host/db.go`
- `apps/lina-core/pkg/pluginfs/pluginfs.go`
- `apps/lina-plugins/content-notice/backend/internal/service/notice/notice.go`
- `apps/lina-plugins/demo-control/backend/internal/service/middleware/middleware_guard.go`
- `apps/lina-plugins/monitor-loginlog/backend/internal/service/loginlog/loginlog.go`
- `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go`
- `apps/lina-plugins/org-center/backend/internal/service/dept/dept.go`
- `apps/lina-plugins/org-center/backend/internal/service/post/post.go`
- `apps/lina-plugins/plugin-demo-dynamic/backend/internal/service/dynamic/dynamic_demo_record.go`
- `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo/demo_record.go`
- `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo/demo_storage.go`

## 调用端可见接口错误 bizerr 审查补充

本轮按新增规范补充审查了已经落入当前实现面的用户、字典、配置导入控制器、统一响应/权限/请求体中间件和任务处理器/任务日志控制器。以下路径已修正为 `bizerr.NewCode` 或 `bizerr.WrapCode`，并补齐 `errorCode`、`messageKey`、`messageParams` 所需的运行时语言包：

- `apps/lina-core/internal/controller/user/user_v1_import.go`
- `apps/lina-core/internal/controller/dict/dict_v1_import.go`
- `apps/lina-core/internal/controller/dict/dict_v1_data_import.go`
- `apps/lina-core/internal/controller/dict/dict_v1_type_import.go`
- `apps/lina-core/internal/controller/config/config_v1_config_import.go`
- `apps/lina-core/internal/controller/jobhandler/jobhandler_v1_detail.go`
- `apps/lina-core/internal/controller/joblog/joblog_v1_cancel.go`
- `apps/lina-core/internal/service/middleware/middleware_response.go`
- `apps/lina-core/internal/service/middleware/middleware_permission.go`
- `apps/lina-core/internal/service/middleware/middleware_request_body_limit.go`

全量扫描仍能发现系统参数、文件、菜单、角色、通知、定时任务、插件 catalog/runtime/wasm、插件桥接、插件文件系统、插件数据库和源码插件后端中的直接 `gerror.New*`/`Wrap*` 返回路径。这些路径已由任务 `3.3` 至 `4.4` 继续跟踪，后续清理时必须按新规范统一改为所属模块 `*_code.go` 中的 `bizerr.Code` 定义。

### 任务 3.3 清理记录

已完成系统参数、配置导入、文件管理、菜单、角色、用户消息和通知模块中直接返回给调用端的用户可见错误治理：

- 为 `sysconfig`、`file`、`menu`、`role`、`usermsg`、`notify` 补齐模块级 `*_code.go` 错误定义。
- 将配置不存在、键名重复、内置参数保护、配置导入失败、文件上传/保存/删除、菜单移动/删除/唯一性、角色不存在/唯一性、用户消息未登录/不存在、通知通道/收件人/载荷等错误改为 `bizerr.NewCode` 或 `bizerr.WrapCode`。
- 配置导入行级失败原因改为运行时 i18n 文案或结构化错误本地化结果，避免继续向调用端返回中文自由文本。
- `role_access_cache.go` 中剩余两个英文 `gerror.New` 仅作为内部访问上下文缓存诊断，调用端路径会被权限中间件包装为 `PERMISSION_CONTEXT_LOAD_FAILED`。

### 任务 3.4 清理记录

已完成定时任务、任务处理器、任务元数据、任务日志取消路径、分布式 KV 缓存、插件宿主分布式锁和运行时参数校验中的调用端可见错误治理：

- 为 `jobmgmt`、`jobhandler`、`jobmeta`、`kvcache`、`hostlock`、`config`、`cron` 补齐或扩展模块级 `*_code.go` 错误定义。
- 将任务分组、任务 CRUD、状态切换、手动触发、Cron 表达式校验、任务处理器注册与 Schema/参数校验、保留策略解析、缓存键/值校验、锁票据校验、运行时参数值校验等错误改为 `bizerr.NewCode` 或 `bizerr.WrapCode`。
- 内部调度器和 Shell 执行器的可传播错误改为共享的 `jobmeta` 结构化错误定义，避免手动触发、取消和执行日志路径继续返回中文自由文本。
- 任务执行日志中由调度器写入的固定失败原因已改为英文源文本；后续若需要按请求语言展示历史执行日志，应继续推进“稳定错误码 + 参数化日志投影”的存储模型。
- 运行时语言包 `zh-CN`、`en-US`、`zh-TW` 及 packed manifest 副本已补齐本批新增错误键。
- `config_duration.go`、`config_metadata.go`、`config_plugin.go`、`config_i18n.go` 中剩余中文 `panic(gerror...)` 属于启动期配置诊断，不进入调用端响应；测试中的 `gerror.New`/`errors.New` 属于 fixture。

### 任务 3.5 阶段清理记录

已完成插件生命周期、启动期自动启用、源码插件升级和动态产物生命周期守卫的第一批治理：

- 为根插件 facade、动态生命周期子包、动态 runtime 子包和源码升级子包分别新增 `plugin_code.go`、`lifecycle_code.go`、`runtime_code.go`、`sourceupgrade_code.go`，所有调用端可见失败统一改为 `bizerr.NewCode` 或 `bizerr.WrapCode`。
- `SourcePluginUpgradeResult` 增加 `MessageKey` 与 `MessageParams`，源码升级的未安装跳过、已是最新版本和升级成功结果不再只返回固定中文 `Message`。
- 宿主运行时语言包新增 `error.plugin.*` 与 `plugin.sourceUpgrade.*` 三语资源，源码升级结果文案放在宿主 `manifest/i18n/<locale>/plugin.json`，错误文案放在 `error.json`。
- `make check-runtime-i18n-messages` 已通过，`go test ./internal/service/plugin ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin/internal/runtime` 已通过。
- 本批清理后`hack/tools/runtime-i18n`扫描总候选数从 814 降至 732；剩余插件 runtime/upload/reconciler、catalog/spec、frontend resource parser 和 wasm host service 仍需后续继续清理，因此任务`3.5`暂不标记完成。

### FB-8 工具迁移记录

运行时 i18n 检查已从临时`hack/scripts/check_runtime_i18n*.py`迁移为`hack/tools/runtime-i18n`下的 Go 工具：

- `make check-runtime-i18n`调用`go run ./hack/tools/runtime-i18n scan`。
- `make check-runtime-i18n-messages`调用`go run ./hack/tools/runtime-i18n messages`。
- allowlist 长期维护在`hack/tools/runtime-i18n/allowlist.json`。
- 迁移后 Go 扫描结果与迁移前 Python 扫描结果一致，当前均为 732 个高风险候选项。

### FB-9 工具使用文档补充记录

`hack/tools`下每个工具目录已补齐中英文使用说明：

- `hack/tools/build-wasm/README.md`与`README.zh_CN.md`说明动态插件 Wasm 构建入口、参数、输出和注意事项。
- `hack/tools/runtime-i18n/README.md`与`README.zh_CN.md`说明扫描、语言包覆盖校验、allowlist 和退出码。
- `hack/tools/upgrade-source/README.md`与`README.zh_CN.md`说明框架升级、源码插件升级、确认参数、dry-run 和升级前检查。
- `hack/tools/README.md`与`README.zh_CN.md`已补充规则，要求每个工具目录同步维护双语使用说明。

## 优先级建议

1. 先治理响应面：API 错误、请求拦截器、插件 demo-control JSON 错误响应。
2. 再治理交付物：用户、字典、系统参数、岗位、操作日志等导入导出表头和失败原因。
3. 再治理插件平台：pluginbridge、pluginfs、plugindb、catalog、runtime、wasm host service 错误契约。
4. 最后治理低风险面：命令行诊断、测试 fixture、非用户可见注释和仅开发期调试文案。
