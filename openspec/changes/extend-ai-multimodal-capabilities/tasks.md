## 1. 变更准备和边界确认

- [x] 1.1 确认本变更依赖`add-ai-text-capability-center`和`refactor-ai-capability-namespace`，并在实施记录中说明两个活跃变更均已完成但尚未归档。
- [x] 1.2 修改`apps/lina-plugins/linapro-ai-core/`前检查插件根目录`AGENTS.md`，若存在则先读取并记录插件本地规范影响。
- [x] 1.3 读取并记录本变更命中的规则域：OpenSpec、文档、架构、插件、API、后端 Go、数据库、数据权限、缓存一致性、前端 UI、测试和`i18n`。
- [x] 1.4 明确记录本变更不触发开发工具跨平台入口变更；若实现期新增脚本或工具入口，则追加读取`.agents/rules/dev-tooling.md`并补充验证。

## 2. 宿主多模态 AI 能力契约

- [x] 2.1 在`apps/lina-core/pkg/plugin/capability/ai`下新增多模态公共值对象，包括`AssetRef`、`AssetResult`、`ProviderOperationRef`、`CapabilityType`、`CapabilityMethod`和通用状态投影。
- [x] 2.2 新增`Image()`、`Embedding()`、`Audio()`、`Vision()`、`Document()`、`Safety()`和`Video()`类型化子能力包与 fallback service，保持`AI().Text()`现有行为不变。
- [x] 2.3 为`image.generate`、`image.edit`、`embedding.create`、`audio.transcribe`、`audio.synthesize`、`vision.analyze`、`document.analyze`、`document.cite`、`safety.moderate`、`video.generate`、`video.edit`、`video.extend`、`video.operation.get`和`video.operation.cancel`定义 Go 命名类型、常量、DTO 和 provider 契约。
- [x] 2.4 明确`computer.act`、`ui.operate`和等价 UI 控制能力不在本轮宿主能力目录中暴露，并补充拒绝路径测试。
- [x] 2.5 通过静态检索确认未引入`AI().Invoke(...)`、`Generate(ctx, capabilityType, payload)`或等价弱类型网关。

## 3. 动态插件 AI Host Service 扩展

- [x] 3.1 扩展`pluginbridge`host service 常量、能力分类映射和清单校验，支持多模态`ai.*`方法并拒绝`computer.act`。
- [x] 3.2 为多模态 host service 定义 DTO 编解码和 payload 上限，确保大对象输入输出只通过`assetRef`或受控临时资产引用表达。
- [x] 3.3 扩展 WASM host service dispatcher，按 service、method、resource、purpose、mime 类型、资产数量、字节数、输出上限和 operation 权限完成调用前校验。
- [x] 3.4 扩展动态插件 guest SDK，使 guest 通过`AI().Image()`、`AI().Embedding()`、`AI().Audio()`、`AI().Vision()`、`AI().Document()`、`AI().Safety()`和`AI().Video()`调用对应 host service 方法。
- [x] 3.5 为未声明方法、未授权 resource、payload 超限、assetRef 不可见、operation 未授权和`computer.act`拒绝路径补充 host service 单元测试。

## 4. 智能中心数据库和 DAO 生成

- [x] 4.1 在`apps/lina-plugins/linapro-ai-core/manifest/sql/`新增当前迭代 SQL，建立 provider endpoint、model capability、method default params 和 provider operation 最小投影表。
- [x] 4.2 将 provider endpoint 表作为端点配置单一事实来源，provider 主表不得保留固定端点列、密钥引用列或面向旧结构的迁移回填逻辑；SQL 必须幂等，且不得显式写入自增`id`。
- [x] 4.3 为 provider、endpoint、model capability、tier、binding、invocation 和 operation 常用筛选、排序、关联和软删除路径补充索引。
- [x] 4.4 执行插件数据库初始化和 DAO 生成流程，确认生成结果位于`apps/lina-plugins/linapro-ai-core/backend/internal/dao`和`internal/model/{do,entity}`，不得手工创建或修改生成工件。
- [x] 4.5 记录 SQL 数据分类、幂等性、自增主键写入、软删除、索引和`N+1`规避验证结果。

## 5. 智能中心后端服务和 Provider Adapter

- [x] 5.1 扩展`linapro-ai-core`后端 API 和 service，支持 provider endpoint 管理、模型能力声明、能力方法档位查询/更新、默认参数和 operation 状态查询。
- [x] 5.2 确保智能中心管理 API 要求平台上下文和对应权限，供应商、endpoint、模型、档位、日志和 operation 查询不得向租户侧开放。
- [x] 5.3 扩展 provider adapter，优先实现文本既有行为无回归，并分批支持 embedding、vision/document、safety、image/audio 和 video operation 协议形态。
- [x] 5.4 实现 capability method 解析缓存，覆盖权威源、写后失效、集群模式同步、最大 30 秒兜底 TTL、缓存不可用降级和脱敏错误日志。
- [x] 5.5 确保调用日志只保存最小摘要、资产引用摘要、operation 摘要、用量、耗时和脱敏错误，不保存完整输入输出、文件内容、音视频内容或密钥。
- [x] 5.6 对新增 Controller、Service、provider adapter、host service handler 和缓存组件完成显式依赖注入检查，记录 owner、创建位置、传递路径和共享实例策略。

## 6. 智能中心前端和 i18n

