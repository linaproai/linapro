## MODIFIED Requirements

### Requirement: The plugin life cycle state machine can be managed

系统 SHALL 提供可审计的插件生命周期状态机，区分源码插件与动态插件的生命周期语义，并允许宿主在启动阶段通过主配置文件中的 `plugin.autoEnable` 把插件推进到启用状态。

#### Scenario: Source code plugin is compiled and integrated with the host
- **WHEN** 宿主编译包含源码插件的源码树并生成 LinaPro 二进制
- **THEN** 该源码插件的后端 Go 代码与宿主源码一起编译
- **AND** 源码插件进入“已发现可治理”的生命周期范围，而不是因为被编译进宿主就自动视为已安装或已启用
- **AND** 管理员或 `plugin.autoEnable` 后续仍可显式推进其安装与启用状态

#### Scenario: Source plugin stays discovered-only after the first synchronization
- **WHEN** 宿主首次发现源码插件并将其写入插件注册表
- **THEN** 该源码插件默认保持“仅发现、未安装、未启用”状态
- **AND** 后续常规同步不会自动把它升级为已安装或已启用

#### Scenario: Auto-enable list installs and enables a source plugin during host startup
- **WHEN** 宿主主配置文件中的 `plugin.autoEnable` 命中一个已发现的源码插件
- **THEN** 宿主在启动阶段先完成该源码插件安装，再推进到启用状态
- **AND** 该源码插件的路由、菜单、cron 与 Hook 只会在成功启用后对外生效

#### Scenario: Install dynamic plugins
- **WHEN** 管理员安装一个合法的 `wasm` 动态插件，或宿主主配置文件中的 `plugin.autoEnable` 要求一个已发现的动态插件在启动时被启用
- **THEN** 宿主创建插件安装记录与当前版本记录
- **AND** 宿主按顺序处理迁移、资源注册、权限接入以及前后端加载准备
- **AND** 在插件被显式推进到启用状态之前，普通用户仍不可见该插件能力

#### Scenario: Auto-enable list can request enabled state for dynamic plugins
- **WHEN** 宿主主配置文件中的 `plugin.autoEnable` 命中一个已发现的动态插件
- **THEN** 宿主把该动态插件推进到启用目标状态所需的共享生命周期流程
- **AND** 只有在安装、授权与 reconcile 成功后，该动态插件才对普通用户可见

#### Scenario: Disable plugin
- **WHEN** 管理员将一个已启用插件切换为禁用状态
- **THEN** 宿主停止暴露该插件的 Hook、Slot、页面与菜单
- **AND** 宿主保留插件业务数据、角色授权关系与安装记录
- **AND** 重新启用后可以恢复已有治理关系

#### Scenario: Uninstall dynamic plugins
- **WHEN** 管理员卸载一个动态插件
- **THEN** 宿主移除该插件在宿主侧注册的菜单、资源引用、运行时产物与挂载信息
- **AND** 默认不删除插件自身业务表或业务数据
- **AND** 宿主保留卸载审计信息
