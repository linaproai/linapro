## 1. 插件 ID 治理契约

- [x] 1.1 新增或抽取插件 ID 解析/校验组件，覆盖运行时基础安全边界，并将 `<author>-<domain>-<capability>` 保留为官方命名建议
- [x] 1.2 将 manifest ID、动态 artifact manifest ID、插件依赖 ID 和源码插件注册 ID 校验统一接入新的插件 ID 治理组件
- [x] 1.3 为插件 ID 解析、运行时安全边界、非三段 ID 接受、依赖 ID 校验和长度限制补充后端单元测试
- [x] 1.4 更新错误信息、接口文档和中英文 i18n 资源，确保插件 ID 校验失败返回稳定错误码、messageKey 和英文 fallback

## 2. 官方插件破坏式改名

- [x] 2.1 按映射重命名官方插件目录、`plugin.yaml` ID、README、manifest README 和示例文本
- [x] 2.2 同步更新官方插件 Go module 名称、import 路径、`apps/lina-plugins/go.mod` replace、`apps/lina-plugins/lina-plugins.go` 和 GoFrame 生成配置
- [x] 2.3 更新官方插件源码注册常量和内部常量，包括 `linapro-ops-demo-guard` 的中间件启用态检测
- [x] 2.4 更新宿主官方插件常量、稳定菜单父级映射、`orgcap.ProviderPluginID`、`tenantcap.ProviderPluginID`、启动一致性检查和 provider 检测逻辑
- [x] 2.5 更新 `plugin.autoEnable` 默认/开发/镜像配置和配置解析测试中的官方插件 ID

## 3. 插件自有存储与生成代码

- [x] 3.1 将官方插件自有 SQL 表、索引、约束、mock 数据、uninstall SQL 和 SQL 注释按新插件 ID snake_case 范围重命名
- [x] 3.2 重新生成或同步官方插件本地 DAO/DO/Entity，确保生成配置、表名、服务代码和测试 fixture 与新表名一致
- [x] 3.3 更新动态插件 host service data 资源授权表名、示例 SQL、资源声明和相关权限审查展示测试
- [x] 3.4 补充 SQL/DAO 静态验证，确认官方插件运行时代码不再引用旧插件表名

## 4. 派生运行时身份同步

- [x] 4.1 更新官方插件 manifest 菜单 key、parent_key、权限字符串、路由 path、cron handlerRef、job i18n key 和菜单 i18n key
- [x] 4.2 更新动态插件 artifact 构建逻辑和测试 fixture，确保 `.wasm` 文件名、manifest、`/plugin-assets/<id>/...` 和 `/api/v1/extensions/<id>/...` 使用新 ID
- [x] 4.3 更新插件管理、菜单、用户首页、API 文档、任务管理、运行时路由、host service 授权和 scheduler 相关测试中的插件 ID
- [x] 4.4 增加旧官方 ID 残留扫描，排除归档历史说明后，运行时代码、配置、测试和活跃 OpenSpec 文档不得再使用旧 ID

## 5. i18n、apidoc 与文档

- [x] 5.1 更新官方插件 `manifest/i18n/<locale>/*.json` 中的 `plugin.<plugin-id>.`、菜单 key、错误 key、job key 和页面文案引用
- [x] 5.2 更新官方插件和宿主 `manifest/i18n/<locale>/apidoc/**/*.json` 中的 `plugins.<plugin_id_snake_case>.` namespace
- [x] 5.3 更新根 README、插件工作区 README、插件开发说明、i18n README、E2E README 和相关 OpenSpec baseline 示例中的插件 ID
- [x] 5.4 运行 JSON 校验和 runtime i18n 治理扫描，确认无硬编码旧插件 ID、无无效翻译键、无旧 apidoc namespace

## 6. 前端与 E2E