- [x] 6.1 扩展供应商页面，支持 provider endpoint 管理、密钥脱敏、模型能力摘要和当前页批量投影渲染，不产生逐行详情补查。
- [x] 6.2 扩展档位管理页面，按`capabilityType + capabilityMethod`分组展示`basic`、`standard`、`advanced`，并按方法显示对应默认参数。
- [x] 6.3 扩展调用日志页面，支持多模态方法筛选、资产引用摘要、operation 摘要和脱敏错误详情。
- [x] 6.4 确认不新增视频任务、音频任务或其他业务任务页面；业务异步任务由调用方业务模块实现。
- [x] 6.5 根据`linapro-ai-core/plugin.yaml`的`i18n.enabled: true`维护插件运行时语言包、菜单文案、页面文案、错误文案和`apidoc`翻译资源。
- [x] 6.6 运行并记录`i18n`静态 key 覆盖和`apidoc`翻译完整性验证。

## 7. 测试覆盖

- [x] 7.1 为宿主多模态能力 fallback、状态、assetRef 边界、ProviderOperationRef 和`computer.act`拒绝路径补充 Go 单元测试。
- [x] 7.2 为动态插件多模态 host service 授权、payload 上限、资源策略、assetRef 可见性和脱敏审计补充 Go 单元测试。
- [x] 7.3 为`linapro-ai-core`provider endpoint、model capability、tier 解析缓存、日志脱敏和 operation 投影补充插件后端单元测试。
- [x] 7.4 创建插件 E2E`apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC004-smart-center-provider-endpoints.ts`，覆盖 provider endpoint、模型能力维护、i18n 文案和关键截图验证。
- [x] 7.5 创建插件 E2E`apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC005-smart-center-multimodal-tiers.ts`，覆盖能力方法筛选、三档绑定、默认参数校验、测试入口和错误路径。
- [x] 7.6 创建插件 E2E`apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC006-smart-center-multimodal-logs.ts`，覆盖多模态日志筛选、资产摘要、operation 摘要、脱敏错误和权限隐藏。
- [x] 7.7 保证新增 E2E 使用插件本地`pages/`和`support/`目录，测试可独立运行、可清理数据，并按截图验证要求把临时截图放到项目根目录`temp/`。

## 8. 验证和审查

- [x] 8.1 运行`openspec validate extend-ai-multimodal-capabilities --strict`并记录结果。
- [x] 8.2 运行覆盖宿主能力包、`pluginbridge`、WASM host service 和启动绑定的 Go 编译门禁。
- [x] 8.3 运行覆盖`linapro-ai-core`插件后端、DAO 使用包和源码插件注册路径的 Go 编译门禁。
- [x] 8.4 运行 SQL 初始化、DAO 生成、SQL 静态检索或等价治理验证，确认新 SQL 可重复执行。
- [x] 8.5 运行新增或更新的插件 E2E，并记录 TC004、TC005、TC006 的执行结果和截图审查结论。
- [x] 8.6 运行静态检索确认没有新增弱类型`AI`网关、没有`/api/ai/video-jobs`业务任务 API、没有大对象 base64 响应、没有明文 API key 返回。
- [x] 8.7 完成`lina-review`审查，覆盖活跃变更状态、任务完成状态、宿主边界、插件边界、API、SQL、数据权限、缓存一致性、`i18n`、测试和 E2E 质量审查。

## Feedback

- [x] **FB-1**: 修复`TC004`供应商端点按钮定位依赖 VXE 主表行子树导致超时的问题。
- [x] **FB-2**: 修复`TC005`多模态档位测试在能力方法筛选、档位操作或测试入口校验中超时的问题。
- [x] **FB-3**: 修复`TC006`调用日志能力方法下拉在 Ant Select 虚拟列表重渲染时出现选项 detached 或点击不稳定的问题。
- [x] **FB-4**: 修复`TC006`调用日志详情抽屉使用未显式注册的`a-descriptions`导致资产摘要、operation 摘要和元数据摘要标签未渲染的问题。
- [x] **FB-5**: 修复完整 Go 测试中源码插件 scoped services fixture 缺少`BizCtx`/`Cache`导致`linapro-ai-core`路由注册失败的问题。
- [x] **FB-6**: 修复`linapro-ai-core`SQL 治理缺少`CREATE TABLE`双语 Purpose 注释和 bare`ON CONFLICT DO NOTHING`的问题。
- [x] **FB-7**: 修复动态插件同版本重建后旧 hostServices 授权快照缺少新增`manifest`服务导致`TC005`页面 manifest demo 接口返回 500 的问题。
- [x] **FB-8**: 修复完整 E2E 中`linapro-org-core`部门 CRUD 与负责人选择测试在全量串行压力下超时和跨用例状态依赖的问题。
- [x] **FB-9**: 修复`linapro-monitor-server`的`TC003`依赖工作台菜单加载导致`beforeEach`超时的问题。
- [x] **FB-10**: 修复宿主`TC001a`系统接口页面首个 iframe 加载 Stoplight Elements 断言在完整 E2E 中失败的问题。
- [x] **FB-11**: 修复宿主菜单 E2E 将源码插件按钮文案和插件路由图标纳入宿主菜单治理断言导致全量运行失败的问题。
- [x] **FB-12**: 记录`linapro-tenant-core`租户插件 provisioning policy 在完整 E2E 中一次性失败的顺序污染诊断和定向验证结论。
- [x] **FB-13**: 修复完整 Go 测试中 plugin 包受共享`linapro-tenant-core`状态影响导致平台权限拒绝和 provider 快照断言失败的问题。
- [x] **FB-14**: 修复完整 Go 测试中`linapro-monitor-server`测试替身缺少新增`AI()`宿主能力方法导致编译失败的问题。
- [x] **FB-15**: 修复完整 E2E 中动态插件`TC005`依赖历史 hostServices 授权快照导致 manifest host service 缺少 capability 的问题。
- [x] **FB-16**: 修复宿主`TC004d`启动 Loading 字体断言在全量串行 E2E 中等待登录表单超时的问题。
- [x] **FB-17**: 删除`linapro-ai-core`中旧供应商固定端点列的历史兼容逻辑，改为 provider endpoint 表单一事实来源。
- [x] **FB-18**: 删除`plugin_linapro_ai_model`中的方法能力字段双源设计，改为 model capability 表单一事实来源。
- [x] **FB-19**: 合并`linapro-ai-core`插件安装 SQL，删除历史数据兼容迁移脚本。

