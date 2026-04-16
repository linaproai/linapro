## 1. 路由合同与产物装载

- [x] 1.1 定义动态插件路由合同结构，补充固定前缀、治理字段与校验规则
- [x] 1.2 在构建链路中从`backend/api/**/*.go`提取`g.Meta`路由元数据，并嵌入运行时产物
- [x] 1.3 在宿主装载链路中解析动态路由合同并回填到`manifest.Routes`

## 2. 宿主分发与治理骨架

- [x] 2.1 实现`/api/v1/extensions/{pluginId}/...`统一分发入口，仅让固定前缀请求进入动态插件链路
- [x] 2.2 实现宿主侧动态路由匹配、登录校验、权限校验和登录态上下文注入
- [x] 2.3 对当前未接入执行器的动态路由返回明确的`501`占位响应
- [x] 2.4 抽象动态路由执行器接口、请求／响应快照骨架与稳定`v1` bridge envelope，先由默认占位执行器承接
- [x] 2.5 将动态路由匹配结果绑定到当前`active release`运行时快照，并按`runtimeKind`选择执行器
- [x] 2.6 在运行时产物中嵌入动态路由 bridge ABI 合同，固定未来执行入口与二进制编解码约定

## 3. 权限与文档投影

- [x] 3.1 基于动态路由`permission`声明自动生成隐藏合成权限菜单，并复用`sys_menu.perms`
- [x] 3.2 在插件启用、禁用、卸载时联动同步合成权限菜单
- [x] 3.3 将已启用动态插件的路由合同投影到宿主`OpenAPI`，展示固定公开路径

## 4. 样例与测试

- [x] 4.1 更新`plugin-demo-dynamic`样例路由元数据，覆盖最小治理字段
- [x] 4.2 补充后端测试，覆盖路由合同提取、产物装载、路径匹配、权限校验规则与合成权限菜单
- [x] 4.3 补充构建器测试，覆盖动态路由合同随运行时产物输出

## 5. 真实执行阶段

- [x] 5.1 补齐真实`Wasm`动态路由执行器，并将当前快照骨架接到实际运行时桥接
- [x] 5.2 在执行器接入后补齐插件本地中间件执行链与真实响应回写
- [x] 5.3 将`OpenAPI`投影升级为“可执行路由展示`200/500`，占位路由展示`501`”的运行时感知模型
- [x] 5.4 更新`hack/tests/e2e/plugin/TC0067-runtime-wasm-lifecycle.ts`，补充`TC-67j`验证动态路由返回真实`Wasm bridge`响应

## Feedback

- [x] **FB-1**：移除`apps/lina-core/internal/service/plugin`中的编译阶段实现与调用逻辑，统一收敛到`hack/build-wasm`
- [x] **FB-2**：将宿主到动态插件的 bridge `DTO`编解码固定为高效二进制协议，禁止使用`json`或纯文本协议
- [x] **FB-3**：将 bridge envelope、codec、guest 侧适配和错误响应等可复用逻辑抽象到`apps/lina-core/pkg`公共组件
- [x] **FB-4**：为本次迭代新增的后端动态插件关键逻辑补充注释，明确运行时分发、Wasm bridge、产物一致性校验与刷新策略的实现意图
- [x] **FB-5**：将动态插件固定前缀路由改为宿主统一`RouterGroup + Middleware`注册方式，复用宿主通用中间件注册链并消除独立分发入口的差异化维护
- [x] **FB-6**：删除`plugin.go`及其引用的孤立`controller/service`代码，消除动态插件中的`SourcePlugin`错误用法和`/api/v1`路由前缀违规
- [x] **FB-7**：修复`E2E`测试`TC-67j`路由路径、权限和响应头与实际`API`定义和`WASM`处理器的不匹配

## 6. Host Functions（宿主回调能力）

- [x] 6.1 在`pluginbridge`包新增 Host Call 协议层：opcode 常量、能力标识、映射函数、校验函数、host call 状态码
- [x] 6.2 在`pluginbridge`包新增 Host Call 编解码层：通用响应信封与各 opcode 的请求/响应消息`protowire`编解码
- [x] 6.3 在`pluginbridge`包新增`Wasm`产物能力声明区段常量`lina.plugin.backend.capabilities`
- [x] 6.4 扩展`pluginbridge` Guest SDK：新增`guestHostCallResponseBuffer`独立缓冲区、`HostCallAlloc`方法、`HostCallResponseBuffer`方法
- [x] 6.5 新建 Guest 侧 Host Call 高级封装（`//go:build wasip1`）：`//go:wasmimport`导入声明与`HostLog`、`HostStateGet/Set/Delete`、`HostDBQuery/Execute`等`API`
- [x] 6.6 新建宿主侧 Host Call 上下文结构体`hostCallContext`：`pluginID`、`capabilities`、`service`引用与`context.Context`注入/提取
- [x] 6.7 新建宿主侧 Host Call 分发器：在`wazero Runtime`注册`lina_env`模块导出`host_call`函数，实现能力校验、opcode 分发、Guest 响应缓冲区分配与内存写入
- [x] 6.8 实现`host:log`处理器：解码日志请求、自动附加`[plugin:{pluginID}]`前缀、按级别调用`logger`组件
- [x] 6.9 实现`host:state`处理器：`sys_plugin_state`表的`Get/Set/Delete`操作，按`pluginID`自动隔离
- [x] 6.10 实现`host:db:query`处理器：`SELECT`前缀校验、`DDL`关键词黑名单、`maxRows`上限 1000
- [x] 6.11 实现`host:db:execute`处理器：`INSERT/UPDATE/DELETE/REPLACE`前缀校验、`DDL`黑名单、`SELECT`拒绝
- [x] 6.12 新建`sys_plugin_state`表`DDL`文件（`012-plugin-host-call.sql`）
- [x] 6.13 修改`Wasm`运行时加载链路：在 WASI 实例化后、模块编译前注册 Host Call 模块；实例化后注入`hostCallContext`
- [x] 6.14 修改`pluginManifest`结构体：新增`HostCapabilities`字段，从产物能力声明段加载
- [x] 6.15 修改`pluginDynamicArtifact`：新增`Capabilities`字段，解析`lina.plugin.backend.capabilities`自定义段
- [x] 6.16 修改构建器：在`pluginManifest`新增`Capabilities`字段、校验能力字符串合法性、嵌入`Wasm`自定义段
- [x] 6.17 更新`plugin-demo-dynamic`清单：在`plugin.yaml`新增`capabilities`声明
- [x] 6.18 更新`plugin-demo-dynamic` Guest 运行时：新增`lina_host_call_alloc`导出、新增`/host-call-demo`路由演示日志+状态存储调用
- [x] 6.19 新建 Host Call Demo 路由合同（`HostCallDemoReq`/`HostCallDemoRes`）
- [x] 6.20 补充 Host Call 编解码往返测试（11 组消息类型全覆盖）
- [x] 6.21 补充 Host Call 宿主函数测试：能力校验、`SQL`关键词检测、`DB`查询/执行校验
- [x] 6.22 修复因路由重命名导致的集成测试回归（`/review-summary` → `/backend-summary`）
