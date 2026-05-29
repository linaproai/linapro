## 1. 规则与历史语义核对

- [x] 1.1 实现前重新读取命中的规则文件：`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`database.md`、`i18n.md`、`cache-consistency.md`，并在任务记录中写明影响判断。
- [x] 1.2 静态检索历史 `Metadata`、`metadata.get`、`service: metadata`、`metadata.yaml`、`声明型资源`、`HostServices.Manifest()` 相关引用，按设计中的语义分流表确认需要删除、重命名、保留或仅更新注释的项目。
- [x] 1.3 确认宿主框架自身 `apps/lina-core/manifest/config/metadata.yaml`、动态路由元数据、审计元数据、错误元数据、数据库表元数据和 cron 展示 metadata 不属于插件 manifest 资源读取服务，避免误删历史功能。

执行记录：已读取命中规则文件。影响判断：本变更属于`apps/lina-core`插件宿主通用能力和动态插件 host service 授权边界调整；不新增 HTTP API、前端页面、数据库迁移、DAO、运行期依赖或缓存域。`i18n`影响仅为允许读取插件`manifest/i18n/`原文，不改变加载、聚合、缓存失效或翻译治理；SQL影响仅为允许读取插件`manifest/sql/`原文，不执行 SQL 或改变迁移账本；数据权限无业务数据读取/写入影响；开发工具跨平台无影响。静态检索确认无插件可见`Metadata()`、`MetadataService`、`metadata.get`、`service: metadata`入口；保留的 metadata 用法属于宿主系统信息、动态路由、审计/错误/数据库表/cron 展示等非插件 manifest 读取语义。

## 2. 插件 Manifest 契约收敛

- [x] 2.1 更新 `pkg/plugin/capability/contract`、`pkg/plugin/capability`、`pkg/plugin/capability/guest` 中 `Manifest()` 的注释和接口语义，明确其为插件自有 `manifest/` 原始资源只读视图。
- [x] 2.2 删除或迁移任何插件可见的旧 `Metadata()`、`MetadataService`、`metadata.get`、`service: metadata` 或等价读取入口，不保留 deprecated alias。
- [x] 2.3 更新 `Scan` 相关注释和测试命名，明确它是 YAML 便捷扫描能力，不代表 `Manifest()` 只能读取 YAML。

执行记录：已更新`Manifest()`契约和实现注释为插件自有`manifest/`原始资源只读视图；未发现可删除的独立旧`Metadata`公开入口；`Scan`继续只作为 YAML 便捷扫描方法。

## 3. 源码插件读取实现

- [x] 3.1 修改源码插件 manifest resource path 规范化逻辑，移除 `config/`、`sql/`、`i18n/` 排除规则，保留空根、绝对路径、URL、Windows drive path、路径穿越、重复 `manifest/` 前缀和跨插件路径拒绝。
- [x] 3.2 保持源码插件读取来源顺序和作用域：优先读取当前插件嵌入文件系统，再读取当前仓库开发目录中的同一插件 `manifest/`，不得读取宿主或其他插件资源。
- [x] 3.3 补充源码插件单元测试，覆盖读取 `metadata.yaml`、`config/config.example.yaml`、`sql/001-schema.sql`、`i18n/zh-CN/plugin.json`、缺失资源、路径穿越和跨插件拒绝。

执行记录：`manifest`路径规范化已移除`config/sql/i18n`保留目录拒绝，读取来源顺序和`readContainedFile`作用域保护保持不变；单元测试已覆盖新允许路径、缺失资源和越界拒绝。

## 4. 动态插件授权与 host service

- [x] 4.1 修改 `pluginbridge` host service path 校验和 WASM host call 授权逻辑，允许 `manifest` 服务声明和访问 `config/`、`sql/`、`i18n/` 相对路径。
- [x] 4.2 保持动态插件 `service: manifest` 的 `methods: [get]` 和 `resources.paths` 授权快照校验，未授权路径必须拒绝。
- [x] 4.3 补充 `pluginbridge` 和 `wasm` 单元测试，覆盖动态插件授权读取 `config/config.example.yaml`、`sql/001-schema.sql`、`i18n/zh-CN/plugin.json`，以及未授权路径拒绝。

执行记录：`pluginbridge`声明校验和 WASM 授权匹配均允许合法专用目录相对路径；动态插件仍必须具备`service: manifest`、`get`方法和`resources.paths`授权，测试覆盖授权读取和未授权拒绝。

## 5. 动态 artifact 资源视图和打包语义

- [x] 5.1 修改 dynamic active release artifact manifest resource projection，完整投影 artifact 中 `manifest/` 下实际携带的资源，不再只投影 `*.yaml` 或排除 `config/`、`sql/`、`i18n/`。
- [x] 5.2 保持 `HostServices.Config()` 的默认配置读取语义，继续从 active release 中识别 `manifest/config/config.yaml`，但不把 `Manifest()` 读取结果作为运行期有效配置自动生效。
- [x] 5.3 检查动态插件构建和 artifact 解析逻辑，确保 `go:embed` 和目录扫描资源来源都保留 `manifest/config/`、`manifest/sql/`、`manifest/i18n/` 及其他 manifest 资源原始路径。
- [x] 5.4 更新资源计数、示例插件 `plugin.yaml` 或测试 fixture，使动态插件通过 `service: manifest` 显式声明需要读取的 manifest 资源路径。

执行记录：artifact 解析不再限制专用目录或`.yaml`扩展名；active release manifest resource projection 完整输出`manifest/`下资源。`buildArtifactDefaultConfig`仍只读取`manifest/config/config.yaml`作为默认配置来源，`Manifest()`只返回原始字节。未修改动态插件构建器或示例插件目录；通过 runtime/WASM 测试 fixture 更新资源计数与`service: manifest`授权路径覆盖。

## 6. 历史命名和文档清理

- [x] 6.1 清理生产代码、测试、fixture 和注释中把 `Manifest()` 描述为 `Metadata` 或“声明型资源读取器”的历史表述。
- [x] 6.2 更新相关 README 或开发说明时同步维护中英文镜像；若最终无需修改目录级说明文档，在任务记录中明确无文档镜像影响。
- [x] 6.3 运行静态检索确认插件资源读取路径中不存在旧 `Metadata()`、`MetadataService`、`metadata.get`、`service: metadata` 或 deprecated alias。

执行记录：已清理`Manifest()`相关旧“declaration resource”表述；未修改 README 或目录级说明文档，无中英文镜像影响。静态检索确认插件资源读取路径中无旧 Metadata 读取入口残留。

## 7. 验证与审查

- [x] 7.1 运行 `go test ./pkg/plugin/capability/manifest ./pkg/plugin/capability/guest ./pkg/plugin/pluginbridge/internal/hostservice -count=1`。
- [x] 7.2 运行 `go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`。
- [x] 7.3 若修改动态插件构建器或官方示例插件，运行对应构建器/示例插件测试，并记录跨平台影响；若未修改开发工具，记录开发工具无影响。
- [x] 7.4 运行 `openspec validate replace-plugin-metadata-with-manifest --strict`。
- [x] 7.5 运行 `git diff --check` 和旧 Metadata 静态检索，确认没有格式问题和旧读取入口残留。
- [x] 7.6 完成任务后调用 `lina-review`，审查结论必须覆盖 `i18n`、缓存一致性、数据权限、SQL、开发工具跨平台、DI 来源和测试策略影响。

执行记录：两组 Go 测试均已通过。未修改动态插件构建器、开发工具、脚本、CI 或官方示例插件目录，开发工具跨平台无影响。`openspec validate replace-plugin-metadata-with-manifest --strict`、`git diff --check`、旧 Metadata 读取入口静态检索和旧“声明型资源/专用管线排除/YAML-only”表述检索均已通过。

审查记录：已按`lina-review`读取`AGENTS.md`以及`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`database.md`、`i18n.md`、`cache-consistency.md`、`data-permission.md`、`documentation.md`、`dev-tooling.md`、`api-contract.md`。未发现阻塞问题。`i18n`影响已限定为读取原文且不注册翻译或触发缓存失效；缓存一致性无新增缓存域，仍依赖 active release artifact 快照；数据权限无业务数据接口或租户/组织数据可见性变化；SQL 仅允许读取原文且不执行 SQL、不改账本；开发工具跨平台无影响；DI 来源无新增运行期依赖或构造函数变更；测试策略由 manifest、pluginbridge、runtime、wasm 单元测试和静态治理校验覆盖。

