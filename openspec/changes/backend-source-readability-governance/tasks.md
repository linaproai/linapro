## 1. 规范与审查标准落地

- [x] 1.1 更新 `AGENTS.md` 后端代码规范，写入主文件契约入口、接口方法详细注释、文件顶部详细说明和 `lina-core/pkg` 公共组件同等治理要求
- [x] 1.2 更新 `.agents/skills/lina-review/SKILL.md`，增加主文件职责、接口方法注释完整度、文件顶部注释质量和分批验证记录审查项
- [x] 1.3 校验 OpenSpec 增量规范与设计文档，确认 `backend-conformance` 变更覆盖宿主、源码插件和 `lina-core/pkg`
- [x] 1.4 记录本轮规范变更的 i18n、缓存一致性、数据权限和开发工具脚本影响判断

执行记录：
- 2026-05-16：已将主文件契约入口、公共组件主文件职责、接口方法详细注释和文件顶部详细说明写入 `AGENTS.md`；已将主文件职责、接口方法注释完整性、文件顶部注释质量和分批验证记录写入 `lina-review` 审查清单。
- 验证通过：`openspec validate backend-source-readability-governance --strict`、`git diff --check -- AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/backend-source-readability-governance`。
- i18n 影响：本轮只修改治理文档和审查规则，不新增或修改前端运行时语言包、manifest i18n 或 apidoc i18n 资源。缓存一致性影响：本轮不修改运行时缓存、失效范围或跨实例同步逻辑。数据权限影响：本轮不新增或修改任何数据操作接口或查询路径。开发工具脚本影响：本轮不新增或修改默认开发、构建、测试脚本入口。

## 2. 基线扫描与任务切片确认

- [x] 2.1 扫描宿主 `apps/lina-core/internal/service/**` 主文件，列出仍包含复杂 receiver 方法实现的组件
- [x] 2.2 扫描 `apps/lina-core/pkg/**` 主文件，列出仍包含复杂实现逻辑的公共组件
- [x] 2.3 扫描源码插件 `apps/lina-plugins/*/backend/internal/service/**` 主文件，列出仍包含复杂 receiver 方法实现的插件组件
- [x] 2.4 将扫描结果按本任务清单的模块批次记录到任务备注，作为后续整改范围和终审对账依据

基线扫描记录：
- 扫描方式：统计目标范围内与目录同名的非测试主文件，并识别 `func (` receiver 方法和包级 `func` 数量，用于定位需要人工判断的复杂实现逻辑。包级构造函数、轻量枚举方法和无复杂逻辑 helper 后续可在分批审查中记录例外。
- 宿主 `internal/service`：49 个主文件，37 个包含 receiver 方法。主要待整改批次：3.x `auth`、`session`、`middleware`、`bizctx`；4.x `user`、`role`、`datascope`、`tenantcap`、`orgcap`；5.x `sysconfig`、`sysinfo`、`dict`、`menu`、`file`、`config`；6.x `cron`、`jobhandler`、`jobmeta`、`jobmgmt`、`startupstats`；7.x `cluster`、`coordination/internal/core`、`coordination/internal/redis`、`cachecoord`、`kvcache`、`locker`、`hostlock`；8.x `i18n`、`usermsg`、`apidoc`、`notify`；9.x/10.x `plugin` 外层和 `plugin/internal/catalog`、`frontend`、`integration`、`lifecycle`、`openapi`、`runtime`、`wasm` 等子组件。
- `lina-core/pkg`：34 个主文件，11 个包含 receiver 方法，25 个包含包级函数。主要待整改批次：11.x 小型公共组件；12.x `pluginhost`、`pluginbridge`、`pluginservice`、`plugindb`、`sourceupgrade`；13.x `dialect`、`i18nresource`。
- 源码插件后端 service：21 个主文件，15 个包含 receiver 方法。主要待整改批次：14.x `org-center` 与 `multi-tenant`；15.x 监控与内容插件；16.x 示例与演示插件。

## 3. 宿主基础与安全服务整改

- [x] 3.1 整改 `auth` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 3.2 整改 `session` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 3.3 整改 `middleware` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 3.4 整改 `bizctx` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 3.5 运行基础与安全服务变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 3.6 调用 `lina-review` 审查本批基础与安全服务整改

