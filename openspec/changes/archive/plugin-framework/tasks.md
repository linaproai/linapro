# Tasks

## Summary

- [x] 建立插件框架核心能力：统一`plugin.yaml`、源码/动态插件生命周期、动态`WASM`运行时、host service、能力目录、插件 UI、菜单权限、启动引导、依赖、升级、工作区和包边界。
- [x] 建立插件宿主领域能力模型：`usercap`、`authzcap`、`dictcap`、`filecap`、`sessioncap`、`configcap`、`notifycap`、`plugincap`、`jobcap`、`infracap`，重整`orgcap`、`tenantcap`和`ai`；清零插件生产代码对宿主`sys_*`核心表的直接依赖；动态`data`服务收窄为当前插件自有表。
- [x] 统一能力组件为`*cap`命名体系：删除`capability/contract`聚合包；认证授权收敛到`authcap`能力族；插件自能力收口到`plugincap`子领域；租户过滤归属`tenantcap`；动态插件数据 SDK 收口到`recordstore`。
- [x] 统一插件资源读取为`Manifest()`：删除旧`Metadata`服务语义；`Manifest()`覆盖`manifest/`下完整资源包括`config/`、`sql/`、`i18n/`；动态 artifact 资源视图完整投影；动态插件构建器完整收集 manifest 资源。
- [x] 移除源码插件宿主配置读取 key 白名单限制，保持动态插件授权快照约束。
- [x] 优化插件运行时复杂度：runtime state 列表批量读取 manifest；WASM host service 依赖显式注入；runtime wiring 必需依赖构造阶段校验；host service 协议单一描述源和覆盖测试。
- [x] 优化插件管理首次加载：列表拆分为分页摘要 DTO 和详情 DTO；cron 和依赖检查移出列表路径；前端弹窗按需异步加载；缓存复用`plugin-runtime`协调。
- [x] 整理插件服务内部文件组织：合并同职责低行数文件；测试按被测职责拆分；helper 收敛到支撑文件。
- [x] 修复和收敛关键演进：源码插件自动安装后同步启动快照；旧`Can*`/guard 生命周期替换为`Before*`/`After*`；动态生命周期由构建期自动发现；插件 API 与资产路径分别收敛到`/x/{plugin-id}/api/v1/...`和`/x-assets/{plugin-id}/{version}/...`。
- [x] 治理：插件公共契约收敛到`pkg/plugin`；宿主私有实现收敛到`internal/service/plugin`；能力目录不暴露`DAO`、`DO`、`Entity`、`*gdb.Model`、`*ghttp.Request`、写入路径或数据权限注入能力。
- [x] 性能：插件列表查询保持只读且使用分页摘要；完整治理读模型可预热并按插件生命周期、动态产物和租户供应策略显式失效；列表装配复用快照并避免逐插件重复扫描。
- [x] 一致性：插件运行时、frontend bundle、runtime i18n、WASM、manifest 资源视图和默认配置视图按插件和资源作用域失效；集群模式通过 coordination revision/event 和 per-plugin 锁收敛。

## 包边界与依赖方向治理

- [x] 修正`pkg/plugin`依赖方向：`capability`成为最底层契约层；`ManifestSnapshotV1`迁入`capmodel`；recordstore SDK 迁移到`pluginbridge/recordstore`；import 边界治理测试持续验证。
- [x] 将 host service 领域桥接收敛为公开 protocol catalog、JSON envelope 和 registry dispatch 三个接缝；descriptor 从公开 catalog 派生；WASM dispatch 消除 service 级 switch。
- [x] 收敛领域能力边界：宿主实现重命名为`capabilityhost`；动态领域只保留一个`ConfigureDomainHostServices`入口；集合型领域 service 名统一为复数；插件配置归属`plugins`；通知归属`notifications`；定时任务归属`jobs`。
- [x] 拆分消费契约与 Provider SPI：`tenantspi`/`orgspi`子包承载 provider 和 scope 接缝；`routecap`/`apidoccap`去除`ghttp`泄漏；provider 声明收敛到`pluginhost.Declarations`；provider manager 由宿主显式持有。
- [x] 简化动态插件 AI 授权：`ai`从`purpose`资源授权改为`service + method`方法授权；普通领域 host service 覆盖源码插件领域能力。
- [x] 统一 host service 单一事实源：descriptor 驱动 README 表格生成和双向覆盖治理；guest client 统一为注入式 domainhostcall 单轨。
- [x] 将 Cache/Lock/Storage 统一为领域能力：源码插件和动态插件共享`cachecap.Service`/`lockcap.Service`/`storagecap.Service`；Storage 新增 provider 扩展机制和默认本地磁盘 provider。

