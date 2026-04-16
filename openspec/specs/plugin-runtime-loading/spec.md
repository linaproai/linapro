# plugin-runtime-loading Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
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

#### Scenario: 构建器优先消费动态插件的嵌入资源声明
- **WHEN** 构建器为动态插件生成运行时 `wasm` 产物
- **THEN** 构建器必须优先从插件声明的嵌入文件系统读取 manifest、前端资源与 SQL 资源
- **AND** 构建器必须把这些资源转换为宿主已发布的自定义节快照
- **AND** 宿主在上传、装载与启用阶段继续只消费该快照，而不是执行 guest 资源读取逻辑

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

### Requirement: 动态插件运行时产物携带可治理的路由合同

系统 SHALL 允许动态插件在运行时产物中携带后端动态路由合同，宿主装载产物后能够恢复这些路由的路径、方法与最小治理元数据，而不需要在请求时再次扫描源码目录。

#### Scenario: 构建阶段提取动态路由合同

- **WHEN** 构建动态插件运行时产物
- **THEN** 构建器从`backend/api/**/*.go`中的请求结构体`g.Meta`提取动态路由元数据
- **AND** 将这些元数据写入运行时产物中的专用区段
- **AND** 宿主加载产物后可恢复为动态插件`manifest.Routes`

#### Scenario: 宿主校验动态路由合同

- **WHEN** 宿主装载一个动态插件的路由合同
- **THEN** 宿主校验内部路径、方法、`access`、`permission`与`operLog`是否合法
- **AND** `access`未声明时按`login`处理
- **AND** `public`路由不得声明`permission`
- **AND** 非法合同会导致该产物装载失败

### Requirement: 宿主按固定前缀分发动态插件路由

系统 SHALL 将动态插件公开接口固定在`/api/v1/extensions/{pluginId}/...`下，并仅让命中该前缀的请求进入动态插件路由分发链路。

#### Scenario: 非插件请求不进入动态分发链路

- **WHEN** 宿主收到一个未命中`/api/v1/extensions/{pluginId}/...`的请求
- **THEN** 该请求继续走宿主原有路由链
- **AND** 不会触发动态插件路由匹配

#### Scenario: 宿主按`pluginId`与内部路径匹配动态路由

- **WHEN** 宿主收到一个命中固定前缀的请求
- **THEN** 宿主先提取`pluginId`
- **AND** 仅在该插件的已启用动态路由集合内按方法和内部路径做匹配
- **AND** 支持`/path/{id}`形式的动态路径段匹配

#### Scenario: 动态插件固定前缀路由复用宿主统一中间件注册方式

- **WHEN** 宿主注册动态插件固定前缀路由
- **THEN** 该入口通过宿主标准`RouterGroup`与`group.Middleware(...)`方式挂载
- **AND** 继续复用宿主通用中间件注册链
- **AND** 动态路由特有的匹配、鉴权与权限校验阶段也以宿主中间件方式编排
- **AND** 宿主不得为动态插件维护一条脱离统一中间件注册方式的独立旁路入口

### Requirement: 动态路由通过受限`Wasm bridge`执行并保留占位回退

系统 SHALL 在完成路由命中与治理校验后，优先通过当前激活版本声明的受限`Wasm bridge`执行业务路由；若当前产物未声明可执行 bridge，则回退到明确的`501`占位响应。

#### Scenario: 命中受保护动态路由

- **WHEN** 一个登录型动态路由被成功匹配
- **THEN** 宿主先完成登录校验
- **AND** 若该路由声明了`permission`，宿主继续完成权限校验
- **AND** 若当前激活产物声明了可执行 bridge，则宿主通过`Wasm bridge`执行该路由并回写真实响应
- **AND** 若当前激活产物未声明可执行 bridge，则宿主返回`501`占位响应

#### Scenario: 命中公开动态路由