## Feedback

- [x] **FB-1**: `linapro-demo-dynamic`缺少`service: manifest`配置示例

执行记录：根因是本变更已要求动态插件通过`plugin.yaml`中的`service: manifest`、`methods: [get]`和`resources.paths`显式授权读取插件`manifest/`原始资源，但官方动态插件示例只提供了`manifest/config/`、SQL 和 i18n 资源文件，未在`hostServices`中展示`manifest.get`授权写法，也未在中英文说明文档中解释`manifest.get`与专用`config`服务的边界。已更新`apps/lina-plugins/linapro-demo-dynamic/plugin.yaml`，新增`service: manifest`示例并授权`config/config.example.yaml`、`config/config.yaml`、SQL 和 i18n JSON 路径；同步更新该插件根目录和`manifest/`目录的中英文 README，说明`manifest.get`仅读取打包原文，不替代配置、SQL 或 i18n 生命周期管线。

影响分析：已按`AGENTS.md`读取`openspec.md`、`plugin.md`、`documentation.md`、`i18n.md`、`architecture.md`、`testing.md`、`cache-consistency.md`、`data-permission.md`、`dev-tooling.md`、`backend-go.md`、`database.md`和`api-contract.md`。插件根目录无本地`AGENTS.md`。本反馈修改官方动态插件示例的`plugin.yaml`和文档，不新增或修改 Go 生产代码、HTTP API、数据库迁移、SQL 内容、前端页面或运行期依赖；DI 来源无新增依赖或构造函数变更。`i18n`影响为文档和插件清单注释说明，不新增运行时用户可见文案或翻译键；缓存一致性无新增缓存域或失效策略变更；数据权限无业务数据读写接口影响；SQL 影响仅为在授权示例中引用已有 SQL 文件路径，不修改 SQL 内容、账本、幂等性或执行管线；API 契约无影响；开发工具跨平台无脚本、CI、构建入口实现变更；测试策略按治理类反馈使用 YAML 解析、OpenSpec 校验、格式检查、清单解析测试和动态插件构建器示例测试覆盖。

