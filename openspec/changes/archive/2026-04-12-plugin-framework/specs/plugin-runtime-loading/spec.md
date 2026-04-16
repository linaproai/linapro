## ADDED Requirements

### Requirement: 源码插件通过目录约定发现资源并由显式注册表装载
系统 SHALL 支持按目录约定发现源码插件资源，并通过集中维护的显式注册表装载源码插件后端入口。

#### Scenario: 扫描源码插件目录资源
- **WHEN** 宿主执行后端或前端构建流程
- **THEN** 宿主扫描 `apps/lina-plugins/` 下所有合法源码插件
- **AND** 按目录约定发现插件清单、SQL、前端页面与 Slot 资源
- **AND** 缺少必需入口或 manifest 不合法的插件会阻止对应插件接入

#### Scenario: 源码插件 Go 后端通过显式注册表参与宿主编译
- **WHEN** 一个源码插件在插件目录内提供后端 Go 入口
- **THEN** 开发者在 `apps/lina-plugins/lina-plugins.go` 中追加该插件后端包的匿名导入
- **AND** 该插件的后端 Go 包与宿主后端一起编译进同一个二进制文件
- **AND** 插件作者不需要手工修改宿主控制器、路由骨架或其他分散装配点来接线该插件

### Requirement: 动态 `wasm` 插件可被校验和装载
系统 SHALL 支持安装 `dynamic` `wasm` 动态插件产物，并在装载前完成完整性与兼容性校验。

#### Scenario: 安装 wasm 单文件插件
- **WHEN** 管理员上传一个单独的 `wasm` 文件
- **THEN** 宿主读取该文件中声明的插件元数据与可选资源信息
- **AND** 若插件仅声明后端能力则无需额外前端资源即可安装
- **AND** 若插件声明前端资源则宿主仅在资源可被正确提取时允许启用

### Requirement: 插件启停与升级无需重启宿主
系统 SHALL 支持在不重启宿主进程的情况下启用、禁用与升级动态插件。

#### Scenario: 热启用插件
- **WHEN** 管理员启用一个已安装但未启用的动态插件
- **THEN** 宿主在当前进程内加载该插件 release 并更新本地插件注册表
- **AND** 新请求可以立即访问该插件提供的页面、Hook 与治理资源
- **AND** 宿主主进程不需要重启

#### Scenario: 热升级插件
- **WHEN** 管理员将动态插件升级到新 release
- **THEN** 宿主为新请求切换到新 release
- **AND** 已经开始处理的旧请求允许自然结束
- **AND** 正在使用该插件页面的用户会收到刷新当前页面的提示

#### Scenario: staged 上传不立即替换当前服务 release
- **WHEN** 管理员上传一个更高版本的动态插件 `wasm`
- **THEN** 宿主先将该产物写入 staging 存储路径并记录为待切换 release
- **AND** 当前 active release 继续通过其稳定归档路径服务已有请求与旧页面
- **AND** 只有在主节点 Reconciler 成功推进代际切换后，新 release 才会成为对外服务的 active release

#### Scenario: 升级失败后继续服务稳定 release
- **WHEN** 动态插件在升级、迁移、菜单切换或前端 bundle 预热阶段失败
- **THEN** 宿主回滚到上一个稳定 release 并恢复其 generation/release_id
- **AND** 失败 release 的静态资源和运行时状态不会继续对普通用户生效
- **AND** 当前稳定 release 的 Hook、资源查询和页面访问能力继续可用

### Requirement: 多节点以代际方式收敛插件状态
系统 SHALL 在多节点部署下通过代际同步机制传播插件变更，并避免重复迁移与双重切换。

#### Scenario: 主节点执行插件升级
- **WHEN** 多节点环境中发生插件安装、启用、禁用或升级
- **THEN** 只有被选举出来的主节点执行共享迁移与 release 切换
- **AND** 其他节点仅根据最新代际收敛本地状态
- **AND** 任一节点都可以上报其当前代际与错误状态

#### Scenario: 当前节点持续上报代际收敛状态
- **WHEN** 主节点已经切换某个动态插件的 active release 或者回滚到稳定 release
- **THEN** 每个节点都会基于最新 `generation/release_id` 更新自己的 `sys_plugin_node_state`
- **AND** 若当前节点无法加载对应 release，则该节点会把本地投影标记为失败并保留诊断信息