- **WHEN** 一个`public`动态路由被成功匹配
- **THEN** 宿主不得解析登录令牌
- **AND** 宿主不得注入用户业务上下文
- **AND** 若当前激活产物声明了可执行 bridge，则宿主通过`Wasm bridge`执行该路由并回写真实响应
- **AND** 若当前激活产物未声明可执行 bridge，则该路由返回`501`占位响应

### Requirement: 动态路由桥接`DTO`使用高效二进制编解码

系统 SHALL 使用版本化二进制编解码协议交换宿主与动态插件之间的请求／响应 bridge envelope，不得使用`json`或纯文本协议作为可执行 bridge 的请求／响应`DTO`编解码方式。

#### Scenario: 可执行 bridge 声明二进制 codec

- **WHEN** 构建器为包含可执行 bridge 的动态插件写入`lina.plugin.backend.bridge`合同
- **THEN** 该合同的`requestCodec`声明为`protobuf`
- **AND** 该合同的`responseCodec`声明为`protobuf`
- **AND** 宿主按二进制协议编码请求信封并解码响应信封

#### Scenario: 宿主拒绝文本类桥接 codec

- **WHEN** 宿主装载一个声明可执行 bridge 的动态插件产物
- **AND** 该产物的`requestCodec`或`responseCodec`声明为`json`、`text`、`plain`或其他纯文本协议
- **THEN** 宿主拒绝将该 bridge 标记为可执行
- **AND** 不会在请求执行链路中使用文本协议对桥接`DTO`进行编解码

#### Scenario: 原始业务请求体按字节透传

- **WHEN** 客户端向动态插件公开接口提交请求体
- **THEN** 宿主将原始请求体作为字节序列放入桥接请求快照
- **AND** 不因业务请求体内容类型为`application/json`而把整个 bridge envelope 改用`json`编码

### Requirement: 宿主运行时插件组件不承载编译阶段逻辑

系统 SHALL 保持`lina-core`的`plugin`组件为运行时业务组件，只消费动态插件运行时产物和嵌入合同；源码扫描、`g.Meta`静态提取、`Wasm`编译和自定义区段写入等编译阶段逻辑必须收敛到`hack/build-wasm`或`hack/`下的独立工具。

#### Scenario: 宿主只消费运行时产物

- **WHEN** 宿主启用、切换或执行一个动态插件
- **THEN** 宿主从当前`active release`产物中读取已嵌入的路由合同与 bridge 合同
- **AND** 宿主不得调用构建器执行源码扫描、编译或产物生成

#### Scenario: 构建阶段由独立工具完成

- **WHEN** 开发者构建动态插件运行时产物
- **THEN** `hack/build-wasm`或`hack/`下的独立工具负责源码扫描、合同提取、`Wasm`编译和自定义区段写入
- **AND** `apps/lina-core/internal/service/plugin`中不包含编译阶段调用链路

### Requirement: 动态插件支持 Host Functions 宿主回调

系统 SHALL 允许`Wasm`动态插件通过受控的 Host Functions 机制回调宿主，在 Guest 运行时内安全地使用宿主提供的日志、状态存储和数据库读写服务。

#### Scenario: 插件声明所需宿主能力

- **WHEN** 开发者在`plugin.yaml`中声明`capabilities`字段
- **THEN** 构建器校验能力字符串合法性
- **AND** 将能力列表嵌入`Wasm`自定义段`lina.plugin.backend.capabilities`
- **AND** 宿主装载产物后将能力列表恢复到`manifest.HostCapabilities`

#### Scenario: 宿主在运行时校验 Host Call 能力

- **WHEN** Guest 通过`lina_env.host_call(opcode, reqPtr, reqLen)`调用宿主
- **THEN** 宿主按 opcode 查找对应能力标识
- **AND** 校验当前插件是否在清单中声明了该能力
- **AND** 未声明的能力调用返回`capability_denied`状态

#### Scenario: Host Call 使用独立响应缓冲区