执行记录：
- `auth`、`session`、`middleware`、`bizctx` 主文件已保留接口、类型、实现结构体、构造函数和编译期断言，receiver 实现迁移到同包 `_impl.go` 职责文件；`auth` 窄依赖接口方法补充输入、输出、错误和缓存/权限约束说明。
- 验证通过：`cd apps/lina-core && go test ./internal/service/auth ./internal/service/session ./internal/service/middleware ./internal/service/bizctx -count=1`。
- 影响判断：仅同包移动和注释补强，不新增 API、路由、数据访问语义、i18n 资源、缓存失效策略或开发工具脚本；数据权限和缓存行为保持不变。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 4. 宿主用户、角色与数据权限服务整改

- [x] 4.1 整改 `user` 主文件职责、接口方法注释和文件顶部说明，按列表、详情、资料、导入导出、批量操作等职责迁移实现
- [x] 4.2 整改 `role` 主文件职责、接口方法注释和文件顶部说明，按角色、菜单、数据权限、访问缓存等职责迁移实现
- [x] 4.3 整改 `datascope` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 4.4 整改 `tenantcap` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 4.5 整改 `orgcap` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 4.6 运行用户、角色、数据权限相关变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 4.7 调用 `lina-review` 审查本批用户、角色与数据权限服务整改

执行记录：
- `user`、`role`、`datascope`、`tenantcap`、`orgcap` 主文件已收敛为契约入口，原业务实现迁移到同包 `_impl.go` 职责文件；接口注释补充数据权限、租户/组织能力、缓存和错误边界。
- 验证通过：`cd apps/lina-core && go test ./internal/service/user ./internal/service/role ./internal/service/datascope ./internal/service/tenantcap ./internal/service/orgcap -count=1`。
- 影响判断：不改变用户、角色、数据权限查询或缓存失效逻辑；未新增 i18n 资源、数据权限策略、API 行为或脚本入口。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 5. 宿主系统治理服务整改

- [x] 5.1 整改 `config` 主文件职责、接口方法注释和文件顶部说明，保持各配置分组实现文件边界清晰
- [x] 5.2 整改 `sysconfig` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 5.3 整改 `sysinfo` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 5.4 整改 `dict` 主文件职责、接口方法注释和文件顶部说明，按类型、数据、导入导出和 i18n 职责迁移实现
- [x] 5.5 整改 `menu` 主文件职责、接口方法注释和文件顶部说明，按权限树、过滤、校验和 i18n 职责迁移实现
- [x] 5.6 整改 `file` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 5.7 运行系统治理服务变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 5.8 调用 `lina-review` 审查本批系统治理服务整改

执行记录：
- `sysconfig`、`sysinfo`、`dict`、`menu`、`file` 等主文件已保留契约定义，导入、校验、查询、上传、打开等实现迁移到职责文件；补齐 `dict_import.go`、`file_storage.go`、`menu_*` 文件顶部职责说明。
- 验证通过：`cd apps/lina-core && go test ./internal/service/sysconfig ./internal/service/sysinfo ./internal/service/dict ./internal/service/menu ./internal/service/file -count=1`。
- 影响判断：不改变系统配置、字典、菜单、文件存储行为；未新增运行时翻译键、缓存策略、数据权限路径或开发工具脚本。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 6. 宿主任务、调度与运行状态服务整改

- [x] 6.1 整改 `cron` 主文件职责、接口方法注释和文件顶部说明，确保调度注册与具体任务逻辑分文件
- [x] 6.2 整改 `jobhandler` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 6.3 整改 `jobmeta` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 6.4 整改 `jobmgmt` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 6.5 整改 `startupstats` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 6.6 运行任务与调度相关变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 6.7 调用 `lina-review` 审查本批任务、调度与运行状态服务整改

执行记录：
- `cron`、`jobhandler`、`jobmeta`、`jobmgmt`、`startupstats` 主文件已收敛；`jobmgmt/internal/scheduler` 和 `jobmgmt/internal/shellexec` receiver 实现迁移到注册、执行、缓冲等职责文件。
- 验证通过：`cd apps/lina-core && go test ./internal/service/cron ./internal/service/jobhandler ./internal/service/jobmeta ./internal/service/jobmgmt ./internal/service/jobmgmt/internal/scheduler ./internal/service/jobmgmt/internal/shellexec ./internal/service/startupstats -count=1`。
- 子代理验证通过：`cd apps/lina-core && go test ./internal/service/jobmgmt/internal/scheduler ./internal/service/jobmgmt/internal/shellexec -count=1`。
- 影响判断：不改变定时任务注册、执行、日志、Shell 进程控制或启动统计行为；未新增 i18n 资源、缓存策略、数据权限路径或脚本入口。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 7. 宿主缓存、协调与分布式基础服务整改

