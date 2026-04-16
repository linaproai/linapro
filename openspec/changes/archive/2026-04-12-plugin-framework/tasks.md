## 0. 当前实现快照（2026-04-06）

> 当前仓库已完成**第一期：源码插件底座**。本节用于说明已交付范围；后续 checklist 继续作为运行时 `wasm`、多节点热更新与开发者工具的路线图。

- [x] 0.1 新增 `apps/lina-plugins/plugin-demo/` 示例目录、`plugin.yaml` 与插件资源约定
- [x] 0.2 新增 `sys_plugin` 并仅为宿主 `sys_*` 表生成 DAO/DO/Entity，插件私有表不进入宿主生成产物
- [x] 0.3 实现源码插件扫描、注册表同步、启用/禁用 API 与插件管理页，源码插件默认作为随宿主编译的已集成插件管理
- [x] 0.4 实现 `plugin-demo` 最小回调注册式扩展示例、插件页面/Slot 挂载与菜单联动隐藏，且插件特定前后端实现收敛在 `apps/lina-plugins/plugin-demo/`
- [x] 0.5 新增并补齐 `TC0066-source-plugin-lifecycle` 与插件管理 POM，覆盖 source plugin 的 sync/enable/disable、编译整合与 slot 渲染/隐藏
- [x] 0.6 完成 source plugin 的免安装闭环、首批通用 Hook/Slot 与一期验收收口；后续进入运行时 `wasm` 与多节点热更新阶段

## 0. 当前实现快照（2026-04-08）

> 本次继续补齐了一期仍缺失的“review 友好型元数据底座”和开发/运维文档，而没有提前承诺二三期 `runtime wasm` 与多节点热更新能力已经完成。

- [x] 0.7 在宿主 `011-plugin-framework.sql` 中新增 `sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state`
- [x] 0.8 在插件服务层补齐发布记录、迁移记录、资源引用与节点状态同步骨架，并接入现有 `sync/install/uninstall/enable/disable` 链路
- [x] 0.9 将源码插件目录约定进一步固化为可校验规则，新增前端页面/Slot 目录发现校验
- [x] 0.10 新增插件运维指南，并将源码插件目录规范继续收敛到 `plugin-demo` 样例与现有开发指南，便于后续人工 review
- [x] 0.11 扩展插件管理后台列表 DTO 与页面治理摘要，补齐生命周期状态、节点状态、资源引用数与最近迁移结果，便于人工 review
- [x] 0.12 补齐运行时 `wasm` 产物契约校验底座，新增嵌入清单、自定义节、ABI 版本与治理快照摘要，避免动态插件长期停留在“仅有空接口”的状态
- [x] 0.13 让动态插件安装/卸载链路优先执行 `wasm` 内嵌 SQL 资源，并将统计口径、迁移记录和 review 文档同步收敛
- [x] 0.14 新增动态插件包上传入口，支持将 `wasm` 产物落盘到插件工作区并立即同步治理元数据，但仍明确禁止用该入口覆盖已安装 release
- [x] 0.15 补齐 `TC0067-runtime-wasm-lifecycle`，验证 runtime `wasm` 上传、安装、启用、禁用与卸载后的宿主状态收敛
- [x] 0.16 补齐运行时 `wasm` 前端静态资源抽取与公开托管基线，支持宿主按插件 ID + 版本稳定暴露资源 URL，供后续 `iframe` / 新标签页 / 宿主内嵌挂载复用

## 0. 当前收尾判断（2026-04-09）

> 本节用于给当前迭代的人工 review 提供一个“除第三期外还剩什么”的明确口径，避免把低 ROI 工具链项和显式依赖三期能力的项继续混在同一层判断里。

- [x] 0.17 复核“除第三期外”的基础范围后，确认当前真正仍需补齐的核心项收敛为 `1.2`、`4.3`、`7.3`、`7.4`
- [x] 0.18 明确 `1.2` 中“标准目录结构”已通过开发文档落地；后续又明确以 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录作为开发样板，从而收敛“源码插件脚手架模板”缺口
- [x] 0.19 明确 `6.3` 属于有意 deferred 的低 ROI 工具链项；当前不再将“模板与打包脚本”作为本迭代基础能力收尾的阻塞条件
- [x] 0.20 明确 `7.5` 虽写在验收章节，但其语义依赖第三期热升级/代际切换能力，因此不纳入当前“除第三期外”的基础收尾阻塞项
- [x] 0.21 收尾阶段已补齐菜单即时刷新回归、动态插件失败隔离回归，并明确以现有样例目录承担样板职责；除第三期与显式 deferred 项外，当前基础范围已完成

## 0. 当前实现快照（2026-04-11）

> 本节记录本次继续实现后真正完成的第三期热升级收尾能力，明确当前迭代已不再停留在“第三期 deferred”口径。

- [x] 0.22 完成动态插件 `desired_state/current_state/generation/release_id` 代际模型、主节点 Reconciler 与 release 归档切换链路
- [x] 0.23 完成热升级失败回滚、失败 release 隐藏与当前稳定 release 继续服务的保护策略
- [x] 0.24 完成当前插件页的刷新提示与点击刷新后的路由/权限重算；非插件页面用户继续保持无感
- [x] 0.25 修复 active release 重新加载时丢失嵌入 Hook/资源契约的问题，并补齐 `TC0068` 与 `TC0070` 回归
- [x] 0.26 明确当前迭代仍不额外维护 `plugin-template`；开发样板继续以 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录为准

## 第一期当前落地快照（2026-04-06）