## Feedback

- [x] **FB-1**: 将`wasm`领域 host service 分发文件按领域能力拆分，评估是否应作为独立组件维护。根因：单文件聚合跨领域分发。修复：拆分为`hostfn_service_<service>.go`。验证：Go 测试通过。不建议提取为子包，因 dispatcher 依赖 wasm 包内可信执行上下文。
- [x] **FB-2**: 统一`wasm`领域 host service 源文件命名，移除额外`domain`前缀。根因：FB-1 拆分后命名不一致。修复：统一为`hostfn_service_<service>.go`。验证：Go 测试和静态检索通过。
- [x] **FB-3**: 统一集合型领域协议字符串为复数形式。根因：`capability.Services`使用复数但协议使用单数。修复：service/capability/常量/文件名统一为`users`/`files`/`jobs`/`notifications`/`plugins`/`sessions`。验证：Go 测试、启动装配和 OpenSpec 校验通过。
- [x] **FB-4**: 将`Plugins().Lifecycle()`作为受治理插件领域能力暴露，收敛`guest.PluginService`特例。根因：`guest.Services.Plugins()`返回公共特例接口。修复：改为返回`plugincap.Service`，删除公共`PluginService`接口。验证：hostservice validation、descriptor、guest stub、WASM dispatcher 测试通过。
- [x] **FB-5**: 删除无业务入口的`file.Service` Markdown 内容读写方法。根因：编译阻断补齐实现扩大了核心合约。修复：删除未发布方法和错误码。验证：Go 测试和静态检索通过。
- [x] **FB-6**: 将插件配置读取从独立`config`收敛到`plugins.config.get`。根因：配置属于插件治理领域但独立暴露。修复：删除`HostServiceConfig`，授权收敛为`service: plugins`和`method: config.get`。验证：descriptor、guest、WASM 和动态示例测试通过。
- [x] **FB-7**: 将通知发送从独立`notify`收敛到`notifications.messages.send`。根因：通知读取和发送分属两个相近领域入口。修复：删除`service: notify`，新增`messages.send`按渠道资源授权。验证：WASM 通知发送测试通过。
- [x] **FB-8/9**: 删除动态插件独立 cron host service，定时任务统一归属`jobs`领域。根因：动态 cron 和运行时 Jobs 混淆。修复：删除 cron host service、发现期 host-call 和 guest SDK 入口；动态插件内置任务改用`jobs.register`发现期声明。验证：Go 测试、E2E 反向断言通过。
- [x] **FB-10**: 将源码插件定时任务注册入口从`Cron`迁移到`Jobs`。根因：源码插件仍保留`Cron`公开契约。修复：统一为`pluginhost.Jobs()`和`JobsRegistrar`。验证：源码插件后端测试通过。
- [x] **FB-11**: 将旧动态插件 cron 声明能力迁移为`Jobs`领域的动态任务声明能力。根因：动态插件声明能力不能丢失。修复：新增`jobs.register`协议方法和 guest adapter；前端授权提交改为包含无资源目标的 service。验证：Go 测试、前端构建和 E2E 通过。
- [x] **FB-12**: 将受治理运行时配置管理收敛到`hostconfigcap.AdminService`。根因：配置管理独立位于`configcap`。修复：删除`configcap`，迁移到`hostconfigcap`。验证：Go 测试和 OpenSpec 校验通过。
- [x] **FB-13**: 新增`DynamicPlugin`声明期契约并与运行时领域能力分离。根因：声明期能力与运行时能力混合。修复：统一为`pluginbridge.Declarations`，运行时只保留普通领域能力。验证：guest、WASM builder 和动态示例测试通过。
- [x] **FB-14**: 将动态插件公开入口从`pluginbridge/guest`收敛到`pluginbridge`。根因：入口命名不对称。修复：迁移 guest SDK 到根包，删除公开 guest 目录。验证：Go 测试和 OpenSpec 校验通过。
- [x] **FB-15**: 将 host service payload codec 所有权从内部`hostservice`迁移到公共`protocol`。根因：protocol 和 internal hostservice 所有权混在一起。修复：迁移 payload struct 和 codec 实现到`protocol`。验证：Go 测试和 OpenSpec 校验通过。
- [x] **FB-16/17/18**: 统一声明期入口命名。根因：`DynamicPlugin`和`SourcePlugin`命名模糊。修复：统一为`Declarations`和`*Declarations`子接口。验证：Go 测试和静态检索通过。
- [x] **FB-19**: 为 WASM `org`和`tenant`用户作用域 host service 补齐目标用户可见性校验。根因：用户作用域方法直接信任 payload 中的任意`userID`。修复：新增`ensureHostCallUsersVisible`前置校验。验证：Go 测试和 race 测试通过。
- [x] **FB-20**: 将 WASM host-call 错误响应载荷结构化。根因：失败时直接写入裸字符串。修复：新增`HostCallErrorPayload`，保留`bizerr`元数据。验证：协议测试通过。
- [x] **FB-21**: 将 WASM host service 运行期依赖收敛为并发安全的显式快照读取。根因：领域服务分散在多个可变包级变量中。修复：新增`hostServiceRuntime`不可变快照和`atomic.Pointer`发布机制。验证：Go 测试和 race 测试通过。
- [x] **FB-22**: 修复 WASM host service race 测试替身的共享状态竞争。根因：测试替身共享父对象写入。修复：改为互斥锁保护的 recorder。验证：race 测试通过。