验证记录：已通过`ruby -e 'require "yaml"; ...'`解析`plugin.yaml`和配置 YAML；`openspec validate replace-plugin-metadata-with-manifest --strict`通过；顶层`git diff --check`和`git -C apps/lina-plugins diff --check`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -run 'Test.*Demo|Test.*Official' -count=1`通过；静态检索确认示例中存在`service: manifest`及关键授权路径，且无`service: metadata`或`metadata.get`残留。

审查记录：已按`lina-review`执行反馈级审查，范围包含`apps/lina-plugins/linapro-demo-dynamic/plugin.yaml`、根目录中英文 README、`manifest/`中英文 README 和本反馈任务记录。未发现阻塞问题。规则域结论：插件目录规范通过，动态插件`manifest`服务使用显式`methods`和`resources.paths`；文档规范通过，中英文镜像同步维护；OpenSpec 记录通过，反馈已先记录、再修复并补充根因、影响和验证；`i18n`、缓存一致性、数据权限、SQL、API、开发工具跨平台和 DI 来源均无运行时行为变更风险；本反馈为治理类示例补齐，无需新增单元测试或 E2E。

- [x] **FB-2**: `linapro-demo-dynamic`的`manifest`示例授权过宽且缺少声明到使用的可见闭环

执行记录：根因是 FB-1 为了展示`manifest.get`可读取专用资源目录，把`config.example`、SQL 和 i18n 资源都放入了官方动态插件示例的`hostServices.resources.paths`，示例授权范围过宽，且页面或接口没有展示这些声明路径被真实读取，无法证明“先声明、后使用”的流程已走通。已新增`manifest/config/profile.yaml`作为专用演示资源，并将`plugin.yaml`中的`service: manifest`授权收敛为仅`config/config.yaml`和`config/profile.yaml`。后端新增只读`GET /api/v1/manifest-demo`接口，复用`Manifest()`host service 读取`config/profile.yaml`与`config/config.yaml`，同时在原`host-call-demo`响应中补充 manifest 摘要；内嵌页面新增`Manifest Host Service`面板，加载`manifest-demo`并展示 profile 路径、profile 值、config 路径和配置预览。同步更新插件根目录与`manifest/`目录中英文 README、插件运行时中英文语言包、中文 apidoc 翻译资源，并新增插件自有 E2E `TC005-manifest-host-service-demo.ts`和 POM 定位器。

影响分析：已按`AGENTS.md`读取`openspec.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`frontend-ui.md`、`testing.md`、`documentation.md`、`i18n.md`、`architecture.md`、`data-permission.md`、`cache-consistency.md`、`dev-tooling.md`和`database.md`，并使用`goframe-v2`、`lina-feedback`、`lina-review`、`lina-e2e`和`frontend-design`技能。插件根目录无本地`AGENTS.md`。本反馈修改动态插件目录、插件 host service 授权、Go 后端 service/controller/API DTO、插件前端页面、插件运行时 i18n、插件 apidoc i18n、E2E 和中英文 README。新增接口为只读`GET /manifest-demo`，不会触发存储、数据或网络副作用；原`host-call-demo`仍为执行型演示接口，仅追加 manifest 响应投影。数据权限影响：`manifest-demo`只读取当前插件打包资源，不读取业务数据、租户数据或组织数据；原`host-call-demo`的数据演示仍沿用既有`data`host service 授权和临时记录清理逻辑。缓存一致性影响：不新增缓存、快照或失效路径；读取来源仍为动态插件 active release artifact/guest host service 当前授权视图。SQL 影响：不修改 SQL 内容、迁移、DAO 或执行账本，且 SQL 资源不再作为示例 manifest 授权路径。开发工具跨平台影响：不修改脚本、Makefile、linactl 或 CI；仅运行既有构建和测试入口。DI 来源：`serviceImpl`新增`manifestSvc`依赖，owner 为插件 guest host service 目录`guest.Default().Manifest()`，在插件 service 构造函数`New()`中与既有`runtime/config/hostConfig`host service client 同源创建，不新增宿主服务图或缓存敏感共享实例。接口性能：页面新增单次只读请求，不产生列表瀑布式调用或`N+1`查询。