- [x] 建立 `apps/lina-plugins/<plugin-id>/` 目录规范，并要求 `plugin-demo` 的插件特定前后端实现收敛在插件目录维护
- [x] 落地源码插件发现、同步、启用/禁用、菜单隐藏与后端扩展点示例，源码插件默认不走安装/卸载流程
- [x] 将 `plugin-demo` 后端示例收敛到插件目录内的 Go 源码实现，并通过构建期静态注册表接入宿主
- [x] 提供 `plugin-demo` 前端页面与 Slot 源码，并通过宿主通用运行时页/Slot 装载器挂载
- [x] 补齐源码插件免安装管理闭环
- [x] 基于插件目录后端注册与 `frontend/slots/**/*.vue` 抽象首批通用 Hook/Slot 总线

## 1. 契约与元数据底座

- [x] 1.1 定义 `plugin.yaml` 清单 schema、版本策略与宿主校验流程
- [x] 1.2 规划 `apps/lina-plugins/<plugin-id>/` 标准目录结构，并补齐源码插件脚手架模板
  - [x] 已完成：标准目录结构、目录职责、最小清单和 review 检查项已写入 `apps/lina-plugins/README.md`
  - [x] 已完成：明确以 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录作为开发样板，无需再单独维护 `hack/plugin-template/`
- [x] 1.3 新增插件元数据 SQL 方案，落地 `sys_plugin`、`sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state` 等基础表
- [x] 1.4 基于新增表生成 DAO/DO/Entity，并建立插件注册、生命周期、资源引用、迁移记录的后端服务骨架
- [x] 1.5 定义插件管理后台 API、DTO、管理页面信息结构以及状态机枚举

## 2. 第一期：源码插件接入

- [x] 2.1 实现源码插件扫描与后端注册表同步，前端资源按目录发现，后端通过集中维护的 `lina-plugins.go` 显式接线
- [x] 2.2 实现前端源码插件清单生成、页面入口发现、Slot 注册与宿主构建集成
- [x] 2.3 打通源码插件的同步发现、启用、禁用管理流程和后台管理界面；运行时安装/卸载留给 `wasm`
- [x] 2.4 实现 `plugin-demo` 源码插件后端能力，覆盖插件目录 Go 源码接入、公开/受保护路由与治理接入
- [x] 2.5 实现 `plugin-demo` 源码插件前端能力，覆盖菜单页展示、宿主页面接入与基本管理交互

## 3. 第一期：治理接入与扩展点发布

- [x] 3.1 扩展菜单、角色与权限链路，使插件菜单和插件权限复用 Lina 通用治理模块
- [x] 3.2 建立宿主后端 Hook 总线，发布首批认证与插件生命周期 Hook，并实现失败隔离与执行观测
- [x] 3.3 建立宿主前端 Slot 注册表，发布首批布局与工作台 Slot，并实现加载失败降级机制
- [x] 3.4 完成插件禁用、重启用及动态插件卸载时的菜单隐藏、权限失效、角色关系保留与资源清理联动

## 4. 第二期：运行时 wasm 插件

- [x] 4.1 定义运行时 `wasm` 产物格式、资源嵌入约定与 ABI 版本策略
- [x] 4.2 实现运行时 `wasm` 插件安装器、校验器、资源提取器与迁移执行器
  - [x] 已完成：上传时校验 `wasm` 文件头、自定义节、嵌入 manifest 与 ABI 版本
  - [x] 已完成：安装/卸载时优先执行内嵌 SQL，并在缺失时回退到目录约定 SQL
  - [x] 已完成：从 `wasm` 中提取前端静态资源，并在启用前校验菜单引用的托管资源契约
- [x] 4.3 基于 WASM Runtime 实现插件加载、Hook 调用、超时控制、错误隔离与卸载回收
  - [x] 已完成：runtime `wasm` 可额外嵌入后端 Hook 与资源声明契约，宿主会在扫描时校验并装载这些声明
  - [x] 已完成：宿主为 runtime Hook 提供 `blocking/async` 执行模式、超时控制与 panic/error 隔离，避免单个动态插件阻断主流程
  - [x] 已完成：禁用或卸载动态插件后，其 Hook 与资源查询能力会从宿主生效链路中退出；重新启用后恢复
- [x] 4.4 实现动态插件静态资源托管与三种前端接入模式：`iframe`、新标签页、宿主内嵌挂载
  - [x] 已完成：从 `wasm` 自定义节提取前端静态资源，并通过 `/plugin-assets/<plugin-id>/<version>/...` 公开托管
  - [x] 已完成：当插件菜单 `path` 指向 `/plugin-assets/...` 托管资源时，宿主会基于 `is_frame` 自动转换为 `iframe` 或新标签页路由语义
  - [x] 已完成：当插件菜单组件为 `system/plugin/dynamic-page` 且 `query_param.pluginAccessMode=embedded-mount` 时，宿主会通过 `dynamic-page` 壳按最小 ESM 挂载协议加载运行时入口
- [x] 4.5 提供独立的 `plugin-demo-dynamic` 动态插件样例，并验证其运行时契约与页面行为
  - [x] 已完成：将旧的 `plugin-demo/runtime-demo` 样例收敛为独立的 `apps/lina-plugins/plugin-demo-dynamic/` 插件目录
  - [x] 已完成：`plugin-demo-dynamic` 仅提供一个左侧菜单，主窗口展示简要页面，并由按钮打开不依赖 Vben 的独立静态页
  - [x] 已完成：自动化测试会验证根据 `backend/`、`frontend/`、`manifest/` 明文源码生成的 `runtime/<plugin-id>.wasm` 与源码树保持一致

## 5. 第三期：多节点热更新与回滚

- [x] 5.1 建立插件 `desired_state/current_state/generation/release_id` 代际模型与主节点切换流程
- [x] 5.2 将主节点选举与节点 Reconciler 接入插件安装、启停、升级与状态收敛链路
- [x] 5.3 实现插件热升级时的新旧代际切换、旧请求自然结束与节点状态上报
- [x] 5.4 实现插件升级失败回滚、迁移异常恢复与前端资源切换失败保护机制
- [x] 5.5 实现前端插件代际感知与“当前插件页面刷新提示”，保证非插件页面用户无感