- [x] 6.1 更新前端插件管理、动态页面、菜单路由、用户首页、测试 page object 和 fixture 中的官方插件 ID
- [x] 6.2 更新或新增 Playwright E2E，覆盖插件列表显示新官方 ID、`linapro-ops-demo-guard` 启用后写请求阻断、动态插件资产路径和扩展 API 新路径
- [x] 6.3 按 `lina-e2e` 规范为新增 E2E 分配 TC ID，并运行 `pnpm -C hack/tests test:validate`
- [x] 6.4 运行前端 typecheck 和受影响 E2E，确认页面无旧插件 ID、无原始 i18n key、无动态资源路径断裂

## 7. 后端编译、缓存与一致性验证

- [x] 7.1 运行受影响宿主包测试，至少覆盖插件 catalog、runtime、integration、jobhandler、jobmgmt、menu、orgcap、tenantcap 和启动绑定包
- [x] 7.2 运行官方插件后端包测试，覆盖 `linapro-content-notice`、`linapro-monitor-loginlog`、`linapro-monitor-operlog`、`linapro-monitor-online`、`linapro-monitor-server`、`linapro-tenant-core`、`linapro-org-core`、`linapro-demo-source` 和 `linapro-ops-demo-guard`
- [x] 7.3 重新构建并验证 `linapro-demo-dynamic` 动态插件 artifact，确认 manifest、host service、frontend、SQL、i18n 和 apidoc 均使用新 ID
- [x] 7.4 明确记录缓存一致性结论：本变更不新增缓存类型，所有状态、菜单、路由、cron、i18n 和 apidoc 刷新继续使用插件 ID scope 精确失效，集群模式沿用既有广播/共享修订号机制
- [x] 7.5 明确记录数据权限结论：本变更不新增业务数据访问接口；官方插件改名后的列表、详情、导出、写操作和插件 host service 访问仍复用既有租户与数据权限边界

## 8. OpenSpec 与最终审查

- [x] 8.1 运行 `openspec validate normalize-plugin-ids --strict`
- [x] 8.2 运行 `git diff --check` 覆盖本变更涉及的代码、测试、i18n、SQL 和 OpenSpec 文档
- [x] 8.3 汇总所有 Go 测试、前端 typecheck、E2E、i18n 扫描、旧 ID 残留扫描和 OpenSpec 校验结果到本任务记录
- [x] 8.4 完成实现后调用 `lina-review`，重点审查官方插件 ID 映射一致性、旧 ID 残留、manifest 校验覆盖、i18n/apidoc namespace、缓存作用域、数据权限边界和后端 Go 编译门禁

## Feedback

- [x] **FB-1**: 插件 ID 命名规范不应作为宿主运行时严格校验，结构化命名仅作为官方建议保留
- [x] **FB-2**: 宿主服务不应硬编码插件 domain、官方保留 capability 或旧官方插件 ID 拒绝表
- [x] **FB-3**: 宿主前端页面不应通过官方租户插件 ID 推导租户管理能力，泛化插件管理 UI 文案与测试标识也不应绑定具体官方插件 ID

## 执行记录

