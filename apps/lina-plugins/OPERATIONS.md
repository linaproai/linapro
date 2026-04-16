# Lina 插件运维与 Review 指南

本文档只描述**当前仓库已经落地的真实能力**，用于帮助维护者和人工 reviewer 在执行插件同步、启用、禁用、安装、卸载相关操作后，快速确认宿主侧元数据是否保持一致。

## 适用范围

当前仓库仍然以**第一期源码插件底座**为主，但 `apps/lina-plugins/` 下已经同时提供：

- `plugin-demo-source`：源码插件样例
- `plugin-demo-dynamic`：动态插件样例

动态插件完整能力尚未交付，但宿主已经先补齐了后续阶段会复用的元数据表：

- `sys_plugin`
- `sys_plugin_release`
- `sys_plugin_migration`
- `sys_plugin_resource_ref`
- `sys_plugin_node_state`

这意味着，即使二三期 `dynamic wasm`、多节点热更新等能力还没有完成，reviewer 也已经可以通过数据库直接看到：

- 宿主当前识别到的插件注册状态
- 宿主认为当前生效的是哪个插件版本
- 插件安装/卸载迁移的执行记录与抽象执行键
- 宿主从插件目录中发现了哪些资源类型，以及对应的数量摘要与抽象标识
- 当前节点对插件状态的本地观测结果

## 当前生命周期模型

## 源码插件

源码插件通过 `apps/lina-plugins/<plugin-id>/plugin.yaml` 被宿主发现。

当前实际行为如下：

1. 宿主扫描插件目录。
2. 宿主校验 `plugin.yaml`、`go.mod`、`backend/plugin.go`、SQL 命名规则以及前端页面/`Slot` 目录约定。
3. 宿主同步 `sys_plugin` 插件注册表。
4. 宿主将当前清单基础字段和资源数量摘要写入 `sys_plugin_release`。
5. 宿主将目录发现到的资源类型与抽象 owner 信息写入 `sys_plugin_resource_ref`。
6. 宿主将当前节点的插件状态投影写入 `sys_plugin_node_state`。

当前约束如下：

- 源码插件视为“随宿主编译即已集成”，不走安装/卸载流程。
- 源码插件禁用后，只会隐藏路由、菜单、页面和 `Slot`，不会删除历史业务数据。
- 当前仓库至少保留 `plugin-demo-source` 与 `plugin-demo-dynamic` 两个样例目录，分别用于 source / dynamic 两种接入模式的 review。
- 当前源码样例 `plugin-demo-source` 只保留一个左侧菜单页，不再额外演示登录页、工作台、CRUD 或右上角 `Slot` 扩展。

## 动态插件

动态插件的完整执行模型仍在后续阶段，但宿主服务层已经先补齐了一部分基础记录能力。

当前已经落地的行为：

- 宿主已提供 `POST /api/v1/plugins/dynamic/package` 上传入口，用于把 dynamic wasm 产物写入 `plugin.dynamic.storagePath`。
- 运维也可以手工将 `.wasm` 文件复制到 `plugin.dynamic.storagePath`，再通过插件管理页执行同步识别。
- 宿主会扫描 `plugin.dynamic.storagePath/*.wasm`，并验证 `wasm` 文件头、自定义节和 ABI 版本。
- 若 `wasm` 内嵌了 `lina.plugin.frontend.assets`，宿主会直接从 `.wasm` 构建内存中的只读前端资源 bundle，而不是先提取到工作区目录。
- 当动态插件处于“已安装 + 已启用”状态时，宿主会通过 `/plugin-assets/<plugin-id>/<version>/...` 对外公开这些静态资源。
- 当插件菜单的 `path` 指向 `/plugin-assets/<plugin-id>/<version>/...` 时，宿主会继续复用现有菜单体系并按 `sys_menu.is_frame` 解释前端接入模式：
  - `is_frame = 0`：在宿主主内容区以内嵌 `iframe` 方式打开托管页面。
  - `is_frame = 1`：点击菜单后直接在新标签页打开托管页面。
- 当插件菜单额外声明 `component = system/plugin/dynamic-page` 且 `query_param.pluginAccessMode = embedded-mount` 时，宿主会改走“宿主内嵌挂载”模式，并在 `dynamic-page` 壳内部动态导入对应的托管 ESM 入口。
- 若 `wasm` 内嵌了 `lina.plugin.install.sql` / `lina.plugin.uninstall.sql`，宿主会优先执行这些嵌入 SQL，并把执行结果记录到 `sys_plugin_migration`。
- 当运行时 `wasm` 未声明嵌入 SQL 时，安装/卸载生命周期仍会回退到目录约定 SQL。
- 当动态插件准备启用时，宿主会再次校验插件菜单引用的托管资源是否已经能从 `.wasm` 中解析出来；若菜单指向不存在的 `/plugin-assets/...` 资源，或者宿主内嵌挂载入口未满足最小 ESM 契约，则启用会被拒绝。
- 安装完成后会同步当前版本快照、资源引用和节点状态。
- 卸载完成后会清理当前版本的资源引用、对应的内存前端资源缓存，并刷新节点状态。

