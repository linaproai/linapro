# 插件运行时加载规范

## Purpose

定义动态插件运行时加载行为、集中式 Wasm 自定义段解析、跨节点派生缓存失效、Wasm 编译缓存键和产物刷新一致性。
## Requirements
### Requirement:WASM 自定义段解析能力必须由 pluginbridge 集中提供
宿主系统 SHALL 通过 `apps/lina-core/pkg/pluginbridge` 体系提供 `ReadCustomSection(content []byte, name string) ([]byte, bool, error)` 和 `ListCustomSections(content []byte) (map[string][]byte, error)` 公共能力，集中实现 `wasm` 文件头验证、段遍历和 ULEB128 解码。该能力可以由 `pluginbridge` 根包 facade 或 `pluginbridge/artifact` 等职责明确的子组件公开，但协议实现必须只有一个权威位置。`apps/lina-core/internal/service/i18n`、`apps/lina-core/internal/service/apidoc` 和插件运行时必须通过此公共能力从动态插件运行时产物中读取自定义段（如 `i18n_assets`、`apidoc_assets`），不得在业务包中维护重复的 WASM 解析实现。`pluginbridge.WasmSection*` 段名常量或其子组件等价常量必须由 `pluginbridge` 体系集中维护。

#### Scenario:i18n 通过 pluginbridge 读取动态插件 i18n 段
- **当** 系统需要从动态插件运行时产物中读取 `i18n_assets` 自定义段时
- **则** 调用方通过 `pluginbridge.ReadCustomSection(content, pluginbridge.WasmSectionI18NAssets)` 或 `pluginbridge/artifact` 的等价入口完成
- **且** `i18n` 包中不存在 `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` 等专用解析函数

#### Scenario:修复 WASM 解析缺陷只需修改 pluginbridge 体系
- **当** WASM 解析需要扩展以支持新段、修复解码错误或添加边界检查时
- **则** 修改 `pkg/pluginbridge` 对应 artifact/wasm section 子组件的权威实现即可
- **且** `i18n` 包和插件运行时不需要重复变更

### Requirement:动态插件运行时派生缓存必须跨节点失效

动态插件安装、启用、禁用、卸载、升级或同版本刷新后，系统 SHALL 使用统一缓存协调机制使所有节点上的插件运行时派生缓存失效或刷新。

#### Scenario:非主节点观察到插件运行时修订号变更

- **当** 集群模式下主节点完成动态插件运行时状态转换并发布插件运行时缓存修订号时
- **则** 非主节点在下一个请求路径或监听路径上观察到新修订号
- **且** 非主节点刷新插件启用快照
- **且** 非主节点使插件前端包、运行时 i18n 包和 Wasm 编译缓存失效

#### Scenario:插件禁用后非主节点不再暴露能力

- **当** 主节点上动态插件被禁用或卸载时
- **则** 非主节点不得在插件运行时缓存域允许的陈旧窗口之外继续从过期本地缓存暴露该插件的菜单、前端资产或动态路由能力

### Requirement:Wasm 编译缓存必须绑定到产物校验和或 generation

系统 SHALL 将动态插件 Wasm 编译缓存绑定到当前活跃发布的产物校验和或 generation。不得仅通过可变产物路径决定缓存复用。

#### Scenario:同版本动态插件刷新重新编译

- **当** 动态插件以相同版本但产物校验和变更进行刷新时
- **则** 节点观察到插件运行时修订号变更后，不得继续命中旧校验和的 Wasm 编译缓存
- **且** 下一次动态路由或动态任务执行必须从新产物编译或加载

#### Scenario:相同产物路径但不同校验和

- **当** 活跃发布产物路径与旧缓存路径相同但校验和不同时
- **则** 系统将其视为不同的编译缓存条目
- **且** 旧条目必须失效或自然清理

### Requirement:动态插件产物归档必须支持同版本刷新一致性

系统 SHALL 确保同版本刷新后的活跃发布指向可验证的新产物内容，并且其他节点可使用共享发布状态判断本地缓存是否过期。

#### Scenario:同版本刷新写入新产物

- **当** 插件同版本刷新提交新产物内容时
- **则** 系统更新活跃发布的校验和或 generation
- **且** 发布插件运行时缓存修订号
- **且** 其他节点可使用活跃发布的校验和或 generation 判断本地缓存是否需要重建

#### Scenario:旧产物清理不影响当前活跃发布

- **当** 系统清理旧动态插件产物时
- **则** 当前活跃发布引用的产物不得被删除
- **且** 仍被本地缓存引用但不再活跃的产物可根据保留策略稍后清理

