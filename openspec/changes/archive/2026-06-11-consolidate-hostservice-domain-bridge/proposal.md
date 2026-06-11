## Why

当前新增一个插件领域能力仍需要在能力契约、能力目录、动态插件 guest 目录、guest client、协议 codec、host service descriptor、WASM dispatch 和双语说明中维护多处手写镜像代码。已有 descriptor 和覆盖测试只能发现漂移，不能减少领域能力扩展时的长期修改扇出。

本变更将动态插件 host service 的领域桥接边界收敛为公开协议 catalog、guest typed client 和宿主 registry dispatch 三个稳定接缝，降低新增领域能力的维护成本，同时保持`capability/<x>cap`作为领域契约 owner。

## What Changes

- 新增`pkg/plugin/pluginbridge/protocol/hostservices`公开 catalog，集中维护 service、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态。
- 将现有 descriptor 治理来源迁移到公开 catalog，避免宿主内部代码依赖`pluginbridge/internal`私有包，也避免 catalog 与 descriptor 形成两份事实源。
- 将普通领域 host service 的 payload 收敛为统一 JSON envelope；仅对`storage`、`cache`、`lock`、`data/recordstore`、`network`等有明确性能、资源或 wire 需求的能力保留专用二进制或`protowire`codec。
- 将宿主 WASM host service 分发改为显式注册的 registry 驱动：入口统一做 envelope 解码、授权和上下文构造，再按`service/method`查找 handler。
- 将宿主 dispatch 入口收敛到`internal/service/plugin/internal/wasm/hostservicedispatch`registry，并通过父包显式注册列表装配，避免`wasm_host_service.go`继续维护 service 级大 switch。
- 保留`pkg/plugin/capability/<x>cap`手写领域契约和`capability.Services`目录扩展；本变更不把领域接口本身改为生成代码。
- 保留双语`README`手写维护方式，不恢复或新增`generated:host-services`标记。
- **BREAKING**：项目无兼容性负担，内部包路径和动态插件 guest SDK 组织结构可按新边界重排；但动态插件`plugin.yaml hostServices`声明、service/method wire 字符串、授权快照、错误 envelope 和已有 payload 语义不得因本变更改变。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `pluginbridge-subcomponent-architecture`：要求 host service 描述源迁移为公开`protocol/hostservices`catalog，并明确普通领域 JSON envelope、特殊二进制 codec、guest client 边界和静态覆盖验证。
- `plugin-host-service-extension`：要求 WASM host service dispatch 由显式注册 registry 驱动，领域 dispatch 通过显式注册适配单元接入，并保持授权、数据权限、缓存敏感依赖和错误 envelope 语义不变。

## Impact

- 影响 Go 包：`apps/lina-core/pkg/plugin/capability`、`apps/lina-core/pkg/plugin/pluginbridge`、`apps/lina-core/pkg/plugin/pluginbridge/protocol`、`apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall`、`apps/lina-core/internal/service/plugin/internal/wasm`。
- 影响动态插件 host service 协议治理：catalog、descriptor、guest client selector、payload codec、host dispatcher 注册和相关测试需要统一校验。
- 影响测试和验证：需要覆盖 catalog/descriptor 一致性、registry lookup 与未知方法拒绝、普通领域 JSON payload round trip、特殊二进制 codec 保留、host dispatch 无 service 级 switch、Go import 边界和动态插件 guest 编译闭包。
- 不影响 HTTP API、SQL 迁移、前端页面、运行时用户可见文案或插件实例目录资源。
- 数据权限影响：不改变具体数据服务语义；所有动态插件通过宿主发布服务访问数据的路径仍必须在 handler 内保持与宿主 API 等价的数据权限和租户边界。
- 缓存一致性影响：不改变 cache、session、权限快照、插件状态或其他缓存权威源；registry handler 必须复用启动期注入的共享服务实例或共享后端。
- `i18n`影响：不新增运行时用户可见文案、API 文档源文本、语言包或翻译缓存；仅 OpenSpec 文档使用中文。