- [x] FB-1: 收敛`Services`公开领域能力目录，移除通知、会话和配置领域的重复公开入口。根因：`capability.Services`同时公开旧入口和新领域入口。修复：删除旧`contract.NotifyService`与`contract.SessionService`，统一使用`Services.Notifications()`、`Services.Sessions()`、`Services.Plugins().Config()`和`Services.HostConfig()`。验证：Go 编译门禁、静态检索和 OpenSpec 严格校验通过。
- [x] FB-2: 通用化插件规范检查入口并移除 Go 语法已阻断的宿主`DAO/DO/Entity`扫描。根因：原扫描器将 Go `internal`目录天然阻断的导入也纳入规则。修复：删除重复规则，公开命令改为`linactl plugins.check`。验证：`go test ./hack/tools/linactl/...`通过。
- [x] FB-3: 公开动态插件语言无关`hostServices`协议目录。根因：开发者无法查看可声明服务、方法与资源边界。修复：新增`pluginbridge/protocol/hostservice_catalog.go`公开协议描述。验证：协议测试通过。
- [x] FB-4: 将动态插件 guest 能力 SDK 合并到`pluginbridge/guest`。根因：`capability/guest`和`pluginbridge/guest`两个包名都叫`guest`造成混淆。修复：删除`capability/guest`目录，合并到`pluginbridge/guest`。验证：Go 测试、治理扫描和`wasip1`构建通过。
- [x] FB-5: 将认证 token 与授权能力收敛到`authcap`能力族子领域。根因：旧结构在根`capability.Services`同时暴露`Auth()`和`Authz()`。修复：token 契约迁移到`authcap/token`，授权迁移到`authcap/authz`，根只暴露`Auth()`。验证：Go 测试和静态检索通过。
- [x] FB-6: 将动态插件受治理数据 SDK 从`capability/data`重命名为`capability/recordstore`。根因：`data`命名过宽。修复：迁移为`recordstore`，Go 入口改为`RecordStore()`。验证：Go 测试和 OpenSpec 严格校验通过。
- [x] FB-7: 将过轻`pkg/authtoken`公共包合并到`authcap/token`子领域。根因：单独保留形成过轻公共包。修复：合并到`authcap/token`。验证：Go 测试和静态检索通过。
- [x] FB-8: 将`hostservices_domain_adapters.go`中的领域适配器实现拆分到各自`*cap`内部组件。根因：单文件聚合跨领域实现。修复：拆分为`hostservices/internal/{authzcap,configcap,...}`组件包。验证：Go 测试通过。
- [x] FB-9: 修复根插件服务测试文件缺少对应生产源码文件的问题。根因：测试命名与生产文件不对应。修复：新增对应生产文件，测试归位。验证：Go 测试通过。
- [x] FB-10: `linapro-demo-dynamic`缺少`service: manifest`配置示例。根因：动态插件示例未展示 manifest 授权写法。修复：更新`plugin.yaml`和文档。验证：YAML 解析、OpenSpec 校验和构建测试通过。
- [x] FB-11: `linapro-demo-dynamic`的 manifest 示例授权过宽。根因：示例授权范围过宽且缺少声明到使用的闭环。修复：收敛授权路径，新增演示接口和页面。验证：Go 测试、TypeScript 编译和 E2E 治理校验通过。
- [x] FB-12: Nightly plugin-full E2E 复用旧动态插件 active release 导致 manifest 配置预览缺失。根因：E2E 未强制重新安装当前构建产物。修复：测试前强制卸载旧 active release 后重新安装。验证：TypeScript 编译和 E2E 治理校验通过。
- [x] FB-13: GitHub Actions clean checkout 缺少被忽略的动态插件 manifest 配置 fixture。根因：`config.yaml`被`.gitignore`忽略。修复：测试前从`config.example.yaml`临时复制生成 fixture。验证：Playwright 测试通过。
- [x] FB-14: 源码插件读取宿主配置被公开 key 白名单错误限制。根因：`valueForKey`通过固定 switch 只允许三个 key。修复：移除固定白名单，按调用方 key 直接读取。验证：Go 测试和 OpenSpec 校验通过。
- [x] FB-15: 动态插件构建器仍沿用旧的 manifest 资源过滤逻辑。根因：`collectManifestResources`只打包少量 YAML 资源。修复：完整收集`manifest/`下实际文件。验证：构建器测试和产物静态检查通过。