验证记录：`GOWORK=off go test ./backend/internal/service/dynamic ./backend/internal/controller/dynamic ./backend/api/dynamic ./backend/api/dynamic/v1 -count=1`通过；`GOWORK=off go test ./backend/api -count=1`通过；`GOWORK=off go test ./backend/internal/service/dynamic -run 'TestRunHostCallDemoManifestReadsAuthorizedResources|TestRunHostCallDemoConfigReadsPluginAndHostConfigValues' -count=1`通过；`cd apps/lina-core && go test ./internal/service/apidoc -count=1`通过；`cd apps/lina-core && go test ./pkg/i18nresource -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -run 'Test.*Demo|Test.*Official' -count=1`通过；`make wasm p=linapro-demo-dynamic out=../../temp/output`通过；`ruby -e 'require "yaml"; ...' plugin.yaml manifest/config/config.yaml manifest/config/config.example.yaml manifest/config/profile.yaml`通过；`node -e ...`解析插件运行时和 apidoc JSON 通过；`cd hack/tests && ./node_modules/.bin/tsc -p tsconfig.json --noEmit --pretty false`通过；`cd hack/tests && pnpm test:validate`通过；`openspec validate replace-plugin-metadata-with-manifest --strict`通过；顶层`git diff --check`和`git -C apps/lina-plugins diff --check`通过。尝试运行`pnpm exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC005-manifest-host-service-demo.ts --config=playwright.config.ts --project=chromium`，因本地`http://127.0.0.1:9120/admin/auth/login`未启动返回`ERR_CONNECTION_REFUSED`，未能执行浏览器端断言；已通过 TypeScript 编译、E2E 治理校验和构建验证覆盖测试文件结构与动态路由构建。

审查记录：已按`lina-review`执行反馈级审查，范围包含本反馈新增/修改的插件`plugin.yaml`、`manifest/config/profile.yaml`、Go API/controller/service/test、前端挂载页、插件中英文 i18n、中文 apidoc、E2E、POM、根目录和`manifest/`中英文 README 以及本任务记录。未发现阻塞问题。规则域结论：插件目录规范通过，动态插件`manifest`服务保留显式`methods: [get]`和最小`resources.paths`授权；API 契约通过，新增只读接口使用`GET`且 DTO 文档源文本为英文；后端 Go 通过，未改生成 DAO/DO/Entity，新增 host service 依赖来源清晰且测试覆盖读取逻辑；前端 UI 通过，展示为单页面局部面板且无逐项补查；`i18n`通过，插件启用多语言，运行时语言包和非英文 apidoc 资源已同步；文档通过，中英文 README 镜像事实一致；测试策略通过，后端单测、E2E 文件、TS 编译和治理校验均覆盖，浏览器 E2E 剩余风险仅为当前本地服务未启动；数据权限、缓存一致性、SQL、开发工具跨平台和模块启停均无新增运行时风险。
