## ADDED Requirements

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
