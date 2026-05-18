## ADDED Requirements

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