### Requirement: 插件运行时变更必须发布 Redis coordination event
系统 SHALL 在集群模式下为插件安装、启用、禁用、卸载、升级、active release 切换、动态插件 artifact 变化发布 `plugin-runtime` Redis revision 和 event。

#### Scenario: 动态插件启用后其他节点刷新
- **WHEN** 主节点启用动态插件 P
- **THEN** 系统发布 `plugin-runtime` Redis revision
- **AND** 其他节点收到 event 后刷新 enabled snapshot
- **AND** 其他节点可路由到插件 P 的 active release

#### Scenario: 动态插件禁用后其他节点隐藏
- **WHEN** 主节点禁用动态插件 P
- **THEN** 系统发布 `plugin-runtime` Redis revision
- **AND** 其他节点失效 frontend bundle、runtime i18n 和 Wasm 派生缓存
- **AND** 后续访问插件 P 路由返回不可用或 404

### Requirement: 插件 runtime freshness 不可确认时必须 conservative-hide
系统 SHALL 在无法确认 `plugin-runtime` revision freshness 且超过最大陈旧窗口时采用 conservative-hide 策略。系统不得暴露可能已禁用、卸载或权限变化的插件能力。

#### Scenario: Redis plugin-runtime revision 不可读
- **WHEN** 请求需要访问动态插件运行时能力
- **AND** Redis `plugin-runtime` revision 不可读
- **AND** 本地插件运行时缓存超过最大陈旧窗口
- **THEN** 系统隐藏或拒绝该插件能力
- **AND** 不使用陈旧缓存继续放行

### Requirement: 动态插件 reconciler 必须由 Redis revision 唤醒
系统 SHALL 在集群模式下使用 Redis revision/event 唤醒动态插件 reconciler。安全扫描或低频 sweep MAY 保留作为兜底。

#### Scenario: active release 变化唤醒 reconciler
- **WHEN** 动态插件 active release 记录变化
- **THEN** 系统发布 reconciler scope 的 `plugin-runtime` revision
- **AND** 需要收敛的节点在观察到 revision 前进后执行收敛

#### Scenario: 事件错过后安全 sweep 兜底
- **WHEN** 节点错过 reconciler event
- **THEN** 节点通过 revision check 或低频 safety sweep 发现需要收敛
- **AND** 最终运行时状态与权威 release 记录一致

### Requirement: 插件派生缓存失效必须按 scope 精细化
系统 SHALL 在插件运行时变更时按插件 ID、sector、locale 或 global scope 精细失效 frontend bundle、runtime i18n 和 Wasm 缓存。普通路径不得无理由清空所有插件所有派生缓存。

#### Scenario: 单插件 frontend bundle 失效
- **WHEN** 动态插件 P 上传新 frontend bundle
- **THEN** 系统仅失效插件 P 相关 frontend bundle cache
- **AND** 其他插件 bundle cache 保持可用

### Requirement: 动态插件生命周期契约必须支持构建期自动发现

系统 SHALL 在动态插件打包阶段自动发现 guest controller 中与源码插件生命周期同名的 bridge handler 方法，并为其生成动态插件生命周期契约。自动发现生成的契约 MUST 写入动态插件 WASM artifact 的生命周期 custom section，宿主运行时 MUST 继续以 artifact 中的显式生命周期契约作为唯一调用依据。

#### Scenario: 构建期发现生命周期方法

- **WHEN** 动态插件 controller 暴露合法 bridge handler 方法 `BeforeInstall`
- **AND** 插件未提供 `backend/lifecycle` override 声明
- **THEN** `build-wasm` 自动生成 `operation=BeforeInstall` 的生命周期契约
- **AND** 生成的契约写入动态插件 WASM artifact 的生命周期 custom section

#### Scenario: 宿主运行时不盲探生命周期方法

- **WHEN** 宿主加载动态插件 artifact
- **THEN** 宿主只读取 artifact 中的生命周期契约
- **AND** 宿主不得通过试探调用 `Before*` 或 `After*` 路径来判断动态插件是否实现生命周期处理器

#### Scenario: 未实现生命周期方法时不生成契约

- **WHEN** 动态插件 controller 未暴露 `BeforeUninstall` 方法
- **THEN** `build-wasm` 不生成 `operation=BeforeUninstall` 的生命周期契约
- **AND** 宿主执行对应生命周期场景时不得调用该动态插件的 `BeforeUninstall` 处理器

### Requirement: 生命周期自动发现必须复用 guest dispatcher 元数据规则

系统 SHALL 使用与动态插件 guest dispatcher 一致的 controller 反射规则发现生命周期 handler 元数据。自动发现 MUST 只接受 guest dispatcher 支持的 bridge handler 签名，并使用同一套 request type 与内部路径推导规则，避免构建期契约与运行时 guest 分发规则不一致。