当前**尚未**落地的行为：

- 真正的运行时装载、热升级、回滚和多节点代际切换

## 上传动态插件包

当前可以通过以下方式上传 dynamic wasm 插件包：

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/plugins/dynamic/package" \
  -H "Authorization: Bearer <token>" \
  -F "file=@./plugin-demo-dynamic.wasm" \
  -F "overwriteSupport=0"
```

当前约束如下：

- 只能上传 `.wasm` 文件。
- 宿主会从上传包中解析插件 ID，因此不接受额外手工指定插件 ID。
- 宿主会把上传产物规范化写入 `plugin.dynamic.storagePath/<plugin-id>.wasm`。
- 若同 ID 动态插件文件已存在，只有在 `overwriteSupport=1` 且该插件尚未安装时才允许覆盖。
- 若目标插件已经处于已安装状态，当前上传接口会拒绝覆盖，因为正式 upgrade/release 切换能力仍未实现。
- 若 `storagePath` 下存在多个嵌入相同 plugin ID 的 `.wasm` 文件，宿主会在同步阶段直接报冲突错误，要求先人工清理。

当前仓库中的动态样例插件不再提交编译后的 `.wasm` 文件，而是统一通过通用命令生成：

```bash
make wasm
make wasm p=plugin-demo-dynamic
```

运维与 review 时需要注意：

- 运行 `make dev` 或 `make build` 前后，宿主都会复用同一个通用构建入口
- 如需单独验证某个动态样例目录，优先使用 `make wasm p=<plugin-id>`
- 通用构建入口当前由根级 `hack/build-wasm/` 独立 Go 工具负责，该工具自身维护 `go.mod`，不依赖主服务模块
- 标准仓库构建流程会把动态插件产物和 guest runtime 中间 `wasm` 统一收敛到仓库根 `temp/output/`，不再回写 `apps/lina-plugins/*/temp/`

## 运行时前端资源访问

如果动态插件声明了 `lina.plugin.frontend.assets`，宿主会在启动预热或首次访问时直接从 `plugin.dynamic.storagePath` 中的 `.wasm` 解析这些资源到内存 bundle，并通过下面的稳定地址对外提供：

```text
http://127.0.0.1:8080/plugin-assets/<plugin-id>/<version>/<relative-path>
```

运维和 review 时需要注意：

- 资源路径中的 `<version>` 必须与当前插件版本一致。
- 插件未安装、未启用、已禁用或已卸载时，该 URL 应返回不可访问。
- 服务重启后，宿主会优先预热“已安装 + 已启用”的 runtime 资源；若预热失败，请求链路仍会按需从 `.wasm` 懒加载。
- 若要走菜单驱动访问，插件菜单 SQL 需要把 `path` 直接写成 `/plugin-assets/<plugin-id>/<version>/...`。
- `is_frame = 0` 与 `is_frame = 1` 当前分别对应 `iframe` 与新标签页两种访问模式，review 时应同时核对菜单记录和最终页面行为。
- 若要走宿主内嵌挂载，菜单还必须声明：
  - `component = system/plugin/dynamic-page`
  - `query_param` 至少包含 `{"pluginAccessMode":"embedded-mount"}`
- 宿主会把原始托管资源地址透传为 `embeddedSrc` 查询参数，并在 `dynamic-page` 壳中按最小 ESM 契约导入该入口。
- 当前 ESM 契约至少要求入口导出 `mount(context)`；`unmount(context)`、`update(context)` 为可选能力。
- 若菜单选择宿主内嵌挂载模式，则入口文件扩展名当前必须为 `.js` 或 `.mjs`。
- 前端资源缓存只是宿主运行态优化，真正的治理依据仍应以 `plugin.dynamic.storagePath` 下的 `.wasm`、`sys_plugin_release`、`sys_plugin_resource_ref` 和当前插件状态为准。

## Review 检查清单

建议人工 review 按下面顺序核对。

## 1. 插件注册表

核对 `sys_plugin`：

- `plugin_id`、`name`、`version`、`type` 是否与源码插件清单或 dynamic wasm 嵌入清单一致
- `installed`、`status` 是否与本次操作预期一致
- 对源码插件，`manifest_path` 应指向正确的 `plugin.yaml`
- 对动态插件，`manifest_path` 当前允许为空，因为宿主以 `.wasm` 嵌入清单为真相源

建议查询：

```sql
SELECT plugin_id, version, type, installed, status, manifest_path, installed_at, enabled_at, disabled_at
FROM sys_plugin
ORDER BY plugin_id;
```

## 2. 发布快照

核对 `sys_plugin_release`：

- `release_version` 是否与当前生效版本一致
- `runtime_kind` 对动态插件应为 `wasm`
- `status` 是否符合当前生命周期动作（如 `active`、`installed`、`uninstalled`）
- `manifest_snapshot` 中应包含基础清单字段和资源数量摘要，而不是具体文件路径
- 若为动态插件，`manifest_snapshot` 中应额外包含 `runtimeKind`、`runtimeAbiVersion`、`runtimeFrontendAssetCount`、`runtimeSqlAssetCount`
- `manifest_snapshot` 中的 `installSqlCount`、`uninstallSqlCount`、`frontendPageCount`、`frontendSlotCount` 是否与目录实际情况一致
- 若动态插件声明了嵌入 SQL，则 `installSqlCount` / `uninstallSqlCount` 当前表示**宿主最终会执行的 SQL 资源数量**，不再局限于目录扫描结果

建议查询：

```sql
SELECT plugin_id, release_version, type, status, manifest_path, package_path, checksum
FROM sys_plugin_release
ORDER BY plugin_id, release_version;
```

## 3. 迁移执行记录

在动态插件安装/卸载链路中，核对 `sys_plugin_migration`：

- `phase` 只能是 `install` 或 `uninstall`
- `migration_key` 应为抽象执行键，例如 `install-step-001`
- `status = succeeded` 表示执行成功
- 如果 SQL 内容变更，`checksum` 也应随之变化

建议查询：

```sql
SELECT plugin_id, release_id, phase, migration_key, execution_order, status, error_message, executed_at
FROM sys_plugin_migration
ORDER BY plugin_id, release_id, phase, execution_order;
```

## 4. 资源引用

核对 `sys_plugin_resource_ref`：

- 每个发现到的资源都应有稳定的 `resource_type` 与 `resource_key`
- `resource_path` 当前应保持为空字符串，或仅作为宿主抽象定位补充信息，不允许保存具体前端/SQL 文件路径
- `remark` 应描述数量摘要或宿主识别结论，便于 review 快速判断发现结果
- 当前源码插件快照至少应覆盖：
  - `manifest`
  - `backend_entry`
  - `install_sql`
  - `uninstall_sql`
  - `frontend_page`
  - `frontend_slot`
- 当前动态插件快照至少还应覆盖：
  - `runtime_wasm`
  - `runtime_frontend`（当运行时产物声明了前端静态资源时）

建议查询：

```sql
SELECT plugin_id, release_id, resource_type, resource_key, resource_path, owner_type, owner_key, remark
FROM sys_plugin_resource_ref
ORDER BY plugin_id, release_id, resource_type, resource_key;
```

## 5. 节点状态投影

核对 `sys_plugin_node_state`：

- `node_key` 应为当前宿主节点主机名
- `release_id` 应指向当前节点观测到的发布记录
- `desired_state` 与 `current_state` 应与 `installed/enabled` 组合保持一致
- `generation` 不应回退；当前实现至少与 `release_id` 对齐或更大

当前映射规则：

- `installed = 0` -> `uninstalled`
- `installed = 1` 且 `enabled = 0` -> `installed`
- `installed = 1` 且 `enabled = 1` -> `enabled`

建议查询：

```sql
SELECT plugin_id, release_id, node_key, desired_state, current_state, generation, last_heartbeat_at, error_message
FROM sys_plugin_node_state
ORDER BY plugin_id, node_key;
```

## 常见 Review 问题

## 为什么资源引用不直接来自 `plugin.yaml`？

因为当前插件设计明确要求 `plugin.yaml` 保持最小化。页面、`Slot`、SQL 等信息应由真实目录结构推导，而不是在清单里重复维护一份配置，避免双真相源。

## 为什么宿主还要额外保存 `manifest_snapshot`？

因为人工 review 不能只依赖磁盘上的当前文件，还需要一个“宿主在某次同步时到底看到了什么”的持久化快照。但当前快照只保存基础清单字段和数量摘要，不保存具体 SQL 文件路径或前端源码路径，避免把框架实现细节硬编码进治理表。

## 为什么多节点运行时能力还没做完，就先加了 `sys_plugin_node_state`？

因为后续多节点阶段一定需要宿主持久化节点侧观测结果。现在先把表结构补齐，可以避免后续继续改底层元数据模型，也能让当前人工 review 立即获得一个节点视角的检查入口。

## 当前已知限制

- 宿主当前还没有把“源码插件目录被删除”自动回收进注册表同步逻辑中。
- 宿主当前还没有把插件菜单 SQL 进一步解析成 `menu_key` 级别的宿主资源引用记录。
- 动态插件的真实产物上传、装载隔离和热升级仍未实现。
- 仓库当前不提供插件打包脚本；这是有意保留的约束，避免出现与真实实现脱节的辅助脚本。

## 当前 review 结论口径

如果人工 reviewer 只想判断“除第三期外，当前基础工作是否已全部完成”，应使用以下口径：

- 不应再把低 ROI 的模板/打包脚本缺失视为当前基础能力阻塞项。
- 不应把依赖第三期热升级/代际切换的验收项误计入当前基础收尾缺口。
- 当前仍然真正阻塞基础收尾的，主要是 runtime `wasm` 的真实执行与隔离能力，以及与之对应的失败隔离/权限治理验收闭环。
