## Why

当前`Lina`动态插件已经具备`wasm`产物治理、前端资源托管、生命周期管理与宿主装载能力，但动态插件后端`REST`扩展仍存在两个现实问题：

- 动态插件还没有稳定、低成本、可治理的动态路由接入面；
- 直接在当前仓库里落地“`GoFrame`私有`ghttp.Server`在`WASI`内真实执行路由”的方案，现阶段存在明显运行时兼容性风险。

结合前序探索，本次实现选择一条更务实的受限路线：让动态插件把路由合同和治理元数据稳定带入运行时产物，由宿主在固定前缀下完成解析、鉴权、权限校验、文档合并与权限物化；同时将真实执行层收敛为`Wasm bridge`受限运行时，而不是追求在`WASI`里完整复刻宿主`GoFrame`路由栈。

这也更符合动态插件的定位：它是面向业务扩展、面向`AI agent`自动生成、面向安全隔离的受限扩展模型，而不是源码插件的等价替代。

## What Changes

- 为动态插件新增基于`api`层`g.Meta`的路由合同提取能力，构建阶段从`backend/api/**/*.go`提取`path`、`method`、`tags`、`summary`、`dc`、`access`、`permission`、`operLog`等最小元数据。
- 在动态插件运行时产物中新增`lina.plugin.backend.routes`区段，宿主加载产物后可直接恢复动态路由合同并挂入`manifest.Routes`。
- 将动态插件编译阶段逻辑统一收敛到`hack/build-wasm`，宿主侧`plugin`组件只消费运行时产物，不再承载构建器实现或调用链路。
- 将动态插件公开路径统一收敛到固定前缀`/api/v1/extensions/{pluginId}/...`，宿主仅对该前缀请求进入动态插件分发链路。
- 在宿主新增动态路由解析与治理骨架：支持`pluginId`快速定位、方法与路径匹配、登录校验、权限校验，以及登录路由的业务上下文注入。
- 在宿主补充统一的动态路由执行器接口、请求／响应快照骨架和稳定`v1` bridge envelope，并接入真实`wasm`执行器。
- 将动态路由执行输入绑定到当前`active release`的运行时快照，并按`runtimeKind`选择执行器，为后续真实`wasm`桥接预留接入点。
- 在动态产物中新增 bridge ABI 合同区段，固定初始化入口、请求缓冲分配入口、执行入口以及请求／响应二进制编解码协议，禁止以`json`或纯文本协议承载桥接`DTO`。
- 在`lina-core/pkg`沉淀动态插件可复用公共组件，封装 bridge envelope、二进制 codec、guest 侧处理器适配和错误响应辅助，降低动态插件业务代码复杂度。
- 明确动态治理元数据规则：`access`默认`login`；`public`路由不能声明`permission`；`operLog`只有显式声明时才保留。
- 将动态路由声明的`permission`自动物化为隐藏的合成权限菜单，复用现有`sys_menu.perms`与角色授权体系。
- 将已启用动态插件的路由合同投影为宿主`OpenAPI`路径；可执行运行时展示真实`200/500`响应语义，未接入执行器时仍保留`501`占位说明。
- 改造`plugin-demo-dynamic`样例与相关测试，验证合同提取、产物装载、权限合成、前缀路由解析、真实`wasm`执行与文档投影闭环。

## Capabilities

### Modified Capabilities

- `plugin-runtime-loading`：动态插件从“仅携带静态清单与资源”扩展为“携带可治理的后端动态路由合同”，宿主可在固定前缀下完成路由解析与治理。
- `plugin-hook-slot-extension`：动态插件后端扩展治理统一收敛到`g.Meta`，并通过宿主自动物化权限项，而不是暴露宿主中间件自由拼装。
- `system-api-docs`：系统接口文档自动合并动态插件路由合同投影出来的公开路径，并展示当前阶段的可访问接口描述。

## Impact

- 后端新增动态路由合同模型、合同校验、产物嵌入与宿主装载逻辑。
- 动态插件源码到`Wasm`产物的构建职责统一由`hack/build-wasm`维护，宿主`plugin`组件与编译阶段解耦。
- 后端新增`lina-core/pkg`级别的动态 bridge 公共组件，宿主执行器、构建器测试与样例插件可复用同一套信封和 codec 合同。
- 动态插件菜单同步逻辑需要额外维护基于`permission`声明生成的隐藏权限节点。
- 宿主`HTTP`启动链路需要注册固定前缀分发入口，并在启动时合并动态插件接口文档。
- 宿主动态路由分发入口需要收敛到统一执行器调用，以便后续平滑替换为真实`Wasm`桥接。
- 后端新增受限`wasm bridge`执行层；只有声明可执行 bridge 的动态产物才会真正执行业务路由，其余产物仍返回`501`占位响应。
- 动态插件样例新增`backend/runtime/wasm/`入口，证明插件本地中间件、业务执行与响应回写可以在受限 bridge 内完成。

### Host Functions（宿主回调能力扩展）

- 为`Wasm`动态插件新增 Host Functions 回调机制，让 Guest 运行时能够安全地调用宿主提供的受控服务（日志、状态存储、数据库读写）。
- 在`pluginbridge`包新增 Host Call 协议层：以单一入口函数`lina_env.host_call(opcode, reqPtr, reqLen)`统一分发，新增能力不改变`Wasm`导入签名。
- 新增能力声明模型：插件在`plugin.yaml`中声明所需`Host Function`能力，构建器嵌入`Wasm`自定义段，宿主运行时按 opcode 映射校验，未声明的能力调用立即拒绝。
- Phase 1 支持四类能力：`host:log`（结构化日志）、`host:state`（插件隔离键值存储）、`host:db:query`（只读`SQL`查询）、`host:db:execute`（写入`SQL`，禁止`DDL`）。
- 新增`sys_plugin_state`表用于插件隔离的键值状态存储，按`pluginID`自动隔离。
- 数据库访问采用开放模式——插件可查询/操作所有表，安全性由管理员决定是否授予该能力来控制。`SQL`语句前缀校验 +`DDL`关键词黑名单防护。
- Guest SDK 提供高级封装`API`（`HostLog`、`HostStateGet/Set/Delete`、`HostDBQuery/Execute`），降低插件开发复杂度。
- 使用独立的`guestHostCallResponseBuffer`响应缓冲区，避免 Host Call 回调中与主请求缓冲区冲突。

## Additional Capabilities

- `plugin-host-call`：动态插件从"单向桥接"扩展为"双向桥接"，Guest 可通过受控的 Host Functions 回调宿主，实现日志输出、状态持久化和数据库读写等业务操作。

## Additional Impact

- 后端新增 Host Call 协议层（opcode 常量、能力校验、请求/响应编解码）。
- 后端新增 Host Call 分发器和四类能力处理器（日志、状态存储、数据库查询、数据库写入）。
- 新增`sys_plugin_state`表`DDL`（`012-plugin-host-call.sql`）。
- `pluginbridge`包新增 Guest 侧 Host Call SDK（`//go:build wasip1`）。
- `Wasm`运行时加载链路需在 WASI 实例化后注册 Host Call 模块。
- 构建器需校验并嵌入能力声明到`Wasm`自定义段。
- 动态插件样例新增`/host-call-demo`路由，演示日志+状态存储的完整 Host Call 工作流。