- [x] 7.1 整改 `cluster` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 7.2 整改 `coordination` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 7.3 整改 `cachecoord` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 7.4 整改 `kvcache` 及其 internal 子组件主文件职责、接口方法注释和文件顶部说明
- [x] 7.5 整改 `locker` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 7.6 整改 `hostlock` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 7.7 运行缓存、协调与锁相关变更包 Go 编译门禁，并记录缓存一致性影响判断
- [x] 7.8 调用 `lina-review` 审查本批缓存、协调与分布式基础服务整改

执行记录：
- `cluster`、`cachecoord`、`kvcache`、`locker`、`hostlock` 主文件已收敛；`coordination/internal/core` 和 `coordination/internal/redis` receiver 实现迁移到 service、lock、kv、revision、event、health 职责文件。
- 验证通过：`cd apps/lina-core && go test ./internal/service/cachecoord ./internal/service/cluster ./internal/service/coordination ./internal/service/coordination/internal/core ./internal/service/coordination/internal/redis ./internal/service/kvcache ./internal/service/locker ./internal/service/hostlock -count=1`。
- 子代理验证通过：`cd apps/lina-core && go test ./internal/service/coordination/internal/core ./internal/service/coordination/internal/redis -count=1`。
- 影响判断：仅移动缓存、协调和锁实现位置，不改变单机/集群分支、Redis 协调、KV 缓存、修订号或失效语义；未新增 i18n、数据权限或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 8. 宿主 i18n、通知与 API 文档服务整改

- [x] 8.1 整改 `i18n` 主文件职责、接口方法注释和文件顶部说明，按 locale、cache、resource、source text、dynamic plugin 等职责保持实现分文件
- [x] 8.2 整改 `notify` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 8.3 整改 `usermsg` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 8.4 整改 `apidoc` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 8.5 运行 i18n、通知和 API 文档相关变更包 Go 编译门禁，并记录 i18n 与缓存一致性影响判断
- [x] 8.6 调用 `lina-review` 审查本批 i18n、通知与 API 文档服务整改

执行记录：
- `i18n` 主文件恢复为接口和构造入口，具体实现迁移到 `i18n_impl.go` 与缓存新鲜度职责文件；`usermsg` i18n 相关投影和基础实现分文件维护。
- 验证通过：`cd apps/lina-core && go test ./internal/service/i18n ./internal/service/notify ./internal/service/usermsg ./internal/service/apidoc -count=1`。
- 影响判断：未新增或删除运行时语言包、manifest i18n 或 apidoc i18n 资源；未改变运行时 bundle、content cache、动态插件翻译失效范围或消息投影行为。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 9. 宿主插件外层服务整改

- [x] 9.1 整改 `plugin` 主文件职责、接口方法注释和文件顶部说明，按列表、生命周期、运行时、升级、前端、OpenAPI、host service 等职责迁移实现
- [x] 9.2 整改 `pluginruntimecache` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 9.3 整改 `pluginhostservices` 主文件职责、接口方法注释和文件顶部说明，具体实现迁移到职责文件
- [x] 9.4 运行插件外层服务变更包 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 9.5 调用 `lina-review` 审查本批宿主插件外层服务整改

执行记录：
- `plugin` 外层服务、`pluginruntimecache`、`pluginhostservices` 主文件已收敛，运行时缓存、目录映射、插件启用快照和 host service 适配实现迁移到职责文件。
- 验证通过：`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/pluginruntimecache ./internal/service/pluginhostservices -count=1`。
- 影响判断：不改变插件安装、启用、运行时缓存、host service 目录或桥接响应语义；未新增 i18n 资源、数据权限路径、缓存拓扑或脚本入口。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 10. 宿主插件内部子组件整改