#### Scenario: 自动发现使用 dispatcher 支持的签名

- **WHEN** 动态插件 controller 方法 `BeforeInstall` 满足 guest dispatcher 支持的 bridge handler 签名
- **THEN** `build-wasm` 可以将该方法识别为生命周期 handler
- **AND** 生成契约中的 `requestType` 与 dispatcher 对该方法的 request type 推导一致

#### Scenario: 自动发现忽略非法签名方法

- **WHEN** 动态插件 controller 存在名为 `BeforeInstall` 但签名不符合 guest dispatcher bridge handler 规则的方法
- **THEN** `build-wasm` 不得为该方法生成生命周期契约
- **AND** 构建结果不得包含无法由 guest dispatcher 执行的生命周期 handler

#### Scenario: 自动发现拒绝旧命名

- **WHEN** 动态插件 controller 暴露 `CanInstall`、`CanUninstall` 或 guard 风格生命周期方法
- **THEN** `build-wasm` 不得为这些方法生成生命周期契约
- **AND** 构建诊断必须继续要求使用源码插件一致的 `Before*` 或 `After*` 生命周期操作名称

### Requirement: 动态插件生命周期声明必须作为自动发现契约的可选覆盖

系统 SHALL 将 `backend/lifecycle/*.yaml` 视为生命周期自动发现结果的可选 override。Override MAY 覆盖已发现 operation 的 `requestType`、`internalPath` 或 `timeoutMs`，但 MUST NOT 为插件中不存在的生命周期 handler 创建新的契约。构建工具 MUST 对重复 operation、非法 operation、非法 timeout 和无法匹配自动发现 handler 的 override 返回失败。

#### Scenario: Override 覆盖生命周期超时

- **WHEN** 动态插件 controller 暴露合法 `BeforeInstall` 生命周期方法
- **AND** `backend/lifecycle/001-before-install.yaml` 声明 `operation=BeforeInstall` 且 `timeoutMs=3000`
- **THEN** `build-wasm` 生成 `BeforeInstall` 生命周期契约
- **AND** 该契约的 timeout 使用 override 声明的 `3000` 毫秒

#### Scenario: Override 声明不存在的方法

- **WHEN** `backend/lifecycle/001-before-install.yaml` 声明 `operation=BeforeInstall`
- **AND** 动态插件 controller 未暴露合法 `BeforeInstall` handler
- **THEN** `build-wasm` 构建失败
- **AND** 错误信息指向该 lifecycle override 找不到对应 handler

#### Scenario: Override 重复声明 operation

- **WHEN** `backend/lifecycle` 下存在两个声明 `operation=BeforeInstall` 的 YAML 文件
- **THEN** `build-wasm` 构建失败
- **AND** 错误信息指向重复的 lifecycle operation

### Requirement: 官方动态示例插件必须通过自动发现声明生命周期

官方动态示例插件 SHALL 依赖 controller 方法自动发现生成生命周期契约，不再要求维护重复的 `backend/lifecycle/*.yaml` 文件。示例插件打包后的 artifact MUST 仍包含与源码插件一致命名的生命周期契约，并覆盖安装、升级、禁用、卸载、租户禁用、租户删除和安装模式切换的前置及后置处理器。

#### Scenario: 示例插件无手写 lifecycle YAML 仍生成完整契约

- **WHEN** 构建 `plugin-demo-dynamic`
- **AND** 示例插件未维护 `backend/lifecycle/*.yaml`
- **THEN** 构建产物包含 `BeforeInstall`、`AfterInstall`、`BeforeUpgrade`、`AfterUpgrade`、`BeforeDisable`、`AfterDisable`、`BeforeUninstall`、`AfterUninstall`、`BeforeTenantDisable`、`AfterTenantDisable`、`BeforeTenantDelete`、`AfterTenantDelete`、`BeforeInstallModeChange` 和 `AfterInstallModeChange` 生命周期契约
- **AND** 宿主运行时解析 artifact 后可以按既有生命周期流程调用这些处理器

### Requirement: 生命周期 manifest snapshot 必须使用共享 typed bridge contract

系统 SHALL 使用 `pluginbridge/contract` 中的 typed manifest snapshot DTO 作为动态插件生命周期请求和源码插件升级回调的唯一 manifest snapshot 发布契约。动态插件 `LifecycleRequest.fromManifest` 与 `LifecycleRequest.toManifest` MUST 使用 typed DTO，不得通过手写 `map[string]interface{}` 字段名构造。源码插件侧 manifest snapshot wrapper MUST 复用同一个 DTO，避免 source plugin 与 dynamic plugin 维护两套字段名。