- Go 单元测试：`go run ./hack/tools/linactl test.go plugins=1 race=false verbose=false` 通过，覆盖 `apps/lina-core`、`hack/tools/linactl`、`apps/lina-plugins` 聚合模块以及 10 个官方源码插件后端包；此前也已用 `-count=1` 覆盖插件 catalog/runtime/integration/dependency/sourceupgrade/wasm、apidoc、cron、jobmgmt、menu、orgcap、tenantcap 和 `apps/lina-core/internal/cmd` 启动绑定包。
- 前端单元与类型检查：`pnpm --dir apps/lina-vben run test:unit` 通过，41 个 test files、344 tests；`pnpm --dir apps/lina-vben run check:type` 通过。
- E2E 治理与类型检查：`pnpm -C hack/tests test:validate` 通过，验证 223 个 E2E 文件、39 个 scope；`pnpm --dir hack/tests exec tsc -p tsconfig.json --noEmit` 通过；`pnpm -C hack/tests test:service-deps` 通过，服务依赖治理扫描为 28 个基线文件、113 个允许构造调用。
- 动态插件构建：`go run ./hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output` 通过，生成 `temp/output/linapro-demo-dynamic.wasm`。
- 全量 E2E：在重建数据库与 mock 数据后，使用 `E2E_BASE_URL=http://127.0.0.1:5667 E2E_BROWSER_CHANNEL=chrome E2E_RETRIES=1 E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests test:full` 通过；parallel 阶段 34 passed、2 skipped，serial 阶段 512 passed、9 skipped。由于 Playwright Chromium CDN 下载多次卡住，本轮按用户要求持续重试后改用本机系统 Chrome 作为浏览器通道。
- OpenSpec 与静态校验：`openspec validate normalize-plugin-ids --strict` 通过；`git diff --check` 与 `git -C apps/lina-plugins diff --check` 均通过；`go run ./hack/tools/linactl i18n.check` 通过，runtime i18n violations 为 0。
- 旧 ID 残留扫描：运行 `rg --pcre2 "(?<!linapro-)(content-notice|monitor-loginlog|monitor-operlog|monitor-online|monitor-server|multi-tenant|org-center|plugin-demo-dynamic|plugin-demo-source|demo-control)" ...`。反馈修正后生产运行时代码不再保留 `legacyOfficialPluginIDs` 拒绝表；剩余旧官方 ID 命中应仅为本变更映射说明、移除场景、测试命名或负向资源一致性场景，配置、manifest、前端正向路径或 E2E 正向路径不得继续依赖旧官方 ID。
- GoFrame DAO 说明：`gf gen dao` 在 `linapro-demo-source` 当前数据库未安装插件自有表 `plugin_linapro_demo_source_record` 时无法直接生成；本轮按新表名手工同步插件本地 DAO/DO/Entity 生成产物，并通过官方插件后端包测试、源码插件生命周期 E2E 和 SQL/旧 ID 扫描验证。
- 缓存一致性结论：本变更不新增缓存类型；插件状态、菜单、路由、cron、i18n、apidoc 与动态运行时刷新继续使用插件 ID scope 精确失效。单机模式沿用本地失效/刷新，集群模式沿用既有广播、共享修订号和共享后端机制，不引入清空所有语言、所有扇区或所有插件缓存的普通业务路径。
- 数据权限结论：本变更不新增业务数据访问接口；官方插件改名后的列表、详情、导出、写操作、插件 host service data 访问和租户治理路径继续复用既有租户隔离与角色数据权限边界。全量 E2E 中已覆盖在线用户、文件、用户、任务、租户、通知、审计日志等数据权限场景。
- i18n 影响结论：本变更涉及插件 manifest/i18n、前端运行时语言包、apidoc i18n、菜单/job/error key 和文档示例，已同步更新并通过 runtime i18n 扫描、前端类型检查、全量 i18n E2E 与旧 apidoc namespace 扫描。

## Feedback 执行记录