## Verification

- [x] `openspec validate`严格校验：所有迭代均通过。
- [x] Go 编译门禁：覆盖`apps/lina-core/pkg/plugin/capability`、`pluginhost`、`pluginbridge`、受影响宿主领域 service、动态 host service 和所有迁移插件后端包。
- [x] 后端单元测试：领域能力单元测试、动态协议测试、`data`服务授权测试、缓存一致性测试、治理扫描测试、import 边界治理测试、descriptor 双向覆盖测试、registry dispatch 测试、WASM host service race 测试。
- [x] 前端验证：TypeScript 类型检查、`pnpm i18n:check`通过。
- [x] E2E：完整 E2E 合计 582 passed、8 skipped、0 failed。
- [x] 静态检索：旧`contract.*Service`、旧非`*cap`包、旧`capability/ai`、旧`Metadata`服务、宿主核心表 DAO 生产导入、旧`capability/recordstore`路径、旧单数集合型 service 名、旧`Configure*HostService`专用入口、旧`pluginbridge/guest`目录、旧`configcap`公开包、旧`Cron`公开契约、旧`DynamicPlugin`/`SourcePlugin`声明期命名均无残留。
- [x] `git diff --check`：所有迭代均通过。
- [x] 动态插件样例构建：`GOWORK=off go test`和`linactl wasm`wasip1 构建均通过。
- [x] 治理测试正反向验证：临时构造违规 import 确认测试能捕获，验证后删除临时文件。

## Governance

- [x] i18n：领域能力返回稳定值和`labelKey`；`make i18n.check`通过；插件运行时文案按宿主和插件 i18n 启用边界维护。后续变更仅涉及技术文档和协议标识符，无运行时用户可见文案或语言包新增。
- [x] 缓存一致性：权限、角色关系、租户成员、插件状态、字典、组织树、运行时配置等关键数据使用共享修订号和事务后失效；单机和集群模式均有覆盖。Cache 继续复用启动期共享后端；Storage 本地 provider 在集群模式下必须提供明确诊断或阻断策略。
- [x] 数据权限：读取、候选、批量、树形、导出和聚合路径均按领域能力边界在查询阶段接入租户与数据权限；批量读取以`MissingIDs`隐藏不存在与不可见差异。租户与组织 scope 过滤只迁移类型归属和注入路径，过滤语义和拒绝策略不变；WASM org/tenant host service 补齐目标用户可见性校验。
- [x] DI 来源：缓存敏感服务复用启动期共享实例或共享后端；WASM host service 和 runtime 依赖显式注入；未在业务路径临时`New()`关键服务图。provider manager 从包级单例迁移到宿主共享实例；WASM 运行期依赖收敛为并发安全的`hostServiceRuntime`快照。
- [x] 跨平台：治理扫描和 WASM 构建均使用 Go 工具链或`linactl`内部组件实现；根`Makefile`和`make.cmd`仅作为薄包装入口。`linactl`动态 WASM builder 继续使用 Go AST 和标准库能力。
- [x] 测试覆盖：后端单元测试、前端类型检查、插件管理与动态插件 E2E、host-only/plugin-full 构建测试、WASM 构建、静态扫描、OpenSpec 校验、import 边界治理测试、descriptor 双向覆盖测试、WASM host service race 测试和发布链路验证。