- [x] 10.1 整改 `plugin/internal/catalog` 主文件职责、接口方法注释和文件顶部说明
- [x] 10.2 整改 `plugin/internal/runtime` 主文件职责、接口方法注释和文件顶部说明
- [x] 10.3 整改 `plugin/internal/integration` 主文件职责、接口方法注释和文件顶部说明
- [x] 10.4 整改 `plugin/internal/frontend`、`openapi`、`lifecycle`、`sourceupgrade` 主文件职责、接口方法注释和文件顶部说明
- [x] 10.5 整改 `plugin/internal/wasm` 主文件职责、接口方法注释和文件顶部说明
- [x] 10.6 运行插件内部子组件变更包 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 10.7 调用 `lina-review` 审查本批宿主插件内部子组件整改

执行记录：
- `catalog`、`runtime`、`integration`、`frontend`、`openapi`、`lifecycle`、`wasm` 主文件已保留接口/契约和构造入口，运行时、启用、OpenAPI、WASM 缓存和桥接实现迁移到职责文件。
- 验证通过：`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/dependency -count=1`。
- 影响判断：不改变插件内部目录、动态路由、Cron、生命周期、WASM host service 或 OpenAPI 投影行为；未新增 i18n 资源、缓存失效语义、数据权限策略或脚本入口。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 11. `lina-core/pkg` 小型公共组件整改

- [x] 11.1 整改 `authtoken`、`bizerr`、`closeutil`、`dbdriver`、`excelutil`、`gdbutil` 主文件职责和文件顶部说明
- [x] 11.2 整改 `logger`、`menutype`、`orgcap`、`pluginfs`、`tenantcap`、`testsupport` 主文件职责和文件顶部说明
- [x] 11.3 补充本批公共组件接口方法注释，确保公共契约可直接指导调用方
- [x] 11.4 运行小型公共组件变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 11.5 调用 `lina-review` 审查本批 `lina-core/pkg` 小型公共组件整改

执行记录：
- 小型公共组件主文件复核完成；`closeutil.Closer` 接口方法补充错误返回契约说明，`menutype.Normalize`/`IsSupported` 迁移到职责文件，`Code.String()` 作为无分支轻量契约方法保留并记录为例外。
- 验证通过：`cd apps/lina-core && go test ./pkg/authtoken ./pkg/bizerr ./pkg/closeutil ./pkg/dbdriver ./pkg/excelutil ./pkg/gdbutil ./pkg/logger ./pkg/menutype ./pkg/orgcap ./pkg/pluginfs ./pkg/tenantcap ./pkg/testsupport -count=1`。
- 影响判断：不改变公共组件导出符号、错误语义、日志行为或菜单类型值；未新增 i18n、缓存、数据权限或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 12. `lina-core/pkg` 插件与桥接公共组件整改

- [x] 12.1 整改 `pluginhost` 主文件职责、接口方法注释和文件顶部说明，确保源码插件公共契约入口清晰
- [x] 12.2 整改 `pluginbridge` 及其子包主文件职责、接口方法注释和文件顶部说明
- [x] 12.3 整改 `pluginservice` 及其子包主文件职责、接口方法注释和文件顶部说明
- [x] 12.4 整改 `plugindb` 及其子包主文件职责、接口方法注释和文件顶部说明
- [x] 12.5 整改 `sourceupgrade` 及其 contract 子包主文件职责、接口方法注释和文件顶部说明
- [x] 12.6 运行插件与桥接公共组件变更包 Go 编译门禁，并记录插件桥接、i18n、缓存一致性和数据权限影响判断
- [x] 12.7 调用 `lina-review` 审查本批插件与桥接公共组件整改

执行记录：
- `pluginbridge`、`pluginservice`、`plugindb`、`sourceupgrade` 主文件已收敛，校验、规范化、运行时调用、配置读取和升级委托实现迁移到职责文件；`pluginservice/contract/doc.go` 补充公共契约包说明。
- 验证通过：`cd apps/lina-core && go test ./pkg/pluginhost/... ./pkg/pluginbridge/... ./pkg/pluginservice/... ./pkg/plugindb ./pkg/sourceupgrade -count=1`。
- 子代理验证通过：`cd apps/lina-core && go test ./pkg/pluginservice/... ./pkg/sourceupgrade ./pkg/menutype -count=1`。
- 影响判断：不改变源码插件公共契约、桥接协议、host service 行为、插件数据库事务或升级代理语义；未新增 i18n、缓存、数据权限或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 13. `lina-core/pkg` 数据库、方言与资源公共组件整改