## 6. 文档、模板与开发者工具

- [x] 6.1 编写插件开发指南，覆盖 `source` 与运行时 `wasm` 两种模式的目录、清单、权限、菜单和扩展点约定
- [x] 6.2 编写插件运维指南，覆盖安装、启停、卸载、升级、回滚、多节点注意事项与故障排查
- [x] 6.3 提供插件模板与打包脚本，帮助开发者快速创建源码插件和运行时产物
  - [x] 当前交付：不再单独维护 `plugin-template`；开发者直接以 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录作为手工创建插件的参考样板
  - [x] 当前交付：运行时产物继续通过现有 `make wasm` / `hack/build-wasm` 入口构建，无需额外新增独立打包脚本目录
- [x] 6.4 补充 `plugin-demo` 的设计说明、发布说明与宿主接入说明，作为后续插件开发参考样板

## 7. E2E 与验收验证

- [x] 7.1 完成 `hack/tests/e2e/plugin/TC0066-source-plugin-lifecycle.ts`，覆盖源码插件 `sync/enable/disable`、编译整合与工作台 slot 渲染切换
  - [x] TC-66a：同步 source 插件后自动处于已集成态，插件管理页无安装按钮
  - [x] TC-66b：启用插件后渲染工作台 slot，并展示左侧插件菜单页
  - [x] TC-66c：启用后可验证插件路由与鉴权访问控制
  - [x] TC-66d：禁用后隐藏工作台 slot，并隐藏插件菜单
  - [x] TC-66e：禁用后源码插件仍保留已集成态，且无需重新安装
- [x] 7.2 创建 `hack/tests/e2e/plugin/TC0067-runtime-wasm-lifecycle.ts`，覆盖 `wasm` 插件安装、启停、卸载与资源托管
  - [x] TC-67a：插件管理页上传入口展示为非白底主按钮，且上传弹窗文案保持精简
  - [x] TC-67b：上传 runtime `wasm` 后，宿主立即识别插件并展示 `WASM / ABI v1` 治理摘要
  - [x] TC-67c：安装并启用 runtime `wasm` 后，宿主状态切换为“已安装 + 已启用”，且公开静态资源可访问
  - [x] TC-67d：禁用并卸载 runtime `wasm` 后，宿主注册态回收到“未安装”或移除运行时条目，且公开静态资源返回不可访问
  - [x] TC-67e：启用后，`iframe` 形态的动态插件菜单会在宿主内容区内嵌打开托管静态资源
  - [x] TC-67f：启用后，新标签页形态的动态插件菜单会直接打开托管静态资源，且当前宿主页保持不变
  - [x] TC-67g：启用后，宿主内嵌挂载形态的动态插件菜单会通过 `dynamic-page` 壳加载 ESM 挂载入口
  - [x] TC-67h：独立的 `plugin-demo-dynamic` 插件启用后，会在宿主页展示简要说明页，并由按钮打开一个不依赖 Vben 的独立静态页面
  - [x] TC-67i：手动删除已启用动态插件产物后，插件列表仍保留该条目、插件菜单立即隐藏，且重新上传同一 `wasm` 后恢复为“未安装 + 未启用”待装态
- [x] 7.3 创建 `hack/tests/e2e/plugin/TC0068-runtime-wasm-failure-isolation.ts`，覆盖 `wasm` 插件失败隔离与回滚
  - [x] TC-68a：runtime Hook 超时或返回错误时，宿主登录链路仍然成功，其他 runtime Hook 继续执行
  - [x] TC-68b：禁用动态插件后其 Hook 停止参与宿主链路，重新启用后恢复执行
- [x] 7.4 创建 `hack/tests/e2e/plugin/TC0069-plugin-permission-governance.ts`，覆盖角色授权、菜单可见性、权限恢复与数据权限上下文
  - [x] TC-69a：动态插件菜单和按钮权限会跟随角色授权、插件禁用隐藏与重新启用恢复，同时资源查询遵循宿主数据权限上下文
- [x] 7.5 创建 `hack/tests/e2e/plugin/TC0070-plugin-hot-upgrade.ts`，覆盖热升级、当前页面刷新提示、多节点代际切换与回退
  - [x] TC-70a：当前插件页热升级时旧版本资源继续可访问，并向当前页面用户提示“刷新当前页面”
  - [x] TC-70b：点击刷新当前页面后，宿主会在不强制跳离当前插件页的前提下重建路由并切换到新代际资源
  - [x] TC-70c：非插件页面用户在插件热升级后保持无感，不出现多余刷新提示或跳转
  - [x] TC-70d：升级失败时宿主会回滚到稳定 release，失败版本资源不会继续公开服务，当前插件页保持可用
- [x] 7.6 为插件管理与插件页面补充所需的 POM（安装/卸载、slot 可见性断言），保证 `TC0066` 可独立运行

## 8. 集群部署与拓扑收敛

- [x] 8.1 新增 `cluster.enabled` 与 `cluster.election.*` 配置语义，明确宿主默认按单节点模式启动
- [x] 8.2 改造 HTTP 启动、领导选举与定时任务调度链路，使主节点专属行为由 `cluster.Service` 统一控制
- [x] 8.3 收敛动态插件在单节点与集群模式下的状态切换、后台 Reconciler 与节点投影行为
- [x] 8.4 收敛 `plugin` / `cluster` / `election` 组件边界，移除 `plugin` 包级集群状态并将 `election` 下沉到 `cluster` 内部实现
- [x] 8.5 补充单节点模式、集群模式、从节点 defer 与拓扑边界收敛的后端测试和相关 OpenSpec 规格

## Feedback