#### Scenario: 动态生命周期请求发布 typed manifest snapshot

- **WHEN** 宿主为动态插件 `BeforeUpgrade`、`Upgrade` 或 `AfterUpgrade` 构建 lifecycle request
- **THEN** `fromManifest` 和 `toManifest` 使用共享 typed manifest snapshot DTO 序列化
- **AND** manifest snapshot 字段由 DTO 的 JSON 标签定义
- **AND** 构建请求的运行时代码不得手写 manifest snapshot map key

#### Scenario: 源码插件和动态插件复用同一 manifest snapshot 契约

- **WHEN** 宿主为源码插件升级回调构建 `ManifestSnapshot`
- **THEN** 源码插件 wrapper 复用与动态插件生命周期请求相同的 typed manifest snapshot DTO
- **AND** 新增、删除或重命名 manifest snapshot 发布字段时必须通过编译期字段引用暴露所有未同步调用点

### Requirement: 动态插件数据面路由必须使用独立宿主命名空间

系统 SHALL 使用 `/x/{pluginId}/...` 作为动态插件数据面路由的 canonical 公开入口。宿主只将 `/x` 识别为动态插件分发命名空间，并只从路径中解析 `pluginId`；`{pluginId}` 之后的路径 SHALL 完全归插件路由声明所有。宿主 MUST NOT 将动态插件数据面路由固定在宿主控制面 `/api/v1` 前缀下，也不得限制插件在自身路径中声明 `/api/v1`、`/api/v2`、`/graphql` 或其他插件自有路径结构。

#### Scenario: 插件声明自己的 API 版本

- **WHEN** 动态插件 `plugin-a` 声明内部路由 `/api/v2/items`
- **THEN** 宿主以 `/x/plugin-a/api/v2/items` 作为 canonical 公开路径
- **AND** 宿主不得生成 `/api/v1/extensions/plugin-a/api/v2/items` 作为 canonical 公开路径

#### Scenario: 插件声明非 REST 版本路径

- **WHEN** 动态插件 `plugin-a` 声明内部路由 `/graphql`
- **THEN** 宿主以 `/x/plugin-a/graphql` 分发该请求
- **AND** 宿主不得要求插件路径包含宿主 API 版本段

### Requirement: 动态插件旧扩展路由不得继续作为分发入口

系统 MUST NOT 继续接受旧 `/api/v1/extensions/{pluginId}/...` 作为动态插件数据面分发入口。OpenAPI 投影、插件资源列表、示例插件前端和新文档 MUST 使用 `/x/{pluginId}/...` 作为公开路径。

#### Scenario: 旧扩展路径不再分发动态插件请求

- **WHEN** 客户端请求 `/api/v1/extensions/plugin-a/backend-summary`
- **THEN** 宿主不得按动态插件 `plugin-a` 的 `/backend-summary` 内部路由执行请求

#### Scenario: 新文档只展示新路径

- **WHEN** 宿主生成动态插件 OpenAPI 文档或插件资源列表
- **THEN** 动态插件公开路径以 `/x/{pluginId}/...` 开头
- **AND** 新生成内容不得把 `/api/v1/extensions/{pluginId}/...` 作为动态插件公开路径

### Requirement: 动态插件根级路由必须保留宿主 HTTP 治理链路

系统 SHALL 在根级 `/x` 动态插件路由上复用宿主统一 HTTP 治理链路。请求在进入动态插件 bridge 执行前 MUST 经过响应包装、CORS、请求体限制、业务上下文初始化、运行时 freshness 检查、动态插件路由准备、登录鉴权和权限校验。路由前缀迁移 MUST NOT 绕过插件启用状态、运行时修订号、数据权限上下文或审计元数据构建。

#### Scenario: 未认证用户访问需要登录的插件路由

- **WHEN** 未认证用户请求 `/x/plugin-a/private-summary`
- **AND** 动态插件路由声明需要登录访问
- **THEN** 宿主拒绝该请求
- **AND** 拒绝结果使用宿主统一响应格式

#### Scenario: 插件禁用后新前缀不可继续暴露能力

- **WHEN** 动态插件 `plugin-a` 被禁用
- **THEN** 后续访问 `/x/plugin-a/backend-summary` 不得继续执行该插件 bridge 路由

#### Scenario: 动态路由元数据使用实际命中路径

- **WHEN** 请求通过 `/x/plugin-a/backend-summary` 命中动态插件路由
- **THEN** 传递给动态插件 bridge 和宿主中间件的 public path 元数据反映实际命中的 `/x/plugin-a/backend-summary`