- [x] 13.1 整改 `dialect` 及其 internal 子包主文件职责、接口方法注释和文件顶部说明
- [x] 13.2 整改 `i18nresource` 主文件职责、接口方法注释和文件顶部说明
- [x] 13.3 运行数据库、方言与资源公共组件变更包 Go 编译门禁，并记录 i18n、缓存一致性和数据权限影响判断
- [x] 13.4 调用 `lina-review` 审查本批数据库、方言与资源公共组件整改

执行记录：
- `dialect`、`i18nresource` 主文件已作为公共契约入口，运行时方言解析、资源目录、加载和选项实现迁移到职责文件；PostgreSQL/SQLite 错误分类接口补充返回码语义注释。
- 验证通过：`cd apps/lina-core && go test ./pkg/dialect/... ./pkg/i18nresource -count=1`。
- 影响判断：不改变 SQL 方言转换、数据库版本读取、i18n 资源加载或缓存语义；未新增数据权限、运行时语言包或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 14. 源码插件组织与租户插件整改

- [x] 14.1 整改 `org-center` 插件 `dept`、`post` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 14.2 整改 `multi-tenant` 插件 `tenant`、`membership`、`tenantplugin` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 14.3 整改 `multi-tenant` 插件 `resolver`、`resolverconfig`、`impersonate`、`provider`、`lifecycleprecondition` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 14.4 运行 `org-center` 与 `multi-tenant` 插件后端 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 14.5 调用 `lina-review` 审查本批源码插件组织与租户插件整改

执行记录：
- `org-center` 与 `multi-tenant` 插件 service 主文件已收敛，receiver 实现迁移到 `_impl.go`；`multi-tenant/impersonate` 的 token signer 窄接口补充签发、解析、错误返回注释。
- 验证通过：基于临时 `go.work` 运行 `org-center`、`multi-tenant` 的 `go test ./backend/internal/service/... -count=1`。
- 影响判断：不改变组织、租户、成员、租户插件、解析器、模拟登录或生命周期前置检查行为；未新增 i18n、缓存、数据权限、插件桥接或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 15. 源码插件监控与内容插件整改

- [x] 15.1 整改 `monitor-loginlog` 插件 `loginlog` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 15.2 整改 `monitor-operlog` 插件 `operlog`、`middleware` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 15.3 整改 `monitor-online` 插件 `monitor` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 15.4 整改 `monitor-server` 插件 `config`、`monitor` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 15.5 整改 `content-notice` 插件 `notice` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 15.6 运行监控与内容插件后端 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 15.7 调用 `lina-review` 审查本批源码插件监控与内容插件整改

执行记录：
- `content-notice`、`monitor-loginlog`、`monitor-online`、`monitor-operlog`、`monitor-server` 插件 service 主文件已收敛，receiver 实现迁移到 `_impl.go`。
- 验证通过：基于临时 `go.work` 运行内容与监控插件的 `go test ./backend/internal/service/... -count=1`。
- 影响判断：不改变通知、登录日志、操作日志、在线用户、服务器监控配置或采集查询行为；未新增 i18n、缓存、数据权限、插件桥接或脚本影响。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## 16. 源码插件示例与演示插件整改

- [x] 16.1 整改 `demo-control` 插件 `middleware` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 16.2 整改 `plugin-demo-source` 插件 `demo` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 16.3 整改 `plugin-demo-dynamic` 插件 `dynamic` 服务主文件职责、接口方法注释和文件顶部说明
- [x] 16.4 运行示例与演示插件后端 Go 编译门禁，并记录 i18n、缓存一致性、数据权限和插件桥接影响判断
- [x] 16.5 调用 `lina-review` 审查本批源码插件示例与演示插件整改

执行记录：
- 示例与演示插件 service 主文件复核完成；当前 `demo-control`、`plugin-demo-source`、`plugin-demo-dynamic` 后端 service 主文件未保留复杂 receiver 实现，符合本轮主文件职责要求。
- 验证通过：基于临时 `go.work` 运行示例与演示插件的 `go test ./backend/internal/service/... -count=1`。
- 影响判断：不改变示例插件运行时行为、动态插件服务契约、i18n、缓存、数据权限、插件桥接或脚本入口。
- 审查结论：纳入本轮 `lina-review` 综合审查，未发现本批阻断问题。