- **WHEN** 宿主处理完 Host Call 请求并需要返回响应
- **THEN** 宿主调用 Guest 导出的`lina_host_call_alloc(size)`分配独立缓冲区
- **AND** 该缓冲区与主请求/响应缓冲区互不干扰
- **AND** 宿主将响应写入该缓冲区并返回打包的指针和长度

### Requirement: Host Call 日志能力按`pluginID`标识来源

系统 SHALL 为声明了`host:log`能力的动态插件提供结构化日志输出，自动附加插件标识前缀。

#### Scenario: Guest 通过 Host Call 输出日志

- **WHEN** Guest 调用`HostLog(level, message, fields)`
- **THEN** 宿主使用项目`logger`组件输出日志
- **AND** 日志自动附加`[plugin:{pluginID}]`前缀
- **AND** 日志级别映射到宿主`logger`对应级别

### Requirement: Host Call 状态存储按`pluginID`隔离

系统 SHALL 为声明了`host:state`能力的动态插件提供键值状态存储，所有操作按`pluginID`自动隔离。

#### Scenario: Guest 通过 Host Call 读写状态

- **WHEN** Guest 调用`HostStateGet/Set/Delete`
- **THEN** 宿主在`sys_plugin_state`表中按`pluginID`和`stateKey`执行操作
- **AND** 插件无法访问其他插件的状态
- **AND** `StateSet`使用`INSERT ... ON DUPLICATE KEY UPDATE`实现幂等写入

### Requirement: Host Call 数据库访问受`SQL`前缀和关键词约束

系统 SHALL 为声明了`host:db:query`或`host:db:execute`能力的动态插件提供受限的数据库访问，通过`SQL`前缀校验和`DDL`关键词黑名单防护。

#### Scenario: 只读查询校验

- **WHEN** Guest 调用`HostDBQuery`
- **THEN** 宿主校验`SQL`语句前缀为`SELECT`
- **AND** 宿主拒绝包含`DDL`关键词（`DROP`、`ALTER`、`CREATE`、`TRUNCATE`、`GRANT`、`REVOKE`）的语句
- **AND** 查询结果受`maxRows`上限（最大 1000）限制

#### Scenario: 写入操作校验

- **WHEN** Guest 调用`HostDBExecute`
- **THEN** 宿主校验`SQL`语句前缀为`INSERT`、`UPDATE`、`DELETE`或`REPLACE`
- **AND** 宿主拒绝`DDL`关键词和`SELECT`语句
- **AND** 返回受影响行数和最后插入`ID`

### Requirement: 动态插件运行时产物携带宿主服务治理快照

系统 SHALL 让动态插件运行时产物携带结构化宿主服务声明、方法授权和资源申请快照，供宿主在装载时恢复当前 release 的宿主服务治理信息；宿主内部 capability 分类快照必须基于该`hostServices`快照自动推导，而不是由 guest 额外嵌入第二份作者侧声明。

#### Scenario: 构建器写入宿主服务快照

- **WHEN** 构建器生成动态插件运行时产物
- **THEN** 构建器将归一化后的`hostServices`声明写入专用自定义节
- **AND** 不再写入作者侧顶层`capabilities`自定义节
- **AND** 对未知 service、method 或非法策略参数直接报错

#### Scenario: 宿主恢复宿主服务快照

- **WHEN** 宿主装载一个动态插件运行时产物
- **THEN** 宿主恢复结构化宿主服务策略并据此推导能力分类集合
- **AND** 将其挂到当前 active release 的运行时 manifest
- **AND** 缺失或损坏的宿主服务快照会阻止对应宿主服务进入可执行状态

### Requirement: 动态插件运行时按统一宿主服务协议执行 Host Call

系统 SHALL 根据动态插件声明的宿主服务协议版本，通过统一宿主服务分发器执行 Host Call，不再为历史探索性实现保留平行公开协议。

#### Scenario: 运行时统一走宿主服务分发器

- **WHEN** 一个动态插件声明了结构化宿主服务协议
- **THEN** 宿主通过统一宿主服务分发器处理该调用
- **AND** 宿主不得再为同类新增能力暴露新的专用 opcode