## 实施记录

- 依赖状态：`add-ai-text-capability-center`和`refactor-ai-capability-namespace`仍为活跃 OpenSpec 变更；本次接续实施时已确认两者前置能力完成但尚未归档。
- 插件本地规范：`apps/lina-plugins/linapro-ai-core/AGENTS.md`不存在，本次插件目录变更按顶层`AGENTS.md`与命中规则执行。
- 规则读取：本次实施、反馈修复、验证和审查前已读取 OpenSpec、文档、架构、插件、API、后端 Go、数据库、数据权限、缓存一致性、前端 UI、测试、`i18n`和开发工具规则。
- 开发工具影响：未新增或修改`Makefile`、`make.cmd`、CI、脚本或默认测试/构建入口；仅新增和调整插件本地 E2E、POM 与支持 helper，不触发跨平台开发工具入口变更。
- SQL 与 DAO：`linapro-ai-core`插件安装基线收敛到`manifest/sql/001-linapro-ai-core-schema.sql`，数据分类为插件 DDL 与 Seed DML；脚本使用存在性保护、稳定业务键和`ON CONFLICT DO NOTHING`，未显式写入自增`id`，且不包含历史数据兼容迁移逻辑。DAO/DO/Entity 通过`make dao p=linapro-ai-core`生成，生成结果位于插件`backend/internal/dao`和`backend/internal/model/{do,entity}`。
- 性能与`N+1`：provider 列表按当前页 provider ID 批量装配 endpoint 和模型摘要；模型能力、档位绑定、调用日志与 provider operation 均使用数据库侧过滤、排序、分页或集合化查询；未引入前端逐行详情补查。
- 数据权限：智能中心 provider、endpoint、model、tier、日志和 provider operation 管理 API 均声明宿主权限标签，并在 service 层通过`ensurePlatform`限制平台上下文；日志首期不向租户侧开放。
- 缓存一致性：能力方法解析缓存权威源为插件数据库；provider、endpoint、model、model capability、tier、binding 和默认参数写入成功后触发失效；缓存使用插件 scoped 共享 revision、方法级 key 和最长 30 秒 TTL，缓存不可用时走数据库重建，数据库不可用时返回结构化不可用错误。
- DI 来源检查：`linapro-ai-core`智能中心 service owner 为源码插件；创建位置为`backend/plugin.go`的`smartCenter(...)`，依赖由宿主 registrar 提供的`BizCtx`和`Cache`逐项传入，并复用包级共享 service 与共享`http.Client`，保证管理写入和框架 provider 调用观察同一解析缓存。宿主多模态能力、guest SDK 与 WASM host service 沿既有 pluginbridge/hostservice 装配路径扩展，未绕过启动期共享实例。
- `i18n`：`linapro-ai-core/plugin.yaml`启用`i18n.enabled: true`；新增运行时文案、菜单文案、错误文案和`apidoc`翻译资源均维护在插件自己的`manifest/i18n`下。`make i18n.check`通过，存在既有 module-level `$t()` warning，未命中本次新增阻断问题。
- 业务任务边界：静态检索确认未新增`/api/ai/video-jobs`、`/api/ai/audio-jobs`或等价视频/音频业务任务页面/API；provider operation 仅作为供应商协议诊断投影。
- 静态检索：未发现新增`AI().Invoke(...)`、`Generate(ctx, capabilityType, payload)`或等价弱类型`AI`网关；未发现多模态 HTTP/host service 响应返回大对象 base64；provider 与 endpoint 响应走脱敏投影，不返回明文 API key。
- 反馈根因：`FB-1`为 VXE 固定操作列与主表行不在同一 DOM 子树；`FB-2`为 Ant Select 内部 input 点击被 selection item 拦截且非文本档位抽屉未按能力方法刷新字段显示；`FB-3`为虚拟列表 option 重渲染导致点击目标 detached；`FB-4`为插件动态页面未解析 kebab-case Ant Design `Descriptions`组件。
- Go 补跑反馈根因：`FB-5`为源码插件 scoped services 测试替身只提供了部分宿主能力，`linapro-ai-core`新增路由注册在启动检查中要求`BizCtx`和`Cache`均非空；`FB-6`为 SQL 治理扫描要求所有`CREATE TABLE`前具备相邻英文`Purpose:`与中文`用途：`注释，且新增 Seed DML 的 bare`ON CONFLICT DO NOTHING`无法声明业务唯一依据。
- Go 补跑影响分析：已读取 OpenSpec、文档、插件、后端 Go、数据库和测试规则；`apps/lina-plugins/linapro-ai-core/AGENTS.md`不存在。`i18n`无运行时文案、语言包或 API 文档源文本影响；缓存一致性仅涉及测试 no-op `Cache` fixture，不改变生产缓存权威源、失效或集群策略；数据权限无业务数据读写、租户边界或可见性变化；开发工具跨平台无脚本、CI 或工具入口变化。
- Go 补跑验证：日志位于`temp/extend-ai-multimodal-capabilities-fix-20260603-021237/`。`go test ./internal/service/plugin/internal/integration -run TestSourcePluginCallbacksUsePluginScopedServices -count=1`、`go test ./pkg/dialect -run 'TestOnConflictTargetsHaveDeclaredIdempotencyBasis|TestSQLCreateTablesHaveBilingualPurposeComments' -count=1`、对应两个完整包测试和`openspec validate extend-ai-multimodal-capabilities --strict`均通过。
- `linapro-demo-dynamic`补跑反馈根因：`FB-7`失败发生在动态插件页面请求`GET /x/linapro-demo-dynamic/api/v1/manifest-demo`时，响应为`host call error (status=1): plugin linapro-demo-dynamic lacks capability host:manifest`。trace 中插件清单已声明`manifest`和`hostconfig`，但 release 的`authorizedHostServices`仍沿用同版本旧 artifact 的授权快照，导致 runtime 派生能力缺少`host:manifest`。
- `linapro-demo-dynamic`补跑修复：`apps/lina-core/internal/service/plugin/internal/catalog/authorization.go`和`release.go`仅在新旧`requestedHostServices`规范化后完全一致时继承旧确认状态与授权服务；同版本 artifact 重建且声明变化时保留卸载清理选项，但不继承旧授权。`PersistReleaseHostServiceAuthorization`在安装/启用流程未显式传入授权输入且快照尚未确认时，按当前声明生成授权快照，避免新增 host service 被旧确认状态遮蔽。
- `linapro-demo-dynamic`补跑影响分析：已读取 OpenSpec、文档、架构、插件、接口契约、后端 Go、数据库、数据权限、缓存一致性、测试和`i18n`规则；`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，且本次未修改插件目录文件。`i18n`无运行时文案、语言包、插件清单或 API 文档源文本影响；缓存一致性无新增缓存或失效策略，仅改变持久化 release snapshot 的继承条件；数据权限无业务数据可见性、租户边界或读写路径变化；开发工具跨平台无脚本、CI 或工具入口变化；DI 来源无新增运行期依赖。
- `linapro-demo-dynamic`补跑验证：原完整 E2E 日志`temp/extend-ai-multimodal-capabilities-final-extra-20260603-020111/02-make-test-full-e2e.log`最终为`529 passed`、`10 failed`、`8 skipped`，其中保留本次修复前的`TC005`失败，后续失败分布在监控、组织、租户、API 文档、菜单治理等独立用例。重启当前工作区服务后，`pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC005-manifest-host-service-demo.ts --project=chromium --workers=1 --grep "TC-5a"`通过，结果为`1 passed`；`go test ./internal/service/plugin -run TestPersistDynamicAuthorizationRefreshesStaleSameVersionHostServices -count=1`、`go test ./internal/service/plugin/internal/catalog -count=1`和`go test ./internal/service/plugin -count=1`均通过。
- `linapro-monitor-server`补跑反馈根因：`FB-9`失败发生在`adminPage` fixture 预加载`/dashboard/analytics`时，trace 显示测试停留在“加载菜单中...”且在进入`TC003`正文前超时；监控页可见性轮询实现本身未触发失败。修复为使用不预加载工作台的`authenticatedPage`，并通过 API 确保`linapro-monitor-server`源码插件启用后直接进入`/monitor/server`。
- `linapro-monitor-server`补跑影响分析：已读取 OpenSpec、文档、插件、前端 UI、测试规则和`lina-e2e`规范；`apps/lina-plugins/linapro-monitor-server/AGENTS.md`不存在。仅调整插件自有 E2E 启动路径，不修改生产前端、后端、API、SQL、运行时文案或语言包；`i18n`、缓存一致性、数据权限、开发工具跨平台、模块接口契约和 DI 来源均无影响。
- `linapro-monitor-server`补跑验证：`E2E_RETRIES=0 pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-monitor-server/hack/tests/e2e/TC003-server-monitor-visibility-aware-polling.ts --project=chromium --workers=1`通过，结果为`1 passed`。
- `linapro-org-core`补跑反馈根因：`FB-8`失败集中在部门页面 E2E。`TC001c` trace 显示进入`/system/dept`后长期停留在 LinaPro 加载态，网络无 4xx/5xx，但 Vite 模块加载在约 71 秒后才继续，原 10 秒表格 ready 等待在完整串行压力下不足；`TC001d/e`依赖前序用例创建的文件级共享部门名称，worker 在失败后重启会重新计算后续名称，导致编辑和删除查找不存在的数据；`TC002d`失败发生在每个子用例重复同步并刷新源码插件投影的`beforeEach`阶段，完整 E2E 压力下放大超时风险。
- `linapro-org-core`补跑影响分析：已读取 OpenSpec、文档、插件、前端 UI、测试、`i18n`规则和`lina-e2e`规范；`apps/lina-plugins/linapro-org-core/AGENTS.md`不存在。仅调整插件自有 E2E 与部门 POM 的等待、数据隔离和清理逻辑，不修改生产前端、后端、API、SQL、运行时文案或语言包；`i18n`、缓存一致性、数据权限、开发工具跨平台、模块接口契约和 DI 来源均无影响。
- `linapro-org-core`补跑验证：`pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-org-core/hack/tests/e2e/dept/TC001-dept-crud.ts /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-org-core/hack/tests/e2e/dept/TC002-dept-leader-select.ts --project=chromium --workers=1`通过，结果为`9 passed`。
- 宿主 API Docs 补跑反馈根因：`FB-10`失败可在单独运行`TC001a`时复现，错误为`frame.getByText('ENDPOINTS')`触发 Playwright strict mode violation；Stoplight 页面实际已经渲染，失败截图显示左侧`ENDPOINTS`标题可见。新增多模态 API schema 名称包含`ListProviderEndpointsReq/Res`，模糊文本定位同时命中章节标题和 schema 项，导致首个用例失败。
- 宿主 API Docs 补跑影响分析：已读取 OpenSpec、文档、前端 UI、测试、接口契约、`i18n`规则和`lina-e2e`规范。仅收窄宿主`about`模块 E2E 断言到 Stoplight 侧边栏章节标题，不修改生产前端、后端、API、SQL、运行时文案、API 文档源文本或语言包；`i18n`、缓存一致性、数据权限、开发工具跨平台、模块接口契约和 DI 来源均无影响。
- 宿主 API Docs 补跑验证：`E2E_RETRIES=0 E2E_WORKERS=1 pnpm -C hack/tests exec playwright test hack/tests/e2e/about/TC001-api-docs-page.ts --project=chromium --workers=1 -g 'TC001a' --output ../../temp/extend-ai-multimodal-capabilities-api-docs-tc001a-20260603-fixed --reporter=list`通过，结果为`1 passed`；`E2E_RETRIES=0 E2E_WORKERS=1 pnpm -C hack/tests exec playwright test hack/tests/e2e/about/TC001-api-docs-page.ts --project=chromium --workers=1 --output ../../temp/extend-ai-multimodal-capabilities-api-docs-tc001-file-20260603-fixed --reporter=list`通过，结果为`10 passed`。
- 宿主菜单补跑反馈根因：`FB-11`失败可通过聚焦运行复现。`TC003`在英文菜单接口中对所有按钮菜单执行宿主短标题断言，完整 E2E 同步`linapro-ai-core`后命中插件自有按钮`Query Providers`；`TC002`对完整管理员路由树执行全局图标唯一性断言，源码插件路由进入菜单后与宿主图标产生合法重复。修复为将两个断言限定到宿主稳定菜单和宿主治理按钮权限范围，避免插件自有菜单资源污染宿主菜单治理回归。
- 宿主菜单补跑影响分析：已读取 OpenSpec、架构、前端 UI、测试、接口契约、后端 Go、数据权限、缓存一致性、`i18n`规则和`lina-e2e`规范。仅调整宿主 E2E 断言边界，不修改生产前端、后端、API、SQL、运行时文案、语言包或插件目录；`i18n`资源、缓存一致性、数据权限、开发工具跨平台、模块接口契约、DI 来源和数据库均无影响。
- 宿主菜单补跑验证：先用`E2E_RETRIES=0 E2E_WORKERS=1 pnpm -C hack/tests exec playwright test hack/tests/e2e/i18n/TC003-menu-governance-localized-titles.ts hack/tests/e2e/iam/menu/TC002-auth-menu.ts --project=chromium --workers=1 -g 'TC-3b|TC002a' --output ../../temp/host-menu-baseline-20260603-033829 --reporter=list`复现`2 failed`；修复后同范围命令通过，结果为`2 passed`；`pnpm -C hack/tests test:validate`通过，结果为`Validated 248 E2E test files across 17 scopes`。
- `linapro-tenant-core`补跑反馈根因：`FB-12`在第一轮完整 E2E 中表现为新建租户后`sys_plugin_state`中`linapro-org-core`的租户启用行数量为`0`，直接断言点位于`apps/lina-plugins/linapro-tenant-core/hack/tests/support/linapro-tenant-core-scenarios.ts`的 provisioning policy 场景。该场景依赖`sys_plugin`中目标插件同时处于已安装、已启用、`install_mode=tenant_scoped`且`auto_enable_for_new_tenants=true`的状态；独立、相邻顺序、目录片段和重复运行均无法复现，判断为第一轮完整 E2E 早期共享插件状态污染引发的一次性失败。
- `linapro-tenant-core`补跑影响分析：已读取 OpenSpec、插件、前端 UI、测试、数据权限、缓存一致性、`i18n`规则和`lina-e2e`规范；本轮未修改`apps/lina-plugins/linapro-tenant-core/`或其他生产/测试文件。`i18n`、缓存一致性、数据权限、开发工具跨平台、模块接口契约、DI 来源、API 和数据库均无新增影响。
- `linapro-tenant-core`补跑验证：`E2E_RETRIES=0 pnpm -C hack/tests exec playwright test apps/lina-plugins/linapro-tenant-core/hack/tests/e2e/plugin-governance/TC005-tenant-provisioning-policy.ts --project=chromium --workers=1`通过，结果为`1 passed`；相邻顺序`TC004-install-mode-migration.ts`加`TC005-tenant-provisioning-policy.ts`通过，结果为`2 passed`；`plugin-governance`目录`TC001`至`TC005`通过，结果为`7 passed`；`TC005`重复运行`--repeat-each=3`通过，结果为`3 passed`。当前数据库关键状态已恢复为`linapro-org-core|1|1|tenant_aware|global|false`、`linapro-tenant-core|1|1|platform_only|global|false`且`org_state_rows=0`。
- 第二轮完整 Go 补跑反馈根因：`FB-13`失败发生在`make test.go`的`lina-core/internal/service/plugin`包。前序 Go 包和官方插件装配会让共享测试库中`linapro-tenant-core`处于已安装启用状态，`newTestService()`默认注入真实`tenantcap`作为平台治理 guard 后，使用`context.Background()`的 plugin 单元测试不再具备`PlatformBypass`，因此 63 个安装、启用、同步、升级路径在真实业务断言前统一被`Platform permission is required`短路。`TestSourceProviderAvailabilityFollowsEnabledSnapshot`同根因受共享官方 tenant provider 影响，在测试插件仍 disabled 时观测到`linapro-tenant-core`作为 active provider。
- 第二轮完整 Go 补跑修复：`apps/lina-core/internal/service/plugin/plugin_test.go`的根测试服务不再默认注入平台治理 guard，保留租户启动一致性和自动 provisioning 能力；平台权限拒绝语义继续由`plugin_platform_guard_test.go`显式注入真实 tenant capability 覆盖。`apps/lina-core/internal/service/plugin/plugin_capability_revision_test.go`改用仅暴露本用例 provider 的 runtime 视图，避免共享官方 provider 状态污染 provider enablement snapshot 断言。未修改`PersistReleaseHostServiceAuthorization`、`authorization.go`或`release.go`，上一轮 stale same-version hostServices 授权修复保持不变。
- 第二轮完整 Go 补跑影响分析：已读取 OpenSpec、文档、架构、插件、后端 Go、数据权限、缓存一致性、测试和`i18n`规则，并使用`goframe-v2`技能。仅修改`apps/lina-core/internal/service/plugin`测试夹具和测试断言隔离，不修改生产代码、API、SQL、插件目录、运行时文案或语言包；`i18n`无影响，缓存一致性无新增缓存或失效路径，数据权限无业务数据可见性或租户边界变化，开发工具跨平台无脚本、CI 或工具入口变化，DI 来源无新增运行期依赖。
- 第二轮完整 Go 补跑验证：`go test -race ./internal/service/plugin -run 'TestBootstrapAutoEnableReusesDynamicAuthorizationSnapshot|TestSourceProviderAvailabilityFollowsEnabledSnapshot|TestInstallBlocksUninstalledSourceDependency|TestExecuteRuntimeUpgradeRequiresConfirmation|TestUpdateTenantProvisioningPolicySurvivesManifestSync|TestSingleNodeModeSkipsPluginNodeProjection' -count=1 -v`通过；`go test ./internal/service/plugin -run TestPersistDynamicAuthorizationRefreshesStaleSameVersionHostServices -count=1 -v`通过，确认 TC005 相关动态 hostServices stale 授权修复仍成立；`go test ./internal/service/plugin -count=1`通过；`go test -race ./internal/service/plugin -count=1`通过；`openspec validate extend-ai-multimodal-capabilities --strict`通过。
- 第三轮完整 Go 补跑反馈根因：`FB-14`失败发生在`apps/lina-plugins/linapro-monitor-server/backend/plugin_test.go`的`fakeCapabilities`测试替身。宿主`pluginhost.Services`经`capability.Services`新增`AI()`能力入口后，该测试替身仍只实现旧接口；`cleanupSnapshots`接收`pluginhost.Services`时触发编译期接口检查，报告缺少`AI()`方法。
- 第三轮完整 Go 补跑修复：仅在`linapro-monitor-server`插件后端测试替身中补充`AI()`方法，返回`capabilityai.New(nil)`默认 fallback namespace，保持测试 no-op 行为，不引入真实 AI provider、生产服务初始化或运行期依赖变化。`apps/lina-plugins/linapro-monitor-server/AGENTS.md`不存在，本次按顶层规范执行。
- 第三轮完整 Go 补跑影响分析：已读取 OpenSpec、文档、插件、后端 Go、测试和`i18n`规则，并使用`goframe-v2`技能。仅修改插件测试替身和 OpenSpec 反馈记录，不修改生产代码、API、SQL、前端、运行时文案、插件清单或语言包；`i18n`无影响，缓存一致性无新增缓存或失效路径，数据权限无业务数据读写或可见性变化，开发工具跨平台无脚本、CI 或工具入口变化，DI 来源无新增运行期依赖，E2E 质量审查未触发。
- 第三轮完整 Go 补跑验证：`GOWORK=off go test ./linapro-monitor-server/backend -count=1`因`apps/lina-plugins`聚合模块不包含该相对包路径而失败；等价命令`cd apps/lina-plugins && GOWORK=off go test lina-plugin-linapro-monitor-server/backend -count=1`通过，结果为`ok lina-plugin-linapro-monitor-server/backend 0.605s`。`openspec validate extend-ai-multimodal-capabilities --strict`通过。
- `FB-15`补跑反馈根因：完整 E2E 旧 trace 中动态插件页面请求`GET /x/linapro-demo-dynamic/api/v1/manifest-demo`返回`plugin linapro-demo-dynamic lacks capability host:manifest`。生产授权继承逻辑已避免同版本重建沿用旧`authorizedHostServices`快照，但该 E2E 仍通过默认安装/启用 helper 触发历史授权状态，未显式提交当前动态插件所需的`manifest`和`hostConfig`方法授权，导致测试在共享环境中不能稳定证明 manifest host service 修复路径。
- `FB-15`补跑修复：`apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC005-manifest-host-service-demo.ts`为动态插件安装和启用请求显式传入`storage`、`network`、`data`、`manifest.get`和`hostConfig.get`授权快照；源码插件依赖仍沿用原 helper，测试恢复逻辑按原始 installed/enabled 状态收敛。
- `FB-15`影响分析：已读取 OpenSpec、文档、插件、前端 UI、测试、`i18n`、缓存一致性、数据权限和开发工具规则，并使用`lina-e2e`规范；`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在。仅修改动态插件自有 E2E 测试，不修改生产前端、后端、API、SQL、运行时文案、插件清单或语言包；`i18n`无运行时文案或翻译资源影响，缓存一致性无新增缓存或失效策略，数据权限无业务数据可见性或租户边界变化，开发工具跨平台无脚本、CI 或工具入口变化，DI 来源无新增运行期依赖。
- `FB-15`补跑验证：历史失败文件定向重跑`temp/extend-ai-multimodal-capabilities-targeted-20260603/current-failed-files-rerun.log`通过，结果为`69 passed`；当前工作区重新运行`pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC005-manifest-host-service-demo.ts --project=chromium --workers=1 --reporter=list`通过，结果为`1 passed`。
- `FB-16`根因：失败 trace 显示`TC004d`在拦截并释放`/admin/src/main.ts`后，`captureLoadingTitleFontOnRefresh`已返回，但随后`/admin/src/bootstrap.ts?t=...`动态导入请求保持未完成状态，页面停留在`#__app-loading__`启动 Loading，测试继续等待用户名输入框直到 180 秒超时。该断言只需要比较刷新时启动 Loading 字体与应用字体；应用字体可以在刷新拦截前的干净登录页完成加载后采样，不应依赖人为延迟启动链路后的页面恢复。
- `FB-16`修复：仅调整宿主`hack/tests/e2e/settings/config/TC004-public-frontend-config.ts`，在`TC004d`执行刷新拦截前先读取应用根字体，刷新拦截只负责捕获启动 Loading 标题字体，移除对拦截后登录表单重新出现的等待。
- `FB-16`影响分析：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/frontend-ui.md`和`.agents/rules/testing.md`，并使用`lina-feedback`与`lina-e2e`规范。仅修改宿主 E2E 和 OpenSpec 反馈记录，不修改生产前端、后端、API、SQL、运行时文案、语言包或插件目录；`i18n`无运行时文案、API 文档源文本或翻译资源影响，缓存一致性无新增缓存或失效路径，数据权限无业务数据读写或可见性变化，开发工具跨平台无脚本、CI 或工具入口变化，DI 来源无新增运行期依赖。
- `FB-16`补跑验证：日志位于`temp/extend-ai-multimodal-capabilities-fix-tc004d-20260603-101514/`。`E2E_RETRIES=0 E2E_WORKERS=1 pnpm -C hack/tests exec playwright test hack/tests/e2e/settings/config/TC004-public-frontend-config.ts --project=chromium --workers=1 -g 'TC004d' --output ../../temp/extend-ai-multimodal-capabilities-fix-tc004d-20260603-101514/test-results --reporter=list`通过，结果为`1 passed`；同文件完整运行`E2E_RETRIES=0 E2E_WORKERS=1 pnpm -C hack/tests exec playwright test hack/tests/e2e/settings/config/TC004-public-frontend-config.ts --project=chromium --workers=1 --output ../../temp/extend-ai-multimodal-capabilities-fix-tc004d-20260603-101514/test-results-file --reporter=list`通过，结果为`6 passed`；`pnpm -C hack/tests test:validate`通过，结果为`Validated 248 E2E test files across 17 scopes`；`openspec validate extend-ai-multimodal-capabilities --strict`通过。
- `FB-17`根因：`linapro-ai-core`在多模态扩展中已经引入 provider endpoint 表，但 SQL、API、service、前端和测试仍保留旧供应商固定端点/密钥字段、从旧字段回填 endpoint 的迁移式逻辑，以及从旧模型表派生 model capability 的回填 SQL；这与项目不考虑历史兼容性的绿色地要求冲突。
- `FB-17`修复：provider 主表收敛为供应商元数据；协议端点、基础地址和密钥引用统一由 provider endpoint 表承载；模型创建和更新必须显式传入有效`endpointId`，基础模型表直接包含`endpoint_id`；模型同步、档位绑定、provider adapter 和调用解析均读取 endpoint 表；删除旧字段相关 DTO、表单、语言包、测试 helper 和旧模型能力回填 SQL，并重新生成插件 DAO/DO/Entity。OpenSpec 提案、设计、增量规格和任务记录同步改为“endpoint 表是单一事实来源”，不再描述旧字段迁移。
- `FB-17`E2E 修复：供应商操作列在`VXE`固定右列中会出现父级`td`拦截按钮点击，插件 POM 改为用主表行`rowid`定位固定右侧目标行；仅在固定列按钮被父级`td`拦截时对目标按钮执行受限 fallback click。同步按钮断言改为验证目标行内按钮存在，不再验证无业务意义的视觉垂直顺序。
- `FB-17`影响分析：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`和`.agents/rules/dev-tooling.md`，并使用`lina-feedback`、`lina-review`、`goframe-v2`、`lina-e2e`和`karpathy-guidelines`规范。`apps/lina-plugins/linapro-ai-core/AGENTS.md`不存在。`i18n`影响为删除旧 provider 字段文案并维护 endpoint 文案和 API 文档翻译；SQL/DAO/API/前端/E2E 均受影响；数据权限仍为平台管理面权限和`ensurePlatform`边界，无租户侧开放；缓存沿既有 provider/endpoint/model/tier 写后失效和方法级解析缓存，无新增缓存类型；开发工具跨平台无脚本、CI 或默认入口变化；DI 来源无新增运行期依赖。
- `FB-17`验证：`GOWORK=off go test ./... -count=1`在`apps/lina-plugins/linapro-ai-core`通过；`pnpm -C apps/lina-vben/apps/web-antd typecheck`通过；`make i18n.check`通过，仅保留既有 module-level`$t()`警告且不命中本次`linapro-ai-core`变更；`pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-ai-core/hack/tests/e2e/TC004-smart-center-provider-endpoints.ts --project=chromium --workers=1 --reporter=list`通过，结果为`1 passed`；`pnpm -C hack/tests test:module -- plugin:linapro-ai-core`通过，结果为`11 passed`；`pnpm -C hack/tests test:validate`通过，结果为`Validated 248 E2E test files across 17 scopes`；`openspec validate extend-ai-multimodal-capabilities --strict`通过；`git diff --check`和`git -C apps/lina-plugins diff --check`通过；兼容残留扫描确认插件目录无旧固定端点字段、旧 provider 密钥字段、旧模型能力回填 SQL、compatibility fallback 或 fallback endpoint 残留。
- `FB-18`根因：`plugin_linapro_ai_model`原本承载模型身份和`text.generate`能力字段，多模态扩展新增`plugin_linapro_ai_model_capability`后未完全移除模型表上的`capability_type`、`capability_method`、token 上限和 thinking effort 字段，形成模型表与 capability 表双事实源，导致模型更新、档位绑定和候选列表可能读取不同权威。
- `FB-18`修复：模型基础表收敛为 provider、endpoint、model name、protocol、source、enabled 和时间字段；`plugin_linapro_ai_model_capability`补齐`supports_thinking`与`supported_efforts`并作为能力方法、token 上限、thinking effort 和 endpoint override 的唯一事实来源。模型创建只在事务内创建模型身份和一条初始 capability；模型更新只修改身份字段。供应商模型列表、供应商摘要、档位绑定校验、档位缓存解析和测试调用均改为从 capability 表集合化装配，不再读取模型表能力字段。前端模型创建移除二次保存 capability 请求，API 文档和插件 apidoc 翻译同步更新。
- `FB-18`影响分析：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`和`.agents/rules/dev-tooling.md`，并使用`lina-feedback`、`lina-review`和`goframe-v2`规范。`apps/lina-plugins/linapro-ai-core/AGENTS.md`不存在。`i18n`影响为 API 文档源文本和插件`zh-CN/apidoc`翻译更新；数据权限仍为平台管理 API 权限和`ensurePlatform`边界，无租户侧开放；缓存沿既有模型、capability、tier 写后失效和共享 revision，不新增缓存类型；开发工具跨平台无脚本、CI 或工具入口变化；DI 来源无新增运行期依赖；前端用户可观察变化为创建模型从双请求收敛为单请求。
- `FB-18`验证：`GOWORK=off go test ./backend/internal/service/ai -run 'TestProviderModelTierAndInvocationFlow|TestMultimodalMetadataManagementFlow|TestTierBindingRequiresMatchingCapabilityMethod' -count=1 -v`实际执行数据库用例并通过；`GOWORK=off go test ./backend/... -count=1`在`apps/lina-plugins/linapro-ai-core`通过；`go test ./pkg/dialect -count=1`在`apps/lina-core`通过；`pnpm -F @lina/web-antd typecheck`和`pnpm -F @lina/web-antd i18n:check`在`apps/lina-vben`通过；`openspec validate extend-ai-multimodal-capabilities --strict`通过；静态扫描确认`plugin_linapro_ai_model`生成的 DAO/DO/Entity 不再包含能力方法、token 或 thinking effort 字段，业务路径的能力方法读取集中到 capability 表。
- `FB-19`根因：当前项目不考虑历史数据兼容，且用户明确允许清理数据；继续保留`002-extend-ai-multimodal-capabilities.sql`中的旧表回填、旧列删除和增量补列逻辑，会让全新插件安装基线携带不必要的历史迁移语义。
- `FB-19`修复：将 provider endpoint、model capability、method default params、provider operation、invocation 多模态摘要字段、多模态档位 seed 和方法默认参数 seed 合并到`manifest/sql/001-linapro-ai-core-schema.sql`；删除`manifest/sql/002-extend-ai-multimodal-capabilities.sql`；同步补全`manifest/sql/uninstall/001-linapro-ai-core-schema.sql`，确保卸载清理所有最终表。
- `FB-19`影响分析：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`，并使用`lina-feedback`规范。`apps/lina-plugins/linapro-ai-core/AGENTS.md`不存在。本次仅调整插件安装/卸载 SQL 与 OpenSpec 记录，不修改 Go 生产代码、API DTO、前端页面、运行时文案、插件清单或语言包；`i18n`无影响，缓存一致性无影响，数据权限无新增数据访问路径或租户边界变化，开发工具跨平台无脚本、CI 或工具入口变化，DI 来源无新增运行期依赖，E2E 质量审查不触发。
- `FB-19`验证：静态扫描确认`linapro-ai-core/manifest/sql`不再包含`DO $$`、`information_schema`、`DROP COLUMN`、`ALTER TABLE ... ADD COLUMN IF NOT EXISTS`、`DROP INDEX`或`002-extend-ai-multimodal-capabilities`残留；`go test ./pkg/dialect -run 'TestOnConflictTargetsHaveDeclaredIdempotencyBasis|TestSQLCreateTablesHaveBilingualPurposeComments' -count=1`和`go test ./pkg/dialect -count=1`在`apps/lina-core`通过；`GOWORK=off go test ./backend/... -count=1`在`apps/lina-plugins/linapro-ai-core`通过；`openspec validate extend-ai-multimodal-capabilities --strict`通过。
- 最终验证：完整验证日志位于`temp/extend-ai-multimodal-capabilities-final-20260603-014508/`。`openspec validate`、宿主 Go 测试、插件 Go 测试、`make i18n.check`、前端构建、`hack/tests test:validate`和`pnpm -C hack/tests test:module -- plugin:linapro-ai-core`全部通过，插件 E2E 结果为`11 passed`。TC004、TC005、TC006 聚焦组合验证为`3 passed`，截图已按规则写入项目根`temp/`目录。