- FB-1/FB-2 规范调整：`plugin-id-governance` 与 `plugin-manifest-lifecycle` 已改为只强制插件 ID 基础安全边界；`<author>-<domain>-<capability>`、domain 分类和 `core` capability 仅作为官方插件命名建议与仓库治理约定，不再作为宿主运行时阻断规则。
- FB-1/FB-2 代码调整：`catalog.ValidatePluginID` 仅校验非空、64 字符长度上限和 lowercase kebab-case；移除宿主运行时的 domain 白名单、官方 `core` capability 判断和 `legacyOfficialPluginIDs` 拒绝表；`demo-control`、`acme-random-report`、`acme-org-core`、`plugin-demo-source` 等运行时安全 ID 均可通过通用校验。
- FB-1/FB-2 依赖治理调整：插件依赖 ID 继续复用基础安全校验；依赖快照构建不再通过旧严格命名规则跳过 registry 行，而是忽略未发现 manifest 且没有 release 引用的无关孤儿 registry 行。已有 release 引用但 snapshot 不可信的已安装插件仍按既有保守反向依赖策略阻断卸载。
- FB-1/FB-2 验证：`go test ./internal/service/plugin/internal/catalog -count=1`、`go test ./internal/service/plugin/internal/dependency -count=1`、`go test ./internal/service/plugin -count=1` 均通过；`openspec validate normalize-plugin-ids --strict` 通过；静态搜索确认宿主运行时代码不再存在 domain/core/legacy ID 硬编码校验入口。
- i18n 影响结论：本反馈未新增、修改或删除用户可见翻译键，已有 `error.plugin.id.invalid` 继续承载基础安全校验失败文案；不需要同步修改运行时语言包或 apidoc i18n。
- 缓存一致性结论：本反馈不新增缓存类型，也不改变插件状态、菜单、路由、cron、i18n 或 apidoc 的缓存失效 scope；集群和单机缓存策略不受影响。
- 数据权限结论：本反馈不新增或修改业务数据访问接口，不改变租户隔离、角色数据权限或插件 host service 授权边界。
- FB-3 代码调整：新增前端管理能力解析门面 `plugins/management-capabilities.ts`，宿主布局、用户管理与角色管理页面改为通过 `tenant.management` / `organization.management` 能力判断页面降级与字段显示，不再直接读取官方租户插件 ID；`slot-registry` 增加能力 provider 运行时状态查询，保留已注册 provider 被禁用时关闭宿主租户 UI 的行为；插件管理泛化列与详情的 `data-testid` 以及英文帮助文案恢复为 multi-tenant 能力概念。
- FB-3 验证：`pnpm --dir apps/lina-vben exec vitest run --dom apps/web-antd/src/plugins/management-capabilities.test.ts apps/web-antd/src/components/tree/src/permission-display.test.ts` 通过，补修后单独重跑 `apps/web-antd/src/plugins/management-capabilities.test.ts` 通过；`pnpm --dir apps/lina-vben run check:type` 通过，补修后重跑仍通过；`pnpm --dir apps/lina-vben run test:unit` 通过，42 个 test files、347 tests；`pnpm -C hack/tests test:validate` 通过；`pnpm --dir hack/tests exec tsc -p tsconfig.json --noEmit` 通过；`E2E_BROWSER_CHANNEL=chrome pnpm --dir hack/tests exec playwright test hack/tests/e2e/extension/plugin/TC0078-plugin-detail-dialog.ts --workers=1` 通过，2 passed；`openspec validate normalize-plugin-ids --strict` 通过；`git diff --check` 覆盖本反馈文件通过；静态扫描确认宿主布局、用户、角色和插件管理泛化 UI 不再残留 `linapro-tenant-core` 能力判断或泛化测试标识。补充运行针对本反馈前端文件的 `pnpm --dir apps/lina-vben exec eslint ...`，结果发现大量既有/已暂存文件的 import 排序、Prettier 与 Tailwind 规则问题；本反馈新增 helper 自身暴露的问题已修复，剩余 lint 输出主要来自既有文件格式治理，不作为本反馈阻断项。
- FB-3 i18n 影响结论：本反馈修改了英文运行时帮助文案，将官方插件 ID 表述恢复为 multi-tenant 能力概念；中文文案已保持能力概念，无需修改；未新增或删除翻译键，未影响插件 manifest i18n 或 apidoc i18n。
- FB-3 缓存一致性结论：本反馈仅调整前端能力解析与页面显示判断，不新增后端缓存、缓存键、失效路径或跨实例同步机制；前端仍复用既有插件运行时状态快照刷新与 registry listener。
- FB-3 数据权限结论：本反馈不新增或修改 HTTP/API 数据操作接口，不改变用户、角色、租户或插件 host service 的数据权限边界；页面字段显示仍由现有后端权限和能力状态共同约束。