- [x] **FB-1**: `gf gen dao` 只处理宿主 `sys_*` 数据表，插件私有 `plugin_*` 表不再生成到 `lina-core` 的 DAO/DO/Entity
- [x] **FB-2**: 合并 `011-plugin-framework.sql` 与 `012-plugin-lifecycle-state.sql`，同一迭代只保留 1 个 SQL 文件
- [x] **FB-3**: 在项目开发规范文档中明确“宿主 `manifest/sql/` 目录下同一迭代只保留 1 个版本 SQL 文件”
- [x] **FB-4**: 精简 `011-plugin-framework.sql` 的表结构变更逻辑，插件一期按新功能处理，仅保留 `CREATE TABLE`，去掉冗余结构 SQL
- [x] **FB-5**: 插件 SQL 采用与宿主一致的版本命名；卸载 SQL 独立放到 `manifest/sql/uninstall/`，避免被初始化顺序执行误扫
- [x] **FB-6**: `plugin.yaml` 作为统一入口索引菜单声明；插件菜单改用 `sys_menu.menu_key` 与 `parent_key` 维护，去掉对 `remark` 和固定整型 `id/parent_id` 的依赖
- [x] **FB-7**: 未交付阶段将 `sys_menu` 的 `menu_key` 结构与宿主插件菜单种子回收到 `008-menu-role-management.sql`，移除 `011-plugin-framework.sql` 中对应冗余 SQL
- [x] **FB-8**: `plugin-demo` 安装 SQL 去掉 `UPDATE/ON DUPLICATE KEY UPDATE`，插件菜单与授权种子统一使用 `INSERT IGNORE INTO` 幂等写入
- [x] **FB-9**: 删除 `plugin-demo` 冗余的 `manifest/menus.json` 与 `resources.menus` 索引，插件一期菜单以 SQL 为单一真相源
- [x] **FB-10**: 源码插件改为随宿主编译即集成，插件管理页不再为 `source` 类型展示安装/卸载按钮，源码插件默认视为已集成
- [x] **FB-11**: 支持插件目录内后端 Go 源码随宿主一起编译接入，并用 `plugin-demo` 走通“开发-编译-展示”完整链路
- [x] **FB-12**: 调整 `TC0066-source-plugin-lifecycle`，改为验证源码插件“同步发现 + 启用/禁用 + 编译接入后的后端扩展点行为”闭环
- [x] **FB-13**: 修复 `make dev` 后端进程后台保活问题，保证源码插件一期“开发-编译-展示”链路可稳定验证
- [x] **FB-14**: 调整 `plugin-demo` 插件首页体验，菜单打开后台 Tab 页后展示更直观的示例信息，明确告知插件已生效
- [x] **FB-15**: 源码插件首次同步后默认启用，且后续同步不覆盖管理员显式禁用状态
- [x] **FB-16**: `plugin-demo` 需提供“左侧主菜单顶部入口 + 右上角菜单栏入口”两个插件示例页面，并均以内页 Tab 方式打开
- [x] **FB-17**: 插件管理页类型展示调整为“源码插件 / 动态插件”，并在治理视图中统一收敛动态插件展示
- [x] **FB-18**: 清理 `plugin-demo` 前端重复实现，仅保留当前真实生效的页面/Slot 源码资源
- [x] **FB-19**: 修复已启用 `plugin-demo` 后左侧插件菜单未展示的问题，并验证菜单可见性与排序
- [x] **FB-20**: 修复右上角“插件示例”入口点击后 404 的问题，并验证入口以内页 Tab 方式正确打开
- [x] **FB-21**: 修复按钮类型菜单被错误返回到左侧导航/动态路由中的问题，并验证按钮权限不再显示为可导航菜单
- [x] **FB-22**: 修复左侧菜单未按菜单管理排序规则展示的问题，并验证同级菜单按排序号稳定输出
- [x] **FB-23**: 修复“插件管理”被展示为独立顶级目录的问题，将其调整为“系统管理”下的直属菜单
- [x] **FB-24**: 修复页面刷新时重复出现两个“加载菜单中”提示的问题，并验证首次加载只触发一次菜单装载提示
- [x] **FB-25**: 修复排序号为 `0` 的顶级菜单在动态路由响应中丢失 `order` 字段，导致“仪表盘”被前端排到菜单底部
- [x] **FB-26**: 将一期源码插件前端从 `pages/slots/*.json` 描述切换为真实前端源码文件实现，并验证 `plugin-demo` 页面与 Slot 仍可正常挂载
- [x] **FB-27**: 简化 `plugin-demo` 示例插件，移除右上角菜单/页面与登录审计页面入口，仅保留左侧菜单页并收敛其展示内容
- [x] **FB-28**: 补齐宿主系统菜单初始化种子数据的 `menu_key` 字段值，避免 `008-menu-role-management.sql` 初始化后出现空业务标识
- [x] **FB-29**: 在无历史数据债务前提下，直接修改 `008-menu-role-management.sql` 原始菜单种子 `INSERT`，为每条宿主菜单显式写入 `menu_key`，并移除初始化后的回填 `UPDATE`
- [x] **FB-30**: 调整宿主系统菜单的 `menu_key` 命名，移除 `host:` 前缀，仅保留插件菜单使用 `plugin:` 命名空间，并避免宿主插件管理菜单与插件命名空间冲突
- [x] **FB-31**: 清理 `plugin-demo` 清单中的冗余 `backend.apis` 声明，并同步更新 README，明确源码插件后端能力由 Go 编译期注册而非 `plugin.yaml` 路由声明驱动
- [x] **FB-32**: 补齐插件前后端插槽目录、类型化安装位置定义与开发者技术文档，禁止在宿主与插件示例中硬编码 Hook/Slot 位置字符串
- [x] **FB-33**: 将源码插件后端扩展模型升级为回调注册式宿主扩展点，补齐 `auth.login.failed`、`http.route.register`、`http.request.after-auth`、`cron.register`、`menu.filter`、`permission.filter`
- [x] **FB-34**: 在布局、登录页、工作台与 CRUD 通用壳层补齐更多前端 Slot，避免扩展点过度集中在少数页面
- [x] **FB-35**: 更新 `plugin-demo` 示例插件，覆盖新的回调注册后端能力与新增前端 Slot 示例
- [x] **FB-36**: 更新 `apps/lina-plugins/README.md` 与相关插件示例文档，明确新的扩展点目录、注册方式与推荐用法
- [x] **FB-37**: 扩展 `TC0066-source-plugin-lifecycle`，验证新增通用扩展点的可见性、路由装配与鉴权后回调行为
- [x] **FB-38**: 统一后端事件 Hook 与注册式回调扩展点的 Go 类型常量目录，禁止宿主与插件代码继续硬编码后端扩展点字符串
- [x] **FB-39**: 为后端回调注册接口补齐执行模式参数，区分 `blocking` 与 `async`，并由宿主校验每个扩展点允许的执行模式
- [x] **FB-40**: 删除 `internal/service/plugin/plugin.go` 内对 Hook 常量的重复别名导出，并同步更新 `plugin-demo` 与 `apps/lina-plugins` 文档示例
- [x] **FB-41**: 调整 `pkg/pluginhost` 包内源码文件命名，统一使用 `pluginhost_*.go` 风格并同步更新文档引用
- [x] **FB-42**: 去掉后端扩展点公开类型与常量中的 `Backend` 前缀，统一收敛为 `ExtensionPoint*` 风格命名
- [x] **FB-43**: 为插件定时任务注册输入对象补齐“当前是否主节点”的识别方法，供插件自行决定主节点专属执行逻辑
- [x] **FB-44**: 将插件回调注册接口中的对象型输入参数抽象为公开接口，降低插件对宿主内部结构体的直接耦合
- [x] **FB-45**: 精简 `plugin-demo` 后端示例，移除登录审计数据库演示代码，为关键逻辑补充注释，并将示例定时任务调整为每分钟执行
- [x] **FB-46**: 同步更新 `apps/lina-plugins` 与 `plugin-demo` 文档、清单及 E2E 用例，匹配新的命名、接口契约与示例行为
- [x] **FB-47**: 将 `plugin-demo` 子目录说明文档收敛到根目录 `README.md`，移除 `backend`、`frontend`、`manifest` 下的介绍性 `README.md`
- [x] **FB-48**: 将源码插件前端目录约定从 `frontend/src/pages|slots` 统一收敛为 `frontend/pages|slots`，同步宿主扫描逻辑、示例文件与文档规格
- [x] **FB-49**: 移除 `plugin-demo` 的 `crud.table.after` 前端示例，避免在所有基于 `useVbenVxeGrid` 的页面下方默认展示说明内容
- [x] **FB-50**: 修复菜单管理页在树表自动高度场景下页面高度持续增长的问题，并补齐页面高度稳定性的回归验证
- [x] **FB-51**: 将列表页高度持续增长问题收敛到 `useVbenVxeGrid` 共享层修复，并补齐菜单管理、角色管理、操作日志页面的高度稳定回归验证
- [x] **FB-52**: 修复页面重新获得焦点时插件注册表无条件触发菜单与路由重算的问题，避免左侧菜单在插件状态未变化时闪烁重渲染
- [x] **FB-53**: 保持 `plugin-demo` 左侧菜单页原有正文内容不变，修复打开页面时错误弹出的“最小源码插件接入”提示框
- [x] **FB-54**: 简化 `plugin-demo` 左侧菜单页内容，仅保留标题和一句简要介绍
- [x] **FB-55**: 让 `plugin-demo` 页面展示的一小部分内容来自插件后端接口，并将插件后端路由注册升级为与宿主一致的对象式 `Bind` 管理方式
- [x] **FB-56**: 收敛 `plugin-demo` 摘要接口，只保留页面实际使用的 `message` 字段，并同步清理冗余 DTO、service 输出与 E2E 断言
- [x] **FB-57**: 将 `plugin-demo` 后端新增对象式路由模块命名从 `example` 统一收敛为 `demo`，保持插件名与源码模块命名一致
- [x] **FB-58**: 插件 HTTP 路由注册改为宿主独立无前缀分组，并向插件公开宿主可选中间件目录，由插件自行决定路由前缀与中间件组合
- [x] **FB-59**: 明确并示范插件可拆分多个治理策略不同的路由分组进行注册，支持同一插件同时暴露免鉴权与需鉴权接口
- [x] **FB-60**: 规范源码插件后端 `api/controller` 目录命名，要求与宿主 GoFrame `gf gen ctrl` 生成风格保持一致
- [x] **FB-61**: 按 `gf gen ctrl` 实际生成结果重整 `plugin-demo` 控制器目录，拆分 public/protected API 模块并删除旧的手写控制器目录
- [x] **FB-62**: 精简插件路由注册契约，移除 `Public/Protected` 封装分组，仅保留根分组与宿主已发布中间件目录供插件自行组合
- [x] **FB-63**: 插件路由注册支持宿主同款 `group.Group(..., func(*ghttp.RouterGroup){...})` 风格，并将 `plugin-demo` 收敛回单一 `demo` API 模块，通过方法级绑定演示匿名与鉴权路由
- [x] **FB-64**: 将插件路由注册输入对象 `RouteRegistrars` 收敛为单数命名 `RouteRegistrar`，并同步更新构造函数、调用点与开发文档
- [x] **FB-65**: 精简插件管理页列表字段，移除“交付方式/接入态/入口”列，新增“安装时间”并将“类型”列标题调整为“插件类型”
- [x] **FB-66**: 调整插件管理页列表列顺序，将“描述”列移动到“版本”和“状态”列之间
- [x] **FB-67**: 修复插件管理页描述列悬停时出现重复提示的问题，仅保留单一提示层
- [x] **FB-68**: 将插件管理页描述提示改为页面内单例自定义浮层，彻底关闭组件库与表格自身的额外提示
- [x] **FB-69**: 修复插件管理页描述列 hover 数秒后仍出现浏览器原生提示的问题，仅保留页面自定义 tooltip
- [x] **FB-70**: 移除插件管理页描述列自定义 tooltip，改为仅保留单一系统默认提示
- [x] **FB-71**: 精简 `plugin-demo` 后端示例，移除 `RegisterAfterAuthHandler` 与 `RegisterCron` 注册，并同步更新文档与回归用例
- [x] **FB-72**: 收敛插件元数据模型，移除 `sys_plugin` 与 `plugin.yaml` 中重复的 `dynamic` 字段，保留单一 `type` 入口并避免重复建模
- [x] **FB-73**: 收敛插件一级类型定义，仅保留 `source` 与 `dynamic` 两类，并将 `wasm` 下沉为运行时产物语义，统一更新文档、接口描述与类型归一化实现
- [x] **FB-74**: 收敛“动态插件当前仅 `wasm` 实现”的事实，移除 `package` 已实现能力的文档与代码暗示，并同步更新规划任务与规格描述
- [x] **FB-75**: 移除 `hack/plugin` 下新增的脚手架、同步与打包脚本，恢复由开发者手工维护 `apps/lina-plugins/lina-plugins.go` 的源码插件注册方式
- [x] **FB-76**: 精简 `plugin.yaml` 基础字段，去掉 `schemaVersion`、`compatibility`、`entry` 等非必要元数据，并同步收敛宿主校验逻辑
- [x] **FB-77**: 将插件 SQL、前端页面与 `Slot` 发现改回目录约定驱动，去掉 `capabilities`、`resources` 中的重复配置
- [x] **FB-78**: 删除 `metadata` 中重复的菜单/权限前缀配置，统一以插件 SQL 和插件代码作为单一真相源
- [x] **FB-79**: 将 `apps/lina-plugins/README.md` 重写为“当前插件机制设计文档 + 开发指南”，补齐目录约定、校验规则、开发步骤、扩展点契约与 review 清单
- [x] **FB-80**: 将插件相关示例中的版本号写法统一为带 `v` 前缀的形式，例如 `v0.1.0`，保持与常见发布标签习惯一致，同时保留宿主对无前缀写法的兼容
- [x] **FB-81**: 为插件机制核心后端源码补齐文件头说明、公开方法/字段注释和关键逻辑英文注释，便于人工 review
- [x] **FB-82**: 调整插件元数据表与服务实现，禁止持久化具体 SQL 文件路径和前端源码文件路径，改为抽象资源标识与数量摘要
- [x] **FB-83**: 审查插件后端实现中的枚举语义字符串硬编码，统一改为 Go 命名类型常量管理，并将该约束写入项目后端代码规范
- [x] **FB-84**: 将动态样例从 `plugin-demo` 的从属变体调整为独立的 `plugin-demo-dynamic` 插件，左侧菜单页展示简要说明与按钮，并由按钮打开一个不依赖 Vben 的独立静态页面
- [x] **FB-85**: 复核当前迭代“除第三期外”的基础收尾范围，将真正仍需补齐的核心项收敛为 `1.2`、`4.3`、`7.3`、`7.4`
- [x] **FB-86**: 在任务清单中明确 `6.3` 为有意 deferred 的低 ROI 工具链项，不再作为当前基础收尾阻塞条件
- [x] **FB-87**: 在任务清单中明确 `7.5` 依赖第三期热升级/代际切换能力，不纳入当前“除第三期外”的基础收尾阻塞项
- [x] **FB-88**: 将 `plugin-demo-dynamic` 的目录结构收敛为与 `plugin-demo` 一致的 `backend/`、`frontend/`、`manifest/`、`runtime/` 布局，并补齐动态样例的后端代码示例
- [x] **FB-89**: 移除 `plugin-demo-dynamic/runtime/review/runtime-metadata.json` 与已提交的 review 中间态目录，改为直接以插件目录下的 clear-text backend/frontend/manifest 源码作为动态样例的单一真相源
- [x] **FB-90**: 将运行时产物命名从固定 `runtime/plugin.wasm` 调整为 `runtime/<plugin-id>.wasm`，并同步更新宿主校验、上传落盘与样例文档
- [x] **FB-91**: 提供通用 `make wasm` / `make wasm p=<plugin-id>` 构建入口，遍历或定向编译 `apps/lina-plugins` 下的 runtime wasm 插件，并在 `.gitignore` 中忽略生成文件
- [x] **FB-92**: 修复动态插件上传后再次查询 `plugins` 列表时 `sys_plugin_resource_ref` 幂等同步冲突的问题，确保重复同步不会因唯一键冲突而失败
- [x] **FB-93**: 修复动态插件卸载后再次安装同一 release 时 install SQL 被历史 migration 成功记录错误短路的问题，确保重新安装会重新执行 install SQL 并恢复菜单等宿主数据
- [x] **FB-94**: 精简插件管理页列表字段，移除“生命周期”和“治理摘要”两列，并补齐列表列可见性的回归验证
- [x] **FB-95**: 删除 `plugin-demo-dynamic/runtime/` 下已提交的生成产物，仅保留运行时工作目录占位，并补齐忽略规则与文档说明
- [x] **FB-96**: 将 `plugin_test_main_test.go` 重命名为符合 Go 测试文件习惯的 `plugin_test.go`，避免冗余和不一致的测试文件命名
- [x] **FB-97**: 将 runtime 打包开发工具改为根级 `hack/build-wasm/` 独立 Go 工具入口，并移出 `apps/lina-core/internal/cmd`，避免把开发脚本混入主服务命令目录
- [x] **FB-98**: 将动态插件静态资源分流并入宿主现有通用静态资源入口，移除 `cmd_http.go` 中冗余的独立 `/plugin-assets/...` 路由定义
- [x] **FB-99**: 为通用静态资源入口补充关键英文注释，说明为什么必须优先判定并处理动态插件静态资源请求
- [x] **FB-100**: 源码插件在插件管理列表中显示灰态“卸载”按钮并提供 hover 提示，明确源码插件不支持页面动态卸载
- [x] **FB-101**: 收敛插件列表接口字段，删除页面已不展示的 `runtimeKind`、`runtimeAbi`、`releaseVersion`、`lifecycleState`、`nodeState`、`resourceCount`、`migrationState` 定义，并同步更新前端类型与回归断言
- [x] **FB-102**: 动态插件缺少本地生成的 `runtime/<plugin-id>.wasm` 时，插件列表与后台页面不应整体报错；但安装和启用动作仍必须明确拒绝缺少产物的动态插件
- [x] **FB-103**: 将 `hack/build-wasm` 收敛为真正独立的开发工具，不再依赖 `apps/lina-core` 模块或其 `pkg/pluginpack` 包
- [x] **FB-104**: 收敛插件 Hook 事件载荷中的 `map[string]interface{}` 硬编码键名，统一 published payload key 常量与 helper，并审查插件相关源码中的同类字符串键硬编码场景
- [x] **FB-105**: 重做 `plugin-demo-dynamic` 的主窗口挂载页与独立静态页视觉，移除宿主 `dynamic-page` 顶部技术说明，主窗口页使用接近 Vben/Ant 风格的卡片与按钮呈现，独立页改为更专业的中文展示
- [x] **FB-106**: 将 `plugin-demo-dynamic` 主窗口页中的“打开独立页面”按钮进一步收敛为更接近 Vben/Ant 的可点击主按钮样式，并在 hover 时明确呈现 pointer 光标
- [x] **FB-107**: 将 `plugin-demo-dynamic` 主窗口页中的“ESM 挂载”术语改为更易懂的“动态加载”，并统一验证范围与特点卡片的描述长度和高度，避免展示参差
- [x] **FB-108**: 取消 runtime 前端资源对 `runtime/frontend-assets/` 磁盘提取目录的运行时依赖，改为以 `runtime/<plugin-id>.wasm` 为真相源，在宿主内存中缓存前端资源并支持服务重启后的启动预热与请求时懒加载兜底
- [x] **FB-109**: 为 runtime 前端内存资源 bundle 补齐关键 debug 日志，覆盖启动预热、缓存命中/重建、请求时懒加载与缓存失效路径，便于后续排障和人工 review
- [x] **FB-110**: 将动态插件发现与上传从 `apps/lina-plugins/<plugin-id>/plugin.yaml` 外层目录解耦，改为统一扫描并写入 `plugin.dynamic.storagePath` 下的 `.wasm` 文件，同时支持手动拷贝后在管理页执行同步识别
- [x] **FB-111**: 调整插件管理页上传按钮为非白底主按钮并更名为“上传插件”，同时精简上传弹窗说明文案
- [x] **FB-147**: 删除已废弃 `GetElection()` 与 `config_election.go`，将宿主配置统一收敛到 `cluster` 配置段
- [x] **FB-148**: 补齐运行配置中的 `cluster` 段，并移除旧版顶层 `election.*` 兼容和二次读取逻辑，降低配置复杂度
- [x] **FB-149**: 移除 `plugin` 包级集群状态，改为 `plugin.Service` 持有显式 topology 抽象并复用统一节点标识
- [x] **FB-150**: 将 `election` 下沉到 `cluster` 组件内部，统一选主实现归属与测试归属
- [x] **FB-151**: 在项目规范文档中补充后端 `service` 层文件顶部注释与主文件注释位置规范
- [x] **FB-152**: 将 `cluster-deployment-toggle` 与 `refine-cluster-service-boundaries` 合并回 `plugin-framework`，清理错误的 plugin archive 痕迹
- [x] **FB-154**: 审查并收敛后端 `service` 层在接口执行链路中的临时 `service.New()` 调用，统一改为构造阶段依赖注入，并将该约束补充到项目规范文档
- [x] **FB-153**: 完成插件机制当前阶段的人工校验，并在确认通过后再执行正式归档
- [x] **FB-112**: 修复动态插件产物被手动删除后宿主注册态未自动收敛的问题，确保插件列表仍可见缺失条目、菜单与路由立即隐藏，且公共运行时状态同步返回“未安装/未启用”
- [x] **FB-113**: 修复动态插件产物缺失时重新上传仍被“已安装不可覆盖”错误拦截的问题，允许将缺失产物作为恢复性重传重新落盘，并补齐对应回归验证
- [x] **FB-114**: 调整 `TC0067-runtime-wasm-lifecycle` 的 bundled runtime 样例准备方式，禁止再向 `apps/lina-plugins/plugin-demo-dynamic/runtime/` 回写生成产物，改为仅写入宿主 `plugin.dynamic.storagePath`
- [x] **FB-115**: 将 `apps/lina-plugins/plugin-demo-dynamic/README.md` 改写为中文文档，并保持与当前动态样例实现和仓库约定一致
- [x] **FB-116**: 精简动态插件上传成功后的提示信息，仅保留“已上传成功，可继续安装/启用”的必要指引，并补齐对应前端回归断言
- [x] **FB-117**: 收敛动态插件上传成功态交互，仅保留单一确认按钮并同步更新插件管理页回归用例
- [x] **FB-118**: 修复菜单管理页将菜单改为隐藏后左侧导航未即时刷新的问题，统一收敛当前登录用户菜单/路由刷新逻辑并补齐菜单可见性回归验证
- [x] **FB-119**: 统一插件机制中的人类可读命名，文档、源码注释、接口文案、页面文案与测试断言统一使用“动态插件”，但保留内部一级类型枚举值 `dynamic`
- [x] **FB-120**: 调查 `TC0068-runtime-wasm-failure-isolation` 中登录审计 Hook 未写入记录的回归，修复动态插件失败隔离链路并恢复对应 E2E 断言
- [x] **FB-121**: 将动态插件内部类型标识从 `runtime` 统一收敛为 `dynamic`，同步更新配置键、API 路径、宿主页壳标识、样例插件 `plugin-demo-dynamic` 与相关文档测试
- [x] **FB-122**: 审查并统一动态插件相关源码文件名，将仍使用 `plugin_runtime` 命名的源码文件收敛为 `plugin_dynamic`
- [x] **FB-123**: 将源码插件资源嵌入责任下沉到插件自身，通过 `//go:embed` 在插件包内嵌入 `plugin.yaml`、约定目录 SQL 与前端资源索引，并移除宿主侧聚合 source catalog 构建
- [x] **FB-124**: 调整 `make build` 输出体验，默认隐藏前端与构建工具详细日志，仅在显式传入 `verbose=1` 时输出完整编译信息
- [x] **FB-125**: 为构建详细日志开关增加 `v=1` 短参数别名，并保持与 `verbose=1` 行为一致
- [x] **FB-126**: 调整 `make build` 默认输出，在静默模式下保留动态插件编译阶段的简短提示信息，同时继续隐藏详细子日志
- [x] **FB-131**: 宿主二进制的 embed 资源需额外包含 `manifest/sql` 与 `manifest/config` 交付资源，但必须排除本地 `manifest/config/config.yaml`
- [x] **FB-127**: 修复登录后插件治理链路中的重复查库与重复插件扫描，消除菜单/权限/Hook 热路径里的 `sys_plugin` N+1 查询
- [x] **FB-128**: 继续收敛插件治理热路径中的 `sys_plugin` 重复查库，改为宿主内存快照缓存并在插件生命周期变更时精确失效
- [x] **FB-129**: 将在线会话活跃时间刷新从 `COUNT + UPDATE` 收敛为单条 `UPDATE`，减少每次鉴权请求的固定 SQL 数量
- [x] **FB-130**: 收敛 `/api/v1/user/info` 角色、权限与菜单装载链路中的重复查询，避免同一次请求重复读取 `sys_user_role`
- [x] **FB-132**: 将宿主 `manifest` 交付资源的 embed 方案收敛为“编译前同步到 `internal/packed` 再统一嵌入”，并将生成文件加入版本忽略
- [x] **FB-133**: 为 `prepare-packed-assets.sh` 与 `hack/makefiles/build.mk` 中新增的 manifest 嵌入准备逻辑补齐关键注释，便于人工 review
- [x] **FB-134**: 将源码样例插件从 `plugin-demo` 统一更名为 `plugin-demo-source`，插件显示名与左侧菜单统一收敛为“源码插件示例”，并移除所有非左侧菜单的前端 Slot 示例
- [x] **FB-135**: 去掉 `internal/service/plugin` 中 `sys_plugin` 查询缓存与相关失效逻辑，改回直接查库以降低当前实现复杂度
- [x] **FB-136**: 将插件菜单注册从 SQL 迁移为 manifest 元数据驱动，源码插件同步与动态插件安装按菜单元数据幂等写入 `sys_menu`，动态插件卸载按同一菜单元数据删除菜单与角色关联
- [x] **FB-137**: 为 `plugin.yaml` 中新增的 `menus` 元数据及其关键字段补齐注释说明，便于开发者理解菜单注册约束与常见写法
- [x] **FB-138**: 移除冗余 `hack/plugin-template` 目录，改为以 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个真实样例目录作为插件开发样板，并同步收敛文档与任务口径
- [x] **FB-139**: 将 active dynamic release 的 manifest 重载统一收敛为“包含嵌入 Hook/资源契约的完整运行时 manifest”，避免热升级或回滚后丢失认证 Hook 与通用资源查询能力
- [x] **FB-140**: 强化 `TC0068-runtime-wasm-failure-isolation` 的自清理与业务成功断言，确保测试会删除 `plugin.dynamic.storagePath` 下的 staged/release 产物，并在 `code != 0` 时立即失败
- [x] **FB-141**: 修复多处系统页面错误等待同步字典 getter 导致筛选/表单选项为空的问题，并补齐监控、部门、岗位、角色、用户相关回归
- [x] **FB-142**: 加固字典、配置、角色、用户角色等 E2E POM 与用例，修复固定列 DOM、工具栏按钮歧义和 worker 重启后跨子用例共享状态导致的脆弱失败
- [x] **FB-143**: 修复插件管理页打开时公共插件状态查询重复触发全量同步的问题，将 `/plugins/dynamic` 收敛为纯读链路并补齐前端请求去重回归
- [x] **FB-144**: 补充多节点部署下本地磁盘上传资源不会自动跨节点分发的设计与使用文档说明，明确 `plugin.dynamic.storagePath`、`upload.path` 等目录需依赖共享存储或外部分发
- [x] **FB-145**: 重建本地 `sys_plugin` 与 `sys_plugin_release` 表结构并重新生成 DAO/DO/Entity，统一 `type` 字段注释为 `source/dynamic`
- [x] **FB-146**: 收敛插件一级类型归一化逻辑，删除对历史 `runtime` 类型常量的兼容，仅保留 `source` 与 `dynamic`