## Feedback

- [x] **FB-1**: 多个宿主与源码插件新增实现文件顶部注释过于泛化，未说明实现文件职责、主要流程和关键约束

执行记录：
- 已将宿主 `auth_impl.go`、`bizctx_impl.go`、`cluster_impl.go`、`datascope_impl.go`、`middleware_impl.go`、`orgcap_impl.go`、`role_impl.go`、`session_impl.go`、`tenantcap_impl.go` 和 `user_impl.go` 顶部注释改为覆盖对应实现切片、主要流程和运行期依赖、数据范围、租户、缓存或可选能力边界的文件职责说明。
- 已将源码插件 `content-notice`、监控插件、`multi-tenant` 和 `org-center` 新增 `_impl.go` 顶部注释改为覆盖插件服务的 CRUD、查询、导入导出、租户过滤、生命周期、能力提供或采集边界说明。
- 该反馈属于项目治理类修复，不改变运行时行为；验证通过：宿主和源码插件 161 个变更 Go 文件顶部注释结构与泛化注释扫描、`gofmt`、`openspec validate backend-source-readability-governance --strict`、`git diff --check`、`git diff --cached --check`、`git -C apps/lina-plugins diff --check`。
- 验证通过：`cd apps/lina-core && go test` 覆盖本变更涉及的宿主 service 与 `pkg` 包；基于临时 `go.work` 运行本次涉及源码插件的 `go test .../backend/internal/service/... -count=1`。

## 17. 全量复核与治理验证

- [x] 17.1 全量扫描宿主、源码插件和 `lina-core/pkg` 主文件，确认复杂实现逻辑已按任务范围迁出或记录明确例外
- [x] 17.2 全量扫描新增或修改的接口定义，确认接口方法注释覆盖功能、输入、输出、错误和关键约束
- [x] 17.3 全量扫描新增或修改的 Go 文件顶部注释，确认主文件和非主文件注释层级正确
- [x] 17.4 运行 `openspec validate backend-source-readability-governance --strict`
- [x] 17.5 运行 `git diff --check` 覆盖本变更所有文档和 Go 文件
- [x] 17.6 汇总所有分批 Go 编译门禁结果，若涉及控制器、路由绑定、启动编排或 API 签名变更则补充 `cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] 17.7 记录本变更最终 i18n、缓存一致性、数据权限、开发工具脚本和 E2E 影响判断
- [x] 17.8 调用 `lina-review` 完成终审并修复审查发现

终审记录：
- 主文件 receiver 扫描结果：宿主、源码插件和 `lina-core/pkg` 目标范围内仅剩 `apps/lina-core/pkg/menutype/menutype.go` 的 `Code.String()`，该方法无分支、无外部调用，属于设计允许的轻量契约方法例外。
- 接口注释扫描结果：新增或修改范围内未发现缺少紧邻注释的接口方法；已补充 `auth` 窄接口、`closeutil.Closer`、方言错误分类接口和 `multi-tenant` token signer 的使用契约说明。
- 文件顶部注释扫描结果：新增和修改 Go 文件均具备顶部职责说明；`pkg/dialect/internal/{postgres,sqlite}/dialect.go` 作为各 internal 子包主说明文件保留 package 注释。
- 最终验证通过：`openspec validate backend-source-readability-governance --strict`。
- 最终验证通过：`git diff --check -- AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/backend-source-readability-governance apps/lina-core apps/lina-plugins`。
- Go 编译门禁汇总：3.x 至 16.x 分批 `go test` 均通过；本变更不涉及 Controller 构造函数、路由绑定、启动编排或 API 签名变更，因此未额外运行 `cd apps/lina-core && go test ./internal/cmd -count=1`。
- 最终影响判断：本变更为源码组织、注释和治理规范更新，不改变 REST API、数据库、权限、数据权限、缓存一致性模型、插件桥接协议、i18n 资源、前端页面、E2E 工作流或开发工具脚本；无需新增 E2E 测试。
- `lina-review` 终审结论：已按主文件职责、接口方法注释、文件顶部注释、Go 编译门禁、i18n、缓存一致性、数据权限和脚本影响进行审查，未发现阻断问题。
